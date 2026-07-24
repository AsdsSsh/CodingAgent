package com.codingagent.llm;

import java.util.List;
import java.util.Map;

/**
 * 所有 LLM 提供商共用的统一消息类型。
 * SYSTEM 消息携带指令，USER/ASSISTANT 携带对话，
 * TOOL 消息通过 {@code toolCallId} 关联对应的工具调用结果。
 * {@code reasoningContent} 用于 DeepSeek 等模型的思考模式。
 */
public record Message(Role role, String content, List<ToolCall> toolCalls,
                      String toolCallId, String reasoningContent) {

    public enum Role { SYSTEM, USER, ASSISTANT, TOOL }

    // -- 向后兼容构造器 --
    public Message(Role role, String content, List<ToolCall> toolCalls, String toolCallId) {
        this(role, content, toolCalls, toolCallId, null);
    }

    public static Message system(String content) {
        return new Message(Role.SYSTEM, content, null, null, null);
    }

    public static Message user(String content) {
        return new Message(Role.USER, content, null, null, null);
    }

    public static Message assistant(String content) {
        return new Message(Role.ASSISTANT, content, null, null, null);
    }

    public static Message assistant(String content, String reasoningContent) {
        return new Message(Role.ASSISTANT, content, null, null, reasoningContent);
    }

    /** 携带工具调用的 assistant 消息。压缩器保证与 toolResult 成对不拆分。 */
    public static Message assistantWithToolCalls(ToolCall toolCall) {
        return new Message(Role.ASSISTANT, null, List.of(toolCall), null, null);
    }

    /** 携带工具调用和思考内容的 assistant 消息。 */
    public static Message assistantWithToolCalls(ToolCall toolCall, String reasoningContent) {
        return new Message(Role.ASSISTANT, null, List.of(toolCall), null, reasoningContent);
    }

    /** 工具结果消息，通过 toolCallId 与对应的 assistant 消息配对。 */
    public static Message toolResult(String toolCallId, String content) {
        return new Message(Role.TOOL, content, null, toolCallId, null);
    }

    /** assistant 消息中的单次工具调用。 */
    public record ToolCall(String id, String name, Map<String, Object> arguments) {}
}
