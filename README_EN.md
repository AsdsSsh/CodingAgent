# CodingAgent

**A ReAct-architecture AI coding assistant in a single binary, zero runtime dependencies. Built with Go.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8)
![Bubbletea](https://img.shields.io/badge/TUI-Bubbletea-FF4E8E)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)
![Binary](https://img.shields.io/badge/binary-11.4MB-blue)

CodingAgent is not a general-purpose chat shell. It focuses on the high-frequency loop of coding tasks: natural-language instruction → ReAct loop of reasoning and tool execution → inline approval for dangerous operations → delivered result.

```text
Natural-language task
  -> ReAct loop (reasoning + parallel tool execution, max 50 iterations)
  -> read / write / edit / search / bash / web fetch
  -> inline permission prompt for high-risk operations
  -> final answer delivered
```

Core design stance: **the model understands, plans and generates; CodingAgent owns tool boundaries, permissions, approvals and execution.** Model output never bypasses the permission system.

## Features

### Coding Loop

- **ReAct loop** — alternating LLM reasoning and tool execution; tool calls run in parallel; hitting the iteration cap converges to a final answer instead of failing outright
- **Six built-in tools** — file read (line ranges + 100K char cap), write (path-escape guard), precise edit (string replacement, ambiguity fails fast), dual-mode search (glob patterns + regex grep, binary files skipped automatically), bash execution, web fetch (tag stripping + 10-redirect cap)
- **Full-screen TUI** — bubbletea REPL: welcome screen, Markdown rendering, color-coded permission badge, live status bar (steps / tools / token estimate), ✓/✗ tool result prefixes, auto-scrolling
- **Cancellable tasks** — Ctrl+C aborts in-flight LLM requests instantly; LLM calls and bash execution carry context timeouts throughout — nothing hangs indefinitely

### Permissions & Security

- **Five permission levels** — `NONE / READ / WRITE / EXECUTE / DANGEROUS`, higher levels subsume lower ones; cycle with `Ctrl+P` in the TUI, the badge changes color live (gray / yellow / orange / red)
- **Three-stage check** — user permanent overrides (remembered via `A` in the prompt) → current-level verdict → inline confirmation when the level is insufficient (`Y` allow once / `N` deny / `A` always allow)
- **Frequency throttling** — 10 calls / 300 seconds per tool by default; exceeding the budget re-prompts even with sufficient permissions, guarding against runaway loops
- **Command blocklist** — high-risk commands like `rm -rf /`, `sudo`, `mkfs` are hard-coded blocked; timeouts kill the entire process tree

### Model Access

- **9 provider presets** — Anthropic, OpenAI, DeepSeek, Ollama, Gemini, Moonshot, Qwen, Zhipu, vLLM; endpoints and protocols auto-detected from the model name
- **Zero SDK dependencies** — Anthropic and OpenAI-compatible protocols are hand-written on `net/http`; the whole project has just 6 direct dependencies
- **Thinking mode** — DeepSeek ReasoningContent is passed through and displayed
- **MCP client** — Model Context Protocol module with stdio + HTTP transports implemented (lazy connection + tool adapter); runtime integration in progress

## Architecture

```text
cmd/coding-agent        Cobra entry point + Windows VT terminal setup
        |
        v
internal/cli            bubbletea TUI (model / update / view)
  - slash commands, key handling, permission prompts, Markdown rendering
  - 4 channel groups connecting the TUI with the agent goroutine
        |
        v
internal/agent          Orchestrator assembling five layers
  - ReActLoop (max 50 iterations) + EventBus (non-blocking pub/sub)
        |
   +----+----+----------------+
   |         |                |
   v         v                v
internal/llm          internal/tool          internal/sandbox
Provider interface    6 built-in tools       PolicyEngine
Anthropic + OpenAI    thread-safe registry   5 permission levels + throttling
hand-written clients  workspace path guard   subprocess sandbox + process-tree kill
        |
        v
internal/config        layered config merging
internal/mcp           MCP client (stdio / HTTP)
```

Config precedence: `hard-coded defaults → ~/.coding-agent/config.yaml (global) → .coding-agent/config.yaml (project) → environment variables → CLI flags`, with non-zero fields overriding at each layer.

## Tech Stack

| Layer | Technology |
| --- | --- |
| Language | Go 1.25 (no cgo, cross-compilation friendly) |
| TUI | bubbletea 1.3, bubbles, lipgloss |
| CLI | Cobra 1.10 |
| LLM | Anthropic protocol + OpenAI-compatible protocol (hand-written on `net/http`, no SDK) |
| Config | Layered YAML merging (gopkg.in/yaml.v3) |
| Testing | go test (6 test files, runs offline, no API key needed) |

## Quick Start

### Prerequisites

- Go 1.25+
- An API key for any supported provider (Ollama / vLLM local setups need none)

### 1. Build

```powershell
git clone https://github.com/AsdsSsh/CodingAgent.git
cd CodingAgent
make build
./coding-agent          # Windows: .\coding-agent.exe
```

### 2. Configure your key

Environment variables (recommended):

```powershell
$env:ANTHROPIC_API_KEY = "sk-ant-..."   # Claude
$env:OPENAI_API_KEY    = "sk-..."       # GPT
$env:DEEPSEEK_API_KEY  = "sk-..."       # DeepSeek
$env:DASHSCOPE_API_KEY = "sk-..."       # Qwen
```

Or via config file (`~/.coding-agent/config.yaml` global, or `.coding-agent/config.yaml` per project):

```yaml
model:
  provider: anthropic
  model: claude-sonnet-4-20250514
  max_tokens: 8192
  context_window: 200000

permissions:
  default_level: READ          # NONE / READ / WRITE / EXECUTE / DANGEROUS
  max_frequency: 10            # per-tool frequency cap
  frequency_window_seconds: 300
```

### 3. Launch

```powershell
coding-agent                              # start TUI with defaults
coding-agent -m gpt-4o                    # specific model (provider auto-detected)
coding-agent -m qwen2.5:7b -u http://localhost:11434/v1   # local Ollama
coding-agent -w C:\path\to\project        # specify workspace
coding-agent -m glm-4 -k <your-key>       # pass key explicitly
```

### 4. TUI Reference

**Slash commands:**

| Command | Alias | Description |
| --- | --- | --- |
| `/help` | `/h` | Show available commands |
| `/model [name]` | `/m` | View or switch model |
| `/permission [level]` | `/perm` `/p` | View or switch permission level (no arg = cycle) |
| `/clear` | | Clear conversation history |
| `/exit` | `/quit` `/q` | Exit |

**Keyboard shortcuts:**

| Key | Action |
| --- | --- |
| `Enter` | Submit single-line task |
| `Ctrl+S` | Submit multi-line task |
| `Ctrl+P` | Cycle permission level (READ → WRITE → EXECUTE → DANGEROUS) |
| `Ctrl+C` | Cancel running task / quit when idle |
| `Tab` | Toggle focus (input ↔ viewport) |

While a permission prompt is up: `Y`/`Enter` allow once, `N`/`Esc` deny, `A` always allow this tool.

## Model Presets

| Provider | Model Prefix | Default Model | Context Window | Env Variable |
| --- | --- | --- | --- | --- |
| Anthropic | `claude-` | claude-sonnet-4-20250514 | 200K | `ANTHROPIC_API_KEY` |
| OpenAI | `gpt-` `o1` `o3` `o4` | gpt-4o | 128K | `OPENAI_API_KEY` |
| DeepSeek | `deepseek-` | deepseek-v4-pro | 1M | `DEEPSEEK_API_KEY` |
| Ollama | (any) | qwen2.5:7b | 32K | — |
| Gemini | `gemini-` | gemini-2.0-flash | 1M | `GEMINI_API_KEY` |
| Moonshot | `moonshot-` `kimi-` | moonshot-v1-8k | 128K | `MOONSHOT_API_KEY` |
| Qwen | `qwen-` | qwen-max | 131K | `DASHSCOPE_API_KEY` |
| Zhipu | `glm-` | glm-4 | 128K | `ZHIPU_API_KEY` |
| vLLM | (any) | (custom) | 32K | — |

Key resolution order: explicit `-k` → preset env variable → `OPENAI_API_KEY` → `ANTHROPIC_API_KEY`.

## Verification & Build

```powershell
make build          # build for current platform (.exe suffix on Windows)
make build-release  # stripped build (smaller)
make build-all      # cross-compile Windows / Linux / macOS (amd64 + arm64)
make test           # all tests (offline, no API key required)
make test-cover     # coverage + HTML report
make lint           # golangci-lint (install separately)
```

## Project Structure

```text
cmd/coding-agent/     Entry: Cobra CLI + Windows VT handling
internal/
  agent/              ReActLoop, Orchestrator, EventBus
  cli/                TUI (bubbletea): model, update, view, Markdown
  config/             Layered YAML loading and merging
  llm/                Provider interface + Anthropic/OpenAI protocols + 9 presets
  mcp/                MCP client (stdio / HTTP, lazy connection)
  sandbox/            Permission levels, PolicyEngine, subprocess sandbox, process-tree kill
  tool/               Tool interface, registry, workspace guard
    builtin/          read / write / edit / search / bash / web_fetch
config/               Default YAML config template
openspec/             Development workflow specs and change archive
```

## License

[MIT](./LICENSE) © 2026 AsdsSsh

---

[中文版本](./README.md)
