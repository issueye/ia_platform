package runtime

import (
	"fmt"
	"iacommon/pkg/host/api"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

const (
	classKindKey        = "__iavm_kind"
	classKindClass      = "class"
	classKindInstance   = "instance"
	classNameKey        = "__iavm_name"
	classParentKey      = "__iavm_parent"
	classMethodsKey     = "__iavm_methods"
	classStaticKey      = "__iavm_static_methods"
	classGettersKey     = "__iavm_getters"
	classSettersKey     = "__iavm_setters"
	instanceClassRefKey = "__iavm_class"
)

func Interpret(vm *VM, entryFuncIndex uint32) error {
	var frame *Frame
	if len(vm.frames) == 0 {
		if entryFuncIndex >= uint32(len(vm.mod.Functions)) {
			return fmt.Errorf("function index %d out of range", entryFuncIndex)
		}

		frame = NewFrame(entryFuncIndex, &vm.mod.Functions[entryFuncIndex], uint32(vm.stack.Size()))
		vm.frames = append(vm.frames, frame)
	}

	for len(vm.frames) > 0 {
		if err := vm.hostContext().Err(); err != nil {
			return err
		}
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
			argCount := int(inst.B)
			args := make([]core.Value, argCount)
			for i := argCount - 1; i >= 0; i-- {
				args[i] = vm.stack.Pop()
			}
			if targetFn.Async {
				result, err := vm.runFunctionSync(fnIdx, args, core.Value{Kind: core.ValueFuncRef, Raw: fnIdx})
				if err != nil {
					return err
				}
				vm.stack.Push(resolvedPromiseValue(result))
				break
			}
			newFrame := NewFrame(fnIdx, targetFn, uint32(vm.stack.Size()))
			for i := 0; i < argCount && i < len(newFrame.Locals); i++ {
				newFrame.Locals[i] = args[i]
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

			fnRef := vm.stack.Pop()
			if err := vm.invokeCallable(fnRef, args); err != nil {
				return err
			}
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

	case core.OpClass:
		classVal, err := vm.buildClassValue(inst.A)
		if err != nil {
			return err
		}
		vm.stack.Push(classVal)

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
		if val, ok := vm.getObjectProperty(obj, name); ok {
			vm.stack.Push(val)
		} else {
			vm.stack.Push(core.Value{Kind: core.ValueNull})
		}

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
		if err := vm.setObjectProperty(obj, name, val); err != nil {
			return err
		}

	case core.OpImportFunc:
		// Simplified: push a placeholder function reference
		vm.stack.Push(core.Value{Kind: core.ValueFuncRef, Raw: inst.A})

	case core.OpImportCap:
		if vm.options.Host == nil {
			return fmt.Errorf("no host configured for capability import")
		}
		// inst.A = capability kind index, inst.B = config index
		fn := &vm.mod.Functions[frame.FunctionIndex]
		capKind, ok := vm.resolveStringConstant(fn, inst.A)
		if !ok || capKind == "" {
			return fmt.Errorf("capability import kind must reference a string constant")
		}
		config := vm.capabilityConfig(module.CapabilityKind(capKind))
		timeoutProfile := vm.capabilityTimeoutProfile(module.CapabilityKind(capKind))
		hostCtx, cancel := vm.hostOperationContext(vm.hostContext(), vm.options.HostTimeout)
		if timeoutProfile.HostTimeout > 0 {
			hostCtx, cancel = vm.hostOperationContext(vm.hostContext(), timeoutProfile.HostTimeout)
		}
		cap, err := vm.options.Host.AcquireCapability(hostCtx, api.AcquireRequest{
			Kind:   api.CapabilityKind(capKind),
			Config: config,
		})
		cancel()
		if err != nil {
			return fmt.Errorf("failed to acquire capability: %w", err)
		}
		// Store capability ID for later use
		if vm.capabilityIDs == nil {
			vm.capabilityIDs = make(map[uint32]string)
		}
		vm.capabilityIDs[inst.A] = cap.ID
		vm.lastCapabilityID = cap.ID
		vm.lastCapabilityKind = module.CapabilityKind(capKind)
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
		callArgs := vm.popHostCallArgs(int(inst.A))

		if vm.lastCapabilityID == "" {
			return fmt.Errorf("no capability imported for host.call")
		}

		req := api.CallRequest{
			CapabilityID: vm.lastCapabilityID,
			Operation:    opName,
			Args:         callArgs,
		}

		timeoutProfile := vm.capabilityTimeoutProfile(vm.lastCapabilityKind)
		hostCtx, cancel := vm.hostOperationContext(vm.hostContext(), vm.options.HostTimeout)
		if timeoutProfile.HostTimeout > 0 {
			hostCtx, cancel = vm.hostOperationContext(vm.hostContext(), timeoutProfile.HostTimeout)
		}
		result, err := vm.options.Host.Call(hostCtx, req)
		cancel()
		if err != nil {
			return fmt.Errorf("host.call failed: %w", err)
		}

		if result.Value != nil {
			vm.stack.Push(core.Value{Kind: core.ValueObjectRef, Raw: result.Value})
		} else {
			vm.stack.Push(core.Value{Kind: core.ValueNull})
		}

	case core.OpHostPoll:
		if vm.options.Host == nil {
			return fmt.Errorf("no host configured for host.poll")
		}
		handleVal := vm.stack.Pop()
		handleID, err := hostHandleID(handleVal)
		if err != nil {
			return err
		}
		timeoutProfile := vm.capabilityTimeoutProfile(vm.lastCapabilityKind)
		hostCtx, cancel := vm.hostOperationContext(vm.hostContext(), vm.options.HostTimeout)
		if timeoutProfile.HostTimeout > 0 {
			hostCtx, cancel = vm.hostOperationContext(vm.hostContext(), timeoutProfile.HostTimeout)
		}
		result, err := vm.options.Host.Poll(hostCtx, handleID)
		cancel()
		if err != nil {
			return fmt.Errorf("host.poll failed: %w", err)
		}
		vm.stack.Push(promiseValueFromHostPoll(handleID, coreValueFromHostPoll(result), result.Done, result.Error, timeoutProfile.HostTimeout, timeoutProfile.WaitTimeout))

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
			if val, ok := vm.getObjectProperty(targetVal, key); ok {
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

	case core.OpNewInstance:
		argCount := int(inst.A)
		args := make([]core.Value, argCount)
		for i := argCount - 1; i >= 0; i-- {
			args[i] = vm.stack.Pop()
		}
		classVal := vm.stack.Pop()
		instance, ctor, err := vm.instantiateClass(classVal)
		if err != nil {
			return err
		}
		vm.stack.Push(instance)
		if ctor.Kind != core.ValueNull {
			if err := vm.invokeCallable(bindReceiver(ctor, instance), args); err != nil {
				return err
			}
		}

	case core.OpSuper:
		propNameVal, err := vm.constantAt(frame, inst.A)
		if err != nil {
			return err
		}
		propName, ok := propNameVal.(string)
		if !ok {
			return fmt.Errorf("super property name at index %d is not a string", inst.A)
		}
		method, err := vm.lookupSuperMethod(frame, propName)
		if err != nil {
			return err
		}
		vm.stack.Push(method)

	case core.OpSuperCall:
		argCount := int(inst.A)
		args := make([]core.Value, argCount)
		for i := argCount - 1; i >= 0; i-- {
			args[i] = vm.stack.Pop()
		}
		result, err := vm.invokeSuperConstructor(frame, args)
		if err != nil {
			return err
		}
		vm.stack.Push(result)

	case core.OpAwait:
		val := vm.stack.Pop()
		resolved, err := awaitValue(val)
		if err != nil {
			if err == ErrPromisePending {
				vm.suspension = &Suspension{
					Reason:     "await_pending_promise",
					AwaitValue: val,
					FrameDepth: len(vm.frames),
				}
			}
			return err
		}
		vm.stack.Push(resolved)

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

func (vm *VM) invokeCallable(fnRef core.Value, args []core.Value) error {
	var fnIdx uint32
	switch fnRef.Kind {
	case core.ValueFuncRef:
		fnIdx = fnRef.Raw.(uint32)
	case core.ValueI64:
		fnIdx = uint32(fnRef.Raw.(int64))
	case core.ValueString:
		name := fnRef.Raw.(string)
		builtin, ok := vm.builtins[name]
		if !ok {
			return fmt.Errorf("builtin function not found: %s", name)
		}
		vm.stack.Push(builtin(args))
		return nil
	default:
		return fmt.Errorf("cannot call value of kind %v", fnRef.Kind)
	}
	if int(fnIdx) < len(vm.mod.Functions) && vm.mod.Functions[fnIdx].Async {
		result, err := vm.runFunctionSync(fnIdx, args, fnRef)
		if err != nil {
			return err
		}
		vm.stack.Push(resolvedPromiseValue(result))
		return nil
	}
	return vm.pushCallFrame(fnIdx, args, fnRef)
}

func (vm *VM) pushCallFrame(fnIdx uint32, args []core.Value, fnRef core.Value) error {
	if int(fnIdx) >= len(vm.mod.Functions) {
		return fmt.Errorf("function index %d out of range", fnIdx)
	}
	targetFn := &vm.mod.Functions[fnIdx]
	newFrame := NewFrame(fnIdx, targetFn, uint32(vm.stack.Size()))

	if fnRef.BoundReceiver != nil && targetFn.HasThis && int(targetFn.ThisLocal) < len(newFrame.Locals) {
		newFrame.Locals[targetFn.ThisLocal] = *fnRef.BoundReceiver
	}

	localIdx := 0
	for _, arg := range args {
		for targetFn.HasThis && localIdx == int(targetFn.ThisLocal) {
			localIdx++
		}
		if localIdx < len(newFrame.Locals) {
			newFrame.Locals[localIdx] = arg
		}
		localIdx++
	}

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
	return nil
}

func bindReceiver(fn core.Value, receiver core.Value) core.Value {
	if fn.Kind != core.ValueFuncRef {
		return fn
	}
	bound := receiver
	fn.BoundReceiver = &bound
	return fn
}

func (vm *VM) buildClassValue(encoded uint32) (core.Value, error) {
	privateFieldCount := int((encoded >> 16) & 0xF)
	hasParent := (encoded >> 20) & 1
	instanceMethodCount := int(encoded & 0xF)
	staticMethodCount := int((encoded >> 4) & 0xF)
	getterCount := int((encoded >> 8) & 0xF)
	setterCount := int((encoded >> 12) & 0xF)

	privateFields := make([]string, 0, privateFieldCount)
	for i := 0; i < privateFieldCount; i++ {
		fieldVal := vm.stack.Pop()
		if fieldVal.Kind != core.ValueString {
			return core.Value{}, fmt.Errorf("private field name must be string, got %v", fieldVal.Kind)
		}
		privateFields = append(privateFields, fieldVal.Raw.(string))
	}

	setters, err := vm.popClassMethodTable(setterCount)
	if err != nil {
		return core.Value{}, err
	}
	getters, err := vm.popClassMethodTable(getterCount)
	if err != nil {
		return core.Value{}, err
	}
	staticMethods, err := vm.popClassMethodTable(staticMethodCount)
	if err != nil {
		return core.Value{}, err
	}
	instanceMethods, err := vm.popClassMethodTable(instanceMethodCount)
	if err != nil {
		return core.Value{}, err
	}

	classNameVal := vm.stack.Pop()
	if classNameVal.Kind != core.ValueString {
		return core.Value{}, fmt.Errorf("class name must be string, got %v", classNameVal.Kind)
	}

	parent := core.Value{Kind: core.ValueNull}
	if hasParent == 1 {
		parentNameVal := vm.stack.Pop()
		if parentNameVal.Kind != core.ValueString {
			return core.Value{}, fmt.Errorf("parent class name must be string, got %v", parentNameVal.Kind)
		}
		parent = vm.lookupGlobalByName(parentNameVal.Raw.(string))
		if !isClassValue(parent) {
			return core.Value{}, fmt.Errorf("parent class not found: %s", parentNameVal.Raw.(string))
		}
	}

	classMap := map[string]core.Value{
		classKindKey:    {Kind: core.ValueString, Raw: classKindClass},
		classNameKey:    classNameVal,
		classParentKey:  parent,
		classMethodsKey: {Kind: core.ValueObjectRef, Raw: instanceMethods},
		classStaticKey:  {Kind: core.ValueObjectRef, Raw: staticMethods},
		classGettersKey: {Kind: core.ValueObjectRef, Raw: getters},
		classSettersKey: {Kind: core.ValueObjectRef, Raw: setters},
	}
	for _, field := range privateFields {
		classMap[field] = core.Value{Kind: core.ValueNull}
	}
	return core.Value{Kind: core.ValueObjectRef, Raw: classMap}, nil
}

func (vm *VM) popClassMethodTable(count int) (map[string]core.Value, error) {
	table := make(map[string]core.Value, count)
	for i := 0; i < count; i++ {
		methodVal := vm.stack.Pop()
		nameVal := vm.stack.Pop()
		if nameVal.Kind != core.ValueString {
			return nil, fmt.Errorf("class method name must be string, got %v", nameVal.Kind)
		}
		if methodVal.Kind != core.ValueFuncRef {
			return nil, fmt.Errorf("class method must be function, got %v", methodVal.Kind)
		}
		table[nameVal.Raw.(string)] = methodVal
	}
	return table, nil
}

func (vm *VM) lookupGlobalByName(name string) core.Value {
	for i, g := range vm.mod.Globals {
		if g.Name == name && i < len(vm.globals) {
			return vm.globals[i]
		}
	}
	return core.Value{Kind: core.ValueNull}
}

func (vm *VM) getObjectProperty(obj core.Value, name string) (core.Value, bool) {
	switch m := obj.Raw.(type) {
	case map[string]core.Value:
		if isInstanceValue(obj) {
			if val, ok := m[name]; ok && !isReservedObjectKey(name) {
				return val, true
			}
			classVal := m[instanceClassRefKey]
			if getter, ok := lookupClassAccessor(classVal, classGettersKey, name); ok {
				return bindReceiver(getter, obj), true
			}
			if method, ok := lookupClassAccessor(classVal, classMethodsKey, name); ok {
				return bindReceiver(method, obj), true
			}
			return core.Value{Kind: core.ValueNull}, false
		}
		if isClassValue(obj) {
			if method, ok := lookupClassAccessor(obj, classStaticKey, name); ok {
				return method, true
			}
		}
		if val, ok := m[name]; ok {
			return val, true
		}
		return core.Value{Kind: core.ValueNull}, false
	case map[string]any:
		val, ok := m[name]
		if !ok {
			return core.Value{Kind: core.ValueNull}, false
		}
		return coreValueFromAny(val), true
	default:
		return core.Value{Kind: core.ValueNull}, false
	}
}

func (vm *VM) setObjectProperty(obj core.Value, name string, val core.Value) error {
	switch m := obj.Raw.(type) {
	case map[string]core.Value:
		if isInstanceValue(obj) {
			classVal := m[instanceClassRefKey]
			if setter, ok := lookupClassAccessor(classVal, classSettersKey, name); ok {
				return vm.invokeCallable(bindReceiver(setter, obj), []core.Value{val})
			}
			m[name] = val
			return nil
		}
		m[name] = val
		return nil
	case map[string]any:
		m[name] = coreValueToHostAny(val)
		return nil
	default:
		return fmt.Errorf("cannot set property on unsupported object backing")
	}
}

func (vm *VM) instantiateClass(classVal core.Value) (core.Value, core.Value, error) {
	if !isClassValue(classVal) {
		return core.Value{}, core.Value{}, fmt.Errorf("cannot instantiate non-class value")
	}
	instanceMap := map[string]core.Value{
		classKindKey:        {Kind: core.ValueString, Raw: classKindInstance},
		instanceClassRefKey: classVal,
	}
	instance := core.Value{Kind: core.ValueObjectRef, Raw: instanceMap}
	ctor, _ := lookupClassAccessor(classVal, classMethodsKey, "constructor")
	return instance, ctor, nil
}

func isReservedObjectKey(name string) bool {
	switch name {
	case classKindKey, classNameKey, classParentKey, classMethodsKey, classStaticKey, classGettersKey, classSettersKey, instanceClassRefKey:
		return true
	default:
		return false
	}
}

func isClassValue(v core.Value) bool {
	return objectKind(v) == classKindClass
}

func isInstanceValue(v core.Value) bool {
	return objectKind(v) == classKindInstance
}

func objectKind(v core.Value) string {
	if v.Kind != core.ValueObjectRef {
		return ""
	}
	m, ok := v.Raw.(map[string]core.Value)
	if !ok {
		return ""
	}
	if kind, ok := m[classKindKey]; ok && kind.Kind == core.ValueString {
		return kind.Raw.(string)
	}
	return ""
}

func lookupClassAccessor(classVal core.Value, tableKey, name string) (core.Value, bool) {
	current := classVal
	for isClassValue(current) {
		classMap := current.Raw.(map[string]core.Value)
		if tableVal, ok := classMap[tableKey]; ok && tableVal.Kind == core.ValueObjectRef {
			if method, ok := tableVal.Raw.(map[string]core.Value)[name]; ok {
				return method, true
			}
		}
		parent, ok := classMap[classParentKey]
		if !ok || parent.Kind == core.ValueNull {
			break
		}
		current = parent
	}
	return core.Value{}, false
}

func (vm *VM) lookupSuperMethod(frame *Frame, name string) (core.Value, error) {
	receiver, classVal, err := vm.currentReceiverAndClass(frame)
	if err != nil {
		return core.Value{}, err
	}

	classMap := classVal.Raw.(map[string]core.Value)
	parent, ok := classMap[classParentKey]
	if !ok || parent.Kind == core.ValueNull {
		return core.Value{}, fmt.Errorf("super property %s not found", name)
	}

	method, ok := lookupClassAccessor(parent, classMethodsKey, name)
	if !ok {
		return core.Value{}, fmt.Errorf("super property %s not found", name)
	}
	return bindReceiver(method, receiver), nil
}

func (vm *VM) invokeSuperConstructor(frame *Frame, args []core.Value) (core.Value, error) {
	receiver, classVal, err := vm.currentReceiverAndClass(frame)
	if err != nil {
		return core.Value{}, err
	}

	classMap := classVal.Raw.(map[string]core.Value)
	parent, ok := classMap[classParentKey]
	if !ok || parent.Kind == core.ValueNull {
		return core.Value{}, fmt.Errorf("parent class has no constructor")
	}

	constructor, ok := lookupClassAccessor(parent, classMethodsKey, "constructor")
	if !ok {
		return core.Value{}, fmt.Errorf("parent class has no constructor")
	}
	if err := vm.invokeCallable(bindReceiver(constructor, receiver), args); err != nil {
		return core.Value{}, err
	}
	return core.Value{Kind: core.ValueNull}, nil
}

func (vm *VM) currentReceiverAndClass(frame *Frame) (core.Value, core.Value, error) {
	fn := &vm.mod.Functions[frame.FunctionIndex]
	if !fn.HasThis || int(fn.ThisLocal) >= len(frame.Locals) {
		return core.Value{}, core.Value{}, fmt.Errorf("super can only be used on instance methods")
	}
	receiver := frame.Locals[fn.ThisLocal]
	if !isInstanceValue(receiver) {
		return core.Value{}, core.Value{}, fmt.Errorf("super can only be used on instance methods")
	}
	instanceMap := receiver.Raw.(map[string]core.Value)
	classVal, ok := instanceMap[instanceClassRefKey]
	if !ok || !isClassValue(classVal) {
		return core.Value{}, core.Value{}, fmt.Errorf("super can only be used on instance methods")
	}
	return receiver, classVal, nil
}

func (vm *VM) popHostCallArgs(argCount int) map[string]any {
	if argCount <= 0 {
		return map[string]any{}
	}

	args := make([]core.Value, argCount)
	for i := argCount - 1; i >= 0; i-- {
		args[i] = vm.stack.Pop()
	}

	if argCount == 1 && args[0].Kind == core.ValueObjectRef {
		if mapped, ok := hostObjectArgs(args[0]); ok {
			return mapped
		}
	}

	result := make(map[string]any, argCount+1)
	ordered := make([]any, 0, argCount)
	for i, arg := range args {
		converted := coreValueToHostAny(arg)
		ordered = append(ordered, converted)
		result[fmt.Sprintf("arg%d", i)] = converted
	}
	result["args"] = ordered
	return result
}

func hostObjectArgs(v core.Value) (map[string]any, bool) {
	if v.Kind != core.ValueObjectRef {
		return nil, false
	}
	switch obj := v.Raw.(type) {
	case map[string]core.Value:
		result := make(map[string]any, len(obj))
		for key, val := range obj {
			if isReservedObjectKey(key) {
				continue
			}
			result[key] = coreValueToHostAny(val)
		}
		return result, true
	case map[string]any:
		result := make(map[string]any, len(obj))
		for key, val := range obj {
			result[key] = val
		}
		return result, true
	default:
		return nil, false
	}
}

func coreValueToHostAny(v core.Value) any {
	switch v.Kind {
	case core.ValueNull:
		return nil
	case core.ValueBool, core.ValueI64, core.ValueF64, core.ValueString, core.ValueBytes, core.ValueHostHandle:
		return v.Raw
	case core.ValueArrayRef:
		arr, ok := v.Raw.([]core.Value)
		if !ok {
			return nil
		}
		result := make([]any, 0, len(arr))
		for _, item := range arr {
			result = append(result, coreValueToHostAny(item))
		}
		return result
	case core.ValueObjectRef:
		if mapped, ok := hostObjectArgs(v); ok {
			return mapped
		}
		return nil
	case core.ValueFuncRef:
		return map[string]any{"kind": "func_ref", "index": v.Raw}
	case core.ValuePromise:
		return nil
	default:
		return v.Raw
	}
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
	case []byte:
		return core.Value{Kind: core.ValueBytes, Raw: val}
	case uint64:
		return core.Value{Kind: core.ValueHostHandle, Raw: val}
	case map[string]any:
		return core.Value{Kind: core.ValueObjectRef, Raw: val}
	case []any:
		result := make([]core.Value, 0, len(val))
		for _, item := range val {
			result = append(result, coreValueFromAny(item))
		}
		return core.Value{Kind: core.ValueArrayRef, Raw: result}
	default:
		return core.Value{Kind: core.ValueNull}
	}
}

func hostHandleID(v core.Value) (uint64, error) {
	switch v.Kind {
	case core.ValueHostHandle:
		switch raw := v.Raw.(type) {
		case uint64:
			return raw, nil
		case int64:
			return uint64(raw), nil
		}
	case core.ValueI64:
		return uint64(v.Raw.(int64)), nil
	case core.ValueF64:
		return uint64(v.Raw.(float64)), nil
	}
	return 0, fmt.Errorf("host.poll requires a handle id")
}

func coreValueFromHostPoll(result api.PollResult) core.Value {
	value := map[string]any{
		"done":  result.Done,
		"error": result.Error,
	}
	for key, item := range result.Value {
		value[key] = item
	}
	return core.Value{Kind: core.ValueObjectRef, Raw: value}
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
	case core.ValuePromise:
		return "promise"
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
	case core.ValuePromise:
		state, ok := a.Raw.(*promiseState)
		if ok && state != nil {
			switch state.Status {
			case promiseStatusResolved:
				return "[Promise resolved]"
			case promiseStatusRejected:
				return "[Promise rejected]"
			default:
				return "[Promise pending]"
			}
		}
		return "[Promise pending]"
	default:
		return fmt.Sprintf("%v", a.Kind)
	}
}
