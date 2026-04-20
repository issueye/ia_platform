package builtin

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	rtvm "ialang/pkg/lang/runtime/vm"
)

type httpRequestConfig struct {
	URL         string
	Method      string
	Body        string
	ContentType string
	Headers     Object
	Timeout     time.Duration
	ChunkSize   int
	Proxy       string
}

type httpServerConfig struct {
	Addr        string
	StatusCode  int
	Body        string
	ContentType string
	Headers     Object
}

type httpProxyServerConfig struct {
	Addr              string
	Target            string
	StripPrefix       string
	PreserveHost      bool
	Headers           Object
	RequestMutations  httpRequestMutationResolver
	ResponseMutations httpResponseMutationResolver
}

type httpForwardServerConfig struct {
	Addr              string
	Target            string
	KeepPath          bool
	Path              string
	PreserveHost      bool
	Headers           Object
	Timeout           time.Duration
	RequestMutations  httpRequestMutationResolver
	ResponseMutations httpResponseMutationResolver
}

type httpBodyMutation struct {
	Enabled bool
	Raw     []byte
}

type httpRequestMutations struct {
	Method        string
	Path          string
	AppendPath    string
	SetQuery      map[string]string
	RemoveQuery   []string
	SetHeaders    map[string]string
	RemoveHeaders []string
	Body          httpBodyMutation
}

type httpResponseMutations struct {
	StatusCode    int
	HasStatusCode bool
	SetHeaders    map[string]string
	RemoveHeaders []string
	Body          httpBodyMutation
}

type httpRequestMutationResolver struct {
	Static    httpRequestMutations
	HasStatic bool
	Dynamic   Value
}

type httpResponseMutationResolver struct {
	Static    httpResponseMutations
	HasStatic bool
	Dynamic   Value
}

func newHTTPModule(asyncRuntime AsyncRuntime) Object {
	client := &http.Client{Timeout: 15 * time.Second}

	requestFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPRequestArgs("http.client.request", args, http.MethodGet)
		if err != nil {
			return nil, err
		}
		return doHTTPRequest(client, cfg)
	})

	getFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPRequestArgs("http.client.get", args, http.MethodGet)
		if err != nil {
			return nil, err
		}
		if cfg.Method != http.MethodGet {
			return nil, fmt.Errorf("http.client.get options.method must be GET, got %s", cfg.Method)
		}
		return doHTTPRequest(client, cfg)
	})

	postFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPRequestArgs("http.client.post", args, http.MethodPost)
		if err != nil {
			return nil, err
		}
		if cfg.Method != http.MethodPost {
			return nil, fmt.Errorf("http.client.post options.method must be POST, got %s", cfg.Method)
		}
		return doHTTPRequest(client, cfg)
	})

	requestAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return requestFn(args)
		}), nil
	})
	streamFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPRequestArgs("http.client.stream", args, http.MethodGet)
		if err != nil {
			return nil, err
		}
		return doHTTPStreamRequest(client, cfg, asyncRuntime)
	})
	streamAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return streamFn(args)
		}), nil
	})
	getAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return getFn(args)
		}), nil
	})
	postAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return postFn(args)
		}), nil
	})

	serverServeFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPServerArgs("http.server.serve", args)
		if err != nil {
			return nil, err
		}
		return startStaticHTTPServer(cfg)
	})
	serverServeAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return serverServeFn(args)
		}), nil
	})
	serverProxyFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPProxyServerArgs("http.server.proxy", args)
		if err != nil {
			return nil, err
		}
		return startReverseProxyHTTPServer(cfg)
	})
	serverProxyAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return serverProxyFn(args)
		}), nil
	})
	serverForwardFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPForwardServerArgs("http.server.forward", args)
		if err != nil {
			return nil, err
		}
		return startForwardHTTPServer(cfg)
	})
	serverForwardAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return serverForwardFn(args)
		}), nil
	})

	clientNamespace := Object{
		"request":      requestFn,
		"stream":       streamFn,
		"get":          getFn,
		"post":         postFn,
		"requestAsync": requestAsyncFn,
		"streamAsync":  streamAsyncFn,
		"getAsync":     getAsyncFn,
		"postAsync":    postAsyncFn,
	}
	serverNamespace := Object{
		"serve":        serverServeFn,
		"serveAsync":   serverServeAsyncFn,
		"proxy":        serverProxyFn,
		"proxyAsync":   serverProxyAsyncFn,
		"forward":      serverForwardFn,
		"forwardAsync": serverForwardAsyncFn,
	}

	namespace := Object{
		"client": clientNamespace,
		"server": serverNamespace,
	}
	module := cloneObject(namespace)
	module["http"] = namespace
	return module
}

