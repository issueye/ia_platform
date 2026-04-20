package runtime

import (
	"fmt"
	"time"
)

// SandboxPolicy defines execution limits and security constraints for running untrusted code.
type SandboxPolicy struct {
	// MaxSteps limits the number of VM instructions. Zero means unlimited.
	MaxSteps int64
	// MaxMemory limits approximate memory usage (stack + constants). Zero means unlimited.
	MaxMemory int64
	// MaxDuration limits total execution time. Zero means unlimited.
	MaxDuration time.Duration
	// AllowImport controls whether import statements are allowed.
	AllowImport bool
	// AllowedModules is a whitelist of module names that can be imported.
	// Empty means all modules are allowed (if AllowImport is true).
	AllowedModules map[string]bool
	// AllowFS controls whether file system operations are allowed.
	AllowFS bool
	// AllowNetwork controls whether network operations are allowed.
	AllowNetwork bool
	// AllowProcess controls whether process operations (exit, env) are allowed.
	AllowProcess bool
}

// DefaultSandboxPolicy returns a restrictive sandbox policy suitable for untrusted code.
func DefaultSandboxPolicy() *SandboxPolicy {
	return &SandboxPolicy{
		MaxSteps:      100000,
		MaxMemory:     10 * 1024 * 1024, // 10MB
		MaxDuration:   5 * time.Second,
		AllowImport:   true,
		AllowFS:       false,
		AllowNetwork:  false,
		AllowProcess:  false,
	}
}

// PermissiveSandboxPolicy returns a sandbox policy that allows most operations.
func PermissiveSandboxPolicy() *SandboxPolicy {
	return &SandboxPolicy{
		MaxSteps:     0,
		MaxMemory:    0,
		MaxDuration:  0,
		AllowImport:  true,
		AllowFS:      true,
		AllowNetwork: true,
		AllowProcess: true,
	}
}

// IsModuleAllowed checks if a module name is allowed by the sandbox policy.
func (p *SandboxPolicy) IsModuleAllowed(moduleName string) bool {
	if !p.AllowImport {
		return false
	}
	if len(p.AllowedModules) == 0 {
		return true
	}
	return p.AllowedModules[moduleName]
}

// AddAllowedModule adds a module to the whitelist.
func (p *SandboxPolicy) AddAllowedModule(name string) {
	if p.AllowedModules == nil {
		p.AllowedModules = make(map[string]bool)
	}
	p.AllowedModules[name] = true
}

// SandboxError is returned when a sandbox policy violation occurs.
type SandboxError struct {
	Violation string
	Limit     string
	Current   string
}

func (e *SandboxError) Error() string {
	return fmt.Sprintf("sandbox violation: %s (limit: %s, current: %s)", e.Violation, e.Limit, e.Current)
}

// StepCounter tracks the number of VM steps executed.
type StepCounter struct {
	count int64
	limit int64
}

// NewStepCounter creates a step counter with an optional limit.
func NewStepCounter(limit int64) *StepCounter {
	return &StepCounter{limit: limit}
}

// Increment increments the step counter and checks against the limit.
func (s *StepCounter) Increment() error {
	s.count++
	if s.limit > 0 && s.count > s.limit {
		return &SandboxError{
			Violation: "max steps exceeded",
			Limit:     fmt.Sprintf("%d", s.limit),
			Current:   fmt.Sprintf("%d", s.count),
		}
	}
	return nil
}

// Count returns the current step count.
func (s *StepCounter) Count() int64 {
	return s.count
}
