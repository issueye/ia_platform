package lang

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestModuleLoaderResolveConcurrentDifferentModules(t *testing.T) {
	dir := t.TempDir()
	moduleAPath := filepath.Join(dir, "a.ia")
	moduleBPath := filepath.Join(dir, "b.ia")

	writeTestModule(t, moduleAPath, "export let value = 1;")
	writeTestModule(t, moduleBPath, "export let value = 2;")

	loader := NewModuleLoader(nil, nil)
	fromPath := filepath.Join(dir, "main.ia")

	var wg sync.WaitGroup
	errCh := make(chan error, 40)

	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			moduleName := "./a"
			if i%2 == 1 {
				moduleName = "./b"
			}
			_, err := loader.Resolve(fromPath, moduleName)
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
}

func TestModuleLoaderResolveCyclicImport(t *testing.T) {
	dir := t.TempDir()
	moduleAPath := filepath.Join(dir, "a.ia")
	moduleBPath := filepath.Join(dir, "b.ia")

	writeTestModule(t, moduleAPath, `
import { b } from "./b";
export let a = b + 1;
`)
	writeTestModule(t, moduleBPath, `
import { a } from "./a";
export let b = a + 1;
`)

	loader := NewModuleLoader(nil, nil)
	fromPath := filepath.Join(dir, "main.ia")

	_, err := loader.Resolve(fromPath, "./a")
	if err == nil {
		t.Fatal("Resolve() expected cyclic import error, got nil")
	}
	if !strings.Contains(err.Error(), "cyclic import detected") {
		t.Fatalf("Resolve() error = %q, want contains %q", err.Error(), "cyclic import detected")
	}
}

func TestModuleLoaderResolveExportAlias(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "mod.ia")
	writeTestModule(t, modulePath, `
let value = 42;
export { value as answer };
`)

	loader := NewModuleLoader(nil, nil)
	fromPath := filepath.Join(dir, "main.ia")

	loaded, err := loader.Resolve(fromPath, "./mod")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	obj, ok := loaded.(Object)
	if !ok {
		t.Fatalf("Resolve() value type = %T, want Object", loaded)
	}
	got, ok := obj["answer"]
	if !ok {
		t.Fatal("expected exported alias answer in module object")
	}
	if got != float64(42) {
		t.Fatalf("module export answer = %#v, want 42", got)
	}
}

func TestModuleLoaderResolveExportDefault(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "mod.ia")
	writeTestModule(t, modulePath, `
let value = 10;
export default value * 2;
`)

	loader := NewModuleLoader(nil, nil)
	fromPath := filepath.Join(dir, "main.ia")

	loaded, err := loader.Resolve(fromPath, "./mod")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	obj, ok := loaded.(Object)
	if !ok {
		t.Fatalf("Resolve() value type = %T, want Object", loaded)
	}
	got, ok := obj["default"]
	if !ok {
		t.Fatal("expected default export in module object")
	}
	if got != float64(20) {
		t.Fatalf("module default export = %#v, want 20", got)
	}
}

func TestModuleLoaderResolveExportNamedClass(t *testing.T) {
	dir := t.TempDir()
	writeTestModule(t, filepath.Join(dir, "mod.ia"), `
export class Counter {
  constructor() { this.value = 7; }
}
`)
	writeTestModule(t, filepath.Join(dir, "main.ia"), `
import { Counter } from "./mod";
let c = new Counter();
export let value = c.value;
`)

	loader := NewModuleLoader(nil, nil)
	loaded, err := loader.Resolve(filepath.Join(dir, "entry.ia"), "./main")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	obj, ok := loaded.(Object)
	if !ok {
		t.Fatalf("Resolve() value type = %T, want Object", loaded)
	}
	if obj["value"] != float64(7) {
		t.Fatalf("module value = %#v, want 7", obj["value"])
	}
}

