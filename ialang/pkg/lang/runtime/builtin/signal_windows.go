//go:build windows

package builtin

import "os"

func platformSignalConstants() Object {
	return Object{
		"SIGTERM": "SIGTERM",
	}
}

func parsePlatformSignalName(normalized string) (os.Signal, bool) {
	switch normalized {
	case "TERM", "SIGTERM":
		return os.Kill, true
	default:
		return nil, false
	}
}

func platformSignalName(sig os.Signal) (string, bool) {
	if sig == os.Kill {
		return "SIGTERM", true
	}
	return "", false
}