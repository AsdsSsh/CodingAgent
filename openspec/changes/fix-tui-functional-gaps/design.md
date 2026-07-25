## Context

TUI 有 4 个 bug blocking 核心交互。

## Goals / Non-Goals

**Goals:** 修复 listenProgress 自毁、权限弹窗死胡同、进度空白工具名、API Key 错误提示

**Non-Goals:** 不重构 Update 架构

## Decisions

### listenProgress: 移除超时

`listenProgress` 的 5s 超时返回 TickMsg 后不会被重新注册。移除超时，纯阻塞 channel 读取。spinner 由独立的 `tickLater` 驱动。

### 权限弹窗: 键盘快捷键

无法使用鼠标的 TUI 中，按键操作：
- `y` 或 `Enter` → Allow
- `n` 或 `Esc` → Deny  
- `a` → Always Allow

### emitProgress: 填充 CurrentTool

在 ReActLoop 执行工具时，先发一条带工具名的进度消息。

### API Key 错误: 全宽高亮提示

在 viewport 中显示带红色高亮的错误消息，引导用户设置环境变量。
