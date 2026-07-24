package com.codingagent.llm;

/**
 * LLM API 每次调用返回的 token 使用统计。
 * 由 {@code ContextManager} 用于判断何时触发上下文压缩。
 */
public record TokenUsage(int inputTokens, int outputTokens) {
    public int total() {
        return inputTokens + outputTokens;
    }
}
