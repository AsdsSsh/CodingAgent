package com.codingagent.tool;

import com.codingagent.sandbox.SandboxBase;

import java.nio.file.Path;

/**
 * 每次工具调用时传入的执行上下文。
 * 携带工作区根路径和沙箱引用，使工具可以解析相对路径
 * 并将命令执行委托给沙箱。
 */
public record ToolContext(Path workspace, SandboxBase sandbox) {

    /** 解析用户提供的路径：绝对路径直接通过，相对路径相对于工作区。 */
    public Path resolvePath(String path) {
        Path p = Path.of(path);
        if (p.isAbsolute()) return p;
        return workspace.resolve(p).normalize();
    }

    /** 检查解析后的路径是否在工作区边界内（路径穿越防护）。 */
    public boolean isInWorkspace(Path path) {
        Path resolved = path.toAbsolutePath().normalize();
        Path ws = workspace.toAbsolutePath().normalize();
        return resolved.startsWith(ws);
    }
}
