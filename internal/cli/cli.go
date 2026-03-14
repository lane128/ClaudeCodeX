package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"claudecodex/internal/config"
	"claudecodex/internal/doctor"
	"claudecodex/internal/i18n"
	"claudecodex/internal/proxy"
)

func Run(args []string) int {
	cfg, path, err := config.Ensure("")
	if err != nil {
		cfg = config.Default()
		cfg.Normalize()
		path = ""
	}
	language := cfg.ResolvedLanguage()

	if len(args) == 0 {
		printUsage(os.Stderr, language)
		return 2
	}

	switch args[0] {
	case "doctor":
		return runDoctor(args[1:], cfg, language)
	case "env":
		return runEnv(args[1:], cfg, language)
	case "test":
		return runTest(args[1:], cfg, language)
	case "language":
		return runLanguage(args[1:], cfg, language)
	case "setting":
		return runSetting(args[1:], cfg, path, language)
	case "-h", "--help", "help":
		printUsage(os.Stdout, language)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printUsage(os.Stderr, language)
		return 2
	}
}

// runDoctor wires CLI flags into the doctor package and is intentionally thin.
// The command layer only resolves input sources; probe execution stays in internal/doctor.
func runDoctor(args []string, cfg config.Config, language string) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		timeoutSeconds = fs.Int("timeout", 5, "probe timeout in seconds")
		proxyOverride  = fs.String("proxy", "", "optional proxy URL override")
		jsonOutput     = fs.Bool("json", false, "print machine-readable JSON output")
		verbose        = fs.Bool("verbose", false, "print detailed probe output")
	)

	if err := fs.Parse(args); err != nil {
		return 2
	}

	activeProfileName := ""
	activeProfileURL := ""
	// doctor uses a clear precedence order:
	// CLI flag > process env > saved active profile.
	if activeProfile, ok := cfg.Active(); ok {
		activeProfileName = activeProfile.Name
		activeProfileURL = activeProfile.ProxyURL
	} else if cfg.LocalVPN.Enabled {
		activeProfileName = "local_vpn"
		activeProfileURL = cfg.LocalVPN.DefaultURL()
	}

	var onProbeStart func(name, target string)
	if !*jsonOutput && isInteractiveTerminal(os.Stdin, os.Stderr) {
		onProbeStart = func(name, target string) {
			fmt.Fprintf(os.Stderr, "\r%-50s", i18n.Text(language, "doctor_checking", name, target))
		}
	}

	result := doctor.Run(doctor.Options{
		Timeout:           time.Duration(*timeoutSeconds) * time.Second,
		ProxyOverride:     strings.TrimSpace(*proxyOverride),
		SavedProfileName:  activeProfileName,
		SavedProfileProxy: activeProfileURL,
		Language:          language,
		Verbose:           *verbose,
		Targets:           cfg.Doctor.Targets,
		OnProbeStart:      onProbeStart,
	})

	if onProbeStart != nil {
		fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 50))
	}

	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode JSON output: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(os.Stdout, doctor.RenderText(result, *verbose, language))
	}

	return result.ExitCode()
}

