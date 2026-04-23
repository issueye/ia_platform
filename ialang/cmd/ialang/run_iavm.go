package main

import (
	"context"
	"fmt"
	"io"

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
	cfg, err := loadCapabilityConfig(cmd.capConfig)
	if err != nil {
		return err
	}
	applyCapabilityConfig(mod, cfg)

	host, err := buildRunIavmHost(cfg)
	if err != nil {
		return err
	}

	vm, err := runtime.New(mod, runtime.Options{
		Host: host,
	})
	if err != nil {
		return fmt.Errorf("[runtime] vm init error: %w", err)
	}

	if err := vm.RunUntilSettled(context.Background()); err != nil {
		return fmt.Errorf("[runtime] %w", err)
	}

	return nil
}
