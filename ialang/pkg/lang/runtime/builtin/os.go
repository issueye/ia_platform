package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	goRuntime "runtime"
	"strings"
)

func newOSModule() Object {
	platformFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("os.platform expects 0 args, got %d", len(args))
		}
		return goRuntime.GOOS, nil
	})
	archFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("os.arch expects 0 args, got %d", len(args))
		}
		return goRuntime.GOARCH, nil
	})
	hostnameFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("os.hostname expects 0 args, got %d", len(args))
		}
		return os.Hostname()
	})
	tmpDirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("os.tmpDir expects 0 args, got %d", len(args))
		}
		return os.TempDir(), nil
	})
	tempDirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("os.tempDir expects 0 args, got %d", len(args))
		}
		return os.TempDir(), nil
	})
	userDirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("os.userDir expects 0 args, got %d", len(args))
		}
		return os.UserHomeDir()
	})
	configDirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("os.configDir expects 0 args, got %d", len(args))
		}
		return os.UserConfigDir()
	})
	cacheDirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("os.cacheDir expects 0 args, got %d", len(args))
		}
		return os.UserCacheDir()
	})
	dataDirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("os.dataDir expects 0 args, got %d", len(args))
		}
		return userDataDir()
	})

	namespace := Object{
		"platform":  platformFn,
		"arch":      archFn,
		"hostname":  hostnameFn,
		"cwd":       makeCwdFn("os"),
		"tmpDir":    tmpDirFn,
		"tempDir":   tempDirFn,
		"userDir":   userDirFn,
		"dataDir":   dataDirFn,
		"configDir": configDirFn,
		"cacheDir":  cacheDirFn,
		"getEnv":    makeGetEnvFn("os"),
		"setEnv":    makeSetEnvFn("os"),
		"env":       makeEnvFn("os"),
	}
	module := cloneObject(namespace)
	module["os"] = namespace
	return module
}

func userDataDir() (string, error) {
	switch goRuntime.GOOS {
	case "windows":
		if v := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); v != "" {
			return v, nil
		}
		if v := strings.TrimSpace(os.Getenv("APPDATA")); v != "" {
			return v, nil
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	default:
		if v := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); v != "" {
			return v, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share"), nil
	}

	return "", fmt.Errorf("unable to resolve user data directory")
}
