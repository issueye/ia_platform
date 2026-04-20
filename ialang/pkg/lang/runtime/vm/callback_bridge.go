package vm

import rt "ialang/pkg/lang/runtime"

func init() {
	rt.RegisterUserFunctionCaller(CallUserFunctionSync)
}