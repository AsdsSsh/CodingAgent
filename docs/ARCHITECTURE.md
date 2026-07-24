# CodingAgent 架构设计文档

## 1. 概述

CodingAgent 是一个轻量级 AI 编程助手，采用 **ReAct（Reasoning + Acting）** 架构。用户通过命令行输入编程任务，Agent 在 ReAct 循环中交替进行推理（Thought）和行动（Action），通过调用文件读写、搜索、Bash 执行、网页抓取等工具完成任务。

### 核心设计原则

- **安全沙箱**：所有工具调用必须经过权限策略引擎的验证
- **提供商无关**：通过抽象接口支持多种 LLM（Anthropic / OpenAI 兼容）
- **不可变配置**：使用 Java Record 实现配置的不可变性
- **分层覆盖**：配置按优先级逐层合并
- **事件驱动**：Agent 状态变更通过 EventBus 传递，UI 层解耦订阅

---

## 2. 五层架构

```
┌──────────────────────────────────────────────────┐
│                   Layer 5: CLI                    │
│         CliApp · Repl · Formatter                 │
│   命令行解析 · 交互式REPL · 结果格式化              │
├──────────────────────────────────────────────────┤
│                  Layer 4: Config                  │
│     AgentConfig · ModelConfig · PermissionConfig  │
│   YAML加载 · 分层合并 · 环境变量覆盖 · CLI覆盖      │
├──────────────────────────────────────────────────┤
│                Layer 3: Orchestrator              │
│   配置加载 → 密钥解析 → 工具注册 → 沙箱初始化        │
│         系统提示构建 → 委托 ReActLoop              │
├──────────────────────────────────────────────────┤
│                 Layer 2: ReAct Loop               │
│   Thought(LLM推理) → Action(工具调用)               │
│        → Observation(观察结果) → 循环               │
│   权限验证 · 步数上限 · 异常恢复                    │
├──────────────────────────────────────────────────┤
│                  Layer 1: Foundation              │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐   │
│  │ LLM 层   │  │ 工具层    │  │  沙箱层       │   │
│  │Provider  │  │Registry  │  │ Sandbox+      │   │
│  │Factory   │  │6个内置工具│  │ PolicyEngine  │   │
│  └──────────┘  └──────────┘  └──────────────┘   │
└──────────────────────────────────────────────────┘
```

---

## 3. ReAct 循环详解

### 3.1 算法

```
输入: task (用户任务), systemPrompt, memories, sandbox, workspace
输出: AgentResult (最终回答 + 统计信息)

1. messages = [system(systemPrompt), ...memories, user(task)]
2. step = 0
3. while step < maxIterations:
   a. step++
   b. emit(AgentState) → 通知 UI 当前状态
   c. response = llm.chat(messages, tools)  ← Thought (推理)
   d. if response 无工具调用:
      - 若 isFinalAnswer 或 stopReason="end_turn": return AgentResult
      - 否则: 追加 assistant 消息, continue
   e. for each toolCall in response.toolCalls:  ← Action (行动)
      i.   若工具未注册: 追加错误 observation
      ii.  decision = policyEngine.check(tool, requiredPermission)
      iii. 若 DENY/PROMPT: 追加错误 observation
      iv.  若 ALLOW: execute(tool, args) → ToolResult → observation  ← Observation
      v.   追加 [assistant_with_tool_call, tool_result] 消息对
4. 达到上限: 追加 "请给出最终回答", 再次调用 LLM, 返回结果
```

### 3.2 关键安全约束

| 约束 | 实现方式 |
|------|---------|
| 步数上限 | `maxIterations`（默认 50），达到后强制结束 |
| 权限拒绝 | 返回错误 observation，Agent 可自主恢复而非崩溃 |
| 未知工具 | 返回包含可用工具列表的错误信息 |
| 消息对完整性 | `tool_call` 和 `tool_result` 始终相邻，永不拆分 |

### 3.3 EventBus 事件流

```
Orchestrator.run()
  └─→ ReActLoop.run()
        └─→ eventBus.emit(AgentState)  ← 每步发出
              └─→ CLI/UI 订阅者接收 → 实时渲染状态
```

---

## 4. 工具系统设计

### 4.1 工具接口

```java
public abstract class ToolBase {
    public abstract String name();           // 工具唯一标识
    public abstract String description();    // 给 LLM 看的描述
    public abstract Map<String, Object> parameters(); // JSON Schema 参数定义
    public abstract PermissionLevel requiredPermission(); // 所需最低权限
    public abstract ToolResult execute(ToolContext ctx, Map<String, Object> args);
}
```

### 4.2 工具注册与 LLM 格式导出

`ToolRegistry` 负责：
- 按名称注册和查找工具
- 检查工具是否存在（容错）
- 导出为 LLM function calling 兼容格式：`[{name, description, parameters}, ...]`

