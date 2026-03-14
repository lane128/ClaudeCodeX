package proxy

import "testing"

func TestNormalizeURL(t *testing.T) {
	normalized, err := NormalizeURL("http://127.0.0.1:7890")
	if err != nil {
		t.Fatalf("normalize url: %v", err)
	}
	if normalized != "http://127.0.0.1:7890" {
		t.Fatalf("unexpected normalized url %q", normalized)
	}
}

func TestRenderEnv(t *testing.T) {
	output, err := RenderEnv("zsh", "http://127.0.0.1:7890")
	if err != nil {
		t.Fatalf("render env: %v", err)
	}
	if output == "" {
		t.Fatal("expected shell output")
	}
}

func TestRenderUnsetEnv(t *testing.T) {
	output, err := RenderUnsetEnv("zsh")
	if err != nil {
		t.Fatalf("render unset env: %v", err)
	}
	if output == "" {
		t.Fatal("expected shell unset output")
	}
}

func TestTestResultExitCode(t *testing.T) {
	result := TestResult{Status: "success"}
	if result.ExitCode() != 0 {
		t.Fatalf("expected success exit code 0, got %d", result.ExitCode())
	}
}

func TestNormalizeURLSupportsSocks5(t *testing.T) {
	normalized, err := NormalizeURL("socks5://user:pass@1.1.1.1:8888")
	if err != nil {
		t.Fatalf("normalize socks5 url: %v", err)
	}
	if normalized != "socks5://user:pass@1.1.1.1:8888" {
		t.Fatalf("unexpected normalized url %q", normalized)
	}
}
