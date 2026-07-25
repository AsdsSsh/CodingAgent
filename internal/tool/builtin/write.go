package builtin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codingagent/coding-agent/internal/sandbox"
	"github.com/codingagent/coding-agent/internal/tool"
)

// FileWriteTool creates or overwrites files. Requires WRITE permission.
type FileWriteTool struct {
	tool.BaseTool
}

func (t *FileWriteTool) Name() string        { return "write" }
func (t *FileWriteTool) Description() string { return "创建新文件或覆盖现有文件，写入指定内容。" }
func (t *FileWriteTool) RequiredPermission() sandbox.PermissionLevel { return sandbox.LevelWrite }

func (t *FileWriteTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{"type": "string", "description": "要写入的文件路径"},
			"content":   map[string]any{"type": "string", "description": "要写入文件的内容"},
		},
		"required": []string{"file_path", "content"},
	}
}

func (t *FileWriteTool) ToLLMFormat() map[string]any {
	return t.BaseTool.ToLLMFormat(t)
}

func (t *FileWriteTool) Execute(ctx tool.ToolContext, args map[string]any) tool.ToolResult {
	filePath, _ := args["file_path"].(string)
	content, _ := args["content"].(string)
	path := ctx.ResolvePath(filePath)

	// Path traversal guard
	if !ctx.IsInWorkspace(path) {
		return tool.Fail("拒绝访问: " + filePath + " 位于工作区之外")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return tool.Fail("创建目录出错: " + err.Error())
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return tool.Fail("写入文件出错: " + err.Error())
	}

	return tool.Ok(fmt.Sprintf("文件已写入: %s (%d 字节)", filePath, len(content)))
}
