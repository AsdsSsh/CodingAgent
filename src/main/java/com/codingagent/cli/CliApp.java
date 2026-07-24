package com.codingagent.cli;

import com.codingagent.config.AgentConfig;
import com.codingagent.config.ModelConfig;
import picocli.CommandLine.Command;
import picocli.CommandLine.Option;

import java.nio.file.Path;
import java.util.concurrent.Callable;

/**
 * CLI 入口——直接启动交互式对话 REPL，类似 Claude Code 的体验。
 *
 * <pre>
 *   coding-agent                启动对话（使用默认配置）
 *   coding-agent -m gpt-4o      切换模型启动
 *   coding-agent -w /my/project 指定工作区
 * </pre>
 */
@Command(
    name = "coding-agent",
    description = "AI Coding Agent - 交互式编程助手",
    mixinStandardHelpOptions = true
)
public class CliApp implements Callable<Integer> {

    @Option(names = {"-c", "--config"}, description = "配置文件路径")
    private Path configPath;

    @Option(names = {"-p", "--provider"}, description = "LLM 提供商: anthropic, openai, custom")
    private String provider;

    @Option(names = {"-m", "--model"}, description = "模型名称")
    private String model;

    @Option(names = {"-k", "--api-key"}, description = "API 密钥（覆盖环境变量和配置文件）")
    private String apiKey;

    @Option(names = {"-t", "--temperature"}, description = "温度: 0.0=精确, 1.0=创意")
    private Double temperature;

    @Option(names = {"-w", "--workspace"}, description = "工作区目录")
    private Path workspace;

    @Override
    public Integer call() {
        try {
            AgentConfig config = configPath != null
                ? AgentConfig.load(configPath) : AgentConfig.load();

            if (workspace != null) {
                config = config.withWorkspace(workspace.toAbsolutePath().toString());
            }
            if (provider != null || model != null || apiKey != null || temperature != null) {
                config = config.withModelOverride(
                    applyCliOverrides(config.model(), provider, model, apiKey, temperature)
                );
            }

            Repl repl = new Repl(config);
            repl.start();
            return 0;
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
            return 1;
        }
    }

    private static ModelConfig applyCliOverrides(
        ModelConfig base, String provider, String model, String apiKey, Double temperature
    ) {
        ModelConfig result = base;
        if (provider != null) result = result.withProvider(provider);
        if (model != null) result = result.withModel(model);
        if (apiKey != null) result = result.withApiKey(apiKey);
        if (temperature != null) {
            result = new ModelConfig(
                result.provider(), result.model(), result.apiKey(), result.apiKeyEnv(),
                result.baseUrl(), result.maxTokens(), result.contextWindow(), temperature
            );
        }
        return result;
    }

    public static void main(String[] args) {
        int exitCode = new picocli.CommandLine(new CliApp()).execute(args);
        System.exit(exitCode);
    }
}