func TestModuleLoaderResolveImportNamespace(t *testing.T) {
	dir := t.TempDir()
	writeTestModule(t, filepath.Join(dir, "dep.ia"), `
export let n = 5;
`)
	writeTestModule(t, filepath.Join(dir, "main.ia"), `
import * as dep from "./dep";
export let value = dep.n + 1;
`)

	loader := NewModuleLoader(nil, nil)
	loaded, err := loader.Resolve(filepath.Join(dir, "entry.ia"), "./main")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	obj, ok := loaded.(Object)
	if !ok {
		t.Fatalf("Resolve() value type = %T, want Object", loaded)
	}
	got, ok := obj["value"]
	if !ok {
		t.Fatal("expected exported value in module object")
	}
	if got != float64(6) {
		t.Fatalf("module value = %#v, want 6", got)
	}
}

func TestModuleLoaderResolveExportAll(t *testing.T) {
	dir := t.TempDir()
	writeTestModule(t, filepath.Join(dir, "dep.ia"), `
export let a = 1;
export let b = 2;
`)
	writeTestModule(t, filepath.Join(dir, "main.ia"), `
export * from "./dep";
`)

	loader := NewModuleLoader(nil, nil)
	loaded, err := loader.Resolve(filepath.Join(dir, "entry.ia"), "./main")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	obj, ok := loaded.(Object)
	if !ok {
		t.Fatalf("Resolve() value type = %T, want Object", loaded)
	}
	if obj["a"] != float64(1) || obj["b"] != float64(2) {
		t.Fatalf("export-all result = %#v, want a=1,b=2", obj)
	}
}

func TestModuleLoaderResolveDynamicImportExpression(t *testing.T) {
	dir := t.TempDir()
	writeTestModule(t, filepath.Join(dir, "dep.ia"), `
export let n = 41;
`)
	writeTestModule(t, filepath.Join(dir, "main.ia"), `
let dep = await import("./dep");
export let value = dep.n + 1;
`)

	loader := NewModuleLoader(nil, nil)
	loaded, err := loader.Resolve(filepath.Join(dir, "entry.ia"), "./main")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	obj, ok := loaded.(Object)
	if !ok {
		t.Fatalf("Resolve() value type = %T, want Object", loaded)
	}
	if obj["value"] != float64(42) {
		t.Fatalf("dynamic-import result = %#v, want 42", obj["value"])
	}
}

func TestModuleLoaderResolveProjectAbsoluteImport(t *testing.T) {
	dir := t.TempDir()
	writeTestModule(t, filepath.Join(dir, "pkg.toml"), "entry = \"main.ia\"\n")
	if err := os.MkdirAll(filepath.Join(dir, "modules"), 0o755); err != nil {
		t.Fatalf("mkdir modules error: %v", err)
	}
	writeTestModule(t, filepath.Join(dir, "modules", "tools.ia"), "export let answer = 42;\n")

	resolverOptions, err := DiscoverModuleResolverOptions(filepath.Join(dir, "main.ia"))
	if err != nil {
		t.Fatalf("DiscoverModuleResolverOptions unexpected error: %v", err)
	}
	loader := NewModuleLoaderWithResolverOptions(nil, nil, VMOptions{}, resolverOptions)
	loaded, err := loader.Resolve(filepath.Join(dir, "entry.ia"), "/modules/tools")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	obj, ok := loaded.(Object)
	if !ok {
		t.Fatalf("Resolve() value type = %T, want Object", loaded)
	}
	if obj["answer"] != float64(42) {
		t.Fatalf("module answer = %#v, want 42", obj["answer"])
	}
}

