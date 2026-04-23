package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"iavm/pkg/module"
)

type cliCommand struct {
	name                string
	file                string
	out                 string
	args                []string
	strict              bool
	verbose             bool
	verify              bool
	profile             string
	maxFunctions        int
	maxConstants        int
	maxCodeSize         int
	maxLocals           int
	maxStack            int
	allowedCapabilities []module.CapabilityKind
	helpShown           bool
}

const usageText = "usage:\n  ialang run <file> [args...]\n  ialang build <entry.ia> [-o output.iapkg]\n  ialang run-pkg <file.iapkg> [args...]\n  ialang build-bin <entry.ia> [-o output.exe]\n  ialang build-iavm <entry.ia> [-o output.iavm]\n  ialang verify-iavm <file.iavm> [--profile default|strict|sandbox] [--strict] [--max-functions N] [--max-constants N] [--max-code-size N] [--max-locals N] [--max-stack N] [--allow-capability fs|network]\n  ialang inspect-iavm <file.iavm> [--verbose] [--verify] [--profile default|strict|sandbox] [--strict] [--max-functions N] [--max-constants N] [--max-code-size N] [--max-locals N] [--max-stack N] [--allow-capability fs|network]\n  ialang run-iavm <file.iavm> [--profile default|strict|sandbox] [--strict] [--max-functions N] [--max-constants N] [--max-code-size N] [--max-locals N] [--max-stack N] [--allow-capability fs|network]\n  ialang init [dir]\n  ialang check [entry.ia|project-dir]\n  ialang fmt [path]  (path can be a file or directory, defaults to current directory)"

