package agent

import (
	"context"

	"github.com/codingagent/coding-agent/internal/config"
	"github.com/codingagent/coding-agent/internal/llm"
	"github.com/codingagent/coding-agent/internal/sandbox"
	"github.com/codingagent/coding-agent/internal/tool"
	"github.com/codingagent/coding-agent/internal/tool/builtin"
)

// Orchestrator wires together all five layers and manages agent lifecycle.
type Orchestrator struct {
	config       *config.Config
	llm          llm.Provider
	toolRegistry *tool.ToolRegistry
	sandbox      *sandbox.SubprocessSandbox
	reactLoop    *ReActLoop
	eventBus     *EventBus
	workspace    string
}

// NewOrchestrator creates a fully wired orchestrator from configuration.
func NewOrchestrator(cfg *config.Config) (*Orchestrator, error) {
	// Create LLM provider
	factory := llm.NewProviderFactory(
		cfg.Model.Model, cfg.Model.Provider, cfg.Model.APIKey, cfg.Model.APIKeyEnv,
		cfg.Model.BaseURL, cfg.Model.MaxTokens, cfg.Model.ContextWindow, cfg.Model.Temperature,
	)
	provider, err := factory.Create()
	if err != nil {
		return nil, err
	}

	// Create tool registry and register built-in tools
	registry := tool.NewToolRegistry()
	registerBuiltinTools(registry)

	// Determine workspace
	workspace := cfg.Workspace
	if workspace == "" {
		workspace = "."
	}

	// Create sandbox
	permLevel := sandbox.FromString(cfg.Permissions.DefaultLevel)
	sb := sandbox.NewSubprocessSandbox(
		workspace,
		cfg.Sandbox.TimeoutSeconds,
		cfg.Sandbox.BlockedCommands,
		permLevel,
	)

	// Create event bus
	bus := NewEventBus()

	// Create ReAct loop
	loop := NewReActLoop(provider, registry, bus, cfg.MaxIter)

	return &Orchestrator{
		config:       cfg,
		llm:          provider,
		toolRegistry: registry,
		sandbox:      sb,
		reactLoop:    loop,
		eventBus:     bus,
		workspace:    workspace,
	}, nil
}

// registerBuiltinTools registers all six built-in tools.
func registerBuiltinTools(registry *tool.ToolRegistry) {
	registry.RegisterAll([]tool.Tool{
		&builtin.FileReadTool{},
		&builtin.FileWriteTool{},
		&builtin.FileEditTool{},
		&builtin.FileSearchTool{},
		&builtin.BashTool{},
		builtin.NewWebFetchTool(),
	})
}

// Run executes a task with the default options (no UI channels).
func (o *Orchestrator) Run(task string) AgentResult {
	return o.reactLoop.Run(RunOptions{
		Task:         task,
		SystemPrompt: buildSystemPrompt(),
		Sandbox:      o.sandbox,
		Workspace:    o.workspace,
	})
}

// RunWithChannels executes a task with progress and permission channels for TUI integration.
func (o *Orchestrator) RunWithChannels(ctx context.Context, task string, progressChan chan<- AgentState, permChan chan<- PermissionRequest, permResp <-chan PermissionResponse) AgentResult {
	return o.reactLoop.Run(RunOptions{
		Task:           task,
		SystemPrompt:   buildSystemPrompt(),
		Sandbox:        o.sandbox,
		Workspace:      o.workspace,
		ProgressChan:   progressChan,
		PermissionChan: permChan,
		PermissionResp: permResp,
		Ctx:            ctx,
	})
}

// EventBus returns the orchestrator's event bus.
func (o *Orchestrator) EventBus() *EventBus { return o.eventBus }

// LLM returns the LLM provider.
func (o *Orchestrator) LLM() llm.Provider { return o.llm }

// Config returns the current configuration.
func (o *Orchestrator) Config() *config.Config { return o.config }

// buildSystemPrompt constructs the system prompt describing available tools and expected behavior.
func buildSystemPrompt() string {
	return `You are an AI coding assistant. You help users with programming tasks using a ReAct architecture.
You have access to tools for reading, writing, editing, searching files, executing bash commands,
and fetching web content.

When you complete a task, provide a clear summary of what you did.
If you encounter errors, explain them and suggest alternatives.
Use the tools available to you to understand the codebase before making changes.`
}
