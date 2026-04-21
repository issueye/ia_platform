package module

import "testing"

func TestBuiltinRegistryResolvesRegisteredValues(t *testing.T) {
	registry := NewBuiltinRegistry[int]()
	registry.RegisterValue("answer", 42)

	value, ok := registry.Resolve("answer")
	if !ok {
		t.Fatal("Resolve(answer) = missing")
	}
	if value != 42 {
		t.Fatalf("Resolve(answer) = %d, want 42", value)
	}
	if !registry.Has("answer") {
		t.Fatal("Has(answer) = false")
	}
}

func TestBuiltinRegistryResolvesProviderValues(t *testing.T) {
	count := 0
	registry := NewBuiltinRegistry[int]()
	registry.RegisterProvider("counter", func() int {
		count++
		return count
	})

	first, ok := registry.Resolve("counter")
	if !ok {
		t.Fatal("Resolve(counter) first = missing")
	}
	second, ok := registry.Resolve("counter")
	if !ok {
		t.Fatal("Resolve(counter) second = missing")
	}
	if first != 1 || second != 2 {
		t.Fatalf("Resolve(counter) sequence = (%d, %d), want (1, 2)", first, second)
	}
}

func TestBuiltinRegistryFromValuesCopiesEntries(t *testing.T) {
	values := map[string]string{"fs": "builtin"}
	registry := BuiltinRegistryFromValues(values)
	values["fs"] = "mutated"

	resolved, ok := registry.Resolve("fs")
	if !ok {
		t.Fatal("Resolve(fs) = missing")
	}
	if resolved != "builtin" {
		t.Fatalf("Resolve(fs) = %q, want %q", resolved, "builtin")
	}
}

func TestBuiltinRegistryNilSafety(t *testing.T) {
	var registry *BuiltinRegistry[int]
	if registry.Has("missing") {
		t.Fatal("nil registry Has(missing) = true")
	}
	if _, ok := registry.Resolve("missing"); ok {
		t.Fatal("nil registry Resolve(missing) = found")
	}
}
