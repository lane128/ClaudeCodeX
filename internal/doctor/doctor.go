package doctor

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"claudecodex/internal/i18n"
)

type Options struct {
	Timeout           time.Duration
	ProxyOverride     string
	SavedProfileName  string
	SavedProfileProxy string
	Language          string
	Verbose           bool
	Targets           []string             // probe targets as URLs; falls back to defaultTargets when empty
	OnProbeStart      func(name, target string) // called before each network probe; may be nil
}

type Result struct {
	Command          string      `json:"command"`
	StartedAt        time.Time   `json:"started_at"`
	DurationMS       int64       `json:"duration_ms"`
	OverallStatus    string      `json:"overall_status"`
	Summary          string      `json:"summary"`
	NextActions      []string    `json:"next_actions"`
	Proxy            ProxyConfig `json:"proxy"`
	ProbeResults     []Probe     `json:"probes"`
	SuccessfulProbes int         `json:"successful_probes"`
	FailedProbes     int         `json:"failed_probes"`
}

type ProxyConfig struct {
	Source string `json:"source"`
	URL    string `json:"url,omitempty"`
	Used   bool   `json:"used"`
}

type Probe struct {
	Name       string            `json:"name"`
	Layer      string            `json:"layer"`
	Target     string            `json:"target"`
	Status     string            `json:"status"`
	DurationMS int64             `json:"duration_ms"`
	Message    string            `json:"message"`
	UsedProxy  bool              `json:"used_proxy"`
	Details    map[string]string `json:"details,omitempty"`
}

type target struct {
	Name string
	Host string
	URL  string
}

var defaultTargets = []target{
	{Name: "Anthropic", Host: "anthropic.com:443", URL: "https://anthropic.com/"},
	{Name: "Claude AI", Host: "claude.ai:443", URL: "https://claude.ai/"},
	{Name: "Claude", Host: "claude.com:443", URL: "https://claude.com/"},
}

// Run executes the full diagnosis as a fixed sequence of layer-by-layer probes.
// Keeping the probe order explicit makes it easier to explain where a failure starts.
func Run(opts Options) Result {
	startedAt := time.Now()
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	result := Result{
		Command:   "ccx doctor",
		StartedAt: startedAt,
		Proxy:     detectProxy(opts.ProxyOverride, opts.SavedProfileName, opts.SavedProfileProxy),
	}

	notify := func(name, target string) {
		if opts.OnProbeStart != nil {
			opts.OnProbeStart(name, target)
		}
	}

	targets := defaultTargets
	if len(opts.Targets) > 0 {
		targets = urlsToTargets(opts.Targets)
	}

	envProbe := probeEnvironment(result.Proxy, opts.Language)
	result.ProbeResults = append(result.ProbeResults, envProbe)

	for _, t := range targets {
		notify("DNS", t.Host)
		result.ProbeResults = append(result.ProbeResults, probeDNS(t, timeout, opts.Language))
		notify("TCP", t.Host)
		result.ProbeResults = append(result.ProbeResults, probeTCP(t, timeout, opts.Language))
		notify("TLS", t.Host)
		result.ProbeResults = append(result.ProbeResults, probeTLS(t, timeout, opts.Language))
		notify("HTTP", t.URL)
		result.ProbeResults = append(result.ProbeResults, probeHTTP(t, timeout, result.Proxy, opts.Language))
	}

	result.DurationMS = time.Since(startedAt).Milliseconds()
	summarize(&result, opts.Language)
	return result
}

func (r Result) ExitCode() int {
	switch r.OverallStatus {
	case "success":
		return 0
	case "degraded":
		return 1
	default:
		return 2
	}
}

// detectProxy applies the same precedence model exposed by the CLI:
// explicit flag first, then environment variables, then the saved active profile.
func detectProxy(override, savedProfileName, savedProfileProxy string) ProxyConfig {
	if override != "" {
		return ProxyConfig{
			Source: "flag",
			URL:    override,
			Used:   true,
		}
	}

	for _, key := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy", "ALL_PROXY", "all_proxy"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return ProxyConfig{
				Source: "env:" + key,
				URL:    value,
				Used:   true,
			}
		}
	}

	if strings.TrimSpace(savedProfileProxy) != "" {
		name := strings.TrimSpace(savedProfileName)
		if name == "" {
			name = "active"
		}
		return ProxyConfig{
			Source: "profile:" + name,
			URL:    strings.TrimSpace(savedProfileProxy),
			Used:   true,
		}
	}

	return ProxyConfig{Source: "none"}
}

