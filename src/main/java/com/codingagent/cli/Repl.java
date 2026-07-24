package com.codingagent.cli;

import com.codingagent.agent.AgentResult;
import com.codingagent.agent.Orchestrator;
import com.codingagent.config.AgentConfig;
import com.codingagent.config.ModelConfig;
import com.codingagent.llm.ModelRegistry;
import com.codingagent.sandbox.PermissionLevel;
import org.jline.reader.EndOfFileException;
import org.jline.reader.LineReader;
import org.jline.reader.LineReaderBuilder;
import org.jline.reader.UserInterruptException;
import org.jline.terminal.Terminal;
import org.jline.terminal.TerminalBuilder;

import java.io.IOException;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

/**
 * 交互式 REPL，类似 Claude Code 的对话体验。
 * 支持多行输入、命令历史、斜杠命令、跨轮次上下文和实时权限切换。
 */
public class Repl {

    private static final String PROMPT = "> ";
    private static final String CONTINUE_PROMPT = "… ";

    private AgentConfig config;
    private PermissionLevel currentLevel;
    private final List<String> history = new ArrayList<>();

    public Repl(AgentConfig config) {
        this.config = config;
        this.currentLevel = PermissionLevel.fromString(config.permissions().defaultLevel());
    }

    public void start() throws IOException {
        Terminal terminal = TerminalBuilder.builder()
            .system(true)
            .build();

        LineReader reader = LineReaderBuilder.builder()
            .terminal(terminal)
            .variable(LineReader.HISTORY_FILE,
                Path.of(System.getProperty("user.home"), ".coding-agent", "repl_history"))
            .build();

        String modelInfo = config.model().model();
        ModelRegistry.ProviderPreset preset = ModelRegistry.detect(config.model().model());
        if (preset != null) modelInfo += " (" + preset.name() + ")";

        System.out.println();
        System.out.println("╔══════════════════════════════════════════════╗");
        System.out.println("║   CodingAgent REPL                           ║");
        System.out.println("║   输入任务开始对话  |  /help 查看帮助         ║");
        System.out.println("║   模型: " + padRight(modelInfo, 36) + "║");
        System.out.println("║   权限: " + padRight(currentLevel.name(), 36) + "║");
        System.out.println("╚══════════════════════════════════════════════╝");
        System.out.println();

        while (true) {
            try {
                String input = readMultiLine(reader);
                if (input == null || input.isBlank()) continue;

                // 斜杠命令
                if (input.startsWith("/")) {
                    if (handleCommand(input)) break;
                    continue;
                }

                // 执行任务（使用当前权限级别）
                System.out.println("⏳ 思考中…");
                System.out.println();

                String taskWithContext = buildTaskWithContext(input);
                AgentConfig taskConfig = config.withPermissionLevel(currentLevel);
                Orchestrator orchestrator = new Orchestrator(taskConfig);
                AgentResult result = orchestrator.run(taskWithContext);

                Formatter formatter = new Formatter();
                System.out.println(formatter.format(result));

                // 保存到历史
                history.add("👤 用户: " + input);
                history.add("🤖 助手: " + result.answer());

            } catch (UserInterruptException e) {
                System.out.println("^C");
            } catch (EndOfFileException e) {
                System.out.println("/exit");
                break;
            }
        }

        System.out.println("再见！");
    }

    /** 读取多行输入：行尾 \\ 续行。 */
    private String readMultiLine(LineReader reader) {
        StringBuilder sb = new StringBuilder();
        String firstLine = reader.readLine(PROMPT);
        if (firstLine == null) return null;

        // 开头空格视为退出续行
        if (firstLine.endsWith("\\")) {
            sb.append(firstLine, 0, firstLine.length() - 1);
            while (true) {
                String next = reader.readLine(CONTINUE_PROMPT);
                if (next == null) break;
                if (next.endsWith("\\")) {
                    sb.append(next, 0, next.length() - 1);
                } else {
                    sb.append(next);
                    break;
                }
            }
            return sb.toString();
        }
        return firstLine;
    }

