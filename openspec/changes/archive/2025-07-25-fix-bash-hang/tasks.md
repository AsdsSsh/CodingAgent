## 1. 重写沙箱执行

- [x] 1.1 将 CombinedOutput() 替换为手动管道读取 + goroutine + select 模式
- [x] 1.2 超时时杀死整个进程树（Windows + Unix）
- [x] 1.3 输出上限 500K，使用 io.CopyN

## 2. 平台特定进程终止

- [x] 2.1 Windows：taskkill /F /T /PID
- [x] 2.2 Unix：syscall.Kill(-pid, SIGKILL)

## 3. 测试

- [x] 3.1 测试超时终止挂起命令
- [x] 3.2 测试正常命令成功完成
- [x] 3.3 测试大输出被截断而非挂起
