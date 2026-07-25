## 1. 项目骨架

- [x] 1.1 初始化 Go module (`go mod init github.com/codingagent/coding-agent`)，Go 1.22+
- [x] 1.2 创建目录结构：`cmd/coding-agent/`、`internal/{agent,cli,config,llm,sandbox,tool/builtin,mcp}`
- [x] 1.3 添加初始依赖：`gopkg.in/yaml.v3`、`github.com/spf13/cobra`、`github.com/charmbracelet/bubbletea`、`github.com/charmbracelet/lipgloss`、`github.com/stretchr/testify`
- [x] 1.4 创建 `cmd/coding-agent/main.go` 桩代码，集成 Cobra 根命令
- [x] 1.5 从 Java 项目复制 `config/default_config.yaml`（配置格式保持兼容）

## 2. LLM 提供商层

- [x] 2.1 定义 `Message` 结构体和 `Role` 枚举（`internal/llm/message.go`）
- [x] 2.2 定义 `LlmResponse` 结构体，含 `ToolCall` 和 `TokenUsage`（`internal/llm/response.go`）
- [x] 2.3 定义 `Provider` 接口：`Chat(messages, tools) → LlmResponse`、`CountTokens(text)`、`ModelName()`、`ContextWindow()`（`internal/llm/provider.go`）
- [x] 2.4 实现 `AnthropicProvider`：HTTP POST 调 Messages API、系统提示提取、content-block 格式转换、tool_use block 解析（`internal/llm/anthropic.go`）
- [x] 2.5 实现 `OpenAIProvider`：HTTP POST 调 Chat Completions API、tool_calls → function 格式转换、DeepSeek 的 reasoning_content 透传（`internal/llm/openai.go`）
- [x] 2.6 实现 `ModelRegistry`，含全部 9 个内置预设（Anthropic、OpenAI、DeepSeek、Ollama、Gemini、Moonshot、Qwen、Zhipu、vLLM）及前缀自动检测（`internal/llm/registry.go`）
- [x] 2.7 实现 `ProviderFactory`：预设优先 → base_url 覆盖 → anthropic 类型 → 错误提示（`internal/llm/factory.go`）

## 3. 工具系统

- [x] 3.1 定义 `Tool` 接口和 `ToolResult` 结构体（`internal/tool/tool.go`、`internal/tool/result.go`）
- [x] 3.2 定义 `ToolContext` 结构体，含工作区路径解析和边界检查（`internal/tool/context.go`）
- [x] 3.3 实现 `ToolRegistry`：register/resolve/has/toLLMFormats 方法，通过 `sync.RWMutex` 保证线程安全（`internal/tool/registry.go`）
- [x] 3.4 实现 `FileReadTool`：文件读取 + 行号、offset/limit、100K 字符截断（`internal/tool/builtin/read.go`）
- [x] 3.5 实现 `FileWriteTool`：文件创建/覆盖 + 工作区边界强制检查（`internal/tool/builtin/write.go`）
- [x] 3.6 实现 `FileEditTool`：精确字符串替换 + 唯一匹配强制检查（`internal/tool/builtin/edit.go`）
- [x] 3.7 实现 `FileSearchTool`：glob 转正则、grep 逐行匹配、200 结果上限（`internal/tool/builtin/search.go`）
- [x] 3.8 实现 `BashTool`：委托沙箱执行，需要 EXECUTE 权限（`internal/tool/builtin/bash.go`）
- [x] 3.9 实现 `WebFetchTool`：HTTP GET、HTML 标签剥离、50K 字符截断（`internal/tool/builtin/webfetch.go`）

## 4. 沙箱与权限引擎

