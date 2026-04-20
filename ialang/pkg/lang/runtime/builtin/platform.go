package builtin

import (
	"context"
	"sync"
	"time"

	bridgeialang "iavm/pkg/bridge/ialang"
	hostapi "iavm/pkg/host/api"
)

const (
	platformFSModuleName   = "@platform/fs"
	platformHTTPModuleName = "@platform/http"
)

type lazyPlatformFSBridge struct {
	host hostapi.Host

	mu     sync.Mutex
	bridge *bridgeialang.PlatformFSBridge
}

type lazyPlatformHTTPBridge struct {
	host hostapi.Host

	mu     sync.Mutex
	bridge *bridgeialang.PlatformHTTPBridge
}

func DefaultModulesWithHost(asyncRuntime AsyncRuntime, host hostapi.Host) map[string]Value {
	modules := DefaultModules(asyncRuntime)
	modules[platformFSModuleName] = newPlatformFSModule(asyncRuntime, host)
	modules[platformHTTPModuleName] = newPlatformHTTPModule(asyncRuntime, host)
	return modules
}

func newPlatformFSModule(asyncRuntime AsyncRuntime, host hostapi.Host) Object {
	bridge := &lazyPlatformFSBridge{host: host}

	readFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, errArgCount("@platform/fs.readFile", len(args), "1")
		}
		path, err := asStringArg("@platform/fs.readFile", args, 0)
		if err != nil {
			return nil, err
		}
		return bridge.ReadFile(path)
	})

	writeFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, errArgCount("@platform/fs.writeFile", len(args), "2")
		}
		path, err := asStringArg("@platform/fs.writeFile", args, 0)
		if err != nil {
			return nil, err
		}
		content, err := asStringArg("@platform/fs.writeFile", args, 1)
		if err != nil {
			return nil, err
		}
		return bridge.WriteFile(path, content)
	})

	appendFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, errArgCount("@platform/fs.appendFile", len(args), "2")
		}
		path, err := asStringArg("@platform/fs.appendFile", args, 0)
		if err != nil {
			return nil, err
		}
		content, err := asStringArg("@platform/fs.appendFile", args, 1)
		if err != nil {
			return nil, err
		}
		return bridge.AppendFile(path, content)
	})

	existsFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, errArgCount("@platform/fs.exists", len(args), "1")
		}
		path, err := asStringArg("@platform/fs.exists", args, 0)
		if err != nil {
			return nil, err
		}
		return bridge.Exists(path)
	})

	mkdirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, errArgCount("@platform/fs.mkdir", len(args), "1-2")
		}
		path, err := asStringArg("@platform/fs.mkdir", args, 0)
		if err != nil {
			return nil, err
		}
		recursive := false
		if len(args) == 2 {
			recursive, err = asBoolArg("@platform/fs.mkdir", args, 1)
			if err != nil {
				return nil, err
			}
		}
		return bridge.Mkdir(path, recursive)
	})

	readDirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, errArgCount("@platform/fs.readDir", len(args), "1")
		}
		path, err := asStringArg("@platform/fs.readDir", args, 0)
		if err != nil {
			return nil, err
		}
		entries, err := bridge.ReadDir(path)
		if err != nil {
			return nil, err
		}
		out := make(Array, 0, len(entries))
		for _, entry := range entries {
			out = append(out, entry)
		}
		return out, nil
	})

	statFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, errArgCount("@platform/fs.stat", len(args), "1")
		}
		path, err := asStringArg("@platform/fs.stat", args, 0)
		if err != nil {
			return nil, err
		}
		info, err := bridge.Stat(path)
		if err != nil {
			return nil, err
		}
		return toRuntimeJSONValue(info), nil
	})

	renameFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, errArgCount("@platform/fs.rename", len(args), "2")
		}
		oldPath, err := asStringArg("@platform/fs.rename", args, 0)
		if err != nil {
			return nil, err
		}
		newPath, err := asStringArg("@platform/fs.rename", args, 1)
		if err != nil {
			return nil, err
		}
		return bridge.Rename(oldPath, newPath)
	})

	removeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, errArgCount("@platform/fs.remove", len(args), "1")
		}
		path, err := asStringArg("@platform/fs.remove", args, 0)
		if err != nil {
			return nil, err
		}
		return bridge.Remove(path)
	})

	removeAllFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, errArgCount("@platform/fs.removeAll", len(args), "1")
		}
		path, err := asStringArg("@platform/fs.removeAll", args, 0)
		if err != nil {
			return nil, err
		}
		return bridge.RemoveAll(path)
	})

	copyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, errArgCount("@platform/fs.copy", len(args), "2")
		}
		src, err := asStringArg("@platform/fs.copy", args, 0)
		if err != nil {
			return nil, err
		}
		dst, err := asStringArg("@platform/fs.copy", args, 1)
		if err != nil {
			return nil, err
		}
		return bridge.Copy(src, dst)
	})

	readFileAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return readFileFn(args)
		}), nil
	})
	writeFileAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return writeFileFn(args)
		}), nil
	})
	appendFileAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return appendFileFn(args)
		}), nil
	})
	existsAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return existsFn(args)
		}), nil
	})
	mkdirAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return mkdirFn(args)
		}), nil
	})
	readDirAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return readDirFn(args)
		}), nil
	})
	statAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return statFn(args)
		}), nil
	})
	renameAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return renameFn(args)
		}), nil
	})
	removeAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return removeFn(args)
		}), nil
	})
	removeAllAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return removeAllFn(args)
		}), nil
	})
	copyAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return copyFn(args)
		}), nil
	})

	namespace := Object{
		"readFile":        readFileFn,
		"writeFile":       writeFileFn,
		"appendFile":      appendFileFn,
		"exists":          existsFn,
		"mkdir":           mkdirFn,
		"readDir":         readDirFn,
		"stat":            statFn,
		"rename":          renameFn,
		"remove":          removeFn,
		"removeAll":       removeAllFn,
		"copy":            copyFn,
		"readFileAsync":   readFileAsyncFn,
		"writeFileAsync":  writeFileAsyncFn,
		"appendFileAsync": appendFileAsyncFn,
		"existsAsync":     existsAsyncFn,
		"mkdirAsync":      mkdirAsyncFn,
		"readDirAsync":    readDirAsyncFn,
		"statAsync":       statAsyncFn,
		"renameAsync":     renameAsyncFn,
		"removeAsync":     removeAsyncFn,
		"removeAllAsync":  removeAllAsyncFn,
		"copyAsync":       copyAsyncFn,
	}
	module := cloneObject(namespace)
	module["fs"] = namespace
	return module
}

