## Why

当用户在权限提示中选择「仅允许一次」(Only Once) 后，第一个 Bash 工具可以正常执行，但后续需要权限的工具调用会导致整个 Agent 永久挂死。原因是 TUI 层的 `listenPermission()` 是一个一次性命令，在消费完第一个权限请求后没有被重新启动，导致 Agent 端发送的第二个权限请求无人接收，Agent 永久阻塞在等待响应上。

## What Changes

- 修复 `update.go` 中权限响应处理（Y/N/A 三个分支），在返回时同时重新启动 `listenPermission()` 命令，确保后续权限请求能被持续监听
- 权限监听命令与进度监听命令 (`listenProgress()`) 采用相同的持续监听模式

## Capabilities

### New Capabilities

无 — 这是一个 bug 修复，不引入新能力。

### Modified Capabilities

无 — 不改变任何规格级别的行为要求，仅修复现有实现中的缺陷。

## Impact

- **受影响代码**: `internal/cli/update.go` 第 42、54、66 行 — 三处权限按键处理分支的返回值
- **不受影响**: `internal/agent/react.go`、`internal/sandbox/policy.go` — Agent 端权限请求/响应逻辑正确，无需修改
