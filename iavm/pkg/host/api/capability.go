package api

type CapabilityKind string

const (
	CapabilityFS      CapabilityKind = "fs"
	CapabilityNetwork CapabilityKind = "network"
)

type AcquireRequest struct {
	Kind   CapabilityKind
	Config map[string]any
}

type CapabilityInstance struct {
	ID     string
	Kind   CapabilityKind
	Rights []string
	Meta   map[string]any
}
