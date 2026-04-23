package api

import hostnet "iacommon/pkg/host/network"

type NetworkHTTPFetchRequest struct {
	Method    string
	URL       string
	Headers   map[string]string
	Body      []byte
	TimeoutMS int64
}

type NetworkHTTPFetchResponse struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

func decodeNetworkHTTPFetchRequest(args map[string]any) (NetworkHTTPFetchRequest, error) {
	requestURL, err := readString(args, "url")
	if err != nil {
		return NetworkHTTPFetchRequest{}, err
	}
	method, err := readOptionalStringAny(args, "method")
	if err != nil {
		return NetworkHTTPFetchRequest{}, err
	}
	headers, err := readStringMap(args, "headers")
	if err != nil {
		return NetworkHTTPFetchRequest{}, err
	}
	body, err := readOptionalBytes(args, "body")
	if err != nil {
		return NetworkHTTPFetchRequest{}, err
	}
	timeoutMS, err := readOptionalInt64Any(args, "timeout_ms", "timeoutMS")
	if err != nil {
		return NetworkHTTPFetchRequest{}, err
	}

	return NetworkHTTPFetchRequest{
		Method:    method,
		URL:       requestURL,
		Headers:   headers,
		Body:      body,
		TimeoutMS: timeoutMS,
	}, nil
}

func (r NetworkHTTPFetchRequest) toProviderRequest() hostnet.HTTPRequest {
	return hostnet.HTTPRequest{
		Method:    r.Method,
		URL:       r.URL,
		Headers:   r.Headers,
		Body:      r.Body,
		TimeoutMS: r.TimeoutMS,
	}
}

func encodeNetworkHTTPFetchResponse(resp *hostnet.HTTPResponse) CallResult {
	if resp == nil {
		return CallResult{Value: map[string]any{}}
	}
	return CallResult{Value: map[string]any{
		"status":  resp.Status,
		"headers": resp.Headers,
		"body":    resp.Body,
	}}
}
