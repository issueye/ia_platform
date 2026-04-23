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

func (t *HandleTable) Add(kind HandleKind, value any) uint64 {
	id := t.nextID
	t.nextID++
	t.entries[id] = HandleEntry{ID: id, Kind: kind, Value: value}
	return id
}

func (t *HandleTable) Get(id uint64) (HandleEntry, bool) {
	entry, ok := t.entries[id]
	return entry, ok
}

func (t *HandleTable) Remove(id uint64) (HandleEntry, bool) {
	entry, ok := t.entries[id]
	if !ok {
		return HandleEntry{}, false
	}
	delete(t.entries, id)
	return entry, true
}
