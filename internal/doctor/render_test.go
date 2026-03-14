package doctor

import (
	"strings"
	"testing"
)

func TestRenderTextDefaultIsCompact(t *testing.T) {
	result := Result{
		OverallStatus: "failed",
		Summary:       "dns resolution failed before transport checks could complete",
		Proxy: ProxyConfig{
			Used:   true,
			URL:    "http://127.0.0.1:7890",
			Source: "env:HTTPS_PROXY",
		},
		ProbeResults: []Probe{
			{
				Layer: "env",
				Status: "success",
				Details: map[string]string{
					"https_proxy": "http://127.0.0.1:7890",
					"http_proxy":  "http://127.0.0.1:7890",
					"all_proxy":   "socks5://127.0.0.1:7890",
				},
			},
			{Layer: "dns", Status: "failed"},
			{Layer: "tcp", Status: "failed"},
			{Layer: "tls", Status: "failed"},
			{Layer: "http", Status: "failed"},
		},
		NextActions: []string{"check your dns resolver"},
	}

	output := RenderText(result, false, "en")
	for _, want := range []string{
		"Status: FAILED",
		"https_proxy: http://127.0.0.1:7890",
		"http_proxy: http://127.0.0.1:7890",
		"all_proxy: socks5://127.0.0.1:7890",
		"Proxy: http://127.0.0.1:7890 (env:HTTPS_PROXY)",
		"DNS: FAILED",
		"TCP: FAILED",
		"TLS: FAILED",
		"HTTP: FAILED",
		"Conclusion: dns resolution failed before transport checks could complete",
		"Suggestion: check your dns resolver",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}

	if strings.Contains(output, "Probes:") {
		t.Fatalf("expected compact output without verbose probe list, got:\n%s", output)
	}
}
