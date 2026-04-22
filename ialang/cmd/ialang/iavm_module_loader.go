package main

import (
	"fmt"
	"os"

	"iavm/pkg/binary"
	"iavm/pkg/module"
)

func loadIavmModule(path string) (*module.Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}

	mod, err := binary.DecodeModule(data)
	if err != nil {
		return nil, fmt.Errorf("decode module error: %w", err)
	}
	return mod, nil
}

func loadAndVerifyIavmModule(path string, opts binary.VerifyOptions) (*module.Module, error) {
	mod, err := loadIavmModule(path)
	if err != nil {
		return nil, err
	}

	result, err := binary.VerifyModule(mod, opts)
	if err != nil {
		return nil, fmt.Errorf("verify module error: %w", err)
	}
	if !result.Valid {
		return nil, fmt.Errorf("module verification failed: %v", result.Errors)
	}
	return mod, nil
}

func buildIavmVerifyOptions(cmd cliCommand, requireEntryDefault bool) binary.VerifyOptions {
	opts := binary.VerifyOptions{}

	switch cmd.profile {
	case "strict":
		opts.RequireEntry = true
	case "sandbox":
		opts.RequireEntry = true
		opts.MaxFunctions = 128
		opts.MaxConstants = 512
		opts.MaxCodeSizePerFunction = 4096
		opts.MaxLocalsPerFunction = 64
		opts.MaxStackPerFunction = 128
		opts.AllowedCapabilities = []module.CapabilityKind{}
	}

	if requireEntryDefault || cmd.strict {
		opts.RequireEntry = true
	}
	if cmd.maxFunctions > 0 {
		opts.MaxFunctions = cmd.maxFunctions
	}
	if cmd.maxConstants > 0 {
		opts.MaxConstants = cmd.maxConstants
	}
	if cmd.maxCodeSize > 0 {
		opts.MaxCodeSizePerFunction = cmd.maxCodeSize
	}
	if cmd.maxLocals > 0 {
		opts.MaxLocalsPerFunction = cmd.maxLocals
	}
	if cmd.maxStack > 0 {
		opts.MaxStackPerFunction = cmd.maxStack
	}
	if len(cmd.allowedCapabilities) > 0 {
		opts.AllowedCapabilities = append([]module.CapabilityKind(nil), cmd.allowedCapabilities...)
	}

	return opts
}