- [x] 4.1 定义 `PermissionLevel` 类型，含有序常量（NONE=0 到 DANGEROUS=4）和 `Covers()` 方法（`internal/sandbox/level.go`）
- [x] 4.2 定义 `PermissionDecision` 结构体，含 Verdict 枚举（ALLOW/DENY/PROMPT）（`internal/sandbox/decision.go`）
- [x] 4.3 定义 `Sandbox` 接口：`ExecuteCommand()`、`IsInWorkspace()`、`PolicyEngine()`（`internal/sandbox/sandbox.go`）
- [x] 4.4 实现 `PolicyEngine`：三层检查（用户覆盖 → 级别+频率 → 升级提示）、滑动窗口频率追踪、线程安全 map（`internal/sandbox/policy.go`）
- [x] 4.5 实现 `SubprocessSandbox`：跨平台命令执行（bash -c / cmd /c）、通过 context 超时控制、命令黑名单、stderr 合并（`internal/sandbox/subprocess.go`）

## 5. 配置系统

- [x] 5.1 定义 `Config` 结构体及所有子配置（`ModelConfig`、`PermissionConfig`、`SandboxConfig`、`MCPServerConfig`），含 YAML 结构体标签（`internal/config/config.go`）
- [x] 5.2 实现 `Defaults()`，返回硬编码的合理默认值
- [x] 5.3 实现 `Load()`：全局配置 → 项目配置 → 环境变量覆盖，嵌套配置深度合并
- [x] 5.4 实现 CLI 覆盖方法（`WithModel()`、`WithWorkspace()`、`WithPermissionLevel()`），直接修改结构体
- [x] 5.5 实现 `ResolveAPIKey()`，优先级：显式密钥 → api_key_env → 提供商默认环境变量

## 6. ReAct 循环与编排器

- [x] 6.1 定义 `AgentState`、`AgentResult` 和 `ToolCallRecord` 结构体（`internal/agent/state.go`、`internal/agent/result.go`）
- [x] 6.2 实现 `EventBus`：类型化订阅/发布，使用 `sync.RWMutex`，基于 channel 的变体用于 bubbletea 集成（`internal/agent/eventbus.go`）
- [x] 6.3 实现 `ReActLoop.Run()`：构建初始消息 → 步骤循环 → LLM 调用 → 解析工具调用 → 权限检查 → 执行 → 追加观察结果 → 达到上限后强制结束（`internal/agent/react.go`）
- [x] 6.4 集成进度 channel：每步发送 `AgentState` 供 TUI 实时更新
- [x] 6.5 集成权限 channel：PROMPT 判定时阻塞等待，等 TUI 返回用户响应
- [x] 6.6 实现并发工具执行：多个独立 tool_call 通过 goroutine + `sync.WaitGroup` 并发执行
- [x] 6.7 实现 `Orchestrator`：串联 Config → ProviderFactory → ToolRegistry → Sandbox → ReActLoop，注册全部 6 个内置工具，构建系统提示（`internal/agent/orchestrator.go`）

## 7. TUI REPL

- [x] 7.1 定义所有 bubbletea 消息类型：`TaskSubmittedMsg`、`AgentProgressMsg`、`AgentResultMsg`、`ToolCallStartedMsg`、`ToolCallFinishedMsg`、`PermissionPromptMsg`、`PermissionResponseMsg`（`internal/cli/messages.go`）
- [x] 7.2 定义 `Model` 结构体，含全部状态（输入 textarea、viewport、spinner、config、focus、channels）（`internal/cli/model.go`）
- [x] 7.3 定义 lipgloss 样式：header、状态栏、输入区、帮助栏、模态覆盖层（`internal/cli/styles.go`）
- [x] 7.4 实现 `View()`：header → viewport → 状态 → 输入 → 帮助栏，权限提示活跃时显示模态覆盖层（`internal/cli/view.go`）
- [x] 7.5 实现 `Update()`：键盘路由（Tab/Ctrl+C/Ctrl+Enter）、agent 消息处理（progress/result/permission）、spinner tick、子组件更新（`internal/cli/update.go`）
- [x] 7.6 实现 `listenProgress()`：消费 progress channel → 返回 `AgentProgressMsg` → 自循环持续监听
- [x] 7.7 实现权限模态渲染：工具名、原因、[Allow] [Deny] [Always Allow] 按钮、120s 超时
- [x] 7.8 实现斜杠命令：`/help`、`/model [name]`、`/permission [level]`、`/clear`、`/history`、`/exit`（`internal/cli/commands.go`）
- [x] 7.9 实现 `Init()`：spinner 启动、输入区获取焦点
- [x] 7.10 实现 `Formatter` 和 `MarkdownRenderer`：粗体/斜体/代码/标题/列表 → lipgloss ANSI 样式（`internal/cli/formatter.go`、`internal/cli/markdown.go`）
- [x] 7.11 实现 `handleTaskSubmit()`：创建 orchestrator、启动 agent goroutine、连接 progress/result channels、返回 tea.Batch（listenProgress + waitForCompletion + spinner）

