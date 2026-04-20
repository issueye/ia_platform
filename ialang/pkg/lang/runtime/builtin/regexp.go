package builtin

import (
	"fmt"
	goRegexp "regexp"
	"strconv"
	"strings"
	"sync"
)

const regexpCompileCacheLimit = 256

type compiledRegexp struct {
	Pattern  string
	Flags    string
	Raw      *goRegexp.Regexp
	lastUsed int64
}

var (
	regexpCompileCacheMu sync.RWMutex
	regexpCompileCache   = map[string]*compiledRegexp{}
	regexpCacheClock     int64
)

func newRegexpModule() Object {
	compileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("regexp.compile expects 1-2 args: pattern, [flags]")
		}
		pattern, err := asStringArg("regexp.compile", args, 0)
		if err != nil {
			return nil, err
		}
		flags := ""
		if len(args) == 2 {
			flags, err = asStringValue("regexp.compile arg[1]", args[1])
			if err != nil {
				return nil, err
			}
		}
		compiled, err := compileRegexp(pattern, flags)
		if err != nil {
			return nil, err
		}
		return newCompiledRegexpObject(compiled), nil
	})

	testFn := NativeFunction(func(args []Value) (Value, error) {
		pattern, text, flags, err := parseRegexpPatternTextFlagsArgs("regexp.test", args, 2, 3)
		if err != nil {
			return nil, err
		}
		compiled, err := compileRegexp(pattern, flags)
		if err != nil {
			return nil, err
		}
		return compiled.Raw.MatchString(text), nil
	})

	findFn := NativeFunction(func(args []Value) (Value, error) {
		pattern, text, flags, err := parseRegexpPatternTextFlagsArgs("regexp.find", args, 2, 3)
		if err != nil {
			return nil, err
		}
		compiled, err := compileRegexp(pattern, flags)
		if err != nil {
			return nil, err
		}
		out := compiled.Raw.FindString(text)
		if out == "" {
			return nil, nil
		}
		return out, nil
	})

	findAllFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 4 {
			return nil, fmt.Errorf("regexp.findAll expects 2-4 args: pattern, text, [n], [flags]")
		}
		pattern, err := asStringArg("regexp.findAll", args, 0)
		if err != nil {
			return nil, err
		}
		text, err := asStringArg("regexp.findAll", args, 1)
		if err != nil {
			return nil, err
		}
		n, flags, err := parseRegexpOptionalNAndFlags("regexp.findAll", args[2:])
		if err != nil {
			return nil, err
		}
		compiled, err := compileRegexp(pattern, flags)
		if err != nil {
			return nil, err
		}
		return stringSliceToRuntimeArray(compiled.Raw.FindAllString(text, n)), nil
	})

	replaceAllFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 3 || len(args) > 4 {
			return nil, fmt.Errorf("regexp.replaceAll expects 3-4 args: pattern, text, replacement, [flags]")
		}
		pattern, err := asStringArg("regexp.replaceAll", args, 0)
		if err != nil {
			return nil, err
		}
		text, err := asStringArg("regexp.replaceAll", args, 1)
		if err != nil {
			return nil, err
		}
		replacement, err := asStringArg("regexp.replaceAll", args, 2)
		if err != nil {
			return nil, err
		}
		flags := ""
		if len(args) == 4 {
			flags, err = asStringValue("regexp.replaceAll arg[3]", args[3])
			if err != nil {
				return nil, err
			}
		}
		compiled, err := compileRegexp(pattern, flags)
		if err != nil {
			return nil, err
		}
		return compiled.Raw.ReplaceAllString(text, replacement), nil
	})

	splitFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 4 {
			return nil, fmt.Errorf("regexp.split expects 2-4 args: pattern, text, [n], [flags]")
		}
		pattern, err := asStringArg("regexp.split", args, 0)
		if err != nil {
			return nil, err
		}
		text, err := asStringArg("regexp.split", args, 1)
		if err != nil {
			return nil, err
		}
		n, flags, err := parseRegexpOptionalNAndFlags("regexp.split", args[2:])
		if err != nil {
			return nil, err
		}
		compiled, err := compileRegexp(pattern, flags)
		if err != nil {
			return nil, err
		}
		return stringSliceToRuntimeArray(compiled.Raw.Split(text, n)), nil
	})

	findSubmatchFn := NativeFunction(func(args []Value) (Value, error) {
		pattern, text, flags, err := parseRegexpPatternTextFlagsArgs("regexp.findSubmatch", args, 2, 3)
		if err != nil {
			return nil, err
		}
		compiled, err := compileRegexp(pattern, flags)
		if err != nil {
			return nil, err
		}
		submatches := compiled.Raw.FindStringSubmatch(text)
		if submatches == nil {
			return nil, nil
		}
		return stringSliceToRuntimeArray(submatches), nil
	})

	findAllSubmatchFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 4 {
			return nil, fmt.Errorf("regexp.findAllSubmatch expects 2-4 args: pattern, text, [n], [flags]")
		}
		pattern, err := asStringArg("regexp.findAllSubmatch", args, 0)
		if err != nil {
			return nil, err
		}
		text, err := asStringArg("regexp.findAllSubmatch", args, 1)
		if err != nil {
			return nil, err
		}
		n, flags, err := parseRegexpOptionalNAndFlags("regexp.findAllSubmatch", args[2:])
		if err != nil {
			return nil, err
		}
		compiled, err := compileRegexp(pattern, flags)
		if err != nil {
			return nil, err
		}
		return stringSlicesToRuntimeArray(compiled.Raw.FindAllStringSubmatch(text, n)), nil
	})

	quoteMetaFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("regexp.quoteMeta expects 1 arg: text")
		}
		text, err := asStringArg("regexp.quoteMeta", args, 0)
		if err != nil {
			return nil, err
		}
		return goRegexp.QuoteMeta(text), nil
	})

	namespace := Object{
		"compile":         compileFn,
		"test":            testFn,
		"find":            findFn,
		"findAll":         findAllFn,
		"replaceAll":      replaceAllFn,
		"split":           splitFn,
		"findSubmatch":    findSubmatchFn,
		"findAllSubmatch": findAllSubmatchFn,
		"quoteMeta":       quoteMetaFn,
	}
	module := cloneObject(namespace)
	module["regexp"] = namespace
	return module
}

