package llm

import (
	"encoding/json"
	"testing"
)

func TestParseAnthropicTextResponse(t *testing.T) {
	body := `{
		"content": [{"type": "text", "text": "Hello, world!"}],
		"stop_reason": "end_turn",
		"usage": {"input_tokens": 10, "output_tokens": 5}
	}`
	resp, err := parseAnthropicResponse([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got '%s'", resp.Content)
	}
	if !resp.IsFinalAnswer {
		t.Error("expected IsFinalAnswer=true for end_turn with no tool calls")
	}
	if resp.Usage.InputTokens != 10 || resp.Usage.OutputTokens != 5 {
		t.Errorf("expected 10/5 tokens, got %d/%d", resp.Usage.InputTokens, resp.Usage.OutputTokens)
	}
}

func TestParseAnthropicToolUseResponse(t *testing.T) {
	body := `{
		"content": [
			{"type": "text", "text": "Let me read that file."},
			{"type": "tool_use", "id": "tool_001", "name": "read", "input": {"file_path": "test.go"}}
		],
		"stop_reason": "tool_use",
		"usage": {"input_tokens": 20, "output_tokens": 15}
	}`
	resp, err := parseAnthropicResponse([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsFinalAnswer {
		t.Error("expected IsFinalAnswer=false for tool_use response")
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "read" {
		t.Errorf("expected tool name 'read', got '%s'", resp.ToolCalls[0].Name)
	}
	if resp.ToolCalls[0].Arguments["file_path"] != "test.go" {
		t.Errorf("expected file_path='test.go', got '%v'", resp.ToolCalls[0].Arguments["file_path"])
	}
}

func TestParseAnthropicMultipleContentBlocks(t *testing.T) {
	body := `{
		"content": [
			{"type": "text", "text": "First."},
			{"type": "text", "text": "Second."}
		],
		"stop_reason": "end_turn",
		"usage": {"input_tokens": 5, "output_tokens": 3}
	}`
	resp, err := parseAnthropicResponse([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "First.\nSecond." {
		t.Errorf("expected concatenated text, got '%s'", resp.Content)
	}
}

func TestExtractSystemPrompt(t *testing.T) {
	messages := []Message{
		System("You are helpful."),
		System("Be concise."),
		User("Hello"),
	}
	result := extractSystemPrompt(messages)
	if result != "You are helpful.\n\nBe concise." {
		t.Errorf("unexpected system prompt: '%s'", result)
	}
}

func TestConvertToAnthropicTools(t *testing.T) {
	tools := []map[string]any{{
		"name":        "read",
		"description": "Read a file",
		"input_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file_path": map[string]any{"type": "string"},
			},
		},
	}}
	result := convertToAnthropicTools(tools)
	if len(result) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(result))
	}

	jsonBytes, _ := json.Marshal(result[0])
	var parsed map[string]any
	json.Unmarshal(jsonBytes, &parsed)

	if parsed["name"] != "read" {
		t.Errorf("expected name 'read', got '%v'", parsed["name"])
	}
	if _, ok := parsed["input_schema"]; !ok {
		t.Error("expected input_schema in tool definition")
	}
}