func newPlatformHTTPModule(asyncRuntime AsyncRuntime, host hostapi.Host) Object {
	bridge := &lazyPlatformHTTPBridge{host: host}

	requestFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPRequestArgs("@platform/http.client.request", args, "GET")
		if err != nil {
			return nil, err
		}
		return bridge.request(cfg)
	})

	getFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPRequestArgs("@platform/http.client.get", args, "GET")
		if err != nil {
			return nil, err
		}
		if cfg.Method != "GET" {
			return nil, errMethod("@platform/http.client.get", "GET", cfg.Method)
		}
		return bridge.request(cfg)
	})

	postFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseHTTPRequestArgs("@platform/http.client.post", args, "POST")
		if err != nil {
			return nil, err
		}
		if cfg.Method != "POST" {
			return nil, errMethod("@platform/http.client.post", "POST", cfg.Method)
		}
		return bridge.request(cfg)
	})

	requestAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return requestFn(args)
		}), nil
	})
	getAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return getFn(args)
		}), nil
	})
	postAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return postFn(args)
		}), nil
	})

	clientNamespace := Object{
		"request":      requestFn,
		"get":          getFn,
		"post":         postFn,
		"requestAsync": requestAsyncFn,
		"getAsync":     getAsyncFn,
		"postAsync":    postAsyncFn,
	}
	moduleNamespace := Object{
		"client": clientNamespace,
	}
	module := cloneObject(moduleNamespace)
	module["http"] = moduleNamespace
	return module
}

