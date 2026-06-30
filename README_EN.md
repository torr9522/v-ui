# v-ui Project

[简体中文](./README.md) | ENGLISH

`v-ui` is the project codename. The runtime layer, service name, command name, install path, and panel display name remain `x-ui`. This codebase already includes outbound management, routing management, an AI traffic split template, and managed Xray config generation driven by `managedOutboundsRouting`.

## Main Features

- Outbounds: list page, CRUD API, frontend CRUD, system entry protection, and routing reference protection
- Routing: list page, CRUD API, reorder support, enable/disable support
- AI traffic split template for OpenAI, Claude, Gemini, Grok, Perplexity, Poe, Cursor, GitHub Copilot, and HuggingFace
- Managed config mode: `x-ui` can generate `outbounds` and `routing.rules` directly
- Source-install bootstrap: when no prebuilt `x-ui` binary exists, `install.sh` automatically installs Go, gcc, git, tar, curl, unzip, and `file`, then builds with `CGO_ENABLED=1`

## Runtime Freeze

Runtime compatibility stays intentionally unchanged:

- binary name: `x-ui`
- systemd service: `x-ui.service`
- install directory: `/usr/local/x-ui`
- web route prefix: `/xui`

This is a compatibility decision, not an error. See [docs/ADR-001-runtime-freeze.md](./docs/ADR-001-runtime-freeze.md).

Additional note:

- repository / branch / documentation may use the `v-ui` project codename
- panel titles, sidebar branding, and runtime-facing names remain `x-ui`

## Installation

Default install command:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/torr9522/c-ui/c-ui/install.sh)
```

English installer:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/torr9522/c-ui/c-ui/install_en.sh)
```

Notes:

- If you clone this repository and run `install.sh`, it will prefer the local source tree and the bundled `bin/` assets
- If you use the raw install command, `install.sh` downloads the source archive from `torr9522/c-ui` and bootstraps the build on the target host
- The installation flow no longer depends on raw links from the legacy `n-ui` repository

## Bundled Assets

The repository now carries the core runtime assets needed for installation and further development:

- `bin/xray-linux-amd64`
- `bin/xray-linux-arm64`
- `bin/geoip.dat`
- `bin/geosite.dat`
- `x-ui.service`
- `x-ui.sh`
- `x-ui_en.sh`
- `xui-portlimit-sync.sh`
- `xui-portlimit-sync.service`
- `xui-portlimit-sync.timer`
- `scripts/bbr.sh`
- `scripts/acme_install.sh`

See [docs/UPSTREAM_ASSETS.md](./docs/UPSTREAM_ASSETS.md) for sizes, source notes, and the remaining external links.

## Validation

The current repository state has already passed real-machine validation on two Debian 11 test servers:

- `5.226.49.140`
- `43.198.88.130`

Validated items include:

- source-install bootstrap
- `managedOutboundsRouting`
- outbound CRUD
- routing CRUD
- AI template deduplication
- `ValidateManagedXrayConfig`
- `RestartXray`

Detailed records:

- [docs/LOCAL_FEATURE_MATRIX.md](./docs/LOCAL_FEATURE_MATRIX.md)
- [docs/LOCAL_RUNTIME_MATRIX.md](./docs/LOCAL_RUNTIME_MATRIX.md)

## Project Docs

- [BASELINE.md](./BASELINE.md)
- [docs/DEVELOPMENT_PLAN.md](./docs/DEVELOPMENT_PLAN.md)
- [docs/ADR-001-runtime-freeze.md](./docs/ADR-001-runtime-freeze.md)
- [docs/UPSTREAM_ASSETS.md](./docs/UPSTREAM_ASSETS.md)

## Credits

- [vaxilu/x-ui](https://github.com/vaxilu/x-ui)
- [XTLS/Xray-core](https://github.com/XTLS/Xray-core)
- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)

## Stargazers

[![Stargazers over time](https://starchart.cc/torr9522/c-ui.svg)](https://starchart.cc/torr9522/c-ui)