### 4.3 工具上下文 (ToolContext)

```java
public record ToolContext(Path workspace, SandboxBase sandbox) {}
```

提供工作区路径和沙箱引用，工具通过沙箱执行受控操作。

---

## 5. 权限系统设计

### 5.1 权限级别

```
NONE ──→ READ ──→ WRITE ──→ EXECUTE ──→ DANGEROUS
 (无)    (读)     (写)       (执行)       (危险)
```

每个级别包含前一级别的所有权限：
- **NONE**: 无任何操作
- **READ**: 文件读取、搜索、网页抓取
- **WRITE**: + 文件写入、编辑
- **EXECUTE**: + Bash 命令执行
- **DANGEROUS**: 所有操作，包括被拦截的危险命令

### 5.2 权限决策流程

```
PolicyEngine.check(toolName, requiredPermission)
          │
          ▼
    tool_overrides 中有配置？
    ├── YES → 直接返回 ALLOW 或 DENY
    └── NO
          │
          ▼
    频率限制检查
    ├── 超出频率 → 返回 PROMPT（需用户确认）
    └── 未超出
          │
          ▼
    currentLevel >= requiredPermission？
    ├── YES → 返回 ALLOW
    └── NO  → 返回 PROMPT
```

### 5.3 沙箱安全机制

`SubprocessSandbox` 实现：
- **命令黑名单**：默认拦截 `rm -rf /`、`sudo`、`dd if=`、`mkfs`、fork bomb、`shutdown`、`reboot`
- **超时控制**：可配置的命令执行超时（默认 120 秒）
- **频率限制**：时间窗口内（默认 300 秒）最多 N 次（默认 10 次），超出后重新弹窗确认

---

## 6. 配置系统设计

### 6.1 配置加载链

```
硬编码默认值
    ↓ 合并
~/.coding-agent/config.yaml     (全局)
    ↓ 合并
.coding-agent/config.yaml       (项目级)
    ↓ 覆盖
环境变量 (ANTHROPIC_API_KEY 等)
    ↓ 覆盖
CLI 参数 (-p, -m, -k, -t, -w)
    ↓
最终 AgentConfig
```

### 6.2 API 密钥解析优先级

```
1. 配置文件中的 api_key 字段
2. 配置文件中的 api_key_env 指定的环境变量
3. provider 默认环境变量:
   - anthropic → ANTHROPIC_API_KEY
   - openai    → OPENAI_API_KEY
   - 其他      → OPENAI_API_KEY 或 ANTHROPIC_API_KEY
```

### 6.3 不可变配置模式

所有配置类使用 Java `record` 类型，通过 `with*` 方法返回修改后的副本：

```java
config.withWorkspace("/new/path")
      .withModelOverride(newModelConfig)
      .withPermissionLevel(PermissionLevel.WRITE)
```

---

## 7. 数据流示意

```
用户输入 "修复 src/App.java 中的空指针异常"
         │
         ▼
    Repl.start()
         │
         ▼
    Orchestrator.run(task)
         │
         ├── 构建系统提示
         ├── 初始 messages: [system, user(task)]
         │
         ▼
    ReActLoop.run()
         │
    ┌────┴─────────────────────────────────────┐
    │ Step 1: LLM → "我需要先读取文件"           │
    │         → tool_call: read("src/App.java") │
    │         → PolicyEngine: READ ✓            │
    │         → execute → observation: 文件内容   │
    ├──────────────────────────────────────────┤
    │ Step 2: LLM → "我发现第42行有空指针风险"    │
    │         → tool_call: edit(修复代码)         │
    │         → PolicyEngine: WRITE ✓           │
    │         → execute → observation: 编辑成功   │
    ├──────────────────────────────────────────┤
    │ Step 3: LLM → "任务完成，已修复空指针异常"   │
    │         → 无工具调用, end_turn             │
    │         → return AgentResult              │
    └──────────────────────────────────────────┘
         │
         ▼
    Formatter.format(result)
         │
         ▼
    "已修复 src/App.java 第42行的空指针异常..."
    Steps: 3 | Tools called: 2
```

---

## 8. 设计模式总结

| 模式 | 应用位置 |
|------|---------|
| **工厂模式** | `ProviderFactory` 根据配置创建 LLM 提供商实例 |
| **策略模式** | `LlmProvider` 接口，Anthropic/OpenAI 不同实现 |
| **注册表模式** | `ToolRegistry` 集中管理工具 |
| **观察者模式** | `EventBus` 发布/订阅 Agent 状态变更 |
| **模板方法** | `ToolBase` 定义工具执行骨架，子类实现具体逻辑 |
| **不可变对象** | 所有 Config 类使用 Java Record |
| **建造者模式** | Config 的 `with*` 方法链式构建新配置 |
