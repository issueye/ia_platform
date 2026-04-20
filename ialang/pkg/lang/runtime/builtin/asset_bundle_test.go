package builtin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

func TestAssetBundleMinifyHashAndVersion(t *testing.T) {
	mod := mustModuleObject(t, DefaultModules(rt.NewGoroutineRuntime()), "asset")

	css := `/* comment */
body {
  color: red;
  margin: 0;
}`
	minifiedCSS := callNative(t, mod, "minify", css, "css").(Object)
	if minifiedCSS["content"] != "body{color:red;margin:0;}" {
		t.Fatalf("unexpected minified css: %q", minifiedCSS["content"])
	}
	if minifiedCSS["original"].(float64) <= minifiedCSS["minified"].(float64) {
		t.Fatalf("expected minified css to be shorter: %#v", minifiedCSS)
	}

	js := "// remove me\nconst url = 'https://example.test'; /* block */\nconst x = 1;"
	minifiedJS := callNative(t, mod, "minify", js, "javascript").(Object)
	jsContent := minifiedJS["content"].(string)
	if strings.Contains(jsContent, "remove me") || strings.Contains(jsContent, "block") {
		t.Fatalf("expected js comments removed, got %q", jsContent)
	}
	if !strings.Contains(jsContent, "https://example.test") {
		t.Fatalf("expected URL-like string to be retained, got %q", jsContent)
	}

	html := "<html> <!-- comment --> <body> hi </body> </html>"
	minifiedHTML := callNative(t, mod, "minify", html, "html").(Object)
	if strings.Contains(minifiedHTML["content"].(string), "comment") {
		t.Fatalf("expected html comment removed, got %q", minifiedHTML["content"])
	}

	hash := callNative(t, mod, "hash", "content", float64(12)).(Object)
	if len(hash["hash"].(string)) != 12 {
		t.Fatalf("expected 12-char hash, got %#v", hash)
	}
	if hash["length"] != float64(12) {
		t.Fatalf("expected length 12, got %#v", hash["length"])
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "app.css")
	if err := os.WriteFile(file, []byte(css), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	version := callNative(t, mod, "version", file).(Object)
	if version["path"] != file {
		t.Fatalf("expected version path %q, got %#v", file, version["path"])
	}
	if !strings.HasPrefix(version["versioned"].(string), "app.") || !strings.HasSuffix(version["versioned"].(string), ".css") {
		t.Fatalf("unexpected versioned filename: %#v", version["versioned"])
	}
}

func TestAssetBundleBundleAnalyzeAndClean(t *testing.T) {
	mod := mustModuleObject(t, DefaultModules(rt.NewGoroutineRuntime()), "asset")

	srcDir := t.TempDir()
	outDir := filepath.Join(t.TempDir(), "dist")
	cssFile := filepath.Join(srcDir, "style.css")
	jsFile := filepath.Join(srcDir, "app.js")
	if err := os.WriteFile(cssFile, []byte("body { color: blue; }"), 0644); err != nil {
		t.Fatalf("write css fixture: %v", err)
	}
	if err := os.WriteFile(jsFile, []byte("// c\nconst x = 1;"), 0644); err != nil {
		t.Fatalf("write js fixture: %v", err)
	}

	bundle := callNative(t, mod, "bundle", outDir, Object{
		"files":  Array{cssFile, jsFile},
		"minify": true,
		"hash":   false,
		"concat": true,
	}).(Object)
	if bundle["count"] != float64(2) {
		t.Fatalf("expected 2 bundled files, got %#v", bundle["count"])
	}
	files := bundle["files"].(Array)
	if len(files) != 2 {
		t.Fatalf("expected processed files length 2, got %d", len(files))
	}
	manifest := bundle["manifest"].(Object)
	if _, ok := manifest[cssFile]; !ok {
		t.Fatalf("expected css manifest entry, got %#v", manifest)
	}
	if _, err := os.Stat(filepath.Join(outDir, "style.css")); err != nil {
		t.Fatalf("expected output css file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "bundle.css")); err != nil {
		t.Fatalf("expected concat output file: %v", err)
	}

	analysis := callNative(t, mod, "analyze", outDir).(Object)
	if analysis["fileCount"].(float64) < 2 {
		t.Fatalf("expected analyzed file count >= 2, got %#v", analysis["fileCount"])
	}
	byExt := analysis["byExtension"].(Object)
	if _, ok := byExt[".css"]; !ok {
		t.Fatalf("expected .css extension stats, got %#v", byExt)
	}

	clean := callNative(t, mod, "clean", outDir, "*.css").(Object)
	if clean["count"].(float64) < 1 {
		t.Fatalf("expected css files removed, got %#v", clean)
	}
}

func TestAssetBundleErrors(t *testing.T) {
	mod := mustModuleObject(t, DefaultModules(rt.NewGoroutineRuntime()), "asset")

	if _, err := callNativeWithError(mod, "minify", "x"); err == nil {
		t.Fatal("expected minify arity error")
	}
	if _, err := callNativeWithError(mod, "minify", "x", "unknown"); err == nil {
		t.Fatal("expected unsupported file type error")
	}
	if _, err := callNativeWithError(mod, "hash"); err == nil {
		t.Fatal("expected hash arity error")
	}
	if _, err := callNativeWithError(mod, "version", filepath.Join(t.TempDir(), "missing.css")); err == nil {
		t.Fatal("expected version missing file error")
	}
	if _, err := callNativeWithError(mod, "bundle", t.TempDir(), "not-options"); err == nil {
		t.Fatal("expected bundle options type error")
	}
	if _, err := callNativeWithError(mod, "analyze", filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected analyze missing directory error")
	}
}
