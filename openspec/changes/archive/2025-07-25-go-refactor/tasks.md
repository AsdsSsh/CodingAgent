## 1. Project Scaffold

- [x] 1.1 Initialize Go module (`go mod init github.com/codingagent/coding-agent`) with Go 1.22+
- [x] 1.2 Create directory structure: `cmd/coding-agent/`, `internal/{agent,cli,config,llm,sandbox,tool/builtin,mcp}`
- [x] 1.3 Add initial dependencies: `gopkg.in/yaml.v3`, `github.com/spf13/cobra`, `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/stretchr/testify`
- [x] 1.4 Create stub `cmd/coding-agent/main.go` with Cobra root command
- [x] 1.5 Copy `config/default_config.yaml` from Java project (config format stays compatible)

## 2. LLM Provider Layer

- [x] 2.1 Define `Message` struct and `Role` enum (`internal/llm/message.go`)
- [x] 2.2 Define `LlmResponse` struct with `ToolCall` and `TokenUsage` (`internal/llm/response.go`)
- [x] 2.3 Define `Provider` interface: `Chat(messages, tools) → LlmResponse`, `CountTokens(text)`, `ModelName()`, `ContextWindow()` (`internal/llm/provider.go`)
- [x] 2.4 Implement `AnthropicProvider`: HTTP POST to Messages API, system prompt extraction, content-block format conversion, tool_use block parsing (`internal/llm/anthropic.go`)
- [x] 2.5 Implement `OpenAIProvider`: HTTP POST to Chat Completions API, tool_calls → function format conversion, reasoning_content passthrough for DeepSeek (`internal/llm/openai.go`)
- [x] 2.6 Implement `ModelRegistry` with all 9 built-in presets (Anthropic, OpenAI, DeepSeek, Ollama, Gemini, Moonshot, Qwen, Zhipu, vLLM) and prefix-based auto-detection (`internal/llm/registry.go`)
- [x] 2.7 Implement `ProviderFactory`: presets first → base_url override → anthropic provider type → error with suggestions (`internal/llm/factory.go`)

## 3. Tool System

- [x] 3.1 Define `Tool` interface and `ToolResult` struct (`internal/tool/tool.go`, `internal/tool/result.go`)
- [x] 3.2 Define `ToolContext` struct with workspace path resolution and boundary check (`internal/tool/context.go`)
- [x] 3.3 Implement `ToolRegistry` with register/resolve/has/toLLMFormats methods, thread-safe via `sync.RWMutex` (`internal/tool/registry.go`)
- [x] 3.4 Implement `FileReadTool`: file read with line numbering, offset/limit, 100K char truncation (`internal/tool/builtin/read.go`)
- [x] 3.5 Implement `FileWriteTool`: file create/overwrite with workspace boundary enforcement (`internal/tool/builtin/write.go`)
- [x] 3.6 Implement `FileEditTool`: exact string replacement with unique-match enforcement (`internal/tool/builtin/edit.go`)
- [x] 3.7 Implement `FileSearchTool`: glob-to-regex conversion, grep with line-by-line matching, 200-result cap (`internal/tool/builtin/search.go`)
- [x] 3.8 Implement `BashTool`: delegate to sandbox, require EXECUTE permission (`internal/tool/builtin/bash.go`)
- [x] 3.9 Implement `WebFetchTool`: HTTP GET, HTML tag stripping, 50K char truncation (`internal/tool/builtin/webfetch.go`)

## 4. Sandbox and Policy Engine

- [x] 4.1 Define `PermissionLevel` type with ordered constants (NONE=0 to DANGEROUS=4) and `Covers()` method (`internal/sandbox/level.go`)
- [x] 4.2 Define `PermissionDecision` struct with Verdict enum (ALLOW/DENY/PROMPT) (`internal/sandbox/decision.go`)
- [x] 4.3 Define `Sandbox` interface: `ExecuteCommand()`, `IsInWorkspace()`, `PolicyEngine()` (`internal/sandbox/sandbox.go`)
- [x] 4.4 Implement `PolicyEngine`: three-tier check (user overrides → level+frequency → escalation prompt), sliding window frequency tracking, thread-safe maps (`internal/sandbox/policy.go`)
- [x] 4.5 Implement `SubprocessSandbox`: cross-platform command execution (bash -c / cmd /c), timeout via context, command blacklist, stderr merging (`internal/sandbox/subprocess.go`)

## 5. Config System

- [x] 5.1 Define `Config` struct and all sub-configs (`ModelConfig`, `PermissionConfig`, `SandboxConfig`, `MCPServerConfig`) with YAML struct tags (`internal/config/config.go`)
- [x] 5.2 Implement `Defaults()` returning hardcoded sensible defaults
- [x] 5.3 Implement `Load()`: global config → project config → env var overrides, with deep merge for nested configs
- [x] 5.4 Implement CLI override methods (`WithModel()`, `WithWorkspace()`, `WithPermissionLevel()`) mutating the struct
- [x] 5.5 Implement `ResolveAPIKey()` with priority: explicit key → api_key_env → provider default env var

## 6. ReAct Loop and Orchestrator

