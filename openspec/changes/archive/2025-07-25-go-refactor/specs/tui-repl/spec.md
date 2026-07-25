## ADDED 需求

### Requirement: 全屏终端界面含 header、viewport 和输入区
系统 SHALL 提供基于 bubbletea 的全屏 TUI，包含 header 栏、可滚动输出 viewport、多行输入区和帮助栏。

#### Scenario: TUI 启动布局
- **WHEN** 应用启动
- **THEN** TUI SHALL 渲染 header（显示模型名、权限级别和 token 数）、对话历史 viewport、多行文本输入区和底部帮助栏（显示快捷键）

#### Scenario: 窗口大小变化自适应
- **WHEN** 终端窗口被调整大小
- **THEN** TUI SHALL 重新计算组件尺寸并适配新尺寸

### Requirement: 通过 Ctrl+Enter 提交多行任务
系统 SHALL 在 textarea 中接受多行输入，用户按 Ctrl+Enter 时提交。

#### Scenario: 单行任务提交
- **WHEN** 用户输入任务并按 Ctrl+Enter
- **THEN** 系统 SHALL 清空输入区、追加用户消息到 viewport、开始 agent 执行

#### Scenario: 多行任务输入
- **WHEN** 用户在 textarea 中输入多行并按 Ctrl+Enter
- **THEN** 系统 SHALL 将整个多行内容作为单个任务提交

### Requirement: 实时 agent 进度显示
系统 SHALL 实时显示 agent 执行进度，包括步骤编号、工具名称和 token 计数。

#### Scenario: 执行期间进度更新
- **WHEN** ReActLoop 发出 AgentState 更新
- **THEN** TUI SHALL 在状态行中更新当前步骤、工具名称和 token 计数

#### Scenario: 工具调用开始通知
- **WHEN** 工具执行开始
- **THEN** TUI SHALL 在 viewport 中追加工具调用条目及运行状态

#### Scenario: 工具调用完成通知
- **WHEN** 工具执行完成
- **THEN** TUI SHALL 更新 viewport 中对应条目的成功/失败状态

#### Scenario: 执行期间动画指示器
- **WHEN** agent 正在运行（进度更新之间）
- **THEN** TUI SHALL 在状态栏中显示动画指示器

### Requirement: 权限提示作为阻塞模态对话框
系统 SHALL 在工具需要权限升级时显示模态，阻塞 ReActLoop 直到用户响应。

#### Scenario: PROMPT 判定时出现权限模态
- **WHEN** PolicyEngine 返回 PROMPT 判定
- **THEN** TUI SHALL 渲染模态显示工具名、所需权限和原因，并 SHALL 阻塞 ReActLoop 直到用户选择 Allow 或 Deny

#### Scenario: 用户本次允许权限
- **WHEN** 用户选择 "Allow"
- **THEN** 系统 SHALL 向 ReActLoop 发送 ALLOW 响应、关闭模态、工具 SHALL 执行

#### Scenario: 用户拒绝权限
- **WHEN** 用户选择 "Deny"
- **THEN** 系统 SHALL 向 ReActLoop 发送 DENY 响应、关闭模态、追加错误观察

#### Scenario: 用户永久允许工具
- **WHEN** 用户选择 "Always Allow"
- **THEN** 系统 SHALL 在 PolicyEngine 中注册永久覆盖、发送 ALLOW 响应、关闭模态

#### Scenario: 权限模态超时
- **WHEN** 权限模态显示 120 秒无响应
- **THEN** 系统 SHALL 自动拒绝请求、关闭模态、解除 ReActLoop 阻塞

### Requirement: REPL 斜杠命令控制
系统 SHALL 支持斜杠命令用于模型切换、权限变更、帮助显示、历史查看和退出。

#### Scenario: 列出和切换模型
- **WHEN** 用户输入 `/model`（无参数）
- **THEN** 系统 SHALL 显示当前模型并列出所有内置预设
- **WHEN** 用户输入 `/model <name>`
- **THEN** 系统 SHALL 切换到指定模型并自动检测端点

#### Scenario: 切换权限级别
- **WHEN** 用户输入 `/permission execute`
- **THEN** 系统 SHALL 更新权限级别为 EXECUTE 并在 header 反映

#### Scenario: 显示帮助
- **WHEN** 用户输入 `/help`
- **THEN** 系统 SHALL 在 viewport 中显示可用命令和快捷键

#### Scenario: 退出 REPL
- **WHEN** 用户输入 `/exit` 或 agent 空闲时按 Ctrl+C
- **THEN** 系统 SHALL 退出 TUI，返回退出码 0

### Requirement: 对话历史滚动和导航
系统 SHALL 在可滚动 viewport 中维护对话历史，Tab 切换输入区和 viewport 焦点。

#### Scenario: 滚动对话历史
- **WHEN** viewport 获得焦点且用户按 Up/Down 或 PageUp/PageDown
- **THEN** viewport SHALL 滚动对话历史

#### Scenario: Tab 切换焦点
- **WHEN** 用户按 Tab
- **THEN** 焦点 SHALL 在输入区和 viewport 之间切换

#### Scenario: Ctrl+C 取消运行中的 agent
- **WHEN** agent 正在执行且用户按 Ctrl+C
- **THEN** 系统 SHALL 通过 context 取消当前 agent 执行并返回空闲状态
