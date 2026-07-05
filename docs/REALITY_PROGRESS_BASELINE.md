# REALITY Progress Baseline

## Baseline

- Code baseline commit: `6f1b5a0`
- Branch: `v-ui`
- Scope: development baseline only
- Stable release status: not part of a new stable release yet

## Completed

- REALITY UI skeleton
- REALITY frontend serialization
- REALITY minimal config generation
- `VLESS + TCP + REALITY + xtls-rprx-vision` live validation passed
- REALITY share metadata via `streamSettings.realityShare`
- REALITY `vless://` share link export
- Custom Share Address compatibility with REALITY share export
- final `config.json` excludes `realityShare`

## Not Completed

- REALITY subscription export
- REALITY QR code export
- full REALITY client export matrix
- XHTTP
- TLS Advanced UI / export

## Remote Validation

Validated host:

- `139.180.135.210`

Validated results:

- REALITY minimal config: `PASS`
- REALITY share link: `PASS`
- Custom Share Address with REALITY: `PASS`

## Current Scope Boundary

Current REALITY support is intentionally limited to:

- protocol: `VLESS`
- transport: `RAW(TCP)`
- security: `REALITY`
- flow: `xtls-rprx-vision`

The current baseline does not include:

- REALITY subscription generation
- REALITY QR export
- Trojan REALITY
- XHTTP

## Next Suggested Branches

- `feature/reality-subscription`
- `feature/reality-qrcode`
- `feature/xhttp-support`

## Notes

- `realityShare` is panel-private metadata for export only.
- `realityShare` may be stored in DB.
- `realityShare` must never appear in final Xray runtime config.
