package com.codingagent.sandbox;

import com.codingagent.tool.ToolResult;

import java.io.IOException;
import java.nio.file.Path;
import java.util.List;
import java.util.concurrent.TimeUnit;

/**
 * 默认沙箱，使用操作系统子进程并施加资源限制。
 *
 * <p>安全机制：</p>
 * <ul>
 *   <li><b>命令黑名单</b> —— 阻止已知的危险模式（sudo、rm -rf /、fork 炸弹）</li>
 *   <li><b>超时</b> —— 终止超过配置时限的进程</li>
 *   <li><b>路径限制</b> —— 所有操作限定在工作区目录内</li>
 *   <li><b>跨平台</b> —— Windows 上使用 cmd，Unix 上使用 bash</li>
 * </ul>
 *
 * <p>这不是真正的安全边界（子进程与宿主机共享操作系统）。
 * 要执行不受信任的代码，请使用 {@code DockerSandbox}（Phase 6）。</p>
 */
public class SubprocessSandbox implements SandboxBase {

    private final Path workspace;
    private final int timeoutSeconds;
    private final List<String> blockedCommands;
    private final PolicyEngine policyEngine;

    /** 无论权限级别如何，始终被阻止的命令。 */
    private static final List<String> DEFAULT_BLOCKED = List.of(
        "rm -rf /", "sudo", "dd if=", "mkfs", ":(){ :|:& };:",
        "shutdown", "reboot", "chmod 777 /"
    );

    public SubprocessSandbox(Path workspace, int timeoutSeconds,
                             List<String> blockedCommands, PermissionLevel permissionLevel) {
        this.workspace = workspace.toAbsolutePath().normalize();
        this.timeoutSeconds = timeoutSeconds;
        this.blockedCommands = blockedCommands != null ? blockedCommands : DEFAULT_BLOCKED;
        this.policyEngine = new PolicyEngine(permissionLevel);
    }

    @Override
    public PolicyEngine policyEngine() { return policyEngine; }

    @Override
    public ToolResult executeCommand(String command) {
        // 生成进程前先做黑名单检查
        for (String blocked : blockedCommands) {
            if (command.contains(blocked)) {
                return ToolResult.fail("Command blocked by security policy: contains '" + blocked + "'");
            }
        }

        try {
            ProcessBuilder pb = new ProcessBuilder();
            // 跨平台：Windows 用 cmd，其他用 bash
            if (System.getProperty("os.name").toLowerCase().contains("win")) {
                pb.command("cmd", "/c", command);
            } else {
                pb.command("bash", "-c", command);
            }
            pb.directory(workspace.toFile());
            pb.redirectErrorStream(true); // 将 stderr 合并到 stdout

            Process process = pb.start();
            boolean finished = process.waitFor(timeoutSeconds, TimeUnit.SECONDS);

            if (!finished) {
                process.destroyForcibly();
                return ToolResult.fail("Command timed out after " + timeoutSeconds + " seconds");
            }

            String output = new String(process.getInputStream().readAllBytes());
            return ToolResult.fromCommandOutput(process.exitValue(), output, "");
        } catch (IOException e) {
            return ToolResult.fail("Command execution error: " + e.getMessage());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return ToolResult.fail("Command interrupted");
        }
    }

    @Override
    public boolean isInWorkspace(String path) {
        Path resolved = workspace.resolve(path).normalize().toAbsolutePath();
        return resolved.startsWith(workspace);
    }

    public Path workspace() { return workspace; }
}
