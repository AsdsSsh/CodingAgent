## ADDED 需求

### Requirement: 标准布局的 Go module 结构
系统 SHALL 使用 Go 标准项目布局，入口点在 `cmd/`，所有应用包在 `internal/`。

#### Scenario: 源码构建
- **WHEN** 执行 `go build ./cmd/coding-agent`
- **THEN** 系统 SHALL 产出一个名为 `coding-agent`（Windows 下为 `coding-agent.exe`）的单一可执行文件

#### Scenario: internal 包不可被外部导入
- **WHEN** 外部 Go module 尝试导入 `internal/` 下的任何包
- **THEN** Go 编译器 SHALL 拒绝该导入

### Requirement: 单二进制文件，零运行时依赖
编译出的二进制文件 SHALL 是自包含的，不需要 JRE、JDK、Maven 或任何其他运行时依赖。

#### Scenario: 二进制文件在目标平台上无需预装工具即可运行
- **WHEN** 二进制文件被复制到仅安装了操作系统的机器上
- **THEN** 二进制文件 SHALL 成功启动并运行 REPL

### Requirement: 跨平台编译
系统 SHALL 支持从任意开发平台交叉编译到 Windows (amd64)、Linux (amd64) 和 macOS (amd64/arm64)。

#### Scenario: 为所有目标交叉编译
- **WHEN** 执行 `make build-all` 或等效命令
- **THEN** 系统 SHALL 为所有支持的平台/架构组合产出二进制文件

#### Scenario: Windows 特定终端初始化
- **WHEN** 二进制文件在 Windows 上运行
- **THEN** 系统 SHALL 启用虚拟终端处理（`ENABLE_VIRTUAL_TERMINAL_PROCESSING`）以支持 ANSI 转义序列

### Requirement: 使用 Makefile 自动化构建
系统 SHALL 提供 Makefile，包含 build、test、lint 和交叉编译目标。

#### Scenario: 默认构建目标
- **WHEN** 执行 `make` 或 `make build`
- **THEN** 系统 SHALL 为当前平台编译二进制文件

#### Scenario: 测试目标
- **WHEN** 执行 `make test`
- **THEN** 系统 SHALL 运行所有单元测试并报告结果

#### Scenario: 清理目标
- **WHEN** 执行 `make clean`
- **THEN** 系统 SHALL 移除所有构建产物
