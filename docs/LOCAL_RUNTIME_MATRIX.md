# C-UI Local Runtime Matrix

- Current commit under test: `0ddfcd6 Document local feature matrix`
- Test time: `2026-06-29 UTC`
- Test directory: `/tmp/c-ui-runtime-test`
- Temporary database: `/tmp/c-ui-runtime-test/runtime.db`
- Temporary binary: `/tmp/c-ui-runtime-test/c-ui-test`
- Temporary runtime helper:
  - `/tmp/c-ui-runtime-test/runtime_matrix.go`
  - `/tmp/c-ui-runtime-test/runtime_matrix_bin`

## Test Scope

- `go test ./...`
- `go build -o /tmp/c-ui-runtime-test/c-ui-test .`
- Isolated service-layer runtime checks against a temporary SQLite DB via direct Go calls
- `xray.TestConfig` verification against the repository-local `bin/xray-linux-amd64`

## Environment Notes

- Used temporary DB: yes
- Started panel service: no
- Started Xray process: no
- Touched systemd: no
- Touched production `/etc/x-ui` or `/usr/local/x-ui`: no
- Touched remote servers: no

## Scenario Results

### A. `managedOutboundsRouting=false`

Result: passed

Observed results:

- Added outbound `ai-proxy` through service layer with `applied=false`
- Added routing rule `domain:openai.com -> ai-proxy` with `applied=false`
- Generated config remained on legacy template mode
- Generated config outbound count remained `2`
- Generated config routing rule count remained `3`
- `ai-proxy` did not appear in final config outbounds
- `domain:openai.com` rule did not appear in final config routing rules

Conclusion:

- Static and runtime behavior both indicate that `managedOutboundsRouting=false` preserves old template mode and ignores Outbound/Routing table data during config generation.

### B. `managedOutboundsRouting=true`

Result: partially passed

Prepared data:

- Outbound:
  - `direct`
  - `blocked`
  - `ai-proxy` (`socks`, local harmless example settings)
- Routing rules:
  - `domain:openai.com -> ai-proxy`
  - `domain:example.com -> direct`
  - intended disabled rule: `domain:disabled.example -> ai-proxy`

Observed results:

- `direct` was the first outbound
- `blocked` was the second outbound
- `ai-proxy` appeared in final outbounds
- `api -> api` was the first routing rule
- user rule order was stable by `sort,id`
- `domain:openai.com` pointed to `ai-proxy`
- `domain:example.com` pointed to `direct`

Unexpected result:

- the intended disabled rule was still present in final config
- temporary DB evidence showed `routing_rules.enabled=1` for the row labeled `Disabled Example`

Conclusion:

- managed config generation mostly works as intended
- there is a real runtime issue around storing or preserving `enabled=false` for routing rules, so the “disabled rule excluded” expectation did not pass in this run

### C. AI Template Dedup

Result: passed

Execution:

- ran `ApplyAITemplate(outboundTag=ai-proxy, sortBase=100, remarkPrefix=AI)` under `managedOutboundsRouting=false`

Observed results:

- first call: `created=9`, `skipped=0`, `applied=false`
- second call: `created=0`, `skipped=9`, `applied=false`
- rules for all target groups were found:
  - OpenAI
  - Anthropic
  - Google AI
  - xAI
  - Perplexity
  - Poe
  - Cursor
  - GitHub Copilot
  - HuggingFace
- all AI template rules used outbound tag `ai-proxy`

Conclusion:

- backend dedup behavior worked in isolated runtime verification

### D. `xray.TestConfig`

Result: passed after correcting working directory

Observed results:

- first attempt from `/tmp/c-ui-runtime-test` failed because `xray.TestConfig` resolves `bin/xray-linux-amd64` relative to current working directory
- second attempt from repository root `/tmp/c-ui-init/c-ui` passed successfully

Conclusion:

- `xray.TestConfig` works for the managed config in this repository when executed from the repo root
- there is path sensitivity in the current `xray.TestConfig` implementation

### E. `RestartXray`

Result: intentionally not executed

Observed results:

- no Xray process was started
- no panel service was started
- no production runtime was touched

Conclusion:

- production runtime behavior remains unverified by design in this matrix run

## Failure Reasons

- `enabled=false` routing rule exclusion did not pass in runtime verification because the row ended up stored with `enabled=1` in the temporary DB.
- initial `xray.TestConfig` attempt failed from the temp directory due relative binary path lookup; rerun from repo root succeeded.

## Unverified Items

- Browser verification was not performed.
- Panel HTTP service startup was not performed.
- `RestartXray` runtime execution was not performed.
- Remote server validation was not performed.
- Production `x-ui` coexistence behavior was not exercised.

## Risk Items

- Routing `enabled=false` persistence appears buggy in runtime conditions and should be treated as a real defect until fixed.
- `xray.TestConfig` depends on current working directory because it uses a relative binary path.
- Snapshot tar emitted the usual `.git` changed-while-reading warning during archive creation, though tag/branch/archive were still created.

## Next Step Suggestions

- Fix the routing `enabled=false` persistence/runtime issue first.
- After that, rerun this same isolated runtime matrix.
- Then do one browser validation pass for:
  - settings toggle
  - outbound CRUD
  - routing CRUD
  - AI template modal

## Bugfix Verification

- Bug identified:
  - `routing_rules.enabled=false` was being persisted as `enabled=1`, so disabled rules still appeared in managed `routing.rules`.
- Root cause:
  - `RoutingRule.Enabled` used `gorm:"default:true"`, which caused GORM create-path zero-value handling to omit explicit `false` on insert.
  - controller binding also used plain `bool`, so add/update requests could not distinguish omitted `enabled` from explicit `enabled=false`.
- Fix applied:
  - removed the `default:true` GORM tag from `RoutingRule.Enabled`
  - changed routing add/update/toggle request DTOs to use `*bool`
  - defaulted add requests to `enabled=true` only when the request omits the field
  - preserved explicit `enabled=false` on add/update
  - updated routing writes to use explicit field updates so `enabled=false` is not skipped
- Regression tests executed:
  - `go test ./web/service -run 'Test(AddRulePersistsExplicitDisabled|UpdateRulePersistsExplicitDisabled|ToggleRulePersistsExplicitDisabled|ManagedRoutingSkipsDisabledRules)$' -v`
  - `go test ./...`
  - `go build ./...`
- Regression results:
  - AddRule with explicit `enabled=false`: passed, DB persisted `false`
  - UpdateRule with explicit `enabled=false`: passed, DB persisted `false`
  - ToggleRule to `false`: passed, DB persisted `false`
  - managed routing build now filters disabled rules correctly: passed
- Current status:
  - `enabled=false` is now correctly persisted and filtered from managed routing config in service-layer regression verification
  - the full isolated local runtime matrix has not yet been rerun end-to-end after this fix
