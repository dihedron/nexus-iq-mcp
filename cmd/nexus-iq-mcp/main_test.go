package main

import (
	"bufio"
	"bytes"
	"context"
	json "encoding/json/v2"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestMain lets the test binary act as the real program when it re-runs itself.
func TestMain(m *testing.M) {
	if os.Getenv("NEXUS_IQ_MCP_TEST_RUN_MAIN") == "1" {
		os.Args = []string{"nexus-iq-mcp", "serve"}
		os.Exit(run())
	}
	os.Exit(m.Run())
}

// TestServeOverStdio runs the program as an MCP client does: the handshake and
// the tool list over stdio, only JSON on stdout, and a clean stop on SIGTERM.
func TestServeOverStdio(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0])
	cmd.Env = append(os.Environ(),
		"NEXUS_IQ_MCP_TEST_RUN_MAIN=1",
		"NEXUS_IQ_URL=http://localhost:1",
		"NEXUS_IQ_USERNAME=admin",
		"NEXUS_IQ_PASSWORD=secret",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	responses := bufio.NewScanner(stdout)
	responses.Buffer(nil, 1024*1024)

	call := func(request string) string {
		t.Helper()
		if _, err := io.WriteString(stdin, request+"\n"); err != nil {
			t.Fatal(err)
		}
		if !responses.Scan() {
			t.Fatalf("no response to %s: %v; stderr: %s", request, responses.Err(), stderr.String())
		}
		var message map[string]any
		if err := json.Unmarshal(responses.Bytes(), &message); err != nil {
			t.Fatalf("stdout is not JSON: %q: %v", responses.Text(), err)
		}
		return responses.Text()
	}

	call(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`)
	tools := call(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	for _, want := range []string{"get_component_details", "get_component_versions", "get_component_remediation"} {
		if !strings.Contains(tools, fmt.Sprintf("%q", want)) {
			t.Errorf("tool %s is missing: %s", want, tools)
		}
	}

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Errorf("SIGTERM is a normal stop, got: %v; stderr: %s", err, stderr.String())
	}
}
