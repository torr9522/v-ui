# Z-UI Certificate Audit For V-UI

## 1. Audit Scope

This round is read-only analysis only.

- `c-ui` source baseline:
  - repo: `https://github.com/torr9522/c-ui`
  - branch: `c-ui`
  - commit: `f4076be4ec0d2fbce0ea2d722eb4da13e83adaa3`
- `z-ui` reference source:
  - repo: `https://github.com/torr9522/z-ui`
  - branch: `z-ui`
  - commit: `29cd5ffe251bbb7eeb34df143cf5299ee44324be`
- local isolated project:
  - path: `/tmp/v-ui-init/v-ui`
  - branch: `v-ui`
  - remote: none

Goal: design how `v-ui` should absorb domain / certificate / HTTPS panel capabilities by borrowing ideas from `z-ui`, without touching `c-ui` or `z-ui`.

## 2. Z-UI Related Files

### 2.1 Shell / install side

- `x-ui.sh`
  - real certificate manager entry
  - ACME issuance
  - renewal timer install
  - panel HTTPS switch
  - import / delete / status
- `install.sh`
  - installs runtime and supporting shell assets
- `knowledge-base/07-certificate-system.md`
  - high-level feature statement

### 2.2 Web / API side

- `web/controller/certificate.go`
  - `/api/certificates`
  - list
  - discover
  - import-discovered
- `web/service/certificate.go`
  - managed certificate directory scan
  - external certificate discovery
  - metadata load/import
- `web/entity/certificate.go`
  - web API certificate DTOs
- `web/web.go`
  - panel HTTPS startup via `webCertFile` / `webKeyFile`
- `web/service/setting.go`
- `web/entity/entity.go`
  - settings only expose panel cert path fields

### 2.3 Web pages using certificate data

- `web/html/xui/setting.html`
  - panel cert path text inputs only
- `web/html/xui/inbound_modal.html`
- `web/html/xui/form/tls_settings.html`
  - inbound TLS selector integrates discovered/imported certificates

## 3. V-UI Current Related Files

### 3.1 Existing cert / HTTPS structure

- `web/service/setting.go`
  - settings include `webCertFile`, `webKeyFile`
- `web/entity/entity.go`
  - validates cert pair with `tls.LoadX509KeyPair`
- `web/web.go`
  - starts HTTPS listener if both cert path settings are set
- `web/html/xui/setting.html`
  - current UI only allows manual file path input
- `web/assets/js/model/models.js`
  - `AllSetting` already models panel cert path settings
- `web/controller/setting.go`
  - single settings save endpoint
- `web/controller/xui.go`
  - current page routing, no cert page
- `web/html/xui/common_sider.html`
  - no certificate menu entry

### 3.2 Shell / helper side

- `x-ui.sh`
  - management script, but no full z-ui style cert manager
- `install.sh`
  - source bootstrap and runtime install only
- `scripts/acme_install.sh`
  - raw acme installer helper, not integrated with panel domain workflows
- `web/network/autp_https_conn.go`
  - auto-HTTPS redirect helper exists, but not domain/cert management

## 4. What Z-UI Actually Implements

## 4.1 Domain input and persistence

- Domain is entered interactively inside `x-ui.sh` during ACME issuance.
- z-ui does not currently have a dedicated persistent `webDomain` setting in the web panel.
- Issued/imported certificate metadata is persisted in `/etc/x-ui/certs/<name>/meta.json`.
- Active panel HTTPS state is persisted only by writing:
  - `webCertFile`
  - `webKeyFile`
  into the settings table.

Implication: z-ui's domain workflow is shell-first, not panel-first.

## 4.2 Certificate issuance

### Implemented

- ACME client: `acme.sh`
- CA selection:
  - explicitly sets default CA to Let's Encrypt
- Domain issuance:
  - HTTP-01 standalone
- IP issuance:
  - experimental path in shell

### Observed constraints

