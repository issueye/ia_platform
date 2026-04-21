package builtin

import (
	"fmt"
	"time"

	hostapi "iacommon/pkg/host/api"
	moduleapi "iacommon/pkg/ialang/module"
)

const (
	platformFSModuleName   = moduleapi.PlatformFSModuleName
	platformHTTPModuleName = moduleapi.PlatformHTTPModuleName
)

var (
	ErrHostNotConfigured       = moduleapi.ErrHostNotConfigured
	ErrCapabilityNotConfigured = moduleapi.ErrCapabilityNotConfigured
	ErrInvalidFSResult         = moduleapi.ErrInvalidFSResult
	ErrInvalidHTTPResult       = moduleapi.ErrInvalidHTTPResult
)

type lazyPlatformFSBridge struct {
	*moduleapi.PlatformFSBridge
}

type lazyPlatformHTTPBridge struct {
	*moduleapi.PlatformHTTPBridge
}

func DefaultModulesWithHost(asyncRuntime AsyncRuntime, host hostapi.Host) map[string]Value {
	modules := DefaultModules(asyncRuntime)
	modules[platformFSModuleName] = newPlatformFSModule(asyncRuntime, host)
	modules[platformHTTPModuleName] = newPlatformHTTPModule(asyncRuntime, host)
	return modules
}

func newPlatformFSModule(asyncRuntime AsyncRuntime, host hostapi.Host) Object {
	bridge := &lazyPlatformFSBridge{PlatformFSBridge: &moduleapi.PlatformFSBridge{Host: host}}

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
	bridge := &lazyPlatformHTTPBridge{PlatformHTTPBridge: &moduleapi.PlatformHTTPBridge{Host: host}}

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

func (b *lazyPlatformHTTPBridge) request(cfg httpRequestConfig) (Value, error) {
	result, err := b.PlatformHTTPBridge.Request(cfg.URL, map[string]any{
		"method":     cfg.Method,
		"headers":    normalizeRequestHeaders(cfg.Headers),
		"body":       cfg.Body,
		"timeoutMS":  int64(cfg.Timeout / time.Millisecond),
		"timeout_ms": int64(cfg.Timeout / time.Millisecond),
	})
	if err != nil {
		return nil, err
	}
	return toRuntimeJSONValue(result), nil
}

func normalizeRequestHeaders(headers Object) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	result := make(map[string]string, len(headers))
	for key, value := range headers {
		text, ok := value.(string)
		if !ok {
			continue
		}
		result[key] = text
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func errArgCount(name string, got int, want string) error {
	return fmt.Errorf("%w: %s expects %s args, got %d", ErrInvalidFSResult, name, want, got)
}

func errMethod(name, want, got string) error {
	return fmt.Errorf("%w: %s options.method must be %s, got %s", ErrInvalidHTTPResult, name, want, got)
}
