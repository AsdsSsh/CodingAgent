package com.codingagent.tool.builtin;

import com.codingagent.sandbox.PermissionLevel;
import com.codingagent.tool.ToolBase;
import com.codingagent.tool.ToolContext;
import com.codingagent.tool.ToolResult;

import java.util.List;
import java.util.Map;

/**
 * 在沙箱内执行 Shell 命令。
 * 需要 {@link PermissionLevel#EXECUTE} 权限——比文件写入高一级。
 * 实际执行委托给沙箱，由沙箱负责超时和黑名单控制。
 */
public class BashTool implements ToolBase {

    @Override
    public String name() { return "bash"; }

    @Override
    public String description() {
        return "在工作目录中执行 bash 命令。命令有超时和输出大小限制。";
    }

    @Override
    @SuppressWarnings("unchecked")
    public Map<String, Object> parameters() {
        return Map.of(
            "type", "object",
            "properties", Map.of(
                "command", Map.of("type", "string", "description", "要执行的 bash 命令"),
                "description", Map.of("type", "string", "description", "该命令的简短说明")
            ),
            "required", List.of("command")
        );
    }

    @Override
    public PermissionLevel requiredPermission() { return PermissionLevel.EXECUTE; }

    @Override
    public ToolResult execute(ToolContext ctx, Map<String, Object> args) {
        String command = (String) args.get("command");
        return ctx.sandbox().executeCommand(command);
    }
}
