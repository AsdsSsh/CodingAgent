## 1. Rewrite sandbox execution

- [x] 1.1 Replace CombinedOutput() with manual pipe read + goroutine + select pattern
- [x] 1.2 Add process tree kill on timeout (Windows + Unix)
- [x] 1.3 Cap output at 500K with io.CopyN

## 2. Platform-specific process kill

- [x] 2.1 Windows: taskkill /F /T /PID
- [x] 2.2 Unix: syscall.Kill(-pid, SIGKILL)

## 3. Testing

- [x] 3.1 Test timeout kills hanging command
- [x] 3.2 Test normal command completes successfully
- [x] 3.3 Test large output is truncated not hung
