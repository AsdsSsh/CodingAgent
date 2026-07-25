package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codingagent/coding-agent/internal/sandbox"
	"github.com/codingagent/coding-agent/internal/tool"
)

type mockSandbox struct {
	workspace string
}

func (m *mockSandbox) PolicyEngine() *sandbox.PolicyEngine { return sandbox.NewPolicyEngineDefault(sandbox.LevelWrite) }
func (m *mockSandbox) ExecuteCommand(cmd string) sandbox.CommandResult { return sandbox.CommandOk("") }
func (m *mockSandbox) IsInWorkspace(path string) bool {
	rel, _ := filepath.Rel(m.workspace, path)
	return rel != "" && rel[0] != '.'
}

func TestFileEditToolUniqueMatch(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	content := "line1\nline2\nline3\n"
	os.WriteFile(filePath, []byte(content), 0644)

	et := &FileEditTool{}
	ctx := tool.ToolContext{
		Workspace: dir,
		Sandbox:   &mockSandbox{workspace: dir},
	}

	result := et.Execute(ctx, map[string]any{
		"file_path":  filePath,
		"old_string": "line2",
		"new_string": "LINE_TWO",
	})

	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}

	updated, _ := os.ReadFile(filePath)
	expected := "line1\nLINE_TWO\nline3\n"
	if string(updated) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(updated))
	}
}

func TestFileEditToolNoMatch(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	os.WriteFile(filePath, []byte("hello world"), 0644)

	et := &FileEditTool{}
	ctx := tool.ToolContext{
		Workspace: dir,
		Sandbox:   &mockSandbox{workspace: dir},
	}

	result := et.Execute(ctx, map[string]any{
		"file_path":  filePath,
		"old_string": "nonexistent",
		"new_string": "replacement",
	})

	if result.Success {
		t.Error("expected failure for no match")
	}
}

func TestFileEditToolMultipleMatch(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	os.WriteFile(filePath, []byte("dup\ndup\n"), 0644)

	et := &FileEditTool{}
	ctx := tool.ToolContext{
		Workspace: dir,
		Sandbox:   &mockSandbox{workspace: dir},
	}

	result := et.Execute(ctx, map[string]any{
		"file_path":  filePath,
		"old_string": "dup",
		"new_string": "unique",
	})

	if result.Success {
		t.Error("expected failure for multiple matches")
	}
}
