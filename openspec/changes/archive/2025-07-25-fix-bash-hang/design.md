## Context

`CombinedOutput()` 内部调用 `cmd.Start()` 然后阻塞读 `stdoutPipe`。当管道的写端（子进程）不被关闭时，读端永久阻塞。

两个死锁场景:
1. Windows: ctx 取消 → kill cmd.exe → 子进程继承管道句柄 → 句柄不关 → Go 永久阻塞
2. 通用: 命令输出 > OS 管道缓冲区(64KB) → 子进程 write 阻塞 → 管道满但没读完 → 死锁

## Decisions

### 手动管道读取 + 超时 select

```go
stdout, _ := cmd.StdoutPipe()
stderr, _ := cmd.StderrPipe()
cmd.Start()

var buf bytes.Buffer
go func() { io.Copy(&buf, stdout) }()  // 独立 goroutine 读管道
go func() { io.Copy(&buf, stderr) }()

select {
case <-ctx.Done():
    killProcessTree(cmd.Process.Pid)
    return timeout
case <-done:
    return output
}
```

### 进程树 Kill

- Windows: `taskkill /F /T /PID <pid>` — 杀进程树
- Unix: `syscall.Kill(-pid, syscall.SIGKILL)` — 杀进程组

### 输出上限 500K

`io.CopyN` 限制最大读取量，防止单个命令撑爆内存。

## Risks

- 管道读取 goroutine 可能泄漏: select 超时后 goroutine 仍在阻塞读
  → 解决: 先 kill 进程树（关闭写端），goroutine 自然退出
