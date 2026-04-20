package binary

import "iavm/pkg/module"

type VerifyOptions struct {
	RequireEntry bool
	AllowCustom  bool
}

func VerifyModule(m *module.Module, opts VerifyOptions) (*VerifyResult, error) {
	_ = m
	_ = opts
	return &VerifyResult{}, nil
}
