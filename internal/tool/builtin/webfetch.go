package builtin

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/codingagent/coding-agent/internal/tool"
)

// WebFetchTool fetches and cleans URL content. Requires READ permission.
type WebFetchTool struct {
	tool.BaseTool
	client *http.Client
}

func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

func (t *WebFetchTool) Name() string        { return "web_fetch" }
func (t *WebFetchTool) Description() string { return "获取 URL 的内容并以文本形式返回。适用于阅读文档或 API 响应。" }

func (t *WebFetchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url":    map[string]any{"type": "string", "description": "要获取的 URL"},
			"prompt": map[string]any{"type": "string", "description": "要从页面中提取的信息内容"},
		},
		"required": []string{"url"},
	}
}

func (t *WebFetchTool) ToLLMFormat() map[string]any {
	return t.BaseTool.ToLLMFormat(t)
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)
var whitespaceRe = regexp.MustCompile(`\s+`)

func (t *WebFetchTool) Execute(ctx tool.ToolContext, args map[string]any) tool.ToolResult {
	urlStr, _ := args["url"].(string)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return tool.Fail("无效的 URL: " + err.Error())
	}
	req.Header.Set("User-Agent", "CodingAgent/1.0")

	resp, err := t.client.Do(req)
	if err != nil {
		return tool.Fail("获取 URL 出错: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return tool.Fail("HTTP " + http.StatusText(resp.StatusCode) + " for URL: " + urlStr)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
	if err != nil {
		return tool.Fail("读取响应出错: " + err.Error())
	}

	// Strip HTML tags and compress whitespace
	cleaned := htmlTagRe.ReplaceAllString(string(body), " ")
	cleaned = whitespaceRe.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	const maxLen = 50_000
	truncated := len(cleaned) > maxLen
	if truncated {
		cleaned = cleaned[:maxLen] + "\n... [已截断]"
	}

	return tool.ToolResult{Success: true, Data: cleaned, Truncated: truncated}
}
