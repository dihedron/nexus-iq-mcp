// Package docs holds the test that keeps the tables of the release documents
// readable: the documents themselves are Markdown, built to PDF by build.sh.
package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var separator = regexp.MustCompile(`^\|( *:?-+:? *\|)+$`)

// TestTablesSetColumnWidths: the dashes set a table's proportions, so equal
// ones give equal columns, which rarely fit.
func TestTablesSetColumnWidths(t *testing.T) {
	files, err := filepath.Glob("*/[0-9][0-9]-*.md")
	if err != nil || len(files) == 0 {
		t.Fatalf("no chapters (%v)", err)
	}
	for _, file := range files {
		data, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if !separator.MatchString(strings.TrimSpace(line)) {
				continue
			}
			lengths := map[int]bool{}
			for _, cell := range strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|") {
				lengths[len(strings.Trim(strings.TrimSpace(cell), ":"))] = true
			}
			if len(lengths) == 1 {
				t.Errorf("%s:%d: the table's columns all have the same width: set their proportions with the dashes' lengths", file, i+1)
			}
		}
	}
}