func parseHTTPRequestArgs(fn string, args []Value, defaultMethod string) (httpRequestConfig, error) {
	if len(args) < 1 || len(args) > 2 {
		return httpRequestConfig{}, fmt.Errorf("%s expects 1-2 args: url, [options]", fn)
	}
	url, err := asStringArg(fn, args, 0)
	if err != nil {
		return httpRequestConfig{}, err
	}
	cfg := httpRequestConfig{
		URL:         url,
		Method:      defaultMethod,
		Body:        "",
		ContentType: "text/plain; charset=utf-8",
		Headers:     Object{},
		Timeout:     15 * time.Second,
		ChunkSize:   4096,
		Proxy:       "",
	}
	if len(args) == 1 || args[1] == nil {
		return cfg, nil
	}
	options, ok := args[1].(Object)
	if !ok {
		return httpRequestConfig{}, fmt.Errorf("%s arg[1] expects object options, got %T", fn, args[1])
	}
	if v, ok := options["method"]; ok && v != nil {
		method, err := asStringValue("http options.method", v)
		if err != nil {
			return httpRequestConfig{}, err
		}
		cfg.Method = strings.ToUpper(method)
	}
	if v, ok := options["body"]; ok {
		body, err := asStringValue("http options.body", v)
		if err != nil {
			return httpRequestConfig{}, err
		}
		cfg.Body = body
	}
	if v, ok := options["contentType"]; ok && v != nil {
		contentType, err := asStringValue("http options.contentType", v)
		if err != nil {
			return httpRequestConfig{}, err
		}
		cfg.ContentType = contentType
	}
	if v, ok := options["headers"]; ok && v != nil {
		headers, ok := v.(Object)
		if !ok {
			return httpRequestConfig{}, fmt.Errorf("http options.headers expects object, got %T", v)
		}
		cfg.Headers = cloneObject(headers)
	}
	if v, ok := options["timeoutMs"]; ok && v != nil {
		timeoutMs, err := asIntValue("http options.timeoutMs", v)
		if err != nil {
			return httpRequestConfig{}, err
		}
		if timeoutMs <= 0 {
			return httpRequestConfig{}, fmt.Errorf("http options.timeoutMs expects positive integer, got %d", timeoutMs)
		}
		cfg.Timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	if v, ok := options["chunkSize"]; ok && v != nil {
		chunkSize, err := asIntValue("http options.chunkSize", v)
		if err != nil {
			return httpRequestConfig{}, err
		}
		if chunkSize <= 0 {
			return httpRequestConfig{}, fmt.Errorf("http options.chunkSize expects positive integer, got %d", chunkSize)
		}
		cfg.ChunkSize = chunkSize
	}
	if v, ok := options["proxy"]; ok && v != nil {
		proxy, err := asStringValue("http options.proxy", v)
		if err != nil {
			return httpRequestConfig{}, err
		}
		cfg.Proxy = strings.TrimSpace(proxy)
	}
	return cfg, nil
}

func parseHTTPServerArgs(fn string, args []Value) (httpServerConfig, error) {
	if len(args) > 1 {
		return httpServerConfig{}, fmt.Errorf("%s expects 0-1 args: [options]", fn)
	}
	cfg := httpServerConfig{
		Addr:        "127.0.0.1:0",
		StatusCode:  http.StatusOK,
		Body:        "ok",
		ContentType: "text/plain; charset=utf-8",
		Headers:     Object{},
	}
	if len(args) == 0 || args[0] == nil {
		return cfg, nil
	}
	options, ok := args[0].(Object)
	if !ok {
		return httpServerConfig{}, fmt.Errorf("%s arg[0] expects object options, got %T", fn, args[0])
	}
	if v, ok := options["addr"]; ok && v != nil {
		addr, err := asStringValue("http.server options.addr", v)
		if err != nil {
			return httpServerConfig{}, err
		}
		cfg.Addr = addr
	}
	if v, ok := options["statusCode"]; ok && v != nil {
		code, err := asIntValue("http.server options.statusCode", v)
		if err != nil {
			return httpServerConfig{}, err
		}
		if code < 100 || code > 599 {
			return httpServerConfig{}, fmt.Errorf("http.server options.statusCode expects 100-599, got %d", code)
		}
		cfg.StatusCode = code
	}
	if v, ok := options["body"]; ok {
		body, err := asStringValue("http.server options.body", v)
		if err != nil {
			return httpServerConfig{}, err
		}
		cfg.Body = body
	}
	if v, ok := options["contentType"]; ok && v != nil {
		contentType, err := asStringValue("http.server options.contentType", v)
		if err != nil {
			return httpServerConfig{}, err
		}
		cfg.ContentType = contentType
	}
	if v, ok := options["headers"]; ok && v != nil {
		headers, ok := v.(Object)
		if !ok {
			return httpServerConfig{}, fmt.Errorf("http.server options.headers expects object, got %T", v)
		}
		cfg.Headers = cloneObject(headers)
	}
	return cfg, nil
}

func parseHTTPProxyServerArgs(fn string, args []Value) (httpProxyServerConfig, error) {
	if len(args) > 1 {
		return httpProxyServerConfig{}, fmt.Errorf("%s expects 0-1 args: [options]", fn)
	}
	cfg := httpProxyServerConfig{
		Addr:              "127.0.0.1:0",
		Target:            "",
		StripPrefix:       "",
		PreserveHost:      false,
		Headers:           Object{},
		RequestMutations:  httpRequestMutationResolver{},
		ResponseMutations: httpResponseMutationResolver{},
	}
	if len(args) == 0 || args[0] == nil {
		return httpProxyServerConfig{}, fmt.Errorf("%s requires options.target", fn)
	}
	options, ok := args[0].(Object)
	if !ok {
		return httpProxyServerConfig{}, fmt.Errorf("%s arg[0] expects object options, got %T", fn, args[0])
	}
	if v, ok := options["addr"]; ok && v != nil {
		addr, err := asStringValue("http.server proxy options.addr", v)
		if err != nil {
			return httpProxyServerConfig{}, err
		}
		cfg.Addr = addr
	}
	if v, ok := options["target"]; ok && v != nil {
		target, err := asStringValue("http.server proxy options.target", v)
		if err != nil {
			return httpProxyServerConfig{}, err
		}
		cfg.Target = strings.TrimSpace(target)
	}
	if cfg.Target == "" {
		return httpProxyServerConfig{}, fmt.Errorf("%s options.target is required", fn)
	}
	if v, ok := options["stripPrefix"]; ok && v != nil {
		prefix, err := asStringValue("http.server proxy options.stripPrefix", v)
		if err != nil {
			return httpProxyServerConfig{}, err
		}
		cfg.StripPrefix = prefix
	}
	if v, ok := options["preserveHost"]; ok && v != nil {
		preserveHost, ok := v.(bool)
		if !ok {
			return httpProxyServerConfig{}, fmt.Errorf("http.server proxy options.preserveHost expects bool, got %T", v)
		}
		cfg.PreserveHost = preserveHost
	}
	if v, ok := options["headers"]; ok && v != nil {
		headers, ok := v.(Object)
		if !ok {
			return httpProxyServerConfig{}, fmt.Errorf("http.server proxy options.headers expects object, got %T", v)
		}
		cfg.Headers = cloneObject(headers)
	}
	if v, ok := options["requestMutations"]; ok && v != nil {
		mutations, err := parseHTTPRequestMutationResolver("http.server proxy options.requestMutations", v)
		if err != nil {
			return httpProxyServerConfig{}, err
		}
		cfg.RequestMutations = mutations
	}
	if v, ok := options["responseMutations"]; ok && v != nil {
		mutations, err := parseHTTPResponseMutationResolver("http.server proxy options.responseMutations", v)
		if err != nil {
			return httpProxyServerConfig{}, err
		}
		cfg.ResponseMutations = mutations
	}
	return cfg, nil
}

func parseHTTPForwardServerArgs(fn string, args []Value) (httpForwardServerConfig, error) {
	if len(args) > 1 {
		return httpForwardServerConfig{}, fmt.Errorf("%s expects 0-1 args: [options]", fn)
	}
	cfg := httpForwardServerConfig{
		Addr:              "127.0.0.1:0",
		Target:            "",
		KeepPath:          true,
		Path:              "",
		PreserveHost:      false,
		Headers:           Object{},
		Timeout:           15 * time.Second,
		RequestMutations:  httpRequestMutationResolver{},
		ResponseMutations: httpResponseMutationResolver{},
	}
	if len(args) == 0 || args[0] == nil {
		return httpForwardServerConfig{}, fmt.Errorf("%s requires options.target", fn)
	}
	options, ok := args[0].(Object)
	if !ok {
		return httpForwardServerConfig{}, fmt.Errorf("%s arg[0] expects object options, got %T", fn, args[0])
	}
	if v, ok := options["addr"]; ok && v != nil {
		addr, err := asStringValue("http.server forward options.addr", v)
		if err != nil {
			return httpForwardServerConfig{}, err
		}
		cfg.Addr = addr
	}
	if v, ok := options["target"]; ok && v != nil {
		target, err := asStringValue("http.server forward options.target", v)
		if err != nil {
			return httpForwardServerConfig{}, err
		}
		cfg.Target = strings.TrimSpace(target)
	}
	if cfg.Target == "" {
		return httpForwardServerConfig{}, fmt.Errorf("%s options.target is required", fn)
	}
	if v, ok := options["keepPath"]; ok && v != nil {
		keepPath, ok := v.(bool)
		if !ok {
			return httpForwardServerConfig{}, fmt.Errorf("http.server forward options.keepPath expects bool, got %T", v)
		}
		cfg.KeepPath = keepPath
	}
	if v, ok := options["path"]; ok && v != nil {
		path, err := asStringValue("http.server forward options.path", v)
		if err != nil {
			return httpForwardServerConfig{}, err
		}
		cfg.Path = path
	}
	if v, ok := options["preserveHost"]; ok && v != nil {
		preserveHost, ok := v.(bool)
		if !ok {
			return httpForwardServerConfig{}, fmt.Errorf("http.server forward options.preserveHost expects bool, got %T", v)
		}
		cfg.PreserveHost = preserveHost
	}
	if v, ok := options["headers"]; ok && v != nil {
		headers, ok := v.(Object)
		if !ok {
			return httpForwardServerConfig{}, fmt.Errorf("http.server forward options.headers expects object, got %T", v)
		}
		cfg.Headers = cloneObject(headers)
	}
	if v, ok := options["timeoutMs"]; ok && v != nil {
		timeoutMs, err := asIntValue("http.server forward options.timeoutMs", v)
		if err != nil {
			return httpForwardServerConfig{}, err
		}
		if timeoutMs <= 0 {
			return httpForwardServerConfig{}, fmt.Errorf("http.server forward options.timeoutMs expects positive integer, got %d", timeoutMs)
		}
		cfg.Timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	if v, ok := options["requestMutations"]; ok && v != nil {
		mutations, err := parseHTTPRequestMutationResolver("http.server forward options.requestMutations", v)
		if err != nil {
			return httpForwardServerConfig{}, err
		}
		cfg.RequestMutations = mutations
	}
	if v, ok := options["responseMutations"]; ok && v != nil {
		mutations, err := parseHTTPResponseMutationResolver("http.server forward options.responseMutations", v)
		if err != nil {
			return httpForwardServerConfig{}, err
		}
		cfg.ResponseMutations = mutations
	}
	return cfg, nil
}

func parseHTTPRequestMutations(label string, raw Value) (httpRequestMutations, error) {
	obj, ok := raw.(Object)
	if !ok {
		return httpRequestMutations{}, fmt.Errorf("%s expects object, got %T", label, raw)
	}

	out := httpRequestMutations{}
	if v, ok := obj["method"]; ok && v != nil {
		method, err := asStringValue(label+".method", v)
		if err != nil {
			return httpRequestMutations{}, err
		}
		out.Method = strings.ToUpper(strings.TrimSpace(method))
	}
	if v, ok := obj["path"]; ok && v != nil {
		path, err := asStringValue(label+".path", v)
		if err != nil {
			return httpRequestMutations{}, err
		}
		out.Path = path
	}
	if v, ok := obj["appendPath"]; ok && v != nil {
		appendPath, err := asStringValue(label+".appendPath", v)
		if err != nil {
			return httpRequestMutations{}, err
		}
		out.AppendPath = appendPath
	}
	if v, ok := obj["setQuery"]; ok && v != nil {
		setQuery, err := parseHTTPStringMap(label+".setQuery", v)
		if err != nil {
			return httpRequestMutations{}, err
		}
		out.SetQuery = setQuery
	}
	if v, ok := obj["removeQuery"]; ok && v != nil {
		removeQuery, err := parseHTTPStringArray(label+".removeQuery", v)
		if err != nil {
			return httpRequestMutations{}, err
		}
		out.RemoveQuery = removeQuery
	}
	if v, ok := obj["setHeaders"]; ok && v != nil {
		setHeaders, err := parseHTTPStringMap(label+".setHeaders", v)
		if err != nil {
			return httpRequestMutations{}, err
		}
		out.SetHeaders = setHeaders
	}
	if v, ok := obj["removeHeaders"]; ok && v != nil {
		removeHeaders, err := parseHTTPStringArray(label+".removeHeaders", v)
		if err != nil {
			return httpRequestMutations{}, err
		}
		out.RemoveHeaders = removeHeaders
	}
	bodyMutation, err := parseHTTPBodyMutation(label, obj)
	if err != nil {
		return httpRequestMutations{}, err
	}
	out.Body = bodyMutation

	return out, nil
}

func parseHTTPResponseMutations(label string, raw Value) (httpResponseMutations, error) {
	obj, ok := raw.(Object)
	if !ok {
		return httpResponseMutations{}, fmt.Errorf("%s expects object, got %T", label, raw)
	}

	out := httpResponseMutations{}
	if v, ok := obj["statusCode"]; ok && v != nil {
		code, err := asIntValue(label+".statusCode", v)
		if err != nil {
			return httpResponseMutations{}, err
		}
		if code < 100 || code > 599 {
			return httpResponseMutations{}, fmt.Errorf("%s.statusCode expects 100-599, got %d", label, code)
		}
		out.StatusCode = code
		out.HasStatusCode = true
	}
	if v, ok := obj["setHeaders"]; ok && v != nil {
		setHeaders, err := parseHTTPStringMap(label+".setHeaders", v)
		if err != nil {
			return httpResponseMutations{}, err
		}
		out.SetHeaders = setHeaders
	}
	if v, ok := obj["removeHeaders"]; ok && v != nil {
		removeHeaders, err := parseHTTPStringArray(label+".removeHeaders", v)
		if err != nil {
			return httpResponseMutations{}, err
		}
		out.RemoveHeaders = removeHeaders
	}
	bodyMutation, err := parseHTTPBodyMutation(label, obj)
	if err != nil {
		return httpResponseMutations{}, err
	}
	out.Body = bodyMutation

	return out, nil
}

func parseHTTPBodyMutation(label string, obj Object) (httpBodyMutation, error) {
	bodyRaw, hasBody := obj["body"]
	bodyBase64Raw, hasBodyBase64 := obj["bodyBase64"]
	if hasBody && hasBodyBase64 {
		return httpBodyMutation{}, fmt.Errorf("%s.body and %s.bodyBase64 are mutually exclusive", label, label)
	}
	if hasBody {
		body, err := asStringValue(label+".body", bodyRaw)
		if err != nil {
			return httpBodyMutation{}, err
		}
		return httpBodyMutation{Enabled: true, Raw: []byte(body)}, nil
	}
	if hasBodyBase64 {
		bodyBase64, err := asStringValue(label+".bodyBase64", bodyBase64Raw)
		if err != nil {
			return httpBodyMutation{}, err
		}
		decoded, err := base64.StdEncoding.DecodeString(bodyBase64)
		if err != nil {
			return httpBodyMutation{}, fmt.Errorf("%s.bodyBase64 decode error: %w", label, err)
		}
		return httpBodyMutation{Enabled: true, Raw: decoded}, nil
	}
	return httpBodyMutation{}, nil
}

func parseHTTPStringMap(label string, raw Value) (map[string]string, error) {
	obj, ok := raw.(Object)
	if !ok {
		return nil, fmt.Errorf("%s expects object, got %T", label, raw)
	}
	out := make(map[string]string, len(obj))
	for k, v := range obj {
		s, err := asStringValue(label+"["+k+"]", v)
		if err != nil {
			return nil, err
		}
		out[k] = s
	}
	return out, nil
}

func parseHTTPStringArray(label string, raw Value) ([]string, error) {
	arr, ok := raw.(Array)
	if !ok {
		return nil, fmt.Errorf("%s expects array, got %T", label, raw)
	}
	out := make([]string, 0, len(arr))
	for i, v := range arr {
		s, err := asStringValue(fmt.Sprintf("%s[%d]", label, i), v)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func parseHTTPRequestMutationResolver(label string, raw Value) (httpRequestMutationResolver, error) {
	switch raw.(type) {
	case NativeFunction, *UserFunction:
		return httpRequestMutationResolver{Dynamic: raw}, nil
	case Object:
		mutations, err := parseHTTPRequestMutations(label, raw)
		if err != nil {
			return httpRequestMutationResolver{}, err
		}
		return httpRequestMutationResolver{
			Static:    mutations,
			HasStatic: true,
		}, nil
	default:
		return httpRequestMutationResolver{}, fmt.Errorf("%s expects object or function, got %T", label, raw)
	}
}

func parseHTTPResponseMutationResolver(label string, raw Value) (httpResponseMutationResolver, error) {
	switch raw.(type) {
	case NativeFunction, *UserFunction:
		return httpResponseMutationResolver{Dynamic: raw}, nil
	case Object:
		mutations, err := parseHTTPResponseMutations(label, raw)
		if err != nil {
			return httpResponseMutationResolver{}, err
		}
		return httpResponseMutationResolver{
			Static:    mutations,
			HasStatic: true,
		}, nil
	default:
		return httpResponseMutationResolver{}, fmt.Errorf("%s expects object or function, got %T", label, raw)
	}
}

func resolveRequestMutations(resolver httpRequestMutationResolver, req *http.Request) (httpRequestMutations, error) {
	if resolver.Dynamic == nil {
		if resolver.HasStatic {
			return resolver.Static, nil
		}
		return httpRequestMutations{}, nil
	}

	body, err := snapshotRequestBody(req)
	if err != nil {
		return httpRequestMutations{}, err
	}
	ctx := requestToMutationContext(req, body)
	obj, err := callMutationCallback(resolver.Dynamic, []Value{ctx}, "requestMutations")
	if err != nil {
		return httpRequestMutations{}, err
	}
	dynamicMutations, err := parseHTTPRequestMutations("http.server requestMutations callback result", obj)
	if err != nil {
		return httpRequestMutations{}, err
	}
	if resolver.HasStatic {
		return mergeHTTPRequestMutations(resolver.Static, dynamicMutations), nil
	}
	return dynamicMutations, nil
}

func resolveResponseMutations(
	resolver httpResponseMutationResolver,
	resp *http.Response,
	requestContext Object,
) (httpResponseMutations, error) {
	if resolver.Dynamic == nil {
		if resolver.HasStatic {
			return resolver.Static, nil
		}
		return httpResponseMutations{}, nil
	}

	body, err := snapshotResponseBody(resp)
	if err != nil {
		return httpResponseMutations{}, err
	}
	responseContext := responseToMutationContext(resp, body)
	if requestContext == nil {
		requestContext = Object{}
	}
	obj, err := callMutationCallback(resolver.Dynamic, []Value{responseContext, requestContext}, "responseMutations")
	if err != nil {
		return httpResponseMutations{}, err
	}
	dynamicMutations, err := parseHTTPResponseMutations("http.server responseMutations callback result", obj)
	if err != nil {
		return httpResponseMutations{}, err
	}
	if resolver.HasStatic {
		return mergeHTTPResponseMutations(resolver.Static, dynamicMutations), nil
	}
	return dynamicMutations, nil
}

func buildRequestContextForResponseMutation(req *http.Request) (Object, error) {
	body, err := snapshotRequestBodyWithGetBody(req)
	if err != nil {
		return nil, err
	}
	return requestToMutationContext(req, body), nil
}

func snapshotRequestBody(req *http.Request) ([]byte, error) {
	if req == nil || req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	resetRequestBody(req, raw)
	return raw, nil
}

func snapshotRequestBodyWithGetBody(req *http.Request) ([]byte, error) {
	if req == nil || req.GetBody == nil {
		return nil, nil
	}
	rc, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	raw, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func snapshotResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil || resp.Body == http.NoBody {
		return nil, nil
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resetResponseBody(resp, raw)
	return raw, nil
}

func requestToMutationContext(req *http.Request, body []byte) Object {
	if req == nil {
		return Object{}
	}
	query := Object{}
	if req.URL != nil {
		query = urlValuesToObject(req.URL.Query())
	}
	path := ""
	rawQuery := ""
	if req.URL != nil {
		path = req.URL.Path
		rawQuery = req.URL.RawQuery
	}
	headerLookupReq := req
	headerFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("request context header(name) expects 1 arg, got %d", len(args))
		}
		name, err := asStringValue("request context header(name)", args[0])
		if err != nil {
			return nil, err
		}
		if headerLookupReq == nil || headerLookupReq.Header == nil {
			return "", nil
		}
		return headerLookupReq.Header.Get(name), nil
	})
	return Object{
		"method":     req.Method,
		"path":       path,
		"rawQuery":   rawQuery,
		"query":      query,
		"headers":    headersToObject(req.Header),
		"body":       string(body),
		"bodyBase64": base64.StdEncoding.EncodeToString(body),
		"header":     headerFn,
		"host":       req.Host,
	}
}

func responseToMutationContext(resp *http.Response, body []byte) Object {
	if resp == nil {
		return Object{}
	}
	return Object{
		"status":     resp.Status,
		"statusCode": float64(resp.StatusCode),
		"headers":    headersToObject(resp.Header),
		"body":       string(body),
		"bodyBase64": base64.StdEncoding.EncodeToString(body),
	}
}

func callMutationCallback(fn Value, args []Value, label string) (Object, error) {
	var (
		result Value
		err    error
	)
	switch cb := fn.(type) {
	case NativeFunction:
		result, err = cb(args)
	case *UserFunction:
		result, err = rtvm.CallUserFunctionSync(cb, args)
	default:
		return nil, fmt.Errorf("http.server %s callback expects function, got %T", label, fn)
	}
	if err != nil {
		return nil, fmt.Errorf("http.server %s callback failed: %w", label, err)
	}
	if result == nil {
		return Object{}, nil
	}
	obj, ok := result.(Object)
	if !ok {
		return nil, fmt.Errorf("http.server %s callback must return object, got %T", label, result)
	}
	return obj, nil
}

func urlValuesToObject(values url.Values) Object {
	out := Object{}
	for k, items := range values {
		out[k] = strings.Join(items, ",")
	}
	return out
}

func mergeHTTPRequestMutations(base httpRequestMutations, override httpRequestMutations) httpRequestMutations {
	out := base
	if strings.TrimSpace(override.Method) != "" {
		out.Method = strings.ToUpper(strings.TrimSpace(override.Method))
	}
	if override.Path != "" {
		out.Path = override.Path
	}
	if override.AppendPath != "" {
		out.AppendPath = override.AppendPath
	}
	if len(override.SetQuery) > 0 {
		out.SetQuery = mergeStringMap(out.SetQuery, override.SetQuery)
	}
	if len(override.RemoveQuery) > 0 {
		out.RemoveQuery = append(append([]string(nil), out.RemoveQuery...), override.RemoveQuery...)
	}
	if len(override.SetHeaders) > 0 {
		out.SetHeaders = mergeStringMap(out.SetHeaders, override.SetHeaders)
	}
	if len(override.RemoveHeaders) > 0 {
		out.RemoveHeaders = append(append([]string(nil), out.RemoveHeaders...), override.RemoveHeaders...)
	}
	if override.Body.Enabled {
		out.Body = override.Body
	}
	return out
}

func mergeHTTPResponseMutations(base httpResponseMutations, override httpResponseMutations) httpResponseMutations {
	out := base
	if override.HasStatusCode {
		out.StatusCode = override.StatusCode
		out.HasStatusCode = true
	}
	if len(override.SetHeaders) > 0 {
		out.SetHeaders = mergeStringMap(out.SetHeaders, override.SetHeaders)
	}
	if len(override.RemoveHeaders) > 0 {
		out.RemoveHeaders = append(append([]string(nil), out.RemoveHeaders...), override.RemoveHeaders...)
	}
	if override.Body.Enabled {
		out.Body = override.Body
	}
	return out
}

func mergeStringMap(base map[string]string, override map[string]string) map[string]string {
	size := len(base) + len(override)
	out := make(map[string]string, size)
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}

func applyRequestMutations(req *http.Request, mutations httpRequestMutations) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	if req.URL == nil {
		return fmt.Errorf("request URL is nil")
	}
	if req.Header == nil {
		req.Header = http.Header{}
	}

	if strings.TrimSpace(mutations.Method) != "" {
		req.Method = strings.ToUpper(strings.TrimSpace(mutations.Method))
	}
	if mutations.Path != "" {
		req.URL.Path = mutations.Path
		req.URL.RawPath = ""
	}
	if mutations.AppendPath != "" {
		basePath := req.URL.Path
		if basePath == "" {
			basePath = "/"
		}
		req.URL.Path = joinURLPath(basePath, mutations.AppendPath)
		req.URL.RawPath = ""
	}
	if len(mutations.SetQuery) > 0 || len(mutations.RemoveQuery) > 0 {
		query := req.URL.Query()
		for k, v := range mutations.SetQuery {
			query.Set(k, v)
		}
		for _, k := range mutations.RemoveQuery {
			query.Del(k)
		}
		req.URL.RawQuery = query.Encode()
	}
	for k, v := range mutations.SetHeaders {
		req.Header.Set(k, v)
	}
	for _, headerName := range mutations.RemoveHeaders {
		deleteHeaderCaseInsensitive(req.Header, headerName)
	}
	if mutations.Body.Enabled {
		replaceRequestBody(req, mutations.Body.Raw)
	}
	removeHopByHopHeaders(req.Header)

	return nil
}

func applyResponseMutations(resp *http.Response, mutations httpResponseMutations) error {
	if resp == nil {
		return fmt.Errorf("response is nil")
	}
	if resp.Header == nil {
		resp.Header = http.Header{}
	}

	if mutations.HasStatusCode {
		resp.StatusCode = mutations.StatusCode
		statusText := http.StatusText(mutations.StatusCode)
		if statusText == "" {
			resp.Status = fmt.Sprintf("%d", mutations.StatusCode)
		} else {
			resp.Status = fmt.Sprintf("%d %s", mutations.StatusCode, statusText)
		}
	}
	for k, v := range mutations.SetHeaders {
		resp.Header.Set(k, v)
	}
	for _, headerName := range mutations.RemoveHeaders {
		deleteHeaderCaseInsensitive(resp.Header, headerName)
	}
	if mutations.Body.Enabled {
		replaceResponseBody(resp, mutations.Body.Raw)
	}

	return nil
}

func replaceRequestBody(req *http.Request, raw []byte) {
	resetRequestBody(req, raw)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", req.ContentLength))
	req.TransferEncoding = nil
	req.Header.Del("Transfer-Encoding")
}

func replaceResponseBody(resp *http.Response, raw []byte) {
	resetResponseBody(resp, raw)
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
	resp.TransferEncoding = nil
	resp.Header.Del("Transfer-Encoding")
	resp.Header.Del("Content-Encoding")
}

func resetRequestBody(req *http.Request, raw []byte) {
	body := append([]byte(nil), raw...)
	if len(body) == 0 {
		req.Body = http.NoBody
	} else {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	req.GetBody = func() (io.ReadCloser, error) {
		if len(body) == 0 {
			return http.NoBody, nil
		}
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	req.ContentLength = int64(len(body))
}

func resetResponseBody(resp *http.Response, raw []byte) {
	body := append([]byte(nil), raw...)
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
	if len(body) == 0 {
		resp.Body = http.NoBody
	} else {
		resp.Body = io.NopCloser(bytes.NewReader(body))
	}
	resp.ContentLength = int64(len(body))
}

func deleteHeaderCaseInsensitive(headers http.Header, name string) {
	for k := range headers {
		if strings.EqualFold(k, name) {
			headers.Del(k)
		}
	}
}

func removeHopByHopHeaders(headers http.Header) {
	for _, connectionValue := range headers.Values("Connection") {
		for _, token := range strings.Split(connectionValue, ",") {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}
			deleteHeaderCaseInsensitive(headers, token)
		}
	}
	for _, hopByHop := range []string{
		"Connection",
		"Proxy-Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	} {
		deleteHeaderCaseInsensitive(headers, hopByHop)
	}
}

func startStaticHTTPServer(cfg httpServerConfig) (Value, error) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for k, v := range cfg.Headers {
			s, err := asStringValue("http.server options.headers["+k+"]", v)
			if err != nil {
				continue
			}
			w.Header().Set(k, s)
		}
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", cfg.ContentType)
		}
		w.WriteHeader(cfg.StatusCode)
		_, _ = w.Write([]byte(cfg.Body))
	})
	return startManagedHTTPServer(cfg.Addr, handler)
}

