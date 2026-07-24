package com.codingagent.config;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.Map;

/**
 * 权限策略配置。
 *
 * @param defaultLevel            初始权限级别（NONE/READ/WRITE/EXECUTE/DANGEROUS）
 * @param autoApprove             可跳过弹窗的工具，按工具名索引
 * @param maxFrequency            时间窗口内的最大使用次数，超出后重新弹窗确认
 * @param frequencyWindowSeconds  频率统计的时间窗口（秒）
 */
public record PermissionConfig(
    @JsonProperty("default_level") String defaultLevel,
    @JsonProperty("auto_approve") Map<String, Boolean> autoApprove,
    @JsonProperty("max_frequency") int maxFrequency,
    @JsonProperty("frequency_window_seconds") int frequencyWindowSeconds
) {
    /** 默认值：READ 级别，无自动批准，每 5 分钟内最多 10 次后重新提示。 */
    public static PermissionConfig defaults() {
        return new PermissionConfig("READ", Map.of(), 10, 300);
    }
}
