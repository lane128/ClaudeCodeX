package i18n

import (
	"errors"
	"fmt"
)

const (
	English = "en"
	Chinese = "zh"
)

var catalog = map[string]map[string]string{
	English: {
		"usage_header":                "Usage:",
		"status_label":                "Status",
		"proxy_label":                 "Proxy",
		"dns_label":                   "DNS",
		"tcp_label":                   "TCP",
		"tls_label":                   "TLS",
		"http_label":                  "HTTP",
		"conclusion_label":            "Conclusion",
		"suggestion_label":            "Suggestion",
		"effective_proxy_label":       "Effective Proxy",
		"source_label":                "Source",
		"active_profile_label":        "Active Profile",
		"target_label":                "Target",
		"level_label":                 "Level",
		"signals_label":               "Signals",
		"language_label":              "Language",
		"shell_label":                 "Shell",
		"profile_label":               "Profile",
		"none":                        "none",
		"status_success":              "SUCCESS",
		"status_degraded":             "DEGRADED",
		"status_failed":               "FAILED",
		"status_skipped":              "SKIPPED",
		"status_detected":             "DETECTED",
		"status_not_detected":         "NOT_DETECTED",
		"level_low":                   "LOW",
		"level_medium":                "MEDIUM",
		"level_high":                  "HIGH",
		"language_current":            "Current language: %s",
		"language_updated":            "Language saved: %s",
		"language_choose_one":         "choose exactly one of --zh or --en",
		"language_unknown":            "unsupported language %q",
		"proxy_none_detected":         "none detected",
		"env_proxy_detected":          "proxy configuration detected",
		"env_proxy_missing":           "no proxy detected",
		"dns_resolved":                "resolved %d IP address(es)",
		"tcp_success":                 "tcp connection succeeded",
		"tls_success":                 "tls handshake succeeded (%s)",
		"http_success":                "http request succeeded with status %d",
		"http_degraded":               "http request reached endpoint but returned status %d",
		"dns_timeout":                 "dns resolution timed out",
		"dns_failed":                  "dns resolution failed",
		"network_timeout":             "network operation timed out",
		"http_timeout":                "http request timed out",
		"http_failed":                 "http request failed",
		"doctor_summary_success":      "all probes succeeded",
		"doctor_summary_degraded":     "%d probe(s) returned degraded results",
		"doctor_cause_dns":            "dns resolution failed before transport checks could complete",
		"doctor_cause_tcp":            "tcp connectivity failed, which often indicates routing, firewall, or proxy issues",
		"doctor_cause_tls":            "tls handshake failed after tcp connect succeeded",
		"doctor_cause_http":           "http request failed after lower-level checks, which may indicate proxy or upstream policy issues",
		"doctor_cause_generic":        "one or more diagnostic probes failed",
		"doctor_action_proxy":         "If you normally rely on a proxy, rerun with --proxy or set HTTPS_PROXY before invoking Claude Code.",
		"doctor_action_dns":           "Check your DNS resolver and compare results with a trusted public resolver to rule out polluted answers.",
		"doctor_action_tcp":           "Verify that your current network or proxy can establish outbound TCP connections to the target host on port 443.",
		"doctor_action_tls":           "Inspect TLS interception, enterprise certificates, or middleboxes that may be interrupting the handshake.",
		"doctor_action_http":          "Confirm the selected proxy can reach the target endpoint and that the upstream service is not rejecting your exit IP.",
		"doctor_action_none":          "No immediate action required.",
		"dns_summary_failed":          "dns resolution failed for one or more targets",
		"dns_summary_degraded":        "dns resolved, but some answers look suspicious",
		"dns_summary_success":         "dns results look normal",
		"dns_suggestion_failed":       "check the current resolver or compare with a trusted public DNS service",
		"dns_suggestion_degraded":     "compare returned IPs across different resolvers before changing your network settings",
		"dns_record_failed":           "lookup failed",
		"dns_record_degraded":         "resolved with suspicious private or loopback addresses",
		"dns_record_success":          "resolved successfully",
		"dns_host_required":           "host is required",
		"risk_low_dns":                "current failure looks more like a DNS or local network issue",
		"risk_low_dns_suggestion":     "fix name resolution first before judging whether the exit IP is being restricted",
		"risk_high_http":              "traffic reached the target path, but the final HTTP step failed",
		"risk_high_http_suggestion":   "try another proxy or exit IP and compare the result",
		"risk_medium_tls":             "tcp is reachable but tls failed, which can indicate filtering or interception",
		"risk_medium_tls_suggestion":  "check proxy egress quality and any tls interception in the current network",
		"risk_medium_tcp":             "connectivity failed before http, so this looks more like routing or firewall trouble",
		"risk_medium_tcp_suggestion":  "verify outbound 443 connectivity before treating this as an IP risk issue",
		"risk_low_none":               "no obvious IP or risk-control signal was detected",
		"env_no_proxy_render":         "no effective proxy is available to render shell exports",
		"env_saved_proxy":             "saved active proxy: %s",
		"env_cleared_proxy":           "cleared saved proxy configuration",
		"env_clear_missing":           "no saved proxy configuration to clear",
		"env_set_requires_proxy":      "provide --set proxy-url or the structured fields --type, --server, and --port",
		"env_set_invalid_port":        "port must be greater than 0",
		"proxy_port_label":            "Proxy Port",
		"expected_ip_label":           "Expected IP",
		"exit_ip_label":               "Exit IP",
		"proxy_port_failed":           "proxy port is not reachable: %s",
		"proxy_exit_ip_mismatch":      "exit IP did not match the expected value %s",
		"proxy_test_success":          "proxy connectivity and exit IP checks passed",
		"proxy_expected_ip_not_set":   "expected_ip is not set in settings.json; configure it to enable exit IP verification",
		"doctor_checking":             "checking %s (%s)...",
	},
	Chinese: {
		"usage_header":                "用法：",
		"status_label":                "状态",
		"proxy_label":                 "代理",
		"dns_label":                   "DNS",
		"tcp_label":                   "TCP",
		"tls_label":                   "TLS",
		"http_label":                  "HTTP",
		"conclusion_label":            "结论",
		"suggestion_label":            "建议",
		"effective_proxy_label":       "当前代理",
		"source_label":                "来源",
		"active_profile_label":        "激活配置",
		"target_label":                "目标",
		"level_label":                 "风险等级",
		"signals_label":               "信号",
		"language_label":              "语言",
		"shell_label":                 "Shell",
		"profile_label":               "配置",
		"none":                        "无",
		"status_success":              "成功",
		"status_degraded":             "降级",
		"status_failed":               "失败",
		"status_skipped":              "跳过",
		"status_detected":             "已检测到",
		"status_not_detected":         "未检测到",
		"level_low":                   "低",
		"level_medium":                "中",
		"level_high":                  "高",
		"language_current":            "当前语言：%s",
		"language_updated":            "语言已保存：%s",
		"language_choose_one":         "请只选择一个参数：--zh 或 --en",
		"language_unknown":            "不支持的语言 %q",
		"proxy_none_detected":         "未检测到代理",
		"env_proxy_detected":          "已检测到代理配置",
		"env_proxy_missing":           "未检测到代理",
		"dns_resolved":                "解析到 %d 个 IP 地址",
		"tcp_success":                 "TCP 连接成功",
		"tls_success":                 "TLS 握手成功（%s）",
		"http_success":                "HTTP 请求成功，状态码 %d",
		"http_degraded":               "HTTP 已到达目标，但返回状态码 %d",
		"dns_timeout":                 "DNS 解析超时",
		"dns_failed":                  "DNS 解析失败",
		"network_timeout":             "网络操作超时",
		"http_timeout":                "HTTP 请求超时",
		"http_failed":                 "HTTP 请求失败",
		"doctor_summary_success":      "所有探测都成功",
		"doctor_summary_degraded":     "%d 个探测结果为降级",
		"doctor_cause_dns":            "DNS 解析失败，后续传输层检查无法正常完成",
		"doctor_cause_tcp":            "TCP 连通失败，通常意味着路由、防火墙或代理存在问题",
		"doctor_cause_tls":            "TCP 已连接成功，但 TLS 握手失败",
		"doctor_cause_http":           "底层检查通过，但 HTTP 请求失败，可能与代理或上游策略有关",
		"doctor_cause_generic":        "一个或多个诊断探测失败",
		"doctor_action_proxy":         "如果你平时依赖代理，请使用 --proxy 重试，或先设置 HTTPS_PROXY 再运行 Claude Code。",
		"doctor_action_dns":           "检查当前 DNS 解析器，并与可信公共 DNS 对比，排除污染或异常解析。",
		"doctor_action_tcp":           "确认当前网络或代理是否能建立到目标主机 443 端口的 TCP 连接。",
		"doctor_action_tls":           "检查是否存在 TLS 劫持、企业证书或中间设备干扰握手。",
		"doctor_action_http":          "确认所选代理能访问目标地址，并检查上游服务是否拒绝当前出口 IP。",
		"doctor_action_none":          "当前没有需要立即处理的问题。",
		"dns_summary_failed":          "一个或多个目标的 DNS 解析失败",
		"dns_summary_degraded":        "DNS 虽然可解析，但结果存在可疑情况",
		"dns_summary_success":         "DNS 解析结果看起来正常",
		"dns_suggestion_failed":       "检查当前 DNS 解析器，或与可信公共 DNS 服务对比结果",
		"dns_suggestion_degraded":     "先对比不同解析器返回的 IP，再决定是否修改网络设置",
		"dns_record_failed":           "查询失败",
		"dns_record_degraded":         "解析到了私有地址或回环地址，结果可疑",
		"dns_record_success":          "解析成功",
		"dns_host_required":           "必须提供主机名",
		"risk_low_dns":                "当前问题更像 DNS 或本地网络故障",
		"risk_low_dns_suggestion":     "先修复解析问题，再判断是否存在出口 IP 或风控限制",
		"risk_high_http":              "链路已到达目标，但最终 HTTP 请求失败",
		"risk_high_http_suggestion":   "建议更换代理或出口 IP 后再对比结果",
		"risk_medium_tls":             "TCP 可达但 TLS 失败，可能存在过滤或中间拦截",
		"risk_medium_tls_suggestion":  "检查代理出口质量，以及当前网络中的 TLS 拦截情况",
		"risk_medium_tcp":             "HTTP 之前就失败了，更像路由或防火墙问题",
		"risk_medium_tcp_suggestion":  "先确认 443 出站连通性，再判断是否属于 IP 风控问题",
		"risk_low_none":               "暂未发现明显的 IP 或风控信号",
		"env_no_proxy_render":         "当前没有可用的代理，无法生成 shell 环境变量片段",
		"env_saved_proxy":             "已保存当前代理：%s",
		"env_cleared_proxy":           "已清除已保存的代理配置",
		"env_clear_missing":           "当前没有可清除的已保存代理配置",
		"env_set_requires_proxy":      "请提供 --set proxy-url，或提供结构化字段 --type、--server、--port",
		"env_set_invalid_port":        "port 必须大于 0",
		"proxy_port_label":            "代理端口",
		"expected_ip_label":           "期望 IP",
		"exit_ip_label":               "出口 IP",
		"proxy_port_failed":           "代理端口不可达：%s",
		"proxy_exit_ip_mismatch":      "出口 IP 未匹配期望值 %s",
		"proxy_test_success":          "代理连通性和出口 IP 检查通过",
		"proxy_expected_ip_not_set":   "settings.json 中未配置 expected_ip，请填写后再运行以启用出口 IP 校验",
		"doctor_checking":             "正在检测 %s (%s)...",
	},
}

func Normalize(language string) (string, error) {
	switch language {
	case "", English:
		return English, nil
	case Chinese:
		return Chinese, nil
	default:
		return "", errors.New(Text(English, "language_unknown", language))
	}
}

func MustNormalize(language string) string {
	normalized, err := Normalize(language)
	if err != nil {
		return English
	}
	return normalized
}

func Text(language, key string, args ...any) string {
	language = MustNormalize(language)
	template, ok := catalog[language][key]
	if !ok {
		template = catalog[English][key]
	}
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func Status(language, status string) string {
	return Text(language, "status_"+status)
}

func Level(language, level string) string {
	return Text(language, "level_"+level)
}
