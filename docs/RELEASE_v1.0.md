# v-ui v1.0 Stable

## Release Summary

- Version: `v-ui v1.0 Stable`
- Stable branch: `v-ui`
- Stable commit baseline: `aa27526`
- Runtime layer remains: `x-ui`

## Completed in v1.0

- Certificate module completed and frozen
- ACME HTTP-01 completed
- Renew timer and renew service completed
- Inbound TLS certificate auto selection completed
- Modern UI Phase 1 completed
- Trojan / Shadowsocks share link fixes completed
- Live UAT on `139.180.135.210` completed

## Stable Scope

Current stable release includes:

- inbound management
- outbound management
- routing management
- HTTPS panel
- manual certificate upload
- ACME HTTP-01 issuance
- fullchain install for panel HTTPS
- startup HTTPS self-heal
- certificate renew timer
- inbound TLS panel-certificate auto fill
- Modern UI Phase 1 for inbound and routing

## Known Issue

- `IssueHttp Idempotent`
- Priority: `P2`
- Status: open
- Release impact: does not block `v1.0 Stable`

## Release Validation

Validated items for this stable line:

- local `go test ./...`
- local `go build ./...`
- `node tests/share_link_smoke.js`
- installer shell syntax validation
- runtime shell syntax validation
- remote clone verification
- live server validation on `139.180.135.210`

## Release Conclusion

`v-ui v1.0 Stable` is ready for production baseline use.

The runtime compatibility surface stays intentionally unchanged:

- `x-ui.service`
- `/usr/local/x-ui`
- `/usr/bin/x-ui`
- `/xui`
