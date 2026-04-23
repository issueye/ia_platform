package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	hostfs "iacommon/pkg/host/fs"
	hostnet "iacommon/pkg/host/network"
)

var (
	ErrCapabilityNotFound    = errors.New("capability not found")
	ErrProviderUnavailable   = errors.New("provider unavailable")
	ErrInvalidCallArgs       = errors.New("invalid call args")
	ErrPollNotSupported      = errors.New("poll is not supported")
	ErrCapabilityUnsupported = errors.New("capability is not supported")
)

type DefaultHost struct {
	FS      hostfs.Provider
	Network hostnet.Provider

	mu           sync.Mutex
	capabilities map[string]CapabilityInstance
	nextCapID    uint64
	fileHandles  map[uint64]hostfs.FileHandle
	nextHandleID uint64
}

func (h *DefaultHost) AcquireCapability(ctx context.Context, req AcquireRequest) (CapabilityInstance, error) {
	if err := ctx.Err(); err != nil {
		return CapabilityInstance{}, err
	}

	instance, err := h.newCapabilityInstance(req)
	if err != nil {
		return CapabilityInstance{}, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.capabilities == nil {
		h.capabilities = map[string]CapabilityInstance{}
	}
	if instance.ID == "" {
		h.nextCapID++
		instance.ID = fmt.Sprintf("cap-%d", h.nextCapID)
	}
	h.capabilities[instance.ID] = instance
	return instance, nil
}

func (h *DefaultHost) ReleaseCapability(ctx context.Context, capID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.capabilities == nil {
		return fmt.Errorf("%w: %s", ErrCapabilityNotFound, capID)
	}
	if _, ok := h.capabilities[capID]; !ok {
		return fmt.Errorf("%w: %s", ErrCapabilityNotFound, capID)
	}
	delete(h.capabilities, capID)
	return nil
}

func (h *DefaultHost) Call(ctx context.Context, req CallRequest) (CallResult, error) {
	if err := ctx.Err(); err != nil {
		return CallResult{}, err
	}

	capability, err := h.lookupCapability(req.CapabilityID)
	if err != nil {
		return CallResult{}, err
	}

	switch capability.Kind {
	case CapabilityFS:
		return h.callFS(ctx, req)
	case CapabilityNetwork:
		return h.callNetwork(ctx, req)
	default:
		return CallResult{}, fmt.Errorf("%w: %s", ErrCapabilityUnsupported, capability.Kind)
	}
}

func (h *DefaultHost) Poll(ctx context.Context, handleID uint64) (PollResult, error) {
	if err := ctx.Err(); err != nil {
		return PollResult{}, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.fileHandles != nil {
		if _, ok := h.fileHandles[handleID]; ok {
			return PollResult{
				Done:  true,
				Value: map[string]any{"ready": true, "handle": handleID},
			}, nil
		}
	}
	return PollResult{}, fmt.Errorf("%w: %d", ErrCapabilityNotFound, handleID)
}

func (h *DefaultHost) newCapabilityInstance(req AcquireRequest) (CapabilityInstance, error) {
	switch req.Kind {
	case CapabilityFS:
		if h == nil || h.FS == nil {
			return CapabilityInstance{}, fmt.Errorf("%w: %s", ErrProviderUnavailable, CapabilityFS)
		}
		return CapabilityInstance{
			Kind:   CapabilityFS,
			Rights: readStringSlice(req.Config, "rights"),
			Meta:   cloneMap(req.Config),
		}, nil
	case CapabilityNetwork:
		if h == nil || h.Network == nil {
			return CapabilityInstance{}, fmt.Errorf("%w: %s", ErrProviderUnavailable, CapabilityNetwork)
		}
		return CapabilityInstance{
			Kind:   CapabilityNetwork,
			Rights: readStringSlice(req.Config, "rights"),
			Meta:   cloneMap(req.Config),
		}, nil
	default:
		return CapabilityInstance{}, fmt.Errorf("%w: %s", ErrCapabilityUnsupported, req.Kind)
	}
}

func (h *DefaultHost) lookupCapability(capID string) (CapabilityInstance, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.capabilities == nil {
		return CapabilityInstance{}, fmt.Errorf("%w: %s", ErrCapabilityNotFound, capID)
	}
	capability, ok := h.capabilities[capID]
	if !ok {
		return CapabilityInstance{}, fmt.Errorf("%w: %s", ErrCapabilityNotFound, capID)
	}
	return capability, nil
}

func (h *DefaultHost) callFS(ctx context.Context, req CallRequest) (CallResult, error) {
	if h == nil || h.FS == nil {
		return CallResult{}, fmt.Errorf("%w: %s", ErrProviderUnavailable, CapabilityFS)
	}

	switch req.Operation {
	case "fs.open":
		parsed, err := decodeFSOpenRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		handle, err := h.FS.Open(ctx, parsed.Path, parsed.Opts)
		if err != nil {
			return CallResult{}, err
		}
		handleID := h.storeFileHandle(handle)
		return encodeFSOpenResponse(handleID), nil
	case "fs.read":
		parsed, err := decodeFSReadRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		handle, err := h.lookupFileHandle(parsed.Handle)
		if err != nil {
			return CallResult{}, err
		}
		buf := make([]byte, parsed.Size)
		n, err := handle.Read(ctx, buf)
		eof := false
		if err != nil {
			if errors.Is(err, io.EOF) {
				eof = true
			} else {
				return CallResult{}, err
			}
		}
		return encodeFSReadResponse(buf[:n], int64(n), eof), nil
	case "fs.write":
		parsed, err := decodeFSWriteRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		handle, err := h.lookupFileHandle(parsed.Handle)
		if err != nil {
			return CallResult{}, err
		}
		n, err := handle.Write(ctx, parsed.Data)
		if err != nil {
			return CallResult{}, err
		}
		return encodeFSWriteResponse(int64(n)), nil
	case "fs.seek":
		parsed, err := decodeFSSeekRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		handle, err := h.lookupFileHandle(parsed.Handle)
		if err != nil {
			return CallResult{}, err
		}
		offset, err := handle.Seek(ctx, parsed.Offset, int(parsed.Whence))
		if err != nil {
			return CallResult{}, err
		}
		return encodeFSSeekResponse(offset), nil
	case "fs.close":
		parsed, err := decodeFSCloseRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		handle, err := h.releaseFileHandle(parsed.Handle)
		if err != nil {
			return CallResult{}, err
		}
		if err := handle.Close(ctx); err != nil {
			return CallResult{}, err
		}
		return emptyCallResult(), nil
	case "fs.read_file":
		parsed, err := decodeFSReadFileRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		data, err := h.FS.ReadFile(ctx, parsed.Path)
		if err != nil {
			return CallResult{}, err
		}
		return encodeFSReadFileResponse(data), nil
	case "fs.write_file":
		parsed, err := decodeFSWriteFileRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		err = h.FS.WriteFile(ctx, parsed.Path, parsed.Data, parsed.Opts)
		if err != nil {
			return CallResult{}, err
		}
		return emptyCallResult(), nil
	case "fs.append_file":
		parsed, err := decodeFSAppendFileRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		if err := h.FS.AppendFile(ctx, parsed.Path, parsed.Data); err != nil {
			return CallResult{}, err
		}
		return emptyCallResult(), nil
	case "fs.read_dir":
		parsed, err := decodeFSReadDirRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		entries, err := h.FS.ReadDir(ctx, parsed.Path)
		if err != nil {
			return CallResult{}, err
		}
		return encodeFSReadDirResponse(entries), nil
	case "fs.stat":
		parsed, err := decodeFSStatRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		info, err := h.FS.Stat(ctx, parsed.Path)
		if err != nil {
			return CallResult{}, err
		}
		return encodeFSStatResponse(info), nil
	case "fs.mkdir":
		parsed, err := decodeFSMkdirRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		if err := h.FS.Mkdir(ctx, parsed.Path, parsed.Opts); err != nil {
			return CallResult{}, err
		}
		return emptyCallResult(), nil
	case "fs.remove":
		parsed, err := decodeFSRemoveRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		if err := h.FS.Remove(ctx, parsed.Path, parsed.Opts); err != nil {
			return CallResult{}, err
		}
		return emptyCallResult(), nil
	case "fs.rename":
		parsed, err := decodeFSRenameRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		if err := h.FS.Rename(ctx, parsed.OldPath, parsed.NewPath); err != nil {
			return CallResult{}, err
		}
		return emptyCallResult(), nil
	default:
		return CallResult{}, fmt.Errorf("unknown fs operation: %w: %s", ErrCapabilityUnsupported, req.Operation)
	}
}

func (h *DefaultHost) storeFileHandle(handle hostfs.FileHandle) uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.fileHandles == nil {
		h.fileHandles = map[uint64]hostfs.FileHandle{}
	}
	h.nextHandleID++
	if h.nextHandleID == 0 {
		h.nextHandleID++
	}
	h.fileHandles[h.nextHandleID] = handle
	return h.nextHandleID
}

func (h *DefaultHost) lookupFileHandle(handleID uint64) (hostfs.FileHandle, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.fileHandles == nil {
		return nil, fmt.Errorf("%w: %d", ErrCapabilityNotFound, handleID)
	}
	handle, ok := h.fileHandles[handleID]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrCapabilityNotFound, handleID)
	}
	return handle, nil
}

