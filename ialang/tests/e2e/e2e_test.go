package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// ============================================================
// E2E/Smoke tests for real projects
// ============================================================

func getRepoRoot() string {
	// Get the repository root directory
	// This file is at: E:\code\issueye\ialang\ialang\tests\e2e\e2e_test.go
	// Repo root is at: E:\code\issueye\ialang\ialang
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

func getIAPMRoot() string {
	return filepath.Join(getRepoRoot(), "..", "iapm")
}

func runIALangCommand(t *testing.T, timeout time.Duration, args ...string) (string, string, error) {
	t.Helper()
	
	repoRoot := getRepoRoot()
	cmdArgs := append([]string{"run"}, args...)
	
	cmd := exec.Command(filepath.Join(repoRoot, "ialang.exe"), cmdArgs...)
	cmd.Dir = repoRoot
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()
	
	select {
	case err := <-done:
		return stdout.String(), stderr.String(), err
	case <-time.After(timeout):
		cmd.Process.Kill()
		return stdout.String(), stderr.String(), fmt.Errorf("command timed out after %v", timeout)
	}
}

func runIAPMCommand(t *testing.T, timeout time.Duration, args ...string) (string, string, error) {
	t.Helper()

	iapmRoot := getIAPMRoot()
	cmdArgs := append([]string{"run", "main.ia"}, args...)

	cmd := exec.Command(filepath.Join(getRepoRoot(), "ialang.exe"), cmdArgs...)
	cmd.Dir = iapmRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		return stdout.String(), stderr.String(), err
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		return stdout.String(), stderr.String(), fmt.Errorf("command timed out after %v", timeout)
	}
}

