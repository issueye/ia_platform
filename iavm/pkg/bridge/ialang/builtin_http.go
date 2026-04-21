package ialang

import (
	"context"

	hostapi "iacommon/pkg/host/api"
	moduleapi "iacommon/pkg/ialang/module"
)

const PlatformHTTPModuleName = moduleapi.PlatformHTTPModuleName

var ErrInvalidHTTPResult = moduleapi.ErrInvalidHTTPResult

type HTTPRequestFunc func(url string, options ...map[string]any) (map[string]any, error)
type HTTPGetFunc func(url string, options ...map[string]any) (map[string]any, error)
type HTTPPostFunc func(url string, options ...map[string]any) (map[string]any, error)

type PlatformHTTPBridge = moduleapi.PlatformHTTPBridge

func BuildPlatformHTTPModule() map[string]any {
	return buildPlatformHTTPModule(&PlatformHTTPBridge{})
}

func BuildPlatformHTTPModuleWithHost(host hostapi.Host) (map[string]any, error) {
	if host == nil {
		return nil, ErrHostNotConfigured
	}

	capability, err := host.AcquireCapability(context.Background(), hostapi.AcquireRequest{Kind: hostapi.CapabilityNetwork})
	if err != nil {
		return nil, err
	}
	return BuildPlatformHTTPModuleWithCapability(host, capability.ID), nil
}

func BuildPlatformHTTPModuleWithCapability(host hostapi.Host, capabilityID string) map[string]any {
	return buildPlatformHTTPModule(&PlatformHTTPBridge{
		Host:         host,
		CapabilityID: capabilityID,
	})
}

func buildPlatformHTTPModule(bridge *PlatformHTTPBridge) map[string]any {
	clientNamespace := map[string]any{
		"request": HTTPRequestFunc(bridge.Request),
		"get":     HTTPGetFunc(bridge.Get),
		"post":    HTTPPostFunc(bridge.Post),
	}

	httpNamespace := map[string]any{
		"client": clientNamespace,
	}
	module := cloneModuleMap(clientNamespace)
	module["http"] = httpNamespace
	return module
}
