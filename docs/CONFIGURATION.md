# 配置参考

## 配置文件

CodingAgent 采用分层配置加载策略，优先级从低到高：

```
硬编码默认值
  ← ~/.coding-agent/config.yaml     (全局配置)
  ← .coding-agent/config.yaml       (项目级配置)
  ← 环境变量
  ← CLI 参数
```

---

## 完整配置模板

```yaml
# ====== CodingAgent 配置 ======

# -------- 模型配置 --------
model:
  provider: anthropic              # 提供商: anthropic | openai
  model: claude-sonnet-4-20250514  # 模型名称
  # api_key: ""                    # API 密钥（不推荐写入文件）
  # api_key_env: ANTHROPIC_API_KEY # 自定义环境变量名
  # base_url: ""                   # 自定义端点 (OpenAI 兼容)
  # temperature: 0.7               # null=默认, 0.0=精确, 1.0=创意
  max_tokens: 8192                 # 单次响应最大 token 数
  context_window: 200000           # 模型总上下文窗口

# -------- MCP 服务器（Phase 3）--------
mcp_servers: []
  # - name: my-server
  #   command: npx
  #   args: ["-y", "@modelcontextprotocol/server-filesystem"]

# -------- 权限配置 --------
permissions:
  default_level: READ              # NONE | READ | WRITE | EXECUTE | DANGEROUS
  max_frequency: 10                # 时间窗口内最大使用次数
  frequency_window_seconds: 300    # 频率统计时间窗口（秒）

# -------- 工作区 --------
workspace: ""                      # 空 = 当前目录

# -------- 执行限制 --------
max_iterations: 50                 # ReAct 循环最大步数

# -------- 记忆存储 --------
memory_path: ""                    # 空 = ~/.coding-agent/memory.jsonl

# -------- 沙箱配置 --------
sandbox:
  type: subprocess                 # 沙箱类型
  timeout_seconds: 120             # 命令超时（秒）
  blocked_commands:                # 拦截的危险命令
    - "rm -rf /"
    - "sudo"
    - "dd if="
    - "mkfs"
    - ":(){ :|:& };:"
    - "shutdown"
    - "reboot"
```

---

## 常用模型配置示例

### Anthropic Claude

```yaml
model:
  provider: anthropic
  model: claude-sonnet-4-20250514        # Sonnet 4（推荐）
  # model: claude-opus-4-20250514        # Opus 4（最强）
  # model: claude-haiku-4-5-20251001     # Haiku 4.5（最快）
  max_tokens: 8192
  context_window: 200000
```

### OpenAI GPT-4o

```yaml
model:
  provider: openai
  model: gpt-4o
  max_tokens: 4096
  context_window: 128000
```

### DeepSeek

```yaml
model:
  provider: openai
  model: deepseek-v4-pro
  base_url: https://api.deepseek.com/v1
  api_key_env: DEEPSEEK_API_KEY
  max_tokens: 32768
  context_window: 1000000
```

### Ollama 本地模型

```yaml
model:
  provider: openai
  model: qwen2.5:7b                    # 需先 ollama pull
  base_url: http://localhost:11434/v1
  api_key: ollama                      # Ollama 不需要真实 Key
  max_tokens: 4096
  context_window: 32768
```

### vLLM 自部署

```yaml
model:
  provider: openai
  model: meta-llama/Meta-Llama-3-8B
  base_url: http://localhost:8000/v1
  api_key: not-needed
  max_tokens: 4096
  context_window: 8192
```

### Gemini (via OpenAI 兼容层)

```yaml
model:
  provider: openai
  model: gemini-2.0-flash
  base_url: https://generativelanguage.googleapis.com/v1beta/openai
  api_key_env: GEMINI_API_KEY
  max_tokens: 4096
  context_window: 1048576
```

---

## CLI 参数

| 参数 | 短参数 | 说明 | 示例 |
|------|--------|------|------|
| `--config` | `-c` | 配置文件路径 | `-c /path/to/config.yaml` |
| `--provider` | `-p` | LLM 提供商 | `-p openai` |
| `--model` | `-m` | 模型名称 | `-m gpt-4o` |
| `--api-key` | `-k` | API 密钥 | `-k "sk-..."` |
| `--temperature` | `-t` | 温度参数 | `-t 0.7` |
| `--workspace` | `-w` | 工作区目录 | `-w /my/project` |
| `--help` | `-h` | 显示帮助 | |

---

## 环境变量

| 变量 | 说明 |
|------|------|
| `ANTHROPIC_API_KEY` | Anthropic API 密钥 |
| `OPENAI_API_KEY` | OpenAI API 密钥 |
| `DEEPSEEK_API_KEY` | DeepSeek API 密钥 |
| `GEMINI_API_KEY` | Gemini API 密钥 |

---

## 权限级别

| 级别 | 值 | 允许的操作 |
|------|-----|-----------|
| `NONE` | 0 | 无任何操作 |
| `READ` | 1 | 文件读取、搜索、网页抓取 |
| `WRITE` | 2 | + 文件写入、编辑 |
| `EXECUTE` | 3 | + Bash 命令执行 |
| `DANGEROUS` | 4 | 所有操作（含危险命令） |

---

## 权限策略配置 (`permissions.yaml`)

```yaml
# 自动批准：设为 true 的工具跳过弹窗确认
auto_approve:
  read: true
  search: true
  web_fetch: true

# 频率限制
max_frequency: 10
frequency_window_seconds: 300

# 工具级别覆盖
tool_overrides:
  # bash: false       # 始终弹窗确认 Bash 命令
  # write: true       # 始终允许文件写入
  # edit: true        # 始终允许文件编辑
```
