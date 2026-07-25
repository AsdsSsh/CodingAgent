package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/codingagent/coding-agent/internal/llm"
	"github.com/codingagent/coding-agent/internal/sandbox"
	"github.com/codingagent/coding-agent/internal/tool"
)

const defaultMaxIterations = 50

type ReActLoop struct {
	llm          llm.Provider
	toolRegistry *tool.ToolRegistry
	eventBus     *EventBus
	maxIter      int
}

func NewReActLoop(llm llm.Provider, registry *tool.ToolRegistry, bus *EventBus, maxIter int) *ReActLoop {
	if maxIter <= 0 {
		maxIter = defaultMaxIterations
	}
	return &ReActLoop{llm: llm, toolRegistry: registry, eventBus: bus, maxIter: maxIter}
}

type RunOptions struct {
	Task           string
	SystemPrompt   string
	Sandbox        sandbox.Sandbox
	Workspace      string
	ProgressChan   chan<- AgentState
	PermissionChan chan<- PermissionRequest
	PermissionResp <-chan PermissionResponse
	Ctx            context.Context // cancellable context for the entire run
}

type PermissionRequest struct {
	ToolName string
	Reason   string
}

type PermissionResponse struct {
	Allowed     bool
	AlwaysAllow bool
}

func (r *ReActLoop) Run(opts RunOptions) AgentResult {
	messages := []llm.Message{llm.System(opts.SystemPrompt)}
	messages = append(messages, llm.User(opts.Task))

	baseCtx := opts.Ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	var toolCallRecords []ToolCallRecord
	policyEngine := opts.Sandbox.PolicyEngine()

	for step := 1; step <= r.maxIter; step++ {
		r.emitProgressWithTool(opts, step, messages, "")

		reqCtx, cancel := context.WithTimeout(baseCtx, 30*time.Second)
		response, err := r.llm.Chat(reqCtx, messages, r.toolRegistry.ToLLMFormats())
		cancel()

		if err != nil {
			if baseCtx.Err() != nil {
				return AgentResult{
					Answer:    "Task cancelled by user.",
					Steps:     step,
					ToolCalls: toolCallRecords,
				}
			}
			messages = append(messages, llm.User(fmt.Sprintf("LLM error: %v. Please try to continue.", err)))
			continue
		}

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

		type indexedResult struct {
			index  int
			result string
		}

		results := make([]indexedResult, len(response.ToolCalls))

		var wg sync.WaitGroup
		for i, tc := range response.ToolCalls {
			r.emitProgressWithTool(opts, step, messages, tc.Name)
			wg.Add(1)
			go func(idx int, call llm.ToolCall) {
				defer wg.Done()
				observation := r.executeTool(opts, call, policyEngine, &toolCallRecords)
				results[idx] = indexedResult{index: idx, result: observation}
			}(i, tc)
		}
		wg.Wait()

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

	messages = append(messages, llm.User("You have reached the maximum number of steps. Please provide your final answer now."))
	reqCtx, cancel := context.WithTimeout(baseCtx, 30*time.Second)
	finalResp, err := r.llm.Chat(reqCtx, messages, nil)
	cancel()

	if err != nil {
		return AgentResult{
			Answer:    "Error: " + err.Error(),
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
		return "Error: Tool '" + call.Name + "' was denied: " + decision.Reason

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
			return "Error: Tool '" + call.Name + "' requires permission escalation: " + decision.Reason
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
