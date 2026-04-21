package module

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverModuleResolverOptions(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "modules"), 0o755); err != nil {
		t.Fatalf("mkdir modules error: %v", err)
	}
	writeProjectFile(t, filepath.Join(dir, ProjectConfigFileName()), "entry = \"main.ia\"\n\n[imports]\nroot_alias = \"@\"\n\n[imports.aliases]\n\"#lib\" = \"modules\"\n")

	opts, err := DiscoverModuleResolverOptions(filepath.Join(dir, "src", "main.ia"))
	if err != nil {
		t.Fatalf("DiscoverModuleResolverOptions unexpected error: %v", err)
	}
	if opts.ProjectRoot != dir {
		t.Fatalf("ProjectRoot = %q, want %q", opts.ProjectRoot, dir)
	}
	if opts.RootAlias != "@" {
		t.Fatalf("RootAlias = %q, want %q", opts.RootAlias, "@")
	}
	if got := opts.Aliases["#lib"]; got != "modules" {
		t.Fatalf("Aliases[#lib] = %q, want %q", got, "modules")
	}
}

func TestReadProjectConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ProjectConfigFileName())
	writeProjectFile(t, cfgPath, "entry = \"app/main.ia\"\n\n[imports]\nroot_alias = \"@\"\n\n[imports.aliases]\n\"#lib\" = \"modules\"\n")

	cfg, err := ReadProjectConfig(cfgPath)
	if err != nil {
		t.Fatalf("ReadProjectConfig unexpected error: %v", err)
	}
	if cfg.Entry != "app/main.ia" {
		t.Fatalf("Entry = %q, want %q", cfg.Entry, "app/main.ia")
	}
	if cfg.Imports.RootAlias != "@" {
		t.Fatalf("Imports.RootAlias = %q, want %q", cfg.Imports.RootAlias, "@")
	}
	if got := cfg.Imports.Aliases["#lib"]; got != "modules" {
		t.Fatalf("Imports.Aliases[#lib] = %q, want %q", got, "modules")
	}
}

func TestResolveModulePathWithOptionsProjectAbsolute(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveModulePathWithOptions(filepath.Join(dir, "main.ia"), "/modules/tools", ModuleResolverOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("ResolveModulePathWithOptions unexpected error: %v", err)
	}
	want := filepath.Join(dir, "modules", "tools.ia")
	assertSamePath(t, got, want)
}

func TestResolveModulePathWithOptionsRootAlias(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveModulePathWithOptions(filepath.Join(dir, "main.ia"), "@/shared/math", ModuleResolverOptions{ProjectRoot: dir, RootAlias: "@"})
	if err != nil {
		t.Fatalf("ResolveModulePathWithOptions unexpected error: %v", err)
	}
	want := filepath.Join(dir, "shared", "math.ia")
	assertSamePath(t, got, want)
}

func TestResolveModulePathWithOptionsNamedAlias(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveModulePathWithOptions(filepath.Join(dir, "main.ia"), "#lib/math", ModuleResolverOptions{ProjectRoot: dir, Aliases: map[string]string{"#lib": "modules"}})
	if err != nil {
		t.Fatalf("ResolveModulePathWithOptions unexpected error: %v", err)
	}
	want := filepath.Join(dir, "modules", "math.ia")
	assertSamePath(t, got, want)
}

func TestResolveModulePathWithOptionsAliasPrecedence(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveModulePathWithOptions(filepath.Join(dir, "main.ia"), "#lib/sub/tool", ModuleResolverOptions{
		ProjectRoot: dir,
		Aliases: map[string]string{
			"#lib":     "modules/root",
			"#lib/sub": "modules/sub",
		},
	})
	if err != nil {
		t.Fatalf("ResolveModulePathWithOptions unexpected error: %v", err)
	}
	want := filepath.Join(dir, "modules", "sub", "tool.ia")
	assertSamePath(t, got, want)
}

func TestResolveModulePathWithOptionsRelativeImport(t *testing.T) {
	dir := t.TempDir()
	from := filepath.Join(dir, "src", "main.ia")
	got, err := ResolveModulePathWithOptions(from, "./utils/math", ModuleResolverOptions{})
	if err != nil {
		t.Fatalf("ResolveModulePathWithOptions unexpected error: %v", err)
	}
	want := filepath.Join(dir, "src", "utils", "math.ia")
	assertSamePath(t, got, want)
}

func TestResolveModulePathWithOptionsWithoutProjectRootNamedAliasFails(t *testing.T) {
	_, err := ResolveModulePathWithOptions("main.ia", "#lib/math", ModuleResolverOptions{Aliases: map[string]string{"#lib": "modules"}})
	if err == nil {
		t.Fatal("ResolveModulePathWithOptions expected error, got nil")
	}
}

func writeProjectFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s error: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s error: %v", path, err)
	}
}

func assertSamePath(t *testing.T, got, want string) {
	t.Helper()
	gotClean := filepath.Clean(got)
	wantClean := filepath.Clean(want)
	if gotClean != wantClean {
		t.Fatalf("path = %q, want %q", gotClean, wantClean)
	}
}
