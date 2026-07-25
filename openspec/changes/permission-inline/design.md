## Decisions

- 内联框：在 renderViewport 中追加带 lipgloss 边框的权限块
- 阻断输入：pendingPermReq != nil 时不渲染 textarea，只显示 "Waiting for permission (Y/N/A)..."
- 结果记录：响应后追加 ChatMessage{Role: "permission", Content: "bash: ALLOWED (once)"}
- Ctrl+P：直接修改 config.Permissions.DefaultLevel + sandbox.Escalate()
- 颜色编码：用 map[PermissionLevel]lipgloss.Color
