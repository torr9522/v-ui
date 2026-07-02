# Changelog v1.0

## v1.0 Stable Release

Current stable version:

- `v-ui v1.0 Stable`
- Runtime compatibility layer: `x-ui`
- Release baseline before tag: `aa27526`

## Development Timeline

### 1. Baseline Import

- `1daa767` Import v-ui source baseline

Created the `v-ui` repository baseline while keeping the `x-ui` runtime compatibility surface.

### 2. Certificate Foundation

- `2ce3944` Add certificate management skeleton
- `e9c7354` Add manual certificate upload and HTTPS toggle

Established the certificate page, manual certificate upload flow, and HTTPS enable/disable baseline.

### 3. Runtime Compatibility Freeze

- `08955c1` Restore x-ui branding for v-ui runtime

Runtime compatibility intentionally remains:

- `x-ui.service`
- `/usr/local/x-ui`
- `/usr/bin/x-ui`
- `/xui`

### 4. HTTPS Toggle Fixes

- `0211ea6` Wait for panel listener after HTTPS toggle
- `67320c0` Fix HTTPS disable readiness detection
- `364f4e2` Document certificate phase 2 completion

Fixed listener wait timing, HTTP readiness detection, and protocol switching edge cases.

### 5. ACME HTTP-01 Mainline

- `22304a6` Plan certificate phase 3 ACME HTTP-01
- `c879fd7` Add ACME HTTP-01 certificate issue
- `1bc444c` Fix acme.sh executable discovery
- `35c234f` Use ACME fullchain for panel certificate

Completed:

- ACME HTTP-01 issuance flow
- `acme.sh` executable discovery
- production / staging issuance
- fullchain install for panel HTTPS

### 6. HTTPS Startup Resilience

- `79e9828` Add HTTPS startup self-heal

If HTTPS stays enabled in settings but certificate files are missing, panel startup now falls back to HTTP instead of failing hard.

### 7. Mainline Re-Alignment

- `7d2816c` Add Cloudflare API integration
- `5be8c63` Add Cloudflare DNS-01 certificate issue
- `1b83d73` Revert "Add Cloudflare API integration"
- `878ef88` Revert "Add Cloudflare DNS-01 certificate issue"
- `386e447` Document z-ui and v-ui certificate alignment

Cloudflare / DNS-01 exploration was intentionally rolled back to keep the certificate path aligned with the simpler `z-ui` style mainline:

- manual certificates
- HTTPS enable / disable
- HTTP-01
- fullchain
- startup self-heal
- renew timer

### 8. Renewal Chain

- `1c95a9b` Add certificate renewal timer

Added:

- `x-ui-cert-renew.service`
- `x-ui-cert-renew.timer`
- `x-ui cert-renew`
- manual certificate auto-skip

### 9. Certificate Module Freeze

- `f3a064e` Freeze certificate module
- `a42fe78` Document curl prerequisite for installer

Certificate module was frozen with only one remaining issue:

- `IssueHttp Idempotent`

Installer documentation was corrected to require `curl` before raw one-line installation.

### 10. Inbound TLS Certificate Reuse

- `d4d0f3d` Audit inbound TLS certificate selection
- `bb08ee9` Add inbound TLS certificate selection
- `27daefd` Fix inbound TLS certificate autofill
- `2039dae` Fix TLS certificate autofill context
- `8021dcb` Document certificate auto selection validation

Completed:

- `POST /xui/cert/listUsable`
- inbound TLS auto-fill of `certFile` / `keyFile`
- panel certificate summary in inbound TLS form
- live validation on `139.180.135.210`

### 11. Xray 26.5.3 and Modern UI Audit

- `8b53931` Audit Xray 26.5.3 protocol parameters
- `914b864` Audit inbound and routing UI modernization

Completed static audits for:

- protocol parameter compatibility against `Xray-core 26.5.3`
- inbound UX modernization gaps
- routing UX modernization gaps

### 12. Modern UI Phase 1

- `d200938` Improve inbound and routing UI layout
- `15a9fa6` Restore Trojan protocol option
- `62b60eb` Restore Trojan TLS tabs

Completed:

- inbound modal tab layout
- routing card view
- routing grouped form layout
- transport naming improvement such as `RAW`
- Trojan option and Trojan TLS tab regression fix

### 13. Inbound Compatibility and Share Link Fixes

- `f2bbd8b` Audit inbound protocol compatibility
- `1ace742` Fix Trojan and Shadowsocks share links

Completed:

- Trojan share link transport / TLS parameter completion
- safer Shadowsocks share link behavior for non-plain transports
- share link smoke test coverage

### 14. Repository Finalization

- `aa27526` Point install and runtime links to v-ui repo

Completed:

- installer links updated to `torr9522/v-ui`
- runtime links updated to `torr9522/v-ui`
- remote clone validation against the independent repository

## Final Capability Matrix

- inbound management
- outbound management
- routing
- AI routing template
- managed routing
- HTTPS panel
- manual certificates
- ACME HTTP-01
- ACME production / staging
- fullchain
- startup self-heal
- renew timer / renew service
- inbound TLS certificate auto selection
- Modern UI Phase 1
- Trojan / Shadowsocks share link fixes

## Live Validation and UAT

Live server validation:

- `5.226.49.140`: `PASS`
- `43.198.88.130`: `PASS`
- `139.180.135.210`: `PASS`

User-acceptance validation summary:

- install: `PASS`
- login: `PASS`
- certificate issue: `PASS`
- TLS inbound creation: `PASS`
- client connectivity: `PASS`
- inbound update: `PASS`
- inbound delete: `PASS`
- renew service path: `PASS`

## Known Issue

- `Issue001` `IssueHttp Idempotent` `P2`

## Stable Conclusion

`v-ui v1.0` is ready as the stable baseline.

Post-release policy:

- Bug Fix only
- Security Fix only
- Compatibility Fix only
- no certificate system expansion

Next development should return to core `v-ui` product features.
