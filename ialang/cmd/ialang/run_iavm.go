package main

import (
	"fmt"
	"io"

	"iacommon/pkg/host/api"
	hostfs "iacommon/pkg/host/fs"
	hostnet "iacommon/pkg/host/network"
	"iavm/pkg/runtime"
)

func executeRunIavmCommand(cmd cliCommand, stderr io.Writer) error {
	opts, err := buildIavmVerifyOptions(cmd, true)
	if err != nil {
		return err
	}
	mod, err := loadAndVerifyIavmModule(cmd.file, opts)
	if err != nil {
		return err
	}

	host := &api.DefaultHost{
		FS:      &hostfs.MemFSProvider{},
		Network: &hostnet.HTTPProvider{},
	}

	vm, err := runtime.New(mod, runtime.Options{
		Host: host,
	})
	if err != nil {
		return fmt.Errorf("[runtime] vm init error: %w", err)
	}

	if err := vm.Run(); err != nil {
		return fmt.Errorf("[runtime] %w", err)
	}

	return nil
}
