package module

type CapabilityKind string

const (
	CapabilityFS      CapabilityKind = "fs"
	CapabilityNetwork CapabilityKind = "network"
)

type CapabilityDecl struct {
	Kind     CapabilityKind
	Required bool
	Config   map[string]any
}
