package com.codingagent.config;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;
import java.util.Map;

/**
 * 单个 MCP（模型上下文协议）服务器的配置。
 * 支持 stdio（子进程）和 HTTP 两种传输方式。
 * MCP 服务器的工具采用延迟加载——启动时仅缓存工具定义，
 * 实际的服务器连接推迟到首次工具调用时才建立。
 *
 * @param name      该服务器的可读名称
 * @param transport 传输方式："stdio" 或 "http"
 * @param command   启动命令（stdio 传输）
 * @param args      启动命令的参数
 * @param url       HTTP 端点地址（http 传输）
 * @param env       服务器进程的额外环境变量
 */
public record McpServerConfig(
    @JsonProperty("name") String name,
    @JsonProperty("transport") String transport,
    @JsonProperty("command") String command,
    @JsonProperty("args") List<String> args,
    @JsonProperty("url") String url,
    @JsonProperty("env") Map<String, String> env
) {
    public static McpServerConfig defaults() {
        return new McpServerConfig("default", "stdio", null, List.of(), null, Map.of());
    }
}
