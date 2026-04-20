package ialang

import (
	"errors"
	"fmt"
)

var ErrUnknownModule = errors.New("unknown module")

func ResolveModule(name string) (any, error) {
	switch name {
	case PlatformFSModuleName:
		return BuildPlatformFSModule(), nil
	case "@platform/http":
		return BuildPlatformHTTPModule(), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownModule, name)
	}
}
