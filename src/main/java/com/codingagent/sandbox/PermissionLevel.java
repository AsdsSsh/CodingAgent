package com.codingagent.sandbox;

/**
 * 安全沙箱的单调递增权限级别。
 *
 * <p>五个级别遵循最小权限原则：</p>
 * <ol>
 *   <li>{@link #NONE} —— 仅观察，不可执行任何工具</li>
 *   <li>{@link #READ} —— 文件读取、网络请求、搜索</li>
 *   <li>{@link #WRITE} —— 文件创建和修改</li>
 *   <li>{@link #EXECUTE} —— Shell 命令、子进程</li>
 *   <li>{@link #DANGEROUS} —— 系统更改、软件包安装</li>
 * </ol>
 *
 * <p>高级别隐式涵盖低级别（例如 EXECUTE 也允许 READ 和 WRITE）。
 * Agent 从配置的默认级别启动，仅在需要时升级。</p>
 */
public enum PermissionLevel {
    NONE(0),
    READ(1),
    WRITE(2),
    EXECUTE(3),
    DANGEROUS(4);

    private final int level;

    PermissionLevel(int level) { this.level = level; }

    public int level() { return level; }

    /** 当前级别是否涵盖了所需级别。 */
    public boolean covers(PermissionLevel required) {
        return this.level >= required.level;
    }

    /** 从配置字符串解析；未知值时默认返回 READ。 */
    public static PermissionLevel fromString(String s) {
        return switch (s.toUpperCase()) {
            case "NONE" -> NONE;
            case "READ" -> READ;
            case "WRITE" -> WRITE;
            case "EXECUTE" -> EXECUTE;
            case "DANGEROUS" -> DANGEROUS;
            default -> READ;
        };
    }
}