func parseRegexpPatternTextFlagsArgs(fn string, args []Value, minLen int, maxLen int) (string, string, string, error) {
	if len(args) < minLen || len(args) > maxLen {
		if maxLen == minLen {
			return "", "", "", fmt.Errorf("%s expects %d args", fn, minLen)
		}
		return "", "", "", fmt.Errorf("%s expects %d-%d args", fn, minLen, maxLen)
	}
	pattern, err := asStringArg(fn, args, 0)
	if err != nil {
		return "", "", "", err
	}
	text, err := asStringArg(fn, args, 1)
	if err != nil {
		return "", "", "", err
	}
	flags := ""
	if len(args) >= 3 {
		flags, err = asStringValue(fn+" arg[2]", args[2])
		if err != nil {
			return "", "", "", err
		}
	}
	return pattern, text, flags, nil
}

func parseRegexpOptionalNAndFlags(fn string, args []Value) (int, string, error) {
	if len(args) == 0 {
		return -1, "", nil
	}
	if len(args) > 2 {
		return 0, "", fmt.Errorf("%s expects optional [n], [flags]", fn)
	}

	n := -1
	flags := ""
	if len(args) == 1 {
		switch v := args[0].(type) {
		case string:
			if parsedN, err := strconv.Atoi(v); err == nil {
				n = parsedN
			} else {
				flags = v
			}
		default:
			parsedN, err := asIntValue(fn+" arg[2]", args[0])
			if err != nil {
				return 0, "", fmt.Errorf("%s optional arg expects number/string flags, got %T", fn, v)
			}
			n = parsedN
		}
		return n, flags, nil
	}

	parsedN, err := asIntValue(fn+" arg[2]", args[0])
	if err != nil {
		return 0, "", err
	}
	flags, err = asStringValue(fn+" arg[3]", args[1])
	if err != nil {
		return 0, "", err
	}
	return parsedN, flags, nil
}

func compileRegexp(pattern string, flags string) (*compiledRegexp, error) {
	normalizedFlags, err := normalizeRegexpFlags(flags)
	if err != nil {
		return nil, err
	}
	cacheKey := pattern + "\x00" + normalizedFlags

	regexpCompileCacheMu.RLock()
	cached := regexpCompileCache[cacheKey]
	regexpCompileCacheMu.RUnlock()
	if cached != nil {
		regexpCompileCacheMu.Lock()
		regexpCacheClock++
		cached.lastUsed = regexpCacheClock
		regexpCompileCacheMu.Unlock()
		return cached, nil
	}

	compiledPattern := pattern
	if normalizedFlags != "" {
		compiledPattern = "(?" + normalizedFlags + ")" + pattern
	}
	raw, err := goRegexp.Compile(compiledPattern)
	if err != nil {
		return nil, fmt.Errorf("regexp compile error: %w", err)
	}

	regexpCompileCacheMu.Lock()
	defer regexpCompileCacheMu.Unlock()

	if existing := regexpCompileCache[cacheKey]; existing != nil {
		return existing, nil
	}
	if len(regexpCompileCache) >= regexpCompileCacheLimit {
		var oldest string
		var oldestTime int64 = 1<<63 - 1
		for k, v := range regexpCompileCache {
			if v.lastUsed < oldestTime {
				oldestTime = v.lastUsed
				oldest = k
			}
		}
		delete(regexpCompileCache, oldest)
	}

	regexpCacheClock++
	compiled := &compiledRegexp{
		Pattern:  pattern,
		Flags:    normalizedFlags,
		Raw:      raw,
		lastUsed: regexpCacheClock,
	}
	regexpCompileCache[cacheKey] = compiled
	return compiled, nil
}

