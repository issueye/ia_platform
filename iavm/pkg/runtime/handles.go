package runtime

type HandleKind uint8

const (
	HandleFile HandleKind = iota
	HandleSocket
	HandleListener
	HandleHTTPStream
)

type HandleEntry struct {
	ID    uint64
	Kind  HandleKind
	Value any
}

type HandleTable struct {
	nextID  uint64
	entries map[uint64]HandleEntry
}

func NewHandleTable() *HandleTable {
	return &HandleTable{
		nextID:  1,
		entries: map[uint64]HandleEntry{},
	}
}
