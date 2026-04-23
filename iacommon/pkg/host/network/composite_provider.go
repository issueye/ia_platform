package network

import "context"

type CompositeProvider struct {
	HTTP   *HTTPProvider
	Socket *SocketProvider
}

func (p *CompositeProvider) HTTPFetch(ctx context.Context, req HTTPRequest) (*HTTPResponse, error) {
	if p != nil && p.HTTP != nil {
		return p.HTTP.HTTPFetch(ctx, req)
	}
	return nil, ErrNetworkOperationNotSupported
}

func (p *CompositeProvider) Dial(ctx context.Context, endpoint Endpoint, opts DialOptions) (SocketHandle, error) {
	if p != nil && p.Socket != nil {
		return p.Socket.Dial(ctx, endpoint, opts)
	}
	return nil, ErrNetworkOperationNotSupported
}

func (p *CompositeProvider) Listen(ctx context.Context, endpoint Endpoint, opts ListenOptions) (ListenerHandle, error) {
	if p != nil && p.Socket != nil {
		return p.Socket.Listen(ctx, endpoint, opts)
	}
	return nil, ErrNetworkOperationNotSupported
}
