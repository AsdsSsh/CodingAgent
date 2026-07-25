## Context

当前权限系统在 TUI 层使用 bubbletea 的 `tea.Cmd` 模式监听 Agent 发送的权限请求。`listenPermission()` 是一个一次性命令，从 `permReqChan` 读取一条权限请求后返回 `PermissionPromptMsg`。

正常模式（以 `listenProgress` 为例）：`AgentProgressMsg` 处理完后重新调用 `m.listenProgress()` 启动下一次监听。但权限响应处理（Y/N/A 按键分支）遗漏了这个重新启动步骤，只返回了 `m.listenProgress()`。

当用户选择「仅允许一次」（Y 键）时，Agent 端不会设置永久覆盖（`AlwaysAllow=false`），因此后续同类型工具调用仍会触发权限请求。但第二次请求到达时 TUI 端已无人监听 `permReqChan`，Agent 在 `react.go:165` 永久阻塞于 `resp := <-opts.PermissionResp`。

## Goals / Non-Goals

**Goals:**
- 修复 Y/N/A 三个权限响应分支，使其在处理完当前权限请求后重新启动 `listenPermission()`
- 确保多次连续的权限请求都能正常弹出提示

**Non-Goals:**
- 不改变 Agent 端（react.go）的权限检查逻辑
- 不改变 PolicyEngine 的覆盖/频率控制机制
- 不引入新的权限模式或 UI 交互

## Decisions

### 修复方案：在三个权限响应分支中补充 `listenPermission()`

**选择**: 将 Y/N/A 三个分支的返回值从 `return m, m.listenProgress()` 改为 `return m, tea.Batch(m.listenProgress(), m.listenPermission())`

**替代方案（不采用）**: 将 `listenPermission()` 改为无限循环模式

将 `listenPermission()` 内部改为 for 循环持续读取 channel 看似更简洁，但 bubbletea 的 Cmd 模式鼓励 one-shot + 重新调度。改为循环会阻塞一个 goroutine，且与其他 Cmd（如 `listenProgress`）的风格不一致。

**影响范围**: 仅 `internal/cli/update.go` 第 42、54、66 行，三处修改模式完全一致。

## Risks / Trade-offs

- **风险**: 如果 Agent 在 TUI 关闭后仍在运行，channel 写入可能 panic。但这已经是已有的风险（与本次修改无关），现有的 channel buffer 设计（cap=5）已提供基本缓冲。
- **权衡**: 每次权限响应后多启动一个 `listenPermission()` goroutine。但 `listenPermission()` 是阻塞读取 channel 的轻量 goroutine，资源开销可忽略不计。