- domain must resolve to current server public IPv4
- port `80` must be free
- ACME is installed on demand by shell logic
- issuance is not exposed as web API in current z-ui

### Not found in the audited z-ui path

- first-class Cloudflare DNS-01 flow
- explicit ZeroSSL issuance flow in cert manager
- panel-side certificate issuance UI
- certificate CRUD web APIs beyond list/discover/import-discovered

## 4.3 Certificate storage

Managed certificates live under:

`/etc/x-ui/certs/<name>/`

Expected files:

- `fullchain.pem`
- `privkey.pem`
- `meta.json`

Permissions observed:

- cert dir: `0700`
- private key: `0600`
- fullchain: `0644`
- meta: `0600`

## 4.4 Panel HTTPS switching

- z-ui writes `webCertFile` and `webKeyFile` into DB settings
- `web/web.go` loads those paths and wraps the panel listener with TLS
- switching can trigger reload/restart of `x-ui`

## 4.5 Renewal

- shell installs:
  - `x-ui-cert-renew.service`
  - `x-ui-cert-renew.timer`
- timer schedule:
  - daily `03:00:00`
- renewal only applies to ACME-managed certificates
- imported certificates are excluded from ACME auto-renew
- renewal reloads/restarts panel if active cert changed

## 4.6 Discovery / import

z-ui can discover existing certificates from:

- `/etc/letsencrypt/live`
- `/root/.acme.sh`
- `/home/*/.acme.sh`

The web API can list and import discovered certificate pairs into the managed directory.

## 4.7 Security characteristics

Positive points:

- API does not return private key contents
- private key stored with `0600`
- managed metadata separated in `meta.json`
- deleted certificates are backed up instead of immediate hard delete

Weak points / design debt:

- issuance and renewal are shell-centric
- panel has no unified certificate state page
- DNS provider credentials flow is absent in audited path
- active cert paths in DB can drift from filesystem reality

## 5. Z-UI Certificate Logic Flow

```text
User / operator
  |
  v
x-ui.sh cert ...
  |
  +--> ensure acme.sh installed
  |
  +--> validate domain A record == server public IP
  |
  +--> validate port 80 free
  |
  +--> acme.sh --issue --standalone
  |
  +--> acme.sh --install-cert
  |
  +--> write /etc/x-ui/certs/<name>/{fullchain.pem,privkey.pem,meta.json}
  |
  +--> optional: write webCertFile/webKeyFile to DB settings
  |
  +--> restart/reload x-ui if cert in use
  |
  +--> install x-ui-cert-renew.timer
```

Separate read-only web path:

```text
Web UI / API
  |
  +--> /api/certificates
  |      list managed certs
  |
  +--> /api/certificates/discover
  |      scan letsencrypt / acme.sh roots
  |
  +--> /api/certificates/import-discovered
         copy cert pair into /etc/x-ui/certs/<name>/
```

## 6. V-UI Current Capability And Gaps

## 6.1 Already present in v-ui

- Panel can already run HTTPS via `webCertFile` + `webKeyFile`
- Settings infrastructure already supports storing cert paths
- Validation already rejects invalid cert/key pairs
- Existing panel setting save/restart flow can be reused
- Menu/page/controller architecture is already clear and extendable

## 6.2 Missing in v-ui

- no `域名证书` page
- no certificate status API
- no managed certificate directory service
- no issue/import/delete/renew controller
- no domain setting
- no DNS challenge provider abstraction
- no renewal timer management
- no operation log surface
- no safe rollback for HTTPS enable failures

## 6.3 Direct carry-over points

- reuse z-ui storage layout:
  - `/etc/x-ui/certs/<name>/`
  - `meta.json`
- reuse z-ui discovery/import logic shape
- reuse current `web/web.go` HTTPS startup
- reuse current settings infrastructure for active panel cert pointers

## 6.4 Parts that should not be copied blindly

