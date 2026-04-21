package runtime

import "iavm/pkg/core"

type Stack struct {
	values []core.Value
}

func NewStack(capacity int) *Stack {
	if capacity < 0 {
		capacity = 0
	}
	return &Stack{values: make([]core.Value, 0, capacity)}
}

func (s *Stack) Push(val core.Value) {
	s.values = append(s.values, val)
}

func (s *Stack) Pop() core.Value {
	if len(s.values) == 0 {
		return core.Value{Kind: core.ValueNull}
	}
	val := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return val
}

func (s *Stack) Peek(offset int) core.Value {
	idx := len(s.values) - 1 - offset
	if idx < 0 || idx >= len(s.values) {
		return core.Value{Kind: core.ValueNull}
	}
	return s.values[idx]
}

func (s *Stack) Size() int {
	return len(s.values)
}
