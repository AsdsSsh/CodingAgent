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
 * OpenAI 及 OpenAI 兼容端点的提供商实现。
 *
 * <p>支持所有实现 OpenAI chat completions API 的服务：
 * GPT-4/GPT-4o、DeepSeek、Ollama、vLLM、LM Studio 等。
 * 通过配置 {@code base_url} 切换端点。</p>
 *
 * <p>与 Anthropic API 的关键差异：
 * <ul>
 *   <li>system prompt 作为普通消息，不在顶层字段</li>
 *   <li>工具定义包裹在 {"type":"function","function":{...}} 中</li>
 *   <li>finish_reason: "stop"=完成, "tool_calls"=有工具调用</li>
 * </ul></p>
 */
public class OpenAIProvider implements LlmProvider {

    /** 默认使用 OpenAI 官方端点，可通过 baseUrl 覆盖。 */
    private static final String DEFAULT_BASE_URL = "https://api.openai.com/v1";
    private static final ObjectMapper MAPPER = new ObjectMapper();

    private final HttpClient httpClient;
    private final String apiKey;
    private final String model;
    private final int maxTokens;
    private final int contextWindow;
    private final String baseUrl;
    private final Double temperature;

    public OpenAIProvider(String apiKey, String model, int maxTokens, int contextWindow,
                          String baseUrl, Double temperature) {
        this.httpClient = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(30))
            .build();
        this.apiKey = apiKey;
        this.model = model;
        this.maxTokens = maxTokens;
        this.contextWindow = contextWindow;
        this.baseUrl = (baseUrl != null && !baseUrl.isBlank()) ? baseUrl : DEFAULT_BASE_URL;
        this.temperature = temperature;
    }

    @Override
    public String modelName() { return model; }

    @Override
    public int contextWindow() { return contextWindow; }

    @Override
    @SuppressWarnings("unchecked")
    public LlmResponse chat(List<Message> messages, List<Map<String, Object>> tools) {
        try {
            Map<String, Object> body = new HashMap<>();
            body.put("model", model);
            body.put("max_tokens", maxTokens);
            if (temperature != null) {
                body.put("temperature", temperature);
            }
            body.put("messages", convertToApiMessages(messages));

            if (tools != null && !tools.isEmpty()) {
                body.put("tools", convertToApiTools(tools));
            }

            String json = MAPPER.writeValueAsString(body);

            HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/chat/completions"))
                .header("Authorization", "Bearer " + apiKey)
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(json))
                .timeout(Duration.ofSeconds(120))
                .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() != 200) {
                throw new RuntimeException("OpenAI API error: " + response.statusCode() + " - " + response.body());
            }

            Map<String, Object> result = MAPPER.readValue(response.body(), Map.class);
            return parseApiResponse(result);

        } catch (IOException | InterruptedException e) {
            throw new RuntimeException("Failed to call OpenAI API", e);
        }
    }

    /**
     * 将统一消息格式转换为 OpenAI 的消息格式。
     * system 消息在 OpenAI 中是普通消息（role="system"），非顶层字段。
     * 包含 tool_calls 的 assistant 消息使用 OpenAI 的 tool_calls 数组格式。
     */
    private List<Map<String, Object>> convertToApiMessages(List<Message> messages) {
        List<Map<String, Object>> apiMessages = new ArrayList<>();
        for (Message msg : messages) {
            Map<String, Object> apiMsg = new HashMap<>();
            apiMsg.put("role", msg.role().name().toLowerCase());

            if (msg.role() == Message.Role.ASSISTANT && msg.toolCalls() != null && !msg.toolCalls().isEmpty()) {
                // DeepSeek 思考模式：需回传 reasoning_content
                if (msg.reasoningContent() != null) {
                    apiMsg.put("reasoning_content", msg.reasoningContent());
                }

                if (msg.content() != null && !msg.content().isEmpty()) {
                    apiMsg.put("content", msg.content());
                } else {
                    apiMsg.put("content", null);
                }

                List<Map<String, Object>> toolCallList = new ArrayList<>();
                for (Message.ToolCall tc : msg.toolCalls()) {
                    Map<String, Object> tcMap = new HashMap<>();
                    tcMap.put("id", tc.id());
                    tcMap.put("type", "function");
                    tcMap.put("function", Map.of(
                        "name", tc.name(),
                        "arguments", MAPPER.valueToTree(tc.arguments()).toString()
                    ));
                    toolCallList.add(tcMap);
                }
                apiMsg.put("tool_calls", toolCallList);

            } else if (msg.role() == Message.Role.TOOL) {
                // tool 消息：role=tool，带 tool_call_id
                apiMsg.put("content", msg.content() != null ? msg.content() : "");
                apiMsg.put("tool_call_id", msg.toolCallId());

            } else {
                // user / system / 纯文本 assistant
                apiMsg.put("content", msg.content() != null ? msg.content() : "");
                // DeepSeek 思考模式：纯文本也要回传
                if (msg.role() == Message.Role.ASSISTANT && msg.reasoningContent() != null) {
                    apiMsg.put("reasoning_content", msg.reasoningContent());
                }
            }

            apiMessages.add(apiMsg);
        }
        return apiMessages;
    }

    /** 转换为 OpenAI 的 function 型工具定义格式。 */
    @SuppressWarnings("unchecked")
    private List<Map<String, Object>> convertToApiTools(List<Map<String, Object>> tools) {
        List<Map<String, Object>> apiTools = new ArrayList<>();
        for (Map<String, Object> tool : tools) {
            Map<String, Object> apiTool = new HashMap<>();
            apiTool.put("type", "function");
            apiTool.put("function", Map.of(
                "name", tool.get("name"),
                "description", tool.getOrDefault("description", ""),
                "parameters", tool.getOrDefault("input_schema",
                    Map.of("type", "object", "properties", Map.of()))
            ));
            apiTools.add(apiTool);
        }
        return apiTools;
    }

    /**
     * 解析 OpenAI 响应。
     * choices[0].message.content → 文本
     * choices[0].message.tool_calls[] → 工具调用列表
     * finish_reason "stop" 且无工具调用 → 最终回答
     */
    @SuppressWarnings("unchecked")
    private LlmResponse parseApiResponse(Map<String, Object> result) {
        List<Map<String, Object>> choices = (List<Map<String, Object>>) result.get("choices");
        if (choices == null || choices.isEmpty()) {
            return new LlmResponse("", null, true, new TokenUsage(0, 0), "empty");
        }

        Map<String, Object> choice = choices.get(0);
        String finishReason = (String) choice.get("finish_reason");
        Map<String, Object> message = (Map<String, Object>) choice.get("message");

        String text = message != null && message.get("content") != null
            ? (String) message.get("content") : "";
        // DeepSeek 思考模式：捕获 reasoning_content
        String reasoning = message != null ? (String) message.get("reasoning_content") : null;

        // 解析 tool_calls
        List<Message.ToolCall> toolCalls = null;
        List<Map<String, Object>> rawToolCalls = null;
        if (message != null) {
            rawToolCalls = (List<Map<String, Object>>) message.get("tool_calls");
        }
        if (rawToolCalls != null && !rawToolCalls.isEmpty()) {
            toolCalls = new ArrayList<>();
            for (Map<String, Object> tc : rawToolCalls) {
                String id = (String) tc.get("id");
                String type = (String) tc.get("type");
                Map<String, Object> function = (Map<String, Object>) tc.get("function");
                String name = function != null ? (String) function.get("name") : "";
                String argsStr = function != null ? (String) function.get("arguments") : "{}";

                Map<String, Object> arguments;
                try {
                    arguments = MAPPER.readValue(argsStr, Map.class);
                } catch (Exception e) {
                    arguments = Map.of();
                }

                toolCalls.add(new Message.ToolCall(id, name, arguments));
            }
        }

        boolean isFinal = "stop".equals(finishReason) && (toolCalls == null || toolCalls.isEmpty());
        boolean hasToolCalls = !"stop".equals(finishReason) && toolCalls != null && !toolCalls.isEmpty();

        Map<String, Object> usage = (Map<String, Object>) result.get("usage");
        int inputTokens = usage != null ? ((Number) usage.get("prompt_tokens")).intValue() : 0;
        int outputTokens = usage != null ? ((Number) usage.get("completion_tokens")).intValue() : 0;

        return new LlmResponse(
            text,
            toolCalls,
            isFinal,
            new TokenUsage(inputTokens, outputTokens),
            finishReason,
            reasoning
        );
    }

    @Override
    public int countTokens(String text) {
        return text == null ? 0 : text.length() / 4;
    }
}
