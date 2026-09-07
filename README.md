# CodingAgent

**单二进制、零运行时依赖的 ReAct 架构 AI 编程助手，Go 实现。**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8)
![Bubbletea](https://img.shields.io/badge/TUI-Bubbletea-FF4E8E)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)
![Binary](https://img.shields.io/badge/binary-11.4MB-blue)

CodingAgent 不做通用聊天，专注编程任务的高频闭环：接收自然语言指令 → ReAct 循环推理与工具执行 → 危险操作内联审批 → 交付结果。

```text
自然语言任务
  -> ReAct 循环（推理 + 工具并行执行，最多 50 轮迭代）
  -> 读 / 写 / 改 / 搜 / bash / 网页抓取
  -> 高危操作弹内联权限确认
  -> 交付最终答案
```

核心设计立场：**模型负责理解、规划和生成；CodingAgent 负责工具边界、权限、审批和执行。** 模型输出永远不会绕过权限系统直接执行。

## 功能特性

### 编程闭环

- **ReAct 循环**——LLM 推理与工具执行交替进行；工具调用并行执行；达到迭代上限时自动收敛为最终答案而非直接失败
- **六个内置工具**——文件读取（行范围 + 10 万字符截断）、写入（路径逃逸守卫）、精确编辑（字符串替换，歧义即失败）、双模式搜索（glob 模式 + 正则 grep，自动跳过二进制文件）、bash 执行、网页抓取（剥标签 + 10 次重定向上限）
- **全屏 TUI**——bubbletea REPL：欢迎页、Markdown 渲染、彩色权限徽标、实时状态栏（步骤 / 工具 / token 估算）、工具结果 ✓/✗ 前缀、自动滚动
- **可取消任务**——Ctrl+C 立即中止在途 LLM 请求；LLM 调用、bash 执行全程带 context 超时，不会无限挂起

### 权限与安全

- **五级权限**——`NONE / READ / WRITE / EXECUTE / DANGEROUS`，高级别覆盖低级别；TUI 内 `Ctrl+P` 一键循环切换，徽标实时变色（灰 / 黄 / 橙 / 红）
- **三级检查**——用户永久覆盖表（弹窗按 `A` 记住）→ 当前级别判定 → 级别不足时弹内联确认框（`Y` 允许一次 / `N` 拒绝 / `A` 永久允许）
- **频率节流**——默认每工具 10 次 / 300 秒，超出后即使权限足够也需重新授权，防止失控循环
- **命令黑名单**——`rm -rf /`、`sudo`、`mkfs` 等高危命令硬编码拦截；超时杀整棵进程树

### 模型接入

- **9 个 Provider 预设**——Anthropic、OpenAI、DeepSeek、Ollama、Gemini、Moonshot、Qwen、Zhipu、vLLM；自动按模型名识别端点与协议
- **无 SDK 依赖**——Anthropic 与 OpenAI 兼容协议均用 `net/http` 手写，整个项目仅 6 个直接依赖
- **思考模式**——DeepSeek ReasoningContent 透传展示
- **MCP 客户端**——stdio + HTTP 双传输的 Model Context Protocol 模块已实现（延迟连接 + 工具适配），运行时集成开发中

## 架构

```text
cmd/coding-agent        Cobra 入口 + Windows VT 终端启用
        |
        v
internal/cli            bubbletea TUI（模型 / 更新 / 视图三件套）
  - 斜杠命令、按键处理、权限弹窗、Markdown 渲染
  - 4 组 channel 连接 TUI 与 agent goroutine
        |
        v
internal/agent          Orchestrator 组装五层
  - ReActLoop（最大 50 迭代）+ EventBus（非阻塞发布订阅）
        |
   +----+----+----------------+
   |         |                |
   v         v                v
internal/llm          internal/tool          internal/sandbox
Provider 接口          6 个内置工具           PolicyEngine
Anthropic + OpenAI     线程安全注册表         5 级权限 + 频率节流
兼容协议手写实现       工作区路径守卫          子进程沙箱 + 进程树终止
        |
        v
internal/config        分层配置合并
internal/mcp           MCP 客户端（stdio / HTTP）
```

配置层级：`硬编码默认值 → ~/.coding-agent/config.yaml（全局） → .coding-agent/config.yaml（项目） → 环境变量 → CLI 参数`，非零字段逐级覆盖。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 语言 | Go 1.25（无 cgo，交叉编译友好） |
| TUI | bubbletea 1.3、bubbles、lipgloss |
| CLI | Cobra 1.10 |
| LLM | Anthropic 协议 + OpenAI 兼容协议（`net/http` 手写，无 SDK） |
| 配置 | YAML 分层合并（gopkg.in/yaml.v3） |
| 验证 | go test（6 个测试文件，离线可跑，无需 API key） |

## 快速开始

### 前置条件

- Go 1.25+
- 任一支持 Provider 的 API key（Ollama / vLLM 本地部署可免 key）

### 1. 构建

```powershell
git clone https://github.com/AsdsSsh/CodingAgent.git
cd CodingAgent
make build
./coding-agent          # Windows: .\coding-agent.exe
```

### 2. 配置密钥

环境变量方式（推荐）：

```powershell
$env:ANTHROPIC_API_KEY = "sk-ant-..."   # Claude
$env:OPENAI_API_KEY    = "sk-..."       # GPT
$env:DEEPSEEK_API_KEY  = "sk-..."       # DeepSeek
$env:DASHSCOPE_API_KEY = "sk-..."       # Qwen
```

或配置文件（`~/.coding-agent/config.yaml` 或项目级 `.coding-agent/config.yaml`）：

```yaml
model:
  provider: anthropic
  model: claude-sonnet-4-20250514
  max_tokens: 8192
  context_window: 200000

permissions:
  default_level: READ          # NONE / READ / WRITE / EXECUTE / DANGEROUS
  max_frequency: 10            # 每工具频率上限
  frequency_window_seconds: 300
```

### 3. 启动

```powershell
coding-agent                              # 默认配置进入 TUI
coding-agent -m gpt-4o                    # 指定模型（自动识别 provider）
coding-agent -m qwen2.5:7b -u http://localhost:11434/v1   # 本地 Ollama
coding-agent -w C:\path\to\project        # 指定工作区
coding-agent -m glm-4 -k <你的key>        # 显式传入密钥
```

### 4. TUI 操作

**斜杠命令：**

| 命令 | 别名 | 说明 |
| --- | --- | --- |
| `/help` | `/h` | 查看可用命令 |
| `/model [name]` | `/m` | 查看或切换模型 |
| `/permission [level]` | `/perm` `/p` | 查看或切换权限级别（无参 = 循环切换） |
| `/clear` | | 清空会话历史 |
| `/exit` | `/quit` `/q` | 退出 |

**快捷键：**

| 键 | 动作 |
| --- | --- |
| `Enter` | 提交单行任务 |
| `Ctrl+S` | 提交多行任务 |
| `Ctrl+P` | 循环切换权限（READ → WRITE → EXECUTE → DANGEROUS） |
| `Ctrl+C` | 运行中取消任务 / 空闲时退出 |
| `Tab` | 焦点切换（输入框 ↔ 视口） |

权限弹窗期间：`Y`/`Enter` 允许一次，`N`/`Esc` 拒绝，`A` 永久允许该工具。

## 模型预设

| Provider | 模型前缀 | 默认模型 | 上下文窗口 | 环境变量 |
| --- | --- | --- | --- | --- |
| Anthropic | `claude-` | claude-sonnet-4-20250514 | 200K | `ANTHROPIC_API_KEY` |
| OpenAI | `gpt-` `o1` `o3` `o4` | gpt-4o | 128K | `OPENAI_API_KEY` |
| DeepSeek | `deepseek-` | deepseek-v4-pro | 1M | `DEEPSEEK_API_KEY` |
| Ollama | （任意） | qwen2.5:7b | 32K | — |
| Gemini | `gemini-` | gemini-2.0-flash | 1M | `GEMINI_API_KEY` |
| Moonshot | `moonshot-` `kimi-` | moonshot-v1-8k | 128K | `MOONSHOT_API_KEY` |
| Qwen | `qwen-` | qwen-max | 131K | `DASHSCOPE_API_KEY` |
| Zhipu | `glm-` | glm-4 | 128K | `ZHIPU_API_KEY` |
| vLLM | （任意） | （自定义） | 32K | — |

密钥解析顺序：显式 `-k` → 预设环境变量 → `OPENAI_API_KEY` → `ANTHROPIC_API_KEY`。

## 验证与构建

```powershell
make build          # 当前平台构建（Windows 自动加 .exe）
make build-release  # 去符号构建（更小）
make build-all      # 交叉编译 Windows / Linux / macOS (amd64 + arm64)
make test           # 全部测试（离线，无需 API key）
make test-cover     # 覆盖率 + HTML 报告
make lint           # golangci-lint（需单独安装）
```

## 项目结构

```text
cmd/coding-agent/     入口：Cobra CLI + Windows VT 处理
internal/
  agent/              ReActLoop、Orchestrator、EventBus
  cli/                TUI（bubbletea）：模型、更新、视图、Markdown
  config/             分层 YAML 加载与合并
  llm/                Provider 接口 + Anthropic/OpenAI 协议 + 9 预设
  mcp/                MCP 客户端（stdio / HTTP，延迟连接）
  sandbox/            权限级别、PolicyEngine、子进程沙箱、进程树终止
  tool/               Tool 接口、注册表、工作区守卫
    builtin/          read / write / edit / search / bash / web_fetch
config/               默认 YAML 配置模板
openspec/             开发流程规格与变更归档
```

## 许可证

[MIT](./LICENSE) © 2026 AsdsSsh

---

[English Version](./README_EN.md)