func runEnv(args []string, cfg config.Config, language string) int {
	fs := flag.NewFlagSet("env", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		jsonOutput = fs.Bool("json", false, "print machine-readable JSON output")
		shell      = fs.String("shell", "", "render shell exports for zsh, bash, sh, or fish")
		unset      = fs.Bool("unset", false, "render shell commands that clear proxy env vars")
	)

	if err := fs.Parse(args); err != nil {
		return 2
	}

	activeProfile := ""
	activeProxy := ""
	if profile, ok := cfg.Active(); ok {
		activeProfile = profile.Name
		activeProxy = profile.ProxyURL
	}
	if activeProxy == "" && cfg.LocalVPN.Enabled {
		activeProfile = "local_vpn"
		activeProxy = cfg.LocalVPN.DefaultURL()
	}

	effectiveSource, effectiveProxy := resolveProxySource(activeProfile, activeProxy, cfg)
	status := "not_detected"
	if effectiveProxy != "" {
		status = "detected"
	}

	info := map[string]string{
		"status":          status,
		"effective_proxy": effectiveProxy,
		"source":          effectiveSource,
		"HTTP_PROXY":      firstEnvValue("http_proxy", "HTTP_PROXY"),
		"HTTPS_PROXY":     firstEnvValue("https_proxy", "HTTPS_PROXY"),
		"ALL_PROXY":       firstEnvValue("all_proxy", "ALL_PROXY"),
		"active_profile": activeProfile,
		"active_proxy":   activeProxy,
	}

	if strings.TrimSpace(*shell) != "" {
		var (
			output string
			err    error
		)
		if *unset {
			output, err = proxy.RenderUnsetEnv(*shell)
		} else {
			if effectiveProxy == "" {
				fmt.Fprintln(os.Stderr, i18n.Text(language, "env_no_proxy_render"))
				return 1
			}
			output, err = proxy.RenderEnv(*shell, effectiveProxy)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to render shell env: %v\n", err)
			return 1
		}
		fmt.Fprint(os.Stdout, output)
		return 0
	}

	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode JSON output: %v\n", err)
			return 1
		}
		return 0
	}

	fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "status_label"), i18n.Status(language, status))
	for _, key := range []string{"HTTPS_PROXY", "HTTP_PROXY", "ALL_PROXY"} {
		value := info[key]
		if value == "" {
			value = "-"
		}
		fmt.Fprintf(os.Stdout, "%s: %s\n", strings.ToLower(key), value)
	}
	if effectiveProxy == "" {
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "effective_proxy_label"), i18n.Text(language, "none"))
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "source_label"), i18n.Text(language, "none"))
	} else {
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "effective_proxy_label"), effectiveProxy)
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "source_label"), effectiveSource)
	}
	if activeProfile != "" {
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "active_profile_label"), activeProfile)
	}
	if profile, ok := cfg.Active(); ok {
		if profile.Type != "" {
			fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "profile_label"), profile.Type)
		}
		if profile.Server != "" {
			fmt.Fprintf(os.Stdout, "Server: %s\n", profile.Server)
		}
		if profile.Port > 0 {
			fmt.Fprintf(os.Stdout, "Port: %d\n", profile.Port)
		}
		if profile.Username != "" {
			fmt.Fprintf(os.Stdout, "Username: %s\n", profile.Username)
		}
	}
	return 0
}

func runTest(args []string, cfg config.Config, language string) int {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		proxyURL       = fs.String("proxy", "", "proxy url override")
		targetURL      = fs.String("target", "", "target url to test through proxy")
		timeoutSeconds = fs.Int("timeout", 5, "probe timeout in seconds")
		jsonOutput     = fs.Bool("json", false, "print machine-readable JSON output")
		expectedIP     = fs.String("expect-ip", "", "expected exit IP")
		ipCheckURL     = fs.String("ip-check-url", "", "comma-separated IP check URLs")
		checkPort      = fs.Bool("check-port", true, "check whether the proxy host:port is reachable before HTTP tests")
	)

	if err := fs.Parse(args); err != nil {
		return 2
	}

	resolvedProxy, _ := resolveProxyURL(*proxyURL, cfg)

	target := *targetURL
	if strings.TrimSpace(target) == "" && len(cfg.Test.Targets) > 0 {
		target = cfg.Test.Targets[0]
	}
	expected := strings.TrimSpace(*expectedIP)
	if expected == "" {
		expected = strings.TrimSpace(cfg.Test.ExpectedIP)
	}
	ipURLs := splitCSV(*ipCheckURL)
	if len(ipURLs) == 0 {
		ipURLs = append([]string{}, cfg.Test.IPCheckURLs...)
	}

	result := proxy.Test(proxy.TestOptions{
		ProxyURL:   resolvedProxy,
		TargetURL:  target,
		Timeout:    time.Duration(*timeoutSeconds) * time.Second,
		Language:   language,
		ExpectedIP: expected,
		IPCheckURL: ipURLs,
		CheckPort:  *checkPort || cfg.Test.CheckProxyPort,
	})
	if *jsonOutput {
		data, err := result.JSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode JSON output: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stdout, "%s\n", data)
	} else {
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "status_label"), colorize(result.Status, i18n.Status(language, result.Status)))
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "proxy_label"), result.ProxyURL)
		if *checkPort && resolvedProxy != "" {
			portStatus := "failed"
			if result.ProxyReachable {
				portStatus = "success"
			}
			fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "proxy_port_label"), colorize(portStatus, i18n.Status(language, portStatus)))
		}
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "target_label"), result.TargetURL)
		if result.ExpectedIP != "" {
			fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "expected_ip_label"), result.ExpectedIP)
		}
		for source, ip := range result.ExitIPs {
			if strings.TrimSpace(ip) == "" {
				ip = "-"
			}
			fmt.Fprintf(os.Stdout, "%s (%s): %s\n", i18n.Text(language, "exit_ip_label"), source, ip)
		}
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "conclusion_label"), colorize(result.Status, result.Message))
		if result.ResponseCode != 0 {
			fmt.Fprintf(os.Stdout, "%s: %d\n", i18n.Text(language, "http_label"), result.ResponseCode)
		}
	}

	return result.ExitCode()
}

