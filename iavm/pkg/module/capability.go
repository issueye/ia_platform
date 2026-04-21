package module

import hostapi "iacommon/pkg/host/api"

type CapabilityKind = hostapi.CapabilityKind

const (
	CapabilityFS      = hostapi.CapabilityFS
	CapabilityNetwork = hostapi.CapabilityNetwork
)

type CapabilityDecl struct {
	Kind     CapabilityKind
	Required bool
	Config   map[string]any
}
