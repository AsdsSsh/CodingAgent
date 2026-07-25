package tool

import (
	"path/filepath"
	"strings"

	"github.com/codingagent/coding-agent/internal/sandbox"
)

// ToolContext carries workspace root and sandbox reference for each tool invocation.
type ToolContext struct {
	Workspace string
	Sandbox   sandbox.Sandbox
}

// ResolvePath resolves a user-provided path: absolute paths pass through,
// relative paths are resolved against the workspace root.
func (ctx ToolContext) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(ctx.Workspace, path)
}

// IsInWorkspace checks whether a resolved path is within the workspace boundary (path traversal guard).
func (ctx ToolContext) IsInWorkspace(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absPath = filepath.Clean(absPath)
	absWS, err := filepath.Abs(ctx.Workspace)
	if err != nil {
		return false
	}
	absWS = filepath.Clean(absWS)

	// On Windows, normalize to same case for comparison
	return strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(absWS)+string(filepath.Separator)) ||
		strings.EqualFold(absPath, absWS)
}
