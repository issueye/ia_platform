package module

type ImportKind uint8

const (
	ImportFunction ImportKind = iota
	ImportGlobal
	ImportCapability
)

type Import struct {
	Module string
	Name   string
	Kind   ImportKind
	Type   uint32
}

type ExportKind uint8

const (
	ExportFunction ExportKind = iota
	ExportGlobal
)

type Export struct {
	Name  string
	Kind  ExportKind
	Index uint32
}
