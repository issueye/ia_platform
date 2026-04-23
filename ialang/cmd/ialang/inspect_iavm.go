package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"iavm/pkg/binary"
	"iavm/pkg/module"
)

func executeInspectIavmCommand(cmd cliCommand, stdout, stderr io.Writer) error {
	_ = stderr

	mod, err := loadIavmModule(cmd.file)
	if err != nil {
		return err
	}

	entry := inspectEntryName(mod)
	fmt.Fprintln(stdout, "IAVM module summary")
	fmt.Fprintf(stdout, "  magic: %s\n", mod.Magic)
	fmt.Fprintf(stdout, "  version: %d\n", mod.Version)
	fmt.Fprintf(stdout, "  target: %s\n", mod.Target)
	fmt.Fprintf(stdout, "  abi_version: %d\n", mod.ABIVersion)
	fmt.Fprintf(stdout, "  feature_flags: %d\n", mod.FeatureFlags)
	fmt.Fprintf(stdout, "  entry: %s\n", entry)
	fmt.Fprintf(stdout, "  types: %d\n", len(mod.Types))
	fmt.Fprintf(stdout, "  imports: %d\n", len(mod.Imports))
	fmt.Fprintf(stdout, "  functions: %d\n", len(mod.Functions))
	fmt.Fprintf(stdout, "  globals: %d\n", len(mod.Globals))
	fmt.Fprintf(stdout, "  exports: %d\n", len(mod.Exports))
	fmt.Fprintf(stdout, "  data_segments: %d\n", len(mod.DataSegments))
	fmt.Fprintf(stdout, "  capabilities: %d\n", len(mod.Capabilities))
	fmt.Fprintf(stdout, "  constants: %d\n", len(mod.Constants))
	fmt.Fprintf(stdout, "  custom_sections: %d\n", len(mod.Custom))

	if len(mod.Capabilities) > 0 {
		fmt.Fprintf(stdout, "  capability_kinds: %s\n", inspectCapabilityKinds(mod.Capabilities))
	}

	if cmd.verify {
		if err := inspectVerifyModule(mod, cmd, stdout); err != nil {
			return err
		}
	}

	if cmd.verbose {
		for i, fn := range mod.Functions {
			fmt.Fprintf(stdout, "  function[%d]: name=%s type=%d locals=%d code=%d max_stack=%d entry=%t\n",
				i, inspectNameOrPlaceholder(fn.Name), fn.TypeIndex, len(fn.Locals), len(fn.Code), fn.MaxStack, fn.IsEntryPoint)
		}
	}

	return nil
}

func inspectVerifyModule(mod *module.Module, cmd cliCommand, stdout io.Writer) error {
	opts, err := buildIavmVerifyOptions(cmd, false)
	if err != nil {
		return err
	}
	result, err := binary.VerifyModule(mod, opts)
	if err != nil {
		return fmt.Errorf("verify module error: %w", err)
	}
	if !result.Valid {
		return fmt.Errorf("module verification failed: %v", result.Errors)
	}

	fmt.Fprintf(stdout, "  verification: passed (mode=%s)\n", inspectVerifyMode(cmd))
	return nil
}

func inspectVerifyMode(cmd cliCommand) string {
	return binary.VerifyOptionsProfileName(binary.VerifyProfile(cmd.profile), cmd.strict)
}

func inspectEntryName(mod *module.Module) string {
	for _, fn := range mod.Functions {
		if fn.IsEntryPoint {
			return inspectNameOrPlaceholder(fn.Name)
		}
	}
	for _, fn := range mod.Functions {
		if fn.Name == "main" || fn.Name == "entry" {
			return fn.Name
		}
	}
	return "<none>"
}

func inspectCapabilityKinds(caps []module.CapabilityDecl) string {
	kinds := make([]string, 0, len(caps))
	for _, cap := range caps {
		kinds = append(kinds, string(cap.Kind))
	}
	sort.Strings(kinds)
	return strings.Join(kinds, ",")
}

func inspectNameOrPlaceholder(name string) string {
	if name == "" {
		return "<anonymous>"
	}
	return name
}
