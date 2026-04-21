package ialang

import (
	"context"

	hostapi "iacommon/pkg/host/api"
	moduleapi "iacommon/pkg/ialang/module"
)

const PlatformFSModuleName = moduleapi.PlatformFSModuleName

var (
	ErrHostNotConfigured       = moduleapi.ErrHostNotConfigured
	ErrCapabilityNotConfigured = moduleapi.ErrCapabilityNotConfigured
	ErrInvalidFSResult         = moduleapi.ErrInvalidFSResult
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

type PlatformFSBridge = moduleapi.PlatformFSBridge

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

func cloneModuleMap(exports map[string]any) map[string]any {
	module := make(map[string]any, len(exports))
	for key, value := range exports {
		module[key] = value
	}
	return module
}
