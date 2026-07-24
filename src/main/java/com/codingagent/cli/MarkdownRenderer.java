package com.codingagent.cli;

/**
 * 轻量级 Markdown→ANSI 终端渲染器。
 * 无外部依赖，只处理 Agent 输出中最常见的格式。
 *
 * <p>支持的语法：
 * <ul>
 *   <li>**粗体** / __粗体__</li>
 *   <li>*斜体* / _斜体_</li>
 *   <li>`行内代码`</li>
 *   <li>### 标题</li>
 *   <li>```代码块```</li>
 *   <li>- 无序列表 / 1. 有序列表</li>
 *   <li>[链接](url) → 仅显示文本</li>
 * </ul>
 */
public class MarkdownRenderer {

    // ANSI 控制码
    private static final String RESET  = "\033[0m";
    private static final String BOLD   = "\033[1m";
    private static final String DIM    = "\033[2m";
    private static final String ITALIC = "\033[3m";
    private static final String UNDER  = "\033[4m";
    private static final String CYAN   = "\033[36m";
    private static final String YELLOW = "\033[33m";
    private static final String GREEN  = "\033[32m";
    private static final String BLUE   = "\033[34m";

    /** 将 Markdown 文本渲染为带 ANSI 转义序列的终端文本。 */
    public String render(String md) {
        if (md == null || md.isEmpty()) return "";

        StringBuilder out = new StringBuilder();
        String[] lines = md.split("\n", -1);
        boolean inCodeBlock = false;

        for (String line : lines) {
            // 代码块开关
            if (line.trim().startsWith("```")) {
                inCodeBlock = !inCodeBlock;
                if (inCodeBlock) {
                    out.append(DIM);
                    String lang = line.trim().substring(3).trim();
                    if (!lang.isEmpty()) out.append("+-- ").append(lang).append(" -------------------------------------").append(RESET).append("\n");
                } else {
                    out.append(DIM).append("+------------------------------------------").append(RESET).append("\n");
                }
                continue;
            }

            if (inCodeBlock) {
                out.append(DIM).append(" | ").append(line).append(RESET).append("\n");
                continue;
            }

            // 标题
            if (line.matches("^#{1,4}\\s+.*")) {
                int level = 0;
                while (level < line.length() && line.charAt(level) == '#') level++;
                String text = line.substring(level).trim();
                if (level <= 2) {
                    out.append(BOLD).append(UNDER).append(text).append(RESET).append("\n");
                } else {
                    out.append(BOLD).append(text).append(RESET).append("\n");
                }
                continue;
            }

            // 分隔线
            if (line.matches("^[-*=]{3,}$")) {
                out.append(DIM).append("-".repeat(60)).append(RESET).append("\n");
                continue;
            }

            // 列表项
            if (line.matches("^\\s*[-*+]\\s+.*") || line.matches("^\\s*\\d+\\.\\s+.*")) {
                out.append("  ").append(GREEN).append("-").append(RESET).append(" ");
                String text = line.replaceFirst("^\\s*(?:[-*+]|\\d+\\.)\\s+", "");
                out.append(renderInline(text)).append("\n");
                continue;
            }

            // 引用
            if (line.startsWith("> ")) {
                out.append(DIM).append("| ").append(renderInline(line.substring(2))).append(RESET).append("\n");
                continue;
            }

            // 普通行
            out.append(renderInline(line)).append("\n");
        }

        return out.toString();
    }

    /** 渲染行内格式：粗体、斜体、代码、链接。 */
    private String renderInline(String text) {
        // 粗体 **text** 或 __text__
        text = text.replaceAll("\\*\\*(.+?)\\*\\*", BOLD + "$1" + RESET);
        text = text.replaceAll("__(.+?)__", BOLD + "$1" + RESET);

        // 斜体 *text* 或 _text_
        text = text.replaceAll("(?<!\\*)\\*(?!\\*)(.+?)(?<!\\*)\\*(?!\\*)", ITALIC + "$1" + RESET);
        text = text.replaceAll("(?<!_)_(?!_)(.+?)(?<!_)_(?!_)", ITALIC + "$1" + RESET);

        // 行内代码 `code`
        text = text.replaceAll("`([^`]+)`", CYAN + "$1" + RESET);

        // 链接 [text](url) → 仅显示文本
        text = text.replaceAll("\\[([^]]+)]\\([^)]+\\)", BLUE + "$1" + RESET);

        return text;
    }
}
