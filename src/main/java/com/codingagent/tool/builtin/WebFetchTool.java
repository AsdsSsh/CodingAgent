package com.codingagent.tool.builtin;

import com.codingagent.sandbox.PermissionLevel;
import com.codingagent.tool.ToolBase;
import com.codingagent.tool.ToolContext;
import com.codingagent.tool.ToolResult;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.List;
import java.util.Map;

/**
 * 抓取并清洗 URL 内容。HTML 标签被剥离以提高可读性；
 * 输出上限 50K 字符。
 *
 * <p>使用 {@link HttpClient}，启用重定向跟随，30 秒超时。
 * 只读操作——需要 {@link PermissionLevel#READ} 权限。</p>
 */
public class WebFetchTool implements ToolBase {

    private final HttpClient httpClient = HttpClient.newBuilder()
        .connectTimeout(Duration.ofSeconds(15))
        .followRedirects(HttpClient.Redirect.NORMAL)
        .build();

    @Override
    public String name() { return "web_fetch"; }

    @Override
    public String description() {
        return "获取 URL 的内容并以文本形式返回。适用于阅读文档或 API 响应。";
    }

    @Override
    @SuppressWarnings("unchecked")
    public Map<String, Object> parameters() {
        return Map.of(
            "type", "object",
            "properties", Map.of(
                "url", Map.of("type", "string", "description", "要获取的 URL"),
                "prompt", Map.of("type", "string", "description", "要从页面中提取的信息内容")
            ),
            "required", List.of("url")
        );
    }

    @Override
    public PermissionLevel requiredPermission() { return PermissionLevel.READ; }

    @Override
    public ToolResult execute(ToolContext ctx, Map<String, Object> args) {
        String urlStr = (String) args.get("url");

        try {
            HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(urlStr))
                .header("User-Agent", "CodingAgent/1.0")
                .timeout(Duration.ofSeconds(30))
                .GET()
                .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() != 200) {
                return ToolResult.fail("HTTP " + response.statusCode() + " for URL: " + urlStr);
            }

            String body = response.body();
            // 剥离 HTML 标签并压缩空白，提高可读性
            String cleaned = body.replaceAll("<[^>]+>", " ").replaceAll("\\s+", " ").trim();
            if (cleaned.length() > 50_000) {
                cleaned = cleaned.substring(0, 50_000) + "\n... [已截断]";
            }

            return ToolResult.ok(cleaned, body.length() > 50_000);
        } catch (Exception e) {
            return ToolResult.fail("获取 URL 出错: " + e.getMessage());
        }
    }
}
