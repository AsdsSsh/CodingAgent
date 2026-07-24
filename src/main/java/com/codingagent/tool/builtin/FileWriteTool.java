package com.codingagent.tool.builtin;

import com.codingagent.sandbox.PermissionLevel;
import com.codingagent.tool.ToolBase;
import com.codingagent.tool.ToolContext;
import com.codingagent.tool.ToolResult;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;

/**
 * 创建或覆盖文件。需要 {@link PermissionLevel#WRITE} 权限。
 * 拒绝在工作区外写入文件。
 */
public class FileWriteTool implements ToolBase {

    @Override
    public String name() { return "write"; }

    @Override
    public String description() {
        return "创建新文件或覆盖现有文件，写入指定内容。";
    }

    @Override
    @SuppressWarnings("unchecked")
    public Map<String, Object> parameters() {
        return Map.of(
            "type", "object",
            "properties", Map.of(
                "file_path", Map.of("type", "string", "description", "要写入的文件路径"),
                "content", Map.of("type", "string", "description", "要写入文件的内容")
            ),
            "required", List.of("file_path", "content")
        );
    }

    @Override
    public PermissionLevel requiredPermission() { return PermissionLevel.WRITE; }

    @Override
    public ToolResult execute(ToolContext ctx, Map<String, Object> args) {
        String filePath = (String) args.get("file_path");
        String content = (String) args.get("content");
        Path path = ctx.resolvePath(filePath);

        // 路径穿越防护
        if (!ctx.isInWorkspace(path)) {
            return ToolResult.fail("拒绝访问: " + filePath + " 位于工作区之外");
        }

        try {
            Files.createDirectories(path.getParent());
            Files.writeString(path, content);
            return ToolResult.ok("文件已写入: " + filePath + " (" + content.length() + " 字节)");
        } catch (IOException e) {
            return ToolResult.fail("写入文件出错: " + e.getMessage());
        }
    }
}
