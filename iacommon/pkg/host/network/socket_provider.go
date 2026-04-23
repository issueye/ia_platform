package network

import (
	"context"
	"net"
	"strconv"
	"time"
)

type SocketProvider struct {
	Policy Policy
}

type netSocketHandle struct {
	conn net.Conn
}

type netListenerHandle struct {
	listener net.Listener
}

func (p *SocketProvider) HTTPFetch(ctx context.Context, req HTTPRequest) (*HTTPResponse, error) {
	_, _, _ = ctx, req, p
	return nil, ErrNetworkOperationNotSupported
}

func (p *SocketProvider) Dial(ctx context.Context, endpoint Endpoint, opts DialOptions) (SocketHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := p.Policy.ValidateEndpoint(endpoint); err != nil {
		return nil, err
	}

	dialer := net.Dialer{}
	if opts.TimeoutMS > 0 {
		dialer.Timeout = time.Duration(opts.TimeoutMS) * time.Millisecond
	}

	conn, err := dialer.DialContext(ctx, endpoint.Network, net.JoinHostPort(endpoint.Host, strconv.Itoa(endpoint.Port)))
	if err != nil {
		return nil, err
	}
	return &netSocketHandle{conn: conn}, nil
}

func (p *SocketProvider) Listen(ctx context.Context, endpoint Endpoint, opts ListenOptions) (ListenerHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := p.Policy.ValidateEndpoint(endpoint); err != nil {
		return nil, err
	}

	config := net.ListenConfig{}
	listener, err := config.Listen(ctx, endpoint.Network, net.JoinHostPort(endpoint.Host, strconv.Itoa(endpoint.Port)))
	if err != nil {
		return nil, err
	}
	return &netListenerHandle{listener: listener}, nil
}

func (h *netSocketHandle) Send(ctx context.Context, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := h.conn.SetWriteDeadline(deadline); err != nil {
			return 0, err
		}
	} else {
		_ = h.conn.SetWriteDeadline(time.Time{})
	}
	return h.conn.Write(data)
}

func (h *netSocketHandle) Recv(ctx context.Context, size int) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if size <= 0 {
		size = 4096
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := h.conn.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
	} else {
		_ = h.conn.SetReadDeadline(time.Time{})
	}
	buf := make([]byte, size)
	n, err := h.conn.Read(buf)
	if n > 0 {
		return buf[:n], nil
	}
	return nil, err
}

func (h *netSocketHandle) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return h.conn.Close()
}

func (h *netListenerHandle) Accept(ctx context.Context) (SocketHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if deadline, ok := ctx.Deadline(); ok {
		if tcpListener, ok := h.listener.(*net.TCPListener); ok {
			if err := tcpListener.SetDeadline(deadline); err != nil {
				return nil, err
			}
		}
	}
	conn, err := h.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &netSocketHandle{conn: conn}, nil
}

func (h *netListenerHandle) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return h.listener.Close()
}