// probeEnvironment records how proxy settings entered the process.
// This gives context for later network failures without touching the network yet.
func probeEnvironment(proxy ProxyConfig, language string) Probe {
	message := i18n.Text(language, "env_proxy_missing")
	status := "success"
	details := map[string]string{
		"goos":        os.Getenv("GOOS"),
		"shell":       os.Getenv("SHELL"),
		"source":      proxy.Source,
		"https_proxy": envProxyValue("https_proxy", "HTTPS_PROXY"),
		"http_proxy":  envProxyValue("http_proxy", "HTTP_PROXY"),
		"all_proxy":   envProxyValue("all_proxy", "ALL_PROXY"),
	}
	if details["goos"] == "" {
		details["goos"] = runtimeGOOS()
	}

	if proxy.Used {
		message = i18n.Text(language, "env_proxy_detected")
		details["proxy_url"] = proxy.URL
	}

	return Probe{
		Name:      "Environment",
		Layer:     "env",
		Target:    "local",
		Status:    status,
		Message:   message,
		UsedProxy: proxy.Used,
		Details:   details,
	}
}

func envProxyValue(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return "-"
}

// probeDNS is deliberately separate from TCP/TLS so the tool can distinguish
// name-resolution failures from transport failures in user-facing output.
func probeDNS(t target, timeout time.Duration, language string) Probe {
	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	host, _, err := net.SplitHostPort(t.Host)
	if err != nil {
		return failedProbe("DNS", "dns", t.Host, false, startedAt, fmt.Errorf("invalid target host: %w", err), language)
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return failedProbe("DNS", "dns", t.Host, false, startedAt, err, language)
	}

	values := make([]string, 0, len(ips))
	for _, ip := range ips {
		values = append(values, ip.IP.String())
	}
	sort.Strings(values)

	return Probe{
		Name:       "DNS",
		Layer:      "dns",
		Target:     t.Host,
		Status:     "success",
		DurationMS: time.Since(startedAt).Milliseconds(),
		Message:    i18n.Text(language, "dns_resolved", len(values)),
		Details: map[string]string{
			"hostname": host,
			"ips":      strings.Join(values, ", "),
		},
	}
}

func probeTCP(t target, timeout time.Duration, language string) Probe {
	startedAt := time.Now()
	conn, err := net.DialTimeout("tcp", t.Host, timeout)
	if err != nil {
		return failedProbe("TCP", "tcp", t.Host, false, startedAt, err, language)
	}
	defer conn.Close()

	return Probe{
		Name:       "TCP",
		Layer:      "tcp",
		Target:     t.Host,
		Status:     "success",
		DurationMS: time.Since(startedAt).Milliseconds(),
		Message:    i18n.Text(language, "tcp_success"),
	}
}

func probeTLS(t target, timeout time.Duration, language string) Probe {
	startedAt := time.Now()
	host, _, err := net.SplitHostPort(t.Host)
	if err != nil {
		return failedProbe("TLS", "tls", t.Host, false, startedAt, fmt.Errorf("invalid target host: %w", err), language)
	}

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", t.Host, &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: host,
	})
	if err != nil {
		return failedProbe("TLS", "tls", t.Host, false, startedAt, err, language)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	version := tlsVersionName(state.Version)
	cipherSuite := tls.CipherSuiteName(state.CipherSuite)

	return Probe{
		Name:       "TLS",
		Layer:      "tls",
		Target:     t.Host,
		Status:     "success",
		DurationMS: time.Since(startedAt).Milliseconds(),
		Message:    i18n.Text(language, "tls_success", version),
		Details: map[string]string{
			"server_name":  host,
			"tls_version":  version,
			"cipher_suite": cipherSuite,
		},
	}
}

// probeHTTP reuses the selected proxy configuration because HTTP is the first layer
// where proxy behavior matters directly for most users.
func probeHTTP(t target, timeout time.Duration, proxy ProxyConfig, language string) Probe {
	startedAt := time.Now()
	client, usedProxy, err := httpClient(timeout, proxy)
	if err != nil {
		return failedProbe("HTTP", "http", t.URL, proxy.Used, startedAt, err, language)
	}

	req, err := http.NewRequest(http.MethodGet, t.URL, nil)
	if err != nil {
		return failedProbe("HTTP", "http", t.URL, usedProxy, startedAt, err, language)
	}
	req.Header.Set("User-Agent", "ccx/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return failedProbe("HTTP", "http", t.URL, usedProxy, startedAt, err, language)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 512))

	status := "success"
	message := i18n.Text(language, "http_success", resp.StatusCode)
	if resp.StatusCode >= 400 {
		status = "degraded"
		message = i18n.Text(language, "http_degraded", resp.StatusCode)
	}

	return Probe{
		Name:       "HTTP",
		Layer:      "http",
		Target:     t.URL,
		Status:     status,
		DurationMS: time.Since(startedAt).Milliseconds(),
		Message:    message,
		UsedProxy:  usedProxy,
		Details: map[string]string{
			"status_code": fmt.Sprintf("%d", resp.StatusCode),
		},
	}
}

func httpClient(timeout time.Duration, proxy ProxyConfig) (*http.Client, bool, error) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
		TLSHandshakeTimeout: timeout,
	}

	usedProxy := false
	if proxy.Used {
		parsed, err := url.Parse(proxy.URL)
		if err != nil {
			return nil, false, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsed)
		usedProxy = true
	}

	return &http.Client{
		Timeout:   timeout + 2*time.Second,
		Transport: transport,
	}, usedProxy, nil
}

