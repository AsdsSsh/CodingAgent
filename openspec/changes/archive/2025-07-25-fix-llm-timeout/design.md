## 方案

RunOptions 新增 `Ctx context.Context`。ReActLoop 中每次 Chat 调用创建 30s 超时子 context。

handleTaskSubmit 创建 `ctx, cancel := context.WithCancel(context.Background())`，Ctrl+C 调用 cancel() 立即中断 HTTP。
