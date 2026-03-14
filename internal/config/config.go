package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tailscale/hujson"
)

const (
	defaultSettingsFile = "settings.json"
	legacyConfigFile    = "config.json"
	defaultDirName      = ".ccx"
)

type Config struct {
	ActiveProfile string                  `json:"active_profile,omitempty"`
	Language      string                  `json:"language,omitempty"`
	Profiles      map[string]ProxyProfile `json:"profiles,omitempty"`
	LocalVPN      LocalVPNSettings        `json:"local_vpn"`
	Test          TestSettings            `json:"test,omitempty"`
	Doctor        DoctorSettings          `json:"doctor,omitempty"`
	Rules         RuleSettings            `json:"rules,omitempty"`
}

type DoctorSettings struct {
	Targets []string `json:"targets,omitempty"`
}

type LocalVPNSettings struct {
	Enabled bool             `json:"enabled"`
	HTTP    LocalProxyConfig `json:"http"`
	SOCKS5  LocalProxyConfig `json:"socks5"`
}

type LocalProxyConfig struct {
	Server   string `json:"server,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type TestSettings struct {
	ExpectedIP     string   `json:"expected_ip,omitempty"`
	IPCheckURLs    []string `json:"ip_check_urls,omitempty"`
	Targets        []string `json:"targets,omitempty"`
	CheckProxyPort bool     `json:"check_proxy_port,omitempty"`
}

type RuleSettings struct {
	Domains []string `json:"domains,omitempty"`
}

type ProxyProfile struct {
	Name      string    `json:"name"`
	ProxyURL  string    `json:"proxy_url,omitempty"`
	Type      string    `json:"type,omitempty"`
	Server    string    `json:"server,omitempty"`
	Port      int       `json:"port,omitempty"`
	Username  string    `json:"username,omitempty"`
	Password  string    `json:"password,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultPath supports explicit overrides first so tests, CI, and local experiments
// can avoid writing into the real user config directory.
func DefaultPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv("CCX_CONFIG_PATH")); override != "" {
		return override, nil
	}
	if xdgHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdgHome != "" {
		return filepath.Join(xdgHome, defaultDirName, defaultSettingsFile), nil
	}
	root, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home dir: %w", err)
	}
	return filepath.Join(root, defaultDirName, defaultSettingsFile), nil
}

// Load treats a missing config file as an empty config.
// That keeps first-run UX simple and avoids forcing a bootstrap step.
func Load(path string) (Config, error) {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return Config{}, err
		}
		if _, statErr := os.Stat(path); errors.Is(statErr, os.ErrNotExist) {
			legacyPath := legacyPathFor(path)
			if _, legacyErr := os.Stat(legacyPath); legacyErr == nil {
				path = legacyPath
			}
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := Default()
			cfg.Normalize()
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	standardJSON, err := hujson.Standardize(data)
	if err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(standardJSON, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	cfg.Normalize()
	return cfg, nil
}

// Save always writes a complete JSON snapshot.
// The config is small enough that full rewrites are simpler than partial mutation.
func Save(path string, cfg Config) error {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return err
		}
	}

	cfg.Normalize()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func Ensure(path string) (Config, string, error) {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return Config{}, "", err
		}
	}

	raw, readErr := os.ReadFile(path)
	fileExists := readErr == nil

	needsSave := !fileExists
	if fileExists && rawNeedsLocalVPN(raw) {
		needsSave = true
	}

	cfg, err := Load(path)
	if err != nil {
		return Config{}, "", err
	}

	if needsSave {
		if err := Save(path, cfg); err != nil {
			return Config{}, "", err
		}
	}
	return cfg, path, nil
}

