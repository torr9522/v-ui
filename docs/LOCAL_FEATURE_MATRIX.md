# C-UI Local Feature Matrix

- Current commit: `ef92b52 Add AI routing template`
- Verification time: `2026-06-29 UTC`
- Verification commands:
  - `go test ./...`
  - `go build ./...`
  - `rg` / `grep` static evidence scans over `database`, `web/service`, `web/controller`, `web/html/xui`, `web/assets/js/model`

## Summary

- Static verification: passed
- `go test ./...`: passed
- `go build ./...`: passed
- Browser verification: not performed
- Runtime verification: not performed

## Passed Items

### Outbound

- Model exists: [database/model/model.go](/tmp/c-ui-init/c-ui/database/model/model.go:69)
- AutoMigrate and system bootstrap exist: [database/db.go](/tmp/c-ui-init/c-ui/database/db.go:240)
- `direct` / `blocked` initialization exists: [database/db.go](/tmp/c-ui-init/c-ui/database/db.go:247)
- List API exists: [web/controller/outbound.go](/tmp/c-ui-init/c-ui/web/controller/outbound.go:34)
- CRUD API exists: [web/controller/outbound.go](/tmp/c-ui-init/c-ui/web/controller/outbound.go:35)
- Frontend CRUD exists: [web/html/xui/outbounds.html](/tmp/c-ui-init/c-ui/web/html/xui/outbounds.html:382)
- `direct` / `blocked` protection exists: [web/service/outbound.go](/tmp/c-ui-init/c-ui/web/service/outbound.go:277)
- Prevent delete when referenced by routing: [web/service/outbound.go](/tmp/c-ui-init/c-ui/web/service/outbound.go:249)

### Routing

- Model exists: [database/model/model.go](/tmp/c-ui-init/c-ui/database/model/model.go:84)
- AutoMigrate exists: [database/db.go](/tmp/c-ui-init/c-ui/database/db.go:288)
- List API exists: [web/controller/routing.go](/tmp/c-ui-init/c-ui/web/controller/routing.go:44)
- CRUD API exists: [web/controller/routing.go](/tmp/c-ui-init/c-ui/web/controller/routing.go:45)
- Reorder API exists: [web/controller/routing.go](/tmp/c-ui-init/c-ui/web/controller/routing.go:49)
- Frontend CRUD exists: [web/html/xui/routing.html](/tmp/c-ui-init/c-ui/web/html/xui/routing.html:68)
- `outboundTag` dropdown exists: [web/html/xui/routing.html](/tmp/c-ui-init/c-ui/web/html/xui/routing.html:162)
- Disabled rules excluded from managed config: [web/service/xray.go](/tmp/c-ui-init/c-ui/web/service/xray.go:290)

### Managed Config

- `managedOutboundsRouting` default false: [web/service/setting.go](/tmp/c-ui-init/c-ui/web/service/setting.go:25)
- Frontend setting field exists: [web/entity/entity.go](/tmp/c-ui-init/c-ui/web/entity/entity.go:37)
- Frontend model default false: [web/assets/js/model/models.js](/tmp/c-ui-init/c-ui/web/assets/js/model/models.js:215)
- `false` path uses legacy logic: [web/service/xray.go](/tmp/c-ui-init/c-ui/web/service/xray.go:89)
- `true` path uses managed logic: [web/service/xray.go](/tmp/c-ui-init/c-ui/web/service/xray.go:94)
- `direct` first outbound and `blocked` second outbound: [web/service/xray.go](/tmp/c-ui-init/c-ui/web/service/xray.go:242)
- API rule first routing rule: [web/service/xray.go](/tmp/c-ui-init/c-ui/web/service/xray.go:286)
- Outbounds sorted by `sort,id`: [web/service/outbound.go](/tmp/c-ui-init/c-ui/web/service/outbound.go:38)
- Routing rules sorted by `sort,id`: [web/service/routing.go](/tmp/c-ui-init/c-ui/web/service/routing.go:33)
- `TestConfig` validation helper exists: [web/service/xray.go](/tmp/c-ui-init/c-ui/web/service/xray.go:542)
- Managed save apply hook exists: [web/service/xray.go](/tmp/c-ui-init/c-ui/web/service/xray.go:558)
- Transactional rollback on managed validation failure exists for outbound: [web/service/outbound.go](/tmp/c-ui-init/c-ui/web/service/outbound.go:301)
- Transactional rollback on managed validation failure exists for routing: [web/service/routing.go](/tmp/c-ui-init/c-ui/web/service/routing.go:284)
- Restart after successful managed validation exists: [web/service/xray.go](/tmp/c-ui-init/c-ui/web/service/xray.go:569)

### AI Template

- `applyAiTemplate` API exists: [web/controller/routing.go](/tmp/c-ui-init/c-ui/web/controller/routing.go:50)
- Backend template apply function exists: [web/service/routing.go](/tmp/c-ui-init/c-ui/web/service/routing.go:301)
- AI template candidate builder exists: [web/service/routing.go](/tmp/c-ui-init/c-ui/web/service/routing.go:359)
- Dedup fingerprint exists: [web/service/routing.go](/tmp/c-ui-init/c-ui/web/service/routing.go:466)
- OpenAI domain evidence exists: [web/service/routing.go](/tmp/c-ui-init/c-ui/web/service/routing.go:372)
- GitHub Copilot domain evidence exists: [web/service/routing.go](/tmp/c-ui-init/c-ui/web/service/routing.go:437)
- Frontend AI template button exists: [web/html/xui/routing.html](/tmp/c-ui-init/c-ui/web/html/xui/routing.html:68)
- Frontend AI template modal submit exists: [web/html/xui/routing.html](/tmp/c-ui-init/c-ui/web/html/xui/routing.html:710)
- `created` / `skipped` / `applied` response handling exists: [web/controller/routing.go](/tmp/c-ui-init/c-ui/web/controller/routing.go:180)

## Not Verified

- Browser interaction was not verified.
  - Outbound and Routing pages were not opened in a real browser.
  - Modal behavior, button states, and success/error toasts were not interactively confirmed.
- Runtime Xray behavior was not verified.
  - Xray process was not started or restarted for this matrix run.
  - Managed config application was not exercised against a live local DB with real outbound/routing records.
  - AI template was not executed against a live DB to observe `created/skipped/applied` in practice.
- No install or deployment flow was verified.
  - `install.sh` was not run.
  - No systemd or production-path behavior was tested.

## Risk Items

- Snapshot tar emitted a common warning while `.git` changed during archive creation; tag, branch, and archive were still created.
- Managed config runtime correctness remains partially unverified without a live Xray binary execution path and seeded outbound/routing records.
- The API first-rule guarantee is statically present, but interaction with future custom system rules still depends on runtime scenarios not exercised here.
- Frontend success messaging for `applied=true/false` is statically present, but exact UX wording and modal refresh flow were not browser-validated.

## Next Step Suggestions

- Run a local seeded DB scenario and verify:
  - create outbound
  - create routing rule
  - enable `managedOutboundsRouting`
  - confirm managed config generation
  - confirm Xray restart path
- Run a local AI template scenario and verify duplicate suppression with repeated clicks.
- Do one browser pass over:
  - Outbound CRUD
  - Routing CRUD
  - AI template modal
  - Setting page managed switch
