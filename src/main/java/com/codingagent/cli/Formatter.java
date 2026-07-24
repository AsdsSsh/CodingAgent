package com.codingagent.cli;

import com.codingagent.agent.AgentResult;

/**
 * 格式化 {@link AgentResult} 并输出。文本模式渲染 Markdown→ANSI，JSON 模式保留原始格式。
 */
public class Formatter {

    private static final MarkdownRenderer MD = new MarkdownRenderer();

    /** 文本格式：Markdown 渲染为 ANSI 终端样式。 */
    public String format(AgentResult result) {
        StringBuilder sb = new StringBuilder();
        sb.append("\n").append("\033[2m").append("-".repeat(60)).append("\033[0m").append("\n");
        sb.append(MD.render(sanitizeUnicode(result.answer())));
        sb.append("\033[2m").append("-".repeat(60)).append("\033[0m").append("\n");
        sb.append(String.format("\033[2mSteps: %d | Tools: %d\033[0m%n",
            result.steps(),
            result.toolCalls() != null ? result.toolCalls().size() : 0));

        if (result.toolCalls() != null && !result.toolCalls().isEmpty()) {
            for (AgentResult.ToolCallRecord tc : result.toolCalls()) {
                String icon = tc.success() ? "✓" : "✗";
                String color = tc.success() ? "\033[32m" : "\033[31m";
                sb.append(String.format("  %s%s\033[0m \033[2m%s\033[0m%n", color, icon, tc.name()));
            }
        }
        return sb.toString();
    }

    /** JSON 格式：保留原始 Markdown。 */
    public String formatJson(AgentResult result) {
        StringBuilder sb = new StringBuilder();
        sb.append("{\n");
        sb.append("  \"answer\": ").append(jsonEscape(result.answer())).append(",\n");
        sb.append("  \"steps\": ").append(result.steps()).append(",\n");
        sb.append("  \"tool_calls\": [\n");
        if (result.toolCalls() != null) {
            for (int i = 0; i < result.toolCalls().size(); i++) {
                var tc = result.toolCalls().get(i);
                sb.append("    {\"name\": \"").append(tc.name())
                    .append("\", \"success\": ").append(tc.success()).append("}");
                if (i < result.toolCalls().size() - 1) sb.append(",");
                sb.append("\n");
            }
        }
        sb.append("  ]\n");
        sb.append("}\n");
        return sb.toString();
    }

    /**
     * 剥离 Windows GBK 控制台无法显示的字符。
     * 只保留：ASCII (0x20-0x7E)、换行/回车/Tab、中文（BMP 内）。
     * 其余全部替换为 ASCII 等价物或剥离。
     */
    private static String sanitizeUnicode(String s) {
        if (s == null) return null;
        StringBuilder sb = new StringBuilder(s.length());
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            // Emoji (surrogate pair) → 直接跳过
            if (Character.isSurrogate(c)) {
                if (Character.isHighSurrogate(c) && i + 1 < s.length()) i++;
                continue;
            }
            // 安全范围：ASCII 可打印 + ASCII 控制（TAB/换行）+ 中文 BMP
            if ((c >= 0x20 && c <= 0x7E) || c == '\n' || c == '\r' || c == '\t') {
                sb.append(c);
            } else if ((c >= 0x4E00 && c <= 0x9FFF)       // CJK 统一汉字
                    || (c >= 0x3400 && c <= 0x4DBF)        // CJK 扩展 A
                    || (c >= 0xF900 && c <= 0xFAFF)        // CJK 兼容汉字
                    || (c >= 0x3000 && c <= 0x303F)        // CJK 标点
                    || (c >= 0xFF00 && c <= 0xFFEF)) {     // 全角/半角形式
                sb.append(c);
            }
            // 其余（emoji、符号、装饰）→ 丢弃
        }
        return sb.toString();
    }

    private String jsonEscape(String s) {
        if (s == null) return "null";
        return "\"" + s.replace("\\", "\\\\").replace("\"", "\\\"")
            .replace("\n", "\\n").replace("\r", "\\r").replace("\t", "\\t") + "\"";
    }
}
