# ADR-001: v-ui Project, x-ui Runtime And Display

- Status: `Accepted`
- Date: `2026-06-29`

## Context

The current codebase is developed in an isolated `v-ui` project workspace, but runtime compatibility remains anchored to `x-ui`.

The original plan included a runtime-layer rename, covering items such as:

- `install.sh`
- `x-ui.sh`
- systemd service names
- install paths
- database paths
- update flow
- release packaging
- test matrix

This work would add maintenance cost but would not directly deliver the product capabilities planned for the `v-ui` project.

## Decision

The project adopts the following strategy:

> Project codename independent, x-ui runtime/display frozen.

Project management layer is allowed to use the `v-ui` codename, including:

- repository identity
- local branch naming
- planning documents
- migration notes

Runtime and display layer remain compatibility-first and are frozen to `x-ui` for now.

## Runtime Freeze Scope

Unless the user explicitly issues `解除 Runtime Freeze`, the following must not be proactively renamed or changed for branding purposes:

- `go.mod`
- import path
- `/xui`
- `x-ui.service`
- `x-ui` command
- panel page titles
- sidebar branding
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

- repository codename and runtime display name are intentionally not unified
- panel-facing naming cannot be used for `v-ui` branding experiments

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
