package builtin

import (
	"encoding/csv"
	"fmt"
	"strings"
)

func newCSVModule() Object {
	parseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("csv.parse expects 1-2 args: text, [options]")
		}
		text, err := asStringArg("csv.parse", args, 0)
		if err != nil {
			return nil, err
		}
		delimiter, err := parseCSVDelimiterOption(args, 1)
		if err != nil {
			return nil, err
		}

		reader := csv.NewReader(strings.NewReader(text))
		reader.Comma = delimiter
		rows, err := reader.ReadAll()
		if err != nil {
			return nil, err
		}
		out := make(Array, 0, len(rows))
		for _, row := range rows {
			line := make(Array, 0, len(row))
			for _, col := range row {
				line = append(line, col)
			}
			out = append(out, line)
		}
		return out, nil
	})

	stringifyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("csv.stringify expects 1-2 args: rows, [options]")
		}
		rows, ok := args[0].(Array)
		if !ok {
			return nil, fmt.Errorf("csv.stringify arg[0] expects array, got %T", args[0])
		}
		delimiter, err := parseCSVDelimiterOption(args, 1)
		if err != nil {
			return nil, err
		}

		var b strings.Builder
		writer := csv.NewWriter(&b)
		writer.Comma = delimiter
		for i, rowValue := range rows {
			row, ok := rowValue.(Array)
			if !ok {
				return nil, fmt.Errorf("csv.stringify row[%d] expects array, got %T", i, rowValue)
			}
			record := make([]string, 0, len(row))
			for j, colValue := range row {
				col, err := asStringValue(fmt.Sprintf("csv.stringify row[%d][%d]", i, j), colValue)
				if err != nil {
					return nil, err
				}
				record = append(record, col)
			}
			if err := writer.Write(record); err != nil {
				return nil, err
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return nil, err
		}
		return b.String(), nil
	})

	namespace := Object{
		"parse":     parseFn,
		"stringify": stringifyFn,
	}
	module := cloneObject(namespace)
	module["csv"] = namespace
	return module
}

func parseCSVDelimiterOption(args []Value, idx int) (rune, error) {
	delimiter := ','
	if len(args) <= idx || args[idx] == nil {
		return delimiter, nil
	}
	options, ok := args[idx].(Object)
	if !ok {
		return delimiter, fmt.Errorf("csv options expects object, got %T", args[idx])
	}
	if v, ok := options["delimiter"]; ok && v != nil {
		s, err := asStringValue("csv options.delimiter", v)
		if err != nil {
			return delimiter, err
		}
		if len([]rune(s)) != 1 {
			return delimiter, fmt.Errorf("csv options.delimiter expects single rune string, got %q", s)
		}
		delimiter = []rune(s)[0]
	}
	return delimiter, nil
}
