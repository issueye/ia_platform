package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ialang/pkg/lang"
	bc "ialang/pkg/lang/bytecode"
	"ialang/pkg/lang/packagefile"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	"ialang/pkg/pool"
)

const defaultBuildOutput = "app.iapkg"

func executeBuildCommand(entryPath, outPath string, stderr io.Writer) error {
	if outPath == "" {
		outPath = defaultBuildOutput
	}
	pkg, err := buildPackage(entryPath, stderr)
	if err != nil {
		return err
	}
	encoded, err := packagefile.Encode(pkg)
	if err != nil {
		return fmt.Errorf("encode package error: %w", err)
	}
	if err := os.WriteFile(outPath, encoded, 0o644); err != nil {
		return fmt.Errorf("write package error: %w", err)
	}
	return nil
}

func executeRunPkgCommand(pkgPath string, scriptArgs []string, _ io.Writer) error {
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return fmt.Errorf("read package error: %w", err)
	}
	pkg, err := packagefile.Decode(data)
	if err != nil {
		return fmt.Errorf("decode package error: %w", err)
	}
	return runDecodedPackage(pkg, pkgPath, scriptArgs)
}

func runDecodedPackage(pkg *packagefile.Package, programPath string, scriptArgs []string) error {
	if pkg == nil {
		return fmt.Errorf("package is nil")
	}
	entryChunk, ok := pkg.Modules[pkg.Entry]
	if !ok {
		return fmt.Errorf("entry module not found in package: %s", pkg.Entry)
	}
	if strings.TrimSpace(programPath) == "" {
		programPath = pkg.Entry
	}

	return withProgramArgs(programPath, scriptArgs, func() error {
		poolOpts := pool.DefaultPoolManagerOptions()
		poolOpts.EnableDefault = true
		poolOpts.EnableIOPool = true
		pm := pool.GetPoolManager()
		if err := pm.EnsureInitialized(); err != nil {
			return fmt.Errorf("pool manager init error: %w", err)
		}

		asyncRuntime, err := buildAsyncRuntimeFromEnv()
		if err != nil {
			return fmt.Errorf("async runtime config error: %w", err)
		}
		vmOptions, err := buildVMOptionsFromEnv()
		if err != nil {
			return fmt.Errorf("vm config error: %w", err)
		}

		modules := rtbuiltin.DefaultModules(asyncRuntime)
		loader := newPackageModuleLoader(pkg, modules, asyncRuntime, vmOptions)
		vm := lang.NewVMWithOptions(entryChunk, modules, loader.Resolve, pkg.Entry, asyncRuntime, vmOptions)
		if err := vm.Run(); err != nil {
			return fmt.Errorf("runtime error: %w", err)
		}
		if err := vm.AutoCallMain(); err != nil {
			return fmt.Errorf("runtime error: %w", err)
		}
		return nil
	})
}

func buildPackage(entryPath string, stderr io.Writer) (*packagefile.Package, error) {
	entryAbs, err := filepath.Abs(entryPath)
	if err != nil {
		return nil, fmt.Errorf("resolve entry path error: %w", err)
	}

	resolverOptions, err := lang.DiscoverModuleResolverOptions(entryAbs)
	if err != nil {
		return nil, err
	}
	builtins := rtbuiltin.DefaultModules(nil)

	modules := map[string]*bc.Chunk{}
	imports := map[string]map[string]string{}
	queue := []string{entryAbs}
	queued := map[string]bool{entryAbs: true}

	for len(queue) > 0 {
		modulePath := queue[0]
		queue = queue[1:]
		delete(queued, modulePath)

		if _, exists := modules[modulePath]; exists {
			continue
		}

		src, err := readRunSource(modulePath)
		if err != nil {
			return nil, fmt.Errorf("read module error (%s): %w", modulePath, err)
		}
		chunk, err := compileRunSourceWithUnit(modulePath, src, stderr)
		if err != nil {
			return nil, fmt.Errorf("compile module error (%s): %w", modulePath, err)
		}
		modules[modulePath] = chunk

		moduleImports, err := extractImportModules(chunk)
		if err != nil {
			return nil, fmt.Errorf("inspect module imports error (%s): %w", modulePath, err)
		}
		if len(moduleImports) == 0 {
			continue
		}

		rules := map[string]string{}
		for _, moduleName := range moduleImports {
			if _, ok := builtins[moduleName]; ok {
				continue
			}
			targetPath, err := lang.ResolveModulePathWithOptions(modulePath, moduleName, resolverOptions)
			if err != nil {
				return nil, err
			}
			rules[moduleName] = targetPath
			if _, exists := modules[targetPath]; !exists && !queued[targetPath] {
				queue = append(queue, targetPath)
				queued[targetPath] = true
			}
		}
		if len(rules) > 0 {
			imports[modulePath] = rules
		}
	}

	return &packagefile.Package{
		Entry:   entryAbs,
		Modules: modules,
		Imports: imports,
	}, nil
}

