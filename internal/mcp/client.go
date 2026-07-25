package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"

	"github.com/codingagent/coding-agent/internal/config"
	"github.com/codingagent/coding-agent/internal/sandbox"
	"github.com/codingagent/coding-agent/internal/tool"
)

// Client is an MCP (Model Context Protocol) client that connects to an MCP server.
type Client struct {
	cfg    config.MCPServerConfig
	conn   MCPConnection
	mu     sync.Mutex
	tools  []MCPToolDef
	closed bool
}

// MCPConnection abstracts the transport layer (stdio or HTTP).
type MCPConnection interface {
	Send(request map[string]any) (map[string]any, error)
	Close() error
}

// MCPToolDef is an MCP tool definition as returned by tools/list.
type MCPToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// NewClient creates an MCP client from server configuration.
func NewClient(cfg config.MCPServerConfig) (*Client, error) {
	var conn MCPConnection
	switch cfg.Transport {
	case "stdio":
		conn = newStdioConnection(cfg.Command, cfg.Args, cfg.Env)
	case "http":
		conn = newHTTPConnection(cfg.URL)
	default:
		return nil, fmt.Errorf("unsupported MCP transport: %s (use 'stdio' or 'http')", cfg.Transport)
	}

	return &Client{
		cfg:  cfg,
		conn: conn,
	}, nil
}

// ListTools retrieves the tool list from the MCP server.
func (c *Client) ListTools() ([]MCPToolDef, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, fmt.Errorf("MCP client is closed")
	}

	resp, err := c.conn.Send(map[string]any{
		"jsonrpc": "2.0",
		"method":  "tools/list",
		"id":      1,
	})
	if err != nil {
		return nil, fmt.Errorf("MCP tools/list failed: %w", err)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("MCP tools/list: unexpected response format")
	}

	toolsRaw, ok := result["tools"].([]any)
	if !ok {
		return nil, fmt.Errorf("MCP tools/list: missing tools array")
	}

	var tools []MCPToolDef
	for _, t := range toolsRaw {
		tm, ok := t.(map[string]any)
		if !ok {
			continue
		}
		td := MCPToolDef{}
		if name, ok := tm["name"].(string); ok {
			td.Name = name
		}
		if desc, ok := tm["description"].(string); ok {
			td.Description = desc
		}
		if schema, ok := tm["inputSchema"].(map[string]any); ok {
			td.InputSchema = schema
		}
		tools = append(tools, td)
	}

	c.tools = tools
	return tools, nil
}

// CallTool invokes a tool on the MCP server.
func (c *Client) CallTool(name string, args map[string]any) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return "", fmt.Errorf("MCP client is closed")
	}

	resp, err := c.conn.Send(map[string]any{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"id":      2,
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	})
	if err != nil {
		return "", fmt.Errorf("MCP tools/call failed: %w", err)
	}

	if errObj, ok := resp["error"]; ok {
		return "", fmt.Errorf("MCP error: %v", errObj)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("MCP tools/call: unexpected response format")
	}

	// Extract content from result
	content, ok := result["content"].([]any)
	if !ok {
		return fmt.Sprintf("%v", result), nil
	}

	var sb strings.Builder
	for _, c := range content {
		if cm, ok := c.(map[string]any); ok {
			if text, ok := cm["text"].(string); ok {
				sb.WriteString(text)
			}
		}
	}
	return sb.String(), nil
}

// Close terminates the MCP connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return c.conn.Close()
}

// ─── MCP Tool Adapter ───

// ToolAdapter wraps an MCP tool as our Tool interface.
type ToolAdapter struct {
	client   *Client
	toolDef  MCPToolDef
	permLevel sandbox.PermissionLevel
}

// NewToolAdapter creates a Tool interface wrapper for an MCP tool.
func NewToolAdapter(client *Client, toolDef MCPToolDef) *ToolAdapter {
	return &ToolAdapter{
		client:    client,
		toolDef:   toolDef,
		permLevel: sandbox.LevelRead,
	}
}

func (a *ToolAdapter) Name() string                        { return a.toolDef.Name }
func (a *ToolAdapter) Description() string                 { return a.toolDef.Description }
func (a *ToolAdapter) RequiredPermission() sandbox.PermissionLevel { return a.permLevel }

func (a *ToolAdapter) Parameters() map[string]any {
	if a.toolDef.InputSchema != nil {
		return a.toolDef.InputSchema
	}
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (a *ToolAdapter) ToLLMFormat() map[string]any {
	return map[string]any{
		"name":         a.toolDef.Name,
		"description":  a.toolDef.Description,
		"input_schema": a.Parameters(),
	}
}

func (a *ToolAdapter) Execute(ctx tool.ToolContext, args map[string]any) tool.ToolResult {
	output, err := a.client.CallTool(a.toolDef.Name, args)
	if err != nil {
		return tool.Fail("MCP tool error: " + err.Error())
	}
	return tool.Ok(output)
}

// ─── stdio connection ───

type stdioConnection struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func newStdioConnection(command string, args []string, env map[string]string) *stdioConnection {
	cmd := exec.Command(command, args...)
	cmd.Env = cmd.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	return &stdioConnection{cmd: cmd}
}

func (s *stdioConnection) Send(request map[string]any) (map[string]any, error) {
	if s.cmd.Process == nil {
		// First call — start the process
		var err error
		s.stdin, err = s.cmd.StdinPipe()
		if err != nil {
			return nil, fmt.Errorf("stdin pipe: %w", err)
		}
		s.stdout, err = s.cmd.StdoutPipe()
		if err != nil {
			return nil, fmt.Errorf("stdout pipe: %w", err)
		}
		if err := s.cmd.Start(); err != nil {
			return nil, fmt.Errorf("start MCP server: %w", err)
		}
	}

	reqJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	_, err = fmt.Fprintf(s.stdin, "%s\n", reqJSON)
	if err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	reader := bufio.NewReader(s.stdout)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return resp, nil
}

func (s *stdioConnection) Close() error {
	if s.stdin != nil {
		s.stdin.Close()
	}
	if s.cmd.Process != nil {
		return s.cmd.Process.Kill()
	}
	return nil
}

// ─── HTTP connection ───

type httpConnection struct {
	url    string
	client *http.Client
}

func newHTTPConnection(url string) *httpConnection {
	return &httpConnection{
		url:    url,
		client: &http.Client{},
	}
}

func (h *httpConnection) Send(request map[string]any) (map[string]any, error) {
	reqJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", h.url, strings.NewReader(string(reqJSON)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}

func (h *httpConnection) Close() error {
	return nil
}