- shell-only interactive prompts
- DB writes done directly via sqlite shell command
- coupling all certificate behavior into `x-ui.sh`
- HTTP-01 only design as the final architecture
- no explicit rollback contract for failed HTTPS switch

## 7. Comparison Table

| 模块 | z-ui 实现 | v-ui 当前实现 | 是否可复用 | 风险 |
| -- | -- | -- | -- | -- |
| 域名输入 | shell prompt in `x-ui.sh` | none | partial | interactive shell flow cannot be reused directly in web UI |
| 域名保存 | no dedicated panel setting; metadata in `meta.json` | none | partial | domain state may be fragmented |
| 证书申请 | `acme.sh` standalone HTTP-01 in shell | none | concept only | shell-first implementation unsuitable for panel API as-is |
| 证书续期 | systemd timer + `acme.sh --renew` | none | yes | renewal hook must avoid invalid restart loops |
| 面板 HTTPS | write `webCertFile` / `webKeyFile`, then reload/restart | same path-based startup logic already exists | yes | bad paths can lock out panel if no rollback |
| 证书路径 | `/etc/x-ui/certs/<name>/` | manual absolute paths only | yes | need consistent ownership and permissions |
| Cloudflare DNS | not found in audited cert path | none | no direct reuse | must be designed fresh |
| 80/443 占用处理 | checks port 80 before standalone issue | none | yes | HTTP-01 may conflict with running web stack |
| 服务重启 | reload/restart from shell | panel restart endpoint exists | partial | restart timing and rollback need controller-safe contract |
| UI 页面 | no dedicated cert page | no dedicated cert page | no | must be built fresh |
| API 接口 | list/discover/import only | none | partial | issuance/renew/delete APIs need new design |
| 安全 | key perms, no private key API output | cert pair validation only | partial | DNS token storage and log redaction still missing |

## 8. Recommended V-UI Design

## 8.1 Module name

- English: `Domain & Certificate`
- Chinese: `域名证书`

Recommendation: use `域名证书` as left menu label and page title.

## 8.2 Design principles

- keep runtime compatibility with `x-ui`
- make certificate management panel-first, not shell-first
- use filesystem as certificate material source of truth
- use settings as active panel cert source of truth
- add rollback points around panel HTTPS enable / issue / renew
- never expose private key contents through web API

## 8.3 Storage model

Recommended managed directory:

`/etc/x-ui/certs/<name>/`

Recommended contents:

- `fullchain.pem`
- `privkey.pem`
- `meta.json`
- optional future:
  - `issue.log`
  - `renew.log`
  - `backup/`

Recommended `meta.json` fields:

- `name`
- `domain`
- `provider`
- `issuer`
- `type`
  - `manual`
  - `http01`
  - `dns01`
  - `imported`
- `certFile`
- `keyFile`
- `createdAt`
- `expireAt`
- `lastRenewAt`
- `autoRenew`
- `dnsProvider`
- `status`
- `lastError`

## 8.4 Functional scope

### Phase-ready scope

1. panel domain setting
2. certificate inventory and status
3. manual upload/import
4. enable HTTPS / disable HTTPS
5. HTTP-01 issuance
6. renewal timer and logs

### Later extension

1. Cloudflare DNS-01
2. staging mode
3. multi-provider CA support
4. operation audit history
5. certificate reuse for inbound TLS selector

## 9. Setting And DB Design

## 9.1 Primary recommendation

For first implementation, do not add a new mandatory certificate table.

Reason:

- certificate files already live on filesystem
- renewal and issuance are file-oriented operations
- active panel HTTPS state is already settings-based
- duplicating certificate inventory into DB creates drift risk
- current scope is panel domain/cert management, not full PKI inventory

Use:

- settings for panel-active state and global preferences
- filesystem `meta.json` for per-certificate metadata

## 9.2 Recommended setting keys

- `webDomain`
- `webCertMode`
  - `http`
  - `https`