func extractImportModules(chunk *bc.Chunk) ([]string, error) {
	seen := map[string]struct{}{}
	for _, ins := range chunk.Code {
		if ins.Op != bc.OpImportName {
			continue
		}
		if ins.A < 0 || ins.A >= len(chunk.Constants) {
			return nil, fmt.Errorf("import module constant index out of range: %d", ins.A)
		}
		moduleName, ok := chunk.Constants[ins.A].(string)
		if !ok {
			return nil, fmt.Errorf("import module constant is not string: %T", chunk.Constants[ins.A])
		}
		seen[moduleName] = struct{}{}
	}
	if len(seen) == 0 {
		return nil, nil
	}
	moduleNames := make([]string, 0, len(seen))
	for moduleName := range seen {
		moduleNames = append(moduleNames, moduleName)
	}
	sort.Strings(moduleNames)
	return moduleNames, nil
}

type packageModuleLoader struct {
	pkg          *packagefile.Package
	builtins     map[string]lang.Value
	asyncRuntime lang.AsyncRuntime
	vmOptions    lang.VMOptions
	cache        map[string]lang.Object
	loading      map[string]bool
}

func newPackageModuleLoader(pkg *packagefile.Package, builtins map[string]lang.Value, asyncRuntime lang.AsyncRuntime, vmOptions lang.VMOptions) *packageModuleLoader {
	return &packageModuleLoader{
		pkg:          pkg,
		builtins:     builtins,
		asyncRuntime: asyncRuntime,
		vmOptions:    vmOptions,
		cache:        map[string]lang.Object{},
		loading:      map[string]bool{},
	}
}

func (m *packageModuleLoader) Resolve(fromPath, moduleName string) (lang.Value, error) {
	if val, ok := m.builtins[moduleName]; ok {
		return val, nil
	}
	targetPath, ok := m.pkg.ResolveImport(fromPath, moduleName)
	if !ok {
		return nil, fmt.Errorf("module not found in package: %s", moduleName)
	}
	if exports, ok := m.cache[targetPath]; ok {
		return exports, nil
	}
	if m.loading[targetPath] {
		return nil, fmt.Errorf("cyclic import detected: %s", targetPath)
	}
	chunk, ok := m.pkg.Modules[targetPath]
	if !ok {
		return nil, fmt.Errorf("module chunk not found in package: %s", targetPath)
	}

	m.loading[targetPath] = true
	defer delete(m.loading, targetPath)

	vm := lang.NewVMWithOptions(chunk, m.builtins, m.Resolve, targetPath, m.asyncRuntime, m.vmOptions)
	if err := vm.Run(); err != nil {
		return nil, fmt.Errorf("runtime module error (%s): %w", targetPath, err)
	}
	exports := clonePackageObject(vm.Exports())
	m.cache[targetPath] = exports
	return exports, nil
}

func clonePackageObject(obj lang.Object) lang.Object {
	out := lang.Object{}
	for k, v := range obj {
		out[k] = v
	}
	return out
}