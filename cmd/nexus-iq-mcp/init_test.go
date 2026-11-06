package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDotEnvLogging runs a real binary: init() derives its variables from the
// binary's name and needs the name of the .env variable stamped at link time.
func TestDotEnvLogging(t *testing.T) {
	const canary = "canary-value"
	dir := t.TempDir()
	bin := filepath.Join(dir, "nexus-iq-mcp")
	build := exec.Command("go", "build", "-ldflags", "-X github.com/dihedron/nexus-iq-mcp/pkg/metadata.DotEnvVarName=NEXUS_IQ_MCP_DOTENV", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	good := write("good.env", "NEXUS_IQ_USERNAME=admin\n")
	setsLevel := write("level.env", "NEXUS_IQ_MCP_LOG_LEVEL=info\n")
	malformed := write("bad.env", "bad line here\nNEXUS_IQ_PASSWORD="+canary+"\n")
	missing := filepath.Join(dir, "missing.env")

	tests := []struct {
		name   string
		args   []string
		env    []string
		want   []string // in stderr; none means stderr must be empty
		absent []string // not in stderr
	}{
		{name: "loaded, no level: silent", env: []string{"NEXUS_IQ_MCP_DOTENV=" + good}},
		{name: "level set by the file itself", env: []string{"NEXUS_IQ_MCP_DOTENV=" + setsLevel}, want: []string{"level=INFO", "successfully loaded"}},
		{name: "missing, no level: still seen", env: []string{"NEXUS_IQ_MCP_DOTENV=" + missing}, want: []string{"no such file"}},
		{name: "missing, unknown level: still seen", env: []string{"NEXUS_IQ_MCP_DOTENV=" + missing, "NEXUS_IQ_MCP_LOG_LEVEL=loud"}, want: []string{"no such file"}},
		{name: "missing, level error: logged", env: []string{"NEXUS_IQ_MCP_DOTENV=" + missing, "NEXUS_IQ_MCP_LOG_LEVEL=error"}, want: []string{"level=ERROR", "no such file"}},
		{name: "missing, level off: silent", env: []string{"NEXUS_IQ_MCP_DOTENV=" + missing, "NEXUS_IQ_MCP_LOG_LEVEL=off"}},
		{name: "malformed: no values shown", env: []string{"NEXUS_IQ_MCP_DOTENV=" + malformed}, want: []string{"malformed"}, absent: []string{canary, "bad line"}},
		{name: "malformed through the flag: no values shown", args: []string{"--verbose", "--dotenv=" + malformed}, env: []string{"NEXUS_IQ_MCP_LOG_LEVEL=error"}, want: []string{"malformed"}, absent: []string{canary, "bad line"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(bin, append([]string{"version"}, tt.args...)...)
			cmd.Dir = dir
			for _, kv := range os.Environ() {
				if !strings.HasPrefix(kv, "NEXUS_IQ_") {
					cmd.Env = append(cmd.Env, kv)
				}
			}
			cmd.Env = append(cmd.Env, tt.env...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("run: %v\n%s", err, stderr.String())
			}
			got := stderr.String()
			if len(tt.want) == 0 && got != "" {
				t.Errorf("stderr should be empty: %s", got)
			}
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("stderr lacks %q: %s", want, got)
				}
			}
			for _, absent := range tt.absent {
				if strings.Contains(got, absent) {
					t.Errorf("stderr shows %q: %s", absent, got)
				}
			}
			if out := stdout.String(); strings.Contains(out, "level=") || strings.Contains(out, canary) {
				t.Errorf("stdout carries a log or a secret: %s", out)
			}
		})
	}
}
