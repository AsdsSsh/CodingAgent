## Why

TUI 交互骨架已可用，但视觉体验粗糙——启动空白、帮助栏过时、对话无分隔、权限弹窗无按键提示。

## What Changes

- 启动欢迎页：logo + 模型信息 + 操作提示
- 帮助栏文案：Ctrl+Enter → Enter / Ctrl+S
- Viewport 自动滚底：新消息自动滚动
- 对话分隔线：每轮之间加 `───`
- 权限弹窗按键：`[Y] Allow [N] Deny [A] Always`
- 思考状态细化：区分 "Thinking" / "Executing tool" / "Done"
- 对话输入回显优化：用户消息加对话气泡风格

## Capabilities

New: `ui-polish` — TUI 视觉打磨

## Impact

修改 `internal/cli/view.go`、`internal/cli/model.go`、`internal/cli/update.go`、`internal/cli/messages.go`