func (h *DefaultHost) releaseFileHandle(handleID uint64) (hostfs.FileHandle, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.fileHandles == nil {
		return nil, fmt.Errorf("%w: %d", ErrCapabilityNotFound, handleID)
	}
	handle, ok := h.fileHandles[handleID]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrCapabilityNotFound, handleID)
	}
	delete(h.fileHandles, handleID)
	return handle, nil
}

func (h *DefaultHost) callNetwork(ctx context.Context, req CallRequest) (CallResult, error) {
	if h == nil || h.Network == nil {
		return CallResult{}, fmt.Errorf("%w: %s", ErrProviderUnavailable, CapabilityNetwork)
	}

	switch req.Operation {
	case "network.http_fetch":
		parsed, err := decodeNetworkHTTPFetchRequest(req.Args)
		if err != nil {
			return CallResult{}, err
		}
		response, err := h.Network.HTTPFetch(ctx, parsed.toProviderRequest())
		if err != nil {
			return CallResult{}, err
		}
		return encodeNetworkHTTPFetchResponse(response), nil
	default:
		return CallResult{}, fmt.Errorf("unknown network operation: %w: %s", ErrCapabilityUnsupported, req.Operation)
	}
}

func readString(args map[string]any, key string) (string, error) {
	return readStringAny(args, key)
}

