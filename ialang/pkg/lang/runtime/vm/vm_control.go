package vm

import "errors"

func (v *VM) execPushTry(catchIP int, catchNameIdx int) error {
	catchName, err := v.stringConstant(catchNameIdx, "catch name")
	if err != nil {
		return err
	}
	v.tryStack = append(v.tryStack, tryFrame{
		catchIP:   catchIP,
		catchName: catchName,
		stackBase: len(v.stack),
	})
	return nil
}

func (v *VM) execPopTry() error {
	if len(v.tryStack) == 0 {
		return errors.New("try stack underflow")
	}
	v.tryStack = v.tryStack[:len(v.tryStack)-1]
	return nil
}

func (v *VM) execThrow() error {
	val, err := v.pop()
	if err != nil {
		return err
	}
	return &thrownError{value: val}
}

func (v *VM) execJumpIfFalse(target int) error {
	cond, err := v.pop()
	if err != nil {
		return err
	}
	if !isTruthy(cond) {
		v.ip = target
	}
	return nil
}

func (v *VM) execJumpIfTrue(target int) error {
	cond, err := v.pop()
	if err != nil {
		return err
	}
	if isTruthy(cond) {
		v.ip = target
	}
	return nil
}

func (v *VM) execJumpIfNullish(target int) error {
	val, err := v.pop()
	if err != nil {
		return err
	}
	// Jump if value is null or undefined (nil in Go)
	if val == nil {
		v.ip = target
	} else {
		// Not null/undefined, push it back and don't jump
		v.push(val)
	}
	return nil
}

func (v *VM) execJumpIfNotNullish(target int) error {
	val, err := v.pop()
	if err != nil {
		return err
	}
	// Always push the value back - we're just checking, not consuming
	v.push(val)
	
	// Jump if value is NOT null or undefined
	if val != nil {
		v.ip = target
	}
	// If it is null/undefined, don't jump - let the next OpPop remove it
	return nil
}

