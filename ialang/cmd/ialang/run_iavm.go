package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"iacommon/pkg/host/api"
	hostfs "iacommon/pkg/host/fs"
	hostnet "iacommon/pkg/host/network"
	"iavm/pkg/binary"
	"iavm/pkg/runtime"
)

func executeRunIavmCommand(path string, stderr io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}

	mod, err := binary.DecodeModule(data)
	if err != nil {
		return fmt.Errorf("decode module error: %w", err)
	}

	result, err := binary.VerifyModule(mod, binary.VerifyOptions{RequireEntry: true})
	if err != nil {
		return fmt.Errorf("verify module error: %w", err)
	}
	if !result.Valid {
		return fmt.Errorf("module verification failed: %v", result.Errors)
	}

	host := &api.DefaultHost{
		FS:      &hostfs.MemFSProvider{},
		Network: &hostnet.HTTPProvider{},
	}

	vm, err := runtime.New(mod, runtime.Options{
		Host: host,
	})
	if err != nil {
		return fmt.Errorf("vm init error: %w", err)
	}

	if err := vm.Run(); err != nil {
		return fmt.Errorf("runtime error: %w", err)
	}

	return nil
}

type simpleHost struct {
	fs      hostfs.Provider
	network hostnet.Provider
}

func (h *simpleHost) AcquireCapability(ctx context.Context, req api.AcquireRequest) (api.CapabilityInstance, error) {
	return api.CapabilityInstance{
		ID:   string(req.Kind),
		Kind: req.Kind,
	}, nil
}

func (h *simpleHost) ReleaseCapability(ctx context.Context, capID string) error {
	return nil
}

func (h *simpleHost) Call(ctx context.Context, req api.CallRequest) (api.CallResult, error) {
	switch req.Operation {
	case "fs.read_file":
		path := req.Args["path"].(string)
		data, err := h.fs.ReadFile(ctx, path)
		if err != nil {
			return api.CallResult{}, err
		}
		return api.CallResult{Value: map[string]any{"data": data}}, nil
	case "fs.write_file":
		path := req.Args["path"].(string)
		data := req.Args["data"].([]byte)
		return api.CallResult{}, h.fs.WriteFile(ctx, path, data, hostfs.WriteOptions{Create: true, Trunc: true})
	default:
		return api.CallResult{Value: map[string]any{}}, nil
	}
}

func (h *simpleHost) Poll(ctx context.Context, handleID uint64) (api.PollResult, error) {
	return api.PollResult{}, nil
}
