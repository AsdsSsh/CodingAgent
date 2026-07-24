package com.codingagent.config;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;

/**
 * 根配置对象，采用分层加载策略：
 * <ol>
 *   <li>代码内硬编码的默认值</li>
 *   <li>全局配置：{@code ~/.coding-agent/config.yaml}</li>
 *   <li>项目级配置：{@code .coding-agent/config.yaml}</li>
 *   <li>环境变量（ANTHROPIC_API_KEY / OPENAI_API_KEY）</li>
 *   <li>CLI 参数（-p / -m / -k / -t / -w）</li>
 * </ol>
 * 后一层逐字段覆盖前一层。
 */
public record AgentConfig(
    @JsonProperty("model") ModelConfig model,
    @JsonProperty("mcp_servers") List<McpServerConfig> mcpServers,
    @JsonProperty("permissions") PermissionConfig permissions,
    @JsonProperty("workspace") String workspace,
    @JsonProperty("max_iterations") int maxIterations,
    @JsonProperty("memory_path") String memoryPath,
    @JsonProperty("sandbox") SandboxConfig sandbox
) {
    private static final ObjectMapper YAML_MAPPER = new ObjectMapper(new YAMLFactory());

    public record SandboxConfig(
        @JsonProperty("type") String type,
        @JsonProperty("timeout_seconds") int timeoutSeconds,
        @JsonProperty("blocked_commands") List<String> blockedCommands
    ) {
        public static SandboxConfig defaults() {
            return new SandboxConfig("subprocess", 120,
                List.of("rm -rf /", "sudo", "dd", "mkfs", ":(){ :|:& };:"));
        }
    }

    /** 返回硬编码的默认配置。 */
    public static AgentConfig defaults() {
        return new AgentConfig(
            ModelConfig.defaults(),
            List.of(),
            PermissionConfig.defaults(),
            System.getProperty("user.dir"),
            50,
            Path.of(System.getProperty("user.home"), ".coding-agent", "memory.jsonl").toString(),
            SandboxConfig.defaults()
        );
    }

    /** 从指定的 YAML 文件路径加载配置。文件缺失时回退到默认值。 */
    public static AgentConfig load(Path configPath) throws IOException {
        if (!Files.exists(configPath)) {
            return defaults();
        }
        String content = Files.readString(configPath);
        AgentConfig fileConfig = YAML_MAPPER.readValue(content, AgentConfig.class);
        return merge(defaults(), fileConfig);
    }

    /**
     * 完整的分层加载：全局 → 项目 → 环境变量。
     * 项目配置覆盖全局配置；环境变量覆盖以上两者。
     */
    public static AgentConfig load() throws IOException {
        Path globalPath = Path.of(System.getProperty("user.home"), ".coding-agent", "config.yaml");
        AgentConfig global = Files.exists(globalPath) ? load(globalPath) : defaults();

        Path projectPath = Path.of(".coding-agent", "config.yaml");
        AgentConfig project = Files.exists(projectPath) ? load(projectPath) : global;

        return applyEnvOverrides(project);
    }

    /** 深度合并 ModelConfig：override 的非 null 字段覆盖 base。 */
    private static ModelConfig mergeModel(ModelConfig base, ModelConfig override) {
        if (override == null) return base;
        if (base == null) return override;
        return new ModelConfig(
            override.provider() != null ? override.provider() : base.provider(),
            override.model() != null ? override.model() : base.model(),
            override.apiKey() != null ? override.apiKey() : base.apiKey(),
            override.apiKeyEnv() != null ? override.apiKeyEnv() : base.apiKeyEnv(),
            override.baseUrl() != null ? override.baseUrl() : base.baseUrl(),
            override.maxTokens() > 0 ? override.maxTokens() : base.maxTokens(),
            override.contextWindow() > 0 ? override.contextWindow() : base.contextWindow(),
            override.temperature() != null ? override.temperature() : base.temperature()
        );
    }

    /** 合并两个配置：{@code override} 中的非空/非空列表字段会覆盖 base。 */
    private static AgentConfig merge(AgentConfig base, AgentConfig override) {
        ModelConfig model = mergeModel(base.model, override.model);
        List<McpServerConfig> mcp = override.mcpServers != null && !override.mcpServers.isEmpty()
            ? override.mcpServers : base.mcpServers;
        PermissionConfig perm = override.permissions != null ? override.permissions : base.permissions;
        String ws = override.workspace != null ? override.workspace : base.workspace;
        int maxIter = override.maxIterations > 0 ? override.maxIterations : base.maxIterations;
        String memPath = override.memoryPath != null ? override.memoryPath : base.memoryPath;
        SandboxConfig sandbox = override.sandbox != null ? override.sandbox : base.sandbox;
        return new AgentConfig(model, mcp, perm, ws, maxIter, memPath, sandbox);
    }

    /** 若配置中未设置 API 密钥，则从环境变量注入。 */
    private static AgentConfig applyEnvOverrides(AgentConfig config) {
        String apiKey = System.getenv("ANTHROPIC_API_KEY");
        if (apiKey != null && config.model != null && !config.model.provider().equals("openai")) {
            config = new AgentConfig(
                config.model.withApiKey(apiKey), config.mcpServers,
                config.permissions, config.workspace, config.maxIterations,
                config.memoryPath, config.sandbox
            );
        }
        return config;
    }

    /** CLI 覆盖：替换整个模型配置。 */
    public AgentConfig withModelOverride(ModelConfig newModel) {
        return new AgentConfig(newModel, mcpServers, permissions, workspace,
            maxIterations, memoryPath, sandbox);
    }

    /** REPL /permission 命令：替换权限级别。 */
    public AgentConfig withPermissionLevel(com.codingagent.sandbox.PermissionLevel level) {
        PermissionConfig newPerm = new PermissionConfig(
            level.name(), permissions.autoApprove(),
            permissions.maxFrequency(), permissions.frequencyWindowSeconds()
        );
        return new AgentConfig(model, mcpServers, newPerm, workspace,
            maxIterations, memoryPath, sandbox);
    }

    /** CLI 覆盖：替换工作区路径。 */
    public AgentConfig withWorkspace(String newWorkspace) {
        return new AgentConfig(model, mcpServers, permissions, newWorkspace,
            maxIterations, memoryPath, sandbox);
    }

    /**
     * 解析 API 密钥：配置文件优先，其次查找各提供商对应的环境变量。
     */
    public String resolveApiKey() {
        return model.resolveApiKey();
    }
}
