package api

import "context"

type Host interface {
	AcquireCapability(ctx context.Context, req AcquireRequest) (CapabilityInstance, error)
	ReleaseCapability(ctx context.Context, capID string) error
	Call(ctx context.Context, req CallRequest) (CallResult, error)
	Poll(ctx context.Context, handleID uint64) (PollResult, error)
}

type Waiter interface {
	Wait(ctx context.Context, handleID uint64) (PollResult, error)
}
