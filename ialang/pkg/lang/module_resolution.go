package lang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const projectConfigFileName = "pkg.toml"

func ProjectConfigFileName() string {
	return projectConfigFileName
}

type ModuleResolverOptions struct {
	ProjectRoot string
	RootAlias   string
	Aliases     map[string]string
}

type ProjectConfig struct {
	Entry   string `toml:"entry"`
	Imports struct {
		RootAlias string            `toml:"root_alias"`
		Aliases   map[string]string `toml:"aliases"`
	} `toml:"imports"`
}

func DiscoverModuleResolverOptions(startPath string) (ModuleResolverOptions, error) {
	searchDir, err := resolverSearchDir(startPath)
	if err != nil {
		return ModuleResolverOptions{}, err
	}
	if searchDir == "" {
		return ModuleResolverOptions{}, nil
	}

	root, cfgPath, found, err := findProjectConfig(searchDir)
	if err != nil {
		return ModuleResolverOptions{}, err
	}
	if !found {
		return ModuleResolverOptions{}, nil
	}

	cfg, err := ReadProjectConfig(cfgPath)
	if err != nil {
		return ModuleResolverOptions{}, err
	}
	return ModuleResolverOptions{
		ProjectRoot: root,
		RootAlias:   strings.TrimSpace(cfg.Imports.RootAlias),
		Aliases:     cloneStringMap(cfg.Imports.Aliases),
	}, nil
}

func ReadProjectConfig(path string) (*ProjectConfig, error) {
	var cfg ProjectConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("read pkg.toml error (%s): %w", path, err)
	}
	return &cfg, nil
}

func ResolveModulePathWithOptions(fromPath, moduleName string, opts ModuleResolverOptions) (string, error) {
	fullPath, err := moduleImportPath(fromPath, moduleName, opts)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("resolve module path error (%s): %w", moduleName, err)
	}
	return abs, nil
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func resolverSearchDir(startPath string) (string, error) {
	trimmed := strings.TrimSpace(startPath)
	if trimmed == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve current directory error: %w", err)
		}
		return cwd, nil
	}
	abs, err := filepath.Abs(trimmed)
	if err != nil {
		return "", fmt.Errorf("resolve path error (%s): %w", trimmed, err)
	}
	info, err := os.Stat(abs)
	if err == nil {
		if info.IsDir() {
			return abs, nil
		}
		return filepath.Dir(abs), nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat path error (%s): %w", abs, err)
	}
	if filepath.Ext(abs) != "" {
		return filepath.Dir(abs), nil
	}
	return abs, nil
}

func findProjectConfig(startDir string) (projectRoot, configPath string, found bool, err error) {
	current := filepath.Clean(startDir)
	for {
		candidate := filepath.Join(current, projectConfigFileName)
		info, statErr := os.Stat(candidate)
		if statErr == nil {
			if info.IsDir() {
				return "", "", false, fmt.Errorf("project config path is directory: %s", candidate)
			}
			return current, candidate, true, nil
		}
		if statErr != nil && !os.IsNotExist(statErr) {
			return "", "", false, fmt.Errorf("stat project config error (%s): %w", candidate, statErr)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", "", false, nil
		}
		current = parent
	}
}

func moduleImportPath(fromPath, moduleName string, opts ModuleResolverOptions) (string, error) {
	if strings.HasPrefix(moduleName, "./") || strings.HasPrefix(moduleName, "../") {
		baseDir := "."
		if fromPath != "" {
			baseDir = filepath.Dir(fromPath)
		}
		return appendModuleExt(filepath.Join(baseDir, moduleName)), nil
	}

	if strings.HasPrefix(moduleName, "/") {
		if strings.TrimSpace(opts.ProjectRoot) != "" {
			trimmed := strings.TrimPrefix(moduleName, "/")
			return appendModuleExt(filepath.Join(opts.ProjectRoot, filepath.FromSlash(trimmed))), nil
		}
	}

	if filepath.IsAbs(moduleName) {
		return appendModuleExt(moduleName), nil
	}

	if target, ok := resolveRootAliasModuleImport(moduleName, opts); ok {
		return appendModuleExt(target), nil
	}

	if target, ok := resolveAliasModuleImport(moduleName, opts); ok {
		return appendModuleExt(target), nil
	}

	return "", fmt.Errorf("module not found: %s", moduleName)
}

func resolveRootAliasModuleImport(moduleName string, opts ModuleResolverOptions) (string, bool) {
	alias := strings.TrimSpace(opts.RootAlias)
	projectRoot := strings.TrimSpace(opts.ProjectRoot)
	if alias == "" || projectRoot == "" {
		return "", false
	}
	if moduleName != alias && !strings.HasPrefix(moduleName, alias+"/") {
		return "", false
	}

	suffix := strings.TrimPrefix(moduleName, alias)
	trimmedSuffix := strings.TrimPrefix(suffix, "/")
	if trimmedSuffix == "" {
		return projectRoot, true
	}
	return filepath.Join(projectRoot, filepath.FromSlash(trimmedSuffix)), true
}

func resolveAliasModuleImport(moduleName string, opts ModuleResolverOptions) (string, bool) {
	if len(opts.Aliases) == 0 {
		return "", false
	}
	alias, target, ok := matchModuleAlias(moduleName, opts.Aliases)
	if !ok {
		return "", false
	}

	suffix := strings.TrimPrefix(moduleName, alias)
	base := target
	if !filepath.IsAbs(base) {
		if strings.TrimSpace(opts.ProjectRoot) == "" {
			return "", false
		}
		base = filepath.Join(opts.ProjectRoot, filepath.FromSlash(base))
	}
	if suffix == "" {
		return base, true
	}
	trimmedSuffix := strings.TrimPrefix(suffix, "/")
	return filepath.Join(base, filepath.FromSlash(trimmedSuffix)), true
}

func matchModuleAlias(moduleName string, aliases map[string]string) (string, string, bool) {
	bestAlias := ""
	bestTarget := ""
	for alias, target := range aliases {
		if alias == "" {
			continue
		}
		if moduleName != alias && !strings.HasPrefix(moduleName, alias+"/") {
			continue
		}
		if len(alias) <= len(bestAlias) {
			continue
		}
		bestAlias = alias
		bestTarget = target
	}
	if bestAlias == "" {
		return "", "", false
	}
	return bestAlias, bestTarget, true
}

func appendModuleExt(path string) string {
	if filepath.Ext(path) != "" {
		return path
	}
	return path + ".ia"
}
