# Inbound Protocol Compatibility Audit

## Scope

审计目标：

- 当前 v-ui 入站协议 UI 与模型层是否一致
- 新增 / 编辑 / 保存 / 回显链路是否存在明显回归
- TLS / XTLS、传输层、证书自动选择、分享链接生成是否自洽
- 与固定内核 `Xray-core 26.5.3` 的兼容边界

本轮结论基于：

- 代码静态审计：
  `web/assets/js/model/xray.js`
  `web/assets/js/model/models.js`
  `web/html/xui/form/*`
  `web/html/xui/inbound_modal.html`
  `web/html/xui/inbounds.html`
- 前序真人 UAT：
  `VLESS + RAW/TLS`
  `VMess + WS + TLS`
  `Trojan + RAW + TLS`
- 前序协议参数审计：
  [XRAY_26_5_3_PROTOCOL_AUDIT.md](/tmp/v-ui-init/v-ui/docs/XRAY_26_5_3_PROTOCOL_AUDIT.md)

说明：

- 本文区分“真人已验证”和“静态推断”。
- 本文不把 `REALITY` / `XHTTP` 缺失当作本轮回归；这属于现代化缺口，而不是 Modern UI Phase 1 的新增回归。

## Summary

- P0：未发现
- P1：发现 3 项，主要集中在分享链接生成与现代 TLS/XTLS 建模
- P2：发现若干后续协议能力缺口
- P3：需要补文档说明的行为边界存在

当前建议：

- 可以推送远程仓库
- 推送前最好修 P1，尤其是 `Trojan` / `Shadowsocks` 的分享链接生成

## Support Matrix

| 协议 / 能力 | 新增/编辑 UI | 传输层 UI | TLS/XTLS UI | 分享链接 | 编辑回显 | 26.5.3 兼容评估 | 验证方式 |
|---|---|---|---|---|---|---|---|
| VMess | 有 | `RAW/WS/gRPC/HTTP/KCP/QUIC` | TLS，XTLS 不适用 | 有 | 有 | 基本兼容；`alterId` 已过时但保留兼容 | 真人 + 静态 |
| VLESS | 有 | `RAW/WS/gRPC/HTTP/KCP/QUIC` | TLS/XTLS | 有 | 有 | 基本兼容；仍沿用旧 `XTLS` 开关建模 | 真人 + 静态 |
| Trojan | 有 | 已恢复 `RAW/WS/gRPC/HTTP/KCP/QUIC` UI | TLS/XTLS | 有，但不完整 | 有 | 基本可用；分享链接对现代 transport/TLS 参数表达不足 | 真人 + 静态 |
| Shadowsocks | 有 | `RAW/WS/gRPC/HTTP/KCP/QUIC` | TLS 可开 | 有，但不完整 | 结构上可回显 | 基础兼容；方法集停留在旧三项 | 静态 |
| SOCKS | 有 | 无 stream UI | 无 TLS UI | 无 | 结构上可回显 | 基本兼容 | 静态 |
| HTTP | 有 | 无 stream UI | 无 TLS UI | 无 | 结构上可回显 | 基本兼容 | 静态 |
| Dokodemo-door | 有 | 无 stream UI | 无 TLS UI | 无 | 结构上可回显 | 基本兼容 | 静态 |
| Mixed | 有 | 无 stream UI | 无 TLS UI | 无 | 结构上可回显 | 基本兼容 | 静态 |
| Tunnel | 有 | 无 stream UI | 无 TLS UI | 无 | 结构上可回显 | 基本兼容 | 静态 |
| TLS 证书自动选择 | 面向可启用 TLS 的入站 | 仅文件路径模式自动填 | 支持面板证书复用 | 不直接影响 | 已验证 | 对 `Manual` / `ACME HTTP-01` 面板证书链路有效 | 真人 |
| RAW / WS / gRPC | 已结构化 | 是 | 按协议条件显示 | VMess/VLESS 基本完整 | 结构上可回显 | 可用 | 真人 + 静态 |
| HTTP / KCP / QUIC | 仍暴露为 `Legacy / Compatibility` | 是 | 受协议条件限制 | 仅 VMess/VLESS 相对完整 | 结构上可回显 | 兼容但非现代主线 | 静态 |

