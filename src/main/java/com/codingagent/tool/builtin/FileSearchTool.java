package com.codingagent.tool.builtin;

import com.codingagent.sandbox.PermissionLevel;
import com.codingagent.tool.ToolBase;
import com.codingagent.tool.ToolContext;
import com.codingagent.tool.ToolResult;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.regex.Pattern;
import java.util.regex.PatternSyntaxException;
import java.util.stream.Stream;

/**
 * 文件搜索工具，支持两种模式：
 * <ul>
 *   <li><b>glob</b> —— 按文件名模式查找（内部将 glob 转为正则表达式）</li>
 *   <li><b>grep</b> —— 按正则表达式搜索文件内容，输出为"文件:行号:内容"格式</li>
 * </ul>
 *
 * <p>二进制和大文件自动跳过以避免噪音。两种模式的结果上限均为 200 条。</p>
 */
public class FileSearchTool implements ToolBase {

    @Override
    public String name() { return "search"; }

    @Override
    public String description() {
        return "按 glob 模式搜索文件名，或按正则表达式搜索文件内容。" +
               "glob 模式查找文件，grep 模式搜索文本内容。";
    }

    @Override
    @SuppressWarnings("unchecked")
    public Map<String, Object> parameters() {
        return Map.of(
            "type", "object",
            "properties", Map.of(
                "pattern", Map.of("type", "string", "description", "glob 模式（如 '**/*.java'）或 grep 的正则表达式"),
                "mode", Map.of("type", "string", "description", "'glob' 按文件名查找，'grep' 搜索文件内容"),
                "path", Map.of("type", "string", "description", "搜索的目录（默认为工作区根目录）")
            ),
            "required", List.of("pattern", "mode")
        );
    }

    @Override
    public PermissionLevel requiredPermission() { return PermissionLevel.READ; }

    @Override
    public ToolResult execute(ToolContext ctx, Map<String, Object> args) {
        String pattern = (String) args.get("pattern");
        String mode = (String) args.get("mode");
        String searchPath = args.containsKey("path") ? (String) args.get("path") : ".";

        Path baseDir = ctx.resolvePath(searchPath);
        if (!Files.exists(baseDir)) return ToolResult.fail("目录不存在: " + searchPath);

        try {
            if ("glob".equals(mode)) return globSearch(baseDir, pattern);
            else if ("grep".equals(mode)) return grepSearch(baseDir, pattern);
            else return ToolResult.fail("未知模式: " + mode + "。请使用 'glob' 或 'grep'。");
        } catch (IOException e) {
            return ToolResult.fail("搜索出错: " + e.getMessage());
        }
    }

    /** 遍历目录树，按 glob 模式匹配文件路径。 */
    private ToolResult globSearch(Path baseDir, String pattern) throws IOException {
        String regex = globToRegex(pattern);
        java.util.regex.Pattern globPattern = java.util.regex.Pattern.compile(regex);
        List<String> results = new ArrayList<>();

        try (Stream<Path> files = Files.walk(baseDir, 20)) {
            files.filter(Files::isRegularFile)
                .filter(p -> globPattern.matcher(baseDir.relativize(p).toString()).matches())
                .limit(200)
                .forEach(p -> results.add(baseDir.relativize(p).toString()));
        }

        if (results.isEmpty()) return ToolResult.ok("未找到匹配的文件: " + pattern);
        return ToolResult.ok(String.join("\n", results));
    }

    /**
     * 将 Shell 风格的 glob 转换为正则表达式。
     * 支持：**（跨目录）、*（单层内）、?（单字符）、{a,b}（选项）。
     */
    private static String globToRegex(String glob) {
        StringBuilder sb = new StringBuilder("^");
        for (int i = 0; i < glob.length(); i++) {
            char c = glob.charAt(i);
            switch (c) {
                case '*' -> {
                    if (i + 1 < glob.length() && glob.charAt(i + 1) == '*') {
                        sb.append(".*"); // ** → 跨路径分隔符
                        i++;
                    } else {
                        sb.append("[^/]*"); // * → 仅匹配单层
                    }
                }
                case '?' -> sb.append("[^/]");
                case '.' -> sb.append("\\.");
                case '{' -> sb.append("(?:");
                case '}' -> sb.append(')');
                case ',' -> sb.append('|');
                default -> sb.append(c);
            }
        }
        sb.append('$');
        return sb.toString();
    }

    /** 使用正则表达式搜索文件内容；跳过已知的二进制/大型文件类型。 */
    private ToolResult grepSearch(Path baseDir, String pattern) throws IOException {
        Pattern regex;
        try {
            regex = Pattern.compile(pattern);
        } catch (PatternSyntaxException e) {
            return ToolResult.fail("无效的正则表达式: " + e.getMessage());
        }

        List<String> results = new ArrayList<>();
        int matchCount = 0;

        // 过滤已知的二进制/非文本文件扩展名
        try (Stream<Path> files = Files.walk(baseDir, 20)) {
            var fileList = files
                .filter(Files::isRegularFile)
                .filter(f -> {
                    String name = f.getFileName().toString().toLowerCase();
                    return !name.endsWith(".class") && !name.endsWith(".jar")
                        && !name.endsWith(".png") && !name.endsWith(".jpg")
                        && !name.endsWith(".exe") && !name.endsWith(".dll");
                })
                .limit(500)
                .toList();

            for (Path file : fileList) {
                if (matchCount >= 200) break;
                try {
                    List<String> lines = Files.readAllLines(file);
                    for (int i = 0; i < lines.size() && matchCount < 200; i++) {
                        if (regex.matcher(lines.get(i)).find()) {
                            results.add(String.format("%s:%d: %s",
                                baseDir.relativize(file), i + 1, lines.get(i).strip()));
                            matchCount++;
                        }
                    }
                } catch (IOException ignored) {
                    // 静默跳过无法读取的文件
                }
            }
        }

        if (results.isEmpty()) return ToolResult.ok("未找到匹配: " + pattern);
        return ToolResult.ok(String.join("\n", results));
    }
}
