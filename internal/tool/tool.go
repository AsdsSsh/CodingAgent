package tool

import "github.com/codingagent/coding-agent/internal/sandbox"

// Tool is the unified contract for all tools (built-in and MCP).
// Each tool declares its name, description, parameter schema, and required permission level.
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	RequiredPermission() sandbox.PermissionLevel
	Execute(ctx ToolContext, args map[string]any) ToolResult
	ToLLMFormat() map[string]any
}

// BaseTool provides default implementations for ToLLMFormat and RequiredPermission.
type BaseTool struct{}

// RequiredPermission defaults to READ — destructive tools override this.
func (BaseTool) RequiredPermission() sandbox.PermissionLevel {
	return sandbox.LevelRead
}

// ToLLMFormat generates the standard LLM tool definition.
func (b BaseTool) ToLLMFormat(t Tool) map[string]any {
	return map[string]any{
		"name":         t.Name(),
		"description":  t.Description(),
		"input_schema": t.Parameters(),
	}
}