func normalizeRegexpFlags(flags string) (string, error) {
	normalizedInput := strings.TrimSpace(flags)
	if normalizedInput == "" {
		return "", nil
	}
	seen := map[rune]bool{}
	for _, ch := range normalizedInput {
		switch ch {
		case 'i', 'm', 's':
			seen[ch] = true
		default:
			return "", fmt.Errorf("regexp flags contains unsupported flag %q (allowed: i, m, s)", string(ch))
		}
	}

	out := make([]rune, 0, len(seen))
	for _, flag := range []rune{'i', 'm', 's'} {
		if seen[flag] {
			out = append(out, flag)
		}
	}
	return string(out), nil
}

func newCompiledRegexpObject(compiled *compiledRegexp) Object {
	return Object{
		"pattern": compiled.Pattern,
		"flags":   compiled.Flags,
		"test": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("regexp compiled.test expects 1 arg: text")
			}
			text, err := asStringArg("regexp compiled.test", args, 0)
			if err != nil {
				return nil, err
			}
			return compiled.Raw.MatchString(text), nil
		}),
		"find": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("regexp compiled.find expects 1 arg: text")
			}
			text, err := asStringArg("regexp compiled.find", args, 0)
			if err != nil {
				return nil, err
			}
			out := compiled.Raw.FindString(text)
			if out == "" {
				return nil, nil
			}
			return out, nil
		}),
		"findAll": NativeFunction(func(args []Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("regexp compiled.findAll expects 1-2 args: text, [n]")
			}
			text, err := asStringArg("regexp compiled.findAll", args, 0)
			if err != nil {
				return nil, err
			}
			n := -1
			if len(args) == 2 {
				n, err = asIntArg("regexp compiled.findAll", args, 1)
				if err != nil {
					return nil, err
				}
			}
			return stringSliceToRuntimeArray(compiled.Raw.FindAllString(text, n)), nil
		}),
		"split": NativeFunction(func(args []Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("regexp compiled.split expects 1-2 args: text, [n]")
			}
			text, err := asStringArg("regexp compiled.split", args, 0)
			if err != nil {
				return nil, err
			}
			n := -1
			if len(args) == 2 {
				n, err = asIntArg("regexp compiled.split", args, 1)
				if err != nil {
					return nil, err
				}
			}
			return stringSliceToRuntimeArray(compiled.Raw.Split(text, n)), nil
		}),
		"replaceAll": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("regexp compiled.replaceAll expects 2 args: text, replacement")
			}
			text, err := asStringArg("regexp compiled.replaceAll", args, 0)
			if err != nil {
				return nil, err
			}
			replacement, err := asStringArg("regexp compiled.replaceAll", args, 1)
			if err != nil {
				return nil, err
			}
			return compiled.Raw.ReplaceAllString(text, replacement), nil
		}),
		"findSubmatch": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("regexp compiled.findSubmatch expects 1 arg: text")
			}
			text, err := asStringArg("regexp compiled.findSubmatch", args, 0)
			if err != nil {
				return nil, err
			}
			submatches := compiled.Raw.FindStringSubmatch(text)
			if submatches == nil {
				return nil, nil
			}
			return stringSliceToRuntimeArray(submatches), nil
		}),
		"findAllSubmatch": NativeFunction(func(args []Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("regexp compiled.findAllSubmatch expects 1-2 args: text, [n]")
			}
			text, err := asStringArg("regexp compiled.findAllSubmatch", args, 0)
			if err != nil {
				return nil, err
			}
			n := -1
			if len(args) == 2 {
				n, err = asIntArg("regexp compiled.findAllSubmatch", args, 1)
				if err != nil {
					return nil, err
				}
			}
			return stringSlicesToRuntimeArray(compiled.Raw.FindAllStringSubmatch(text, n)), nil
		}),
	}
}

func stringSliceToRuntimeArray(items []string) Array {
	out := make(Array, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func stringSlicesToRuntimeArray(items [][]string) Array {
	out := make(Array, 0, len(items))
	for _, group := range items {
		out = append(out, stringSliceToRuntimeArray(group))
	}
	return out
}
