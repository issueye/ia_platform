package builtin

import (
	"fmt"
	gostrc "strconv"
)

func newStrconvModule() Object {
	atoiFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("strconv.atoi expects 1 arg: text")
		}
		text, err := asStringArg("strconv.atoi", args, 0)
		if err != nil {
			return nil, err
		}
		n, err := gostrc.Atoi(text)
		if err != nil {
			return nil, err
		}
		return float64(n), nil
	})

	itoaFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("strconv.itoa expects 1 arg: number")
		}
		n, err := asIntArg("strconv.itoa", args, 0)
		if err != nil {
			return nil, err
		}
		return gostrc.Itoa(n), nil
	})

	parseFloatFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("strconv.parseFloat expects 1 arg: text")
		}
		text, err := asStringArg("strconv.parseFloat", args, 0)
		if err != nil {
			return nil, err
		}
		n, err := gostrc.ParseFloat(text, 64)
		if err != nil {
			return nil, err
		}
		return n, nil
	})

	formatFloatFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("strconv.formatFloat expects 1-2 args: number, [precision]")
		}
		n, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("strconv.formatFloat arg[0] expects number, got %T", args[0])
		}
		prec := -1
		if len(args) == 2 {
			p, err := asIntArg("strconv.formatFloat", args, 1)
			if err != nil {
				return nil, err
			}
			prec = p
		}
		return gostrc.FormatFloat(n, 'f', prec, 64), nil
	})

	parseBoolFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("strconv.parseBool expects 1 arg: text")
		}
		text, err := asStringArg("strconv.parseBool", args, 0)
		if err != nil {
			return nil, err
		}
		v, err := gostrc.ParseBool(text)
		if err != nil {
			return nil, err
		}
		return v, nil
	})

	formatBoolFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("strconv.formatBool expects 1 arg: bool")
		}
		v, ok := args[0].(bool)
		if !ok {
			return nil, fmt.Errorf("strconv.formatBool arg[0] expects bool, got %T", args[0])
		}
		return gostrc.FormatBool(v), nil
	})

	namespace := Object{
		"atoi":        atoiFn,
		"itoa":        itoaFn,
		"parseFloat":  parseFloatFn,
		"formatFloat": formatFloatFn,
		"parseBool":   parseBoolFn,
		"formatBool":  formatBoolFn,
	}
	module := cloneObject(namespace)
	module["strconv"] = namespace
	return module
}
