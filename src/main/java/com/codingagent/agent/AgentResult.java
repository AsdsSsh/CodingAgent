package com.codingagent.agent;

import com.codingagent.llm.Message;

import java.util.List;

/**
 * Agent 完成一次运行后的结果：最终回答文本、执行统计信息
 * 以及完整的对话记录（用于记忆提取）。
 */
public record AgentResult(
    String answer,
    int steps,
    List<ToolCallRecord> toolCalls,
    List<Message> conversation
) {
    /** 本次运行中每次工具调用的记录。 */
    public record ToolCallRecord(String name, java.util.Map<String, Object> args, boolean success) {}
}
