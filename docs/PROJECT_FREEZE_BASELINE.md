# Project Freeze Baseline

## 1. Current Recoverable Development Baseline

- Branch: `v-ui`
- Development baseline tag: `v1.0.2-reality-baseline`
- Pre-freeze code baseline: `6f1b5a0 Preserve REALITY share metadata`
- Purpose: freeze the current REALITY progress as a recoverable development point without creating a new stable release

## 2. Current Git Tag

- Development baseline tag: `v1.0.2-reality-baseline`
- Stable tags retained:
  - `v1.0-stable`
  - `v1.0.1-stable`
  - `v1.0.2-stable`

## 3. Current Release

- Current stable GitHub Release: `v1.0.2-stable`
- Release URL: `https://github.com/torr9522/v-ui/releases/tag/v1.0.2-stable`
- Development baseline `v1.0.2-reality-baseline` is intentionally a tag only, not a stable release

## 4. Release / Tag / Doc Mapping

| Stage | Git Tag | GitHub Release | Repo Doc | Purpose |
| --- | --- | --- | --- | --- |
| Stable v1.0 | `v1.0-stable` | yes | `docs/RELEASE_v1.0.md` | first stable line |
| Stable v1.0.1 | `v1.0.1-stable` | yes | `docs/RELEASE_v1.0.1.md` | installer branding hotfix |
| Stable v1.0.2 | `v1.0.2-stable` | yes | `docs/RELEASE_v1.0.2.md` | current stable release |
| REALITY dev baseline | `v1.0.2-reality-baseline` | no | `docs/REALITY_PROGRESS_BASELINE.md` | current development freeze |

This mapping is the required recovery index for restoring any major stage through Git tag + GitHub Release + repository docs.

## 5. Completed Features

- independent `torr9522/v-ui` repository and installer chain
- runtime freeze on `x-ui` service / path / route compatibility
- certificate module
- ACME HTTP-01 flow
- renew timer and renew status path
- panel HTTPS self-heal path
- inbound TLS certificate auto selection
- Modern UI Phase 1
- Trojan / Shadowsocks share link fixes
- Custom Share Address Override
- REALITY UI skeleton
- REALITY frontend serialization
- REALITY minimal config generation for `VLESS + TCP + REALITY`
- REALITY share metadata retention via `realityShare`
- REALITY minimal `vless://` share export

## 6. Not Completed

- REALITY subscription export
- REALITY QR export
- full REALITY client export matrix
- XHTTP
- broader REALITY protocol matrix
- TLS Advanced export surface

## 7. Recommended Next Development Order

1. `feature/reality-subscription`
2. `feature/reality-qrcode`
3. `feature/xhttp-support`
4. REALITY export matrix hardening
5. TLS Advanced / client export cleanup

## 8. Repository Recovery Steps

Recover source:

```bash
git clone https://github.com/torr9522/v-ui.git
cd v-ui
git checkout v-ui
```

Recover build verification:

```bash
node tests/share_link_smoke.js
go test ./...
go build ./...
bash -n install.sh
bash -n x-ui.sh
```

Recover install verification:

```bash
apt update && apt install -y curl
bash <(curl -Ls https://raw.githubusercontent.com/torr9522/v-ui/v-ui/install.sh)
```

Recover a specific historical stage:

```bash
git checkout <tag>
```

Recommended tags:

- `v1.0-stable`
- `v1.0.1-stable`
- `v1.0.2-stable`
- `v1.0.2-reality-baseline`

## 9. Local Artifact Policy

The project does not require local snapshots, local tarballs, local build caches, local temporary browser scripts, or local Codex work directories to continue development.

Local paths such as the following are disposable after this freeze:

- `/tmp/v-ui-init`
- `/tmp/v-ui-snapshots`
- local `backup/*` branches
- local `snapshot-before-*` tags
- local build artifacts
- local browser / UAT scratch directories

They are development conveniences only, not required repository state.

## 10. Recoverable Development Baseline Statement

This repository is now expected to be recoverable from GitHub alone:

- source code is tracked in Git
- frontend code is tracked in Git
- Go backend code is tracked in Git
- installer / management scripts are tracked in Git
- tests are tracked in Git
- design / audit / freeze / roadmap documents are tracked in Git
- stable historical releases are preserved in GitHub Releases
- major stages are anchored by Git tags

After deleting local source, local snapshots, local caches, and local build outputs, development should still be restorable from:

- GitHub repository
- Git tags
- GitHub Releases
- repository documentation