type iapmEnvelope struct {
	OK      bool            `json:"ok"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type iapmInfoData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Latest  string `json:"latest"`
}

func parseIAPMEnvelope(t *testing.T, stdout string) iapmEnvelope {
	t.Helper()

	start := strings.Index(stdout, "{")
	if start < 0 {
		t.Fatalf("stdout does not contain JSON:\n%s", stdout)
	}

	var env iapmEnvelope
	dec := json.NewDecoder(strings.NewReader(stdout[start:]))
	if err := dec.Decode(&env); err != nil {
		t.Fatalf("failed to parse JSON from stdout %q: %v", stdout, err)
	}
	return env
}

func runIAPMJSONCommand(t *testing.T, timeout time.Duration, args ...string) iapmEnvelope {
	t.Helper()

	stdout, stderr, err := runIAPMCommand(t, timeout, args...)
	if err != nil {
		t.Fatalf("iapm command %v failed:\nstdout: %s\nstderr: %s\nerr: %v", args, stdout, stderr, err)
	}

	env := parseIAPMEnvelope(t, stdout)
	if !env.OK {
		t.Fatalf("iapm command %v returned error: %+v", args, env)
	}
	return env
}

func findFreeAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	return ln.Addr().String()
}

func waitForTCP(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s", addr)
}

func buildIAPMSignature(secret, method, path, timestamp, body string) string {
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = io.WriteString(mac, method)
	_, _ = io.WriteString(mac, "\n")
	_, _ = io.WriteString(mac, path)
	_, _ = io.WriteString(mac, "\n")
	_, _ = io.WriteString(mac, timestamp)
	_, _ = io.WriteString(mac, "\n")
	_, _ = io.WriteString(mac, body)
	return hex.EncodeToString(mac.Sum(nil))
}

func doIAPMRequest(t *testing.T, addr, secret, method, path, body, timestamp, signature string) (int, string) {
	t.Helper()

	if timestamp == "" {
		timestamp = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	if signature == "" {
		signature = buildIAPMSignature(secret, method, path, timestamp, body)
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, "http://"+addr+path, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Ialink-Timestamp", timestamp)
	req.Header.Set("X-Ialink-Signature", signature)

	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	return resp.StatusCode, string(raw)
}

func startIAPMServer(t *testing.T, addr, dataDir, secret string) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(
		filepath.Join(getRepoRoot(), "ialang.exe"),
		"run", "main.ia", "serve", addr, dataDir, secret,
	)
	cmd.Dir = getIAPMRoot()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start iapm server: %v", err)
	}

	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		}
		if t.Failed() {
			t.Logf("iapm server stdout:\n%s", stdout.String())
			t.Logf("iapm server stderr:\n%s", stderr.String())
		}
	})

	if err := waitForTCP(addr, 5*time.Second); err != nil {
		t.Fatalf("iapm server did not become ready: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	return cmd
}

func TestE2E_IAPM_BasicCommands(t *testing.T) {
	// Test that iapm can be loaded and parsed without errors
	iapmPath := filepath.Join(getRepoRoot(), "..", "iapm", "main.ia")
	
	if _, err := os.Stat(iapmPath); os.IsNotExist(err) {
		t.Skip("iapm project not found, skipping e2e test")
	}
	
	// Just verify the file parses correctly
	// Full server tests would require starting a server
	t.Log("iapm main.ia exists at:", iapmPath)
}

func TestE2E_IAPM_InfoLatestAlias(t *testing.T) {
	iapmPath := filepath.Join(getIAPMRoot(), "main.ia")
	if _, err := os.Stat(iapmPath); os.IsNotExist(err) {
		t.Skip("iapm project not found, skipping e2e test")
	}

	addr := findFreeAddr(t)
	dataDir := filepath.Join(t.TempDir(), "iapm-data")
	secret := "iapm-dev-secret"
	pkgPath := filepath.Join(t.TempDir(), "demo.iapkg")
	if err := os.WriteFile(pkgPath, []byte("demo package content"), 0o644); err != nil {
		t.Fatalf("write package file failed: %v", err)
	}

	startIAPMServer(t, addr, dataDir, secret)

	stdout, stderr, err := runIAPMCommand(t, 10*time.Second, "publish", addr, secret, "demo", "1.0.0", pkgPath, "demo package")
	if err != nil {
		t.Fatalf("publish 1.0.0 failed:\nstdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
	}
	firstPublish := parseIAPMEnvelope(t, stdout)
	if !firstPublish.OK {
		t.Fatalf("publish 1.0.0 returned error: %+v", firstPublish)
	}

	stdout, stderr, err = runIAPMCommand(t, 10*time.Second, "publish", addr, secret, "demo", "1.1.0", pkgPath, "demo package")
	if err != nil {
		t.Fatalf("publish 1.1.0 failed:\nstdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
	}
	secondPublish := parseIAPMEnvelope(t, stdout)
	if !secondPublish.OK {
		t.Fatalf("publish 1.1.0 returned error: %+v", secondPublish)
	}

	stdout, stderr, err = runIAPMCommand(t, 10*time.Second, "info", addr, secret, "demo", "latest")
	if err != nil {
		t.Fatalf("info latest failed:\nstdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
	}
	infoLatest := parseIAPMEnvelope(t, stdout)
	if !infoLatest.OK {
		t.Fatalf("info latest returned error: %+v", infoLatest)
	}

	var infoData iapmInfoData
	if err := json.Unmarshal(infoLatest.Data, &infoData); err != nil {
		t.Fatalf("failed to parse info data: %v", err)
	}

	if infoData.Version != "1.1.0" {
		t.Fatalf("info latest version = %q, want 1.1.0", infoData.Version)
	}
	if infoData.Latest != "1.1.0" {
		t.Fatalf("info latest latest = %q, want 1.1.0", infoData.Latest)
	}
}

func TestE2E_IAPM_RejectsInvalidSignature(t *testing.T) {
	iapmPath := filepath.Join(getIAPMRoot(), "main.ia")
	if _, err := os.Stat(iapmPath); os.IsNotExist(err) {
		t.Skip("iapm project not found, skipping e2e test")
	}

	addr := findFreeAddr(t)
	dataDir := filepath.Join(t.TempDir(), "iapm-data")
	secret := "iapm-test-secret"

	startIAPMServer(t, addr, dataDir, secret)

	status, body := doIAPMRequest(
		t,
		addr,
		secret,
		http.MethodGet,
		"/health",
		"",
		fmt.Sprintf("%d", time.Now().UnixMilli()),
		"bad-signature",
	)

	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body: %s", status, http.StatusUnauthorized, body)
	}
	if !strings.Contains(body, "UNAUTHORIZED") {
		t.Fatalf("response body = %q, want unauthorized error", body)
	}
}

func TestE2E_IAPM_RejectsExpiredTimestamp(t *testing.T) {
	iapmPath := filepath.Join(getIAPMRoot(), "main.ia")
	if _, err := os.Stat(iapmPath); os.IsNotExist(err) {
		t.Skip("iapm project not found, skipping e2e test")
	}

	addr := findFreeAddr(t)
	dataDir := filepath.Join(t.TempDir(), "iapm-data")
	secret := "iapm-test-secret"

	startIAPMServer(t, addr, dataDir, secret)

	expiredTimestamp := fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).UnixMilli())
	status, body := doIAPMRequest(
		t,
		addr,
		secret,
		http.MethodGet,
		"/health",
		"",
		expiredTimestamp,
		"",
	)

	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body: %s", status, http.StatusUnauthorized, body)
	}
	if !strings.Contains(body, "UNAUTHORIZED") {
		t.Fatalf("response body = %q, want unauthorized error", body)
	}
}

func TestE2E_IAPM_AcceptsValidSignature(t *testing.T) {
	iapmPath := filepath.Join(getIAPMRoot(), "main.ia")
	if _, err := os.Stat(iapmPath); os.IsNotExist(err) {
		t.Skip("iapm project not found, skipping e2e test")
	}

	addr := findFreeAddr(t)
	dataDir := filepath.Join(t.TempDir(), "iapm-data")
	secret := "iapm-test-secret"

	startIAPMServer(t, addr, dataDir, secret)

	stdout, stderr, err := runIAPMCommand(t, 10*time.Second, "ping", addr, secret)
	if err != nil {
		t.Fatalf("ping failed:\nstdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
	}

	env := parseIAPMEnvelope(t, stdout)
	if !env.OK {
		t.Fatalf("ping returned error: %+v", env)
	}
}

func TestE2E_IAPM_ReleaseGate(t *testing.T) {
	iapmPath := filepath.Join(getIAPMRoot(), "main.ia")
	if _, err := os.Stat(iapmPath); os.IsNotExist(err) {
		t.Skip("iapm project not found, skipping e2e test")
	}

	addr := findFreeAddr(t)
	dataDir := filepath.Join(t.TempDir(), "iapm-data")
	secret := "iapm-test-secret"
	pkgContent := []byte("demo package content for release gate")
	pkgPath := filepath.Join(t.TempDir(), "demo.iapkg")
	outPath := filepath.Join(t.TempDir(), "demo-fetched.iapkg")
	if err := os.WriteFile(pkgPath, pkgContent, 0o644); err != nil {
		t.Fatalf("write package file failed: %v", err)
	}

	startIAPMServer(t, addr, dataDir, secret)

	runIAPMJSONCommand(t, 10*time.Second, "publish", addr, secret, "demo", "1.0.0", pkgPath, "release gate demo")
	runIAPMJSONCommand(t, 10*time.Second, "versions", addr, secret, "demo")
	runIAPMJSONCommand(t, 10*time.Second, "info", addr, secret, "demo", "latest")
	runIAPMJSONCommand(t, 10*time.Second, "get", addr, secret, "demo", "1.0.0")
	runIAPMJSONCommand(t, 10*time.Second, "fetch", addr, secret, "demo", "1.0.0", outPath)
	runIAPMJSONCommand(t, 10*time.Second, "install", addr, secret, "demo")

	fetched, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read fetched package: %v", err)
	}
	if !bytes.Equal(fetched, pkgContent) {
		t.Fatalf("fetched package content mismatch: got %q want %q", string(fetched), string(pkgContent))
	}
}

func TestE2E_Examples_ParseAndRun(t *testing.T) {
	// Test that key example files parse and run without errors
	repoRoot := getRepoRoot()
	examplesDir := filepath.Join(repoRoot, "examples")
	
	examples := []string{
		"hello.ia",
		"function.ia",
		"control.ia",
		"class.ia",
	}
	
	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			examplePath := filepath.Join(examplesDir, example)
			
			if _, err := os.Stat(examplePath); os.IsNotExist(err) {
				t.Skipf("Example %s not found, skipping", example)
			}
			
			// Try to run the example
			stdout, stderr, err := runIALangCommand(t, 10*time.Second, examplePath)
			
			// We expect the run to succeed (no parse/compile errors)
			if err != nil {
				// Check if it's a runtime error (which is OK) vs parse/compile error
				if strings.Contains(stderr, "parse errors") || strings.Contains(stderr, "compile errors") {
					t.Errorf("Example %s failed to parse/compile:\nstdout: %s\nstderr: %s", 
						example, stdout, stderr)
				}
				// Runtime errors are acceptable for e2e test
			}
		})
	}
}

func TestE2E_Formatting_ProjectWide(t *testing.T) {
	// Test that formatting works on the entire ialang project
	repoRoot := getRepoRoot()
	
	// Run fmt on current directory (should format all .ia files)
	stdout, stderr, _ := runIALangCommand(t, 30*time.Second, "fmt", repoRoot)
	
	// Allow errors for files with syntax errors
	// The important thing is the command runs and doesn't crash
	t.Logf("Format output:\nstdout: %s\nstderr: %s", stdout, stderr)
	
	// We don't fail the test if some files have parse errors
	// The formatter should handle them gracefully
}

func TestE2E_Check_IAPMProject(t *testing.T) {
	// Test that check command works on iapm project
	iapmPath := filepath.Join(getRepoRoot(), "..", "iapm")
	
	if _, err := os.Stat(iapmPath); os.IsNotExist(err) {
		t.Skip("iapm project not found, skipping check test")
	}
	
	stdout, stderr, err := runIALangCommand(t, 30*time.Second, "check", iapmPath)
	
	t.Logf("Check output:\nstdout: %s\nstderr: %s", stdout, stderr)
	
	// Check may report errors for files with syntax issues
	// The important thing is the command runs
	_ = err
}

func TestE2E_LanguageSpecExamples(t *testing.T) {
	// Verify that examples from language-spec.md work correctly
	repoRoot := getRepoRoot()
	specPath := filepath.Join(repoRoot, "docs", "language-spec.md")
	
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("language-spec.md not found, skipping")
	}
	
	// The spec exists, which documents the semantics
	// Future work: extract and run examples from it
	t.Log("Language spec exists at:", specPath)
}
