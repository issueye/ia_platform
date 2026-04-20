package vm

import (
	"fmt"
	"sync"
	"time"
)

type VMOptions struct {
	StructuredRuntimeErrors bool
	Sandbox                 *SandboxPolicy
}

type VM struct {
	chunk         *Chunk
	ip            int
	stack         []Value
	globals       map[string]Value
	modules       map[string]Value
	modulePath    string
	resolveImport ImportResolver
	asyncRuntime  AsyncRuntime
	options       VMOptions
	env           *Environment
	exports       Object
	tryStack      []tryFrame
	sandbox       *SandboxPolicy
	stepCounter   *StepCounter
	startTime     time.Time
}

var vmPool = sync.Pool{
	New: func() any {
		return &VM{
			stack:    make([]Value, 0, 64),
			tryStack: make([]tryFrame, 0, 8),
		}
	},
}

type ImportResolver func(fromPath, moduleName string) (Value, error)

type tryFrame struct {
	catchIP   int
	catchName string
	stackBase int
}

func NewVM(chunk *Chunk, modules map[string]Value, resolve ImportResolver, modulePath string, asyncRuntime AsyncRuntime) *VM {
	return NewVMWithOptions(chunk, modules, resolve, modulePath, asyncRuntime, VMOptions{})
}

// Globals returns the VM's global variables map.
// This is primarily useful for testing.
func (v *VM) Globals() map[string]Value {
	return v.globals
}

// GetEnv returns a variable from the VM's environment, if it exists.
// This is primarily useful for testing.
func (v *VM) GetEnv(name string) (Value, bool) {
	if v.env != nil {
		val, ok := v.env.Get(name)
		return val, ok
	}
	return nil, false
}

func NewVMWithOptions(chunk *Chunk, modules map[string]Value, resolve ImportResolver, modulePath string, asyncRuntime AsyncRuntime, options VMOptions) *VM {
	globals := map[string]Value{}
	globals["print"] = NativeFunction(func(args []Value) (Value, error) {
		for _, arg := range args {
			fmt.Print(toString(arg))
		}
		fmt.Println()
		return nil, nil
	})
	if asyncRuntime == nil {
		asyncRuntime = NewGoroutineRuntime()
	}

	sandbox := options.Sandbox
	var stepCounter *StepCounter
	if sandbox != nil && sandbox.MaxSteps > 0 {
		stepCounter = NewStepCounter(sandbox.MaxSteps)
	}

	vm := &VM{
		chunk:         chunk,
		stack:         make([]Value, 0, 256),
		globals:       globals,
		modules:       modules,
		modulePath:    modulePath,
		resolveImport: resolve,
		asyncRuntime:  asyncRuntime,
		options:       options,
		env:           NewEnvironment(nil),
		exports:       Object{},
		sandbox:       sandbox,
		stepCounter:   stepCounter,
		startTime:     time.Now(),
	}
	return vm
}

func (v *VM) Run() error {
	_, err := v.runChunk()
	return err
}

func (v *VM) AutoCallMain() error {
	if v == nil {
		return fmt.Errorf("vm is nil")
	}
	mainValue, ok := v.lookupTopLevelName("main")
	if !ok {
		return nil
	}
	ret, err := v.callEntryCallable(mainValue)
	if err != nil {
		return fmt.Errorf("entry main() call error: %w", err)
	}
	if v.asyncRuntime == nil {
		return nil
	}
	if _, err := v.asyncRuntime.AwaitValue(ret); err != nil {
		return fmt.Errorf("entry main() await error: %w", err)
	}
	return nil
}

func (v *VM) Exports() Object {
	return v.exports
}

func (v *VM) lookupTopLevelName(name string) (Value, bool) {
	if v.env != nil {
		if val, ok := v.env.Get(name); ok {
			return val, true
		}
	}
	val, ok := v.globals[name]
	return val, ok
}