func runLanguage(args []string, cfg config.Config, language string) int {
	fs := flag.NewFlagSet("language", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	zh := fs.Bool("zh", false, "use Chinese output")
	en := fs.Bool("en", false, "use English output")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if !*zh && !*en {
		if isInteractiveTerminal(os.Stdin, os.Stdout) {
			selected, err := runLanguageSelector(os.Stdin, os.Stdout, cfg.ResolvedLanguage())
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to choose language: %v\n", err)
				return 1
			}
			return saveLanguageSelection(cfg, selected)
		}
		fmt.Fprintf(os.Stdout, "%s: %s\n", i18n.Text(language, "language_label"), cfg.ResolvedLanguage())
		return 0
	}
	if *zh == *en {
		fmt.Fprintln(os.Stderr, i18n.Text(language, "language_choose_one"))
		return 2
	}

	selected := "en"
	if *zh {
		selected = "zh"
	}
	if _, err := i18n.Normalize(selected); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}

	return saveLanguageSelection(cfg, selected)
}

func saveLanguageSelection(cfg config.Config, selected string) int {
	cfg.SetLanguage(selected)
	path, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve config path: %v\n", err)
		return 1
	}
	if err := config.Save(path, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stdout, "%s\n", i18n.Text(selected, "language_updated", selected))
	return 0
}

func runSetting(args []string, cfg config.Config, path, language string) int {
	fs := flag.NewFlagSet("setting", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	jsonOutput := fs.Bool("json", false, "print full settings as JSON")
	validate := fs.Bool("validate", false, "validate settings")
	showPath := fs.Bool("path", false, "print settings file path")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *validate {
		if err := cfg.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "settings invalid: %v\n", err)
			return 1
		}
		fmt.Fprintln(os.Stdout, "settings ok")
		return 0
	}

	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode JSON output: %v\n", err)
			return 1
		}
		return 0
	}

	if *showPath || true {
		fmt.Fprintf(os.Stdout, "Settings: %s\n", path)
	}
	return 0
}

// resolveProxyURL centralizes proxy lookup rules for test-related commands.
// A direct --proxy override wins first, then current environment, then local settings.
func resolveProxyURL(override string, cfg config.Config) (string, error) {
	if strings.TrimSpace(override) != "" {
		return proxy.NormalizeURL(override)
	}

	for _, key := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy", "ALL_PROXY", "all_proxy"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return proxy.NormalizeURL(value)
		}
	}

	profile, ok := cfg.Active()
	if !ok {
		if cfg.LocalVPN.Enabled {
			if value := strings.TrimSpace(cfg.LocalVPN.DefaultURL()); value != "" {
				return proxy.NormalizeURL(value)
			}
		}
		return "", fmt.Errorf("no proxy found in flags, environment, local_vpn, or active profile")
	}
	return profile.ProxyURL, nil
}

func resolveProxySource(activeProfile, activeProxy string, cfg config.Config) (string, string) {
	for _, key := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy", "ALL_PROXY", "all_proxy"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return "env:" + key, value
		}
	}
	if cfg.LocalVPN.Enabled {
		if value := strings.TrimSpace(cfg.LocalVPN.DefaultURL()); value != "" {
			return "settings:local_vpn", value
		}
	}
	if strings.TrimSpace(activeProxy) != "" {
		if strings.TrimSpace(activeProfile) == "" {
			return "profile:active", activeProxy
		}
		return "profile:" + activeProfile, activeProxy
	}
	return "none", ""
}

func splitHosts(value string) []string {
	return splitCSV(value)
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			values = append(values, item)
		}
	}
	return values
}

func firstEnvValue(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func printUsage(out *os.File, language string) {
	fmt.Fprintln(out, "ClaudeCodeX (ccx)")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, i18n.Text(language, "usage_header"))
	fmt.Fprintln(out, "  ccx doctor [--timeout seconds] [--proxy url] [--json] [--verbose]")
	fmt.Fprintln(out, "  ccx env [--json] [--shell zsh|bash|sh|fish] [--unset]")
	fmt.Fprintln(out, "  ccx test [--proxy url] [--target https://www.google.com/] [--expect-ip ip] [--ip-check-url url1,url2] [--check-port] [--timeout seconds] [--json]")
	fmt.Fprintln(out, "  ccx language [--zh|--en]")
	fmt.Fprintln(out, "  ccx setting [--path] [--json] [--validate]")
}

func isInteractiveTerminal(input *os.File, output *os.File) bool {
	inputInfo, err := input.Stat()
	if err != nil {
		return false
	}
	outputInfo, err := output.Stat()
	if err != nil {
		return false
	}
	return (inputInfo.Mode()&os.ModeCharDevice) != 0 && (outputInfo.Mode()&os.ModeCharDevice) != 0
}
