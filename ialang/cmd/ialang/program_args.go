package main

import "os"

func withProgramArgs(programPath string, scriptArgs []string, fn func() error) error {
	oldArgs := append([]string(nil), os.Args...)
	hostArg := "ialang"
	if len(oldArgs) > 0 {
		hostArg = oldArgs[0]
	}

	newArgs := []string{hostArg, programPath}
	if len(scriptArgs) > 0 {
		newArgs = append(newArgs, scriptArgs...)
	}

	os.Args = newArgs
	defer func() {
		os.Args = oldArgs
	}()

	return fn()
}