- `webCertProvider`
  - `manual`
  - `letsencrypt-http`
  - `letsencrypt-dns`
- `webCertEmail`
- `webCertFile`
- `webKeyFile`
- `webCertStatus`
- `webCertExpireAt`
- `webCertIssuer`
- `webCertAutoRenew`
- `webCertDnsProvider`
- `webCertDnsConfigEncrypted`
- `webCertLastError`
- `webCertLastRenewAt`
- `webCertStaging`

## 9.3 Optional future table design

If later phases need searchable history, multi-cert dashboard performance, or audit retention, an optional `certificates` table can be added.

Suggested shape:

| 字段 | 类型 | 说明 |
| -- | -- | -- |
| `id` | int | 主键 |
| `name` | varchar(128) | managed cert name |
| `domain` | varchar(255) | 主域名 |
| `provider` | varchar(64) | manual / letsencrypt-http / letsencrypt-dns |
| `issuer` | varchar(255) | issuer text |
| `type` | varchar(32) | imported / http01 / dns01 / manual |
| `cert_file` | varchar(512) | fullchain path |
| `key_file` | varchar(512) | private key path |
| `dns_provider` | varchar(64) | cf etc |
| `auto_renew` | bool | renewal switch |
| `active_for_panel` | bool | current panel cert |
| `status` | varchar(32) | ready / issuing / error / expired |
| `expire_at` | bigint | unix ms or sec, choose one consistently |
| `last_renew_at` | bigint | renew timestamp |
| `last_error` | text | last failure summary |
| `created_at` | bigint | create timestamp |
| `updated_at` | bigint | update timestamp |

Recommendation: keep this optional until audit/history requirements justify it.

## 10. API Design

All APIs should remain under `/xui` to preserve runtime route freeze.

## 10.1 `POST /xui/cert/status`

Request:

```json
{}
```

Response:

```json
{
  "success": true,
  "msg": "get certificate status success",
  "obj": {
    "httpsEnabled": true,
    "webDomain": "panel.example.com",
    "activeCert": {
      "name": "panel.example.com",
      "domain": "panel.example.com",
      "issuer": "Let's Encrypt",
      "expireAt": 1760000000
    },
    "certificates": []
  }
}
```

- Restart x-ui: no
- Rollback: no
- Risks:
  - active cert path may point to missing file
  - status must handle drift between settings and filesystem

## 10.2 `POST /xui/cert/setDomain`

Request:

```json
{
  "domain": "panel.example.com"
}
```

Response:

```json
{
  "success": true,
  "msg": "set domain success",
  "obj": {
    "domain": "panel.example.com"
  }
}
```

- Restart x-ui: no
- Rollback: no
- Risks:
  - invalid FQDN
  - saved domain not yet resolved to current server

## 10.3 `POST /xui/cert/checkDomain`

Request:

```json
{
  "domain": "panel.example.com"
}
```

Response:

```json
{
  "success": true,
  "msg": "check domain success",
  "obj": {
    "domain": "panel.example.com",
    "publicIp": "1.2.3.4",
    "resolvedIps": ["1.2.3.4"],
    "matched": true,
    "port80Free": true
  }
}
```

- Restart x-ui: no
- Rollback: no
- Risks:
  - outbound IP lookup may fail
  - multi-A record deployments need explicit UX language

## 10.4 `POST /xui/cert/issueHttp`

Request:

```json
{
  "domain": "panel.example.com",
  "email": "admin@example.com",
  "staging": false,
  "enableHttps": true,
  "autoRenew": true
}
```

Response:

```json
{
  "success": true,
  "msg": "issue http certificate success",
  "obj": {
    "name": "panel.example.com",
    "domain": "panel.example.com",
    "applied": true,
    "httpsEnabled": true
  }
}
```

