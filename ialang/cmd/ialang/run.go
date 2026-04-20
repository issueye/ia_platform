package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ialang/pkg/lang"
	bc "ialang/pkg/lang/bytecode"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	"ialang/pkg/pool"
)

func executeRunCommand(path string, scriptArgs []string, stderr io.Writer) error {
	src, err := readRunSource(path)
	if err != nil {
		return err
	}

	chunk, err := compileRunSourceWithUnit(path, src, stderr)
	if err != nil {
		return err
	}

	if err := executeRunChunk(path, chunk, scriptArgs); err != nil {
		return err
	}
	return nil
}

func readRunSource(path string) (string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file error: %w", err)
	}
	return string(src), nil
}

func compileRunSource(src string, stderr io.Writer) (*bc.Chunk, error) {
	return compileRunSourceWithUnit("", src, stderr)
}

func compileRunSourceWithUnit(unitPath, src string, stderr io.Writer) (*bc.Chunk, error) {
	l := lang.NewLexer(src)
	p := lang.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		if strings.TrimSpace(unitPath) != "" {
			fmt.Fprintf(stderr, "parse errors in %s:\n", unitPath)
		} else {
			fmt.Fprintln(stderr, "parse errors:")
		}
		for _, e := range p.Errors() {
			fmt.Fprintf(stderr, "- %s\n", e)
		}
		return nil, fmt.Errorf("parse failed")
	}

	c := lang.NewCompiler()
	chunk, compileErrs := c.Compile(program)
	if len(compileErrs) > 0 {
		if strings.TrimSpace(unitPath) != "" {
			fmt.Fprintf(stderr, "compile errors in %s:\n", unitPath)
		} else {
			fmt.Fprintln(stderr, "compile errors:")
		}
		for _, e := range compileErrs {
			fmt.Fprintf(stderr, "- %s\n", e)
		}
		return nil, fmt.Errorf("compile failed")
	}
	return chunk, nil
}

func executeRunChunk(path string, chunk *bc.Chunk, scriptArgs []string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path error: %w", err)
	}

	return withProgramArgs(absPath, scriptArgs, func() error {
		// 初始化协程池管理器
		poolOpts := pool.DefaultPoolManagerOptions()
		poolOpts.EnableDefault = true
		poolOpts.EnableIOPool = true
		pm := pool.GetPoolManager()
		if err := pm.EnsureInitializedWithOptions(poolOpts); err != nil {
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
		resolverOptions, err := lang.DiscoverModuleResolverOptions(absPath)
		if err != nil {
			return fmt.Errorf("module resolver config error: %w", err)
		}
		loader := lang.NewModuleLoaderWithResolverOptions(modules, asyncRuntime, vmOptions, resolverOptions)
		vm := lang.NewVMWithOptions(chunk, modules, loader.Resolve, absPath, asyncRuntime, vmOptions)
		if err := vm.Run(); err != nil {
			return fmt.Errorf("runtime error: %w", err)
		}
		if err := vm.AutoCallMain(); err != nil {
			return fmt.Errorf("runtime error: %w", err)
		}
		return nil
	})
}
