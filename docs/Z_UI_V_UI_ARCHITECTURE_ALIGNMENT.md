# Z-UI / V-UI Certificate Architecture Alignment

## 1. Summary

This review answers one narrow question:

- can `z-ui` certificate / renewal logic be copied directly into `v-ui`?

Short answer:

- **partially**
- the runtime core is still highly compatible
- the certificate management shape is **not fully isomorphic**
- `z-ui` is **shell-first**
- `v-ui` is already **panel-first**

So the correct direction is:

- reuse the same runtime assumptions
- reuse the same HTTPS settings model
- reuse the same `acme.sh` + `systemd timer` pattern
- do **not** blindly transplant the whole `z-ui x-ui.sh cert` subsystem

## 2. High-Level Comparison

### 2.1 Common foundation

`z-ui` and `v-ui` still share the same core runtime model:

- Go backend
- Gin controllers
- service layer
- SQLite settings table
- `webCertFile` / `webKeyFile` as active panel HTTPS pointers
- `SIGHUP`-based `RestartPanel`
- `x-ui.service`
- `x-ui` runtime naming
- `Vue 2.6.12 + Ant Design Vue 1.7.2`

This means the following are structurally compatible:

- panel HTTPS boot logic
- settings-driven TLS switch
- `x-ui` shell command entry
- `systemd`-driven scheduled renewal

### 2.2 Main architectural divergence

The certificate workflow shape is different:

`z-ui`:

- shell-first
- certificate inventory under `/etc/x-ui/certs/<name>/`
- per-certificate `meta.json`
- issue / renew / import / set-panel-cert all centered in `x-ui.sh`
- panel web side only exposes certificate discovery/import helper APIs

`v-ui`:

- panel-first
- active panel certificate model centered in `web/service/cert.go`
- certificate files currently centered around:
  - `/usr/local/x-ui/cert/panel.crt`
  - `/usr/local/x-ui/cert/panel.key`
  - `/usr/local/x-ui/cert/acme-panel.crt`
  - `/usr/local/x-ui/cert/acme-panel.key`
- state stored in settings:
  - `webDomain`
  - `webCertStatus`
  - `webCertMode`
  - `webCertProvider`
  - `webCertExpireAt`
  - `webCertIssuer`
  - `webCertAutoRenew`
- HTTP-01 issue logic already lives in Go service, not shell menu

That is the critical split:

- same runtime
- different certificate orchestration layer

## 3. Detailed Comparison

| 模块 | z-ui | v-ui | 是否同构 | 能否直接照搬 | 原因 |
|---|---|---|---|---|---|
| settings 表 | SQLite `settings`，`webCertFile/webKeyFile` 驱动 HTTPS | SQLite `settings`，且已扩展 `webDomain/webCertStatus/webCertMode/...` | 部分同构 | 部分可 | 基础键兼容，v-ui 多了面板化状态键 |
| `webCertFile/webKeyFile` | 直接决定 HTTPS 是否启用 | 同样直接决定 HTTPS 是否启用 | 同构 | 可 | 这是最稳定的共通层 |
| web server HTTPS 启动 | 启动时读取 cert/key，存在即 TLS listener | 同样读取 cert/key；额外有 startup self-heal | 同构 | 可 | v-ui 还更安全 |
| `RestartPanel` | 发送 `SIGHUP`，主进程 Stop + Start | 同样 `SIGHUP` | 同构 | 可 | 信号和重建 listener 的机制一致 |
| `x-ui.sh` 菜单 | 完整 shell-first 证书管理入口 | 仍有旧脚本能力，但不是当前证书主线 | 非同构 | 不建议直接整体照搬 | v-ui 当前证书主线已经在 Go service / panel |
| `acme_install.sh` | `acme.sh` installer helper | 同样存在 installer helper | 同构 | 可 | 适合作为底层依赖安装 |
| systemd timer | `x-ui-cert-renew.service` + `x-ui-cert-renew.timer` | 当前无同名 renew timer | 可兼容 | 可借鉴 | 模式可直接借鉴，文件内容需按 v-ui 当前路径调整 |
| renew service | shell 调 `x-ui cert renew` | 当前未实现 | 非同构 | 不能原样照搬 | v-ui 没有 z-ui 的 cert inventory / meta.json 子系统 |
| 证书目录 | `/etc/x-ui/certs/<name>/` + `meta.json` | `/usr/local/x-ui/cert/` 单活动面板证书文件 | 非同构 | 不能直接照搬 | 存储模型不同 |
| 续期后重启 | `acme.sh --install-cert --reloadcmd ...` + shell `reload_after_cert_change` | 当前 Go service 负责 enable/disable 与 listener wait | 部分同构 | 部分可 | 策略可借鉴，但不应绕过现有 Go toggle 状态 |
| 前端页面 | 无完整 panel cert manager 主页面，更多是发现/导入辅助 | 已有完整 `/xui/cert` 页面 | 非同构 | 不需要照搬 | v-ui 已经比 z-ui 更适合当前目标 |

## 4. Isomorphism Judgment

### 4.1 What is isomorphic

The following layers are effectively isomorphic:

