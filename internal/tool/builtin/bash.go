package builtin

import (
	"github.com/codingagent/coding-agent/internal/sandbox"
	"github.com/codingagent/coding-agent/internal/tool"
)

// BashTool executes shell commands within the sandbox. Requires EXECUTE permission.
type BashTool struct {
	tool.BaseTool
}

func (t *BashTool) Name() string        { return "bash" }
func (t *BashTool) Description() string { return "在工作目录中执行 bash 命令。命令有超时和输出大小限制。" }
func (t *BashTool) RequiredPermission() sandbox.PermissionLevel { return sandbox.LevelExecute }

func (t *BashTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command":     map[string]any{"type": "string", "description": "要执行的 bash 命令"},
			"description": map[string]any{"type": "string", "description": "该命令的简短说明"},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) ToLLMFormat() map[string]any {
	return t.BaseTool.ToLLMFormat(t)
}

func (t *BashTool) Execute(ctx tool.ToolContext, args map[string]any) tool.ToolResult {
	command, _ := args["command"].(string)
	result := ctx.Sandbox.ExecuteCommand(command)
	if result.Success {
		return tool.Ok(result.Output)
	}
	return tool.Fail(result.Error)
}
