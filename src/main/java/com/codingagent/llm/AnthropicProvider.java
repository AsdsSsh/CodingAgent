package com.codingagent.llm;

import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 基于 {@link java.net.http.HttpClient} 直接调用 Anthropic Messages API 的提供商实现。
 *
 * <p>绕开 Anthropic Java SDK 的原因：我们可以完全控制 HTTP 层（超时、重试），
 * 并且无需等待 SDK 更新即可访问全部 API 特性。
 * 代价是需要手动构建 JSON 请求体和解析响应。</p>
 *
 * <p>线程安全：实例在构造后不可变，可安全地并发使用。
 * {@link HttpClient} 本身即为线程安全。</p>
 */
public class AnthropicProvider implements LlmProvider {

    private static final String API_URL = "https://api.anthropic.com/v1/messages";
    /** Anthropic API 版本头——固定版本以保证稳定性。 */
    private static final String VERSION = "2023-06-01";
    private static final ObjectMapper MAPPER = new ObjectMapper();

    private final HttpClient httpClient;
    private final String apiKey;
    private final String model;
    private final int maxTokens;
    private final int contextWindow;

    public AnthropicProvider(String apiKey, String model, int maxTokens, int contextWindow) {
        this.httpClient = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(30))
            .build();
        this.apiKey = apiKey;
        this.model = model;
        this.maxTokens = maxTokens;
        this.contextWindow = contextWindow;
    }

    @Override
    public String modelName() { return model; }

    @Override
    public int contextWindow() { return contextWindow; }

    /**
     * 向 Anthropic Messages API 发送聊天请求。
     * System 消息被提取到顶层的 "system" 参数发送；
     * 其余消息通过 "messages" 数组发送。
     */
    @Override
    @SuppressWarnings("unchecked")
    public LlmResponse chat(List<Message> messages, List<Map<String, Object>> tools) {
        try {
            Map<String, Object> body = new HashMap<>();
            body.put("model", model);
            body.put("max_tokens", maxTokens);

            // Anthropic 要求 system prompt 使用顶层字段，而非放在消息列表中
            String systemPrompt = extractSystemPrompt(messages);
            if (systemPrompt != null) {
                body.put("system", systemPrompt);
            }

            body.put("messages", convertToApiMessages(messages));

            if (tools != null && !tools.isEmpty()) {
                body.put("tools", convertToApiTools(tools));
            }

            String json = MAPPER.writeValueAsString(body);

            HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(API_URL))
                .header("x-api-key", apiKey)
                .header("anthropic-version", VERSION)
                .header("content-type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(json))
                .timeout(Duration.ofSeconds(120))
                .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() != 200) {
                throw new RuntimeException("Anthropic API error: " + response.statusCode() + " - " + response.body());
            }

            Map<String, Object> result = MAPPER.readValue(response.body(), Map.class);
            return parseApiResponse(result);

        } catch (IOException | InterruptedException e) {
            throw new RuntimeException("Failed to call Anthropic API", e);
        }
    }

    /**
     * 将统一消息格式转换为 Anthropic 的 content-block 格式。
     * 包含工具调用的 assistant 消息被转换为 content block 列表（text + tool_use）；
     * tool 消息被转换为 tool_result block。
     */
    private List<Map<String, Object>> convertToApiMessages(List<Message> messages) {
        List<Map<String, Object>> apiMessages = new ArrayList<>();
        for (Message msg : messages) {
            if (msg.role() == Message.Role.SYSTEM) continue; // 已单独处理

            Map<String, Object> apiMsg = new HashMap<>();
            apiMsg.put("role", msg.role().name().toLowerCase());

            if (msg.role() == Message.Role.ASSISTANT && msg.toolCalls() != null && !msg.toolCalls().isEmpty()) {
                List<Map<String, Object>> contentBlocks = new ArrayList<>();
                if (msg.content() != null && !msg.content().isEmpty()) {
                    contentBlocks.add(Map.of("type", "text", "text", msg.content()));
                }
                for (Message.ToolCall tc : msg.toolCalls()) {
                    Map<String, Object> toolBlock = new HashMap<>();
                    toolBlock.put("type", "tool_use");
                    toolBlock.put("id", tc.id());
                    toolBlock.put("name", tc.name());
                    toolBlock.put("input", tc.arguments() != null ? tc.arguments() : Map.of());
                    contentBlocks.add(toolBlock);
                }
                apiMsg.put("content", contentBlocks);
            } else if (msg.role() == Message.Role.TOOL) {
                Map<String, Object> toolResult = new HashMap<>();
                toolResult.put("type", "tool_result");
                toolResult.put("tool_use_id", msg.toolCallId());
                toolResult.put("content", msg.content());
                apiMsg.put("content", List.of(toolResult));
            } else {
                apiMsg.put("content", msg.content() != null ? msg.content() : "");
            }

            apiMessages.add(apiMsg);
        }
        return apiMessages;
    }

    /** 收集所有 SYSTEM 消息，合并为单个 system prompt 字符串。 */
    private String extractSystemPrompt(List<Message> messages) {
        StringBuilder sb = new StringBuilder();
        for (Message msg : messages) {
            if (msg.role() == Message.Role.SYSTEM) {
                if (!sb.isEmpty()) sb.append("\n\n");
                sb.append(msg.content());
            }
        }
        return !sb.isEmpty() ? sb.toString() : null;
    }

    @SuppressWarnings("unchecked")
    private List<Map<String, Object>> convertToApiTools(List<Map<String, Object>> tools) {
        List<Map<String, Object>> apiTools = new ArrayList<>();
        for (Map<String, Object> tool : tools) {
            Map<String, Object> apiTool = new HashMap<>();
            apiTool.put("name", tool.get("name"));
            apiTool.put("description", tool.getOrDefault("description", ""));
            apiTool.put("input_schema", tool.getOrDefault("input_schema",
                Map.of("type", "object", "properties", Map.of())));
            apiTools.add(apiTool);
        }
        return apiTools;
    }

    /**
     * 将 Anthropic 响应 JSON 解析为统一的响应模型。
     * text block 拼接为文本；tool_use block 转换为工具调用。
     * stop_reason 为 "end_turn" 且无工具调用时表示最终回答。
     */
    @SuppressWarnings("unchecked")
    private LlmResponse parseApiResponse(Map<String, Object> result) {
        List<Map<String, Object>> contentBlocks = (List<Map<String, Object>>) result.get("content");
        StringBuilder text = new StringBuilder();
        List<Message.ToolCall> toolCalls = new ArrayList<>();

        for (Map<String, Object> block : contentBlocks) {
            String type = (String) block.get("type");
            if ("text".equals(type)) {
                text.append((String) block.get("text"));
            } else if ("tool_use".equals(type)) {
                Map<String, Object> input = (Map<String, Object>) block.get("input");
                toolCalls.add(new Message.ToolCall(
                    (String) block.get("id"),
                    (String) block.get("name"),
                    input != null ? input : Map.of()
                ));
            }
        }

        String stopReason = (String) result.get("stop_reason");
        Map<String, Object> usage = (Map<String, Object>) result.get("usage");
        int inputTokens = usage != null ? ((Number) usage.get("input_tokens")).intValue() : 0;
        int outputTokens = usage != null ? ((Number) usage.get("output_tokens")).intValue() : 0;

        // 仅当模型自然结束且未发起工具调用时，才视为最终回答
        boolean isFinal = "end_turn".equals(stopReason) && toolCalls.isEmpty();

        return new LlmResponse(
            text.toString(),
            toolCalls.isEmpty() ? null : toolCalls,
            isFinal,
            new TokenUsage(inputTokens, outputTokens),
            stopReason
        );
    }

    /**
     * 粗略的 token 估算：英文文本约每 4 个字符对应 1 个 token。
     * 精确计数需使用模型的分词器；此估算足以供 ContextManager 的压缩阈值判断使用。
     */
    @Override
    public int countTokens(String text) {
        return text == null ? 0 : text.length() / 4;
    }
}
