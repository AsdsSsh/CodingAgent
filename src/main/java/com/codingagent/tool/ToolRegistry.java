package com.codingagent.tool;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Agent 可用所有工具的中央注册表。
 *
 * <p>内置工具在启动时立即注册。MCP 工具（Phase 3）
 * 将采用延迟注册——其定义被缓存以便出现在 LLM 的工具列表中，
 * 但 MCP 服务器连接推迟到首次实际调用时才建立。</p>
 *
 * <p>线程安全：使用 {@link ConcurrentHashMap}；工具可以在
 * 任何虚拟线程中注册和解析。</p>
 */
public class ToolRegistry {

    private final Map<String, ToolBase> tools = new ConcurrentHashMap<>();
    /** 跟踪 MCP 工具名称，供延迟加载路径使用（Phase 3）。 */
    private final List<String> mcpToolNames = new ArrayList<>();

    public void register(ToolBase tool) {
        tools.put(tool.name(), tool);
    }

    public void registerAll(List<ToolBase> toolList) {
        for (ToolBase tool : toolList) {
            register(tool);
        }
    }

    /** 将工具名称标记为由 MCP 服务器提供（延迟加载钩子）。 */
    public void registerMcpToolName(String name) {
        mcpToolNames.add(name);
    }

    /** 按名称解析工具。未找到时抛出异常——调用方应先调用 {@link #has} 检查。 */
    public ToolBase resolve(String name) {
        ToolBase tool = tools.get(name);
        if (tool == null) {
            throw new IllegalArgumentException("Tool not found: " + name);
        }
        return tool;
    }

    public boolean has(String name) {
        return tools.containsKey(name);
    }

    /** 生成完整的工具列表，供 LLM API 调用使用。 */
    public List<Map<String, Object>> toLlmFormats() {
        List<Map<String, Object>> formats = new ArrayList<>();
        for (ToolBase tool : tools.values()) {
            formats.add(tool.toLlmFormat());
        }
        return formats;
    }

    public int size() { return tools.size(); }
}
