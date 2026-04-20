package network

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPProvider struct {
	Policy Policy
	Client *http.Client
}

func (p *HTTPProvider) HTTPFetch(ctx context.Context, req HTTPRequest) (*HTTPResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	req.Method = method
	if _, err := p.Policy.ValidateHTTPRequest(req); err != nil {
		return nil, err
	}

	var body io.Reader
	if len(req.Body) > 0 {
		body = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, req.URL, body)
	if err != nil {
		return nil, err
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	client := p.httpClient(req.TimeoutMS)
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		Status:  resp.StatusCode,
		Headers: flattenHeaders(resp.Header),
		Body:    responseBody,
	}, nil
}

func (p *HTTPProvider) Dial(ctx context.Context, endpoint Endpoint, opts DialOptions) (SocketHandle, error) {
	_, _, _, _ = ctx, endpoint, opts, p
	return nil, ErrNetworkOperationNotSupported
}

func (p *HTTPProvider) Listen(ctx context.Context, endpoint Endpoint, opts ListenOptions) (ListenerHandle, error) {
	_, _, _, _ = ctx, endpoint, opts, p
	return nil, ErrNetworkOperationNotSupported
}

func (p *HTTPProvider) httpClient(timeoutMS int64) *http.Client {
	base := p.Client
	if base == nil {
		base = &http.Client{}
	}
	clone := *base
	if timeoutMS > 0 {
		clone.Timeout = time.Duration(timeoutMS) * time.Millisecond
	} else if clone.Timeout <= 0 {
		clone.Timeout = 15 * time.Second
	}
	return &clone
}

func flattenHeaders(headers http.Header) map[string]string {
	result := make(map[string]string, len(headers))
	for key, values := range headers {
		result[key] = strings.Join(values, ",")
	}
	return result
}
