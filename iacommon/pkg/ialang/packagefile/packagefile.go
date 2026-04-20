package packagefile

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	bc "iacommon/pkg/ialang/bytecode"
	"iacommon/pkg/ialang/chunkcodec"
)

const (
	PackageFormatMagic   = "IALANG_PKG"
	PackageFormatVersion = 1
)

type Package struct {
	Entry   string
	Modules map[string]*bc.Chunk
	Imports map[string]map[string]string
}

type packageEnvelope struct {
	Magic   string                       `json:"magic"`
	Version int                          `json:"version"`
	Entry   string                       `json:"entry"`
	Modules []encodedModule              `json:"modules"`
	Imports map[string]map[string]string `json:"imports,omitempty"`
}

type encodedModule struct {
	Path  string `json:"path"`
	Chunk string `json:"chunk"`
}

func Encode(pkg *Package) ([]byte, error) {
	if pkg == nil {
		return nil, fmt.Errorf("package is nil")
	}
	if pkg.Entry == "" {
		return nil, fmt.Errorf("package entry is empty")
	}
	if len(pkg.Modules) == 0 {
		return nil, fmt.Errorf("package modules is empty")
	}
	if _, ok := pkg.Modules[pkg.Entry]; !ok {
		return nil, fmt.Errorf("package entry module not found: %s", pkg.Entry)
	}

	modulePaths := make([]string, 0, len(pkg.Modules))
	for path := range pkg.Modules {
		modulePaths = append(modulePaths, path)
	}
	sort.Strings(modulePaths)

	modules := make([]encodedModule, 0, len(modulePaths))
	for _, path := range modulePaths {
		chunk := pkg.Modules[path]
		if chunk == nil {
			return nil, fmt.Errorf("module chunk is nil: %s", path)
		}
		chunkBytes, err := chunkcodec.Serialize(chunk)
		if err != nil {
			return nil, fmt.Errorf("serialize module %s error: %w", path, err)
		}
		modules = append(modules, encodedModule{
			Path:  path,
			Chunk: base64.StdEncoding.EncodeToString(chunkBytes),
		})
	}

	envelope := packageEnvelope{
		Magic:   PackageFormatMagic,
		Version: PackageFormatVersion,
		Entry:   pkg.Entry,
		Modules: modules,
		Imports: cloneImports(pkg.Imports),
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("encode package envelope error: %w", err)
	}
	return data, nil
}

func Decode(data []byte) (*Package, error) {
	var envelope packageEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode package envelope error: %w", err)
	}
	if envelope.Magic != PackageFormatMagic {
		return nil, fmt.Errorf("invalid package magic: %q", envelope.Magic)
	}
	if envelope.Version != PackageFormatVersion {
		return nil, fmt.Errorf("unsupported package format version: got %d, want %d", envelope.Version, PackageFormatVersion)
	}
	if envelope.Entry == "" {
		return nil, fmt.Errorf("package entry is empty")
	}
	if len(envelope.Modules) == 0 {
		return nil, fmt.Errorf("package modules is empty")
	}

	modules := make(map[string]*bc.Chunk, len(envelope.Modules))
	for _, module := range envelope.Modules {
		if module.Path == "" {
			return nil, fmt.Errorf("module path is empty")
		}
		if _, exists := modules[module.Path]; exists {
			return nil, fmt.Errorf("duplicate module path: %s", module.Path)
		}
		chunkBytes, err := base64.StdEncoding.DecodeString(module.Chunk)
		if err != nil {
			return nil, fmt.Errorf("decode module %s error: %w", module.Path, err)
		}
		chunk, err := chunkcodec.Deserialize(chunkBytes)
		if err != nil {
			return nil, fmt.Errorf("deserialize module %s error: %w", module.Path, err)
		}
		modules[module.Path] = chunk
	}

	if _, ok := modules[envelope.Entry]; !ok {
		return nil, fmt.Errorf("package entry module not found: %s", envelope.Entry)
	}

	imports := cloneImports(envelope.Imports)
	for fromPath, importRules := range imports {
		if _, exists := modules[fromPath]; !exists {
			return nil, fmt.Errorf("import map refers to unknown module: %s", fromPath)
		}
		for importName, targetPath := range importRules {
			if importName == "" {
				return nil, fmt.Errorf("empty import name in module: %s", fromPath)
			}
			if _, exists := modules[targetPath]; !exists {
				return nil, fmt.Errorf("import target not found: %s -> %s", fromPath, targetPath)
			}
		}
	}

	return &Package{
		Entry:   envelope.Entry,
		Modules: modules,
		Imports: imports,
	}, nil
}

func (p *Package) ResolveImport(fromPath, moduleName string) (string, bool) {
	if p == nil {
		return "", false
	}
	rules, ok := p.Imports[fromPath]
	if !ok {
		return "", false
	}
	targetPath, ok := rules[moduleName]
	if !ok {
		return "", false
	}
	if _, exists := p.Modules[targetPath]; !exists {
		return "", false
	}
	return targetPath, true
}

func cloneImports(in map[string]map[string]string) map[string]map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]map[string]string, len(in))
	for fromPath, rules := range in {
		clonedRules := make(map[string]string, len(rules))
		for name, target := range rules {
			clonedRules[name] = target
		}
		out[fromPath] = clonedRules
	}
	return out
}
