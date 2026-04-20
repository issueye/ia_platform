package network

import "context"

type SocketProvider struct {
	Policy Policy
}

func (p *SocketProvider) HTTPFetch(ctx context.Context, req HTTPRequest) (*HTTPResponse, error) {
	_, _, _ = ctx, req, p
	return nil, nil
}

func (p *SocketProvider) Dial(ctx context.Context, endpoint Endpoint, opts DialOptions) (SocketHandle, error) {
	_, _, _, _ = ctx, endpoint, opts, p
	return nil, nil
}

func (p *SocketProvider) Listen(ctx context.Context, endpoint Endpoint, opts ListenOptions) (ListenerHandle, error) {
	_, _, _, _ = ctx, endpoint, opts, p
	return nil, nil
}
