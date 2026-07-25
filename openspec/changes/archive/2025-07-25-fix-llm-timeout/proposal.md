## 问题

LLM HTTP 请求超过 60s 后仍不返回，Ctrl+C 无法中断。根因：ReActLoop 使用 `context.Background()` 传给 HTTP 请求，无 deadline 也无法取消。HTTP Client 的 60s timeout 在某些网络状态下不触发。

## 修改

- ReActLoop.RunOptions 新增 `Ctx context.Context`
- `handleTaskSubmit` 创建 `context.WithCancel`，Ctrl+C 时 cancel
- 每次 LLM 调用创建 `context.WithTimeout(ctx, 30s)` 子 context
- HTTP 请求被 cancel 时立即返回错误而非挂起

## 影响

修改 `internal/agent/react.go`、`internal/cli/update.go`
