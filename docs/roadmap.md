# Roadmap

## Phase 0: Foundation

- Initialize repository and project conventions
- Define CLI information architecture
- Define feature specs and delivery sequence

## Phase 1: Connectivity Diagnostics

Goal: help users determine why Claude Code cannot connect.

Deliverables:

- `ccx doctor`
- endpoint reachability checks
- proxy-aware connectivity detection
- structured output and exit codes

## Phase 2: Proxy Configuration

Goal: make proxy setup consistent and low-risk.

Deliverables:

- `ccx proxy init`
- `ccx proxy test`
- profile persistence
- shell integration snippets

## Phase 3: DNS and IP Risk Troubleshooting

Goal: identify DNS pollution, poor exit IP quality, and region mismatches.

Deliverables:

- `ccx dns inspect`
- `ccx dns compare`
- `ccx risk check`

## Phase 4: Profiles and Workflow Automation

Goal: reduce repeated setup cost across multiple network environments.

Deliverables:

- named profiles
- one-command switching
- preflight checks
- report export

## Phase 5: Experience Hardening

- better error taxonomy
- richer debug bundle
- docs and onboarding polish
- cross-platform verification
