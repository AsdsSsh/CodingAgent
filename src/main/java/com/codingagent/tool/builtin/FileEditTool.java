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
 * 精确编辑：通过精确字符串替换修改文件。
 *
 * <p><b>安全约束：</b>若 old_string 出现 0 次或超过 1 次，编辑会失败。
 * 这迫使 LLM 提供足够的上下文以唯一标识编辑位置，
 * 防止因模糊匹配导致意外破坏。</p>
 */
public class FileEditTool implements ToolBase {

    @Override
    public String name() { return "edit"; }

    @Override
    public String description() {
        return "在文件中执行精确的字符串替换。" +
               "查找 old_string 并将其替换为 new_string。" +
               "若 old_string 未找到或出现多次则操作失败。";
    }

    @Override
    @SuppressWarnings("unchecked")
    public Map<String, Object> parameters() {
        return Map.of(
            "type", "object",
            "properties", Map.of(
                "file_path", Map.of("type", "string", "description", "要编辑的文件路径"),
                "old_string", Map.of("type", "string", "description", "要查找并替换的文本"),
                "new_string", Map.of("type", "string", "description", "替换后的文本")
            ),
            "required", List.of("file_path", "old_string", "new_string")
        );
    }

    @Override
    public PermissionLevel requiredPermission() { return PermissionLevel.WRITE; }

    @Override
    public ToolResult execute(ToolContext ctx, Map<String, Object> args) {
        String filePath = (String) args.get("file_path");
        String oldString = (String) args.get("old_string");
        String newString = (String) args.get("new_string");
        Path path = ctx.resolvePath(filePath);

        if (!Files.exists(path)) return ToolResult.fail("文件不存在: " + filePath);

        try {
            String content = Files.readString(path);
            int count = countOccurrences(content, oldString);

            if (count == 0) return ToolResult.fail("在文件中未找到 old_string");
            if (count > 1) {
                return ToolResult.fail(
                    "old_string 在文件中出现了 " + count + " 次。" +
                    "请使用更长的字符串，包含更多上下文以确保唯一匹配。"
                );
            }

            String updated = content.replace(oldString, newString);
            Files.writeString(path, updated);
            return ToolResult.ok("文件已编辑: " + filePath);
        } catch (IOException e) {
            return ToolResult.fail("编辑文件出错: " + e.getMessage());
        }
    }

    /** 统计 search 在 content 中非重叠出现的次数。 */
    private int countOccurrences(String content, String search) {
        int count = 0;
        int idx = 0;
        while ((idx = content.indexOf(search, idx)) != -1) {
            count++;
            idx += search.length();
        }
        return count;
    }
}
