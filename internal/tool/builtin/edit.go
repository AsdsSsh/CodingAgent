package builtin

import (
	"fmt"
	"os"
	"strings"

	"github.com/codingagent/coding-agent/internal/sandbox"
	"github.com/codingagent/coding-agent/internal/tool"
)

// FileEditTool performs exact string replacement in files.
// Fails if old_string appears zero or more than one time.
type FileEditTool struct {
	tool.BaseTool
}

func (t *FileEditTool) Name() string        { return "edit" }
func (t *FileEditTool) Description() string { return "在文件中执行精确的字符串替换。查找 old_string 并将其替换为 new_string。若 old_string 未找到或出现多次则操作失败。" }
func (t *FileEditTool) RequiredPermission() sandbox.PermissionLevel { return sandbox.LevelWrite }

func (t *FileEditTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path":  map[string]any{"type": "string", "description": "要编辑的文件路径"},
			"old_string": map[string]any{"type": "string", "description": "要查找并替换的文本"},
			"new_string": map[string]any{"type": "string", "description": "替换后的文本"},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}
}

func (t *FileEditTool) ToLLMFormat() map[string]any {
	return t.BaseTool.ToLLMFormat(t)
}

func (t *FileEditTool) Execute(ctx tool.ToolContext, args map[string]any) tool.ToolResult {
	filePath, _ := args["file_path"].(string)
	oldString, _ := args["old_string"].(string)
	newString, _ := args["new_string"].(string)
	path := ctx.ResolvePath(filePath)

	data, err := os.ReadFile(path)
	if err != nil {
		return tool.Fail("文件不存在: " + filePath)
	}

	content := string(data)
	count := strings.Count(content, oldString)

	if count == 0 {
		return tool.Fail("在文件中未找到 old_string")
	}
	if count > 1 {
		return tool.Fail(fmt.Sprintf(
			"old_string 在文件中出现了 %d 次。请使用更长的字符串，包含更多上下文以确保唯一匹配。", count,
		))
	}

	updated := strings.Replace(content, oldString, newString, 1)
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return tool.Fail("编辑文件出错: " + err.Error())
	}

	return tool.Ok("文件已编辑: " + filePath)
}
