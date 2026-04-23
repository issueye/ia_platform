package runtime

import (
	"context"
	"fmt"
	"iacommon/pkg/host/api"
	"iavm/pkg/core"
)

func Interpret(vm *VM, entryFuncIndex uint32) error {
	if entryFuncIndex >= uint32(len(vm.mod.Functions)) {
		return fmt.Errorf("function index %d out of range", entryFuncIndex)
	}

	frame := NewFrame(entryFuncIndex, &vm.mod.Functions[entryFuncIndex], uint32(vm.stack.Size()))
	vm.frames = append(vm.frames, frame)

	for len(vm.frames) > 0 {
		frame = vm.frames[len(vm.frames)-1]
		fn := &vm.mod.Functions[frame.FunctionIndex]

		if frame.IP >= uint32(len(fn.Code)) {
			// Function completed, pop frame
			vm.frames = vm.frames[:len(vm.frames)-1]
			continue
		}

		vm.stepCount++
		if vm.options.MaxSteps > 0 && vm.stepCount > vm.options.MaxSteps {
			return core.ErrResourceExhausted
		}

		inst := fn.Code[frame.IP]
		frame.IP++

		if err := vm.dispatch(inst, frame); err != nil {
			if err == core.ErrUncaughtException {
				handled := false
				for len(vm.frames) > 0 {
					currentFrame := vm.frames[len(vm.frames)-1]
					if len(currentFrame.TryHandlers) > 0 {
						handlerIdx := len(currentFrame.TryHandlers) - 1
						handler := currentFrame.TryHandlers[handlerIdx]
						currentFrame.TryHandlers = currentFrame.TryHandlers[:handlerIdx]
						currentFrame.IP = handler.HandlerIP
						if handler.HasCatchVar {
							if int(handler.CatchLocalIdx) < len(currentFrame.Locals) {
								currentFrame.Locals[handler.CatchLocalIdx] = vm.exception
							} else {
								vm.stack.Push(vm.exception)
							}
						} else {
							vm.stack.Push(vm.exception)
						}
						vm.exception = core.Value{Kind: core.ValueNull}
						handled = true
						break
					}
					vm.frames = vm.frames[:len(vm.frames)-1]
				}
				if handled {
					continue
				}
				return err
			}
			return err
		}
	}

	return nil
}

