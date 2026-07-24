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
 * 读取文件内容，支持可选的行范围选择。
 * 输出带行号，上限 100K 字符。
 */
public class FileReadTool implements ToolBase {

    @Override
    public String name() { return "read"; }

    @Override
    public String description() {
        return "读取文件内容。对于大文件可通过 offset 和 limit 指定读取范围。";
    }

    @Override
    @SuppressWarnings("unchecked")
    public Map<String, Object> parameters() {
        return Map.of(
            "type", "object",
            "properties", Map.of(
                "file_path", Map.of("type", "string", "description", "文件的绝对或相对路径"),
                "offset", Map.of("type", "integer", "description", "起始行号（从 1 开始）"),
                "limit", Map.of("type", "integer", "description", "要读取的行数")
            ),
            "required", List.of("file_path")
        );
    }

    @Override
    public PermissionLevel requiredPermission() { return PermissionLevel.READ; }

    @Override
    public ToolResult execute(ToolContext ctx, Map<String, Object> args) {
        String filePath = (String) args.get("file_path");
        Path path = ctx.resolvePath(filePath);

        if (!Files.exists(path)) return ToolResult.fail("文件不存在: " + filePath);
        if (!Files.isRegularFile(path)) return ToolResult.fail("不是文件: " + filePath);

        try {
            List<String> lines = Files.readAllLines(path);
            int offset = args.containsKey("offset") ? ((Number) args.get("offset")).intValue() : 1;
            int limit = args.containsKey("limit") ? ((Number) args.get("limit")).intValue() : lines.size();

            if (offset < 1) offset = 1;
            int start = offset - 1;
            int end = Math.min(start + limit, lines.size());

            // 带行号的输出，类似 cat -n
            StringBuilder sb = new StringBuilder();
            for (int i = start; i < end; i++) {
                sb.append(String.format("%6d\t%s%n", i + 1, lines.get(i)));
            }

            String result = sb.toString();
            boolean truncated = result.length() > 100_000;
            if (truncated) {
                result = result.substring(0, 100_000) + "\n... [已截断]";
            }
            return new ToolResult(true, result, null, truncated);
        } catch (IOException e) {
            return ToolResult.fail("读取文件出错: " + e.getMessage());
        }
    }
}
