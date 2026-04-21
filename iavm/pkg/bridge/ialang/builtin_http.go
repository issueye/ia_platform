package ialang

import (
	"context"
	"errors"

	hostapi "iacommon/pkg/host/api"
	moduleapi "iacommon/pkg/ialang/module"
)

const PlatformHTTPModuleName = moduleapi.PlatformHTTPModuleName

var ErrInvalidHTTPResult = errors.New("invalid http result")

type HTTPRequestFunc func(url string, options ...map[string]any) (map[string]any, error)
type HTTPGetFunc func(url string, options ...map[string]any) (map[string]any, error)
type HTTPPostFunc func(url string, options ...map[string]any) (map[string]any, error)

type PlatformHTTPBridge struct {
	Host         hostapi.Host
	CapabilityID string
}

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

func (b *PlatformHTTPBridge) Request(rawURL string, options ...map[string]any) (map[string]any, error) {
	req := map[string]any{"url": rawURL}
	if len(options) > 0 && options[0] != nil {
		for key, value := range options[0] {
			req[key] = value
		}
	}

	result, err := b.call("network.http_fetch", map[string]any{
		"url":        rawURL,
		"method":     readBridgeString(req, "method", "GET"),
		"headers":    readBridgeHeaders(req),
		"body":       readBridgeBody(req),
		"timeout_ms": readBridgeTimeout(req),
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

	normalized := map[string]any{
		"ok":      httpStatusOK(status),
		"status":  status,
		"headers": readBridgeHeaders(map[string]any{"headers": result["headers"]}),
		"body":    bodyText,
	}
	return normalized, nil
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
