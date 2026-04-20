package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type threshold struct {
	maxNsOp     float64
	maxAllocsOp float64
}

var limits = map[string]threshold{
	"BenchmarkVMFunctionCall-":      {maxNsOp: 700000, maxAllocsOp: 7000},
	"BenchmarkVMRecursiveFunction-": {maxNsOp: 1000000, maxAllocsOp: 11000},
	"BenchmarkVMClosureExecution-":  {maxNsOp: 700000, maxAllocsOp: 7000},
	"BenchmarkVMClassMethodCall-":   {maxNsOp: 900000, maxAllocsOp: 8000},
	"BenchmarkVMArrayPush-":         {maxNsOp: 900000, maxAllocsOp: 7000},
	"BenchmarkCompilerSmallAST-":    {maxNsOp: 7000, maxAllocsOp: 90},
	"BenchmarkCompilerMediumAST-":   {maxNsOp: 80000, maxAllocsOp: 1300},
}

var lineRe = regexp.MustCompile(`^(Benchmark\S+)\s+\d+\s+([0-9.]+)\s+ns/op(?:\s+([0-9.]+)\s+B/op\s+([0-9.]+)\s+allocs/op)?`)

func normalizeName(name string) string {
	if idx := strings.LastIndex(name, "-"); idx >= 0 {
		return name[:idx+1]
	}
	return name
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./scripts/bench_guard.go <benchmark_output_file>")
		os.Exit(2)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open benchmark output: %v\n", err)
		os.Exit(2)
	}
	defer file.Close()

	seen := map[string]bool{}
	var failures []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		m := lineRe.FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}

		rawName := m[1]
		name := normalizeName(rawName)
		limit, ok := limits[name]
		if !ok {
			continue
		}
		seen[name] = true

		nsOp, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s parse ns/op error: %v", rawName, err))
			continue
		}
		if nsOp > limit.maxNsOp {
			failures = append(failures, fmt.Sprintf("%s ns/op %.0f > %.0f", rawName, nsOp, limit.maxNsOp))
		}

		if len(m) >= 5 && m[4] != "" {
			allocsOp, err := strconv.ParseFloat(m[4], 64)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s parse allocs/op error: %v", rawName, err))
				continue
			}
			if allocsOp > limit.maxAllocsOp {
				failures = append(failures, fmt.Sprintf("%s allocs/op %.0f > %.0f", rawName, allocsOp, limit.maxAllocsOp))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read benchmark output: %v\n", err)
		os.Exit(2)
	}

	for name := range limits {
		if !seen[name] {
			failures = append(failures, fmt.Sprintf("missing benchmark result: %s*", name))
		}
	}

	if len(failures) > 0 {
		fmt.Fprintln(os.Stderr, "performance guard failed:")
		for _, f := range failures {
			fmt.Fprintf(os.Stderr, "- %s\n", f)
		}
		os.Exit(1)
	}

	fmt.Println("performance guard passed")
}