func (vm *VM) dispatch(inst core.Instruction, frame *Frame) error {
	switch inst.Op {
	case core.OpNop:
		return nil

	case core.OpConst:
		val, err := vm.constantAt(frame, inst.A)
		if err != nil {
			return err
		}
		vm.stack.Push(coreValueFromAny(val))

	case core.OpLoadLocal:
		if int(inst.A) >= len(frame.Locals) {
			return fmt.Errorf("local index %d out of range", inst.A)
		}
		if frame.UpvalueMap != nil {
			if uv, ok := frame.UpvalueMap[inst.A]; ok {
				vm.stack.Push(uv.Value)
				break
			}
		}
		vm.stack.Push(frame.Locals[inst.A])

	case core.OpStoreLocal:
		if int(inst.A) >= len(frame.Locals) {
			return fmt.Errorf("local index %d out of range", inst.A)
		}
		val := vm.stack.Pop()
		frame.Locals[inst.A] = val
		if frame.UpvalueMap != nil {
			if uv, ok := frame.UpvalueMap[inst.A]; ok {
				uv.Value = val
			}
		}

	case core.OpLoadGlobal:
		if int(inst.A) >= len(vm.globals) {
			vm.stack.Push(core.Value{Kind: core.ValueNull})
		} else {
			vm.stack.Push(vm.globals[inst.A])
		}

	case core.OpStoreGlobal:
		val := vm.stack.Pop()
		if int(inst.A) >= len(vm.globals) {
			newGlobals := make([]core.Value, inst.A+1)
			copy(newGlobals, vm.globals)
			vm.globals = newGlobals
		}
		vm.globals[inst.A] = val

	case core.OpAdd:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(addValues(a, b))

	case core.OpSub:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(subValues(a, b))

	case core.OpMul:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(mulValues(a, b))

	case core.OpDiv:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(divValues(a, b))

	case core.OpMod:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(modValues(a, b))

	case core.OpNeg:
		a := vm.stack.Pop()
		vm.stack.Push(negValue(a))

	case core.OpNot:
		a := vm.stack.Pop()
		vm.stack.Push(notValue(a))

	case core.OpTruthy:
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueBool, Raw: isTruthy(a)})

	case core.OpEq:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueBool, Raw: valuesEqual(a, b)})

	case core.OpNe:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueBool, Raw: !valuesEqual(a, b)})

	case core.OpLt:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueBool, Raw: valuesLess(a, b)})

	case core.OpGt:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueBool, Raw: valuesLess(b, a)})

	case core.OpLe:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueBool, Raw: valuesEqual(a, b) || valuesLess(a, b)})

	case core.OpGe:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueBool, Raw: valuesEqual(a, b) || valuesLess(b, a)})

	case core.OpJump:
		frame.IP = inst.A

	case core.OpJumpIfFalse:
		val := vm.stack.Pop()
		if !isTruthy(val) {
			frame.IP = inst.A
		}

	case core.OpJumpIfTrue:
		val := vm.stack.Pop()
		if isTruthy(val) {
			frame.IP = inst.A
		}

	case core.OpClosure:
		fnIdx := inst.A
		if int(fnIdx) < len(vm.mod.Functions) {
			fn := &vm.mod.Functions[fnIdx]
			if len(fn.Captures) > 0 {
				upvalues := make([]*core.Upvalue, len(fn.Captures))
				for i, outerLocalIdx := range fn.Captures {
					val := core.Value{Kind: core.ValueNull}
					if int(outerLocalIdx) < len(frame.Locals) {
						val = frame.Locals[outerLocalIdx]
					}
					upvalues[i] = &core.Upvalue{Value: val}

					if frame.UpvalueMap == nil {
						frame.UpvalueMap = make(map[uint32]*core.Upvalue)
					}
					frame.UpvalueMap[outerLocalIdx] = upvalues[i]
				}
				vm.stack.Push(core.Value{Kind: core.ValueFuncRef, Raw: fnIdx, Upvalues: upvalues})
			} else {
				vm.stack.Push(core.Value{Kind: core.ValueFuncRef, Raw: fnIdx})
			}
		} else {
			vm.stack.Push(core.Value{Kind: core.ValueFuncRef, Raw: fnIdx})
		}

	case core.OpJumpIfNullish:
		val := vm.stack.Peek(0)
		if val.Kind == core.ValueNull {
			vm.stack.Pop()
			frame.IP = inst.A
		}

	case core.OpJumpIfNotNullish:
		val := vm.stack.Peek(0)
		if val.Kind != core.ValueNull {
			frame.IP = inst.A
		}

	case core.OpCall:
		if inst.B > 0 {
			// Direct function call: A = function index, B = arg count
			fnIdx := inst.A
			if int(fnIdx) >= len(vm.mod.Functions) {
				return fmt.Errorf("function index %d out of range", fnIdx)
			}
			targetFn := &vm.mod.Functions[fnIdx]
			newFrame := NewFrame(fnIdx, targetFn, uint32(vm.stack.Size()))
			argCount := int(inst.B)
			for i := argCount - 1; i >= 0; i-- {
				if i < len(newFrame.Locals) {
					newFrame.Locals[i] = vm.stack.Pop()
				}
			}
			vm.frames = append(vm.frames, newFrame)
		} else {
			// Stack-based call: function is UNDER args on stack
			// Stack layout: [..., function, arg1, arg2, ...]
			argCount := int(inst.A)

			// Pop args first
			args := make([]core.Value, argCount)
			for i := argCount - 1; i >= 0; i-- {
				args[i] = vm.stack.Pop()
			}

			// Now pop function reference
			fnRef := vm.stack.Pop()
			var fnIdx uint32
			switch fnRef.Kind {
			case core.ValueFuncRef:
				fnIdx = fnRef.Raw.(uint32)
			case core.ValueI64:
				fnIdx = uint32(fnRef.Raw.(int64))
			case core.ValueString:
				// Builtin function name
				name := fnRef.Raw.(string)
				builtin, ok := vm.builtins[name]
				if !ok {
					return fmt.Errorf("builtin function not found: %s", name)
				}
				result := builtin(args)
				vm.stack.Push(result)
				return nil
			default:
				return fmt.Errorf("cannot call value of kind %v", fnRef.Kind)
			}
			if int(fnIdx) >= len(vm.mod.Functions) {
				return fmt.Errorf("function index %d out of range", fnIdx)
			}
			targetFn := &vm.mod.Functions[fnIdx]
			newFrame := NewFrame(fnIdx, targetFn, uint32(vm.stack.Size()))
			for i := 0; i < argCount; i++ {
				if i < len(newFrame.Locals) {
					newFrame.Locals[i] = args[i]
				}
			}
			// Propagate upvalues from closure to new frame
			if len(fnRef.Upvalues) > 0 {
				if newFrame.UpvalueMap == nil {
					newFrame.UpvalueMap = make(map[uint32]*core.Upvalue)
				}
				for i, uv := range fnRef.Upvalues {
					if i < len(newFrame.Locals) {
						newFrame.UpvalueMap[uint32(i)] = uv
					}
				}
			}
			vm.frames = append(vm.frames, newFrame)
		}

	case core.OpReturn:
		vm.frames = vm.frames[:len(vm.frames)-1]

	case core.OpMakeArray:
		count := int(inst.A)
		arr := make([]core.Value, count)
		for i := count - 1; i >= 0; i-- {
			arr[i] = vm.stack.Pop()
		}
		vm.stack.Push(core.Value{Kind: core.ValueArrayRef, Raw: arr})

	case core.OpMakeObject:
		m := make(map[string]core.Value)
		if inst.A > 0 {
			// Pop key-value pairs: stack has [..., key1, val1, key2, val2, ...]
			// Popping order: val2, key2, val1, key1
			for i := 0; i < int(inst.A); i++ {
				val := vm.stack.Pop()
				keyVal := vm.stack.Pop()
				if keyVal.Kind == core.ValueString {
					m[keyVal.Raw.(string)] = val
				}
			}
		}
		vm.stack.Push(core.Value{Kind: core.ValueObjectRef, Raw: m})

	case core.OpGetProp:
		obj := vm.stack.Pop()
		if obj.Kind != core.ValueObjectRef {
			return fmt.Errorf("cannot get property from non-object")
		}
		nameVal, err := vm.constantAt(frame, inst.A)
		if err != nil {
			return fmt.Errorf("property name constant lookup failed: %w", err)
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("property name at index %d is not a string", inst.A)
		}
		m := obj.Raw.(map[string]core.Value)
		vm.stack.Push(m[name])

	case core.OpSetProp:
		val := vm.stack.Pop()
		obj := vm.stack.Pop()
		if obj.Kind != core.ValueObjectRef {
			return fmt.Errorf("cannot set property on non-object")
		}
		nameVal, err := vm.constantAt(frame, inst.A)
		if err != nil {
			return fmt.Errorf("property name constant lookup failed: %w", err)
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("property name at index %d is not a string", inst.A)
		}
		m := obj.Raw.(map[string]core.Value)
		m[name] = val

	case core.OpImportFunc:
		// Simplified: push a placeholder function reference
		vm.stack.Push(core.Value{Kind: core.ValueFuncRef, Raw: inst.A})

	case core.OpImportCap:
		if vm.options.Host == nil {
			return fmt.Errorf("no host configured for capability import")
		}
		// inst.A = capability kind index, inst.B = config index
		fn := &vm.mod.Functions[frame.FunctionIndex]
		var capKind string
		if int(inst.A) < len(fn.Constants) {
			if s, ok := fn.Constants[inst.A].(string); ok {
				capKind = s
			}
		}
		cap, err := vm.options.Host.AcquireCapability(context.Background(), api.AcquireRequest{
			Kind:   api.CapabilityKind(capKind),
			Config: map[string]any{},
		})
		if err != nil {
			return fmt.Errorf("failed to acquire capability: %w", err)
		}
		// Store capability ID for later use
		if vm.capabilityIDs == nil {
			vm.capabilityIDs = make(map[uint32]string)
		}
		vm.capabilityIDs[inst.A] = cap.ID
		vm.stack.Push(core.Value{Kind: core.ValueHostHandle, Raw: cap.ID})

	case core.OpHostCall:
		if vm.options.Host == nil {
			return fmt.Errorf("no host configured for host.call")
		}
		// Pop operation name from stack
		opVal := vm.stack.Pop()
		if opVal.Kind != core.ValueString {
			return fmt.Errorf("host.call operation must be string")
		}
		opName := opVal.Raw.(string)

		// Get capability ID from the last imported capability
		var capID string
		for _, id := range vm.capabilityIDs {
			capID = id
			break
		}

		req := api.CallRequest{
			CapabilityID: capID,
			Operation:    opName,
			Args:         map[string]any{},
		}

		result, err := vm.options.Host.Call(context.Background(), req)
		if err != nil {
			return fmt.Errorf("host.call failed: %w", err)
		}

		if result.Value != nil {
			vm.stack.Push(core.Value{Kind: core.ValueObjectRef, Raw: result.Value})
		} else {
			vm.stack.Push(core.Value{Kind: core.ValueNull})
		}

	case core.OpHostPoll:
		vm.stack.Push(core.Value{Kind: core.ValueNull})

	case core.OpDup:
		val := vm.stack.Peek(0)
		vm.stack.Push(val)

	case core.OpPop:
		vm.stack.Pop()

	case core.OpBitAnd:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(bitAndValues(a, b))

	case core.OpBitOr:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(bitOrValues(a, b))

	case core.OpBitXor:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(bitXorValues(a, b))

	case core.OpShl:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(shlValues(a, b))

	case core.OpShr:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(shrValues(a, b))

	case core.OpAnd:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueBool, Raw: isTruthy(a) && isTruthy(b)})

	case core.OpOr:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueBool, Raw: isTruthy(a) || isTruthy(b)})

	case core.OpTypeof:
		a := vm.stack.Pop()
		vm.stack.Push(core.Value{Kind: core.ValueString, Raw: typeOfValue(a)})

	case core.OpObjectKeys:
		a := vm.stack.Pop()
		if a.Kind != core.ValueObjectRef {
			vm.stack.Push(core.Value{Kind: core.ValueArrayRef, Raw: []core.Value{}})
		} else {
			m := a.Raw.(map[string]core.Value)
			keys := make([]core.Value, 0, len(m))
			for k := range m {
				keys = append(keys, core.Value{Kind: core.ValueString, Raw: k})
			}
			vm.stack.Push(core.Value{Kind: core.ValueArrayRef, Raw: keys})
		}

	case core.OpPushTry:
		handler := TryHandler{HandlerIP: inst.A}
		if inst.B > 0 {
			handler.CatchLocalIdx = inst.B - 1
			handler.HasCatchVar = true
		}
		frame.TryHandlers = append(frame.TryHandlers, handler)

	case core.OpPopTry:
		if len(frame.TryHandlers) > 0 {
			frame.TryHandlers = frame.TryHandlers[:len(frame.TryHandlers)-1]
		}

	case core.OpIndex:
		indexVal := vm.stack.Pop()
		targetVal := vm.stack.Pop()
		switch targetVal.Kind {
		case core.ValueArrayRef:
			idx, ok := toIntIndex(indexVal)
			if !ok {
				return fmt.Errorf("array index must be numeric, got %v", indexVal.Kind)
			}
			arr := targetVal.Raw.([]core.Value)
			if idx < 0 || idx >= len(arr) {
				vm.stack.Push(core.Value{Kind: core.ValueNull})
			} else {
				vm.stack.Push(arr[idx])
			}
		case core.ValueObjectRef:
			if indexVal.Kind != core.ValueString {
				return fmt.Errorf("object index must be string, got %v", indexVal.Kind)
			}
			key := indexVal.Raw.(string)
			m := targetVal.Raw.(map[string]core.Value)
			if val, ok := m[key]; ok {
				vm.stack.Push(val)
			} else {
				vm.stack.Push(core.Value{Kind: core.ValueNull})
			}
		case core.ValueString:
			idx, ok := toIntIndex(indexVal)
			if !ok {
				return fmt.Errorf("string index must be numeric, got %v", indexVal.Kind)
			}
			s := targetVal.Raw.(string)
			if idx < 0 || idx >= len(s) {
				vm.stack.Push(core.Value{Kind: core.ValueNull})
			} else {
				vm.stack.Push(core.Value{Kind: core.ValueString, Raw: string(s[idx])})
			}
		default:
			return fmt.Errorf("index operator not supported on %v", targetVal.Kind)
		}

	case core.OpThrow:
		vm.exception = vm.stack.Pop()
		return core.ErrUncaughtException

	default:
		return fmt.Errorf("unimplemented opcode: %v", inst.Op)
	}

	return nil
}