// summarize collapses per-probe results into a single status and user guidance.
// The goal is fast triage rather than exposing every low-level error equally.
func summarize(result *Result, language string) {
	var failed []Probe
	var degraded []Probe

	for _, probe := range result.ProbeResults {
		switch probe.Status {
		case "success":
			result.SuccessfulProbes++
		case "degraded":
			degraded = append(degraded, probe)
		default:
			failed = append(failed, probe)
		}
	}

	result.FailedProbes = len(failed)

	switch {
	case len(failed) == 0 && len(degraded) == 0:
		result.OverallStatus = "success"
		result.Summary = i18n.Text(language, "doctor_summary_success")
	case len(failed) == 0:
		result.OverallStatus = "degraded"
		result.Summary = i18n.Text(language, "doctor_summary_degraded", len(degraded))
	default:
		result.OverallStatus = "failed"
		result.Summary = probableCause(failed[0], language)
	}

	result.NextActions = recommendActions(result.ProbeResults, result.Proxy, language)
}

func probableCause(probe Probe, language string) string {
	switch probe.Layer {
	case "dns":
		return i18n.Text(language, "doctor_cause_dns")
	case "tcp":
		return i18n.Text(language, "doctor_cause_tcp")
	case "tls":
		return i18n.Text(language, "doctor_cause_tls")
	case "http":
		return i18n.Text(language, "doctor_cause_http")
	default:
		return i18n.Text(language, "doctor_cause_generic")
	}
}

func recommendActions(probes []Probe, proxy ProxyConfig, language string) []string {
	actions := []string{}

	if !proxy.Used {
		actions = append(actions, i18n.Text(language, "doctor_action_proxy"))
	}

	for _, probe := range probes {
		if probe.Status == "success" {
			continue
		}
		switch probe.Layer {
		case "dns":
			actions = append(actions, i18n.Text(language, "doctor_action_dns"))
		case "tcp":
			actions = append(actions, i18n.Text(language, "doctor_action_tcp"))
		case "tls":
			actions = append(actions, i18n.Text(language, "doctor_action_tls"))
		case "http":
			actions = append(actions, i18n.Text(language, "doctor_action_http"))
		}
	}

	if len(actions) == 0 {
		actions = append(actions, i18n.Text(language, "doctor_action_none"))
	}

	return dedupe(actions)
}

func failedProbe(name, layer, target string, usedProxy bool, startedAt time.Time, err error, language string) Probe {
	return Probe{
		Name:       name,
		Layer:      layer,
		Target:     target,
		Status:     "failed",
		DurationMS: time.Since(startedAt).Milliseconds(),
		Message:    classifyError(err, language),
		UsedProxy:  usedProxy,
		Details: map[string]string{
			"error": err.Error(),
		},
	}
}

func classifyError(err error, language string) string {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsTimeout {
			return i18n.Text(language, "dns_timeout")
		}
		return i18n.Text(language, "dns_failed")
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Timeout() {
			return i18n.Text(language, "network_timeout")
		}
	}

	urlErr := new(url.Error)
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return i18n.Text(language, "http_timeout")
		}
		return i18n.Text(language, "http_failed")
	}

	return err.Error()
}

func dedupe(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

// urlsToTargets converts a list of URL strings into probe targets.
// The host and port are derived from the URL; HTTPS defaults to port 443.
func urlsToTargets(urls []string) []target {
	result := make([]target, 0, len(urls))
	for _, raw := range urls {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Hostname() == "" {
			continue
		}
		host := parsed.Hostname()
		port := parsed.Port()
		if port == "" {
			if parsed.Scheme == "https" {
				port = "443"
			} else {
				port = "80"
			}
		}
		result = append(result, target{
			Name: host,
			Host: host + ":" + port,
			URL:  raw,
		})
	}
	return result
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS12:
		return "TLS1.2"
	case tls.VersionTLS13:
		return "TLS1.3"
	default:
		return fmt.Sprintf("0x%x", version)
	}
}
