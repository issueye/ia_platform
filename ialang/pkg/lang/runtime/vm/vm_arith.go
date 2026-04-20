package vm

import "fmt"

func (v *VM) execAdd() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}

	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if lok && rok {
		v.push(ln + rn)
		return nil
	}

	v.push(toString(left) + toString(right))
	return nil
}

func (v *VM) execSub() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator - expects numbers, got %T and %T", left, right)
	}
	v.push(ln - rn)
	return nil
}

func (v *VM) execMul() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator * expects numbers, got %T and %T", left, right)
	}
	v.push(ln * rn)
	return nil
}

func (v *VM) execDiv() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator / expects numbers, got %T and %T", left, right)
	}
	if rn == 0 {
		return fmt.Errorf("division by zero")
	}
	v.push(ln / rn)
	return nil
}

func (v *VM) execMod() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator %% expects numbers, got %T and %T", left, right)
	}
	if rn == 0 {
		return fmt.Errorf("modulo by zero")
	}
	leftInt := int(ln)
	rightInt := int(rn)
	if rightInt == 0 {
		return fmt.Errorf("modulo by zero")
	}
	v.push(float64(leftInt % rightInt))
	return nil
}

func (v *VM) execNeg() error {
	val, err := v.pop()
	if err != nil {
		return err
	}
	n, ok := val.(float64)
	if !ok {
		return fmt.Errorf("unary - expects number, got %T", val)
	}
	v.push(-n)
	return nil
}

func (v *VM) execNot() error {
	val, err := v.pop()
	if err != nil {
		return err
	}
	v.push(!isTruthy(val))
	return nil
}

func (v *VM) execAnd() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	// JS-style: returns left if falsy, otherwise returns right
	if !isTruthy(left) {
		v.push(left)
	} else {
		v.push(right)
	}
	return nil
}

func (v *VM) execOr() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	// JS-style: returns left if truthy, otherwise returns right
	if isTruthy(left) {
		v.push(left)
	} else {
		v.push(right)
	}
	return nil
}

func (v *VM) execBitAnd() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator & expects numbers, got %T and %T", left, right)
	}
	v.push(float64(int(ln) & int(rn)))
	return nil
}

func (v *VM) execBitOr() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator | expects numbers, got %T and %T", left, right)
	}
	v.push(float64(int(ln) | int(rn)))
	return nil
}

func (v *VM) execBitXor() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator ^ expects numbers, got %T and %T", left, right)
	}
	v.push(float64(int(ln) ^ int(rn)))
	return nil
}

func (v *VM) execShl() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator << expects numbers, got %T and %T", left, right)
	}
	v.push(float64(int(ln) << uint(int(rn))))
	return nil
}

func (v *VM) execShr() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator >> expects numbers, got %T and %T", left, right)
	}
	v.push(float64(int(ln) >> uint(int(rn))))
	return nil
}

func (v *VM) execTruthy() error {
	val, err := v.pop()
	if err != nil {
		return err
	}
	v.push(isTruthy(val))
	return nil
}

func (v *VM) execEqual() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	v.push(valueEqual(left, right))
	return nil
}

func (v *VM) execNotEqual() error {
	if err := v.execEqual(); err != nil {
		return err
	}
	val, err := v.pop()
	if err != nil {
		return err
	}
	eq, ok := val.(bool)
	if !ok {
		return fmt.Errorf("internal error: equality result is not bool: %T", val)
	}
	v.push(!eq)
	return nil
}

func (v *VM) execGreater() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator > expects numbers, got %T and %T", left, right)
	}
	v.push(ln > rn)
	return nil
}

func (v *VM) execLess() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator < expects numbers, got %T and %T", left, right)
	}
	v.push(ln < rn)
	return nil
}

func (v *VM) execGreaterEqual() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator >= expects numbers, got %T and %T", left, right)
	}
	v.push(ln >= rn)
	return nil
}

func (v *VM) execLessEqual() error {
	right, err := v.pop()
	if err != nil {
		return err
	}
	left, err := v.pop()
	if err != nil {
		return err
	}
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		return fmt.Errorf("operator <= expects numbers, got %T and %T", left, right)
	}
	v.push(ln <= rn)
	return nil
}

