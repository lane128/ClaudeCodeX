package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("default path: %v", err)
	}

	cfg := Config{}
	cfg.UpsertProfile("default", "http://127.0.0.1:7890", true)
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.ActiveProfile != "default" {
		t.Fatalf("unexpected active profile %q", loaded.ActiveProfile)
	}
	profile, ok := loaded.Active()
	if !ok {
		t.Fatal("expected active profile")
	}
	if profile.ProxyURL != "http://127.0.0.1:7890" {
		t.Fatalf("unexpected proxy url %q", profile.ProxyURL)
	}
}

func TestLoadMissingConfig(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("load missing config: %v", err)
	}
	if len(cfg.Profiles) != 0 {
		t.Fatalf("expected empty profiles, got %d", len(cfg.Profiles))
	}
	if cfg.ResolvedLanguage() != "en" {
		t.Fatalf("expected default language en, got %q", cfg.ResolvedLanguage())
	}
}

func TestSaveAndLoadLanguage(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("default path: %v", err)
	}

	cfg := Config{}
	cfg.SetLanguage("zh")
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.ResolvedLanguage() != "zh" {
		t.Fatalf("expected language zh, got %q", loaded.ResolvedLanguage())
	}
}

func TestRemoveProfile(t *testing.T) {
	cfg := Config{}
	cfg.UpsertProfile("default", "http://127.0.0.1:7890", true)

	if !cfg.RemoveProfile("default") {
		t.Fatal("expected profile removal to succeed")
	}
	if cfg.ActiveProfile != "" {
		t.Fatalf("expected active profile to be cleared, got %q", cfg.ActiveProfile)
	}
	if len(cfg.Profiles) != 0 {
		t.Fatalf("expected no profiles left, got %d", len(cfg.Profiles))
	}
}

func TestUpsertStructuredProfile(t *testing.T) {
	cfg := Config{}
	profile := cfg.UpsertStructuredProfile("home", "socks5", "1.1.1.1", 8888, "user", "pass", true)

	if profile.Type != "socks5" {
		t.Fatalf("expected type socks5, got %q", profile.Type)
	}
	if profile.Server != "1.1.1.1" || profile.Port != 8888 {
		t.Fatalf("unexpected server info: %+v", profile)
	}
	if profile.URL() != "socks5://user:pass@1.1.1.1:8888" {
		t.Fatalf("unexpected proxy url %q", profile.URL())
	}
}

func TestLoadFallsBackToLegacyConfigPath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)

	legacyPath := filepath.Join(root, ".ccx", "config.json")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("mkdir legacy dir: %v", err)
	}
	if err := os.WriteFile(legacyPath, []byte("{\"language\":\"zh\"}\n"), 0o644); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.ResolvedLanguage() != "zh" {
		t.Fatalf("expected language zh from legacy config, got %q", cfg.ResolvedLanguage())
	}
}

func TestEnsureCreatesSettingsFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)

	cfg, path, err := Ensure("")
	if err != nil {
		t.Fatalf("ensure settings: %v", err)
	}
	if path != filepath.Join(root, ".ccx", "settings.json") {
		t.Fatalf("unexpected settings path %q", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected settings file to exist: %v", err)
	}
	if cfg.ResolvedLanguage() != "en" {
		t.Fatalf("expected default language en, got %q", cfg.ResolvedLanguage())
	}
	if cfg.LocalVPN.HTTP.Server != "127.0.0.1" || cfg.LocalVPN.HTTP.Port != 7890 {
		t.Fatalf("unexpected default local vpn config: %+v", cfg.LocalVPN)
	}
}

func TestLocalVPNDefaultURL(t *testing.T) {
	vpn := LocalVPNSettings{
		Enabled:  true,
		HTTP: LocalProxyConfig{
			Server:   "127.0.0.1",
			Port:     7890,
			Username: "user",
			Password: "pass",
		},
		SOCKS5: LocalProxyConfig{
			Server: "127.0.0.1",
			Port:   7891,
		},
	}
	if vpn.DefaultURL() != "http://user:pass@127.0.0.1:7890" {
		t.Fatalf("unexpected local vpn url %q", vpn.DefaultURL())
	}
}
