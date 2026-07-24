package com.codingagent.tool;

/**
 * 工具执行的结果。成功/失败通过 {@code success} 标志编码；
 * LLM 通过 {@link #formatForLlm()} 接收格式化后的输出。
 *
 * <p>输出上限约 100K 字符（~25K token），防止单次工具结果
 * 消耗整个上下文窗口。被截断的结果会包含提示，告知 Agent 缩小查询范围。</p>
 */
public record ToolResult(boolean success, String data, String error, boolean truncated) {

    public static ToolResult ok(String data) {
        return new ToolResult(true, data, null, false);
    }

    public static ToolResult ok(String data, boolean truncated) {
        return new ToolResult(true, data, null, truncated);
    }

    public static ToolResult fail(String error) {
        return new ToolResult(false, "", error, false);
    }

    /** 工具输出硬上限，保护上下文窗口不被撑爆。 */
    private static final int MAX_OUTPUT_LENGTH = 100_000;

    /**
     * 由命令执行的 stdout/stderr 创建结果。合并两个流；
     * 非零退出码产生失败结果。
     */
    public static ToolResult fromCommandOutput(int exitCode, String stdout, String stderr) {
        StringBuilder sb = new StringBuilder();
        if (stdout != null && !stdout.isEmpty()) sb.append(stdout);
        if (stderr != null && !stderr.isEmpty()) {
            if (!sb.isEmpty()) sb.append("\n");
            sb.append(stderr);
        }
        String output = sb.toString();
        boolean truncated = output.length() > MAX_OUTPUT_LENGTH;
        if (truncated) {
            output = output.substring(0, MAX_OUTPUT_LENGTH) + "\n... [output truncated]";
        }
        if (exitCode != 0) {
            return new ToolResult(false, output, "Exit code: " + exitCode, truncated);
        }
        return new ToolResult(true, output, null, truncated);
    }

    /** 格式化此结果，用于插入 LLM 对话。 */
    public String formatForLlm() {
        if (!success) {
            return "Error: " + (error != null ? error : "Unknown error");
        }
        if (truncated) {
            return data + "\n\n[Output truncated — use more specific parameters]";
        }
        return data;
    }
}
