package serve

import "testing"

func TestExecuteRejectsIncompleteConfiguration(t *testing.T) {
	tests := []struct {
		name, url, username, password string
	}{
		{"no url", "", "user", "pass"},
		{"no username", "http://localhost:8070", "", "pass"},
		{"no password", "http://localhost:8070", "user", ""},
		{"bad url", "ftp://localhost", "user", "pass"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NEXUS_IQ_URL", tt.url)
			t.Setenv("NEXUS_IQ_USERNAME", tt.username)
			t.Setenv("NEXUS_IQ_PASSWORD", tt.password)
			if err := (&Serve{}).Execute(nil); err == nil {
				t.Error("expected an error")
			}
		})
	}
}