func (vm *VM) constantAt(frame *Frame, index uint32) (any, error) {
	if len(vm.mod.Constants) > 0 {
		if int(index) >= len(vm.mod.Constants) {
			return nil, fmt.Errorf("module constant index %d out of range (constants: %d)", index, len(vm.mod.Constants))
		}
		return vm.mod.Constants[index], nil
	}

	fn := &vm.mod.Functions[frame.FunctionIndex]
	if int(index) >= len(fn.Constants) {
		return nil, fmt.Errorf("constant index %d out of range", index)
	}
	return fn.Constants[index], nil
}

func coreValueFromAny(v any) core.Value {
	if v == nil {
		return core.Value{Kind: core.ValueNull}
	}
	switch val := v.(type) {
	case core.Value:
		return val
	case bool:
		return core.Value{Kind: core.ValueBool, Raw: val}
	case int:
		return core.Value{Kind: core.ValueI64, Raw: int64(val)}
	case int64:
		return core.Value{Kind: core.ValueI64, Raw: val}
	case float64:
		return core.Value{Kind: core.ValueF64, Raw: val}
	case string:
		return core.Value{Kind: core.ValueString, Raw: val}
	default:
		return core.Value{Kind: core.ValueNull}
	}
}

func toFloat(v core.Value) (float64, bool) {
	switch v.Kind {
	case core.ValueF64:
		return v.Raw.(float64), true
	case core.ValueI64:
		return float64(v.Raw.(int64)), true
	}
	return 0, false
}

