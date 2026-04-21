package main

import (
	"fmt"
	"io"
	"os"

	bc "iacommon/pkg/ialang/bytecode"
	"iavm/pkg/binary"
	bridge_ialang "iavm/pkg/bridge/ialang"
)

func executeBuildIavmCommand(entryPath, outPath string, stderr io.Writer) error {
	if outPath == "" {
		base := entryPath
		if len(base) > 3 && base[len(base)-3:] == ".ia" {
			outPath = base[:len(base)-3] + ".iavm"
		} else {
			outPath = base + ".iavm"
		}
	}

	chunk, err := compileRunSourceWithUnit(entryPath, readFileOrEmpty(entryPath), stderr)
	if err != nil {
		return err
	}

	mod, err := bridge_ialang.LowerToModule(chunk)
	if err != nil {
		return fmt.Errorf("lowering failed: %w", err)
	}

	data, err := binary.EncodeModule(mod)
	if err != nil {
		return fmt.Errorf("encoding failed: %w", err)
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("write output failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "built %s (%d bytes)\n", outPath, len(data))
	return nil
}

func readFileOrEmpty(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func compileIavmSource(path string, stderr io.Writer) (*bc.Chunk, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}
	return compileRunSourceWithUnit(path, string(src), stderr)
}
