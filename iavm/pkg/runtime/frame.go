package runtime

import (
	"iavm/pkg/core"
	"iavm/pkg/module"
)

type Frame struct {
	FunctionIndex uint32
	IP            uint32
	Locals        []core.Value
	BasePointer   uint32
	TryHandlers   []uint32 // stack of try handler IP addresses
}

func NewFrame(fnIndex uint32, fn *module.Function, baseSP uint32) *Frame {
	locals := make([]core.Value, len(fn.Locals))
	for i := range locals {
		locals[i] = core.Value{Kind: core.ValueNull}
	}
	return &Frame{
		FunctionIndex: fnIndex,
		IP:            0,
		Locals:        locals,
		BasePointer:   baseSP,
		TryHandlers:   make([]uint32, 0, 4),
	}
}