- Restart x-ui: yes if `enableHttps=true`
- Rollback: yes
- Risks:
  - port 80 occupied
  - domain not resolved
  - ACME issue succeeds but HTTPS switch fails
  - failed switch must restore previous `webCertFile` / `webKeyFile`

## 10.5 `POST /xui/cert/issueDns`

Request:

```json
{
  "domain": "panel.example.com",
  "email": "admin@example.com",
  "provider": "cloudflare",
  "apiToken": "secret",
  "zoneApiToken": "",
  "staging": false,
  "enableHttps": true,
  "autoRenew": true
}
```

Response:

```json
{
  "success": true,
  "msg": "issue dns certificate success",
  "obj": {
    "name": "panel.example.com",
    "provider": "cloudflare",
    "applied": true
  }
}
```

- Restart x-ui: yes if enabling HTTPS
- Rollback: yes
- Risks:
  - DNS token leakage
  - provider API misconfiguration
  - renew hook must not require exposing raw token again

## 10.6 `POST /xui/cert/upload`

Request:

```json
{
  "name": "panel-manual",
  "domain": "panel.example.com",
  "certPem": "-----BEGIN CERTIFICATE-----...",
  "keyPem": "-----BEGIN PRIVATE KEY-----...",
  "enableHttps": true
}
```

Response:

```json
{
  "success": true,
  "msg": "upload certificate success",
  "obj": {
    "name": "panel-manual",
    "imported": true,
    "applied": true
  }
}
```

- Restart x-ui: yes if enabling HTTPS
- Rollback: yes
- Risks:
  - invalid cert/key pair
  - storing PEM in logs must be forbidden

## 10.7 `POST /xui/cert/delete`

Request:

```json
{
  "name": "panel.example.com"
}
```

Response:

```json
{
  "success": true,
  "msg": "delete certificate success",
  "obj": {
    "deleted": true,
    "backupPath": "/etc/x-ui/certs/deleted.panel.example.com.20260630000000"
  }
}
```

- Restart x-ui: only if deleting active panel cert and switching HTTPS off or to another cert
- Rollback: yes when active cert involved
- Risks:
  - deleting active cert can lock panel out
  - prefer soft-delete backup like z-ui

## 10.8 `POST /xui/cert/renew`

Request:

```json
{
  "name": "panel.example.com"
}
```

Response:

```json
{
  "success": true,
  "msg": "renew certificate success",
  "obj": {
    "name": "panel.example.com",
    "renewed": true,
    "applied": true
  }
}
```

- Restart x-ui: maybe
- Rollback: yes if active cert replacement fails
- Risks:
  - imported/manual cert must be blocked from ACME renew
  - renewal should be idempotent

## 10.9 `POST /xui/cert/enableHttps`

Request:

```json
{
  "name": "panel.example.com"
}
```

Response:

```json
{
  "success": true,
  "msg": "enable https success",
  "obj": {
    "applied": true,
    "httpsEnabled": true
  }
}
```

- Restart x-ui: yes
- Rollback: yes
- Risks:
  - restart succeeds but TLS listener load fails
  - must validate cert pair before writing settings

## 10.10 `POST /xui/cert/disableHttps`

Request:

```json
{}
```

Response:

```json
{
  "success": true,
  "msg": "disable https success",
  "obj": {
    "applied": true,
    "httpsEnabled": false
  }
}
```

- Restart x-ui: yes
- Rollback: yes
- Risks:
  - restart failure after clearing cert paths

## 10.11 `POST /xui/cert/logs`

Request:

```json
{
  "name": "panel.example.com",
  "limit": 200
}
```

Response:

```json
{
  "success": true,
  "msg": "get cert logs success",
  "obj": {
    "lines": [
      "2026-06-30T00:00:00Z renew success"
    ]
  }
}
```

- Restart x-ui: no
- Rollback: no
- Risks:
  - log output must redact secrets and token values

## 11. Frontend Page Design

New page:

- `web/html/xui/cert.html`

New left menu entry:

- `域名证书`

