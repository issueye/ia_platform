package main

import (
	"fmt"
	"io"
)

func executeVerifyIavmCommand(cmd cliCommand, stdout, stderr io.Writer) error {
	_ = stderr

	mod, err := loadAndVerifyIavmModule(cmd.file, buildIavmVerifyOptions(cmd, false))
	if err != nil {
		return err
	}

	mode := "default"
	if cmd.profile != "" {
		mode = cmd.profile
	}
	if cmd.strict && mode == "default" {
		mode = "strict"
	}
	fmt.Fprintf(stdout, "module verification passed: target=%s functions=%d mode=%s\n", mod.Target, len(mod.Functions), mode)
	return nil
}
