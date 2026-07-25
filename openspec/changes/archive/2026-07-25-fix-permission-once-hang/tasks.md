## 1. 修复权限响应分支

- [x] 1.1 修复「仅允许一次」(Y/Enter) 分支：`update.go:42` 将 `return m, m.listenProgress()` 改为 `return m, tea.Batch(m.listenProgress(), m.listenPermission())`
- [x] 1.2 修复「拒绝」(N/Esc) 分支：`update.go:54` 将 `return m, m.listenProgress()` 改为 `return m, tea.Batch(m.listenProgress(), m.listenPermission())`
- [x] 1.3 修复「始终允许」(A) 分支：`update.go:66` 将 `return m, m.listenProgress()` 改为 `return m, tea.Batch(m.listenProgress(), m.listenPermission())`

## 2. 验证

- [x] 2.1 编译项目确保无编译错误：`go build ./...`
- [x] 2.2 运行现有测试确保无回归：`go test ./...`
