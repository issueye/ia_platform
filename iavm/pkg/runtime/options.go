package runtime

import (
	"iacommon/pkg/host/api"
	"time"
)

type Options struct {
	MaxSteps        int64
	MaxMemory       int64
	MaxDuration     time.Duration
	HostTimeout     time.Duration
	WaitTimeout     time.Duration
	RetryCount      int
	RetryBackoff    time.Duration
	RetryMaxBackoff time.Duration
	RetryMultiplier float64
	RetryJitter     float64
	RetryCallOps    []string
	WaitInterval    time.Duration
	Host            api.Host
	EnableTracing   bool
}