func readStringAny(args map[string]any, keys ...string) (string, error) {
	for _, key := range keys {
		value, ok := args[key]
		if !ok {
			continue
		}
		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("%w: %s must be a string", ErrInvalidCallArgs, key)
		}
		return text, nil
	}
	return "", fmt.Errorf("%w: missing %v", ErrInvalidCallArgs, keys)
}

func readOptionalStringAny(args map[string]any, keys ...string) (string, error) {
	for _, key := range keys {
		value, ok := args[key]
		if !ok || value == nil {
			continue
		}
		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("%w: %s must be a string", ErrInvalidCallArgs, key)
		}
		return text, nil
	}
	return "", nil
}

func readBytes(args map[string]any, key string) ([]byte, error) {
	value, ok := args[key]
	if !ok {
		return nil, fmt.Errorf("%w: missing %s", ErrInvalidCallArgs, key)
	}
	switch typed := value.(type) {
	case []byte:
		return typed, nil
	case string:
		return []byte(typed), nil
	default:
		return nil, fmt.Errorf("%w: %s must be []byte or string", ErrInvalidCallArgs, key)
	}
}

func readOptionalBytes(args map[string]any, key string) ([]byte, error) {
	value, ok := args[key]
	if !ok || value == nil {
		return nil, nil
	}
	switch typed := value.(type) {
	case []byte:
		return typed, nil
	case string:
		return []byte(typed), nil
	default:
		return nil, fmt.Errorf("%w: %s must be []byte or string", ErrInvalidCallArgs, key)
	}
}

func readBool(args map[string]any, key string) bool {
	value, ok := args[key]
	if !ok {
		return false
	}
	flag, ok := value.(bool)
	if !ok {
		return false
	}
	return flag
}

func readOptionalInt64Any(args map[string]any, keys ...string) (int64, error) {
	for _, key := range keys {
		value, ok := args[key]
		if !ok || value == nil {
			continue
		}
		return readInt64Value(key, value)
	}
	return 0, nil
}

func readInt64Value(key string, value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int8:
		return int64(typed), nil
	case int16:
		return int64(typed), nil
	case int32:
		return int64(typed), nil
	case int64:
		return typed, nil
	case uint:
		return int64(typed), nil
	case uint8:
		return int64(typed), nil
	case uint16:
		return int64(typed), nil
	case uint32:
		return int64(typed), nil
	case uint64:
		return int64(typed), nil
	case float32:
		return int64(typed), nil
	case float64:
		return int64(typed), nil
	default:
		return 0, fmt.Errorf("%w: %s must be a number", ErrInvalidCallArgs, key)
	}
}

func readRequiredUint64(args map[string]any, key string) (uint64, error) {
	value, ok := args[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("%w: missing %s", ErrInvalidCallArgs, key)
	}
	switch typed := value.(type) {
	case int:
		return uint64(typed), nil
	case int64:
		return uint64(typed), nil
	case uint64:
		return typed, nil
	case uint32:
		return uint64(typed), nil
	case float64:
		return uint64(typed), nil
	default:
		return 0, fmt.Errorf("%w: %s must be a handle id", ErrInvalidCallArgs, key)
	}
}

func readStringMap(args map[string]any, key string) (map[string]string, error) {
	value, ok := args[key]
	if !ok || value == nil {
		return nil, nil
	}
	switch typed := value.(type) {
	case map[string]string:
		result := make(map[string]string, len(typed))
		for k, v := range typed {
			result[k] = v
		}
		return result, nil
	case map[string]any:
		result := make(map[string]string, len(typed))
		for k, v := range typed {
			text, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("%w: %s[%s] must be a string", ErrInvalidCallArgs, key, k)
			}
			result[k] = text
		}
		return result, nil
	default:
		return nil, fmt.Errorf("%w: %s must be a string map", ErrInvalidCallArgs, key)
	}
}

func readStringSlice(values map[string]any, key string) []string {
	if values == nil {
		return nil
	}
	value, ok := values[key]
	if !ok {
		return nil
	}
	return toStringSlice(value)
}

func toStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		result := make([]string, len(typed))
		copy(result, typed)
		return result
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func cloneMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