func toIntIndex(v core.Value) (int, bool) {
	switch v.Kind {
	case core.ValueI64:
		return int(v.Raw.(int64)), true
	case core.ValueF64:
		return int(v.Raw.(float64)), true
	}
	return 0, false
}

func addValues(a, b core.Value) core.Value {
	if a.Kind == core.ValueString || b.Kind == core.ValueString {
		return core.Value{Kind: core.ValueString, Raw: valueToString(a) + valueToString(b)}
	}
	if a.Kind == core.ValueI64 && b.Kind == core.ValueI64 {
		return core.Value{Kind: core.ValueI64, Raw: a.Raw.(int64) + b.Raw.(int64)}
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return core.Value{Kind: core.ValueF64, Raw: af + bf}
	}
	return core.Value{Kind: core.ValueNull}
}

func subValues(a, b core.Value) core.Value {
	if a.Kind == core.ValueI64 && b.Kind == core.ValueI64 {
		return core.Value{Kind: core.ValueI64, Raw: a.Raw.(int64) - b.Raw.(int64)}
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return core.Value{Kind: core.ValueF64, Raw: af - bf}
	}
	return core.Value{Kind: core.ValueNull}
}

func mulValues(a, b core.Value) core.Value {
	if a.Kind == core.ValueI64 && b.Kind == core.ValueI64 {
		return core.Value{Kind: core.ValueI64, Raw: a.Raw.(int64) * b.Raw.(int64)}
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return core.Value{Kind: core.ValueF64, Raw: af * bf}
	}
	return core.Value{Kind: core.ValueNull}
}

