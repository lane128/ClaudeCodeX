# Product Definition

## Product Positioning

ClaudeCodeX (`ccx`) is a command-line toolkit focused on helping developers use Claude Code reliably in network-constrained environments, especially in China.

It is not a generic proxy manager. Its value is that it understands the failure modes developers hit specifically when running Claude Code and can detect, explain, and remediate them with a guided workflow.

## Core Problems

### 1. Network Connectivity

Users often cannot determine whether failures are caused by outbound blocking, TLS handshake issues, unstable routes, or local environment mistakes.

### 2. Proxy Configuration

Proxy settings are fragmented across shell environments, package managers, GUI apps, and Claude Code runtime requirements. Users need a unified setup flow.

### 3. IP / DNS Risk Control

Even when connectivity exists, requests may fail because of DNS poisoning, incorrect resolver choice, low-quality exit IPs, or region-based risk controls.

## Product Principles

- Diagnose before mutating environment
- Prefer reversible changes
- Show exact cause, not vague status
- Provide both interactive and scriptable CLI modes
- Make all changes inspectable and auditable

## User Outcomes

- Quickly identify why Claude Code cannot connect
- Apply the correct proxy or DNS fix in minutes
- Switch between home, office, and global network profiles safely
- Export enough diagnostics for self-debugging or support

## Functional Areas

### A. Diagnostics

- Detect local network basics
- Check Claude Code dependent endpoints
- Validate TCP, TLS, HTTP, DNS, and proxy layers
- Produce a human-readable diagnosis report

### B. Proxy Setup

- Configure HTTP/HTTPS/SOCKS proxy
- Generate shell-specific export commands
- Persist profile-based settings
- Validate proxy reachability and authentication

### C. DNS / IP Risk Control

- Inspect resolver chain and effective DNS
- Compare DNS results across resolvers
- Detect polluted or mismatched answers
- Recommend safer resolver and route choices

### D. Profiles and Automation

- Save multiple environment profiles
- Switch profiles with one command
- Run preflight checks before launching workflows
- Support CI and non-interactive usage

### E. Troubleshooting and Recovery

- Explain common error signatures
- Suggest ordered remediation steps
- Roll back previously applied settings
- Export debug bundle

## Suggested Command Surface

- `ccx doctor`
- `ccx ping`
- `ccx proxy init`
- `ccx proxy test`
- `ccx proxy use <profile>`
- `ccx dns inspect`
- `ccx dns compare`
- `ccx risk check`
- `ccx profile list`
- `ccx profile switch <name>`
- `ccx report export`

## Non-Goals For V1

- Building or bundling a VPN service
- Bypassing legal or enterprise security controls
- Supporting every editor or AI provider on day one
- Managing OS network stack settings with privileged automation by default
