package doctor

import "testing"

func TestSummarizeSuccess(t *testing.T) {
	result := Result{
		ProbeResults: []Probe{
			{Layer: "env", Status: "success"},
			{Layer: "dns", Status: "success"},
		},
	}

	summarize(&result, "en")

	if result.OverallStatus != "success" {
		t.Fatalf("expected success, got %q", result.OverallStatus)
	}
	if result.ExitCode() != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode())
	}
}

func TestSummarizeFailure(t *testing.T) {
	result := Result{
		ProbeResults: []Probe{
			{Layer: "env", Status: "success"},
			{Layer: "dns", Status: "failed"},
		},
	}

	summarize(&result, "en")

	if result.OverallStatus != "failed" {
		t.Fatalf("expected failed, got %q", result.OverallStatus)
	}
	if result.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %d", result.ExitCode())
	}
	if result.Summary == "" {
		t.Fatal("expected summary to be populated")
	}
}

func TestDetectProxyOverride(t *testing.T) {
	proxy := detectProxy("http://127.0.0.1:7890", "", "")
	if !proxy.Used {
		t.Fatal("expected override proxy to be used")
	}
	if proxy.Source != "flag" {
		t.Fatalf("expected source flag, got %q", proxy.Source)
	}
}

func TestDetectProxySavedProfile(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("https_proxy", "")
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("http_proxy", "")
	t.Setenv("ALL_PROXY", "")
	t.Setenv("all_proxy", "")

	proxy := detectProxy("", "office", "http://127.0.0.1:7890")
	if !proxy.Used {
		t.Fatal("expected saved profile proxy to be used")
	}
	if proxy.Source != "profile:office" {
		t.Fatalf("unexpected source %q", proxy.Source)
	}
}
