package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codingagent/coding-agent/internal/tool"
)

// FileSearchTool supports two modes: glob (file name matching) and grep (content searching).
type FileSearchTool struct {
	tool.BaseTool
}

func (t *FileSearchTool) Name() string { return "search" }
func (t *FileSearchTool) Description() string {
	return "按 glob 模式搜索文件名，或按正则表达式搜索文件内容。"
}

func (t *FileSearchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{"type": "string", "description": "glob 模式（如 '**/*.java'）或 grep 的正则表达式"},
			"mode":    map[string]any{"type": "string", "description": "'glob' 按文件名查找，'grep' 搜索文件内容"},
			"path":    map[string]any{"type": "string", "description": "搜索的目录（默认为工作区根目录）"},
		},
		"required": []string{"pattern", "mode"},
	}
}

func (t *FileSearchTool) ToLLMFormat() map[string]any {
	return t.BaseTool.ToLLMFormat(t)
}

func (t *FileSearchTool) Execute(ctx tool.ToolContext, args map[string]any) tool.ToolResult {
	pattern, _ := args["pattern"].(string)
	mode, _ := args["mode"].(string)
	searchPath := "."
	if sp, ok := args["path"].(string); ok && sp != "" {
		searchPath = sp
	}

	baseDir := ctx.ResolvePath(searchPath)
	if info, err := os.Stat(baseDir); err != nil || !info.IsDir() {
		return tool.Fail("目录不存在: " + searchPath)
	}

	switch mode {
	case "glob":
		return globSearch(baseDir, pattern)
	case "grep":
		return grepSearch(baseDir, pattern)
	default:
		return tool.Fail("未知模式: " + mode + "。请使用 'glob' 或 'grep'。")
	}
}

func globSearch(baseDir, pattern string) tool.ToolResult {
	regex := globToRegex(pattern)
	re, err := regexp.Compile(regex)
	if err != nil {
		return tool.Fail("无效的 glob 模式: " + err.Error())
	}

	var results []string
	count := 0

	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if count >= 200 {
			return filepath.SkipAll
		}
		rel, _ := filepath.Rel(baseDir, path)
		// Normalize to forward slashes for matching
		rel = filepath.ToSlash(rel)
		if re.MatchString(rel) {
			results = append(results, rel)
			count++
		}
		return nil
	})

	if len(results) == 0 {
		return tool.Ok("未找到匹配的文件: " + pattern)
	}
	return tool.Ok(strings.Join(results, "\n"))
}

func grepSearch(baseDir, pattern string) tool.ToolResult {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return tool.Fail("无效的正则表达式: " + err.Error())
	}

	skipExts := map[string]bool{
		".class": true, ".jar": true, ".png": true, ".jpg": true,
		".exe": true, ".dll": true, ".so": true, ".o": true,
		".zip": true, ".gz": true, ".tar": true,
	}

	var results []string
	matchCount := 0
	fileCount := 0

	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if fileCount >= 500 || matchCount >= 200 {
			return filepath.SkipAll
		}

		ext := strings.ToLower(filepath.Ext(path))
		if skipExts[ext] {
			return nil
		}

		fileCount++
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // silently skip unreadable files
		}

		lines := strings.Split(string(data), "\n")
		rel, _ := filepath.Rel(baseDir, path)
		for i, line := range lines {
			if matchCount >= 200 {
				break
			}
			if re.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d: %s", rel, i+1, strings.TrimSpace(line)))
				matchCount++
			}
		}
		return nil
	})

	if len(results) == 0 {
		return tool.Ok("未找到匹配: " + pattern)
	}
	return tool.Ok(strings.Join(results, "\n"))
}

// globToRegex converts a shell-style glob to a regex pattern.
// Supports: ** (cross-directory), * (single-level), ? (single char), {a,b} (alternation).
func globToRegex(glob string) string {
	var sb strings.Builder
	sb.WriteByte('^')
	for i := 0; i < len(glob); i++ {
		c := glob[i]
		switch c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				sb.WriteString(".*")
				i++
			} else {
				sb.WriteString("[^/]*")
			}
		case '?':
			sb.WriteString("[^/]")
		case '.':
			sb.WriteString("\\.")
		case '{':
			sb.WriteString("(?:")
		case '}':
			sb.WriteByte(')')
		case ',':
			sb.WriteByte('|')
		default:
			sb.WriteByte(c)
		}
	}
	sb.WriteByte('$')
	return sb.String()
}
