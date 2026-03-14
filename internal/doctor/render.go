package doctor

import (
	"fmt"
	"strings"

	"claudecodex/internal/i18n"
)

func RenderText(result Result, verbose bool, language string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "ClaudeCodeX doctor\n")
	fmt.Fprintf(&b, "%s: %s\n", i18n.Text(language, "status_label"), i18n.Status(language, result.OverallStatus))
	env := environmentDetails(result.ProbeResults)
	fmt.Fprintf(&b, "https_proxy: %s\n", env["https_proxy"])
	fmt.Fprintf(&b, "http_proxy: %s\n", env["http_proxy"])
	fmt.Fprintf(&b, "all_proxy: %s\n", env["all_proxy"])
	if result.Proxy.Used {
		fmt.Fprintf(&b, "%s: %s (%s)\n", i18n.Text(language, "proxy_label"), result.Proxy.URL, result.Proxy.Source)
	} else {
		fmt.Fprintf(&b, "%s: %s\n", i18n.Text(language, "proxy_label"), i18n.Text(language, "proxy_none_detected"))
	}
	for _, layer := range []string{"dns", "tcp", "tls", "http"} {
		fmt.Fprintf(&b, "%s: %s\n", i18n.Text(language, layer+"_label"), i18n.Status(language, layerStatus(result.ProbeResults, layer)))
	}
	fmt.Fprintf(&b, "%s: %s\n", i18n.Text(language, "conclusion_label"), result.Summary)

	if len(result.NextActions) > 0 {
		fmt.Fprintf(&b, "%s: %s\n", i18n.Text(language, "suggestion_label"), result.NextActions[0])
	}

	if verbose {
		fmt.Fprintf(&b, "\nDuration: %dms\n", result.DurationMS)
		fmt.Fprintf(&b, "\nProbes:\n")
		for _, probe := range result.ProbeResults {
			fmt.Fprintf(&b, "- [%s] %s %s: %s\n", strings.ToUpper(probe.Status), probe.Layer, probe.Target, probe.Message)
			if len(probe.Details) > 0 {
				for key, value := range probe.Details {
					fmt.Fprintf(&b, "  %s: %s\n", key, value)
				}
			}
		}

		if len(result.NextActions) > 1 {
			fmt.Fprintf(&b, "\nMore suggestions:\n")
			for _, action := range result.NextActions[1:] {
				fmt.Fprintf(&b, "- %s\n", action)
			}
		}
	}

	return b.String()
}

func layerStatus(probes []Probe, layer string) string {
	status := "skipped"
	for _, probe := range probes {
		if probe.Layer != layer {
			continue
		}
		switch probe.Status {
		case "failed":
			return "failed"
		case "degraded":
			status = "degraded"
		case "success":
			if status == "skipped" {
				status = "success"
			}
		}
	}
	return status
}

func environmentDetails(probes []Probe) map[string]string {
	result := map[string]string{
		"https_proxy": "-",
		"http_proxy":  "-",
		"all_proxy":   "-",
	}
	for _, probe := range probes {
		if probe.Layer != "env" || len(probe.Details) == 0 {
			continue
		}
		for _, key := range []string{"https_proxy", "http_proxy", "all_proxy"} {
			if value := probe.Details[key]; value != "" {
				result[key] = value
			}
		}
		break
	}
	return result
}
