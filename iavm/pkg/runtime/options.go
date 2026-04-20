package runtime

import (
	"time"
	"iavm/pkg/host/api"
)

type Options struct {
	MaxSteps      int64
	MaxMemory     int64
	MaxDuration   time.Duration
	Host          api.Host
	EnableTracing bool
}
