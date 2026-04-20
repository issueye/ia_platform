package ialang

import (
	"context"
	"errors"
	"fmt"

	hostapi "iavm/pkg/host/api"
)

const PlatformHTTPModuleName = "@platform/http"

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

	namespace := map[string]any{
		"client": clientNamespace,
	}
	module := cloneModuleMap(namespace)
	module["http"] = namespace
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
	requestOptions := cloneOptionalOptions(options)
	requestOptions["method"] = "GET"
	return b.Request(rawURL, requestOptions)
}

func (b *PlatformHTTPBridge) Post(rawURL string, options ...map[string]any) (map[string]any, error) {
	requestOptions := cloneOptionalOptions(options)
	requestOptions["method"] = "POST"
	return b.Request(rawURL, requestOptions)
}

func (b *PlatformHTTPBridge) call(operation string, args map[string]any) (map[string]any, error) {
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

func normalizeHTTPResult(result map[string]any) (map[string]any, error) {
	statusValue, ok := result["status"]
	if !ok {
		return nil, fmt.Errorf("%w: missing status", ErrInvalidHTTPResult)
	}
	status, err := bridgeInt(statusValue)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidHTTPResult, err)
	}

	headers, err := bridgeStringMap(result["headers"])
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidHTTPResult, err)
	}
	body, err := bridgeBytesToString(result["body"])
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidHTTPResult, err)
	}

	return map[string]any{
		"ok":         status >= 200 && status < 300,
		"status":     status,
		"statusCode": float64(status),
		"body":       body,
		"headers":    headers,
	}, nil
}

func cloneOptionalOptions(options []map[string]any) map[string]any {
	if len(options) == 0 || options[0] == nil {
		return map[string]any{}
	}
	return cloneModuleMap(options[0])
}

func readBridgeString(values map[string]any, key, fallback string) string {
	if values == nil {
		return fallback
	}
	value, ok := values[key]
	if !ok || value == nil {
		return fallback
	}
	text, ok := value.(string)
	if !ok {
		return fallback
	}
	if text == "" {
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
	mapped, err := bridgeStringMap(value)
	if err != nil {
		return nil
	}
	return mapped
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
	return values["timeoutMS"]
}

func bridgeStringMap(value any) (map[string]string, error) {
	if value == nil {
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
				return nil, fmt.Errorf("header %s is not string", k)
			}
			result[k] = text
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unexpected headers type %T", value)
	}
}

func bridgeBytesToString(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "", nil
	case string:
		return typed, nil
	case []byte:
		return string(typed), nil
	default:
		return "", fmt.Errorf("unexpected body type %T", value)
	}
}

func bridgeInt(value any) (int, error) {
	switch typed := value.(type) {
	case int:
		return typed, nil
	case int8:
		return int(typed), nil
	case int16:
		return int(typed), nil
	case int32:
		return int(typed), nil
	case int64:
		return int(typed), nil
	case uint:
		return int(typed), nil
	case uint8:
		return int(typed), nil
	case uint16:
		return int(typed), nil
	case uint32:
		return int(typed), nil
	case uint64:
		return int(typed), nil
	case float32:
		return int(typed), nil
	case float64:
		return int(typed), nil
	default:
		return 0, fmt.Errorf("unexpected status type %T", value)
	}
}
