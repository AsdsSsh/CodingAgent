## ADDED 需求

### Requirement: 通过 YAML 配置 MCP 服务器
系统 SHALL 支持在 YAML 配置文件中配置 MCP（模型上下文协议）服务器，支持 stdio 和 HTTP 传输方式。

#### Scenario: Stdio 传输配置
- **WHEN** MCP 服务器配置了 `transport: stdio`、`command` 和可选的 `args`
- **THEN** 系统 SHALL 将命令作为子进程启动，通过 stdin/stdout 使用 MCP 协议通信

#### Scenario: HTTP 传输配置
- **WHEN** MCP 服务器配置了 `transport: http` 和 `url`
- **THEN** 系统 SHALL 使用 HTTP 流式传输连接该 URL

#### Scenario: MCP 服务器进程环境变量
- **WHEN** MCP 服务器配置包含 `env` 键值对
- **THEN** 系统 SHALL 在服务器子进程上设置这些环境变量

### Requirement: MCP 工具与 ToolRegistry 集成
系统 SHALL 将 MCP 提供的工具注册到现有 ToolRegistry 中，使其与内置工具一起对 LLM 可用。

#### Scenario: MCP 工具出现在 LLM 工具列表中
- **WHEN** MCP 服务器连接成功且其工具已注册
- **THEN** `ToLLMFormats()` 方法 SHALL 同时包含内置工具和 MCP 工具

#### Scenario: MCP 工具执行委托给 MCP 服务器
- **WHEN** LLM 调用一个 MCP 提供的工具
- **THEN** 系统 SHALL 将调用路由到对应的 MCP 服务器，并将响应作为工具结果观察返回

#### Scenario: MCP 工具名与内置工具冲突
- **WHEN** MCP 服务器提供与内置工具同名的工具
- **THEN** 系统 SHALL 记录警告并跳过重复的 MCP 工具注册

### Requirement: MCP 服务器延迟连接
系统 SHALL 将 MCP 服务器连接推迟到该服务器的第一个工具被实际调用时，以最小化启动时间。

#### Scenario: 启动时服务器未连接
- **WHEN** agent 启动且配置了 MCP 服务器
- **THEN** 系统 SHALL 缓存服务器配置但 SHALL NOT 建立连接，直到该服务器的工具被调用

#### Scenario: 首次工具调用时建立连接
- **WHEN** LLM 请求一个尚未连接的 MCP 服务器的工具
- **THEN** 系统 SHALL 建立连接、获取工具列表并执行请求的工具

#### Scenario: 连接失败产生错误观察
- **WHEN** MCP 服务器连接失败或子进程异常退出
- **THEN** 系统 SHALL 向 LLM 返回错误观察，允许其尝试替代方案
