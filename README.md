# CodingAgent — Go Edition

A lightweight AI coding assistant with ReAct (Reasoning + Acting) architecture, written in Go.

> Refactored from the original Java implementation. Single binary, zero runtime dependencies.

## Features

- **ReAct Loop**: Alternating LLM reasoning and tool execution
- **Full-Screen TUI**: Interactive REPL built with [bubbletea](https://github.com/charmbracelet/bubbletea)
- **Multi-Provider**: Anthropic, OpenAI, DeepSeek, Ollama, Gemini, Moonshot, Qwen, Zhipu, vLLM
- **Six Built-in Tools**: File read/write/edit, file search (glob + grep), bash execution, web fetch
- **Permission System**: Five-tier security levels (NONE/READ/WRITE/EXECUTE/DANGEROUS) with frequency throttling
- **MCP Support**: Model Context Protocol for extensible tools (stdio + HTTP transports)
- **Single Binary**: ~12MB executable, no JRE/JDK required

## Quick Start

### Download

```bash
# Build from source
git clone https://github.com/codingagent/coding-agent.git
cd coding-agent
make build
./coding-agent
```

### Prerequisites

- Go 1.22+
- API key for your preferred LLM provider

### Configuration

Set your API key:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."   # For Claude
export OPENAI_API_KEY="sk-..."          # For GPT-4o
export DEEPSEEK_API_KEY="sk-..."        # For DeepSeek
```

Or use a config file at `~/.coding-agent/config.yaml` or `.coding-agent/config.yaml`:

```yaml
model:
  provider: anthropic
  model: claude-sonnet-4-20250514
  max_tokens: 8192
  context_window: 200000

permissions:
  default_level: READ
```

### Usage

```bash
# Start interactive REPL
coding-agent

# Use a specific model
coding-agent -m gpt-4o

# Set workspace directory
coding-agent -w /path/to/project

# Custom API endpoint (local models)
coding-agent -m qwen2.5:7b -u http://localhost:11434/v1
```

### REPL Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/model [name]` | View or switch model |
| `/permission [level]` | View or switch permission level |
| `/clear` | Clear conversation history |
| `/exit` | Exit the REPL |

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Ctrl+Enter` | Submit task |
| `Ctrl+C` | Cancel task / Quit |
| `Tab` | Toggle focus (input ↔ viewport) |

## Supported Models

| Provider | Model Prefix | Default Model | Context Window |
|----------|-------------|---------------|----------------|
| Anthropic | `claude-` | claude-sonnet-4-20250514 | 200K |
| OpenAI | `gpt-`, `o1`, `o3`, `o4` | gpt-4o | 128K |
| DeepSeek | `deepseek-` | deepseek-v4-pro | 1M |
| Ollama | (any) | qwen2.5:7b | 32K |
| Gemini | `gemini-` | gemini-2.0-flash | 1M |
| Moonshot | `moonshot-`, `kimi-` | moonshot-v1-8k | 128K |
| Qwen | `qwen-` | qwen-max | 131K |
| Zhipu | `glm-` | glm-4 | 128K |
| vLLM | (any) | (custom) | 32K |

## Project Structure

```
cmd/coding-agent/    Entry point, CLI (Cobra)
internal/
  agent/             ReActLoop, Orchestrator, EventBus
  cli/               TUI REPL (bubbletea), Formatter, Markdown
  config/            YAML loading, layered merge
  llm/               Provider interface, Anthropic/OpenAI implementations
  mcp/               MCP client, lazy connection, tool adapter
  sandbox/           SubprocessSandbox, PolicyEngine, PermissionLevel
  tool/              Tool interface, ToolRegistry
    builtin/         FileRead, FileWrite, FileEdit, FileSearch, Bash, WebFetch
config/              Default YAML configuration
```

## Build

```bash
make build          # Build for current platform
make build-release  # Build with stripped symbols (smaller)
make build-all      # Cross-compile for Windows/Linux/macOS
make test           # Run all tests
make lint           # Run golangci-lint
```

## License

MIT
