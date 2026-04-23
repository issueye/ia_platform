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

func buildIavmVerifyOptions(cmd cliCommand, requireEntryDefault bool) (binary.VerifyOptions, error) {
	requireEntry := requireEntryDefault || cmd.strict
	overrides := binary.VerifyPolicyOverrides{
		MaxFunctions:           cmd.maxFunctions,
		MaxConstants:           cmd.maxConstants,
		MaxCodeSizePerFunction: cmd.maxCodeSize,
		MaxLocalsPerFunction:   cmd.maxLocals,
		MaxStackPerFunction:    cmd.maxStack,
		AllowedCapabilities:    append([]module.CapabilityKind(nil), cmd.allowedCapabilities...),
		CapabilityAllowlistSet: cmd.capabilityAllowlistSet,
	}
	if requireEntry {
		overrides.RequireEntry = &requireEntry
	}
	return binary.BuildVerifyOptions(binary.VerifyProfile(cmd.profile), overrides)
}
