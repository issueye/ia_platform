package main

import "os"

func main() {
	if handled, code := maybeRunEmbeddedPackage(os.Stderr); handled {
		os.Exit(code)
	}
	os.Exit(runCLI(os.Args, os.Stdout, os.Stderr))
}
