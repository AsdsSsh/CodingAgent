package com.codingagent.agent;

import com.codingagent.config.AgentConfig;
import com.codingagent.llm.LlmProvider;
import com.codingagent.llm.ProviderFactory;
import com.codingagent.sandbox.PermissionLevel;
import com.codingagent.sandbox.SubprocessSandbox;
import com.codingagent.tool.ToolRegistry;
import com.codingagent.tool.builtin.*;

import java.nio.file.Path;

/**
 * 顶层编排器，串联全部五个层次：
 * <ol>
 *   <li>加载配置并解析 API 密钥</li>
 *   <li>创建 LLM 提供商（Anthropic / OpenAI）</li>
 *   <li>将所有内置工具注册到 {@link ToolRegistry}</li>
 *   <li>以配置的权限级别初始化沙箱</li>
 *   <li>构建系统提示并委托给 {@link ReActLoop} 执行</li>
 * </ol>
 *
 * <p>这是主要的扩展点：记忆检索、指令加载、
 * MCP 服务器注册均在此处接入。</p>
 */
public class Orchestrator {

    private final AgentConfig config;
    private final LlmProvider llm;
    private final ToolRegistry toolRegistry;
    private final SubprocessSandbox sandbox;
    private final ReActLoop reactLoop;
    private final EventBus eventBus;
    private final Path workspace;

    public Orchestrator(AgentConfig config) {
        this.config = config;
        this.eventBus = new EventBus();

        // ProviderFactory 负责密钥解析 + 模型自动检测 + 错误提示
        this.llm = ProviderFactory.create(config.model());
        this.toolRegistry = new ToolRegistry();
        registerBuiltinTools();

        this.workspace = Path.of(config.workspace()).toAbsolutePath().normalize();
        PermissionLevel permLevel = PermissionLevel.fromString(config.permissions().defaultLevel());
        this.sandbox = new SubprocessSandbox(
            workspace,
            config.sandbox().timeoutSeconds(),
            config.sandbox().blockedCommands(),
            permLevel
        );

        this.reactLoop = new ReActLoop(llm, toolRegistry, eventBus, config.maxIterations());
    }

    /** 注册全部六个内置工具。Phase 3 中将在此处添加 MCP 工具。 */
    private void registerBuiltinTools() {
        toolRegistry.register(new FileReadTool());
        toolRegistry.register(new FileWriteTool());
        toolRegistry.register(new FileEditTool());
        toolRegistry.register(new FileSearchTool());
        toolRegistry.register(new BashTool());
        toolRegistry.register(new WebFetchTool());
    }

    /**
     * 端到端执行一次编程任务。
     * Phase 5 中此方法还将：查询跨会话记忆 → 注入上下文 →
     * 完成后 → 提取并持久化新的记忆。
     */
    public AgentResult run(String task) {
        String systemPrompt = buildSystemPrompt();
        return reactLoop.run(task, systemPrompt, null, sandbox, workspace);
    }

    public EventBus eventBus() { return eventBus; }
    public LlmProvider llm() { return llm; }

    /** 构建简洁的系统提示，描述可用工具及期望行为。 */
    private String buildSystemPrompt() {
        return "You are an AI coding assistant. You help users with programming tasks using a ReAct architecture.\n"
            + "You have access to tools for reading, writing, editing, searching files, executing bash commands,\n"
            + "and fetching web content.\n\n"
            + "When you complete a task, provide a clear summary of what you did.\n"
            + "If you encounter errors, explain them and suggest alternatives.\n"
            + "Use the tools available to you to understand the codebase before making changes.";
    }

}