## Verified vs Inferred

真人已验证：

- `VLESS + RAW + TLS`
  新增、TLS 自动证书、保存、分享链接、删除通过
- `VMess + WS + TLS`
  新增、保存、编辑回显、删除通过
- `Trojan + RAW + TLS`
  新增、TLS 自动证书、保存、编辑回显、删除通过

静态推断为可用，但未做人机链路验证：

- `Shadowsocks`
- `SOCKS`
- `HTTP`
- `Dokodemo-door`
- `Mixed`
- `Tunnel`
- `HTTP/KCP/QUIC` 旧 transport 的真人链路

## Findings

### P0

无。

说明：

- 本轮没有再发现“新增入站直接缺协议”“保存即崩”“编辑回显丢失基础字段”这类阻断推送的问题。
- `Trojan` 下拉缺失与 `Trojan` 的 `Transport/Security` tab 缺失，已由 `15a9fa6` 与 `62b60eb` 修复并经远端 UAT 通过。

### P1

#### P1-1 Trojan 分享链接生成不完整

影响：

- 当前 `Trojan` 链接生成只输出 `trojan://password@address:port#remark`
- 未编码 `type`、`security`、`sni`、`host`、`path`、`serviceName` 等 transport/TLS 参数
- 当用户把 `Trojan` 配成 `WS/gRPC/HTTP/QUIC`，或面板访问地址与 SNI 域名不一致时，复制链接很容易不够用

证据：

- `genTrojanLink()` 仅返回最简 URI
  [xray.js](/tmp/v-ui-init/v-ui/web/assets/js/model/xray.js:1033)
- `Trojan` 现在已经能在 UI 中选择 transport 与 TLS
  [inbound.html](/tmp/v-ui-init/v-ui/web/html/xui/form/inbound.html:109)
  [tls_settings.html](/tmp/v-ui-init/v-ui/web/html/xui/form/tls_settings.html:1)

建议：

- 推前修更稳妥
- 最小修法是把 `Trojan` 链接生成补齐到至少覆盖：
  `security` `sni` `type` `host` `path` `serviceName`

#### P1-2 Shadowsocks 分享链接与 UI 可配置能力不一致

影响：

- UI 允许 `Shadowsocks` 选择 stream/TLS
- 但 `genSSLink()` 只输出 `method:password@address:port`
- 未表达 stream、TLS、插件或其它补充参数

结果：

- 纯基础 `SS` 可能可用
- 一旦使用 `TLS`、`WS`、`gRPC`、旧 `HTTP/KCP/QUIC` 等组合，复制出来的分享链接大概率不能完整复原入站配置

证据：

- `Shadowsocks` 生成链接仅使用 method/password/address/port
  [xray.js](/tmp/v-ui-init/v-ui/web/assets/js/model/xray.js:1023)
- `Shadowsocks` 可启用 stream/TLS
  [xray.js](/tmp/v-ui-init/v-ui/web/assets/js/model/xray.js:812)
  [inbound.html](/tmp/v-ui-init/v-ui/web/html/xui/form/inbound.html:109)

建议：

- 推前修更稳妥
- 若短期不修，至少应在文档中明确：
  `SS` 分享链接只保证基础 TCP/UDP 场景，不保证带 TLS/自定义 stream 的复杂入站

#### P1-3 VLESS/Trojan 的现代安全面仍以旧 XTLS 开关建模

影响：