- [x] 6.1 Define `AgentState`, `AgentResult`, and `ToolCallRecord` structs (`internal/agent/state.go`, `internal/agent/result.go`)
- [x] 6.2 Implement `EventBus`: typed subscribe/emit with `sync.RWMutex`, channel-based variant for bubbletea integration (`internal/agent/eventbus.go`)
- [x] 6.3 Implement `ReActLoop.Run()`: build initial messages → step loop → LLM call → parse tool calls → permission check → execute → append observations → max-iteration fallback (`internal/agent/react.go`)
- [x] 6.4 Integrate progress channel: send `AgentState` on each step for real-time TUI updates
- [x] 6.5 Integrate permission channel: block on PROMPT verdict, wait for user response from TUI
- [x] 6.6 Implement concurrent tool execution: when multiple tool_calls are independent, execute via goroutines with `sync.WaitGroup`
- [x] 6.7 Implement `Orchestrator`: wire Config → ProviderFactory → ToolRegistry → Sandbox → ReActLoop, register all 6 built-in tools, build system prompt (`internal/agent/orchestrator.go`)

## 7. TUI REPL

- [x] 7.1 Define all bubbletea message types: `TaskSubmittedMsg`, `AgentProgressMsg`, `AgentResultMsg`, `ToolCallStartedMsg`, `ToolCallFinishedMsg`, `PermissionPromptMsg`, `PermissionResponseMsg` (`internal/cli/messages.go`)
- [x] 7.2 Define `Model` struct with all state (input textarea, viewport, spinner, config, focus, channels) (`internal/cli/model.go`)
- [x] 7.3 Define lipgloss `Styles` for header, status bar, input area, help bar, modal overlay (`internal/cli/styles.go`)
- [x] 7.4 Implement `View()`: header → viewport → status → input → help bar, with modal overlay when permission prompt is active (`internal/cli/view.go`)
- [x] 7.5 Implement `Update()`: keyboard routing (Tab/Ctrl+C/Ctrl+Enter), agent message handling (progress/result/permission), spinner tick, sub-component updates (`internal/cli/update.go`)
- [x] 7.6 Implement `listenProgress()`: consume progress channel → return `AgentProgressMsg` → self-loop for continuous listening
- [x] 7.7 Implement permission modal rendering: tool name, reason, [Allow] [Deny] [Always Allow] buttons, 120s timeout
- [x] 7.8 Implement slash commands: `/help`, `/model [name]`, `/permission [level]`, `/clear`, `/history`, `/exit` (`internal/cli/commands.go`)
- [x] 7.9 Implement `Init()`: spinner start, input focus
- [x] 7.10 Implement `Formatter` and `MarkdownRenderer`: bold/italic/code/headings/lists → lipgloss ANSI styles (`internal/cli/formatter.go`, `internal/cli/markdown.go`)
- [x] 7.11 Implement `handleTaskSubmit()`: create orchestrator, start agent goroutine, wire progress/result channels, return tea.Batch of listenProgress + waitForCompletion + spinner

## 8. MCP Integration

- [x] 8.1 Add `github.com/modelcontextprotocol/go-sdk` dependency
- [x] 8.2 Implement `MCPServerConfig` loader from YAML: name, transport, command/args (stdio), url (http), env vars (`internal/config/mcp.go`)
- [x] 8.3 Implement MCP client wrapper: connect via stdio subprocess or HTTP transport, retrieve tool list, register into ToolRegistry (`internal/mcp/client.go`)
- [x] 8.4 Implement lazy connection: cache MCP tool metadata at startup, defer actual server connection to first tool invocation (`internal/mcp/lazy.go`)
- [x] 8.5 Implement MCP tool adapter: wrap MCP tool as `Tool` interface, route execute calls to MCP server, convert responses to `ToolResult`
- [x] 8.6 Handle MCP error cases: connection failure → error observation, server crash → reconnect attempt, duplicate tool name → skip with warning

## 9. Testing

- [x] 9.1 Test `AnthropicProvider` JSON parsing: fixture-based tests for text response, tool_use response, mixed content blocks, error responses
- [x] 9.2 Test `OpenAIProvider` JSON parsing: fixture-based tests for standard response, tool_calls response, DeepSeek reasoning_content
- [x] 9.3 Test `ModelRegistry` auto-detection: exact match, prefix match, unknown model fallback
- [x] 9.4 Test `ReActLoop`: mock LLM provider, verify step counting, tool execution, max-iteration fallback, permission denial handling
- [x] 9.5 Test `PolicyEngine`: level coverage checks, frequency window throttling, user override precedence
- [x] 9.6 Test `Config` loading: defaults, YAML file merge, env var override, CLI override
- [x] 9.7 Test `FileEditTool`: unique match, no match, multiple match
- [x] 9.8 Test `FileSearchTool`: glob-to-regex conversion, grep result formatting
- [x] 9.9 Test `SubprocessSandbox`: command execution, timeout, blacklist blocking

## 10. Build, Polish, and Documentation

- [x] 10.1 Create `Makefile`: build, test, lint, clean, build-all targets
- [x] 10.2 Add `//go:build` constraints for platform-specific code (Windows terminal init)
- [x] 10.3 Write `README.md`: installation, configuration, usage examples, model presets table
- [x] 10.4 Cross-platform smoke test: Windows Terminal, macOS Terminal, Linux (WSL/GNOME Terminal)
- [x] 10.5 Manual integration test: full ReAct loop with real LLM API, verify tool calls execute correctly
- [x] 10.6 Manual TUI test: multi-line input, real-time progress, permission modal, slash commands, Ctrl+C cancellation, resize behavior
- [x] 10.7 Edge case testing: empty tool response, very long output truncation, concurrent tool execution ordering, MCP server disconnect mid-task
