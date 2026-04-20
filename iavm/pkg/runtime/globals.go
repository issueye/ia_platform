package runtime

type Globals struct {
	values map[string]any
}

func NewGlobals() *Globals {
	return &Globals{values: map[string]any{}}
}
