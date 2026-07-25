package llm

import "maps"

// Role represents the role of a message in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is the unified message type shared across all LLM providers.
type Message struct {
	Role             Role
	Content          string
	ToolCalls        []ToolCall
	ToolCallID       string
	ReasoningContent string // DeepSeek thinking mode passthrough
}

// ToolCall represents a single tool invocation within an assistant message.
type ToolCall struct {
	ID        string
	Name      string
	Arguments map[string]any
}

// System creates a system message.
func System(content string) Message {
	return Message{Role: RoleSystem, Content: content}
}

// User creates a user message.
func User(content string) Message {
	return Message{Role: RoleUser, Content: content}
}

// Assistant creates a plain assistant text message.
func Assistant(content string) Message {
	return Message{Role: RoleAssistant, Content: content}
}

// AssistantWithReasoning creates an assistant message with reasoning content (DeepSeek).
func AssistantWithReasoning(content, reasoning string) Message {
	return Message{Role: RoleAssistant, Content: content, ReasoningContent: reasoning}
}

// AssistantWithToolCalls creates an assistant message carrying tool calls.
func AssistantWithToolCalls(tc ToolCall) Message {
	return Message{Role: RoleAssistant, ToolCalls: []ToolCall{tc}}
}

// AssistantWithToolCallsAndReasoning creates an assistant message with tool calls and reasoning.
func AssistantWithToolCallsAndReasoning(tc ToolCall, reasoning string) Message {
	return Message{Role: RoleAssistant, ToolCalls: []ToolCall{tc}, ReasoningContent: reasoning}
}

// ToolResult creates a tool result message paired with its tool call via ID.
func ToolResult(toolCallID, content string) Message {
	return Message{Role: RoleTool, Content: content, ToolCallID: toolCallID}
}

// CloneToolCall returns a deep copy of a ToolCall (arguments map is copied).
func CloneToolCall(tc ToolCall) ToolCall {
	cloned := ToolCall{ID: tc.ID, Name: tc.Name}
	if tc.Arguments != nil {
		cloned.Arguments = maps.Clone(tc.Arguments)
	}
	return cloned
}
