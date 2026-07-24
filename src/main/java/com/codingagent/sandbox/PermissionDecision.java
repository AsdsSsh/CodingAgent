package com.codingagent.sandbox;

/**
 * 工具调用权限检查的结果。
 *
 * <p>三种可能的裁定：</p>
 * <ul>
 *   <li>{@link Verdict#ALLOW} —— 允许继续执行</li>
 *   <li>{@link Verdict#DENY} —— 已阻止；Agent 收到错误观察结果</li>
 *   <li>{@link Verdict#PROMPT} —— 需要用户升级权限；Agent 收到提示，引导其重新请求</li>
 * </ul>
 */
public record PermissionDecision(Verdict verdict, String reason) {

    public enum Verdict { ALLOW, DENY, PROMPT }

    public static PermissionDecision allow(String reason) {
        return new PermissionDecision(Verdict.ALLOW, reason);
    }

    public static PermissionDecision deny(String reason) {
        return new PermissionDecision(Verdict.DENY, reason);
    }

    public static PermissionDecision prompt(String reason) {
        return new PermissionDecision(Verdict.PROMPT, reason);
    }
}
