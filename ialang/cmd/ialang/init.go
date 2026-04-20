package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type initTemplateFile struct {
	path    string
	content string
}

type initTemplateSpec struct {
	source string
	target string
}

type initTemplateData struct {
	ProjectName string
}

//go:embed templates/*
var initTemplatesFS embed.FS

var initTemplateSpecs = []initTemplateSpec{
	{source: "templates/main.ia.tmpl", target: "main.ia"},
	{source: "templates/app.config.json.tmpl", target: "config/app.json"},
	{source: "templates/modules_utils_index.ia.tmpl", target: "modules/utils/index.ia"},
	{source: "templates/modules_pkg_index.ia.tmpl", target: "modules/pkg/index.ia"},
	{source: "templates/README.md.tmpl", target: "README.md"},
	{source: "templates/pkg.toml.tmpl", target: "pkg.toml"},
	{source: "templates/gitignore.tmpl", target: ".gitignore"},
}

var runGitInit = func(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg != "" {
			return fmt.Errorf("git init error: %w: %s", err, msg)
		}
		return fmt.Errorf("git init error: %w", err)
	}
	return nil
}

func executeInitCommand(targetDir string, _ io.Writer) error {
	targetDir = strings.TrimSpace(targetDir)
	if targetDir == "" {
		targetDir = "."
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create target directory error: %w", err)
	}

	projectName := inferProjectName(targetDir)
	files, err := buildInitTemplateFiles(projectName)
	if err != nil {
		return err
	}
	if err := ensureInitFilesDoNotExist(targetDir, files); err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(targetDir, "config"),
		filepath.Join(targetDir, "modules"),
		filepath.Join(targetDir, "modules", "utils"),
		filepath.Join(targetDir, "modules", "pkg"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory error (%s): %w", dir, err)
		}
	}

	for _, file := range files {
		fullPath := filepath.Join(targetDir, filepath.FromSlash(file.path))
		if err := os.WriteFile(fullPath, []byte(file.content), 0o644); err != nil {
			return fmt.Errorf("write file error (%s): %w", fullPath, err)
		}
	}

	if err := ensureGitRepository(targetDir); err != nil {
		return err
	}
	return nil
}

func inferProjectName(targetDir string) string {
	name := filepath.Base(filepath.Clean(targetDir))
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "ialang-app"
	}
	name = strings.ReplaceAll(name, " ", "-")
	return name
}

func buildInitTemplateFiles(projectName string) ([]initTemplateFile, error) {
	data := initTemplateData{ProjectName: projectName}
	files := make([]initTemplateFile, 0, len(initTemplateSpecs))
	for _, spec := range initTemplateSpecs {
		raw, err := initTemplatesFS.ReadFile(spec.source)
		if err != nil {
			return nil, fmt.Errorf("read init template error (%s): %w", spec.source, err)
		}
		rendered, err := renderInitTemplate(spec.source, string(raw), data)
		if err != nil {
			return nil, err
		}
		files = append(files, initTemplateFile{
			path:    spec.target,
			content: rendered,
		})
	}
	return files, nil
}

func renderInitTemplate(sourcePath, source string, data initTemplateData) (string, error) {
	tmpl, err := template.New(sourcePath).Option("missingkey=error").Parse(source)
	if err != nil {
		return "", fmt.Errorf("parse init template error (%s): %w", sourcePath, err)
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("render init template error (%s): %w", sourcePath, err)
	}
	return out.String(), nil
}

func ensureInitFilesDoNotExist(targetDir string, files []initTemplateFile) error {
	for _, file := range files {
		fullPath := filepath.Join(targetDir, filepath.FromSlash(file.path))
		info, err := os.Stat(fullPath)
		if err == nil {
			if info.IsDir() {
				return fmt.Errorf("path already exists as directory: %s", fullPath)
			}
			return fmt.Errorf("file already exists: %s", fullPath)
		}
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("check target file error (%s): %w", fullPath, err)
		}
	}
	return nil
}

func ensureGitRepository(targetDir string) error {
	gitDir := filepath.Join(targetDir, ".git")
	info, err := os.Stat(gitDir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf(".git exists and is not a directory: %s", gitDir)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("check .git error (%s): %w", gitDir, err)
	}
	if err := runGitInit(targetDir); err != nil {
		return err
	}
	return nil
}
