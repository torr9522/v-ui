# v1.0.2 Release Candidate Freeze Audit

Date: 2026-07-03 UTC

Audit target: `v-ui` release candidate after `3d5d313 Remove Cloudflare DNS certificate flow from scripts`

Scope:

- install / update / management scripts
- runtime download chain
- frontend templates and embedded assets
- inbound / routing / cert / settings entry points
- share link generation regression check
- docs and release-facing references

## 1. Result Summary

- `P0`: 0
- `P1`: 0
- `P2`: present
- `P3`: present

Conclusion:

- no release blocker was found in runtime, installer, script, or frontend entry paths
- `v1.0.2-stable` can proceed from the current code state

## 2. Blocking Audit Items

### P0

None.

### P1

None.

The previous blocker was removed:

- `x-ui.sh` no longer exposes Cloudflare DNS certificate mode
- `x-ui_en.sh` no longer exposes Cloudflare DNS certificate mode
- no `dns_cf` / `CF_*` certificate issue path remains in the management scripts

## 3. P2 Items

These items do not block release, but should be cleaned in a later documentation pass:

1. Historical release and changelog files still mention old project names.
   - `CHANGELOG.md`
   - `docs/RELEASE_v0.1.0.md`

2. Some archived audit documents still contain `c-ui` / `n-ui` / `z-ui` / `3x-ui` comparison history.
   - examples: `docs/LOCAL_FEATURE_MATRIX.md`, `docs/LOCAL_RUNTIME_MATRIX.md`, `docs/Z_UI_*`, `docs/MODERN_UI_REVIEW.md`

3. `config/version` is still `0.3.2`, which is currently only used as a cache-busting string for static assets.
   - not a runtime blocker
   - should be normalized in a later release hygiene pass

## 4. P3 Items

1. `web/html/xui/index.html` still logs polling failures through `console.error(e)`.
   - useful during failures
   - not a release blocker

2. `web/html/xui/form/reality_settings.html` uses `www.cloudflare.com:443` as an example `dest`.
   - this is only placeholder text
   - not related to Cloudflare certificate flow

## 5. Download Chain Review

Confirmed acceptable:

- raw installer entry points to `torr9522/v-ui`
- source archive points to `https://github.com/torr9522/v-ui/archive/refs/heads/v-ui.tar.gz`
- release fallback points to `torr9522/v-ui` / `v-ui-assets`
- Xray update fallback in `web/service/server.go` points to `torr9522/v-ui/releases/download/v-ui-assets`
- bundled `bin/` assets are present and embedded runtime templates are included in the final build

Confirmed removed from active install/runtime paths:

- `c-ui`
- `n-ui`
- `z-ui`
- `3x-ui`
- `bin456789`
- `MHSanaei`
- Cloudflare DNS script certificate flow

## 6. Frontend / Embed Review

Confirmed:

- `web/web.go` embeds `web/assets/*`, `web/html/*`, and `web/translation/*`
- current HTML additions such as:
  - `web/html/xui/cert.html`
  - `web/html/xui/inbound_info_modal.html`
  - `web/html/common/qrcode_modal.html`
  - `web/html/xui/form/reality_settings.html`
  are inside the embedded tree
- `/xui` routes exist for:
  - dashboard
  - inbounds
  - outbounds
  - routing
  - cert
  - access-ips
  - setting

No missing template reference or broken page entry was found in this static audit.

## 7. Share Link Regression Review

Validated with `tests/share_link_smoke.js`:

- VMess remains `vmess://` + Base64(JSON)
- VLESS share links pass
- Trojan share links pass
- Shadowsocks share links pass
- custom share address override still passes current constraints

## 8. Validation Commands

Passed:

- `go test ./...`
- `go build ./...`
- `bash -n install.sh`
- `bash -n install_en.sh`
- `bash -n x-ui.sh`
- `bash -n x-ui_en.sh`
- `node tests/share_link_smoke.js`

## 9. Final Recommendation

- release status: ready
- recommended tag target: `v1.0.2-stable`
- if development stops today, this state is acceptable as a stable baseline
- next cleanup should focus on historical docs and release-note consistency, not runtime code
