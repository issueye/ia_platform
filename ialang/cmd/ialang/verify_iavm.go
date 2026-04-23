package main

import (
	"fmt"
	"io"

	"iavm/pkg/binary"
)

func executeVerifyIavmCommand(cmd cliCommand, stdout, stderr io.Writer) error {
	_ = stderr

	opts, err := buildIavmVerifyOptions(cmd, false)
	if err != nil {
		return err
	}
	mod, err := loadAndVerifyIavmModule(cmd.file, opts)
	if err != nil {
		return err
	}

	mode := binary.VerifyOptionsProfileName(binary.VerifyProfile(cmd.profile), cmd.strict)
	fmt.Fprintf(stdout, "module verification passed: target=%s functions=%d mode=%s\n", mod.Target, len(mod.Functions), mode)
	fmt.Fprintf(stdout, "  %s\n", opts.PolicySummary())
	return nil
}
