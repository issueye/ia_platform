package network

import "context"

type Provider interface {
	HTTPFetch(ctx context.Context, req HTTPRequest) (*HTTPResponse, error)
	Dial(ctx context.Context, endpoint Endpoint, opts DialOptions) (SocketHandle, error)
	Listen(ctx context.Context, endpoint Endpoint, opts ListenOptions) (ListenerHandle, error)
}

type HTTPRequest struct {
	Method    string
	URL       string
	Headers   map[string]string
	Body      []byte
	TimeoutMS int64
}

type HTTPResponse struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

type Endpoint struct {
	Network string
	Host    string
	Port    int
}

type DialOptions struct {
	TimeoutMS int64
}

type ListenOptions struct {
	Backlog int
}

type SocketHandle interface {
	Send(ctx context.Context, data []byte) (int, error)
	Recv(ctx context.Context, size int) ([]byte, error)
	Close(ctx context.Context) error
}

type ListenerHandle interface {
	Accept(ctx context.Context) (SocketHandle, error)
	Close(ctx context.Context) error
}
