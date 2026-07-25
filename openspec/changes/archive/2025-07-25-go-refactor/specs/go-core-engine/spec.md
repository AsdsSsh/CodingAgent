## ADDED 需求

### Requirement: ReAct 循环执行 Thought-Action-Observation 周期
系统 SHALL 实现一个 ReAct（推理+行动）循环，在 LLM 推理和工具执行之间交替进行，直至任务完成或达到步骤上限。

#### Scenario: 任务以最终回答完成
- **WHEN** LLM 返回一个响应，其中 `stop_reason: "end_turn"`（Anthropic）或 `finish_reason: "stop"`（OpenAI）且无工具调用
- **THEN** ReActLoop SHALL 返回一个 AgentResult，包含回答文本、步骤计数和工具调用记录

#### Scenario: 任务达到最大迭代次数
- **WHEN** 步骤计数达到 `maxIterations`（默认 50）
- **THEN** ReActLoop SHALL 追加一条消息指示模型给出最终回答，再调用一次 LLM，然后返回结果

#### Scenario: LLM 请求工具执行
- **WHEN** LLM 响应包含一个或多个工具调用
- **THEN** ReActLoop SHALL 执行每个工具，将工具结果作为观察追加，然后继续循环

#### Scenario: 请求了未知工具
- **WHEN** LLM 请求了一个未在 ToolRegistry 中注册的工具
- **THEN** ReActLoop SHALL 追加一条列出可用工具的错误观察，而不是崩溃

#### Scenario: 权限拒绝某个工具
- **WHEN** PolicyEngine 对某个工具调用返回 DENY 判定
- **THEN** ReActLoop SHALL 追加一条包含拒绝原因的错误观察，并继续循环

#### Scenario: 并发工具执行
- **WHEN** 同一步骤中的多个工具调用互不依赖
- **THEN** ReActLoop SHALL 通过 goroutine 并发执行它们，在消息列表中保持原始调用顺序

### Requirement: LLM 提供商抽象支持多种后端
系统 SHALL 提供一个通用的 `Provider` 接口，支持 Anthropic Messages API 和 OpenAI 兼容的 Chat Completions API，可通过配置选择。

#### Scenario: Anthropic API 调用
- **WHEN** 配置了 provider 为 `anthropic` 的模型
- **THEN** 系统 SHALL 向 `https://api.anthropic.com/v1/messages` 发送请求，带上 `x-api-key` 和 `anthropic-version` 头部，并将系统消息提取到顶层 `system` 字段

#### Scenario: OpenAI API 调用
- **WHEN** 配置了 provider 为 `openai` 或自定义 `base_url` 的模型
- **THEN** 系统 SHALL 向 `{base_url}/chat/completions` 发送请求，带上 `Authorization: Bearer` 头部

#### Scenario: 工具定义翻译
- **WHEN** 向 chat 方法提供了工具
- **THEN** AnthropicProvider SHALL 将其翻译为 `{name, description, input_schema}` 格式，OpenAIProvider SHALL 将其翻译为 `{type: "function", function: {name, description, parameters}}` 格式

#### Scenario: 根据预设自动检测模型
- **WHEN** 模型名称匹配内置预设前缀（如 "claude-"、"gpt-"、"deepseek-"）
- **THEN** 系统 SHALL 自动配置正确的 base_url、API 格式和密钥环境变量

#### Scenario: 未知模型且无 base_url
- **WHEN** 模型名称不匹配任何预设且未配置 base_url
- **THEN** 系统 SHALL 返回错误，列出可用的预设

### Requirement: 工具系统含注册表和六个内置工具
系统 SHALL 提供 `Tool` 接口和 `ToolRegistry`，后者管理工具并以 LLM 兼容格式导出其定义。

#### Scenario: 工具注册与查找
- **WHEN** 通过 `Register(tool)` 注册工具
- **THEN** ToolRegistry SHALL 能通过 `Resolve(name)` 按名称解析，通过 `Has(name)` 检查是否存在

#### Scenario: LLM 格式导出
- **WHEN** 调用 `ToLLMFormats()`
- **THEN** ToolRegistry SHALL 返回所有已注册工具的列表，格式为 `{name, description, input_schema}` map

#### Scenario: 带范围选择的文件读取
- **WHEN** 调用 `read` 工具，传 `file_path`、可选 `offset` 和可选 `limit`
- **THEN** 系统 SHALL 读取文件内容，标记行号，输出截断在 100,000 字符以内

#### Scenario: 带工作区边界检查的文件写入
- **WHEN** 调用 `write` 工具，传 `file_path` 和 `content`
- **THEN** 系统 SHALL 拒绝工作区外的路径，并按需创建父目录

#### Scenario: 强制唯一匹配的文件编辑
- **WHEN** 调用 `edit` 工具，传 `file_path`、`old_string` 和 `new_string`
- **THEN** 系统 SHALL 在 `old_string` 出现零次或多次时失败，仅在精确单次匹配时替换为 `new_string`

#### Scenario: glob 和 grep 模式的文件搜索
- **WHEN** 调用 `search` 工具，传 `pattern` 和 `mode`
- **THEN** 系统 SHALL 在 "glob" 模式下执行基于 glob 的文件名匹配，在 "grep" 模式下执行基于正则的内容匹配，结果上限 200 条

#### Scenario: 带沙箱强制的 Bash 执行
- **WHEN** 调用 `bash` 工具，传 `command`
- **THEN** 系统 SHALL 将执行委托给沙箱，沙箱会应用命令黑名单、超时和工作目录约束

#### Scenario: 带 HTML 剥离的网页抓取
- **WHEN** 调用 `web_fetch` 工具，传一个 URL
- **THEN** 系统 SHALL 抓取页面内容、剥离 HTML 标签、压缩空白、输出截断在 50,000 字符以内

### Requirement: 沙箱和权限引擎强制执行安全约束
系统 SHALL 提供 SubprocessSandbox，在超时、命令黑名单和工作区路径约束下执行命令，由带权限级别的 PolicyEngine 治理。

#### Scenario: 命令黑名单阻止危险模式
- **WHEN** 命令包含被阻止的模式（如 `rm -rf /`、`sudo`、`mkfs`）
- **THEN** 沙箱 SHALL 拒绝执行并返回失败结果

#### Scenario: 命令超时
- **WHEN** 命令超过配置的超时时间（默认 120s）
- **THEN** 沙箱 SHALL 强制终止进程并返回超时错误

#### Scenario: 权限级别升级
- **WHEN** PolicyEngine 检查一个需要 EXECUTE 级别的工具，但当前为 READ 级别
- **THEN** PolicyEngine SHALL 返回 PROMPT 判定，要求用户升级

#### Scenario: 基于频率的重新提示
- **WHEN** 某个工具在 `frequencyWindowSeconds` 内的使用次数超过 `maxFrequency`
- **THEN** PolicyEngine SHALL 返回 PROMPT 判定，防止静默过度使用

#### Scenario: 用户永久覆盖
- **WHEN** 用户对某个特定工具设置了永久 ALLOW 或 DENY
- **THEN** PolicyEngine SHALL 直接返回覆盖判定，无需进一步检查
