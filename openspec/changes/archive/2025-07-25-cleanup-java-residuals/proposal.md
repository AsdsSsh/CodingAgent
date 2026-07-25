## Why

Go 重构完成后，Java 源码 (`src/`)、Maven 构建文件 (`pom.xml`)、编译产物 (`target/`)、IDE 配置 (`.idea/`) 和旧文档 (`docs/`) 等残留文件不再需要，增加仓库噪音。清理后项目更简洁，只保留 Go 项目必需的文件。

## What Changes

- 删除 `pom.xml` — Maven 项目定义
- 删除 `src/` — 37 个 Java 源文件（已全部翻译为 Go）
- 删除 `target/` — Maven 编译产物（`.jar`、`.class`）
- 删除 `.idea/` — IntelliJ IDEA 项目配置
- 删除 `docs/` — Java 版文档（ARCHITECTURE.md、CONFIGURATION.md 等）
- 删除 `.github/modernize/` — Java 升级 CI 工作流
- 删除 `config/permissions.yaml` — 已合并到 `default_config.yaml`
- 删除 `config/default_config.yaml.bak` — 重构备份
- 删除 `./-p` — 幽灵空目录
- 将 `coding-agent.exe` 加入 `.gitignore`

## Capabilities

### New Capabilities

- `repo-cleanup`: 仓库清理 — 移除 Java 重构残留文件和目录，添加 Go 编译产物到 gitignore。

### Modified Capabilities

<!-- 无 -->

## Impact

- 删除约 10 个路径，释放 ~50MB 磁盘空间（主要是 `target/` 下的 jar）
- `.gitignore` 新增 `coding-agent.exe` 条目
- Go 项目源码、配置、文档不受影响
