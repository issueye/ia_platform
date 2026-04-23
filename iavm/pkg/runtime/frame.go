package runtime

import (
	"iavm/pkg/core"
	"iavm/pkg/module"
)

type TryHandler struct {
	HandlerIP     uint32
	CatchLocalIdx uint32
	HasCatchVar   bool
}

type Frame struct {
	FunctionIndex uint32
	IP            uint32
	Locals        []core.Value
	BasePointer   uint32
	TryHandlers   []TryHandler
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
		TryHandlers:   make([]TryHandler, 0, 4),
	}
}
