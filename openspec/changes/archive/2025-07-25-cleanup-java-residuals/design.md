## Context

Go 重构 (`go-refactor`) 完成后，仓库中残留 Java 项目文件。Java 源码和 Maven 构建已完全被 Go 替代，这些文件不再有存在价值。

## Goals / Non-Goals

**Goals:**
- 删除所有 Java 构建和源码文件
- 删除 IDE 配置（`.idea/`）
- 删除过时文档
- 清理零散杂物（备份文件、幽灵目录）
- 将 Go 编译产物加入 `.gitignore`

**Non-Goals:**
- 不修改任何 Go 源代码
- 不修改配置文件内容（仅删除冗余的 `permissions.yaml`）
- 不修改 `.claude/` 或 `openspec/` 目录

## Decisions

**删除策略：直接 `rm -rf`**，不使用 `git rm`。原因：`.idea/` 和 `target/` 可能已在 `.gitignore` 中，`git rm` 会失败。用 `rm -rf` 删除后，Git 会自然检测到文件消失。

**不归档旧代码**：Java 源码在 git 历史中可追溯，不需要保留 `_archive` 副本。

## Risks / Trade-offs

- [误删 Go 文件] → 清理列表精确限定，不触碰 `cmd/`、`internal/`、`go.*`、`Makefile`
- [`.idea/` 是用户本地配置] → 确认 `.idea/` 不在版本控制中，删除不影响其他开发者
