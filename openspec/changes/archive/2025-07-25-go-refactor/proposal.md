## Why

将 CodingAgent 从 Java 21 + Maven 重构为 Go，实现零依赖单二进制分发、更低运行时开销、以及更自然的并发模型。当前 Java 实现已完成架构验证（3021 行，5 层 ReAct 架构），但分发要求用户安装 JDK 21 + Maven，启动慢，且 JLine REPL 在跨平台终端体验上存在局限。Go 的原生编译、goroutine 并发、以及 bubbletea 全屏 TUI 生态恰好弥补这些短板。同时这也是深入学习 Go 的实践机会。

## What Changes

- **语言切换**：全部 37 个 Java 源文件 → ~30 个 Go 源文件，保持 1:1 功能复刻
- **编译模型**：Maven 多插件构建 → `go build` 单二进制（~15MB），无需 JRE/JDK
- **REPL 体验升级**：JLine 行编辑 → bubbletea 全屏 TUI，支持实时进度、权限弹窗阻断、对话历史滚动、命令补全
- **LLM API 调用**：Java `java.net.http.HttpClient` → Go `net/http`，手写 Anthropic Messages API + OpenAI Chat Completions API JSON 往返
- **CLI 框架**：Picocli → Cobra + stdlib flag
- **配置系统**：Jackson YAML + Java Record → `gopkg.in/yaml.v3` + Go struct，保持分层合并逻辑
- **MCP 集成**：从注释占位变为正式功能，使用 `github.com/modelcontextprotocol/go-sdk`
- **并发工具执行**：ReAct 循环中互不依赖的 tool_call 通过 goroutine 并发执行（Java 版严格串行）
- **测试**：从零测试覆盖到关键路径测试（ReAct 循环、LLM JSON 解析、PolicyEngine）

## Capabilities

### New Capabilities

- `go-core-engine`: Go 重构后的核心引擎——ReAct 循环、LLM 提供商抽象、工具注册表、沙箱与权限策略。功能与 Java 版等价。
- `tui-repl`: bubbletea 全屏终端 REPL——多行输入、实时进度展示、权限弹窗阻断、对话历史滚动、命令补全与快捷键。
- `mcp-integration`: MCP（模型上下文协议）集成——支持 stdio 和 HTTP 传输，工具延迟加载，与内置工具共用注册表。
- `single-binary-build`: 单二进制构建与分发——`go build` 产出独立可执行文件，跨平台编译（Windows/Linux/macOS），零运行时依赖。

### Modified Capabilities

<!-- 无现有 spec 需要修改——这是全新重构 -->

## Impact

- **新增依赖（Go module）**：`gopkg.in/yaml.v3`、`github.com/spf13/cobra`、`github.com/charmbracelet/bubbletea`、`github.com/charmbracelet/lipgloss`、`github.com/modelcontextprotocol/go-sdk`、`github.com/stretchr/testify`
- **移除依赖（Java）**：Picocli、JLine、Jackson、anthropic-java SDK、openai-java SDK、Maven 构建链
- **项目结构**：新增 `cmd/`、`internal/`、`go.mod`、`go.sum`、`Makefile`；保留 `config/`（YAML 配置文件跨语言通用）；归档 Java 源码
- **用户体验**：从"装 JDK → mvn package → java -jar"变为"下载二进制 → 直接运行"；REPL 从行编辑升级为全屏 TUI
- **API 兼容**：LLM API 调用逻辑不变（Anthropic Messages API + OpenAI Chat Completions），HTTP 层从 Java HttpClient 换为 Go net/http
