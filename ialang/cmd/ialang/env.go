package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"ialang/pkg/lang"
)

func buildAsyncRuntimeFromEnv() (lang.AsyncRuntime, error) {
	taskTimeout, err := readTimeoutMsEnv("IALANG_ASYNC_TASK_TIMEOUT_MS")
	if err != nil {
		return nil, err
	}
	awaitTimeout, err := readTimeoutMsEnv("IALANG_ASYNC_AWAIT_TIMEOUT_MS")
	if err != nil {
		return nil, err
	}

	return lang.NewGoroutineRuntimeWithOptions(lang.GoroutineRuntimeOptions{
		TaskTimeout:  taskTimeout,
		AwaitTimeout: awaitTimeout,
	}), nil
}

func buildVMOptionsFromEnv() (lang.VMOptions, error) {
	structuredErrors, err := readBoolEnv("IALANG_STRUCTURED_RUNTIME_ERRORS")
	if err != nil {
		return lang.VMOptions{}, err
	}
	return lang.VMOptions{
		StructuredRuntimeErrors: structuredErrors,
	}, nil
}

func readTimeoutMsEnv(name string) (time.Duration, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return 0, nil
	}
	ms, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be integer milliseconds, got %q", name, raw)
	}
	if ms < 0 {
		return 0, fmt.Errorf("%s must be >= 0, got %d", name, ms)
	}
	return time.Duration(ms) * time.Millisecond, nil
}

func readBoolEnv(name string) (bool, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return false, nil
	}
	switch raw {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true, nil
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be boolean (1/0/true/false/yes/no/on/off), got %q", name, raw)
	}
}
