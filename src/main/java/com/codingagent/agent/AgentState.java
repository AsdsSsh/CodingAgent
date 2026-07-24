package com.codingagent.agent;

/**
 * Agent 在单步执行中的不可变状态快照。
 * 通过 {@link EventBus} 发送，供实时显示和日志记录使用。
 *
 * <p>各字段有意保持粗粒度——这是供人类阅读的状态记录，
 * 而非机器可读的审计日志。</p>
 */
public record AgentState(
    String task,
    int step,
    int conversationTokens,
    String currentTool,
    String permissionLevel,
    int compressionCount
) {}
