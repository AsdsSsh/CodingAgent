package llm

import (
	"testing"
)

func TestParseOpenAITextResponse(t *testing.T) {
	body := `{
		"choices": [{
			"finish_reason": "stop",
			"message": {"content": "Hello, world!"}
		}],
		"usage": {"prompt_tokens": 10, "completion_tokens": 5}
	}`
	resp, err := parseOpenAIResponse([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got '%s'", resp.Content)
	}
	if !resp.IsFinalAnswer {
		t.Error("expected IsFinalAnswer=true for stop with no tool calls")
	}
	if resp.Usage.InputTokens != 10 || resp.Usage.OutputTokens != 5 {
		t.Errorf("expected 10/5 tokens, got %d/%d", resp.Usage.InputTokens, resp.Usage.OutputTokens)
	}
}

func TestParseOpenAIToolCallsResponse(t *testing.T) {
	body := `{
		"choices": [{
			"finish_reason": "tool_calls",
			"message": {
				"content": null,
				"tool_calls": [{
					"id": "call_001",
					"type": "function",
					"function": {
						"name": "read",
						"arguments": "{\"file_path\":\"test.go\"}"
					}
				}]
			}
		}],
		"usage": {"prompt_tokens": 20, "completion_tokens": 15}
	}`
	resp, err := parseOpenAIResponse([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsFinalAnswer {
		t.Error("expected IsFinalAnswer=false for tool_calls response")
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "read" {
		t.Errorf("expected 'read', got '%s'", resp.ToolCalls[0].Name)
	}
}

func TestParseOpenAIDeepSeekReasoningContent(t *testing.T) {
	body := `{
		"choices": [{
			"finish_reason": "stop",
			"message": {
				"content": "The answer is 42.",
				"reasoning_content": "Let me think about this..."
			}
		}],
		"usage": {"prompt_tokens": 10, "completion_tokens": 20}
	}`
	resp, err := parseOpenAIResponse([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ReasoningContent != "Let me think about this..." {
		t.Errorf("expected reasoning_content, got '%s'", resp.ReasoningContent)
	}
	if resp.Content != "The answer is 42." {
		t.Errorf("expected content, got '%s'", resp.Content)
	}
}

func TestParseOpenAIEmptyChoices(t *testing.T) {
	body := `{"choices": [], "usage": {"prompt_tokens": 0, "completion_tokens": 0}}`
	resp, err := parseOpenAIResponse([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsFinalAnswer {
		t.Error("expected IsFinalAnswer=true for empty choices")
	}
	if resp.StopReason != "empty" {
		t.Errorf("expected stop_reason='empty', got '%s'", resp.StopReason)
	}
}
