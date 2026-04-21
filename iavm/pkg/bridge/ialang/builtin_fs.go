package ialang

import (
	"context"
	"errors"
	"os"

	hostapi "iacommon/pkg/host/api"
	hostfs "iacommon/pkg/host/fs"
	moduleapi "iacommon/pkg/ialang/module"
)

const PlatformFSModuleName = moduleapi.PlatformFSModuleName

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
	readFileFn := FSReadFileFunc(bridge.ReadFile)
	writeFileFn := FSWriteFileFunc(bridge.WriteFile)
	appendFileFn := FSAppendFileFunc(bridge.AppendFile)
	existsFn := FSExistsFunc(bridge.Exists)
	mkdirFn := FSMkdirFunc(bridge.Mkdir)
	readDirFn := FSReadDirFunc(bridge.ReadDir)
	statFn := FSStatFunc(bridge.Stat)
	renameFn := FSRenameFunc(bridge.Rename)
	removeFn := FSRemoveFunc(bridge.Remove)
	removeAllFn := FSRemoveAllFunc(bridge.RemoveAll)
	copyFn := FSCopyFunc(bridge.Copy)

	fsNamespace := map[string]any{
		"readFile":   readFileFn,
		"writeFile":  writeFileFn,
		"appendFile": appendFileFn,
		"exists":     existsFn,
		"mkdir":      mkdirFn,
		"readDir":    readDirFn,
		"stat":       statFn,
		"rename":     renameFn,
		"remove":     removeFn,
		"removeAll":  removeAllFn,
		"copy":       copyFn,
	}

	module := cloneModuleMap(fsNamespace)
	module["fs"] = fsNamespace
	return module
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
	if b.Host == nil {
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
	if result.Value == nil {
		return map[string]any{}, nil
	}
	return result.Value, nil
}

func cloneModuleMap(exports map[string]any) map[string]any {
	module := make(map[string]any, len(exports))
	for key, value := range exports {
		module[key] = value
	}
	return module
}
