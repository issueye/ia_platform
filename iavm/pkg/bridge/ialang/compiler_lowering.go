package ialang

import "iavm/pkg/module"

func LowerToModule(input any) (*module.Module, error) {
	_ = input
	return &module.Module{}, nil
}
