## Why

Bash 命令执行偶尔永久卡死。根因是 `CombinedOutput()` 在两种场景下死锁：Windows 子进程孤儿（父进程被 kill 后子进程继承管道句柄不释放）、长输出填满 OS 管道缓冲区。必须改用手动管道读取 + 进程树 kill。

## What Changes

- `SubprocessSandbox.ExecuteCommand()` 重写：`cmd.Start()` + goroutine 读管道 + select 超时
- 新增 `killProcessTree()`：Windows 用 `taskkill /F /T /PID`，Unix 用 `syscall.Kill(-pid)`
- 输出大小上限从 100K 收紧到 500K（防止 OOM），超出截断

## Capabilities

New: `fix-bash-hang` — 修复 bash 命令执行死锁

## Impact

修改 `internal/sandbox/subprocess.go`，新增平台适配代码
