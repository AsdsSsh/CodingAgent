## Context

CodingAgent 是一个轻量级 AI 编程助手，采用 ReAct (Reasoning + Acting) 架构。当前 Java 实现共 37 个源文件、3021 行代码，包含 5 层架构（CLI → Config → Orchestrator → ReAct Loop → Foundation）、6 个内置工具、2 个 LLM 提供商（Anthropic + OpenAI 兼容）。

目标是用 Go 1:1 复刻全部功能，在此基础上将 REPL 升级为全屏 TUI、集成 MCP 协议、并利用 goroutine 实现并发工具执行。分发形态从 `java -jar` 变为单二进制文件。

## Goals / Non-Goals

**Goals:**
- 1:1 功能复刻 Java 版所有核心能力（ReAct 循环、6 个内置工具、权限策略引擎、分层配置）
- 全屏 TUI REPL，支持实时进度、权限弹窗阻断、对话历史滚动
- MCP 协议集成（stdio + HTTP transport），与内置工具共用 ToolRegistry
- 单二进制构建（`go build`），跨平台（Windows/Linux/macOS）
- 关键路径单元测试（ReAct 循环、LLM JSON 解析、PolicyEngine）

**Non-Goals:**
- 功能增强（如记忆系统、上下文压缩、Docker 沙箱——这些是后续 phase 的事）
- 向后兼容 Java 版（代码结构、类名、配置文件格式不变）
- gRPC/HTTP API 服务器（保持 CLI 工具定位）
- 流式 LLM 响应（v1 用非流式，和 Java 版一致）

## Decisions

### 1. 项目布局：`cmd/` + `internal/` 标准结构

```
coding-agent/
├── cmd/coding-agent/main.go     # 入口
├── internal/
│   ├── agent/                   # ReActLoop, Orchestrator, EventBus
│   ├── cli/                     # REPL (bubbletea), Formatter, Markdown
│   ├── config/                  # YAML 加载 + 分层合并
│   ├── llm/                     # Provider 接口 + Anthropic/OpenAI 实现
│   ├── sandbox/                 # SubprocessSandbox, PolicyEngine
│   ├── tool/                    # Tool 接口 + ToolRegistry
│   │   └── builtin/             # 6 个内置工具
│   └── mcp/                     # MCP SDK 封装
├── config/                      # YAML 配置文件（与 Java 版共享格式）
├── go.mod / go.sum
├── Makefile
└── README.md
```

**选择原因**：Go 社区标准做法。`internal/` 防止外部 import，`cmd/` 单一入口。与 Java 的 `com.codingagent.*` 包结构有清晰的 1:1 映射关系。

**替代方案**：`pkg/` 暴露公共 API——但当前是 CLI 工具而非库，不需要。

### 2. 配置：可变 Struct + 显式流水线，放弃不可变性

Java 版用 `record` + `with*()` 实现不可变配置。Go 中无语法等价物。

```go
type Config struct {
    Model       ModelConfig
    MCPServers  []MCPServerConfig
    Permissions PermissionConfig
    Workspace   string
    MaxIter     int
    MemoryPath  string
    Sandbox     SandboxConfig
}

// 分层加载流水线
func Load() (*Config, error) {
    cfg := Defaults()           // 硬编码默认值
    cfg.Merge(loadGlobal())     // ~/.coding-agent/config.yaml
    cfg.Merge(loadProject())    // .coding-agent/config.yaml
    cfg.ApplyEnv()              // ANTHROPIC_API_KEY 等
    return cfg, nil
}
```

**选择原因**：对于一个 sprint 的项目，可变 struct 简单直观。分层合并的流水线顺序明确，易于调试和测试。

**替代方案**：手动 builder + copy-on-write——但对于配置这种"加载一次、多次读取"的场景，不可变性的收益不足以抵消复杂度。

### 3. bubbletea 全屏 TUI 架构

```
┌─ Header ───────────────────────────────────┐
│  CodingAgent │ model │ permission │ tokens  │
├─────────────────────────────────────────────┤
│                                             │
│  ⏳ Step 3 · read · examining App.java     │  ← 状态行
│  ─────────────────────────────────         │
│  对话输出...                                │  ← Viewport
│                                             │
├─────────────────────────────────────────────┤
│  > 用户输入...                              │  ← Textarea
├─────────────────────────────────────────────┤
│  Ctrl+Enter │ Ctrl+C │ Tab │ /commands     │  ← 帮助栏
└─────────────────────────────────────────────┘
```

关键设计决策：

- **bubbletea 与 ReActLoop 的桥接**：ReActLoop 在独立 goroutine 中运行，通过 buffered channel 向 Model 发送 AgentProgressMsg。Update 函数处理消息后返回一个新的 listenProgress 命令以持续监听，形成自循环。
- **权限弹窗用 Modal 阻断**：PROMPT 时 ReActLoop 向 permission channel 发送请求并**阻塞等待**——Model 渲染 Modal → 用户选择 → 响应发回 channel → ReActLoop 继续。这替代了 Java 版的"返回错误 observation 让 Agent 自行绕过"。
- **焦点管理用 enum**：`FocusInput` 和 `FocusViewport` 两个状态，Tab 切换。键盘消息根据焦点路由到不同子组件。
- **组件选择**：`textarea.Model`（输入）、`viewport.Model`（可滚动输出）、`spinner.Model`（加载动画）。

