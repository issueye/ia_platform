package runtime

import (
	"iacommon/pkg/host/api"
	"time"
)

type Options struct {
	MaxSteps      int64
	MaxMemory     int64
	MaxDuration   time.Duration
	Host          api.Host
	EnableTracing bool
}
