package proxy

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"claudecodex/internal/i18n"
)

const DefaultShell = "zsh"

var DefaultIPCheckURLs = []string{
	"http://icanhazip.com",
	"https://ping0.cc",
}

type TestOptions struct {
	ProxyURL   string
	TargetURL  string
	Timeout    time.Duration
	Language   string
	CheckPort  bool
	ExpectedIP string
	IPCheckURL []string
}

type TestResult struct {
	ProxyURL       string            `json:"proxy_url"`
	TargetURL      string            `json:"target_url"`
	Status         string            `json:"status"`
	DurationMS     int64             `json:"duration_ms"`
	Message        string            `json:"message"`
	ResponseCode   int               `json:"response_code,omitempty"`
	ProxyReachable bool              `json:"proxy_reachable"`
	ExpectedIP     string            `json:"expected_ip,omitempty"`
	ExitIPMatched  bool              `json:"exit_ip_matched"`
	ExitIPs        map[string]string `json:"exit_ips,omitempty"`
	Details        map[string]string `json:"details,omitempty"`
}

// NormalizeURL validates user input early so invalid proxy settings never reach
// profile storage or transport construction.
func NormalizeURL(proxyURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(proxyURL))
	if err != nil {
		return "", fmt.Errorf("parse proxy url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("proxy url must include scheme and host")
	}
	switch parsed.Scheme {
	case "http", "https", "socks5":
	default:
		return "", fmt.Errorf("unsupported proxy scheme %q", parsed.Scheme)
	}
	return parsed.String(), nil
}

// RenderEnv prints shell code instead of mutating the parent process environment.
// That makes the command predictable and safe to inspect before evaluation.
func RenderEnv(shell, proxyURL string) (string, error) {
	normalized, err := NormalizeURL(proxyURL)
	if err != nil {
		return "", err
	}

	switch shellName(shell) {
	case "zsh", "bash", "sh":
		return fmt.Sprintf("export HTTPS_PROXY=%q\nexport HTTP_PROXY=%q\nexport ALL_PROXY=%q\n", normalized, normalized, normalized), nil
	case "fish":
		return fmt.Sprintf("set -x HTTPS_PROXY %q\nset -x HTTP_PROXY %q\nset -x ALL_PROXY %q\n", normalized, normalized, normalized), nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shell)
	}
}

func RenderUnsetEnv(shell string) (string, error) {
	switch shellName(shell) {
	case "zsh", "bash", "sh":
		return "unset HTTPS_PROXY\nunset HTTP_PROXY\nunset ALL_PROXY\n", nil
	case "fish":
		return "set -e HTTPS_PROXY\nset -e HTTP_PROXY\nset -e ALL_PROXY\n", nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shell)
	}
}

// Test performs connectivity, proxy-port, and exit-IP checks through the selected proxy.
// If ProxyURL is empty, checks are performed via direct connection.
func Test(opts TestOptions) TestResult {
	startedAt := time.Now()
	result := TestResult{
		ProxyURL:   strings.TrimSpace(opts.ProxyURL),
		TargetURL:  strings.TrimSpace(opts.TargetURL),
		Status:     "failed",
		ExpectedIP: strings.TrimSpace(opts.ExpectedIP),
		ExitIPs:    map[string]string{},
	}

	if result.TargetURL == "" {
		result.TargetURL = "https://www.google.com/"
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}

	var client *http.Client

	if strings.TrimSpace(opts.ProxyURL) == "" {
		result.ProxyURL = "direct"
		result.ProxyReachable = true
		client = directClient(opts.Timeout)
	} else {
		normalized, err := NormalizeURL(opts.ProxyURL)
		if err != nil {
			result.DurationMS = time.Since(startedAt).Milliseconds()
			result.Message = err.Error()
			return result
		}
		result.ProxyURL = normalized

		parsed, err := url.Parse(normalized)
		if err != nil {
			result.DurationMS = time.Since(startedAt).Milliseconds()
			result.Message = err.Error()
			return result
		}

		if opts.CheckPort {
			result.ProxyReachable = checkPort(parsed.Host, opts.Timeout)
			if !result.ProxyReachable {
				result.DurationMS = time.Since(startedAt).Milliseconds()
				result.Message = i18n.Text(opts.Language, "proxy_port_failed", parsed.Host)
				result.Details = map[string]string{"proxy_host": parsed.Host}
				return result
			}
		} else {
			result.ProxyReachable = true
		}

		client, err = proxyClient(parsed, opts.Timeout)
		if err != nil {
			result.DurationMS = time.Since(startedAt).Milliseconds()
			result.Message = err.Error()
			return result
		}
	}

	if len(opts.IPCheckURL) == 0 {
		opts.IPCheckURL = DefaultIPCheckURLs
	}
	checkExitIPs(client, opts, &result)

	resp, err := requestThroughProxy(client, result.TargetURL, opts.Timeout)
	if err != nil {
		result.DurationMS = time.Since(startedAt).Milliseconds()
		result.Message = err.Error()
		result.Details = map[string]string{"proxy_url": result.ProxyURL}
		return result
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 512))

	result.DurationMS = time.Since(startedAt).Milliseconds()
	result.ResponseCode = resp.StatusCode
	result.Details = map[string]string{"proxy_url": result.ProxyURL}

	if resp.StatusCode >= 400 {
		result.Status = "failed"
		result.Message = i18n.Text(opts.Language, "http_degraded", resp.StatusCode)
		return result
	}

	if result.ExpectedIP == "" {
		result.Status = "failed"
		result.Message = i18n.Text(opts.Language, "proxy_expected_ip_not_set")
		return result
	}

	if !result.ExitIPMatched {
		result.Status = "failed"
		result.Message = i18n.Text(opts.Language, "proxy_exit_ip_mismatch", result.ExpectedIP)
		return result
	}

	result.Status = "success"
	result.Message = i18n.Text(opts.Language, "proxy_test_success")
	return result
}

func (r TestResult) ExitCode() int {
	if r.Status == "success" {
		return 0
	}
	return 2
}

func (r TestResult) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func checkPort(host string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func directClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout + 2*time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
			TLSHandshakeTimeout: timeout,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
}

func proxyClient(proxyURL *url.URL, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
		TLSHandshakeTimeout: timeout,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return &http.Client{
		Timeout:   timeout + 2*time.Second,
		Transport: transport,
	}, nil
}

func requestThroughProxy(client *http.Client, targetURL string, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout+2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ccx/0.1")
	return client.Do(req)
}

func checkExitIPs(client *http.Client, opts TestOptions, result *TestResult) {
	for _, rawURL := range opts.IPCheckURL {
		target := strings.TrimSpace(rawURL)
		if target == "" {
			continue
		}
		ip, err := fetchPlainText(client, target, opts.Timeout)
		if err != nil {
			result.ExitIPs[target] = ""
			continue
		}
		result.ExitIPs[target] = ip
		if result.ExpectedIP != "" && ip == result.ExpectedIP {
			result.ExitIPMatched = true
		}
	}
}

func fetchPlainText(client *http.Client, targetURL string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout+2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "curl/7.88.1")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func shellName(shell string) string {
	shell = strings.TrimSpace(shell)
	if shell == "" {
		return DefaultShell
	}
	return shell
}