    /** 处理 / 开头的斜杠命令。返回 true 表示退出。 */
    private boolean handleCommand(String input) {
        String[] parts = input.split("\\s+", 2);
        String cmd = parts[0].toLowerCase();

        switch (cmd) {
            case "/exit", "/quit", "/q" -> {
                return true;
            }
            case "/help", "/h" -> {
                System.out.println("""

                    📖 可用命令:
                      /help         显示此帮助
                      /model        查看或切换模型（自动识别端点）
                      /baseurl      设置自定义 API 端点
                      /permission   查看或切换权限级别
                      /clear        清除对话历史
                      /history      显示对话历史
                      /exit         退出 REPL

                    🤖 模型切换:
                      /model                  列出所有内置预设
                      /model deepseek-v4-pro  自动识别 DeepSeek
                      /model gpt-4o           自动识别 OpenAI
                      /model claude-sonnet-4  自动识别 Anthropic
                      /model qwen2.5:7b       需配合 /baseurl http://localhost:11434/v1

                    🔐 权限级别:
                      /permission read / write / execute / dangerous

                    💡 多行输入: 在行尾加 \\ 续行
                    💡 Ctrl+C     中断当前任务
                    💡 Ctrl+D     退出 REPL
                    """);
            }
            case "/model", "/m" -> {
                if (parts.length < 2) {
                    System.out.println("当前模型: " + config.model().model());
                    System.out.println("用法: /model <模型名>");
                    System.out.println("内置预设 (自动识别端点):");
                    for (var p : ModelRegistry.all()) {
                        System.out.printf("  %-12s %-25s (密钥: %s)%n",
                            p.name(), p.defaultModel(),
                            p.envVar() != null ? p.envVar() : "无需");
                    }
                    System.out.println("也可直接输入任何 OpenAI 兼容模型名 + /baseurl 指定端点");
                } else {
                    String newModel = parts[1];
                    config = config.withModelOverride(config.model().withModel(newModel));
                    var detected = ModelRegistry.detect(newModel);
                    if (detected != null) {
                        System.out.println("已切换: " + newModel + " → " + detected.name()
                            + " (" + detected.baseUrl() + ")");
                    } else {
                        System.out.println("已切换: " + newModel + " (未匹配预设，若为 OpenAI 兼容端点请用 /baseurl 设置)");
                    }
                }
            }
            case "/baseurl" -> {
                if (parts.length < 2) {
                    System.out.println("当前 base_url: " + config.model().baseUrl());
                    System.out.println("用法: /baseurl <URL>   (设置 OpenAI 兼容端点)");
                } else {
                    config = config.withModelOverride(new ModelConfig(
                        config.model().provider(), config.model().model(),
                        config.model().apiKey(), config.model().apiKeyEnv(),
                        parts[1], config.model().maxTokens(), config.model().contextWindow(),
                        config.model().temperature()
                    ));
                    System.out.println("base_url 已设置: " + parts[1]);
                }
            }
            case "/permission", "/perm", "/p" -> {
                if (parts.length < 2) {
                    System.out.println("当前权限: " + currentLevel);
                    System.out.println("用法: /permission <read|write|execute|dangerous>");
                } else {
                    try {
                        PermissionLevel newLevel = PermissionLevel.fromString(parts[1].toUpperCase());
                        currentLevel = newLevel;
                        System.out.println("权限已切换为: " + currentLevel);
                    } catch (IllegalArgumentException e) {
                        System.out.println("无效的权限级别。可选: read, write, execute, dangerous");
                    }
                }
            }
            case "/clear" -> {
                history.clear();
                System.out.println("对话历史已清除。");
            }
            case "/history" -> {
                if (history.isEmpty()) {
                    System.out.println("（无对话历史）");
                } else {
                    System.out.println("\n--- 对话历史 ---");
                    for (String h : history) {
                        System.out.println(h);
                        System.out.println();
                    }
                    System.out.println("--- 共 " + history.size() / 2 + " 轮 ---");
                }
            }
            default -> {
                System.out.println("未知命令: " + cmd + "（输入 /help 查看帮助）");
            }
        }
        return false;
    }

    /** 将之前的对话历史作为上下文追加到当前任务。 */
    private String buildTaskWithContext(String task) {
        if (history.isEmpty()) return task;

        StringBuilder sb = new StringBuilder();
        sb.append("以下是我们之前的对话记录，请参考这些上下文：\n\n");
        for (int i = Math.max(0, history.size() - 10); i < history.size(); i++) {
            sb.append(history.get(i)).append("\n\n");
        }
        sb.append("---\n");
        sb.append("当前任务: ").append(task);
        return sb.toString();
    }

    private static String padRight(String s, int n) {
        if (s.length() >= n) return s;
        return s + " ".repeat(n - s.length());
    }
}
