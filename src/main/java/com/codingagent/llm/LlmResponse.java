package com.codingagent.llm;

import java.util.List;

/**
 * 解析后的 LLM 响应。最终文本回答 或 工具调用列表，两者互斥。
 * {@code reasoningContent} 用于 DeepSeek 思考模式的回传。
 */
public record LlmResponse(
    String content,
    List<Message.ToolCall> toolCalls,
    boolean isFinalAnswer,
    TokenUsage usage,
    String stopReason,
    String reasoningContent
) {
    public LlmResponse {
        if (usage == null) usage = new TokenUsage(0, 0);
        if (stopReason == null) stopReason = "";
    }

    /** 向后兼容构造器（无 reasoningContent）。 */
    public LlmResponse(String content, List<Message.ToolCall> toolCalls,
                       boolean isFinalAnswer, TokenUsage usage, String stopReason) {
        this(content, toolCalls, isFinalAnswer, usage, stopReason, null);
    }

    public boolean hasToolCalls() {
        return toolCalls != null && !toolCalls.isEmpty();
    }
}