func divValues(a, b core.Value) core.Value {
	if a.Kind == core.ValueI64 && b.Kind == core.ValueI64 {
		bv := b.Raw.(int64)
		if bv == 0 {
			return core.Value{Kind: core.ValueNull}
		}
		return core.Value{Kind: core.ValueI64, Raw: a.Raw.(int64) / bv}
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return core.Value{Kind: core.ValueF64, Raw: af / bf}
	}
	return core.Value{Kind: core.ValueNull}
}

func modValues(a, b core.Value) core.Value {
	if a.Kind == core.ValueI64 && b.Kind == core.ValueI64 {
		bv := b.Raw.(int64)
		if bv == 0 {
			return core.Value{Kind: core.ValueNull}
		}
		return core.Value{Kind: core.ValueI64, Raw: a.Raw.(int64) % bv}
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return core.Value{Kind: core.ValueF64, Raw: float64(int64(af) % int64(bf))}
	}
	return core.Value{Kind: core.ValueNull}
}

func negValue(a core.Value) core.Value {
	if a.Kind == core.ValueI64 {
		return core.Value{Kind: core.ValueI64, Raw: -a.Raw.(int64)}
	}
	if a.Kind == core.ValueF64 {
		return core.Value{Kind: core.ValueF64, Raw: -a.Raw.(float64)}
	}
	return core.Value{Kind: core.ValueNull}
}

