package builtin

import (
	"os"
	"strings"
	"testing"
)

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		glob   string
		should string
	}{
		{"*.go", `^[^/]*\.go$`},
		{"**/*.go", `^.*/[^/]*\.go$`},
		{"src/**", `^src/.*$`},
		{"file?.txt", `^file[^/]\.txt$`},
		{"{a,b}.go", `^(?:a|b)\.go$`},
	}

	for _, tt := range tests {
		result := globToRegex(tt.glob)
		if result != tt.should {
			t.Errorf("globToRegex(%q):\n  expected %q\n  got      %q", tt.glob, tt.should, result)
		}
	}
}

func TestGrepSearchFindsMatches(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/test.txt", []byte("hello world\nfoo bar\nhello again\n"), 0644)

	result := grepSearch(dir, "hello")
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	lines := strings.Split(strings.TrimSpace(result.Data), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 matches, got %d:\n%s", len(lines), result.Data)
	}
}

func TestGrepSearchNoMatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/test.txt", []byte("foo bar baz\n"), 0644)

	result := grepSearch(dir, "nonexistent")
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if !strings.Contains(result.Data, "未找到匹配") {
		t.Errorf("expected 'no matches' message, got: %s", result.Data)
	}
}

func TestGrepSearchInvalidRegex(t *testing.T) {
	dir := t.TempDir()
	result := grepSearch(dir, "[invalid")
	if result.Success {
		t.Error("expected failure for invalid regex")
	}
}