// rawNeedsLocalVPN reports whether the raw JSON bytes are missing the local_vpn
// block, which happens when the config was created before local_vpn was introduced.
func rawNeedsLocalVPN(raw []byte) bool {
	var probe struct {
		LocalVPN *json.RawMessage `json:"local_vpn"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	return probe.LocalVPN == nil
}

func Default() Config {
	cfg := Config{
		Language: "en",
		Profiles: map[string]ProxyProfile{},
		LocalVPN: LocalVPNSettings{
			Enabled: false,
			HTTP: LocalProxyConfig{
				Server: "127.0.0.1",
				Port:   7890,
			},
			SOCKS5: LocalProxyConfig{
				Server: "127.0.0.1",
				Port:   7891,
			},
		},
		Test: TestSettings{
			IPCheckURLs: []string{
				"http://icanhazip.com",
				"https://ping0.cc",
			},
			Targets: []string{
				"https://claude.ai/",
				"https://api.anthropic.com/",
			},
			CheckProxyPort: true,
		},
		Doctor: DoctorSettings{
			Targets: []string{
				"https://anthropic.com/",
				"https://claude.ai/",
				"https://claude.com/",
			},
		},
		Rules: RuleSettings{
			Domains: []string{
				"api.anthropic.com",
				"anthropic.com",
				"claude.ai",
				"claudeusercontent.com",
				"claude.com",
				"ping0.cc",
				"icanhazip.com",
				"proxy-cheap.com",
			},
		},
	}
	return cfg
}

func (c *Config) Normalize() {
	if c.Language == "" {
		c.Language = "en"
	}
	if c.Profiles == nil {
		c.Profiles = map[string]ProxyProfile{}
	}
	if len(c.Test.IPCheckURLs) == 0 {
		c.Test.IPCheckURLs = append([]string{}, Default().Test.IPCheckURLs...)
	}
	if len(c.Test.Targets) == 0 {
		c.Test.Targets = append([]string{}, Default().Test.Targets...)
	}
	if len(c.Doctor.Targets) == 0 {
		c.Doctor.Targets = append([]string{}, Default().Doctor.Targets...)
	}
	if !c.Test.CheckProxyPort {
		c.Test.CheckProxyPort = Default().Test.CheckProxyPort
	}
	if len(c.Rules.Domains) == 0 {
		c.Rules.Domains = append([]string{}, Default().Rules.Domains...)
	}
	if c.LocalVPN.HTTP.Server == "" {
		c.LocalVPN.HTTP.Server = Default().LocalVPN.HTTP.Server
	}
	if c.LocalVPN.HTTP.Port == 0 {
		c.LocalVPN.HTTP.Port = Default().LocalVPN.HTTP.Port
	}
	if c.LocalVPN.SOCKS5.Server == "" {
		c.LocalVPN.SOCKS5.Server = Default().LocalVPN.SOCKS5.Server
	}
	if c.LocalVPN.SOCKS5.Port == 0 {
		c.LocalVPN.SOCKS5.Port = Default().LocalVPN.SOCKS5.Port
	}
}

func (c Config) Validate() error {
	if _, err := NormalizeLanguage(c.Language); err != nil {
		return err
	}
	for name, profile := range c.Profiles {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("profile name cannot be empty")
		}
		if profile.URL() == "" {
			return fmt.Errorf("profile %q has no usable proxy configuration", name)
		}
	}
	if c.ActiveProfile != "" {
		if _, ok := c.Profiles[c.ActiveProfile]; !ok {
			return fmt.Errorf("active profile %q does not exist", c.ActiveProfile)
		}
	}
	if c.LocalVPN.Enabled && c.LocalVPN.DefaultURL() == "" {
		return fmt.Errorf("local_vpn is enabled but has no usable proxy configuration")
	}
	return nil
}

// UpsertProfile is the only mutation helper for profile writes in the current design.
// This keeps activation and timestamp updates consistent across commands.
func (c *Config) UpsertProfile(name, proxyURL string, activate bool) ProxyProfile {
	if c.Profiles == nil {
		c.Profiles = map[string]ProxyProfile{}
	}

	name = strings.TrimSpace(name)
	profile := ProxyProfile{
		Name:      name,
		ProxyURL:  strings.TrimSpace(proxyURL),
		UpdatedAt: time.Now(),
	}
	profile.fillFieldsFromURL()
	c.Profiles[name] = profile
	if activate {
		c.ActiveProfile = name
	}
	return profile
}

func (c *Config) UpsertStructuredProfile(name, proxyType, server string, port int, username, password string, activate bool) ProxyProfile {
	if c.Profiles == nil {
		c.Profiles = map[string]ProxyProfile{}
	}

	profile := ProxyProfile{
		Name:      strings.TrimSpace(name),
		Type:      strings.TrimSpace(proxyType),
		Server:    strings.TrimSpace(server),
		Port:      port,
		Username:  strings.TrimSpace(username),
		Password:  strings.TrimSpace(password),
		UpdatedAt: time.Now(),
	}
	profile.ProxyURL = profile.URL()
	c.Profiles[profile.Name] = profile
	if activate {
		c.ActiveProfile = profile.Name
	}
	return profile
}

func (c Config) Profile(name string) (ProxyProfile, bool) {
	profile, ok := c.Profiles[strings.TrimSpace(name)]
	return profile, ok
}

func (c Config) Active() (ProxyProfile, bool) {
	if c.ActiveProfile == "" {
		return ProxyProfile{}, false
	}
	return c.Profile(c.ActiveProfile)
}

func (c Config) SortedProfiles() []ProxyProfile {
	result := make([]ProxyProfile, 0, len(c.Profiles))
	for _, profile := range c.Profiles {
		result = append(result, profile)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (c Config) ResolvedLanguage() string {
	language := strings.TrimSpace(c.Language)
	if language == "" {
		return "en"
	}
	return language
}

func NormalizeLanguage(language string) (string, error) {
	language = strings.TrimSpace(language)
	switch language {
	case "", "en":
		return "en", nil
	case "zh":
		return "zh", nil
	default:
		return "", fmt.Errorf("unsupported language %q", language)
	}
}

func (c *Config) SetLanguage(language string) {
	c.Language = strings.TrimSpace(language)
}

func (c *Config) RemoveProfile(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	if c.Profiles == nil {
		return false
	}
	if _, ok := c.Profiles[name]; !ok {
		return false
	}
	delete(c.Profiles, name)
	if c.ActiveProfile == name {
		c.ActiveProfile = ""
	}
	return true
}

func (p ProxyProfile) URL() string {
	if strings.TrimSpace(p.ProxyURL) != "" {
		return strings.TrimSpace(p.ProxyURL)
	}
	scheme := strings.TrimSpace(p.Type)
	host := strings.TrimSpace(p.Server)
	if scheme == "" || host == "" || p.Port <= 0 {
		return ""
	}
	value := &url.URL{
		Scheme: scheme,
		Host:   host + ":" + strconv.Itoa(p.Port),
	}
	if strings.TrimSpace(p.Username) != "" {
		if strings.TrimSpace(p.Password) != "" {
			value.User = url.UserPassword(strings.TrimSpace(p.Username), strings.TrimSpace(p.Password))
		} else {
			value.User = url.User(strings.TrimSpace(p.Username))
		}
	}
	return value.String()
}

func (p *ProxyProfile) fillFieldsFromURL() {
	parsed, err := url.Parse(strings.TrimSpace(p.ProxyURL))
	if err != nil {
		return
	}
	p.Type = parsed.Scheme
	p.Server = parsed.Hostname()
	port, err := strconv.Atoi(parsed.Port())
	if err == nil {
		p.Port = port
	}
	if parsed.User != nil {
		p.Username = parsed.User.Username()
		password, ok := parsed.User.Password()
		if ok {
			p.Password = password
		}
	}
}


func legacyPathFor(path string) string {
	if filepath.Base(path) != defaultSettingsFile {
		return path
	}
	return filepath.Join(filepath.Dir(path), legacyConfigFile)
}

func (v LocalVPNSettings) DefaultURL() string {
	if value := v.HTTP.URL("http"); value != "" {
		return value
	}
	return v.SOCKS5.URL("socks5")
}

func (v LocalProxyConfig) URL(proxyType string) string {
	profile := ProxyProfile{
		Type:     strings.TrimSpace(proxyType),
		Server:   strings.TrimSpace(v.Server),
		Port:     v.Port,
		Username: strings.TrimSpace(v.Username),
		Password: strings.TrimSpace(v.Password),
	}
	return profile.URL()
}
