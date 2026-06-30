# ADR-001: Brand Independent, Runtime Frozen

- Status: `Accepted`
- Date: `2026-06-29`

## Context

`c-ui` has already been isolated from the original `n-ui` repository and now has its own branding baseline and source baseline.

The original plan included a runtime-layer rename, covering items such as:

- `install.sh`
- `x-ui.sh`
- systemd service names
- install paths
- database paths
- update flow
- release packaging
- test matrix

This work would add maintenance cost but would not directly deliver the core product capabilities planned for `c-ui`.

## Decision

`c-ui` adopts the following strategy:

> Brand independent, Runtime frozen.

Brand layer is allowed to be `C-UI`, including:

- repository identity
- README
- logo and naming
- web page titles
- web menu branding

Runtime layer remains compatibility-first and is frozen for now.

## Runtime Freeze Scope

Unless the user explicitly issues `解除 Runtime Freeze`, the following must not be proactively renamed or changed for branding purposes:

- `go.mod`
- import path
- `/xui`
- `x-ui.service`
- `x-ui` command
- database tables
- setting keys
- `install.sh`
- `x-ui.sh`
- Xray config schema

## Consequences

Benefits:

- avoids non-functional rename churn
- keeps upgrade and compatibility risk low
- keeps effort focused on feature delivery

Tradeoffs:

- runtime naming remains mixed with `x-ui` / `n-ui` compatibility surfaces
- brand and runtime names are intentionally not fully unified

## Feature Priority After This ADR

1. Phase 2: `Outbounds`
2. Phase 3: `Routing`
3. Phase 4: `Template`
4. Phase 5: `AI Traffic Split`
5. Phase 6: `Node`
6. Phase 7: `Runtime API`

## Delivery Rules

Each development stage should follow this order:

1. safety check
2. create git tag
3. create backup branch
4. create `tar.gz` snapshot
5. development
6. `go test ./...`
7. `go build ./...`
8. minimal verification
9. git commit
10. rollback instructions

If a stage fails, prefer rollback using:

- `git reset`
- git tag
- backup branch
