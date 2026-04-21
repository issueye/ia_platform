package module

import (
	"context"
	"errors"
	"os"
	"sync"

	hostapi "iacommon/pkg/host/api"
	hostfs "iacommon/pkg/host/fs"
)

var (
	ErrHostNotConfigured       = errors.New("host is not configured")
	ErrCapabilityNotConfigured = errors.New("capability is not configured")
	ErrInvalidFSResult         = errors.New("invalid fs result")
	ErrInvalidHTTPResult       = errors.New("invalid http result")
)

type PlatformFSBridge struct {
	Host         hostapi.Host
	CapabilityID string

	mu sync.Mutex
}

func (b *PlatformFSBridge) ReadFile(path string) (string, error) {
	result, err := b.call("fs.read_file", map[string]any{"path": path})
	if err != nil {
		return "", err
	}
	data, ok := result["data"]
	if !ok {
		return "", ErrInvalidFSResult
	}
	switch typed := data.(type) {
	case []byte:
		return string(typed), nil
	case string:
		return typed, nil
	default:
		return "", ErrInvalidFSResult
	}
}

func (b *PlatformFSBridge) WriteFile(path, content string) (bool, error) {
	_, err := b.call("fs.write_file", map[string]any{
		"path":   path,
		"data":   []byte(content),
		"create": true,
		"trunc":  true,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) AppendFile(path, content string) (bool, error) {
	_, err := b.call("fs.append_file", map[string]any{"path": path, "data": []byte(content)})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) Exists(path string) (bool, error) {
	_, err := b.call("fs.stat", map[string]any{"path": path})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) Mkdir(path string, recursive ...bool) (bool, error) {
	recursiveValue := false
	if len(recursive) > 0 {
		recursiveValue = recursive[0]
	}
	_, err := b.call("fs.mkdir", map[string]any{"path": path, "recursive": recursiveValue})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) ReadDir(path string) ([]string, error) {
	result, err := b.call("fs.read_dir", map[string]any{"path": path})
	if err != nil {
		return nil, err
	}
	entriesAny, ok := result["entries"]
	if !ok {
		return nil, ErrInvalidFSResult
	}
	switch typed := entriesAny.(type) {
	case []hostfs.DirEntry:
		entries := make([]string, 0, len(typed))
		for _, entry := range typed {
			entries = append(entries, entry.Name)
		}
		return entries, nil
	case []any:
		entries := make([]string, 0, len(typed))
		for _, entry := range typed {
			switch item := entry.(type) {
			case hostfs.DirEntry:
				entries = append(entries, item.Name)
			case map[string]any:
				name, ok := item["name"].(string)
				if !ok {
					return nil, ErrInvalidFSResult
				}
				entries = append(entries, name)
			default:
				return nil, ErrInvalidFSResult
			}
		}
		return entries, nil
	default:
		return nil, ErrInvalidFSResult
	}
}

func (b *PlatformFSBridge) Stat(path string) (map[string]any, error) {
	result, err := b.call("fs.stat", map[string]any{"path": path})
	if err != nil {
		return nil, err
	}
	info, ok := result["info"]
	if !ok {
		return nil, ErrInvalidFSResult
	}
	switch typed := info.(type) {
	case hostfs.FileInfo:
		return map[string]any{
			"name":    typed.Name,
			"size":    typed.Size,
			"mode":    typed.Mode,
			"isDir":   typed.IsDir,
			"modUnix": typed.ModUnix,
		}, nil
	case map[string]any:
		return typed, nil
	default:
		return nil, ErrInvalidFSResult
	}
}

func (b *PlatformFSBridge) Rename(oldPath, newPath string) (bool, error) {
	_, err := b.call("fs.rename", map[string]any{"old_path": oldPath, "new_path": newPath})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) Remove(path string) (bool, error) {
	_, err := b.call("fs.remove", map[string]any{"path": path, "recursive": false})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) RemoveAll(path string) (bool, error) {
	_, err := b.call("fs.remove", map[string]any{"path": path, "recursive": true})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) Copy(src, dst string) (bool, error) {
	data, err := b.ReadFile(src)
	if err != nil {
		return false, err
	}
	return b.WriteFile(dst, data)
}

func (b *PlatformFSBridge) call(operation string, args map[string]any) (map[string]any, error) {
	capabilityID, err := b.capabilityIDForCall()
	if err != nil {
		return nil, err
	}
	if capabilityID == "" {
		return nil, ErrCapabilityNotConfigured
	}

	result, err := b.Host.Call(context.Background(), hostapi.CallRequest{
		CapabilityID: capabilityID,
		Operation:    operation,
		Args:         args,
	})
	if err != nil {
		return nil, err
	}
	if result.Value == nil {
		return map[string]any{}, nil
	}
	return result.Value, nil
}

func (b *PlatformFSBridge) capabilityIDForCall() (string, error) {
	if b == nil || b.Host == nil {
		return "", ErrHostNotConfigured
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.CapabilityID != "" {
		return b.CapabilityID, nil
	}

	capability, err := b.Host.AcquireCapability(context.Background(), hostapi.AcquireRequest{Kind: hostapi.CapabilityFS})
	if err != nil {
		return "", err
	}
	b.CapabilityID = capability.ID
	return b.CapabilityID, nil
}

type PlatformHTTPBridge struct {
	Host         hostapi.Host
	CapabilityID string

	mu sync.Mutex
}

func (b *PlatformHTTPBridge) Request(rawURL string, options ...map[string]any) (map[string]any, error) {
	req := map[string]any{"url": rawURL}
	if len(options) > 0 && options[0] != nil {
		for key, value := range options[0] {
			req[key] = value
		}
	}

	timeout := readBridgeTimeout(req)
	result, err := b.call("network.http_fetch", map[string]any{
		"url":        rawURL,
		"method":     readBridgeString(req, "method", "GET"),
		"headers":    readBridgeHeaders(req),
		"body":       readBridgeBody(req),
		"timeout_ms": timeout,
		"timeoutMS":  timeout,
	})
	if err != nil {
		return nil, err
	}
	return normalizeHTTPResult(result)
}

func (b *PlatformHTTPBridge) Get(rawURL string, options ...map[string]any) (map[string]any, error) {
	requestOptions := mergeHTTPOptions(options...)
	requestOptions["method"] = "GET"
	return b.Request(rawURL, requestOptions)
}

func (b *PlatformHTTPBridge) Post(rawURL string, options ...map[string]any) (map[string]any, error) {
	requestOptions := mergeHTTPOptions(options...)
	requestOptions["method"] = "POST"
	return b.Request(rawURL, requestOptions)
}

func (b *PlatformHTTPBridge) call(operation string, args map[string]any) (map[string]any, error) {
	capabilityID, err := b.capabilityIDForCall()
	if err != nil {
		return nil, err
	}
	if capabilityID == "" {
		return nil, ErrCapabilityNotConfigured
	}

	result, err := b.Host.Call(context.Background(), hostapi.CallRequest{
		CapabilityID: capabilityID,
		Operation:    operation,
		Args:         args,
	})
	if err != nil {
		return nil, err
	}
	if result.Value == nil {
		return map[string]any{}, nil
	}
	return result.Value, nil
}

func (b *PlatformHTTPBridge) capabilityIDForCall() (string, error) {
	if b == nil || b.Host == nil {
		return "", ErrHostNotConfigured
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.CapabilityID != "" {
		return b.CapabilityID, nil
	}

	capability, err := b.Host.AcquireCapability(context.Background(), hostapi.AcquireRequest{Kind: hostapi.CapabilityNetwork})
	if err != nil {
		return "", err
	}
	b.CapabilityID = capability.ID
	return b.CapabilityID, nil
}

func readBridgeString(values map[string]any, key string, fallback string) string {
	if values == nil {
		return fallback
	}
	value, ok := values[key]
	if !ok || value == nil {
		return fallback
	}
	text, ok := value.(string)
	if !ok || text == "" {
		return fallback
	}
	return text
}

func readBridgeHeaders(values map[string]any) map[string]string {
	if values == nil {
		return nil
	}
	value, ok := values["headers"]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case map[string]string:
		result := make(map[string]string, len(typed))
		for key, item := range typed {
			result[key] = item
		}
		return result
	case map[string]any:
		result := make(map[string]string, len(typed))
		for key, item := range typed {
			text, ok := item.(string)
			if ok {
				result[key] = text
			}
		}
		return result
	default:
		return nil
	}
}

func readBridgeBody(values map[string]any) any {
	if values == nil {
		return nil
	}
	return values["body"]
}

func readBridgeTimeout(values map[string]any) any {
	if values == nil {
		return nil
	}
	if value, ok := values["timeout_ms"]; ok {
		return value
	}
	if value, ok := values["timeoutMS"]; ok {
		return value
	}
	return nil
}

func normalizeHTTPResult(result map[string]any) (map[string]any, error) {
	status, ok := result["status"]
	if !ok {
		return nil, ErrInvalidHTTPResult
	}
	bodyValue, ok := result["body"]
	if !ok {
		return nil, ErrInvalidHTTPResult
	}

	bodyText, ok := normalizeHTTPBody(bodyValue)
	if !ok {
		return nil, ErrInvalidHTTPResult
	}

	return map[string]any{
		"ok":      httpStatusOK(status),
		"status":  status,
		"headers": normalizeHTTPHeaders(result["headers"]),
		"body":    bodyText,
	}, nil
}

func normalizeHTTPHeaders(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]string:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = item
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = item
		}
		return result
	default:
		return map[string]any{}
	}
}

func normalizeHTTPBody(value any) (string, bool) {
	switch typed := value.(type) {
	case []byte:
		return string(typed), true
	case string:
		return typed, true
	default:
		return "", false
	}
}

func httpStatusOK(value any) bool {
	status, ok := toHTTPStatusCode(value)
	if !ok {
		return false
	}
	return status >= 200 && status < 300
}

func toHTTPStatusCode(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int8:
		return int64(typed), true
	case int16:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case uint:
		return int64(typed), true
	case uint8:
		return int64(typed), true
	case uint16:
		return int64(typed), true
	case uint32:
		return int64(typed), true
	case uint64:
		return int64(typed), true
	case float32:
		return int64(typed), true
	case float64:
		return int64(typed), true
	default:
		return 0, false
	}
}

func mergeHTTPOptions(options ...map[string]any) map[string]any {
	merged := map[string]any{}
	for _, option := range options {
		if option == nil {
			continue
		}
		for key, value := range option {
			merged[key] = value
		}
	}
	return merged
}
