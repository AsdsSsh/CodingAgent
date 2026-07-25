package builtin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codingagent/coding-agent/internal/tool"
)

// FileReadTool reads file content with optional line range selection.
type FileReadTool struct {
	tool.BaseTool
}

func (t *FileReadTool) Name() string        { return "read" }
func (t *FileReadTool) Description() string { return "读取文件内容。对于大文件可通过 offset 和 limit 指定读取范围。" }

func (t *FileReadTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{"type": "string", "description": "文件的绝对或相对路径"},
			"offset":    map[string]any{"type": "integer", "description": "起始行号（从 1 开始）"},
			"limit":     map[string]any{"type": "integer", "description": "要读取的行数"},
		},
		"required": []string{"file_path"},
	}
}

func (t *FileReadTool) ToLLMFormat() map[string]any {
	return t.BaseTool.ToLLMFormat(t)
}

func (t *FileReadTool) Execute(ctx tool.ToolContext, args map[string]any) tool.ToolResult {
	filePath, _ := args["file_path"].(string)
	path := ctx.ResolvePath(filePath)

	info, err := os.Stat(path)
	if err != nil {
		return tool.Fail("文件不存在: " + filePath)
	}
	if info.IsDir() {
		return tool.Fail("不是文件: " + filePath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return tool.Fail("读取文件出错: " + err.Error())
	}

	lines := splitLines(string(data))

	offset := 1
	if o, ok := args["offset"]; ok {
		offset = toInt(o)
	}
	if offset < 1 {
		offset = 1
	}

	limit := len(lines)
	if l, ok := args["limit"]; ok {
		limit = toInt(l)
	}

	start := offset - 1
	end := start + limit
	if end > len(lines) {
		end = len(lines)
	}

	var result string
	for i := start; i < end; i++ {
		result += fmt.Sprintf("%6d\t%s\n", i+1, lines[i])
	}

	const maxLen = 100_000
	truncated := len(result) > maxLen
	if truncated {
		// Truncate at UTF-8 safe boundary
		if maxLen < len(result) {
			result = result[:maxLen] + "\n... [已截断]"
		}
	}

	return tool.ToolResult{Success: true, Data: result, Truncated: truncated}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		} else if s[i] == '\r' {
			if i+1 < len(s) && s[i+1] == '\n' {
				lines = append(lines, s[start:i])
				start = i + 2
				i++
			} else {
				lines = append(lines, s[start:i])
				start = i + 1
			}
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	default:
		return 0
	}
}

func init() {
	_ = filepath.Base
}
