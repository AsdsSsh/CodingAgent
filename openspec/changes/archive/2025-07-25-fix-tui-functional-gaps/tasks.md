## 1. 修复 listenProgress 超时自毁

- [x] 1.1 移除 `listenProgress` 的 5s 超时，改用纯阻塞 channel 读取

## 2. 修复 emitProgress 缺失 CurrentTool

- [x] 2.1 在 ReActLoop 工具执行循环中添加工具名发射

## 3. 修复权限弹窗死胡同

- [x] 3.1 为权限模态添加键盘处理：y/Enter=Allow、n/Esc=Deny、a=Always Allow

## 4. 改善 API Key 错误可见性

- [x] 4.1 将初始化错误改为红色横幅样式显示