func (b *lazyPlatformFSBridge) ReadFile(path string) (string, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return "", err
	}
	return bridge.ReadFile(path)
}

func (b *lazyPlatformFSBridge) WriteFile(path, content string) (bool, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return false, err
	}
	return bridge.WriteFile(path, content)
}

func (b *lazyPlatformFSBridge) AppendFile(path, content string) (bool, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return false, err
	}
	return bridge.AppendFile(path, content)
}

func (b *lazyPlatformFSBridge) Exists(path string) (bool, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return false, err
	}
	return bridge.Exists(path)
}

func (b *lazyPlatformFSBridge) Mkdir(path string, recursive bool) (bool, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return false, err
	}
	return bridge.Mkdir(path, recursive)
}

func (b *lazyPlatformFSBridge) ReadDir(path string) ([]string, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return nil, err
	}
	return bridge.ReadDir(path)
}

func (b *lazyPlatformFSBridge) Stat(path string) (map[string]any, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return nil, err
	}
	return bridge.Stat(path)
}

func (b *lazyPlatformFSBridge) Rename(oldPath, newPath string) (bool, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return false, err
	}
	return bridge.Rename(oldPath, newPath)
}

func (b *lazyPlatformFSBridge) Remove(path string) (bool, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return false, err
	}
	return bridge.Remove(path)
}

func (b *lazyPlatformFSBridge) RemoveAll(path string) (bool, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return false, err
	}
	return bridge.RemoveAll(path)
}

func (b *lazyPlatformFSBridge) Copy(src, dst string) (bool, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return false, err
	}
	return bridge.Copy(src, dst)
}

func (b *lazyPlatformFSBridge) bridgeForCall() (*bridgeialang.PlatformFSBridge, error) {
	if b == nil || b.host == nil {
		return nil, bridgeialang.ErrHostNotConfigured
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.bridge != nil {
		return b.bridge, nil
	}

	capability, err := b.host.AcquireCapability(context.Background(), hostapi.AcquireRequest{Kind: hostapi.CapabilityFS})
	if err != nil {
		return nil, err
	}
	b.bridge = &bridgeialang.PlatformFSBridge{Host: b.host, CapabilityID: capability.ID}
	return b.bridge, nil
}

func (b *lazyPlatformHTTPBridge) request(cfg httpRequestConfig) (Value, error) {
	bridge, err := b.bridgeForCall()
	if err != nil {
		return nil, err
	}

	options := map[string]any{
		"method":    cfg.Method,
		"body":      cfg.Body,
		"timeoutMS": float64(cfg.Timeout / time.Millisecond),
	}
	if len(cfg.Headers) > 0 {
		headers := make(map[string]any, len(cfg.Headers))
		for key, value := range cfg.Headers {
			headers[key] = yamlToGoValue(value)
		}
		options["headers"] = headers
	}
	if cfg.ContentType != "" {
		headers, _ := options["headers"].(map[string]any)
		if headers == nil {
			headers = map[string]any{}
		}
		if _, exists := headers["Content-Type"]; !exists && cfg.Body != "" {
			headers["Content-Type"] = cfg.ContentType
		}
		options["headers"] = headers
	}

	result, err := bridge.Request(cfg.URL, options)
	if err != nil {
		return nil, err
	}
	return toRuntimeJSONValue(result), nil
}

func (b *lazyPlatformHTTPBridge) bridgeForCall() (*bridgeialang.PlatformHTTPBridge, error) {
	if b == nil || b.host == nil {
		return nil, bridgeialang.ErrHostNotConfigured
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.bridge != nil {
		return b.bridge, nil
	}

	capability, err := b.host.AcquireCapability(context.Background(), hostapi.AcquireRequest{Kind: hostapi.CapabilityNetwork})
	if err != nil {
		return nil, err
	}
	b.bridge = &bridgeialang.PlatformHTTPBridge{Host: b.host, CapabilityID: capability.ID}
	return b.bridge, nil
}

func errArgCount(name string, got int, want string) error {
	return bridgeialang.ErrInvalidFSResult
}

func errMethod(name, want, got string) error {
	return bridgeialang.ErrInvalidHTTPResult
}
