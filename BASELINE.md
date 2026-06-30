## C-UI Baseline

- Source repository: https://github.com/torr9522/c-ui
- Source branch: `c-ui`
- Source commit: `2768ef19ba3587783c139495e5f722bed157ba28`
- New project name: `c-ui`
- Display name: `C-UI`
- Current phase: `Phase 1.5 Baseline Imported`
- Current remote: `none`
- Current strategy: `Brand independent, Runtime frozen`
- Current goal: start `Outbounds` architecture and database design without runtime-layer renaming

### Not Changed In This Phase

- `go.mod` module
- import path
- `/xui` route
- systemd service
- install directory
- database path
- runtime scripts

### Runtime Freeze

Unless the user explicitly issues `解除 Runtime Freeze`, the following remain frozen:

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

### Development Priority

1. `Outbounds`
2. `Routing`
3. `Template`
4. `AI Traffic Split`
5. `Node`
6. `Runtime API`
