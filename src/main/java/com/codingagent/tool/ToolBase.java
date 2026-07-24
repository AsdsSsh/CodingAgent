package com.codingagent.tool;

import com.codingagent.sandbox.PermissionLevel;

import java.util.Map;

/**
 * Agent 可用所有工具（内置和 MCP）的统一契约。
 *
 * <p>每个工具声明：</p>
 * <ul>
 *   <li>{@link #name()} —— LLM 工具调用中使用的唯一标识</li>
 *   <li>{@link #description()} —— 供 LLM 理解的自然语言描述</li>
 *   <li>{@link #parameters()} —— 参数的 JSON Schema</li>
 *   <li>{@link #requiredPermission()} —— 所需最低权限级别（默认为 READ）</li>
 * </ul>
 *
 * <p>{@link #toLlmFormat()} 生成与提供商无关的工具定义，
 * 随每次 LLM 请求一起发送。MCP 工具遵循同一接口，因此
 * ReAct 循环无需知道工具的具体来源。</p>
 */
public interface ToolBase {

    String name();
    String description();
    Map<String, Object> parameters();

    /** 所需的最低权限级别。破坏性工具需覆盖此方法。 */
    default PermissionLevel requiredPermission() {
        return PermissionLevel.READ;
    }

    /** 执行工具。沙箱在调用此方法前已执行权限检查。 */
    ToolResult execute(ToolContext ctx, Map<String, Object> args);

    /** 与提供商无关的工具定义，供 LLM API 调用使用。 */
    default Map<String, Object> toLlmFormat() {
        return Map.of(
            "name", name(),
            "description", description(),
            "input_schema", parameters()
        );
    }
}
