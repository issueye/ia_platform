package lang

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageBoundaries(t *testing.T) {
	type rule struct {
		pkgDir    string
		forbidden []string
	}

	rules := []rule{
		{
			pkgDir: "runtime",
			forbidden: []string{
				"ialang/pkg/lang",
				"ialang/pkg/lang/compiler",
				"ialang/pkg/lang/frontend",
			},
		},
		{
			pkgDir: "compiler",
			forbidden: []string{
				"ialang/pkg/lang",
				"ialang/pkg/lang/frontend",
			},
		},
		{
			pkgDir: "frontend",
			forbidden: []string{
				"ialang/pkg/lang",
				"ialang/pkg/lang/compiler",
				"ialang/pkg/lang/runtime",
			},
		},
		{
			pkgDir: "ast",
			forbidden: []string{
				"ialang/pkg/lang",
				"ialang/pkg/lang/compiler",
				"ialang/pkg/lang/frontend",
				"ialang/pkg/lang/runtime",
			},
		},
		{
			pkgDir: "bytecode",
			forbidden: []string{
				"ialang/pkg/lang",
				"ialang/pkg/lang/ast",
				"ialang/pkg/lang/compiler",
				"ialang/pkg/lang/frontend",
				"ialang/pkg/lang/runtime",
				"ialang/pkg/lang/token",
			},
		},
		{
			pkgDir: "token",
			forbidden: []string{
				"ialang/pkg/lang",
				"ialang/pkg/lang/ast",
				"ialang/pkg/lang/bytecode",
				"ialang/pkg/lang/compiler",
				"ialang/pkg/lang/frontend",
				"ialang/pkg/lang/runtime",
			},
		},
	}

	for _, r := range rules {
		checkPackageImportRule(t, r.pkgDir, r.forbidden)
	}
}

func checkPackageImportRule(t *testing.T, pkgDir string, forbidden []string) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
	if err != nil {
		t.Fatalf("glob error for %s: %v", pkgDir, err)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse imports error %s: %v", path, err)
		}

		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			for _, p := range forbidden {
				if isForbiddenImport(importPath, p) {
					t.Fatalf("forbidden import in %s: %s", path, importPath)
				}
			}
		}
	}
}

func isForbiddenImport(importPath, forbidden string) bool {
	if importPath == forbidden {
		return true
	}
	// "ialang/pkg/lang" means facade package only, not all subpackages.
	if forbidden == "ialang/pkg/lang" {
		return false
	}
	return strings.HasPrefix(importPath, forbidden+"/")
}
