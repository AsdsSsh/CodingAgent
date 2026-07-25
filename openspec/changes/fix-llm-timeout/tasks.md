## 1. ReActLoop 支持可取消 context

- [ ] 1.1 RunOptions 新增 Ctx context.Context 字段
- [ ] 1.2 每次 Chat 调用用 context.WithTimeout(runOpts.Ctx, 30s)

## 2. TUI 侧创建可取消 context

- [ ] 2.1 handleTaskSubmit 中创建 context.WithCancel
- [ ] 2.2 Ctrl+C 时调用 cancel()，立即中断 HTTP
- [ ] 2.3 ReActLoop.Run 需要 ctx 时使用 runOpts.Ctx 而非 context.Background()
