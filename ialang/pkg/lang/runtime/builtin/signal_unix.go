//go:build !windows

package builtin

import (
	"os"
	"syscall"
)

func platformSignalConstants() Object {
	return Object{
		"SIGTERM": "SIGTERM",
		"SIGHUP":  "SIGHUP",
		"SIGQUIT": "SIGQUIT",
	}
}

func parsePlatformSignalName(normalized string) (os.Signal, bool) {
	switch normalized {
	case "TERM", "SIGTERM":
		return syscall.SIGTERM, true
	case "HUP", "SIGHUP":
		return syscall.SIGHUP, true
	case "QUIT", "SIGQUIT":
		return syscall.SIGQUIT, true
	default:
		return nil, false
	}
}

func platformSignalName(sig os.Signal) (string, bool) {
	switch sig {
	case syscall.SIGTERM:
		return "SIGTERM", true
	case syscall.SIGHUP:
		return "SIGHUP", true
	case syscall.SIGQUIT:
		return "SIGQUIT", true
	default:
		return "", false
	}
}