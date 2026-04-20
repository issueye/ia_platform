package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	bc "iacommon/pkg/ialang/bytecode"
	"ialang/pkg/lang"
)

const defaultProjectEntry = "main.ia"

func executeCheckCommand(target string, stdout, stderr io.Writer) error {
	entryPath, err := resolveCheckEntryPath(target)
	if err != nil {
		return err
	}
	pkg, err := buildPackageForCheck(entryPath, stderr)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "syntax check passed: %d module(s), entry=%s\n", len(pkg.Modules), pkg.Entry)
	return nil
}

func buildPackageForCheck(entryPath string, stderr io.Writer) (*checkPackageSummary, error) {
	pkg, err := buildPackage(entryPath, stderr)
	if err != nil {
		return nil, err
	}
	return &checkPackageSummary{
		Entry:   pkg.Entry,
		Modules: pkg.Modules,
		Imports: pkg.Imports,
	}, nil
}

type checkPackageSummary struct {
	Entry   string
	Modules map[string]*bc.Chunk
	Imports map[string]map[string]string
}

func resolveCheckEntryPath(target string) (string, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve current directory error: %w", err)
		}
		return resolveEntryFromProjectDir(cwd)
	}

	info, err := os.Stat(trimmed)
	if err == nil && info.IsDir() {
		return resolveEntryFromProjectDir(trimmed)
	}
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("stat target error: %w", err)
	}
	abs, err := filepath.Abs(trimmed)
	if err != nil {
		return "", fmt.Errorf("resolve entry path error: %w", err)
	}
	return abs, nil
}

func resolveEntryFromProjectDir(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve project directory error: %w", err)
	}

	pkgTomlPath := filepath.Join(absDir, lang.ProjectConfigFileName())
	if fileExists(pkgTomlPath) {
		cfg, err := lang.ReadProjectConfig(pkgTomlPath)
		if err != nil {
			return "", err
		}
		entry := strings.TrimSpace(cfg.Entry)
		if entry != "" {
			if filepath.IsAbs(entry) {
				return entry, nil
			}
			return filepath.Join(absDir, filepath.FromSlash(entry)), nil
		}
	}

	fallback := filepath.Join(absDir, defaultProjectEntry)
	if fileExists(fallback) {
		return fallback, nil
	}
	return "", fmt.Errorf("cannot determine project entry in %s: define entry in %s or provide entry file path", absDir, lang.ProjectConfigFileName())
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
