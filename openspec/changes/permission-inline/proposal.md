## Why

权限弹窗用全屏 Modal 打断对话流且无法主动切换权限。改为对话内联提示 + Ctrl+P 循环切换 + Header 彩色状态。

## What Changes

- 被动权限请求：对话区内联边框块，输入区阻断，结果留在对话历史
- Ctrl+P 循环切换权限：READ→WRITE→EXECUTE→DANGEROUS，实时更新 sandbox
- Header 权限颜色编码：灰/黄/橙/红 + 状态栏实时显示
- /permission 无参数时行为改为循环切换

## Capabilities

New: `permission-inline` — 权限系统交互重构

## Impact

修改 model.go / view.go / update.go / styles.go
