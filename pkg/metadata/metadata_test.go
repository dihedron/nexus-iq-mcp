package metadata

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	VersionMajor, VersionMinor, VersionPatch, Description = "1", "2", "3", "test build"
	var b bytes.Buffer
	Print(&b)
	if !strings.Contains(b.String(), "v1.2.3") || !strings.Contains(b.String(), "test build") {
		t.Errorf("Print wrote %q", b.String())
	}
	b.Reset()
	PrintFull(&b)
	if !strings.Contains(b.String(), "Build Commit") {
		t.Errorf("PrintFull did not use its writer: %q", b.String())
	}
	if n := strings.Count(b.String(), "Git Time"); n != 1 {
		t.Errorf("PrintFull wrote Git Time %d times", n)
	}
}