func startReverseProxyHTTPServer(cfg httpProxyServerConfig) (Value, error) {
	targetURL, err := url.Parse(cfg.Target)
	if err != nil {
		return nil, fmt.Errorf("http.server.proxy options.target parse error: %w", err)
	}
	if targetURL.Scheme == "" || targetURL.Host == "" {
		return nil, fmt.Errorf("http.server.proxy options.target must include scheme and host, got %q", cfg.Target)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		originalPath := req.URL.Path
		if cfg.StripPrefix != "" && strings.HasPrefix(originalPath, cfg.StripPrefix) {
			originalPath = strings.TrimPrefix(originalPath, cfg.StripPrefix)
			if originalPath == "" {
				originalPath = "/"
			} else if !strings.HasPrefix(originalPath, "/") {
				originalPath = "/" + originalPath
			}
		}

		outURL := *targetURL
		outURL.Path = joinURLPath(targetURL.Path, originalPath)
		outURL.RawPath = ""
		outURL.RawQuery = mergeURLQuery(targetURL.RawQuery, req.URL.RawQuery)

		outReq, buildErr := http.NewRequestWithContext(req.Context(), req.Method, outURL.String(), req.Body)
		if buildErr != nil {
			http.Error(w, buildErr.Error(), http.StatusBadGateway)
			return
		}
		outReq.Header = req.Header.Clone()
		if cfg.PreserveHost {
			outReq.Host = req.Host
		} else {
			outReq.Host = targetURL.Host
		}

		for k, v := range cfg.Headers {
			s, convErr := asStringValue("http.server proxy options.headers["+k+"]", v)
			if convErr != nil {
				continue
			}
			outReq.Header.Set(k, s)
		}

		requestMutations, mutationErr := resolveRequestMutations(cfg.RequestMutations, outReq)
		if mutationErr != nil {
			http.Error(w, mutationErr.Error(), http.StatusBadGateway)
			return
		}
		if err := applyRequestMutations(outReq, requestMutations); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		requestContext, requestContextErr := buildRequestContextForResponseMutation(outReq)
		if requestContextErr != nil {
			http.Error(w, requestContextErr.Error(), http.StatusBadGateway)
			return
		}

		resp, doErr := client.Do(outReq)
		if doErr != nil {
			http.Error(w, doErr.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		responseMutations, mutationErr := resolveResponseMutations(cfg.ResponseMutations, resp, requestContext)
		if mutationErr != nil {
			http.Error(w, mutationErr.Error(), http.StatusBadGateway)
			return
		}
		if err := applyResponseMutations(resp, responseMutations); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		removeHopByHopHeaders(resp.Header)
		writeHTTPResponse(w, resp)
	})

	return startManagedHTTPServer(cfg.Addr, handler)
}

func startForwardHTTPServer(cfg httpForwardServerConfig) (Value, error) {
	targetURL, err := url.Parse(cfg.Target)
	if err != nil {
		return nil, fmt.Errorf("http.server.forward options.target parse error: %w", err)
	}
	if targetURL.Scheme == "" || targetURL.Host == "" {
		return nil, fmt.Errorf("http.server.forward options.target must include scheme and host, got %q", cfg.Target)
	}

	client := &http.Client{Timeout: cfg.Timeout}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		outPath := req.URL.Path
		if !cfg.KeepPath {
			outPath = "/"
		}
		if cfg.Path != "" {
			outPath = cfg.Path
		}

		outURL := *targetURL
		outURL.Path = joinURLPath(targetURL.Path, outPath)
		outURL.RawPath = ""
		outURL.RawQuery = mergeURLQuery(targetURL.RawQuery, req.URL.RawQuery)

		outReq, buildErr := http.NewRequestWithContext(req.Context(), req.Method, outURL.String(), req.Body)
		if buildErr != nil {
			http.Error(w, buildErr.Error(), http.StatusBadGateway)
			return
		}
		outReq.Header = req.Header.Clone()
		if cfg.PreserveHost {
			outReq.Host = req.Host
		} else {
			outReq.Host = targetURL.Host
		}
		for k, v := range cfg.Headers {
			s, convErr := asStringValue("http.server forward options.headers["+k+"]", v)
			if convErr != nil {
				continue
			}
			outReq.Header.Set(k, s)
		}

		requestMutations, mutationErr := resolveRequestMutations(cfg.RequestMutations, outReq)
		if mutationErr != nil {
			http.Error(w, mutationErr.Error(), http.StatusBadGateway)
			return
		}
		if err := applyRequestMutations(outReq, requestMutations); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		requestContext, requestContextErr := buildRequestContextForResponseMutation(outReq)
		if requestContextErr != nil {
			http.Error(w, requestContextErr.Error(), http.StatusBadGateway)
			return
		}

		resp, doErr := client.Do(outReq)
		if doErr != nil {
			http.Error(w, doErr.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		responseMutations, mutationErr := resolveResponseMutations(cfg.ResponseMutations, resp, requestContext)
		if mutationErr != nil {
			http.Error(w, mutationErr.Error(), http.StatusBadGateway)
			return
		}
		if err := applyResponseMutations(resp, responseMutations); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		removeHopByHopHeaders(resp.Header)
		writeHTTPResponse(w, resp)
	})

	return startManagedHTTPServer(cfg.Addr, handler)
}

func writeHTTPResponse(w http.ResponseWriter, resp *http.Response) {
	for k, vv := range resp.Header {
		for _, hv := range vv {
			w.Header().Add(k, hv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func startManagedHTTPServer(addr string, handler http.Handler) (Value, error) {
	server := &http.Server{Handler: handler}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	go func() {
		_ = server.Serve(ln)
	}()
	return Object{
		"addr": ln.Addr().String(),
		"close": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("http.server.close expects 0 args, got %d", len(args))
			}
			closeErr := server.Close()
			if closeErr != nil && !errors.Is(closeErr, http.ErrServerClosed) {
				return nil, closeErr
			}
			return true, nil
		}),
	}, nil
}

func joinURLPath(basePath string, requestPath string) string {
	if requestPath == "" {
		requestPath = "/"
	}
	if basePath == "" {
		return requestPath
	}
	baseHasSlash := strings.HasSuffix(basePath, "/")
	reqHasSlash := strings.HasPrefix(requestPath, "/")
	switch {
	case baseHasSlash && reqHasSlash:
		return basePath + requestPath[1:]
	case !baseHasSlash && !reqHasSlash:
		return basePath + "/" + requestPath
	default:
		return basePath + requestPath
	}
}

func mergeURLQuery(baseRawQuery string, reqRawQuery string) string {
	switch {
	case baseRawQuery == "":
		return reqRawQuery
	case reqRawQuery == "":
		return baseRawQuery
	default:
		return baseRawQuery + "&" + reqRawQuery
	}
}

func doHTTPRequest(client *http.Client, cfg httpRequestConfig) (Value, error) {
	var bodyReader io.Reader
	if cfg.Method == http.MethodGet || cfg.Body == "" {
		bodyReader = nil
	} else {
		bodyReader = bytes.NewBufferString(cfg.Body)
	}
	req, err := http.NewRequest(cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return nil, err
	}
	for k, v := range cfg.Headers {
		s, err := asStringValue("http options.headers["+k+"]", v)
		if err != nil {
			return nil, err
		}
		req.Header.Set(k, s)
	}
	if cfg.Body != "" && cfg.Method != http.MethodGet && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", cfg.ContentType)
	}

	reqClient, err := buildHTTPClient(client, cfg.Timeout, cfg.Proxy)
	if err != nil {
		return nil, err
	}
	resp, err := reqClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return Object{
		"ok":         resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status":     resp.Status,
		"statusCode": float64(resp.StatusCode),
		"body":       string(raw),
		"headers":    headersToObject(resp.Header),
	}, nil
}

func doHTTPStreamRequest(client *http.Client, cfg httpRequestConfig, asyncRuntime AsyncRuntime) (Value, error) {
	var bodyReader io.Reader
	if cfg.Method == http.MethodGet || cfg.Body == "" {
		bodyReader = nil
	} else {
		bodyReader = bytes.NewBufferString(cfg.Body)
	}
	req, err := http.NewRequest(cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return nil, err
	}
	for k, v := range cfg.Headers {
		s, err := asStringValue("http options.headers["+k+"]", v)
		if err != nil {
			return nil, err
		}
		req.Header.Set(k, s)
	}
	if cfg.Body != "" && cfg.Method != http.MethodGet && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", cfg.ContentType)
	}

	reqClient, err := buildHTTPClient(client, cfg.Timeout, cfg.Proxy)
	if err != nil {
		return nil, err
	}
	resp, err := reqClient.Do(req)
	if err != nil {
		return nil, err
	}

	var ioMu sync.Mutex
	var closeOnce sync.Once
	closeBody := func() error {
		var closeErr error
		closeOnce.Do(func() {
			closeErr = resp.Body.Close()
		})
		return closeErr
	}
	recvFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("http.client.stream.recv expects 0 args, got %d", len(args))
		}
		ioMu.Lock()
		defer ioMu.Unlock()
		buf := make([]byte, cfg.ChunkSize)
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			return Object{
				"chunk": string(buf[:n]),
				"done":  false,
			}, nil
		}
		if errors.Is(readErr, io.EOF) {
			_ = closeBody()
			return Object{
				"chunk": "",
				"done":  true,
			}, nil
		}
		if readErr != nil {
			_ = closeBody()
			return nil, readErr
		}
		return Object{
			"chunk": "",
			"done":  false,
		}, nil
	})
	recvAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return recvFn(args)
		}), nil
	})
	closeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("http.client.stream.close expects 0 args, got %d", len(args))
		}
		if err := closeBody(); err != nil {
			return nil, err
		}
		return true, nil
	})

	return Object{
		"ok":         resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status":     resp.Status,
		"statusCode": float64(resp.StatusCode),
		"headers":    headersToObject(resp.Header),
		"recv":       recvFn,
		"recvAsync":  recvAsyncFn,
		"close":      closeFn,
	}, nil
}

