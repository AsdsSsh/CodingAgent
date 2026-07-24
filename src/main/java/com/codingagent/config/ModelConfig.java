package com.codingagent.config;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * 不可变的模型配置。YAML 加载 + 环境变量覆盖 + CLI 参数覆盖。
 *
 * <p>提供商可以是任意字符串——若 {@code provider} 为 "openai" 或设置了 {@code base_url}，
 * 则自动使用 OpenAI 兼容的 API 格式。只有 "anthropic" 走 Anthropic 专有格式。</p>
 *
 * @param provider      提供商："anthropic"、"openai" 或任意自定义名称
 * @param model         模型标识（如 "claude-sonnet-4-20250514"、"gpt-4o"、"deepseek-chat"）
 * @param apiKey        API 密钥——优先使用 CLI 参数，其次配��文件，最后环境变量
 * @param apiKeyEnv     读取 API 密钥的环境变量名（不设置则按 provider 自动推断）
 * @param baseUrl       自定义端点 URL（如 DeepSeek: https://api.deepseek.com/v1, Ollama: http://localhost:11434/v1）
 * @param maxTokens     单次响应的最大 token 数
 * @param contextWindow 该模型的总上下文窗口大小
 * @param temperature   温度参数（null 表示使用 API 默认值，0.0=精确确定，1.0=最大创意）
 */
public record ModelConfig(
    @JsonProperty("provider") String provider,
    @JsonProperty("model") String model,
    @JsonProperty("api_key") String apiKey,
    @JsonProperty("api_key_env") String apiKeyEnv,
    @JsonProperty("base_url") String baseUrl,
    @JsonProperty("max_tokens") int maxTokens,
    @JsonProperty("context_window") int contextWindow,
    @JsonProperty("temperature") Double temperature
) {
    /** 默认配置：Anthropic Claude Sonnet 4。 */
    public static ModelConfig defaults() {
        return new ModelConfig(
            "anthropic", "claude-sonnet-4-20250514",
            null, null, null,
            8192, 200_000, null
        );
    }

    /** 返回替换了 API 密钥的副本（环境变量/CLI 覆盖用）。 */
    public ModelConfig withApiKey(String apiKey) {
        return new ModelConfig(provider, model, apiKey, apiKeyEnv, baseUrl,
            maxTokens, contextWindow, temperature);
    }

    /** 返回替换了模型的副本（CLI --model 覆盖用）。 */
    public ModelConfig withModel(String model) {
        return new ModelConfig(provider, model, apiKey, apiKeyEnv, baseUrl,
            maxTokens, contextWindow, temperature);
    }

    /** 返回替换了提供商的副本。 */
    public ModelConfig withProvider(String provider) {
        return new ModelConfig(provider, model, apiKey, apiKeyEnv, baseUrl,
            maxTokens, contextWindow, temperature);
    }

    /**
     * 解析有效的 API 密钥，按优先级：配置文件 > 指定的环境变量 > provider 默认环境变量。
     */
    public String resolveApiKey() {
        if (apiKey != null && !apiKey.isBlank()) return apiKey;

        // 用户指定的环境变量名优先
        if (apiKeyEnv != null && !apiKeyEnv.isBlank()) {
            String val = System.getenv(apiKeyEnv);
            if (val != null) return val;
        }

        // provider 默认环境变量（provider 可能为 null——YAML 未配置时）
        String prov = provider != null ? provider.toLowerCase() : "";
        return switch (prov) {
            case "anthropic" -> System.getenv("ANTHROPIC_API_KEY");
            case "openai" -> System.getenv("OPENAI_API_KEY");
            default -> {
                // 尝试 OPENAI_API_KEY（兼容端点通常也接受此变量）
                String key = System.getenv("OPENAI_API_KEY");
                yield key != null ? key : System.getenv("ANTHROPIC_API_KEY");
            }
        };
    }
}
