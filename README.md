# ClaudeCodeX

[English](./README.md) | [简体中文](./README.zh-CN.md)

ClaudeCodeX (`ccx`) is a CLI toolkit that helps developers reliably use Claude Code in restricted or complex network environments.

Persistent local settings live in `~/.ccx/settings.json`. The file is created automatically on first run, and legacy `config.json` files are still read as fallback.

The settings file now includes a `local_vpn` block for user-managed local proxy data such as Clash or other VPN client listeners, with separate `http` and `socks5` entries.

## Core Goals

- Improve network connectivity for Claude Code in restricted regions
- Simplify proxy configuration across common developer environments
- Reduce failures caused by IP and DNS risk-control issues

## Target Users

- Developers in mainland China using Claude Code
- Engineers working in enterprise or campus networks with strict egress rules
- Users who need a repeatable setup instead of ad hoc shell scripts

## Initial Product Scope

1. Network diagnostics
2. Proxy configuration assistant
3. DNS and IP risk troubleshooting
4. Environment profiles and one-command switching
5. Health checks and guided fixes

## Installation

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/lane128/ClaudeCodeX/main/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/lane128/ClaudeCodeX/main/install.ps1 | iex
```

### Build from source

```bash
go build -o ./bin/ccx ./cmd/ccx
```

## Quick Start

```bash
ccx doctor
ccx env
ccx test --proxy http://127.0.0.1:7890
ccx test          # compares exit IP against expected_ip in settings.json
ccx setting
ccx language      # interactive picker; use --zh or --en to skip
```

## Current Status

The current command surface is intentionally compact:

- `ccx doctor`
- `ccx env`
- `ccx test`
- `ccx setting`
- `ccx language`

`ccx doctor` currently checks:

- proxy-related environment variables
- DNS resolution for `anthropic.com`, `claude.ai`, and `claude.com` (configurable via `doctor.targets` in settings.json)
- direct TCP connectivity on port 443
- direct TLS handshake status
- HTTP reachability, optionally through a configured proxy

`ccx env` currently shows:

- the currently effective proxy
- the source of that proxy: environment or saved active profile
- active saved proxy profile, when present
- shell export snippets with `--shell`
- shell unset snippets with `--shell ... --unset`
- storing long-lived defaults in local `settings.json`
- reading fallback proxy data from `settings.local_vpn` when configured

`ccx test` currently supports:

- validating proxy reachability against a target URL (default: `https://www.anthropic.com/`)
- reading the proxy from `--proxy`, environment variables, or the active saved profile
- two-state result: success (green) or failed (red), with color output in interactive terminals
- falling back to direct connection when no proxy is configured (useful for global VPN or TUN mode)
- checking whether the proxy host:port is reachable
- checking exit IP through multiple IP check URLs
- verifying the exit IP matches `expected_ip` in settings.json (set it manually to enable the check)

`ccx language` currently supports:

- opening an interactive language picker when you run `ccx language`
- switching output language between English and Chinese
- saving the choice into local config
- keeping `--zh` and `--en` as a non-interactive fallback
- defaulting to English when no language has been configured

`ccx setting` currently supports:

- showing the current settings file path
- printing the full settings file as JSON
- validating whether the current settings are usable
- automatically migrating existing config files to include the `local_vpn` block on first run

The `local_vpn` block is always present in the settings file and contains separate `http` and `socks5` entries for user-managed local proxy listeners such as Clash. Set `enabled: true` and update the server/port values to activate it.

## Documentation

- Product definition: `docs/product.md`
- Roadmap: `docs/roadmap.md`
- Feature specs: `docs/features/`
