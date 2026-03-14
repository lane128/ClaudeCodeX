# Feature 002: Proxy Configuration Assistant

## Objective

Provide a safe and repeatable way for users to save, inspect, test, and export proxy settings needed by Claude Code without manually editing shell startup files every time.

## Primary Commands

- `ccx proxy init`
- `ccx proxy list`
- `ccx proxy env`
- `ccx proxy test`

## User Stories

- As a developer, I want to save a proxy profile once and reuse it across sessions.
- As a developer, I want to quickly export the right shell variables for my current terminal.
- As a developer, I want to verify whether a specific proxy can actually reach known public test endpoints before I use it in Claude Code.
- As a developer with multiple networks, I want to mark one proxy profile as active so diagnostics can use it by default.

## Functional Requirements

### Profile Management

- save a named proxy profile with URL
- support `http`, `https`, and `socks5` proxy schemes
- mark a profile as active during creation
- list saved profiles and clearly mark the active one
- store configuration in a user-level config file

### Shell Integration

- output shell-specific environment variable snippets
- support `zsh`, `bash`, `sh`, and `fish`
- export `HTTPS_PROXY`, `HTTP_PROXY`, and `ALL_PROXY`
- avoid mutating the parent shell process directly

### Proxy Testing

- validate proxy URL syntax before probing
- send an HTTP request through the proxy to a target URL
- report success, degraded, or failed result
- include duration and HTTP status when available
- support JSON output for automation

### Integration With Diagnostics

- `ccx doctor` should use the CLI `--proxy` override first
- then prefer existing environment variables
- then fall back to the active saved profile

## Non-Functional Requirements

- configuration format should remain human-readable
- saved config should be easy to back up or inspect
- shell output should be copy-paste friendly
- the command set should work without third-party Go dependencies

## Acceptance Criteria

- user can create a proxy profile in one command
- user can list existing profiles and see the active profile
- user can output shell exports for a chosen profile
- user can test a profile and get an actionable status
- `ccx doctor` can consume the active profile when env vars are absent
