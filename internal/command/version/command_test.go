package version

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRelevantEnv(t *testing.T) {
	tests := []struct {
		env, want string
		ok        bool
	}{
		{"NEXUS_IQ_MCP_DOTENV=/etc/nexus-iq-mcp/.env", "NEXUS_IQ_MCP_DOTENV=/etc/nexus-iq-mcp/.env", true},
		{"NEXUS_IQ_MCP_LOG_LEVEL=debug", "NEXUS_IQ_MCP_LOG_LEVEL=debug", true},
		{"NEXUS_IQ_PASSWORD=hunter2", "NEXUS_IQ_PASSWORD=********", true},
		{"NEXUS_IQ_URL=http://localhost:8070", "NEXUS_IQ_URL=http://localhost:8070", true},
		{"NEXUS_IQ_MCP_CLIENT_SECRET=s3cr3t", "NEXUS_IQ_MCP_CLIENT_SECRET=********", true},
		{"OS_AUTH_URL=https://keystone:5000/v3", "", false},
		{"HOME=/root", "", false},
		{"MIDPOINT_PASSWORD=x", "", false},
	}
	for _, tt := range tests {
		got, ok := relevantEnv(tt.env)
		if got != tt.want || ok != tt.ok {
			t.Errorf("relevantEnv(%q) = %q, %v; want %q, %v", tt.env, got, ok, tt.want, tt.ok)
		}
	}
}

// TestExecute runs the command in both modes and checks that the verbose
// listing masks secrets.
func TestExecute(t *testing.T) {
	dotenv := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(dotenv, []byte("NEXUS_IQ_MCP_TEST_SETTING=on\nNEXUS_IQ_MCP_TEST_PASSWORD=hunter2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("NEXUS_IQ_MCP_TEST_SETTING"); _ = os.Unsetenv("NEXUS_IQ_MCP_TEST_PASSWORD") })
	missing := filepath.Join(t.TempDir(), "missing")
	for _, cmd := range []Version{{}, {Verbose: true}, {Verbose: true, DotEnv: &dotenv}, {Verbose: true, DotEnv: &missing}} {
		out := capture(t, func() error { return cmd.Execute(nil) })
		if cmd.DotEnv == &dotenv && (!strings.Contains(out, "NEXUS_IQ_MCP_TEST_SETTING=on") || !strings.Contains(out, "NEXUS_IQ_MCP_TEST_PASSWORD=********") || strings.Contains(out, "hunter2")) {
			t.Errorf("verbose output:\n%s", out)
		}
	}
}

// capture returns what f writes to standard output.
func capture(t *testing.T, f func() error) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	err = f()
	os.Stdout = stdout
	_ = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	out, _ := io.ReadAll(r)
	return string(out)
}
