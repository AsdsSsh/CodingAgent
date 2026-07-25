package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/codingagent/coding-agent/internal/llm"
	"github.com/codingagent/coding-agent/internal/sandbox"
	"github.com/codingagent/coding-agent/internal/tool"
)

const defaultMaxIterations = 50

// ReActLoop executes the core ReAct (Reasoning + Acting) cycle.
type ReActLoop struct {
	llm          llm.Provider
	toolRegistry *tool.ToolRegistry
	eventBus     *EventBus
	maxIter      int
}

// NewReActLoop creates a new ReAct loop with the given components.
func NewReActLoop(llm llm.Provider, registry *tool.ToolRegistry, bus *EventBus, maxIter int) *ReActLoop {
	if maxIter <= 0 {
		maxIter = defaultMaxIterations
	}
	return &ReActLoop{
		llm:          llm,
		toolRegistry: registry,
		eventBus:     bus,
		maxIter:      maxIter,
	}
}

// RunOptions controls the behavior of a ReActLoop run.
type RunOptions struct {
	Task         string
	SystemPrompt string
	Sandbox      sandbox.Sandbox
	Workspace    string

	// ProgressChan receives AgentState updates on each step (non-blocking send).
	ProgressChan chan<- AgentState

	// PermissionChan is used for blocking permission prompts.
	PermissionChan chan<- PermissionRequest
	PermissionResp <-chan PermissionResponse
}

// PermissionRequest is sent when a tool requires user approval.
type PermissionRequest struct {
	ToolName string
	Reason   string
}

// PermissionResponse is the user's response to a permission request.
type PermissionResponse struct {
	Allowed     bool
	AlwaysAllow bool
}

// Run executes a full ReAct cycle for a single task.
func (r *ReActLoop) Run(opts RunOptions) AgentResult {
	messages := []llm.Message{llm.System(opts.SystemPrompt)}
	messages = append(messages, llm.User(opts.Task))

	var toolCallRecords []ToolCallRecord
	policyEngine := opts.Sandbox.PolicyEngine()

	for step := 1; step <= r.maxIter; step++ {
		r.emitProgressWithTool(opts, step, messages, "")

		// === THOUGHT: Call LLM ===
		response, err := r.llm.Chat(context.Background(), messages, r.toolRegistry.ToLLMFormats())
		if err != nil {
			messages = append(messages, llm.User(fmt.Sprintf("LLM error: %v. Please try to continue.", err)))
			continue
		}

		// No tool calls → text response or final answer
		if !response.HasToolCalls() {
			messages = append(messages, llm.Assistant(response.Content))

			if response.IsFinalAnswer {
				return AgentResult{
					Answer:    response.Content,
					Steps:     step,
					ToolCalls: toolCallRecords,
				}
			}
			continue
		}

		// === ACTION + OBSERVATION: Execute tool calls ===
		type indexedResult struct {
			index  int
			result string
		}

		results := make([]indexedResult, len(response.ToolCalls))

		var wg sync.WaitGroup
		for i, tc := range response.ToolCalls {
			// Emit progress with tool name before each tool execution
			r.emitProgressWithTool(opts, step, messages, tc.Name)
			wg.Add(1)
			go func(idx int, call llm.ToolCall) {
				defer wg.Done()
				observation := r.executeTool(opts, call, policyEngine, &toolCallRecords)
				results[idx] = indexedResult{index: idx, result: observation}
			}(i, tc)
		}
		wg.Wait()

		// Append tool calls and results in original order
		for i := range response.ToolCalls {
			tc := response.ToolCalls[i]
			messages = append(messages, llm.AssistantWithToolCalls(llm.CloneToolCall(tc)))
			for _, r := range results {
				if r.index == i {
					messages = append(messages, llm.ToolResult(tc.ID, r.result))
					break
				}
			}
		}
	}

	// Max iterations reached
	messages = append(messages, llm.User("You have reached the maximum number of steps. Please provide your final answer now."))
	finalResp, err := r.llm.Chat(context.Background(), messages, nil)
	if err != nil {
		return AgentResult{
			Answer:    "Error: Failed to get final answer after max iterations: " + err.Error(),
			Steps:     r.maxIter,
			ToolCalls: toolCallRecords,
			Error:     err,
		}
	}

	return AgentResult{
		Answer:    finalResp.Content,
		Steps:     r.maxIter,
		ToolCalls: toolCallRecords,
	}
}

func (r *ReActLoop) executeTool(opts RunOptions, call llm.ToolCall, pe *sandbox.PolicyEngine, records *[]ToolCallRecord) string {
	if !r.toolRegistry.Has(call.Name) {
		available := r.toolNames()
		*records = append(*records, ToolCallRecord{Name: call.Name, Args: call.Arguments, Success: false})
		return fmt.Sprintf("Error: Tool '%s' not found. Available: %s", call.Name, strings.Join(available, ", "))
	}

	t := r.toolRegistry.Resolve(call.Name)
	decision := pe.Check(call.Name, t.RequiredPermission())

	switch decision.Verdict {
	case sandbox.VerdictDeny:
		*records = append(*records, ToolCallRecord{Name: call.Name, Args: call.Arguments, Success: false})
		return "Error: Tool '" + call.Name + "' was denied by security policy: " + decision.Reason

	case sandbox.VerdictPrompt:
		if opts.PermissionChan != nil && opts.PermissionResp != nil {
			opts.PermissionChan <- PermissionRequest{ToolName: call.Name, Reason: decision.Reason}
			resp := <-opts.PermissionResp
			if !resp.Allowed {
				*records = append(*records, ToolCallRecord{Name: call.Name, Args: call.Arguments, Success: false})
				return "Error: Tool '" + call.Name + "' requires permission escalation: " + decision.Reason
			}
			if resp.AlwaysAllow {
				pe.AddOverride(call.Name, true)
			}
		} else {
			*records = append(*records, ToolCallRecord{Name: call.Name, Args: call.Arguments, Success: false})
			return "Error: Tool '" + call.Name + "' requires permission escalation: " + decision.Reason +
				". Use a different approach or ask the user to escalate permissions."
		}

	case sandbox.VerdictAllow:
	}

	ctx := tool.ToolContext{Workspace: opts.Workspace, Sandbox: opts.Sandbox}
	result := t.Execute(ctx, call.Arguments)
	*records = append(*records, ToolCallRecord{Name: call.Name, Args: call.Arguments, Success: result.Success})

	if result.Success {
		return result.FormatForLLM()
	}
	return "Error executing " + call.Name + ": " + result.Error
}

func (r *ReActLoop) emitProgressWithTool(opts RunOptions, step int, messages []llm.Message, toolName string) {
	if opts.ProgressChan == nil {
		return
	}
	state := AgentState{
		Task:               opts.Task,
		Step:               step,
		ConversationTokens: r.llm.CountMessagesTokens(messages),
		CurrentTool:        toolName,
		PermissionLevel:    opts.Sandbox.PolicyEngine().CurrentLevel().String(),
	}
	select {
	case opts.ProgressChan <- state:
	default:
	}
}

func (r *ReActLoop) toolNames() []string {
	formats := r.toolRegistry.ToLLMFormats()
	names := make([]string, len(formats))
	for i, f := range formats {
		if n, ok := f["name"].(string); ok {
			names[i] = n
		}
	}
	return names
}
