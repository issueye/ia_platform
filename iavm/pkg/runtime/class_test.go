package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	compiler "ialang/pkg/lang/compiler"
	frontend "ialang/pkg/lang/frontend"
	bridge_ialang "iavm/pkg/bridge/ialang"
	"iavm/pkg/core"
)

func TestClassInheritanceSuperRuntime(t *testing.T) {
	source := `
class Animal {
  constructor(name) {
    this.name = name;
  }
  speak() {
    return this.name + " makes a sound";
  }
}

class Dog extends Animal {
  constructor(name) {
    super(name);
  }
  speak() {
    return super.speak();
  }
}

let dog = new Dog("Buddy");
let result = dog.speak();
`

	lexer := frontend.NewLexer(source)
	parser := frontend.NewParser(lexer)
	program := parser.ParseProgram()
	if errs := parser.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	compiled, errs := compiler.NewCompiler().Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	mod, err := bridge_ialang.LowerToModule(compiled)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	var got core.Value
	found := false
	for i, global := range mod.Globals {
		if global.Name == "result" && i < len(vm.globals) {
			got = vm.globals[i]
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected global result")
	}
	if got.Kind != core.ValueString || !strings.Contains(got.Raw.(string), "Buddy makes a sound") {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestClassInheritanceExampleFileRuntime(t *testing.T) {
	sourcePath := filepath.Join("..", "..", "..", "ialang", "examples", "inheritance.ia")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	lexer := frontend.NewLexer(string(source))
	parser := frontend.NewParser(lexer)
	program := parser.ParseProgram()
	if errs := parser.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	compiled, errs := compiler.NewCompiler().Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	mod, err := bridge_ialang.LowerToModule(compiled)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}
