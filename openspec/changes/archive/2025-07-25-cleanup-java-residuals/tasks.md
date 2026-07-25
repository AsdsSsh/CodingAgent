## 1. 移除 Java 构建和源码文件

- [x] 1.1 删除 `pom.xml` — Maven 项目定义
- [x] 1.2 删除 `src/` — 全部 37 个 Java 源文件
- [x] 1.3 删除 `target/` — Maven 构建产物（`.jar`、`.class` 等）

## 2. 移除 IDE 和 CI 配置

- [x] 2.1 删除 `.idea/` — IntelliJ IDEA 项目文件
- [x] 2.2 删除 `.github/modernize/` — Java 升级 CI 工作流

## 3. 移除旧文档

- [x] 3.1 删除 `docs/` — Java 版文档

## 4. 移除冗余配置和零散文件

- [x] 4.1 删除 `config/permissions.yaml` — 已合并到 `default_config.yaml`
- [x] 4.2 删除 `config/default_config.yaml.bak` — 重构备份
- [x] 4.3 删除 `./-p` — 幽灵空目录

## 5. 将 Go 编译产物加入 gitignore

- [x] 5.1 将 `coding-agent.exe` 加入 `.gitignore`

## 6. 验证

- [x] 6.1 验证清理后 `go build ./...` 仍然成功
- [x] 6.2 验证清理后 `go test ./...` 仍然通过
- [x] 6.3 验证 `config/default_config.yaml` 完好且可加载