func (v *VM) runChunk() (Value, error) {
	for v.ip < len(v.chunk.Code) {
		// Check step counter
		if v.stepCounter != nil {
			if err := v.stepCounter.Increment(); err != nil {
				return nil, err
			}
		}
		// Check duration
		if v.sandbox != nil && v.sandbox.MaxDuration > 0 {
			if time.Since(v.startTime) > v.sandbox.MaxDuration {
				return nil, &SandboxError{
					Violation: "max duration exceeded",
					Limit:     v.sandbox.MaxDuration.String(),
					Current:   time.Since(v.startTime).String(),
				}
			}
		}

		ins := v.chunk.Code[v.ip]
		v.ip++
		var execErr error

		switch ins.Op {
		case OpConstant:
			v.push(v.chunk.Constants[ins.A])
		case OpAdd:
			execErr = v.execAdd()
		case OpSub:
			execErr = v.execSub()
		case OpMul:
			execErr = v.execMul()
		case OpDiv:
			execErr = v.execDiv()
		case OpMod:
			execErr = v.execMod()
		case OpNeg:
			execErr = v.execNeg()
		case OpNot:
			execErr = v.execNot()
		case OpAnd:
			execErr = v.execAnd()
		case OpOr:
			execErr = v.execOr()
		case OpBitAnd:
			execErr = v.execBitAnd()
		case OpBitOr:
			execErr = v.execBitOr()
		case OpBitXor:
			execErr = v.execBitXor()
		case OpShl:
			execErr = v.execShl()
		case OpShr:
			execErr = v.execShr()
		case OpTruthy:
			execErr = v.execTruthy()
		case OpDup:
			execErr = v.execDup()
		case OpEqual:
			execErr = v.execEqual()
		case OpNotEqual:
			execErr = v.execNotEqual()
		case OpGreater:
			execErr = v.execGreater()
		case OpLess:
			execErr = v.execLess()
		case OpGreaterEqual:
			execErr = v.execGreaterEqual()
		case OpLessEqual:
			execErr = v.execLessEqual()
		case OpPop:
			_, execErr = v.pop()
		case OpGetName:
			var name string
			name, execErr = v.stringConstant(ins.A, "name")
			if execErr == nil {
				var val Value
				val, execErr = v.resolveName(name)
				if execErr == nil {
					v.push(val)
				}
			}
		case OpDefineName:
			var name string
			name, execErr = v.stringConstant(ins.A, "define name")
			if execErr == nil {
				var val Value
				val, execErr = v.pop()
				if execErr == nil {
					v.defineName(name, val)
				}
			}
		case OpSetName:
			var name string
			name, execErr = v.stringConstant(ins.A, "set name")
			if execErr == nil {
				var val Value
				val, execErr = v.pop()
				if execErr == nil {
					execErr = v.setName(name, val)
				}
			}
		case OpClass:
			execErr = v.execClass(ins.A)
		case OpSetProperty:
			var prop string
			prop, execErr = v.stringConstant(ins.A, "set property")
			if execErr == nil {
				execErr = v.execSetProperty(prop)
			}
		case OpNew:
			execErr = v.execNew(ins.A)
		case OpClosure:
			var tmpl *FunctionTemplate
			tmpl, execErr = v.functionConstant(ins.A)
			if execErr == nil {
				// Bind function template to current lexical environment.
				bound := v.bindFunctionTemplate(tmpl)
				v.push(bound)
			}
		case OpGetGlobal:
			var name string
			name, execErr = v.stringConstant(ins.A, "global name")
			if execErr == nil {
				var val Value
				var exists bool
				val, exists = v.globals[name]
				if !exists {
					execErr = fmt.Errorf("undefined variable: %s", name)
				} else {
					v.push(val)
				}
			}
		case OpDefineGlobal:
			var name string
			name, execErr = v.stringConstant(ins.A, "define name")
			if execErr == nil {
				var val Value
				val, execErr = v.pop()
				if execErr == nil {
					v.globals[name] = val
				}
			}
		case OpArray:
			execErr = v.execArray(ins.A)
		case OpObject:
			execErr = v.execObject(ins.A)
		case OpSpreadArray:
			execErr = v.execSpreadArray(ins.A, ins.B)
		case OpSpreadObject:
			execErr = v.execSpreadObject(ins.A, ins.B)
		case OpGetProperty:
			var name string
			name, execErr = v.stringConstant(ins.A, "property name")
			if execErr == nil {
				execErr = v.execGetProperty(name)
			}
		case OpIndex:
			execErr = v.execIndex()
		case OpCall:
			execErr = v.execCall(ins.A)
		case OpSpreadCall:
			execErr = v.execSpreadCall(ins.A, ins.B)
		case OpAwait:
			execErr = v.execAwait()
		case OpPushTry:
			execErr = v.execPushTry(ins.A, ins.B)
		case OpPopTry:
			execErr = v.execPopTry()
		case OpThrow:
			execErr = v.execThrow()
		case OpJumpIfFalse:
			execErr = v.execJumpIfFalse(ins.A)
		case OpJumpIfTrue:
			execErr = v.execJumpIfTrue(ins.A)
		case OpJumpIfNullish:
			execErr = v.execJumpIfNullish(ins.A)
		case OpJumpIfNotNullish:
			execErr = v.execJumpIfNotNullish(ins.A)
		case OpJump:
			v.ip = ins.A
		case OpImportName:
			execErr = v.execImportName(ins.A, ins.B)
		case OpImportNamespace:
			execErr = v.execImportNamespace(ins.A, ins.B)
		case OpImportDynamic:
			execErr = v.execImportDynamic()
		case OpExportName:
			var exportName string
			exportName, execErr = v.stringConstant(ins.A, "export name")
			if execErr == nil {
				var exportVal Value
				exportVal, execErr = v.resolveName(exportName)
				if execErr == nil {
					v.exports[exportName] = exportVal
				}
			}
		case OpExportAs:
			var localName string
			localName, execErr = v.stringConstant(ins.A, "local export name")
			if execErr == nil {
				var exportName string
				exportName, execErr = v.stringConstant(ins.B, "export alias name")
				if execErr == nil {
					var exportVal Value
					exportVal, execErr = v.resolveName(localName)
					if execErr == nil {
						v.exports[exportName] = exportVal
					}
				}
			}
		case OpExportDefault:
			var exportVal Value
			exportVal, execErr = v.pop()
			if execErr == nil {
				v.exports["default"] = exportVal
			}
		case OpExportAll:
			execErr = v.execExportAll(ins.A)
		case OpSuper:
			var propName string
			propName, execErr = v.stringConstant(ins.A, "super property name")
			if execErr == nil {
				execErr = v.execSuperProperty(propName)
			}
		case OpSuperCall:
			execErr = v.execSuperCall(ins.A)
		case OpTypeof:
			execErr = v.execTypeof()
		case OpObjectKeys:
			execErr = v.execObjectKeys()
		case OpReturn:
			if ins.A == 1 {
				ret, err := v.pop()
				if err != nil {
					return nil, err
				}
				return ret, nil
			}
			return nil, nil
		default:
			execErr = fmt.Errorf("unknown opcode: %d", ins.Op)
		}

		if execErr != nil {
			contextualErr := v.attachRuntimeContext(execErr, ins.Op)
			if v.handleRuntimeError(contextualErr) {
				continue
			}
			return nil, contextualErr
		}
	}
	return nil, nil
}