**替代方案**：go-readline + lipgloss（Level 2）——功能完整但无法做到输入/输出区分离和实时进度。用户明确要求一步到位做全屏 TUI。

### 4. Anthropic Messages API：两遍 JSON 解析

Anthropic API 的 content block 是异构数组（text + tool_use 混在同一 `content` 数组中），Go 的静态类型无法直接反序列化。

```go
type ContentBlock struct {
    Type string `json:"type"`
}

type TextBlock struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

type ToolUseBlock struct {
    Type  string         `json:"type"`
    ID    string         `json:"id"`
    Name  string         `json:"name"`
    Input map[string]any `json:"input"`
}

// 解析流程：
// 1. json.Unmarshal → []ContentBlock (只解析 Type)
// 2. 遍历，按 Type 分发：
//    - "text"      → json.Unmarshal → TextBlock
//    - "tool_use"  → json.Unmarshal → ToolUseBlock
```

**选择原因**：这是处理异构 JSON 数组的标准 Go 模式（"先看类型，再完整解析"）。避免了 `map[string]any` 地狱，保持类型安全。

**替代方案**：`json.RawMessage` + 延迟解析——更灵活但调试困难。用明确的结构体更好。

### 5. 工具执行并发化

Java 版在一个 step 中有多个 tool_call 时**串行执行**。Go 版改为：互不依赖的 tool_call 通过 goroutine 并发执行。

```go
var wg sync.WaitGroup
results := make([]ToolResult, len(toolCalls))

for i, tc := range toolCalls {
    if tc.canRunConcurrently(previousResults) {
        wg.Add(1)
        go func(idx int, call ToolCall) {
            defer wg.Done()
            results[idx] = tool.Execute(ctx, call.Arguments)
        }(i, tc)
    } else {
        results[i] = tool.Execute(ctx, tc.Arguments)
    }
}
wg.Wait()
```

**选择原因**：Go 的 goroutine 让并发工具执行零成本。当 LLM 返回多个独立的文件读取或搜索请求时，并发能将等待时间从 sum(t1..tn) 降到 max(t1..tn)。

**注意**：需要保留顺序信息——`results` 用 index 对齐，追加到消息列表时保持原始顺序。

### 6. 依赖选择

| 用途 | Go 库 | 选择原因 |
|------|-------|---------|
| YAML | `gopkg.in/yaml.v3` | Go 生态事实标准 |
| CLI | `github.com/spf13/cobra` | 最成熟的 Go CLI 框架 |
| TUI | `github.com/charmbracelet/bubbletea` | Elm 架构，组件生态完善 |
| 样式 | `github.com/charmbracelet/lipgloss` | 声明式 ANSI，和 bubbletea 配合 |
| MCP | `github.com/modelcontextprotocol/go-sdk` | 官方维护，stdio+HTTP transport |
| 测试 | `github.com/stretchr/testify` | assert/mock 实用工具 |
| HTTP | stdlib `net/http` | LLM API 调用，无需第三方 HTTP 库 |
| 子进程 | stdlib `os/exec` | 沙箱命令执行 |

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|---------|
| Anthropic API JSON 解析出错 | 关键路径单元测试覆盖 JSON 往返（真实 API 响应 fixture）；边界：空 content、多 tool_use、reasoning_content |
| bubbletea 自循环消息可能泄漏 goroutine | progress channel 用 `close()` 信号结束；AgentResultMsg 到达后停止返回 listenProgress 命令；Ctrl+C 走 context 取消 |
| 权限弹窗阻塞导致 ReActLoop 永久挂起 | 设置弹窗超时（默认 120s），超时后自动 DENY；用户关闭弹窗等同于 DENY |
| Windows 终端兼容（ConPTY 老版本） | bubbletea v1 已处理大部分 Windows 兼容；init() 中显式 `ENABLE_VIRTUAL_TERMINAL_PROCESSING`；CI 矩阵测试 Windows Terminal + cmd.exe |
| Sprint 范围内 TUI 代码量可能超预期 | TUI 预估 ~780 行；若 Week 1 发现进度落后，先把 TUI 降级到 go-readline（接口不变，换实现） |
| MCP Go SDK API 不稳定 | 封装 `internal/mcp` 层，对上层只暴露 `ToolBase` 接口；SDK 变更只影响封装层 |

## Open Questions

1. **流式 LLM 响应**：v1 保持非流式（和 Java 一致），但 bubbletea 的 Model/Update 架构天然支持 SSE 流。是否在 v1.1 加上？
2. **记忆系统**：Java 版预留了 `memoryPath` 配置和 `memories` 参数，但未实现。Go 版同样先留钩子还是直接做？
3. **配置文件格式兼容**：Go 版的 YAML 配置是否 100% 兼容 Java 版？当前设计是"兼容"，但字段名可能需要调整（如 `snake_case` vs `camelCase` 的 yaml tag）。