Recommended page sections:

1. 当前状态
   - HTTP / HTTPS
   - active cert
   - expire time
   - issuer
   - auto renew state
2. 域名设置
   - domain input
   - domain check button
   - clear domain button
3. 证书申请
   - HTTP-01 issue form
   - email
   - staging
   - enable HTTPS after issue
4. Cloudflare DNS 配置
   - provider selector
   - token input
   - save masked state
5. 手动上传证书
   - PEM textarea or file upload
   - optional apply immediately
6. 证书列表
   - managed certs
   - active badge
   - expireAt
   - issuer
   - operations:
     - enable HTTPS
     - renew
     - delete
     - view details
7. 续期状态
   - next renew time
   - timer status
   - last renew result
8. 操作日志
   - issue / renew / apply results

UI style:

- keep Vue2 + Ant Design Vue
- follow current `setting.html`, `outbounds.html`, `routing.html` style
- do not add new frontend dependencies

## 12. Implementation Notes For V-UI

## 12.1 Recommended service split

- `web/service/certificate.go`
  - certificate inventory
  - filesystem storage
  - issue/import/delete/renew
  - domain checks
- `web/controller/certificate.go`
  - `/xui/cert/*`
- `web/entity/entity.go`
  - extend `AllSetting`
  - add request/response DTOs if needed
- `web/service/setting.go`
  - new setting getters/setters
- `web/controller/xui.go`
  - add cert page route
- `web/html/xui/common_sider.html`
  - add menu entry
- `web/html/xui/cert.html`
  - dedicated UI
- `x-ui.sh`
  - keep CLI compatibility layer later if needed, but not as primary implementation target

## 12.2 HTTPS apply strategy

Recommended safe sequence:

1. validate cert pair
2. persist cert files and metadata
3. snapshot current `webCertFile` / `webKeyFile`
4. write new settings
5. restart panel
6. verify panel process active
7. on failure:
   - restore previous settings
   - restart again
   - return failure

## 12.3 Renewal strategy

Recommendation:

- continue using `acme.sh`
- use systemd timer or cron only after first successful managed issuance
- store renew command/provider mode in metadata
- restart panel only when active cert actually changed

## 13. Security Risk List

1. Cloudflare or other DNS provider credentials are the highest-risk new input.
2. Panel HTTPS switch can lock operators out if restart succeeds with broken cert paths.
3. HTTP-01 may fail when port 80 is occupied by another service.
4. Logs must never print raw PEM, API tokens, or full secret config.
5. Manual upload must validate cert/key pair before persisting.
6. Renewal hooks must not create restart loops.
7. Discover/import must restrict allowed filesystem roots to avoid arbitrary file read.
8. Domain check must handle CDN or multi-A record cases without false promises.
9. Cert deletion must be soft-delete or backup-first when active cert is involved.

## 14. Phased Commit Plan

### Phase 0

`v-ui` project isolation and baseline commit

### Phase 1

`z-ui` certificate logic audit report

### Phase 2

settings / model / service / controller skeleton

### Phase 3

domain check + certificate status read APIs and page skeleton

### Phase 4

manual upload certificate + enable HTTPS + rollback

### Phase 5

HTTP-01 automated issuance

### Phase 6

Cloudflare DNS-01

### Phase 7

auto renew + renew logs + timer management

### Phase 8

real machine verification matrix

## 15. First Coding Recommendation

Do not begin with ACME issuance.

Recommended first coding slice:

1. add setting keys for domain and cert status
2. add certificate service skeleton with managed directory scan
3. add `POST /xui/cert/status`
4. add `POST /xui/cert/setDomain`
5. add `POST /xui/cert/checkDomain`
6. add `web/html/xui/cert.html` read-only status page
7. add left menu entry

Reason:

- lowest operational risk
- validates page/service/controller wiring first
- reuses current HTTPS startup path
- avoids introducing restart-sensitive issuance logic too early
