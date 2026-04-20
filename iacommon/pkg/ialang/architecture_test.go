package ialang_test

import (
	"go/parser"
	gotoken "go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedPackageBoundaries(t *testing.T) {
	type rule struct {
		pkgDir    string
		forbidden []string
	}

	rules := []rule{
		{
			pkgDir: "pkg/ialang/value",
			forbidden: []string{
				"iacommon/pkg/ialang/ast",
				"iacommon/pkg/ialang/bytecode",
				"iacommon/pkg/ialang/chunkcodec",
				"iacommon/pkg/ialang/packagefile",
				"iacommon/pkg/ialang/token",
				"ialang/pkg/lang",
			},
		},
		{
			pkgDir: "pkg/ialang/token",
			forbidden: []string{
				"iacommon/pkg/ialang/ast",
				"iacommon/pkg/ialang/bytecode",
				"iacommon/pkg/ialang/chunkcodec",
				"iacommon/pkg/ialang/packagefile",
				"ialang/pkg/lang",
			},
		},
		{
			pkgDir: "pkg/ialang/ast",
			forbidden: []string{
				"iacommon/pkg/ialang/bytecode",
				"iacommon/pkg/ialang/chunkcodec",
				"iacommon/pkg/ialang/packagefile",
				"ialang/pkg/lang",
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
		fset := gotoken.NewFileSet()
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
	if forbidden == "ialang/pkg/lang" {
		return false
	}
	return strings.HasPrefix(importPath, forbidden+"/")
}