- 当前 UI 仍把 `XTLS` 当作独立安全面开关
- `VLESS flow=xtls-rprx-vision` 只有在 `XTLS` 打开时才可见
- 这与 Xray 26.5.3 的现代文档表达已经不完全一致，容易让用户把“Vision/现代 TLS”理解成旧 XTLS 模式

证据：

- `XTLS (Deprecated)` 仍是独立开关
  [tls_settings.html](/tmp/v-ui-init/v-ui/web/html/xui/form/tls_settings.html:5)
- `flow` 下拉受 `inbound.xtls` 控制
  [vless.html](/tmp/v-ui-init/v-ui/web/html/xui/form/protocol/vless.html:6)

建议：

- 这不是当前推送阻断项
- 但若准备公开推送并承诺“面向 Xray 26.5.3 的现代 UI”，建议推前修或至少加文档说明

### P2

#### P2-1 缺少 REALITY / XHTTP / HTTPUpgrade 结构化支持

状态：

- 属于当前固定 scope 之外
- 但从 `Xray 26.5.3` 视角看，这是现代主线能力缺口

建议：

- 作为后续 issue

#### P2-2 Shadowsocks 方法集仍停留在旧三项

当前仅暴露：

- `chacha20-poly1305`
- `aes-256-gcm`
- `aes-128-gcm`

缺失：

- `2022-*`
- `xchacha20-poly1305`
- `none`

证据：

- [xray.js](/tmp/v-ui-init/v-ui/web/assets/js/model/xray.js:21)

#### P2-3 多 client / 多 account 能力未结构化暴露

现状：

- 模型层大量协议仍以数组承载
- UI 实际只编辑第一个 client/account

影响协议：

- `VMess`
- `VLESS`
- `Trojan`
- `SOCKS`
- `HTTP`
- `Mixed`

建议：

- 不阻断推送
- 后续是否支持多 client，应单独做需求决策

#### P2-4 旧 transport 仍保留，但仅适合作为兼容模式

涉及：

- `HTTP (Legacy)`
- `mKCP (Legacy)`
- `QUIC (Legacy)`

现状：

- UI 已有 `Legacy / Compatibility` 标记
- 这部分与 `Xray 26.5.3` 现代主线并不一致，但仍有兼容价值

### P3

#### P3-1 需要文档明确“哪些协议有分享链接，哪些没有”

当前只有以下协议会出现二维码/分享链接：

- `VMess`
- `VLESS`
- `Trojan`
- `Shadowsocks`

证据：

- `hasLink()` 只返回上述四类
  [models.js](/tmp/v-ui-init/v-ui/web/assets/js/model/models.js:187)

建议：

- README 或入站文档中写清楚这是设计行为，不是缺陷

#### P3-2 需要文档明确“真人已验证范围”

已真人验证：

- `VLESS + RAW + TLS`
- `VMess + WS + TLS`
- `Trojan + RAW + TLS`

未真人验证但结构化存在：

- `Shadowsocks`
- `SOCKS`
- `HTTP`
- `Dokodemo-door`
- `Mixed`
- `Tunnel`

## Push Recommendation

当前判断：

- 可以推送远程仓库
- 不存在必须挡住推送的 P0

更稳妥的推前修复项：

1. `Trojan` 分享链接补齐 transport/TLS/SNI 参数
2. `Shadowsocks` 分享链接与 UI 可配置能力对齐
3. 若准备强调“面向 Xray 26.5.3 的现代安全面”，补一版 `XTLS/Vision` 文案或建模说明

若本轮选择直接推送，建议至少同步文档说明：

- `Trojan` / `SS` 复杂 transport/TLS 组合的分享链接仍存在边界
- 旧 `HTTP/KCP/QUIC` 为兼容模式，不代表现代推荐主线
- `REALITY` / `XHTTP` 暂未进入当前稳定 scope

## Final Verdict

- P0：无
- P1：有，主要是分享链接与现代 TLS/XTLS 建模
- 结论：`可推送，但建议先修 P1-1 与 P1-2`
