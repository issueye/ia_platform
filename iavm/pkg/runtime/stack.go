package runtime

type Stack struct {
	values []any
}

func NewStack(capacity int) *Stack {
	if capacity < 0 {
		capacity = 0
	}
	return &Stack{values: make([]any, 0, capacity)}
}
