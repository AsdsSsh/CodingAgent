package com.codingagent.sandbox;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 中央权限决策引擎。
 *
 * <p>三层检查（首个匹配即返回）：</p>
 * <ol>
 *   <li><b>永久用户覆盖</b> —— 由用户在任何时刻显式允许/拒绝</li>
 *   <li><b>权限级别足够 + 频率检查</b> —— 若当前级别覆盖所需级别，
 *       检查该工具近期是否被调用过多次（防审批疲劳）</li>
 *   <li><b>权限级别不足</b> —— 提示用户升级</li>
 * </ol>
 *
 * <p>基于频率的重新弹窗：在 {@code windowMs} 内使用超过 {@code maxFrequency} 次后，
 * 该工具会触发重新授权提示。这防止 Agent 在用户不知情的情况下静默运行数百条命令。</p>
 *
 * <p>线程安全：所有状态位于并发集合中；check/recordUse 不是原子操作
 *（对于频率计数场景可接受）。</p>
 */
public class PolicyEngine {

    /** 用户对特定工具设置的永久允许/拒绝。 */
    private final Map<String, Boolean> userOverrides = new ConcurrentHashMap<>();
    /** 滑动窗口内的工具使用时间戳。 */
    private final Map<String, List<Long>> useTimestamps = new ConcurrentHashMap<>();
    private PermissionLevel currentLevel;
    private final int maxFrequency;
    private final long windowMs;

    public PolicyEngine(PermissionLevel currentLevel, int maxFrequency, long windowMs) {
        this.currentLevel = currentLevel;
        this.maxFrequency = maxFrequency;
        this.windowMs = windowMs;
    }

    /** 便捷构造器，默认频率限制：5 分钟内 10 次。 */
    public PolicyEngine(PermissionLevel currentLevel) {
        this(currentLevel, 10, 300_000);
    }

    /**
     * 检查工具调用应被允许、拒绝还是需要弹窗确认。
     * 在 ReAct 循环的每次工具执行前调用。
     */
    public PermissionDecision check(String toolName, PermissionLevel required) {
        // 第一层：永久覆盖
        Boolean override = userOverrides.get(toolName);
        if (override != null) {
            if (override) {
                recordUse(toolName);
                return PermissionDecision.allow("permanently allowed by user");
            }
            return PermissionDecision.deny("permanently denied by user");
        }

        // 第二层：权限级别足够 → 检查频率
        if (currentLevel.covers(required)) {
            if (shouldPrompt(toolName)) {
                return PermissionDecision.prompt(
                    "工具 '" + toolName + "' 在过去 " + (windowMs / 1000) +
                    " 秒内已使用 " + maxFrequency + " 次，请重新授权？"
                );
            }
            recordUse(toolName);
            return PermissionDecision.allow("permission level " + currentLevel);
        }

        // 第三层：权限级别不足
        return PermissionDecision.prompt(
            "工具 '" + toolName + "' 需要 " + required +
            " 级别，当前级别为 " + currentLevel
        );
    }

    /** 在当前窗口内，该工具是否已超频。 */
    private boolean shouldPrompt(String toolName) {
        List<Long> timestamps = useTimestamps.get(toolName);
        if (timestamps == null) return false;

        long cutoff = System.currentTimeMillis() - windowMs;
        long recentCount = timestamps.stream().filter(t -> t > cutoff).count();
        return recentCount >= maxFrequency;
    }

    /** 记录一次工具使用时间戳，供频率追踪使用。 */
    private void recordUse(String toolName) {
        useTimestamps.computeIfAbsent(toolName, k -> new ArrayList<>())
            .add(System.currentTimeMillis());
    }

    /** 提升当前权限级别（单调递增——不可降级）。 */
    public void escalate(PermissionLevel newLevel) {
        if (newLevel.level() > this.currentLevel.level()) {
            this.currentLevel = newLevel;
        }
    }

    /** 永久允许或拒绝某个特定工具。 */
    public void addOverride(String toolName, boolean allow) {
        userOverrides.put(toolName, allow);
    }

    /** 移除永久覆盖，恢复正常的策略判定。 */
    public void removeOverride(String toolName) {
        userOverrides.remove(toolName);
    }

    public PermissionLevel currentLevel() { return currentLevel; }
}