func TestModuleLoaderResolveRootAliasImport(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "shared"), 0o755); err != nil {
		t.Fatalf("mkdir shared error: %v", err)
	}
	writeTestModule(t, filepath.Join(dir, "pkg.toml"), "entry = \"main.ia\"\n\n[imports]\nroot_alias = \"@\"\n")
	writeTestModule(t, filepath.Join(dir, "shared", "math.ia"), "export let answer = 42;\n")

	resolverOptions, err := DiscoverModuleResolverOptions(filepath.Join(dir, "main.ia"))
	if err != nil {
		t.Fatalf("DiscoverModuleResolverOptions unexpected error: %v", err)
	}
	loader := NewModuleLoaderWithResolverOptions(nil, nil, VMOptions{}, resolverOptions)
	loaded, err := loader.Resolve(filepath.Join(dir, "main.ia"), "@/shared/math")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	obj, ok := loaded.(Object)
	if !ok {
		t.Fatalf("Resolve() value type = %T, want Object", loaded)
	}
	if obj["answer"] != float64(42) {
		t.Fatalf("module answer = %#v, want 42", obj["answer"])
	}
}

func TestModuleLoaderResolveAliasImport(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "modules"), 0o755); err != nil {
		t.Fatalf("mkdir modules error: %v", err)
	}
	writeTestModule(t, filepath.Join(dir, "pkg.toml"), "entry = \"main.ia\"\n\n[imports.aliases]\n\"@\" = \"modules\"\n")
	writeTestModule(t, filepath.Join(dir, "modules", "math.ia"), "export let answer = 42;\n")

	resolverOptions, err := DiscoverModuleResolverOptions(filepath.Join(dir, "main.ia"))
	if err != nil {
		t.Fatalf("DiscoverModuleResolverOptions unexpected error: %v", err)
	}
	loader := NewModuleLoaderWithResolverOptions(nil, nil, VMOptions{}, resolverOptions)
	loaded, err := loader.Resolve(filepath.Join(dir, "main.ia"), "@/math")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	obj, ok := loaded.(Object)
	if !ok {
		t.Fatalf("Resolve() value type = %T, want Object", loaded)
	}
	if obj["answer"] != float64(42) {
		t.Fatalf("module answer = %#v, want 42", obj["answer"])
	}
}

func TestModuleLoaderResolveRootAliasAndNamedAliasesTogether(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "shared"), 0o755); err != nil {
		t.Fatalf("mkdir shared error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "modules"), 0o755); err != nil {
		t.Fatalf("mkdir modules error: %v", err)
	}
	writeTestModule(t, filepath.Join(dir, "pkg.toml"), "entry = \"main.ia\"\n\n[imports]\nroot_alias = \"@\"\n\n[imports.aliases]\n\"#lib\" = \"modules\"\n")
	writeTestModule(t, filepath.Join(dir, "shared", "config.ia"), "export let kind = \"root\";\n")
	writeTestModule(t, filepath.Join(dir, "modules", "math.ia"), "export let kind = \"alias\";\n")

	resolverOptions, err := DiscoverModuleResolverOptions(filepath.Join(dir, "main.ia"))
	if err != nil {
		t.Fatalf("DiscoverModuleResolverOptions unexpected error: %v", err)
	}
	loader := NewModuleLoaderWithResolverOptions(nil, nil, VMOptions{}, resolverOptions)

	rootLoaded, err := loader.Resolve(filepath.Join(dir, "main.ia"), "@/shared/config")
	if err != nil {
		t.Fatalf("Resolve() root alias unexpected error: %v", err)
	}
	rootObj, ok := rootLoaded.(Object)
	if !ok {
		t.Fatalf("Resolve() root alias value type = %T, want Object", rootLoaded)
	}
	if rootObj["kind"] != "root" {
		t.Fatalf("root alias module kind = %#v, want %q", rootObj["kind"], "root")
	}

	aliasLoaded, err := loader.Resolve(filepath.Join(dir, "main.ia"), "#lib/math")
	if err != nil {
		t.Fatalf("Resolve() named alias unexpected error: %v", err)
	}
	aliasObj, ok := aliasLoaded.(Object)
	if !ok {
		t.Fatalf("Resolve() named alias value type = %T, want Object", aliasLoaded)
	}
	if aliasObj["kind"] != "alias" {
		t.Fatalf("named alias module kind = %#v, want %q", aliasObj["kind"], "alias")
	}
}

func writeTestModule(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write module %s error: %v", path, err)
	}
}