func buildHTTPClient(base *http.Client, timeout time.Duration, proxy string) (*http.Client, error) {
	needTimeout := timeout > 0 && timeout != base.Timeout
	needProxy := strings.TrimSpace(proxy) != ""
	if !needTimeout && !needProxy {
		return base, nil
	}

	clone := *base
	if needTimeout {
		clone.Timeout = timeout
	}

	if needProxy {
		parsedProxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("http options.proxy parse error: %w", err)
		}
		if parsedProxyURL.Scheme == "" || parsedProxyURL.Host == "" {
			return nil, fmt.Errorf("http options.proxy must include scheme and host, got %q", proxy)
		}

		var baseTransport *http.Transport
		switch t := clone.Transport.(type) {
		case nil:
			defaultTransport, ok := http.DefaultTransport.(*http.Transport)
			if !ok {
				return nil, fmt.Errorf("http default transport is not *http.Transport")
			}
			baseTransport = defaultTransport.Clone()
		case *http.Transport:
			baseTransport = t.Clone()
		default:
			return nil, fmt.Errorf("http options.proxy requires *http.Transport, got %T", clone.Transport)
		}
		baseTransport.Proxy = http.ProxyURL(parsedProxyURL)
		clone.Transport = baseTransport
	}

	return &clone, nil
}
