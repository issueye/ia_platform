package ialang

import (
	"context"
	"errors"
	"fmt"
	"os"

	hostapi "iavm/pkg/host/api"
	hostfs "iavm/pkg/host/fs"
)

const PlatformFSModuleName = "@platform/fs"

var (
	ErrHostNotConfigured       = errors.New("host is not configured")
	ErrCapabilityNotConfigured = errors.New("capability is not configured")
	ErrInvalidFSResult         = errors.New("invalid fs result")
)

type FSReadFileFunc func(path string) (string, error)
type FSWriteFileFunc func(path, content string) (bool, error)
type FSAppendFileFunc func(path, content string) (bool, error)
type FSExistsFunc func(path string) (bool, error)
type FSMkdirFunc func(path string, recursive ...bool) (bool, error)
type FSReadDirFunc func(path string) ([]string, error)
type FSStatFunc func(path string) (map[string]any, error)
type FSRenameFunc func(oldPath, newPath string) (bool, error)
type FSRemoveFunc func(path string) (bool, error)
type FSRemoveAllFunc func(path string) (bool, error)
type FSCopyFunc func(src, dst string) (bool, error)

type PlatformFSBridge struct {
	Host         hostapi.Host
	CapabilityID string
}

func BuildPlatformFSModule() map[string]any {
	return buildPlatformFSModule(&PlatformFSBridge{})
}

func BuildPlatformFSModuleWithHost(host hostapi.Host) (map[string]any, error) {
	if host == nil {
		return nil, ErrHostNotConfigured
	}

	capability, err := host.AcquireCapability(context.Background(), hostapi.AcquireRequest{Kind: hostapi.CapabilityFS})
	if err != nil {
		return nil, err
	}
	return BuildPlatformFSModuleWithCapability(host, capability.ID), nil
}

func BuildPlatformFSModuleWithCapability(host hostapi.Host, capabilityID string) map[string]any {
	return buildPlatformFSModule(&PlatformFSBridge{
		Host:         host,
		CapabilityID: capabilityID,
	})
}

func buildPlatformFSModule(bridge *PlatformFSBridge) map[string]any {
	namespace := map[string]any{
		"readFile":   FSReadFileFunc(bridge.ReadFile),
		"writeFile":  FSWriteFileFunc(bridge.WriteFile),
		"appendFile": FSAppendFileFunc(bridge.AppendFile),
		"exists":     FSExistsFunc(bridge.Exists),
		"mkdir":      FSMkdirFunc(bridge.Mkdir),
		"readDir":    FSReadDirFunc(bridge.ReadDir),
		"stat":       FSStatFunc(bridge.Stat),
		"rename":     FSRenameFunc(bridge.Rename),
		"remove":     FSRemoveFunc(bridge.Remove),
		"removeAll":  FSRemoveAllFunc(bridge.RemoveAll),
		"copy":       FSCopyFunc(bridge.Copy),
	}

	module := cloneModuleMap(namespace)
	module["fs"] = namespace
	return module
}

func (b *PlatformFSBridge) ReadFile(path string) (string, error) {
	result, err := b.call("fs.read_file", map[string]any{"path": path})
	if err != nil {
		return "", err
	}

	data, ok := result["data"]
	if !ok {
		return "", fmt.Errorf("%w: missing data", ErrInvalidFSResult)
	}

	switch typed := data.(type) {
	case []byte:
		return string(typed), nil
	case string:
		return typed, nil
	default:
		return "", fmt.Errorf("%w: unexpected readFile payload %T", ErrInvalidFSResult, data)
	}
}

func (b *PlatformFSBridge) WriteFile(path, content string) (bool, error) {
	_, err := b.call("fs.write_file", map[string]any{
		"path":   path,
		"data":   content,
		"create": true,
		"trunc":  true,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) AppendFile(path, content string) (bool, error) {
	_, err := b.call("fs.append_file", map[string]any{
		"path": path,
		"data": content,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) Exists(path string) (bool, error) {
	_, err := b.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (b *PlatformFSBridge) Mkdir(path string, recursive ...bool) (bool, error) {
	_, err := b.call("fs.mkdir", map[string]any{
		"path":      path,
		"recursive": len(recursive) > 0 && recursive[0],
	})
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

	entries, ok := result["entries"]
	if !ok {
		return nil, fmt.Errorf("%w: missing entries", ErrInvalidFSResult)
	}

	switch typed := entries.(type) {
	case []hostfs.DirEntry:
		out := make([]string, 0, len(typed))
		for _, entry := range typed {
			out = append(out, entry.Name)
		}
		return out, nil
	case []any:
		out := make([]string, 0, len(typed))
		for _, entry := range typed {
			switch item := entry.(type) {
			case hostfs.DirEntry:
				out = append(out, item.Name)
			case string:
				out = append(out, item)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%w: unexpected readDir payload %T", ErrInvalidFSResult, entries)
	}
}

func (b *PlatformFSBridge) Stat(path string) (map[string]any, error) {
	result, err := b.call("fs.stat", map[string]any{"path": path})
	if err != nil {
		return nil, err
	}

	info, ok := result["info"]
	if !ok {
		return nil, fmt.Errorf("%w: missing info", ErrInvalidFSResult)
	}

	switch typed := info.(type) {
	case hostfs.FileInfo:
		return fileInfoToMap(typed), nil
	case map[string]any:
		return typed, nil
	default:
		return nil, fmt.Errorf("%w: unexpected stat payload %T", ErrInvalidFSResult, info)
	}
}

func (b *PlatformFSBridge) Rename(oldPath, newPath string) (bool, error) {
	_, err := b.call("fs.rename", map[string]any{
		"old_path": oldPath,
		"new_path": newPath,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) Remove(path string) (bool, error) {
	_, err := b.call("fs.remove", map[string]any{"path": path})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) RemoveAll(path string) (bool, error) {
	_, err := b.call("fs.remove", map[string]any{
		"path":      path,
		"recursive": true,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *PlatformFSBridge) Copy(src, dst string) (bool, error) {
	content, err := b.ReadFile(src)
	if err != nil {
		return false, err
	}
	return b.WriteFile(dst, content)
}

func (b *PlatformFSBridge) call(operation string, args map[string]any) (map[string]any, error) {
	if b == nil || b.Host == nil {
		return nil, ErrHostNotConfigured
	}
	if b.CapabilityID == "" {
		return nil, ErrCapabilityNotConfigured
	}

	result, err := b.Host.Call(context.Background(), hostapi.CallRequest{
		CapabilityID: b.CapabilityID,
		Operation:    operation,
		Args:         args,
	})
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}

func fileInfoToMap(info hostfs.FileInfo) map[string]any {
	return map[string]any{
		"name":        info.Name,
		"isDir":       info.IsDir,
		"size":        float64(info.Size),
		"mode":        info.Mode,
		"modTimeUnix": float64(info.ModUnix),
	}
}

func cloneModuleMap(values map[string]any) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
