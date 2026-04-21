package module

type BuiltinProvider[T any] func() T

type BuiltinRegistry[T any] struct {
	providers map[string]BuiltinProvider[T]
}

func NewBuiltinRegistry[T any]() *BuiltinRegistry[T] {
	return &BuiltinRegistry[T]{
		providers: map[string]BuiltinProvider[T]{},
	}
}

func BuiltinRegistryFromValues[T any](values map[string]T) *BuiltinRegistry[T] {
	registry := NewBuiltinRegistry[T]()
	for name, value := range values {
		registry.RegisterValue(name, value)
	}
	return registry
}

func (r *BuiltinRegistry[T]) RegisterValue(name string, value T) {
	r.RegisterProvider(name, func() T { return value })
}

func (r *BuiltinRegistry[T]) RegisterProvider(name string, provider BuiltinProvider[T]) {
	if r.providers == nil {
		r.providers = map[string]BuiltinProvider[T]{}
	}
	r.providers[name] = provider
}

func (r *BuiltinRegistry[T]) Has(name string) bool {
	if r == nil {
		return false
	}
	_, ok := r.providers[name]
	return ok
}

func (r *BuiltinRegistry[T]) Resolve(name string) (T, bool) {
	var zero T
	if r == nil {
		return zero, false
	}
	provider, ok := r.providers[name]
	if !ok {
		return zero, false
	}
	return provider(), true
}
