package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"iacommon/pkg/ialang/packagefile"
)

var embeddedPackageMagic = []byte("IALANG_EMBEDDED_PKG_V1")

const embeddedPackageLengthBytes = 8

func executeBuildBinCommand(entryPath, outPath string, stderr io.Writer) error {
	if outPath == "" {
		outPath = defaultStandaloneOutput(entryPath)
	}
	pkg, err := buildPackage(entryPath, stderr)
	if err != nil {
		return err
	}
	pkgBytes, err := packagefile.Encode(pkg)
	if err != nil {
		return fmt.Errorf("encode package error: %w", err)
	}
	if err := writeStandaloneBinary(outPath, pkgBytes); err != nil {
		return err
	}
	return nil
}

func defaultStandaloneOutput(entryPath string) string {
	base := strings.TrimSuffix(filepath.Base(entryPath), filepath.Ext(entryPath))
	if base == "" {
		base = "app"
	}
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func writeStandaloneBinary(outPath string, pkgBytes []byte) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve current executable path error: %w", err)
	}

	exeAbs, err := filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("resolve executable abs path error: %w", err)
	}
	outAbs, err := filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("resolve output abs path error: %w", err)
	}
	if strings.EqualFold(exeAbs, outAbs) {
		return fmt.Errorf("output path must not overwrite current executable: %s", outPath)
	}

	exeBytes, err := os.ReadFile(exePath)
	if err != nil {
		return fmt.Errorf("read current executable error: %w", err)
	}
	payload, err := appendEmbeddedPackagePayload(exeBytes, pkgBytes)
	if err != nil {
		return err
	}

	fileMode := os.FileMode(0o755)
	if stat, statErr := os.Stat(exePath); statErr == nil {
		fileMode = stat.Mode()
	}
	if err := os.WriteFile(outPath, payload, fileMode); err != nil {
		return fmt.Errorf("write standalone binary error: %w", err)
	}
	return nil
}

func appendEmbeddedPackagePayload(executableBytes, pkgBytes []byte) ([]byte, error) {
	if len(executableBytes) == 0 {
		return nil, fmt.Errorf("current executable bytes are empty")
	}
	if len(pkgBytes) == 0 {
		return nil, fmt.Errorf("package payload is empty")
	}

	lengthBuf := make([]byte, embeddedPackageLengthBytes)
	binary.LittleEndian.PutUint64(lengthBuf, uint64(len(pkgBytes)))

	out := make([]byte, 0, len(executableBytes)+len(pkgBytes)+len(embeddedPackageMagic)+embeddedPackageLengthBytes)
	out = append(out, executableBytes...)
	out = append(out, pkgBytes...)
	out = append(out, embeddedPackageMagic...)
	out = append(out, lengthBuf...)
	return out, nil
}

func extractEmbeddedPackagePayload(executableBytes []byte) ([]byte, bool, error) {
	footerLen := len(embeddedPackageMagic) + embeddedPackageLengthBytes
	if len(executableBytes) < footerLen {
		return nil, false, nil
	}

	lengthStart := len(executableBytes) - embeddedPackageLengthBytes
	magicStart := lengthStart - len(embeddedPackageMagic)
	if magicStart < 0 {
		return nil, false, nil
	}

	if !bytes.Equal(executableBytes[magicStart:lengthStart], embeddedPackageMagic) {
		return nil, false, nil
	}

	payloadLen := binary.LittleEndian.Uint64(executableBytes[lengthStart:])
	if payloadLen == 0 {
		return nil, true, fmt.Errorf("embedded package length is zero")
	}
	if payloadLen > uint64(magicStart) {
		return nil, true, fmt.Errorf("embedded package length is invalid: %d", payloadLen)
	}

	payloadStart := magicStart - int(payloadLen)
	payload := make([]byte, int(payloadLen))
	copy(payload, executableBytes[payloadStart:magicStart])
	return payload, true, nil
}

func maybeRunEmbeddedPackage(stderr io.Writer) (handled bool, exitCode int) {
	exePath, err := os.Executable()
	if err != nil {
		return false, 0
	}
	exeBytes, err := os.ReadFile(exePath)
	if err != nil {
		return false, 0
	}
	pkgBytes, found, err := extractEmbeddedPackagePayload(exeBytes)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return true, 1
	}
	if !found {
		return false, 0
	}

	pkg, err := packagefile.Decode(pkgBytes)
	if err != nil {
		fmt.Fprintln(stderr, fmt.Errorf("decode embedded package error: %w", err).Error())
		return true, 1
	}
	if err := runDecodedPackage(pkg, exePath, os.Args[1:]); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return true, 1
	}
	return true, 0
}
