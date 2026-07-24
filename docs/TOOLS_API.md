# 工具 API 参考

## 概述

CodingAgent 提供 6 个内置工具，供 LLM 在 ReAct 循环中调用。每个工具有对应的最低权限级别要求。

---

## 工具列表

### 1. `search` — 文件搜索

**权限**: `READ`

**功能**: 按 glob 模式搜索文件名，或按正则表达式搜索文件内容。

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `pattern` | string | 是 | glob 模式（如 `**/*.java`）或 grep 的正则表达式 |
| `mode` | string | 是 | `"glob"` 按文件名查找，`"grep"` 搜索文件内容 |
| `path` | string | 否 | 搜索的目录（默认为工作区根目录） |

**示例**:
```json
{"pattern": "**/*.java", "mode": "glob"}
{"pattern": "class ReActLoop", "mode": "grep", "path": "src/"}
```

---

### 2. `read` — 文件读取

**权限**: `READ`

**功能**: 读取文件内容。对于大文件可通过 offset 和 limit 指定读取范围。

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_path` | string | 是 | 文件的绝对或相对路径 |
| `offset` | integer | 否 | 起始行号（从 1 开始） |
| `limit` | integer | 否 | 要读取的行数 |

**示例**:
```json
{"file_path": "src/main/java/com/codingagent/CodingAgentApp.java"}
{"file_path": "pom.xml", "offset": 1, "limit": 30}
```

---

### 3. `write` — 文件写入

**权限**: `WRITE`

**功能**: 创建新文件或覆盖现有文件，写入指定内容。

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_path` | string | 是 | 要写入的文件路径 |
| `content` | string | 是 | 要写入文件的内容 |

**示例**:
```json
{"file_path": "docs/README.md", "content": "# My Project\n\n..."}
```

---

### 4. `edit` — 文件编辑

**权限**: `WRITE`

**功能**: 在文件中执行精确的字符串替换。若 `old_string` 未找到或出现多次则操作失败。

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_path` | string | 是 | 要编辑的文件路径 |
| `old_string` | string | 是 | 要查找并替换的文本 |
| `new_string` | string | 是 | 替换后的文本 |

**示例**:
```json
{
  "file_path": "src/App.java",
  "old_string": "System.out.println(\"Hello\");",
  "new_string": "System.out.println(\"Hello, World!\");"
}
```

**注意**: 必须精确匹配（包括空格和缩进）。若匹配不唯一则替换失败。

---

### 5. `bash` — Bash 命令执行

**权限**: `EXECUTE`

**功能**: 在工作目录中执行 Bash 命令。命令有超时和输出大小限制。受沙箱安全策略约束。

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `command` | string | 是 | 要执行的 Bash 命令 |
| `description` | string | 是 | 该命令的简短说明（用于审计） |

**安全限制**:
- 默认超时: 120 秒
- 危险命令自动拦截: `rm -rf /`、`sudo`、`dd if=`、`mkfs`、fork bomb、`shutdown`、`reboot`
- 频率限制: 300 秒内最多 10 次（可配置）

**示例**:
```json
{"command": "mvn clean compile", "description": "编译项目"}
{"command": "ls -la src/main/java/", "description": "列出 Java 源文件"}
```

---

### 6. `web_fetch` — 网页抓取

**权限**: `READ`

**功能**: 获取 URL 的内容并以文本形式返回。适用于阅读文档或 API 响应。

**参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `url` | string | 是 | 要获取的 URL |
| `prompt` | string | 否 | 要从页面中提取的信息内容 |

**示例**:
```json
{"url": "https://docs.oracle.com/en/java/javase/21/docs/api/"}
{"url": "https://api.github.com/repos/user/repo", "prompt": "提取仓库信息"}
```

---

## 工具执行流程

```
LLM 返回 tool_calls
        │
        ▼
  ToolRegistry.resolve(name)
        │
        ▼
  PolicyEngine.check(name, requiredPermission)
        │
   ┌────┼────┐
   ▼    ▼    ▼
 ALLOW PROMPT DENY
   │    │     │
   ▼    ▼     ▼
execute  返回错误  返回错误
   │
   ▼
ToolResult.formatForLlm()
   │
   ▼
追加到 messages: [assistant(tool_call), tool(observation)]
```

---

## 扩展：自定义工具

```java
public class MyTool extends ToolBase {
    @Override
    public String name() { return "my_tool"; }

    @Override
    public String description() { return "我的自定义工具"; }

    @Override
    public Map<String, Object> parameters() {
        return Map.of(
            "type", "object",
            "properties", Map.of(
                "param1", Map.of("type", "string", "description", "参数1")
            ),
            "required", List.of("param1")
        );
    }

    @Override
    public PermissionLevel requiredPermission() { return PermissionLevel.READ; }

    @Override
    public ToolResult execute(ToolContext ctx, Map<String, Object> args) {
        // 实现工具逻辑
        return ToolResult.success("执行结果");
    }
}

// 注册: toolRegistry.register(new MyTool());
```
