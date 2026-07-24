package com.codingagent.agent;

import com.codingagent.llm.LlmProvider;
import com.codingagent.llm.LlmResponse;
import com.codingagent.llm.Message;
import com.codingagent.sandbox.PermissionDecision;
import com.codingagent.sandbox.SandboxBase;
import com.codingagent.tool.ToolBase;
import com.codingagent.tool.ToolContext;
import com.codingagent.tool.ToolRegistry;
import com.codingagent.tool.ToolResult;

import java.nio.file.Path;
import java.util.*;

/**
 * 核心 ReAct（Reasoning + Acting）循环。
 *
 * <p>算法流程：
 * <ol>
 *   <li>构建初始上下文：系统提示 +（可选）记忆 + 用户任务</li>
 *   <li>调用 LLM → 获取文本响应或工具调用列表</li>
 *   <li>若是最终回答 → 返回</li>
 *   <li>若是工具调用 → 检查权限 → 执行 → 追加观察结果 → 继续循环</li>
 *   <li>达到 {@link #MAX_ITERATIONS} 步后，强制要求给出最终回答</li>
 * </ol>
 *
 * <p>本循环保证的安全约束：</p>
 * <ul>
 *   <li>硬性步数上限防止无限工具调用循环</li>
 *   <li>权限拒绝返回错误观察结果——Agent 可自行恢复</li>
 *   <li>未知工具以错误形式报告，而非静默忽略</li>
 *   <li>tool_call / tool_result 消息对始终相邻（压缩时永不拆分）</li>
 * </ul>
 */
public class ReActLoop {

    /** 安全上限：防止工具调用循环失控。 */
    private static final int MAX_ITERATIONS = 50;

    private final LlmProvider llm;
    private final ToolRegistry toolRegistry;
    private final EventBus eventBus;
    private final int maxIterations;

    public ReActLoop(LlmProvider llm, ToolRegistry toolRegistry, EventBus eventBus) {
        this(llm, toolRegistry, eventBus, MAX_ITERATIONS);
    }

    public ReActLoop(LlmProvider llm, ToolRegistry toolRegistry, EventBus eventBus, int maxIterations) {
        this.llm = llm;
        this.toolRegistry = toolRegistry;
        this.eventBus = eventBus;
        this.maxIterations = maxIterations;
    }

    /**
     * 为单个任务执行完整的 ReAct 循环。
     *
     * @param task         用户的任务描述
     * @param systemPrompt 合并后的系统指令
     * @param memories     跨会话记忆条目（可为 null）
     * @param sandbox       权限和执行沙箱
     * @param workspace     文件操作的工作目录
     * @return 最终 Agent 结果，包含回答、统计信息和对话记录
     */
    public AgentResult run(String task, String systemPrompt,
                           List<Message> memories, SandboxBase sandbox, Path workspace) {
        // 构建初始消息上下文
        List<Message> messages = new ArrayList<>();
        messages.add(Message.system(systemPrompt));
        if (memories != null) {
            messages.addAll(memories);
        }
        messages.add(Message.user(task));

        List<AgentResult.ToolCallRecord> toolCallRecords = new ArrayList<>();
        int step = 0;

        while (step < maxIterations) {
            step++;

            // 发送状态事件供实时显示
            eventBus.emit(new AgentState(task, step, countTokens(messages),
                null, sandbox.policyEngine().currentLevel().name(), 0));

            // === THOUGHT：调用 LLM，传入所有已知工具 ===
            LlmResponse response = llm.chat(messages, toolRegistry.toLlmFormats());

            // === PARSE：无工具调用 → 文本回复或最终回答 ===
            if (response.toolCalls() == null || response.toolCalls().isEmpty()) {
                messages.add(Message.assistant(response.content(), response.reasoningContent()));

                // "end_turn" 且无工具调用表示模型认为任务已完成
                if (response.isFinalAnswer() || response.stopReason() != null
                    && response.stopReason().contains("end_turn")) {
                    return new AgentResult(response.content(), step, toolCallRecords, messages);
                }
                // 自由文本推理——继续循环
                continue;
            }

            // === ACTION + OBSERVATION：执行每个请求的工具 ===
            for (Message.ToolCall tc : response.toolCalls()) {
                // 优雅处理未知工具——Agent 可换另一种方式尝试
                if (!toolRegistry.has(tc.name())) {
                    String obs = "Error: Tool '" + tc.name() + "' not found. Available: " +
                        toolRegistry.toLlmFormats().stream().map(m -> (String) m.get("name")).toList();
                    messages.add(Message.assistantWithToolCalls(tc, response.reasoningContent()));
                    messages.add(Message.toolResult(tc.id(), obs));
                    toolCallRecords.add(new AgentResult.ToolCallRecord(tc.name(), tc.arguments(), false));
                    continue;
                }

                ToolBase tool = toolRegistry.resolve(tc.name());

                // 权限关卡——拒绝/弹窗返回错误观察结果，而非崩溃
                PermissionDecision decision = sandbox.policyEngine().check(tc.name(), tool.requiredPermission());
                String observation;

                if (decision.verdict() == PermissionDecision.Verdict.DENY) {
                    observation = "Error: Tool '" + tc.name() + "' was denied by security policy: " + decision.reason();
                    toolCallRecords.add(new AgentResult.ToolCallRecord(tc.name(), tc.arguments(), false));
                } else if (decision.verdict() == PermissionDecision.Verdict.PROMPT) {
                    observation = "Error: Tool '" + tc.name() + "' requires permission escalation: " + decision.reason() +
                        ". Use a different approach or ask the user to escalate permissions.";
                    toolCallRecords.add(new AgentResult.ToolCallRecord(tc.name(), tc.arguments(), false));
                } else {
                    // 在沙箱内执行工具
                    try {
                        ToolContext ctx = new ToolContext(workspace, sandbox);
                        ToolResult result = tool.execute(ctx, tc.arguments());
                        observation = result.formatForLlm();
                        toolCallRecords.add(new AgentResult.ToolCallRecord(tc.name(), tc.arguments(), result.success()));
                    } catch (Exception e) {
                        observation = "Error executing " + tc.name() + ": " + e.getMessage();
                        toolCallRecords.add(new AgentResult.ToolCallRecord(tc.name(), tc.arguments(), false));
                    }
                }

                eventBus.emit(new AgentState(task, step, countTokens(messages),
                    tc.name(), sandbox.policyEngine().currentLevel().name(), 0));

                // 追加 tool_call / tool_result 消息对，二者相邻
                messages.add(Message.assistantWithToolCalls(tc, response.reasoningContent()));
                messages.add(Message.toolResult(tc.id(), observation));
            }
        }

        // 达到步数上限——强制模型给出最终回答
        messages.add(Message.user("You have reached the maximum number of steps. Please provide your final answer now."));
        LlmResponse finalResponse = llm.chat(messages, null);
        return new AgentResult(finalResponse.content(), step, toolCallRecords, messages);
    }

    private int countTokens(List<Message> messages) {
        return llm.countMessagesTokens(messages);
    }
}
