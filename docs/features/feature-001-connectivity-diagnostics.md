# Feature 001: Connectivity Diagnostics

## Objective

Provide a reliable first-step diagnostic command so users can identify whether Claude Code failures are caused by local networking, proxy setup, DNS issues, TLS problems, or upstream endpoint reachability.

## Primary Command

- `ccx doctor`

## User Stories

- As a developer, I want to know whether my machine can reach required endpoints so I can separate local problems from upstream ones.
- As a developer using a proxy, I want to know whether the proxy is actually being used and whether it can reach the target.
- As a developer in China, I want to know whether the failure is more likely due to DNS pollution, blocked transport, or TLS handshake issues.
- As a support/debug user, I want exportable diagnostic output so I can share facts instead of screenshots.

## Functional Requirements

### Input

- Run without arguments for standard checks
- Accept `--proxy <url>` to override environment proxy
- Accept `--json` for machine-readable output
- Accept `--verbose` for detailed trace
- Accept `--timeout <seconds>` to control probe timeout

### Checks

- detect OS and shell environment
- inspect relevant proxy environment variables
- resolve configured target hostnames through current resolver
- test DNS answers for consistency and timing
- test TCP connectivity to required ports
- test TLS handshake and certificate chain visibility
- test HTTP(S) request success for selected endpoints
- indicate whether a proxy was used for each probe

### Output

- overall status: success, degraded, failed
- per-layer status: env, dns, tcp, tls, http, proxy
- probable root-cause summary
- actionable next-step suggestions
- optional JSON schema stable enough for automation

## Non-Functional Requirements

- default runtime under 10 seconds on a healthy network
- each failed probe must return a reason, not only a boolean
- command must exit non-zero on failed or degraded states
- output should be understandable without network expertise

## Error Classification

- no proxy configured
- proxy configured but unreachable
- proxy authentication failed
- DNS resolution timeout
- DNS answer mismatch
- TCP connect timeout
- TLS handshake failed
- HTTP blocked or reset
- upstream returned non-success response

## Acceptance Criteria

- user can run `ccx doctor` and receive a summary in one command
- user can tell whether proxy settings were detected and used
- output differentiates DNS failure from TCP or TLS failure
- JSON output includes probe-level statuses and timestamps
- failure messages include at least one next action

## Open Design Questions

- whether default endpoints should stay on neutral public targets like Google or become provider-specific later
- whether to support parallel probes in V1
- whether degraded should map to exit code `1` and failed to `2`

## Implementation Notes

- build this feature first
- keep probe architecture extensible so later commands can reuse it
- separate probe execution from presentation layer
