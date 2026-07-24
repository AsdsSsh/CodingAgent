package com.codingagent.llm;

import java.util.List;
import java.util.Map;

/**
 * 所有 LLM 提供商（Anthropic、OpenAI、本地模型）的抽象接口。
 *
 * <p>我们直接使用 HTTP/SDK 调用而非 LangChain4j 等包装框架，
 * 以便访问各提供商的专有特性（prompt caching、thinking blocks、
 * 结构化输出），而不受抽象层的限制。</p>
 *
 * <p>所有方法均为阻塞调用——使用虚拟线程时无需 async/await 即可获得并发能力。</p>
 */
public interface LlmProvider extends AutoCloseable {

    /** 人类可读的模型标识。 */
    String modelName();

    /** 总上下文窗口大小（token 数），供 ContextManager 计算压缩阈值使用。 */
    int contextWindow();

    /**
     * 向 LLM 发送消息，返回文本或工具调用。
     * 工具以 JSON Schema 映射的形式传入——各提供商将其转换为对应的 API 格式。
     */
    LlmResponse chat(List<Message> messages, List<Map<String, Object>> tools);

    /** 单字符串的近似 token 计数。 */
    int countTokens(String text);

    /** 完整消息列表（内容 + 工具调用）的近似 token 计数。 */
    default int countMessagesTokens(List<Message> messages) {
        int total = 0;
        for (Message msg : messages) {
            if (msg.content() != null) total += countTokens(msg.content());
            if (msg.toolCalls() != null) {
                for (Message.ToolCall tc : msg.toolCalls()) {
                    total += countTokens(tc.name());
                    if (tc.arguments() != null) {
                        total += countTokens(tc.arguments().toString());
                    }
                }
            }
        }
        return total;
    }

    /** 默认无操作；若提供商持有资源则覆盖此方法。 */
    @Override
    default void close() {}
}
