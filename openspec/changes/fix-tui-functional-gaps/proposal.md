## Why

TUI 审计发现 4 个功能缺口：API Key 错误不可见、进度监听器自毁、工具名不显示、权限弹窗无法操作。这些导致核心交互链路断裂。

## What Changes

- 修复 `listenProgress` 超时后不再重新监听（移除 5s 超时，改为纯阻塞等待）
- 修复 `emitProgress` 不填 `CurrentTool` 字段
- 修复权限弹窗缺少键盘操作（Enter=Allow, Esc=Deny, Ctrl+A=Always Allow）
- API Key 初始化失败时不再静默显示，改为覆盖式错误提示

## Capabilities

### New Capabilities

- `tui-fixes`: TUI 功能缺口修复，涉及进度监听、权限弹窗、错误提示、工具名显示。

## Impact

- 修改 `internal/cli/update.go` — listenProgress、权限弹窗按键、错误提示
- 修改 `internal/agent/react.go` — emitProgress 填充 CurrentTool
