package com.codingagent.sandbox;

import com.codingagent.tool.ToolResult;

/**
 * 执行环境的抽象。计划提供两种实现：
 * <ul>
 *   <li>{@link SubprocessSandbox} —— 带超时、路径约束和命令黑名单的子进程</li>
 *   <li>{@code DockerSandbox}（Phase 6）—— 完整操作系统级隔离的 Docker 容器</li>
 * </ul>
 */
public interface SandboxBase {
    /** 该沙箱实例的权限引擎。 */
    PolicyEngine policyEngine();
    /** 在沙箱约束下执行 Shell 命令。 */
    ToolResult executeCommand(String command);
    /** 检查路径是否在允许的工作区范围内。 */
    boolean isInWorkspace(String path);
}
