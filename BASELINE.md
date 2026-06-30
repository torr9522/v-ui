## V-UI Baseline

- Source project: `torr9522/c-ui`
- Source repository: `https://github.com/torr9522/c-ui`
- Source branch: `c-ui`
- Source commit: `f4076be4ec0d2fbce0ea2d722eb4da13e83adaa3` (`f4076be`)
- New project name: `v-ui`
- Display name: `V-UI`

## Runtime Freeze

- Runtime binary remains `x-ui`
- systemd service remains `x-ui.service`
- CLI command remains `x-ui`
- Install directory remains `/usr/local/x-ui`
- Web route remains `/xui`
- Go module name and import path stay unchanged for now

## Current Product Base

- Outbound CRUD
- Routing CRUD
- AI Traffic Split template
- `managedOutboundsRouting`
- Source install bootstrap with Go + CGO toolchain handling

## New Project Goal

Build `v-ui` as an isolated successor project on top of the current `c-ui` baseline, while keeping `x-ui` runtime compatibility, and add domain / certificate management inspired by `torr9522/z-ui`.

## Scope Of First Round

- Project isolation only
- Read-only audit of `z-ui` certificate/domain logic
- Migration design for `v-ui`
- No feature code yet
- No remote configured in this local repository
