package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

// AnthropicProvider implements Provider for the Anthropic Messages API.
// Uses net/http directly — no SDK dependency needed.
type AnthropicProvider struct {
	httpClient   *http.Client
	apiKey       string
	model        string
	maxTokens    int
	contextWindow int
}

// NewAnthropicProvider creates a provider for the Anthropic Messages API.
func NewAnthropicProvider(apiKey, model string, maxTokens, contextWindow int) *AnthropicProvider {
	return &AnthropicProvider{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		apiKey:        apiKey,
		model:         model,
		maxTokens:     maxTokens,
		contextWindow: contextWindow,
	}
}

func (p *AnthropicProvider) ModelName() string     { return p.model }
func (p *AnthropicProvider) ContextWindow() int     { return p.contextWindow }

func (p *AnthropicProvider) Chat(ctx context.Context, messages []Message, tools []map[string]any) (LlmResponse, error) {
	body := map[string]any{
		"model":      p.model,
		"max_tokens": p.maxTokens,
	}

	// Anthropic requires system prompt as a top-level field
	if sysPrompt := extractSystemPrompt(messages); sysPrompt != "" {
		body["system"] = sysPrompt
	}

	body["messages"] = convertToAnthropicMessages(messages)

	if len(tools) > 0 {
		body["tools"] = convertToAnthropicTools(tools)
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return LlmResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		return LlmResponse{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return LlmResponse{}, fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return LlmResponse{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return LlmResponse{}, fmt.Errorf("anthropic api error %d: %s", resp.StatusCode, string(respBody))
	}

	return parseAnthropicResponse(respBody)
}

// extractSystemPrompt collects all SYSTEM messages into a single string.
func extractSystemPrompt(messages []Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		if msg.Role == RoleSystem {
			if sb.Len() > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString(msg.Content)
		}
	}
	return sb.String()
}

// convertToAnthropicMessages transforms unified messages to Anthropic content-block format.
func convertToAnthropicMessages(messages []Message) []map[string]any {
	var apiMsgs []map[string]any
	for _, msg := range messages {
		if msg.Role == RoleSystem {
			continue // handled in top-level system field
		}

		apiMsg := map[string]any{
			"role": string(msg.Role),
		}

		switch {
		case msg.Role == RoleAssistant && len(msg.ToolCalls) > 0:
			var blocks []map[string]any
			if msg.Content != "" {
				blocks = append(blocks, map[string]any{
					"type": "text",
					"text": msg.Content,
				})
			}
			for _, tc := range msg.ToolCalls {
				args := tc.Arguments
				if args == nil {
					args = map[string]any{}
				}
				blocks = append(blocks, map[string]any{
					"type":  "tool_use",
					"id":    tc.ID,
					"name":  tc.Name,
					"input": args,
				})
			}
			apiMsg["content"] = blocks

		case msg.Role == RoleTool:
			apiMsg["content"] = []map[string]any{{
				"type":         "tool_result",
				"tool_use_id":  msg.ToolCallID,
				"content":      msg.Content,
			}}

		default:
			content := msg.Content
			if content == "" {
				content = ""
			}
			apiMsg["content"] = content
		}

		apiMsgs = append(apiMsgs, apiMsg)
	}
	return apiMsgs
}

// convertToAnthropicTools transforms provider-agnostic tool defs to Anthropic format.
func convertToAnthropicTools(tools []map[string]any) []map[string]any {
	var apiTools []map[string]any
	for _, tool := range tools {
		schema := map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
		if is, ok := tool["input_schema"]; ok {
			if m, ok := is.(map[string]any); ok {
				schema = m
			}
		}
		apiTools = append(apiTools, map[string]any{
			"name":         tool["name"],
			"description":  tool["description"],
			"input_schema": schema,
		})
	}
	return apiTools
}

// parseAnthropicResponse parses the Anthropic Messages API JSON response.
func parseAnthropicResponse(body []byte) (LlmResponse, error) {
	var result struct {
		Content    []json.RawMessage `json:"content"`
		StopReason string            `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return LlmResponse{}, fmt.Errorf("parse response: %w", err)
	}

	var textBuilder strings.Builder
	var toolCalls []ToolCall

	for _, raw := range result.Content {
		// Peek at the type field to dispatch
		var blockType struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &blockType); err != nil {
			continue
		}

		switch blockType.Type {
		case "text":
			var tb struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(raw, &tb); err == nil {
				if textBuilder.Len() > 0 {
					textBuilder.WriteByte('\n')
				}
				textBuilder.WriteString(tb.Text)
			}

		case "tool_use":
			var tu struct {
				ID    string         `json:"id"`
				Name  string         `json:"name"`
				Input map[string]any `json:"input"`
			}
			if err := json.Unmarshal(raw, &tu); err == nil {
				input := tu.Input
				if input == nil {
					input = map[string]any{}
				}
				toolCalls = append(toolCalls, ToolCall{
					ID:        tu.ID,
					Name:      tu.Name,
					Arguments: input,
				})
			}
		}
	}

	isFinal := result.StopReason == "end_turn" && len(toolCalls) == 0

	return LlmResponse{
		Content:       textBuilder.String(),
		ToolCalls:     toolCalls,
		IsFinalAnswer: isFinal,
		Usage: TokenUsage{
			InputTokens:  result.Usage.InputTokens,
			OutputTokens: result.Usage.OutputTokens,
		},
		StopReason: result.StopReason,
	}, nil
}

func (p *AnthropicProvider) CountTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(text) / 4
}

func (p *AnthropicProvider) CountMessagesTokens(messages []Message) int {
	return BaseProvider{}.CountMessagesTokens(messages, p.CountTokens)
}

func (p *AnthropicProvider) Close() error { return nil }
