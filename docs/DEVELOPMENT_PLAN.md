# v-ui Development Plan

## Strategy

`v-ui` is the project codename, while the runtime and display layer remain `x-ui`.

Project layer:

- `v-ui` naming in repository and documentation
- migration and planning records

Runtime and display layer:

- remains compatibility-first
- keeps `x-ui` page titles
- keeps `x-ui` menu branding
- does not enter runtime rename work in the current roadmap

See [ADR-001 Runtime Freeze](./ADR-001-runtime-freeze.md).

## Roadmap

### Phase 2: Outbounds

Status: completed for database, backend CRUD, and frontend CRUD.

Add outbound management.

### Phase 3: Routing

Status: completed for database, backend CRUD, and frontend CRUD.

Add routing rule management.

### Phase 3.5: Managed Config

Status: in progress.

Add the `managedOutboundsRouting` switch and let `x-ui` generate Xray `outbounds` and `routing.rules` from the Outbounds and Routing pages while keeping legacy template mode as the default.

### Phase 4: Template

Status: next.

Add routing templates, including:

- AI
- Streaming
- Telegram
- GitHub
- Google
- Gaming

### Phase 5: AI Traffic Split

Complete AI traffic splitting for:

- OpenAI
- ChatGPT
- Claude
- Gemini
- Grok
- Cursor
- Copilot
- Perplexity
- HuggingFace

### Phase 6: Node

Add node management and outbound generation from imported nodes.

Target protocols:

- `vmess`
- `vless`
- `trojan`
- `socks`
- `ss`
- `hy2`
- `tuic`

### Phase 7: Runtime API

Gradually add runtime hot-update support without breaking the runtime freeze boundary unless explicitly approved.

## Change Control

The following are frozen unless the user explicitly issues `解除 Runtime Freeze`:

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

## Standard Delivery Flow

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
