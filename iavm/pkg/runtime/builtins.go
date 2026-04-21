package runtime

import (
	"fmt"
	"strconv"

	"iavm/pkg/core"
)

// builtinPrint prints all arguments separated by spaces, followed by a newline.
// Returns null.
func builtinPrint(args []core.Value) core.Value {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(valueToString(arg))
	}
	fmt.Println()
	return core.Value{Kind: core.ValueNull}
}

// builtinLen returns the length of a string or array as an i64.
// Returns null for unsupported types.
func builtinLen(args []core.Value) core.Value {
	if len(args) != 1 {
		return core.Value{Kind: core.ValueNull}
	}
	switch args[0].Kind {
	case core.ValueString:
		return core.Value{Kind: core.ValueI64, Raw: int64(len(args[0].Raw.(string)))}
	case core.ValueArrayRef:
		if arr, ok := args[0].Raw.([]core.Value); ok {
			return core.Value{Kind: core.ValueI64, Raw: int64(len(arr))}
		}
	}
	return core.Value{Kind: core.ValueNull}
}

// builtinTypeof returns the type name of a value as a string.
func builtinTypeof(args []core.Value) core.Value {
	if len(args) != 1 {
		return core.Value{Kind: core.ValueNull}
	}
	return core.Value{Kind: core.ValueString, Raw: typeOfValue(args[0])}
}

// builtinStr converts a value to its string representation.
func builtinStr(args []core.Value) core.Value {
	if len(args) != 1 {
		return core.Value{Kind: core.ValueNull}
	}
	return core.Value{Kind: core.ValueString, Raw: valueToString(args[0])}
}

// builtinInt converts a value to an integer.
// Supports i64 (identity), f64 (truncation), and string (parsing).
func builtinInt(args []core.Value) core.Value {
	if len(args) != 1 {
		return core.Value{Kind: core.ValueNull}
	}
	switch args[0].Kind {
	case core.ValueI64:
		return args[0]
	case core.ValueF64:
		return core.Value{Kind: core.ValueI64, Raw: int64(args[0].Raw.(float64))}
	case core.ValueString:
		if v, err := strconv.ParseInt(args[0].Raw.(string), 10, 64); err == nil {
			return core.Value{Kind: core.ValueI64, Raw: v}
		}
	}
	return core.Value{Kind: core.ValueNull}
}

// builtinFloat converts a value to a float.
// Supports f64 (identity), i64 (conversion), and string (parsing).
func builtinFloat(args []core.Value) core.Value {
	if len(args) != 1 {
		return core.Value{Kind: core.ValueNull}
	}
	switch args[0].Kind {
	case core.ValueF64:
		return args[0]
	case core.ValueI64:
		return core.Value{Kind: core.ValueF64, Raw: float64(args[0].Raw.(int64))}
	case core.ValueString:
		if v, err := strconv.ParseFloat(args[0].Raw.(string), 64); err == nil {
			return core.Value{Kind: core.ValueF64, Raw: v}
		}
	}
	return core.Value{Kind: core.ValueNull}
}