- process model
- signal reload model
- HTTPS startup model
- active cert path model
- `acme.sh` as ACME client
- `systemd timer` as renewal scheduler

These can be treated as direct carry-over assumptions.

### 4.2 What is not isomorphic

The following layers are not isomorphic:

- certificate inventory storage model
- shell-first certificate management
- `meta.json`-centric renewal bookkeeping
- `x-ui.sh cert` interactive certificate manager
- certificate discovery/import APIs as the only web-side interface

This means:

- `z-ui` renew timer idea is reusable
- `z-ui` full shell cert manager is not directly reusable

## 5. What Can Be Carried Over Directly

These are safe direct-carry items from `z-ui` into `v-ui`:

1. `acme.sh` remains the only supported ACME client.
2. Renewal should be scheduled by `systemd timer`, not by a new complex provider framework.
3. Renewal should target only ACME-managed certificates.
4. Imported/manual certificates should not enable ACME auto-renew.
5. Renewal should reinstall:
   - fullchain
   - key
6. Renewal should restart or reload `x-ui` only when the active panel certificate is involved.
7. Renewal should remain `x-ui` runtime compatible and shell-invokable.
8. The system should keep `webCertFile` / `webKeyFile` as the final active truth for panel HTTPS.

## 6. What Must Not Be Carried Over Directly

These items should **not** be copied directly:

1. The whole `z-ui x-ui.sh cert` menu tree.
2. `/etc/x-ui/certs/<name>/meta.json` as a forced new storage model for this phase.
3. `discover-import` certificate inventory system as a prerequisite for renewal.
4. `issue-ip` experimental branch.
5. Any historical Cloudflare / DNS API shell path from old scripts.
6. Any design that turns `v-ui` into a multi-provider certificate platform.

## 7. Ceiling Line

This is the hard scope ceiling for `v-ui` certificate work.

### 7.1 Allowed

- domain save
- domain DNS resolve check
- manual certificate upload
- HTTPS enable / disable
- `acme.sh` standalone HTTP-01
- fullchain install
- HTTPS startup self-heal
- z-ui-style renew timer
- renew then restart/reload `x-ui`

### 7.2 Forbidden

- Cloudflare DNS-01
- multi-DNS-provider abstraction
- ZeroSSL UI flow
- certificate asset platform
- multi-certificate repository productization
- tenant certificate system
- automatic DNS record management
- complex certificate observability platform
- new provider credentials storage surface

### 7.3 Operational interpretation

If a future certificate feature needs:

- DNS provider credentials
- TXT automation
- provider switching
- multi-cert inventory UX

then it is already beyond the ceiling and should stop.

## 8. Renewal Implementation Recommendation

## 8.1 Should we directly copy z-ui timer/service unit files?

**Yes, conceptually.**

`systemd` timer/service is the most reusable part.

But the file contents should be adapted to `v-ui` current command surface, not copied byte-for-byte without review.

## 8.2 Should we directly reuse z-ui `x-ui.sh` renew functions?

**No, not wholesale.**

Reason:

- z-ui renew functions assume managed certificate inventory under `/etc/x-ui/certs/<name>/`
- v-ui currently does not use that inventory model as its main cert source
- v-ui currently uses fixed active panel cert paths and Go-side status state

So direct reuse would mix two certificate state models and create drift.

## 8.3 Does v-ui already have a better place to host renewal logic?

**Partially yes.**

Best split:

- shell layer: timer entry and operational command
- Go service layer: current certificate status / active path / validation semantics
- `systemd` layer: scheduling

That means the best minimal design is **mixed**, not pure shell and not pure Go.

## 8.4 Recommended minimal renewal design

Recommended minimal design:

1. Add a narrow renewal command entry in `x-ui.sh`.
2. That command should only handle:
   - invoke renewal
   - call `acme.sh --renew`
   - reinstall fullchain/key to the existing active ACME panel paths
3. Keep active path model unchanged:
   - `/usr/local/x-ui/cert/acme-panel.crt`
   - `/usr/local/x-ui/cert/acme-panel.key`
4. Install:
   - `x-ui-cert-renew.service`
   - `x-ui-cert-renew.timer`
5. On renewal success:
   - restart/reload `x-ui`
   - preserve existing startup self-heal behavior
6. Do not introduce:
   - per-domain inventory
   - provider registry
   - DNS API layer

This is the smallest design that matches both:

- z-ui philosophy
- current v-ui architecture

## 9. Final Decision

### 9.1 Is z-ui / v-ui certificate architecture fully isomorphic?

No.

### 9.2 Is it aligned enough to continue the z-ui-style renewal phase?

Yes.

But only under a strict boundary:

- copy the renewal **pattern**
- not the full `z-ui` certificate manager subsystem

## 10. Next Coding Recommendation

The next coding round should do only this:

1. add `x-ui-cert-renew.service`
2. add `x-ui-cert-renew.timer`
3. add minimal `x-ui.sh` renewal command
4. renew only the current ACME HTTP-01 managed panel certificate
5. on success reinstall fullchain/key and restart/reload `x-ui`
6. do not add any new provider, inventory, or certificate platform concept
