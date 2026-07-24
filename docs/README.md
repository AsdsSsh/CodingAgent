# CodingAgent 项目文档

> **轻量级 AI 编程助手** — 基于 ReAct (Reasoning + Acting) 架构的 Java 命令行工具，支持多种 LLM 提供商。

---

## 目录

- [架构概览](#架构概览)
- [模块说明](#模块说明)
- [配置指南](#配置指南)
- [工具系统](#工具系统)
- [权限与沙箱](#权限与沙箱)
- [快速开始](#快速开始)
- [开发指南](#开发指南)

---

## 架构概览

```
┌─────────────────────────────────────────────────────┐
│                     用户输入                          │
└──────────────────────┬──────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────────┐
│  CodingAgentApp (入口)  →  CliApp (Picocli CLI)      │
└──────────────────────┬──────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────────┐
│               Repl (交互式 REPL)                      │
│   JLine 多行输入 · 斜杠命令 · 对话历史 · 权限切换      │
└──────────────────────┬──────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────────┐
│            Orchestrator (编排器)                      │
│   配置加载 → API密钥解析 → 工具注册 → 沙箱初始化        │
└──────────────────────┬──────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────────┐
│              ReActLoop (核心循环)                     │
│   Thought → Action → Observation → 循环, 最多50步      │
└──────────┬─────────────┬──────────────┬─────────────┘
           ▼             ▼              ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  LlmProvider │ │  ToolRegistry│ │  SandboxBase │
│  (LLM调用)   │ │  (工具注册)   │ │  (沙箱执行)  │
└──────┬───────┘ └──────┬───────┘ └──────┬───────┘
       ▼                ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ Anthropic    │ │ FileRead     │ │ Subprocess   │
│ OpenAI       │ │ FileWrite    │ │ Sandbox      │
│ (扩展点)     │ │ FileEdit     │ │ +PolicyEngine│
└──────────────┘ │ FileSearch   │ └──────────────┘
                 │ BashTool     │
                 │ WebFetchTool │
                 └──────────────┘
```

---

## 模块说明

### 1. 入口与应用层 (`com.codingagent`, `com.codingagent.cli`)

| 文件 | 说明 |
|------|------|
| `CodingAgentApp.java` | 应用入口，委托给 `CliApp` |
| `CliApp.java` | Picocli 命令行解析，支持 `-p`（提供商）、`-m`（模型）、`-k`（API密钥）、`-t`（温度）、`-w`（工作区）等参数 |
| `Repl.java` | 交互式 REPL，基于 JLine 3，支持多行输入（行尾 `\`）、斜杠命令（`/help`、`/permission`、`/clear`、`/history`、`/exit`）、对话历史和上下文携带 |
| `Formatter.java` | 结果格式化，支持人类可读（文本）和机器可读（JSON）两种模式 |

**REPL 命令列表：**
| 命令 | 功能 |
|------|------|
| `/help` 或 `/h` | 显示帮助 |
| `/permission <level>` | 查看/切换权限级别 (read/write/execute/dangerous) |
| `/clear` | 清除对话历史 |
| `/history` | 显示对话历史 |
| `/exit` 或 `/quit` 或 `/q` | 退出 REPL |
| `Ctrl+C` | 中断当前任务 |
| `Ctrl+D` | 退出 REPL |

---

### 2. 核心 Agent 层 (`com.codingagent.agent`)

| 文件 | 说明 |
|------|------|
| `Orchestrator.java` | **顶层编排器**。串联五个层次：加载配置 → 创建 LLM 提供商 → 注册工具 → 初始化沙箱 → 构建系统提示并委托 ReActLoop 执行。是记忆检索、指令加载、MCP 服务器注册的主要扩展点 |
| `ReActLoop.java` | **核心 ReAct 循环**。算法流程：构建上下文 → 调用 LLM → 解析响应（文本/工具调用）→ 权限检查 → 执行工具 → 追加观察结果 → 循环。最多执行 50 步（可配置），超出后强制要求最终回答 |
| `AgentResult.java` | 不可变结果记录，包含最终回答文本、步数、工具调用记录列表、完整对话 |
| `AgentState.java` | 单步状态快照，通过 EventBus 发送供实时显示 |
| `EventBus.java` | 轻量级进程内发布/订阅事件总线。CLI 层订阅事件实现实时渲染（Thought → Tool Call → Observation），线程安全 |

**ReAct 循环流程图：**

```
用户任务
   │
   ▼
构建消息上下文 [系统提示 + 记忆 + 用户任务]
   │
   ▼  ┌─────────────────────────┐
   └─▶│  LLM 调用（带工具定义）   │◀──── 追加 tool_result
      └───────────┬─────────────┘
                  │
          ┌───────┴───────┐
          ▼               ▼
    有工具调用？       无工具调用？
          │               │
          ▼               ▼
    遍历每个 tool_call   返回最终回答
          │
          ▼
    PolicyEngine 权限检查
          │
    ┌─────┼─────┐
    ▼     ▼     ▼
  ALLOW PROMPT DENY
    │     │     │
    ▼     ▼     ▼
 执行工具 返回   返回
    │   错误    错误
    ▼
 追加 observation
```

---

### 3. LLM 层 (`com.codingagent.llm`)

| 文件 | 说明 |
|------|------|
| `LlmProvider.java` | LLM 提供商接口，定义 `chat()`（对话）和 `countMessagesTokens()`（Token 计数） |
| `AnthropicProvider.java` | Anthropic API 实现（Claude 系列：Sonnet / Opus / Haiku） |
| `OpenAIProvider.java` | OpenAI 兼容 API 实现（GPT-4、DeepSeek、Ollama、vLLM、Gemini 等） |
| `ProviderFactory.java` | 工厂类，根据 `ModelConfig.provider()` 自动选择实现。`"anthropic"` 走 Anthropic 专有格式，其余走 OpenAI 兼容格式 |
| `Message.java` | 消息对象（system / user / assistant / tool_result），包含 ToolCall 子类 |
| `LlmResponse.java` | LLM 响应封装，包含文本内容、工具调用列表、停止原因、推理内容、Token 用量 |
| `TokenUsage.java` | Token 用量统计 |

**支持的 LLM 提供商：**

| 提供商 | 模型示例 | 说明 |
|--------|---------|------|
| **Anthropic** | claude-sonnet-4-20250514, claude-opus-4-20250514, claude-haiku-4-5-20251001 | 官方 API |
| **OpenAI** | gpt-4o, gpt-4-turbo | 官方 API |
| **DeepSeek** | deepseek-v4-pro | 通过 OpenAI 兼容层 + base_url |
| **Ollama** | qwen2.5:7b 等 | 本地部署，OpenAI 兼容 |
| **vLLM** | meta-llama/Meta-Llama-3-8B 等 | 自部署，OpenAI 兼容 |
| **Gemini** | gemini-2.0-flash | 通过 OpenAI 兼容层 |

---

### 4. 工具系统 (`com.codingagent.tool`)

| 文件 | 说明 |
|------|------|
| `ToolBase.java` | 工具抽象基类，定义 `name()`、`description()`、`parameters()`、`requiredPermission()`、`execute()` |
| `ToolRegistry.java` | 工具注册中心，支持按名称查找、检查存在性、导出为 LLM 兼容的 JSON Schema 格式 |
| `ToolContext.java` | 工具执行上下文，包含工作区路径和沙箱引用 |
| `ToolResult.java` | 工具执行结果，提供 `formatForLlm()` 方法序列化为 LLM 可读文本 |

#### 内置工具

| 工具 | 类 | 功能 | 所需权限 |
|------|-----|------|---------|
| **read** | `FileReadTool` | 读取文件内容，支持 offset/limit 范围读取 | READ |
| **write** | `FileWriteTool` | 创建新文件或覆盖现有文件 | WRITE |
| **edit** | `FileEditTool` | 在文件中执行精确的字符串替换 | WRITE |
| **search** | `FileSearchTool` | 按 glob 模式搜索文件名，或按正则搜索文件内容 | READ |
| **bash** | `BashTool` | 执行 Bash 命令（受沙箱约束） | EXECUTE |
| **web_fetch** | `WebFetchTool` | 获取 URL 内容（用于阅读文档或 API 响应） | READ |

---

### 5. 权限与沙箱 (`com.codingagent.sandbox`)

| 文件 | 说明 |
|------|------|
| `PermissionLevel.java` | 权限级别枚举：`NONE` < `READ` < `WRITE` < `EXECUTE` < `DANGEROUS` |
| `PermissionDecision.java` | 权限判断结果：`ALLOW`（允许）、`DENY`（拒绝）、`PROMPT`（需用户确认） |
| `PolicyEngine.java` | 权限策略引擎，根据工具所需权限与当前权限级别做对比决策 |
| `SandboxBase.java` | 沙箱抽象基类 |
| `SubprocessSandbox.java` | 子进程沙箱实现，支持命令超时和危险命令拦截 |

**安全特性：**
- 默认拦截的命令：`rm -rf /`、`sudo`、`dd if=`、`mkfs`、fork bomb、`shutdown`、`reboot`
- 可配置的命令超时（默认 120 秒）
- 频率限制：时间窗口内超出次数后重新弹窗确认（默认 5 分钟内 10 次）
- REPL 中可实时切换权限级别：`/permission read|write|execute|dangerous`

---

### 6. 配置系统 (`com.codingagent.config`)

| 文件 | 说明 |
|------|------|
| `AgentConfig.java` | 根配置对象（Java Record），支持分层加载策略 |
| `ModelConfig.java` | 不可变模型配置，支持 API 密钥多级解析 |
| `PermissionConfig.java` | 权限策略配置 |
| `McpServerConfig.java` | MCP 服务器配置（Phase 3 预留） |

**配置加载优先级（后覆盖前）：**
1. 代码内硬编码默认值
2. 全局配置 `~/.coding-agent/config.yaml`
3. 项目级配置 `.coding-agent/config.yaml`
4. 环境变量（`ANTHROPIC_API_KEY` / `OPENAI_API_KEY`）
5. CLI 参数（`-p` / `-m` / `-k` / `-t` / `-w`）

---

## 快速开始

### 环境要求

- **Java 21+**
- **Maven 3.8+**

### 构建

```bash
mvn clean package
```

### 运行

```bash
# 使用默认配置（Anthropic Claude Sonnet 4）
java -jar target/coding-agent-1.0-SNAPSHOT-jar-with-dependencies.jar

# 指定模型
java -jar target/coding-agent-1.0-SNAPSHOT-jar-with-dependencies.jar -m gpt-4o -p openai

# 指定工作区
java -jar target/coding-agent-1.0-SNAPSHOT-jar-with-dependencies.jar -w /path/to/project

# 指定 API 密钥
java -jar target/coding-agent-1.0-SNAPSHOT-jar-with-dependencies.jar -k "sk-..."
```

### 配置 API 密钥

推荐方式：设置环境变量

```bash
export ANTHROPIC_API_KEY="sk-ant-..."   # Anthropic
export OPENAI_API_KEY="sk-..."          # OpenAI
export DEEPSEEK_API_KEY="sk-..."        # DeepSeek
```

---

## 开发指南

### 项目结构

```
coding-agent/
├── config/
│   ├── default_config.yaml    # 默认配置模板
│   └── permissions.yaml       # 权限策略模板
├── src/main/java/com/codingagent/
│   ├── CodingAgentApp.java    # 入口
│   ├── agent/                 # 核心 Agent 逻辑
│   │   ├── Orchestrator.java  # 编排器
│   │   ├── ReActLoop.java     # ReAct 循环
│   │   ├── AgentResult.java   # 结果记录
│   │   ├── AgentState.java    # 状态快照
│   │   └── EventBus.java      # 事件总线
│   ├── cli/                   # CLI / REPL
│   │   ├── CliApp.java        # Picocli 命令行
│   │   ├── Repl.java          # JLine 交互式 REPL
│   │   └── Formatter.java     # 输出格式化
│   ├── config/                # 配置加载
│   │   ├── AgentConfig.java   # 根配置
│   │   ├── ModelConfig.java   # 模型配置
│   │   ├── PermissionConfig.java
│   │   └── McpServerConfig.java
│   ├── llm/                   # LLM 提供商抽象
│   │   ├── LlmProvider.java   # 接口
│   │   ├── AnthropicProvider.java
│   │   ├── OpenAIProvider.java
│   │   ├── ProviderFactory.java
│   │   ├── Message.java
│   │   ├── LlmResponse.java
│   │   └── TokenUsage.java
│   ├── sandbox/               # 权限与沙箱
│   │   ├── SandboxBase.java
│   │   ├── SubprocessSandbox.java
│   │   ├── PolicyEngine.java
│   │   ├── PermissionLevel.java
│   │   └── PermissionDecision.java
│   └── tool/                  # 工具系统
│       ├── ToolBase.java
│       ├── ToolRegistry.java
│       ├── ToolContext.java
│       ├── ToolResult.java
│       └── builtin/
│           ├── FileReadTool.java
│           ├── FileWriteTool.java
│           ├── FileEditTool.java
│           ├── FileSearchTool.java
│           ├── BashTool.java
│           └── WebFetchTool.java
├── pom.xml                    # Maven 配置
└── docs/                      # 文档（本目录）
```

### 技术栈

| 组件 | 版本 | 用途 |
|------|------|------|
| Java | 21 | 运行环境 |
| Picocli | 4.7.6 | CLI 参数解析 |
| JLine 3 | 3.26.3 | REPL 交互终端 |
| Jackson | 2.18.0 | JSON/YAML 解析 |
| Anthropic Java SDK | 0.6.0 | Anthropic API 调用 |
| OpenAI Java SDK | 0.18.0 | OpenAI 兼容 API 调用 |
| JUnit 5 | 5.11.0 | 单元测试 |
| Mockito | 5.12.0 | 测试 Mock |

### 扩展点

1. **新增 LLM 提供商**：实现 `LlmProvider` 接口，在 `ProviderFactory` 中注册
2. **新增工具**：继承 `ToolBase`，在 `Orchestrator.registerBuiltinTools()` 中注册
3. **MCP 服务器集成**（Phase 3 计划）：通过 `McpServerConfig` 配置，在 Orchestrator 中动态加载
4. **记忆系统**（Phase 5 计划）：跨会话记忆的存储和检索，通过 `memory.jsonl` 持久化
5. **自定义沙箱**：继承 `SandboxBase`，在配置中指定 `sandbox.type`

### 设计理念

- **安全优先**：多级权限 + 命令黑名单 + 频率限制，所有工具调用需经过 PolicyEngine
- **可扩展**：LLM 提供商、工具、沙箱均可插拔替换
- **不可变配置**：Java Record 保证配置一经创建不可变
- **分层加载**：配置从默认值 → 全局 → 项目 → 环境变量 → CLI 参数逐层覆盖
- **前后端解耦**：EventBus 实现 Agent 状态与 CLI 渲染的解耦
