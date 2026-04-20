package compiler

import (
	"testing"

	"ialang/pkg/lang/frontend"
)

func TestCompileBasicProgram(t *testing.T) {
	src := `
let x = 1 + 2;
let y = x - 1;
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	if chunk == nil {
		t.Fatal("chunk is nil")
	}
	if len(chunk.Code) == 0 {
		t.Fatal("chunk has no instructions")
	}
}

func TestCompileAsyncAwaitProgram(t *testing.T) {
	src := `
async function f() { return 1; }
let p = f();
let v = await p;
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	foundAwait := false
	for _, ins := range chunk.Code {
		if ins.Op == OpAwait {
			foundAwait = true
			break
		}
	}
	if !foundAwait {
		t.Fatal("compiled chunk does not contain OpAwait")
	}
}

func TestCompileExportAlias(t *testing.T) {
	src := `
let value = 1;
export { value as answer };
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	foundExportAs := false
	for _, ins := range chunk.Code {
		if ins.Op == OpExportAs {
			foundExportAs = true
			break
		}
	}
	if !foundExportAs {
		t.Fatal("compiled chunk does not contain OpExportAs")
	}
}

func TestCompileTemplateAndRestParams(t *testing.T) {
	src := `
function join(head, ...rest) {
  return ` + "`" + `h=${head}, n=${rest.length}` + "`" + `;
}
let result = join("x", 1, 2);
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	_, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
}

func TestCompileImportNamespaceAndExportDefault(t *testing.T) {
	src := `
import * as mod from "m";
let x = mod.value;
export default x;
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}

	foundImportNS := false
	foundExportDefault := false
	for _, ins := range chunk.Code {
		if ins.Op == OpImportNamespace {
			foundImportNS = true
		}
		if ins.Op == OpExportDefault {
			foundExportDefault = true
		}
	}
	if !foundImportNS {
		t.Fatal("compiled chunk does not contain OpImportNamespace")
	}
	if !foundExportDefault {
		t.Fatal("compiled chunk does not contain OpExportDefault")
	}
}

func TestCompileExportDefaultClass(t *testing.T) {
	src := `
export default class Counter {
  constructor() { this.v = 1; }
}
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	foundExportDefault := false
	for _, ins := range chunk.Code {
		if ins.Op == OpExportDefault {
			foundExportDefault = true
			break
		}
	}
	if !foundExportDefault {
		t.Fatal("compiled chunk does not contain OpExportDefault for default class")
	}
}

func TestCompileExportNamedClass(t *testing.T) {
	src := `
export class Counter {
  constructor() { this.v = 1; }
}
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	foundClass := false
	foundExportName := false
	for _, ins := range chunk.Code {
		if ins.Op == OpClass {
			foundClass = true
		}
		if ins.Op == OpExportName {
			foundExportName = true
		}
	}
	if !foundClass {
		t.Fatal("compiled chunk does not contain OpClass for named class export")
	}
	if !foundExportName {
		t.Fatal("compiled chunk does not contain OpExportName for named class export")
	}
}

func TestCompileDestructuringLet(t *testing.T) {
	src := `
let [a, b] = arr;
let {x, y: z} = obj;
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	_, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
}

func TestCompileDestructuringAssign(t *testing.T) {
	src := `
let a = 0;
let b = 0;
[a, b] = arr;
let x = 0;
let z = 0;
({x, y: z} = obj);
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	_, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
}

func TestCompileExportAllAndDynamicImport(t *testing.T) {
	src := `
export * from "./dep";
let mod = import("./dep");
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}

	foundExportAll := false
	foundImportDynamic := false
	for _, ins := range chunk.Code {
		if ins.Op == OpExportAll {
			foundExportAll = true
		}
		if ins.Op == OpImportDynamic {
			foundImportDynamic = true
		}
	}
	if !foundExportAll {
		t.Fatal("compiled chunk does not contain OpExportAll")
	}
	if !foundImportDynamic {
		t.Fatal("compiled chunk does not contain OpImportDynamic")
	}
}

func TestCompileClassPrivateField(t *testing.T) {
	src := `
class Counter {
  #value;
  constructor() { this.#value = 1; }
  readValue() { return this.#value; }
}
let c = new Counter();
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	_, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
}
