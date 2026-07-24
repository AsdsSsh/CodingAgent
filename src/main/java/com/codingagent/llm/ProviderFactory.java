package com.codingagent.llm;

import com.codingagent.config.ModelConfig;

/**
 * 从配置创建 LLM 提供商——内置自动模型检测，
 * 大多数情况只需配置模型名即可自动匹配端点。
 */
public class ProviderFactory {

    /**
     * 自动检测 + 创建提供商。
     *
     * <p>检测优先级：
     * <ol>
     *   <li>base_url 明确配置 → 用该 URL + OpenAI 兼容格式</li>
     *   <li>模型名匹配内置预设 → 用预设的 base_url 和 api_format</li>
     *   <li>均不匹配 → 报错，提示设置 base_url</li>
     * </ol>
     */
    public static LlmProvider create(ModelConfig config) {
        ModelRegistry.ProviderPreset preset = ModelRegistry.detect(config.model());

        // 1. 用户显式指定了 base_url → 优先使用
        if (config.baseUrl() != null && !config.baseUrl().isBlank()) {
            return createOpenAI(config, config.baseUrl());
        }

        // 2. 模型名匹配到了预设
        if (preset != null) {
            String apiKey = ModelRegistry.resolveApiKey(preset, config.apiKey());
            if (apiKey == null || apiKey.isBlank()) {
                throw missingKeyError(preset);
            }
            return switch (preset.apiFormat()) {
                case "anthropic" -> new AnthropicProvider(
                    apiKey, config.model(), config.maxTokens(), config.contextWindow()
                );
                default -> new OpenAIProvider(
                    apiKey, config.model(), config.maxTokens(), config.contextWindow(),
                    preset.baseUrl(), config.temperature()
                );
            };
        }

        // 3. provider 明确指定了 anthropic
        if ("anthropic".equalsIgnoreCase(config.provider())) {
            String apiKey = ModelRegistry.resolveApiKey(null, config.apiKey());
            if (apiKey == null || apiKey.isBlank()) {
                throw new IllegalStateException("未找到 Anthropic API 密钥，请设置 ANTHROPIC_API_KEY 环境变量");
            }
            return new AnthropicProvider(
                apiKey, config.model(), config.maxTokens(), config.contextWindow()
            );
        }

        // 4. 无法识别
        throw new IllegalArgumentException(
            "无法识别模型 '" + config.model() + "'。\n" +
            "  可用预设: " + ModelRegistry.all().stream()
                .map(ModelRegistry.ProviderPreset::name).toList() + "\n" +
            "  或在配置中设置 base_url 使用自定义端点。"
        );
    }

    private static LlmProvider createOpenAI(ModelConfig config, String baseUrl) {
        String apiKey = ModelRegistry.resolveApiKey(null, config.apiKey());
        if (apiKey == null || apiKey.isBlank()) {
            throw new IllegalStateException("未找到 API 密钥，请设置 OPENAI_API_KEY 环境变量或在配置中设置 api_key");
        }
        return new OpenAIProvider(
            apiKey, config.model(), config.maxTokens(), config.contextWindow(),
            baseUrl, config.temperature()
        );
    }

    private static IllegalStateException missingKeyError(ModelRegistry.ProviderPreset preset) {
        return new IllegalStateException(
            "未找到 " + preset.name() + " 的 API 密钥。\n" +
            "  export " + preset.envVar() + "=你的密钥\n" +
            "  或在配置文件中设置 api_key 字段"
        );
    }
}
