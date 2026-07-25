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

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

// OpenAIProvider implements Provider for OpenAI and OpenAI-compatible APIs.
// Supports GPT-4/GPT-4o, DeepSeek, Ollama, vLLM, LM Studio, etc.
type OpenAIProvider struct {
	httpClient    *http.Client
	apiKey        string
	model         string
	maxTokens     int
	contextWindow int
	baseURL       string
	temperature   *float64
}

// NewOpenAIProvider creates a provider for OpenAI-compatible APIs.
func NewOpenAIProvider(apiKey, model string, maxTokens, contextWindow int, baseURL string, temperature *float64) *OpenAIProvider {
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &OpenAIProvider{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		apiKey:        apiKey,
		model:         model,
		maxTokens:     maxTokens,
		contextWindow: contextWindow,
		baseURL:       baseURL,
		temperature:   temperature,
	}
}

func (p *OpenAIProvider) ModelName() string     { return p.model }
func (p *OpenAIProvider) ContextWindow() int     { return p.contextWindow }

func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, tools []map[string]any) (LlmResponse, error) {
	body := map[string]any{
		"model":      p.model,
		"max_tokens": p.maxTokens,
	}

	if p.temperature != nil {
		body["temperature"] = *p.temperature
	}

	body["messages"] = convertToOpenAIMessages(messages)

	if len(tools) > 0 {
		body["tools"] = convertToOpenAITools(tools)
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return LlmResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return LlmResponse{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

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
		return LlmResponse{}, fmt.Errorf("openai api error %d: %s", resp.StatusCode, string(respBody))
	}

	return parseOpenAIResponse(respBody)
}

// convertToOpenAIMessages transforms unified messages to OpenAI format.
func convertToOpenAIMessages(messages []Message) []map[string]any {
	var apiMsgs []map[string]any
	for _, msg := range messages {
		apiMsg := map[string]any{
			"role": string(msg.Role),
		}

		switch {
		case msg.Role == RoleAssistant && len(msg.ToolCalls) > 0:
			// DeepSeek thinking mode: passthrough reasoning_content
			if msg.ReasoningContent != "" {
				apiMsg["reasoning_content"] = msg.ReasoningContent
			}

			if msg.Content != "" {
				apiMsg["content"] = msg.Content
			} else {
				apiMsg["content"] = nil
			}

			var toolCallList []map[string]any
			for _, tc := range msg.ToolCalls {
				argsJSON, _ := json.Marshal(tc.Arguments)
				toolCallList = append(toolCallList, map[string]any{
					"id":   tc.ID,
					"type": "function",
					"function": map[string]any{
						"name":      tc.Name,
						"arguments": string(argsJSON),
					},
				})
			}
			apiMsg["tool_calls"] = toolCallList

		case msg.Role == RoleTool:
			apiMsg["content"] = msg.Content
			apiMsg["tool_call_id"] = msg.ToolCallID

		default:
			content := msg.Content
			if content == "" {
				content = ""
			}
			apiMsg["content"] = content
			// DeepSeek: passthrough reasoning_content for plain assistant messages too
			if msg.Role == RoleAssistant && msg.ReasoningContent != "" {
				apiMsg["reasoning_content"] = msg.ReasoningContent
			}
		}

		apiMsgs = append(apiMsgs, apiMsg)
	}
	return apiMsgs
}

// convertToOpenAITools transforms provider-agnostic tool defs to OpenAI function format.
func convertToOpenAITools(tools []map[string]any) []map[string]any {
	var apiTools []map[string]any
	for _, tool := range tools {
		params := map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
		if is, ok := tool["input_schema"]; ok {
			if m, ok := is.(map[string]any); ok {
				params = m
			}
		}
		apiTools = append(apiTools, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool["name"],
				"description": tool["description"],
				"parameters":  params,
			},
		})
	}
	return apiTools
}

// parseOpenAIResponse parses the OpenAI Chat Completions API JSON response.
func parseOpenAIResponse(body []byte) (LlmResponse, error) {
	var result struct {
		Choices []struct {
			FinishReason *string `json:"finish_reason"`
			Message      struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content"` // DeepSeek
				ToolCalls        []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return LlmResponse{}, fmt.Errorf("parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return LlmResponse{
			Content:       "",
			IsFinalAnswer: true,
			StopReason:    "empty",
		}, nil
	}

	choice := result.Choices[0]
	msg := choice.Message
	finishReason := ""
	if choice.FinishReason != nil {
		finishReason = *choice.FinishReason
	}

	// Parse tool calls
	var toolCalls []ToolCall
	for _, tc := range msg.ToolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			args = map[string]any{}
		}
		toolCalls = append(toolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		})
	}

	isFinal := finishReason == "stop" && len(toolCalls) == 0

	return LlmResponse{
		Content:          msg.Content,
		ToolCalls:        toolCalls,
		IsFinalAnswer:    isFinal,
		StopReason:       finishReason,
		ReasoningContent: msg.ReasoningContent,
		Usage: TokenUsage{
			InputTokens:  result.Usage.PromptTokens,
			OutputTokens: result.Usage.CompletionTokens,
		},
	}, nil
}

func (p *OpenAIProvider) CountTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(text) / 4
}

func (p *OpenAIProvider) CountMessagesTokens(messages []Message) int {
	return BaseProvider{}.CountMessagesTokens(messages, p.CountTokens)
}

func (p *OpenAIProvider) Close() error { return nil }
