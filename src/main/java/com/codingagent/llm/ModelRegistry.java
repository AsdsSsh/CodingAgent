package com.codingagent.llm;

import java.util.*;

/**
 * 内置模型注册中心——用户只需提供模型名，其余参数自动补全。
 *
 * <p>自动检测规则：
 * <ol>
 *   <li>精确匹配预设中的模型名</li>
 *   <li>模型名前缀匹配（如 deepseek-* → DeepSeek, gpt-* → OpenAI）</li>
 *   <li>未匹配 → OpenAI 兼容格式，需 base_url</li>
 * </ol>
 */
public class ModelRegistry {

    public record ProviderPreset(
        String name,            // "DeepSeek" / "OpenAI" / "Anthropic"
        String apiFormat,       // "anthropic" / "openai"
        String baseUrl,          // API 端点
        String defaultModel,    // 该提供商的推荐模型
        List<String> modelPrefixes, // 模型名前缀，用于自动匹配
        String envVar,           // API 密钥环境变量
        int contextWindow,
        int maxTokens
    ) {}

    /** 所有内置预设，按优先级排序（精确匹配优先）。 */
    private static final List<ProviderPreset> PRESETS = List.of(
        // --- Anthropic ---
        new ProviderPreset("Anthropic", "anthropic",
            "https://api.anthropic.com/v1", "claude-sonnet-4-20250514",
            List.of("claude-"),
            "ANTHROPIC_API_KEY", 200_000, 8192),

        // --- OpenAI ---
        new ProviderPreset("OpenAI", "openai",
            "https://api.openai.com/v1", "gpt-4o",
            List.of("gpt-", "o1", "o3", "o4"),
            "OPENAI_API_KEY", 128_000, 4096),

        // --- DeepSeek ---
        new ProviderPreset("DeepSeek", "openai",
            "https://api.deepseek.com/v1", "deepseek-v4-pro",
            List.of("deepseek-"),
            "DEEPSEEK_API_KEY", 1_000_000, 32768),

        // --- Ollama ---
        new ProviderPreset("Ollama", "openai",
            "http://localhost:11434/v1", "qwen2.5:7b",
            List.of(),  // 手动匹配（模型名无统一前缀）
            null, 32768, 4096),

        // --- Gemini (OpenAI 兼容) ---
        new ProviderPreset("Gemini", "openai",
            "https://generativelanguage.googleapis.com/v1beta/openai", "gemini-2.0-flash",
            List.of("gemini-"),
            "GEMINI_API_KEY", 1_048_576, 4096),

        // --- 月之暗面 Moonshot / Kimi ---
        new ProviderPreset("Moonshot", "openai",
            "https://api.moonshot.cn/v1", "moonshot-v1-8k",
            List.of("moonshot-", "kimi-"),
            "MOONSHOT_API_KEY", 128_000, 4096),

        // --- 通义千问 (阿里云 DashScope) ---
        new ProviderPreset("Qwen", "openai",
            "https://dashscope.aliyuncs.com/compatible-mode/v1", "qwen-max",
            List.of("qwen-"),
            "DASHSCOPE_API_KEY", 131_072, 4096),

        // --- Zhipu / 智谱 GLM ---
        new ProviderPreset("Zhipu", "openai",
            "https://open.bigmodel.cn/api/paas/v4", "glm-4",
            List.of("glm-"),
            "ZHIPU_API_KEY", 128_000, 4096),

        // --- vLLM 自部署 ---
        new ProviderPreset("vLLM", "openai",
            "http://localhost:8000/v1", "",
            List.of(), null, 32768, 4096)
    );

    /**
     * 根据模型名自动匹配预设。
     * 规则：精确匹配预设 defaultModel → 前缀匹配 → 返回 null（调用方需自行处理）
     */
    public static ProviderPreset detect(String modelName) {
        if (modelName == null || modelName.isBlank()) return null;

        String m = modelName.toLowerCase();

        // 1. 精确匹配预设的 defaultModel
        for (ProviderPreset p : PRESETS) {
            if (m.equals(p.defaultModel.toLowerCase())) return p;
        }

        // 2. 前缀匹配
        for (ProviderPreset p : PRESETS) {
            for (String prefix : p.modelPrefixes) {
                if (m.startsWith(prefix.toLowerCase())) return p;
            }
        }

        return null; // 未匹配
    }

    /** 按名称查找预设。 */
    public static ProviderPreset findByName(String name) {
        return PRESETS.stream()
            .filter(p -> p.name.equalsIgnoreCase(name))
            .findFirst().orElse(null);
    }

    /** 列出所有预设供 /model 命令使用。 */
    public static List<ProviderPreset> all() {
        return PRESETS;
    }

    /** 解析 API 密钥：按预设 envVar → OPENAI_API_KEY → ANTHROPIC_API_KEY 顺序查找。 */
    public static String resolveApiKey(ProviderPreset preset, String explicitKey) {
        if (explicitKey != null && !explicitKey.isBlank()) return explicitKey;
        if (preset != null && preset.envVar != null) {
            String val = System.getenv(preset.envVar);
            if (val != null) return val;
        }
        String key = System.getenv("OPENAI_API_KEY");
        if (key != null) return key;
        return System.getenv("ANTHROPIC_API_KEY");
    }
}