func runCLI(args []string, stdout, stderr io.Writer) int {
	cmd, err := parseCLIArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		fmt.Fprintln(stderr, usageText)
		return 1
	}

	if cmd.helpShown {
		fmt.Fprintln(stdout, usageText)
		return 0
	}

	switch cmd.name {
	case "run":
		if err := executeRunCommand(cmd.file, cmd.args, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "build":
		if err := executeBuildCommand(cmd.file, cmd.out, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "run-pkg":
		if err := executeRunPkgCommand(cmd.file, cmd.args, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "build-bin":
		if err := executeBuildBinCommand(cmd.file, cmd.out, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "init":
		if err := executeInitCommand(cmd.file, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "check":
		if err := executeCheckCommand(cmd.file, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "fmt":
		if cmd.helpShown {
			return 0
		}
		if err := executeFmtCommand(cmd.file, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "build-iavm":
		if err := executeBuildIavmCommand(cmd.file, cmd.out, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "verify-iavm":
		if err := executeVerifyIavmCommand(cmd, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "inspect-iavm":
		if err := executeInspectIavmCommand(cmd, stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	case "run-iavm":
		if err := executeRunIavmCommand(cmd, stderr); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	default:
		fmt.Fprintf(stderr, "unsupported command: %s\n", cmd.name)
		fmt.Fprintln(stderr, usageText)
		return 1
	}
}

func parseCLIArgs(args []string) (cliCommand, error) {
	if len(args) < 2 {
		return cliCommand{}, fmt.Errorf("missing command")
	}
	if args[1] == "--help" || args[1] == "-h" || args[1] == "help" {
		return cliCommand{name: "help", helpShown: true}, nil
	}
	switch args[1] {
	case "run":
		if len(args) < 3 {
			return cliCommand{}, fmt.Errorf("run expects a file argument")
		}
		return cliCommand{name: "run", file: args[2], args: append([]string(nil), args[3:]...)}, nil
	case "run-pkg":
		if len(args) < 3 {
			return cliCommand{}, fmt.Errorf("run-pkg expects a package file argument")
		}
		return cliCommand{name: "run-pkg", file: args[2], args: append([]string(nil), args[3:]...)}, nil
	case "build":
		return parseBuildCLIArgs(args)
	case "build-bin":
		return parseBuildBinCLIArgs(args)
	case "init":
		return parseInitCLIArgs(args)
	case "check":
		return parseCheckCLIArgs(args)
	case "fmt":
		return parseFmtCLIArgs(args)
	case "build-iavm":
		return parseBuildIavmCLIArgs(args)
	case "verify-iavm":
		return parseIavmVerifyCLIArgs("verify-iavm", args)
	case "inspect-iavm":
		return parseInspectIavmCLIArgs(args)
	case "run-iavm":
		return parseIavmVerifyCLIArgs("run-iavm", args)
	default:
		return cliCommand{}, fmt.Errorf("unsupported command: %s", args[1])
	}
}

func parseBuildCLIArgs(args []string) (cliCommand, error) {
	return parseBuildLikeCLIArgs("build", args)
}

func parseBuildBinCLIArgs(args []string) (cliCommand, error) {
	return parseBuildLikeCLIArgs("build-bin", args)
}

func parseBuildLikeCLIArgs(command string, args []string) (cliCommand, error) {
	if len(args) < 3 {
		return cliCommand{}, fmt.Errorf("%s expects an entry file", command)
	}
	cmd := cliCommand{name: command, file: args[2]}
	remaining := args[3:]

	for i := 0; i < len(remaining); i++ {
		tok := remaining[i]
		switch tok {
		case "-o", "--out":
			if i+1 >= len(remaining) {
				return cliCommand{}, fmt.Errorf("%s requires an output file", tok)
			}
			i++
			if cmd.out != "" {
				return cliCommand{}, fmt.Errorf("output file provided multiple times")
			}
			cmd.out = remaining[i]
		default:
			if strings.HasPrefix(tok, "-") {
				return cliCommand{}, fmt.Errorf("unknown %s option: %s", command, tok)
			}
			if cmd.out != "" {
				return cliCommand{}, fmt.Errorf("too many %s arguments", command)
			}
			cmd.out = tok
		}
	}
	return cmd, nil
}

func parseInitCLIArgs(args []string) (cliCommand, error) {
	if len(args) > 3 {
		return cliCommand{}, fmt.Errorf("init expects at most one target directory")
	}
	target := "."
	if len(args) == 3 {
		target = args[2]
	}
	return cliCommand{name: "init", file: target}, nil
}

func parseCheckCLIArgs(args []string) (cliCommand, error) {
	if len(args) > 3 {
		return cliCommand{}, fmt.Errorf("check expects at most one target (entry file or project directory)")
	}
	target := ""
	if len(args) == 3 {
		target = args[2]
	}
	return cliCommand{name: "check", file: target}, nil
}

func parseFmtCLIArgs(args []string) (cliCommand, error) {
	// fmt expects zero or one argument (file or directory path)
	// If no argument provided, use current directory
	if len(args) > 3 {
		return cliCommand{}, fmt.Errorf("fmt expects at most one path argument")
	}

	// Handle --help and -h flags
	if len(args) == 3 && (args[2] == "--help" || args[2] == "-h") {
		fmt.Println("Format .ia source files")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  ialang fmt [path]")
		fmt.Println("")
		fmt.Println("Arguments:")
		fmt.Println("  path    File or directory path. If a directory, formats all .ia files recursively.")
		fmt.Println("          Defaults to current directory if omitted.")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  ialang fmt              # Format all .ia files in current directory")
		fmt.Println("  ialang fmt ./examples   # Format all .ia files in ./examples")
		fmt.Println("  ialang fmt file.ia      # Format a single file")
		// Return a special command to indicate help was shown
		return cliCommand{name: "fmt", file: "", helpShown: true}, nil
	}

	target := "."
	if len(args) == 3 {
		target = args[2]
	}
	return cliCommand{name: "fmt", file: target}, nil
}

func parseBuildIavmCLIArgs(args []string) (cliCommand, error) {
	if len(args) < 3 {
		return cliCommand{}, fmt.Errorf("build-iavm expects an entry file")
	}
	cmd := cliCommand{name: "build-iavm", file: args[2]}
	remaining := args[3:]

	for i := 0; i < len(remaining); i++ {
		tok := remaining[i]
		switch tok {
		case "-o", "--out":
			if i+1 >= len(remaining) {
				return cliCommand{}, fmt.Errorf("build-iavm requires an output file")
			}
			i++
			if cmd.out != "" {
				return cliCommand{}, fmt.Errorf("output file provided multiple times")
			}
			cmd.out = remaining[i]
		default:
			if strings.HasPrefix(tok, "-") {
				return cliCommand{}, fmt.Errorf("unknown build-iavm option: %s", tok)
			}
			if cmd.out != "" {
				return cliCommand{}, fmt.Errorf("too many build-iavm arguments")
			}
			cmd.out = tok
		}
	}
	return cmd, nil
}

func parseIavmVerifyCLIArgs(command string, args []string) (cliCommand, error) {
	if len(args) < 3 {
		return cliCommand{}, fmt.Errorf("%s expects a file argument", command)
	}
	cmd := cliCommand{name: command, file: args[2]}
	remaining := args[3:]

	for i := 0; i < len(remaining); i++ {
		tok := remaining[i]
		handled, err := parseIavmVerifyOption(command, &cmd, tok, remaining, &i)
		if err != nil {
			return cliCommand{}, err
		}
		if !handled {
			return cliCommand{}, fmt.Errorf("unknown %s option: %s", command, tok)
		}
	}
	return cmd, nil
}

func parseIavmVerifyOption(command string, cmd *cliCommand, option string, args []string, index *int) (bool, error) {
	switch option {
	case "--profile":
		profile, err := parseIavmProfileOption(command, option, args, index)
		if err != nil {
			return false, err
		}
		if cmd.profile != "" {
			return false, fmt.Errorf("profile provided multiple times")
		}
		cmd.profile = profile
		return true, nil
	case "--strict":
		if cmd.strict {
			return false, fmt.Errorf("strict mode provided multiple times")
		}
		cmd.strict = true
		return true, nil
	case "--max-functions":
		value, err := parsePositiveIntOption(command, option, args, index)
		if err != nil {
			return false, err
		}
		cmd.maxFunctions = value
		return true, nil
	case "--max-constants":
		value, err := parsePositiveIntOption(command, option, args, index)
		if err != nil {
			return false, err
		}
		cmd.maxConstants = value
		return true, nil
	case "--max-code-size":
		value, err := parsePositiveIntOption(command, option, args, index)
		if err != nil {
			return false, err
		}
		cmd.maxCodeSize = value
		return true, nil
	case "--max-locals":
		value, err := parsePositiveIntOption(command, option, args, index)
		if err != nil {
			return false, err
		}
		cmd.maxLocals = value
		return true, nil
	case "--max-stack":
		value, err := parsePositiveIntOption(command, option, args, index)
		if err != nil {
			return false, err
		}
		cmd.maxStack = value
		return true, nil
	case "--allow-capability":
		capability, err := parseCapabilityOption(command, option, args, index)
		if err != nil {
			return false, err
		}
		cmd.allowedCapabilities = append(cmd.allowedCapabilities, capability)
		return true, nil
	default:
		return false, nil
	}
}

func parsePositiveIntOption(command, option string, args []string, index *int) (int, error) {
	if *index+1 >= len(args) {
		return 0, fmt.Errorf("%s requires a value for %s", command, option)
	}
	*index++
	value, err := strconv.Atoi(args[*index])
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s requires a positive integer for %s", command, option)
	}
	return value, nil
}

func parseCapabilityOption(command, option string, args []string, index *int) (module.CapabilityKind, error) {
	if *index+1 >= len(args) {
		return "", fmt.Errorf("%s requires a value for %s", command, option)
	}
	*index++
	switch args[*index] {
	case string(module.CapabilityFS):
		return module.CapabilityFS, nil
	case string(module.CapabilityNetwork):
		return module.CapabilityNetwork, nil
	default:
		return "", fmt.Errorf("%s requires a supported capability for %s", command, option)
	}
}

func parseIavmProfileOption(command, option string, args []string, index *int) (string, error) {
	if *index+1 >= len(args) {
		return "", fmt.Errorf("%s requires a value for %s", command, option)
	}
	*index++
	switch args[*index] {
	case "default", "strict", "sandbox":
		return args[*index], nil
	default:
		return "", fmt.Errorf("%s requires a supported profile for %s", command, option)
	}
}

func parseInspectIavmCLIArgs(args []string) (cliCommand, error) {
	if len(args) < 3 {
		return cliCommand{}, fmt.Errorf("inspect-iavm expects a file argument")
	}
	cmd := cliCommand{name: "inspect-iavm", file: args[2]}
	remaining := args[3:]

	for i := 0; i < len(remaining); i++ {
		tok := remaining[i]
		switch tok {
		case "--verbose":
			if cmd.verbose {
				return cliCommand{}, fmt.Errorf("verbose mode provided multiple times")
			}
			cmd.verbose = true
		case "--verify":
			if cmd.verify {
				return cliCommand{}, fmt.Errorf("verify mode provided multiple times")
			}
			cmd.verify = true
		default:
			handled, err := parseIavmVerifyOption("inspect-iavm", &cmd, tok, remaining, &i)
			if err != nil {
				return cliCommand{}, err
			}
			if !handled {
				return cliCommand{}, fmt.Errorf("unknown inspect-iavm option: %s", tok)
			}
			cmd.verify = true
		}
	}
	return cmd, nil
}
