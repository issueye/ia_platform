package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteInitCommandCreatesProjectTemplate(t *testing.T) {
	restore := stubGitInit(t)
	defer restore()

	target := filepath.Join(t.TempDir(), "demo-app")
	var stderr bytes.Buffer
	if err := executeInitCommand(target, &stderr); err != nil {
		t.Fatalf("executeInitCommand unexpected error: %v", err)
	}

	mustExistFile(t, filepath.Join(target, "main.ia"))
	mustExistFile(t, filepath.Join(target, "config", "app.json"))
	mustExistFile(t, filepath.Join(target, "modules", "utils", "index.ia"))
	mustExistFile(t, filepath.Join(target, "modules", "pkg", "index.ia"))
	mustExistFile(t, filepath.Join(target, "README.md"))
	mustExistFile(t, filepath.Join(target, "pkg.toml"))
	mustExistFile(t, filepath.Join(target, ".gitignore"))
	mustExistDir(t, filepath.Join(target, ".git"))

	mainData, err := os.ReadFile(filepath.Join(target, "main.ia"))
	if err != nil {
		t.Fatalf("read main.ia error: %v", err)
	}
	if !strings.Contains(string(mainData), `import { greet } from "@/modules/utils/index";`) {
		t.Fatalf("main.ia content unexpected: %q", string(mainData))
	}
	if !strings.Contains(string(mainData), `import { pkgName, pkgVersion } from "@/modules/pkg/index";`) {
		t.Fatalf("main.ia should use root alias imports: %q", string(mainData))
	}
	if !strings.Contains(string(mainData), `import { fromFile } from "json";`) {
		t.Fatalf("main.ia should load config file: %q", string(mainData))
	}
	if !strings.Contains(string(mainData), `import { setLevel, setJSON, setOutputPath, info, warn } from "log";`) {
		t.Fatalf("main.ia should configure logging from config: %q", string(mainData))
	}
	if !strings.Contains(string(mainData), `function main()`) {
		t.Fatalf("main.ia missing main function: %q", string(mainData))
	}
	if strings.Contains(string(mainData), `main();`) {
		t.Fatalf("main.ia should not include explicit main() call: %q", string(mainData))
	}

	pkgData, err := os.ReadFile(filepath.Join(target, "pkg.toml"))
	if err != nil {
		t.Fatalf("read pkg.toml error: %v", err)
	}
	if !strings.Contains(string(pkgData), `entry = "main.ia"`) {
		t.Fatalf("pkg.toml content unexpected: %q", string(pkgData))
	}
	if !strings.Contains(string(pkgData), `root_alias = "@"`) {
		t.Fatalf("pkg.toml should configure root alias: %q", string(pkgData))
	}

	cfgData, err := os.ReadFile(filepath.Join(target, "config", "app.json"))
	if err != nil {
		t.Fatalf("read config/app.json error: %v", err)
	}
	if !strings.Contains(string(cfgData), `"level": "info"`) {
		t.Fatalf("config/app.json should include default log level: %q", string(cfgData))
	}
	if !strings.Contains(string(cfgData), `"outputPath": ""`) {
		t.Fatalf("config/app.json should include default log outputPath: %q", string(cfgData))
	}
}

func TestExecuteInitCommandExistingFileError(t *testing.T) {
	restore := stubGitInit(t)
	defer restore()

	target := filepath.Join(t.TempDir(), "demo-app")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "main.ia"), []byte("let x = 1;"), 0o644); err != nil {
		t.Fatalf("write main.ia error: %v", err)
	}

	var stderr bytes.Buffer
	err := executeInitCommand(target, &stderr)
	if err == nil {
		t.Fatal("executeInitCommand expected error, got nil")
	}
	if !strings.Contains(err.Error(), "file already exists:") {
		t.Fatalf("executeInitCommand error = %v, want existing file error", err)
	}
}

func TestRunCLIInitCommand(t *testing.T) {
	restore := stubGitInit(t)
	defer restore()

	target := filepath.Join(t.TempDir(), "demo-app")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI([]string{"ialang", "init", target}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI init code = %d, want 0, stderr=%q", code, stderr.String())
	}
	mustExistFile(t, filepath.Join(target, "README.md"))
	mustExistDir(t, filepath.Join(target, ".git"))
}

func stubGitInit(t *testing.T) func() {
	t.Helper()
	original := runGitInit
	runGitInit = func(dir string) error {
		return os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	}
	return func() {
		runGitInit = original
	}
}

func mustExistFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat file error (%s): %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("path is directory, want file: %s", path)
	}
}

func mustExistDir(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat dir error (%s): %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("path is file, want directory: %s", path)
	}
}
