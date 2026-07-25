## ADDED Requirements

### Requirement: ReAct Loop executes Thought-Action-Observation cycle
The system SHALL implement a ReAct (Reasoning + Acting) loop that alternates between LLM reasoning and tool execution until the task is complete or the step limit is reached.

#### Scenario: Task completes with final answer
- **WHEN** the LLM returns a response with `stop_reason: "end_turn"` (Anthropic) or `finish_reason: "stop"` (OpenAI) and no tool calls
- **THEN** the ReActLoop SHALL return an AgentResult containing the answer text, step count, and tool call records

#### Scenario: Task reaches max iterations
- **WHEN** the step count reaches `maxIterations` (default 50)
- **THEN** the ReActLoop SHALL append a message instructing the model to provide a final answer, make one final LLM call, and return the result

#### Scenario: LLM requests tool execution
- **WHEN** the LLM response contains one or more tool calls
- **THEN** the ReActLoop SHALL execute each tool, append the tool results as observations, and continue the loop

#### Scenario: Unknown tool is requested
- **WHEN** the LLM requests a tool not registered in the ToolRegistry
- **THEN** the ReActLoop SHALL append an error observation listing available tools, rather than crashing

#### Scenario: Permission denies a tool
- **WHEN** the PolicyEngine returns a DENY verdict for a tool call
- **THEN** the ReActLoop SHALL append an error observation with the denial reason and continue the loop

#### Scenario: Concurrent tool execution
- **WHEN** multiple tool calls in one step do not depend on each other's results
- **THEN** the ReActLoop SHALL execute them concurrently via goroutines, preserving the original call order in the message list

### Requirement: LLM Provider abstraction supports multiple backends
The system SHALL provide a common `Provider` interface that supports Anthropic Messages API and OpenAI-compatible Chat Completions API, selectable by configuration.

#### Scenario: Anthropic API call
- **WHEN** a model with provider `anthropic` is configured
- **THEN** the system SHALL send requests to `https://api.anthropic.com/v1/messages` with the `x-api-key` and `anthropic-version` headers, and extract system messages to the top-level `system` field

#### Scenario: OpenAI API call
- **WHEN** a model with provider `openai` or a custom `base_url` is configured
- **THEN** the system SHALL send requests to `{base_url}/chat/completions` with the `Authorization: Bearer` header

#### Scenario: Tool definition translation
- **WHEN** tools are provided to the chat method
- **THEN** AnthropicProvider SHALL translate them to `{name, description, input_schema}` format, and OpenAIProvider SHALL translate them to `{type: "function", function: {name, description, parameters}}` format

#### Scenario: Model auto-detection from preset
- **WHEN** a model name matches a built-in preset prefix (e.g., "claude-", "gpt-", "deepseek-")
- **THEN** the system SHALL automatically configure the correct base_url, API format, and key environment variable

#### Scenario: Unknown model with no base_url
- **WHEN** a model name does not match any preset and no base_url is configured
- **THEN** the system SHALL return an error listing available presets

### Requirement: Tool system with registry and six built-in tools
The system SHALL provide a `Tool` interface and a `ToolRegistry` that manages tools and exports their definitions in LLM-compatible format.

#### Scenario: Tool registration and lookup
- **WHEN** tools are registered via `Register(tool)`
- **THEN** the ToolRegistry SHALL resolve them by name via `Resolve(name)` and check existence via `Has(name)`

#### Scenario: LLM format export
- **WHEN** `ToLLMFormats()` is called
- **THEN** the ToolRegistry SHALL return a list of all registered tools formatted as `{name, description, input_schema}` maps

#### Scenario: File read with range selection
- **WHEN** the `read` tool is called with `file_path`, optional `offset`, and optional `limit`
- **THEN** the system SHALL read the file content, number each line, and return the output truncated at 100,000 characters

#### Scenario: File write with workspace boundary check
- **WHEN** the `write` tool is called with `file_path` and `content`
- **THEN** the system SHALL reject paths outside the workspace directory and create parent directories as needed

#### Scenario: File edit with unique match enforcement
- **WHEN** the `edit` tool is called with `file_path`, `old_string`, and `new_string`
- **THEN** the system SHALL fail if `old_string` appears zero or more than one time in the file, and replace it with `new_string` only on exact single match

#### Scenario: File search with glob and grep modes
- **WHEN** the `search` tool is called with `pattern` and `mode`
- **THEN** the system SHALL perform glob-based file name matching in "glob" mode or regex-based content matching in "grep" mode, with results capped at 200 matches

#### Scenario: Bash execution with sandbox enforcement
- **WHEN** the `bash` tool is called with `command`
- **THEN** the system SHALL delegate execution to the sandbox, which applies the command blacklist, timeout, and working directory constraints

#### Scenario: Web fetch with HTML stripping
- **WHEN** the `web_fetch` tool is called with a URL
- **THEN** the system SHALL fetch the page content, strip HTML tags, collapse whitespace, and return output truncated at 50,000 characters

### Requirement: Sandbox and Policy Engine enforce security constraints
The system SHALL provide a SubprocessSandbox that executes commands under timeout, command blacklist, and workspace path constraints, governed by a PolicyEngine with permission levels.

#### Scenario: Command blacklist blocks dangerous patterns
- **WHEN** a command contains blocked patterns (e.g., `rm -rf /`, `sudo`, `mkfs`)
- **THEN** the sandbox SHALL refuse execution and return a failure result

#### Scenario: Command timeout
- **WHEN** a command exceeds the configured timeout (default 120s)
- **THEN** the sandbox SHALL forcibly terminate the process and return a timeout error

#### Scenario: Permission level escalation
- **WHEN** the PolicyEngine checks a tool requiring EXECUTE against a READ-level session
- **THEN** the PolicyEngine SHALL return a PROMPT verdict requiring user escalation

#### Scenario: Frequency-based re-prompt
- **WHEN** a tool has been used more than `maxFrequency` times within `frequencyWindowSeconds`
- **THEN** the PolicyEngine SHALL return a PROMPT verdict to prevent silent overuse

#### Scenario: User permanent override
- **WHEN** the user has set a permanent ALLOW or DENY for a specific tool
- **THEN** the PolicyEngine SHALL return the override verdict without further checks