func notValue(a core.Value) core.Value {
	return core.Value{Kind: core.ValueBool, Raw: !isTruthy(a)}
}

func valuesEqual(a, b core.Value) bool {
	if a.Kind == b.Kind {
		switch a.Kind {
		case core.ValueNull:
			return true
		case core.ValueBool:
			return a.Raw.(bool) == b.Raw.(bool)
		case core.ValueI64:
			return a.Raw.(int64) == b.Raw.(int64)
		case core.ValueF64:
			return a.Raw.(float64) == b.Raw.(float64)
		case core.ValueString:
			return a.Raw.(string) == b.Raw.(string)
		default:
			return a.Raw == b.Raw
		}
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return af == bf
	}
	return false
}

func valuesLess(a, b core.Value) bool {
	if a.Kind == core.ValueString && b.Kind == core.ValueString {
		return a.Raw.(string) < b.Raw.(string)
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return af < bf
	}
	return false
}

func isTruthy(val core.Value) bool {
	switch val.Kind {
	case core.ValueNull:
		return false
	case core.ValueBool:
		return val.Raw.(bool)
	case core.ValueI64:
		return val.Raw.(int64) != 0
	case core.ValueF64:
		return val.Raw.(float64) != 0
	case core.ValueString:
		return val.Raw.(string) != ""
	default:
		return true
	}
}

func toInt64Bitwise(v core.Value) (int64, bool) {
	switch v.Kind {
	case core.ValueI64:
		return v.Raw.(int64), true
	case core.ValueF64:
		return int64(v.Raw.(float64)), true
	}
	return 0, false
}

func bitAndValues(a, b core.Value) core.Value {
	ai, aok := toInt64Bitwise(a)
	bi, bok := toInt64Bitwise(b)
	if aok && bok {
		return core.Value{Kind: core.ValueI64, Raw: ai & bi}
	}
	return core.Value{Kind: core.ValueNull}
}

func bitOrValues(a, b core.Value) core.Value {
	ai, aok := toInt64Bitwise(a)
	bi, bok := toInt64Bitwise(b)
	if aok && bok {
		return core.Value{Kind: core.ValueI64, Raw: ai | bi}
	}
	return core.Value{Kind: core.ValueNull}
}

func bitXorValues(a, b core.Value) core.Value {
	ai, aok := toInt64Bitwise(a)
	bi, bok := toInt64Bitwise(b)
	if aok && bok {
		return core.Value{Kind: core.ValueI64, Raw: ai ^ bi}
	}
	return core.Value{Kind: core.ValueNull}
}

func shlValues(a, b core.Value) core.Value {
	ai, aok := toInt64Bitwise(a)
	bi, bok := toInt64Bitwise(b)
	if aok && bok {
		return core.Value{Kind: core.ValueI64, Raw: ai << bi}
	}
	return core.Value{Kind: core.ValueNull}
}

func shrValues(a, b core.Value) core.Value {
	ai, aok := toInt64Bitwise(a)
	bi, bok := toInt64Bitwise(b)
	if aok && bok {
		return core.Value{Kind: core.ValueI64, Raw: ai >> bi}
	}
	return core.Value{Kind: core.ValueNull}
}

func typeOfValue(a core.Value) string {
	switch a.Kind {
	case core.ValueNull:
		return "null"
	case core.ValueBool:
		return "boolean"
	case core.ValueI64, core.ValueF64:
		return "number"
	case core.ValueString:
		return "string"
	case core.ValueArrayRef:
		return "array"
	case core.ValueObjectRef:
		return "object"
	case core.ValueFuncRef:
		return "function"
	case core.ValueHostHandle:
		return "handle"
	default:
		return "unknown"
	}
}

func valueToString(a core.Value) string {
	switch a.Kind {
	case core.ValueNull:
		return "null"
	case core.ValueBool:
		if a.Raw.(bool) {
			return "true"
		}
		return "false"
	case core.ValueI64:
		return fmt.Sprintf("%d", a.Raw.(int64))
	case core.ValueF64:
		return fmt.Sprintf("%v", a.Raw.(float64))
	case core.ValueString:
		return a.Raw.(string)
	default:
		return fmt.Sprintf("%v", a.Kind)
	}
}