## 8. MCP 集成

- [x] 8.1 添加 `github.com/modelcontextprotocol/go-sdk` 依赖
- [x] 8.2 实现 `MCPServerConfig` 的 YAML 加载：name、transport、command/args（stdio）、url（http）、env vars
- [x] 8.3 实现 MCP 客户端封装：通过 stdio 子进程或 HTTP transport 连接、获取工具列表、注册进 ToolRegistry（`internal/mcp/client.go`）
- [x] 8.4 实现延迟连接：启动时缓存 MCP 工具元数据，实际服务器连接推迟到首次工具调用时（`internal/mcp/lazy.go`）
- [x] 8.5 实现 MCP 工具适配器：将 MCP 工具包装为 `Tool` 接口，execute 调用路由到 MCP 服务器，响应转换为 `ToolResult`
- [x] 8.6 处理 MCP 错误情况：连接失败 → error observation、服务器崩溃 → 重连尝试、工具名重复 → 跳过并警告

## 9. 测试

- [x] 9.1 测试 `AnthropicProvider` JSON 解析：基于 fixture 的文本响应、tool_use 响应、混合 content block、错误响应测试
- [x] 9.2 测试 `OpenAIProvider` JSON 解析：标准响应、tool_calls 响应、DeepSeek reasoning_content 测试
- [x] 9.3 测试 `ModelRegistry` 自动检测：精确匹配、前缀匹配、未知模型回退
- [x] 9.4 测试 `ReActLoop`：mock LLM provider、验证步骤计数、工具执行、达到上限后强制结束、权限拒绝处理
- [x] 9.5 测试 `PolicyEngine`：级别覆盖检查、频率窗口节流、用户覆盖优先级
- [x] 9.6 测试 `Config` 加载：默认值、YAML 文件合并、环境变量覆盖、CLI 覆盖
- [x] 9.7 测试 `FileEditTool`：唯一匹配、无匹配、多次匹配
- [x] 9.8 测试 `FileSearchTool`：glob 转正则转换、grep 结果格式化
- [x] 9.9 测试 `SubprocessSandbox`：命令执行、超时、黑名单阻止

## 10. 构建、打磨与文档

- [x] 10.1 创建 `Makefile`：build、test、lint、clean、build-all 目标
- [x] 10.2 添加 `//go:build` 平台特定约束（Windows 终端初始化）
- [x] 10.3 编写 `README.md`：安装、配置、使用示例、模型预设表
- [x] 10.4 跨平台冒烟测试：Windows Terminal、macOS Terminal、Linux（WSL/GNOME Terminal）
- [x] 10.5 手动集成测试：使用真实 LLM API 完成完整 ReAct 循环，验证工具调用正确执行
- [x] 10.6 手动 TUI 测试：多行输入、实时进度、权限模态、斜杠命令、Ctrl+C 取消、窗口缩放行为
- [x] 10.7 边界情况测试：空工具响应、超长输出截断、并发工具执行顺序、MCP 服务器任务中断开连接
