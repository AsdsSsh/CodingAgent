## Decisions

- Welcome: 在 renderViewport 检测空 messages，渲染 logo 和帮助文本
- Auto-scroll: 每次 SetContent 后调用 viewport.GotoBottom()
- Dividers: 追加 assistant 消息时自动插入分隔符
- Permission keys: modal 渲染添加按键标签
- Thinking state: statusBar 显示 "Thinking..." vs "Executing read..."
- Input echo: 用户消息加 `▌` 前缀和 bold 样式
