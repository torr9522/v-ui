# Xray 26.5.3 Protocol Parameter Audit

## 1. 当前 v-ui 支持矩阵

说明：本节只统计 v-ui 当前“结构化面板”实际暴露的能力，不把原始 JSON 文本框和 `xrayTemplateConfig` 逃生口混同为“已支持”。

| 类别 | v-ui 当前支持 | 文件位置 | 备注 |
|---|---|---|---|
| 入站协议 | `vmess` `vless` `trojan` `shadowsocks` `dokodemo-door` `socks` `http` `mixed` `tunnel` | `database/model/model.go`, `web/assets/js/model/xray.js`, `web/html/xui/form/inbound.html` | `mtproto` 仅残留在前端模型常量中，当前 UI 未正常暴露。 |
| 出站协议 | `freedom` `blackhole` `socks` `vmess` `vless` `trojan` `shadowsocks` | `web/html/xui/outbounds.html`, `web/service/outbound.go` | 出站编辑主要是协议下拉 + `settings` / `streamSettings` / `mux` 原始 JSON。 |
| 传输层 / network | `tcp` `kcp` `ws` `http` `quic` `grpc` | `web/html/xui/form/stream/stream_settings.html`, `web/assets/js/model/xray.js` | 仍使用旧命名 `tcp` / `http`；未暴露 `raw` / `httpupgrade` / `xhttp`。 |
| TCP/RAW 伪装 | `header.type=none/http`，并暴露 HTTP request/response 伪装字段 | `web/assets/js/model/xray.js`, `web/html/xui/form/stream/stream_tcp.html` | 对应旧 TCP HTTP camouflage。 |
| mKCP/KCP | `mtu` `tti` `uplinkCapacity` `downlinkCapacity` `congestion` `readBufferSize` `writeBufferSize` `header.type` `seed` | `web/assets/js/model/xray.js`, `web/html/xui/form/stream/stream_kcp.html` | 仍完整暴露。 |
| WebSocket | `path` `headers` | `web/assets/js/model/xray.js`, `web/html/xui/form/stream/stream_ws.html` | 仅传统 WS。 |
| HTTP 传输 | `path` `host[]` | `web/assets/js/model/xray.js`, `web/html/xui/form/stream/stream_http.html` | 这是旧 H2 `httpSettings`，不是 XHTTP。 |
| QUIC | `security` `key` `header.type` | `web/assets/js/model/xray.js`, `web/html/xui/form/stream/stream_quic.html` | 仍保留旧 QUIC 表单。 |
| gRPC | `serviceName` | `web/assets/js/model/xray.js`, `web/html/xui/form/stream/stream_grpc.html` | 未见更细粒度 transport 参数。 |
| TLS | `tls` 开关、`serverName`、证书文件路径或证书内容 | `web/html/xui/form/tls_settings.html`, `web/assets/js/model/xray.js` | 没有结构化 `alpn` / `fingerprint` / `minVersion` / `maxVersion`。 |
| XTLS / flow | `xtls` 开关；VLESS flow 仅 `xtls-rprx-vision` | `web/html/xui/form/tls_settings.html`, `web/html/xui/form/protocol/vless.html`, `web/assets/js/model/xray.js` | UI 仍以“xtls 开关”表达旧安全面。 |
| REALITY | 无结构化支持 | 无 | 面板没有 `security=reality`、`realitySettings`、公私钥、shortId 等表单。 |
| VMess | `id` `alterId` `security`，并保留 `disableInsecureEncryption` 模型字段 | `web/html/xui/form/protocol/vmess.html`, `web/assets/js/model/xray.js` | `alterId` 仍暴露。 |
| VLESS | `id` `flow` `fallbacks` | `web/html/xui/form/protocol/vless.html`, `web/assets/js/model/xray.js` | `decryption` 在模型中存在，UI 并未重点暴露。 |
| Trojan | `password` `fallbacks` | `web/html/xui/form/protocol/trojan.html`, `web/assets/js/model/xray.js` | 仅传统字段。 |
| Shadowsocks | 仅 `chacha20-poly1305` `aes-256-gcm` `aes-128-gcm` | `web/html/xui/form/protocol/shadowsocks.html`, `web/assets/js/model/xray.js` | 未暴露 SS 2022 方法。 |
| Sniffing | `enabled` `metadataOnly` `routeOnly` | `web/html/xui/form/sniffing.html`, `web/assets/js/model/xray.js` | 模型内默认 `destOverride=["http","tls"]`，但 UI 不可编辑。 |
| Routing | `field` 规则；`domain` `ip` `port` `protocol` `network` `source` `inboundTag` `outboundTag` | `web/html/xui/routing.html`, `web/service/xray.go`, `database/model/model.go` | 结构化管理不支持更高级 Rule 字段。 |
| DNS / FakeDNS | 无独立结构化页面 | `xray/config.go`, `web/html/xui/setting.html` | 只能靠 `xrayTemplateConfig` 原始模板承载。 |
| Mux | 无独立结构化表单，但出站 JSON 字段可手填 `mux` | `web/html/xui/outbounds.html`, `web/service/outbound.go` | 不是“面板表单支持”，只是原始 JSON 透传。 |
| 原始逃生口 | `xrayTemplateConfig`、出站 `settings` / `streamSettings` / `mux` 文本框 | `web/html/xui/setting.html`, `web/html/xui/outbounds.html` | 这保证了很多 26.5.3 新字段“理论可写”，但并不等于“面板已支持”。 |

## 2. Xray 26.5.3 官方支持矩阵

只采用官方 / 上游依据：

- Xray 26.5.3 二进制：本仓库 `./bin/xray-linux-amd64 version`
- Project X 官方配置文档：
  - Inbound Proxy: <https://xtls.github.io/en/config/inbound.html>
  - Outbound Proxy (Mux, XUDP): <https://xtls.github.io/en/config/outbound.html>
  - Transport Configuration: <https://xtls.github.io/en/config/transport.html>
  - RAW: <https://xtls.github.io/en/config/transports/raw.html>
  - HTTPUpgrade: <https://xtls.github.io/en/config/transports/httpupgrade.html>
  - XHTTP: <https://xtls.github.io/en/config/transports/xhttp.html>
  - TLS: <https://xtls.github.io/en/config/transports/tls.html>
  - REALITY: <https://xtls.github.io/en/config/transports/reality.html>
  - Routing: <https://xtls.github.io/en/config/routing.html>
  - VLESS / VMess / Shadowsocks 对应协议文档

| 类别 | 26.5.3 支持参数 | 官方依据 | 备注 |
|---|---|---|---|
| 入站协议 | `tunnel` `http` `shadowsocks` `socks` `vless` `vmess` `trojan` `wireguard` `hysteria` `tun` | Inbound Proxy 文档的 `protocol` 枚举 | v-ui 当前缺 `wireguard` `hysteria` `tun` 结构化支持。 |
| 出站协议 | `blackhole` `dns` `freedom` `http` `loopback` `shadowsocks` `socks` `trojan` `vless` `vmess` `hysteria` `wireguard` | Outbound Proxy 文档的 `protocol` 枚举 | v-ui 当前缺 `dns` `http` `loopback` `wireguard` `hysteria`。 |
| 传输方法 | `RAW` `XHTTP` `mKCP` `gRPC` `WebSocket` `HTTPUpgrade` `Hysteria` | Transport Configuration 文档 | v-ui 当前结构化表单没有 `raw` / `xhttp` / `httpupgrade` / `hysteria`。 |
| RAW | `network=raw`，`rawSettings.acceptProxyProtocol`，`rawSettings.header` | RAW 文档 | RAW 是“原 TCP transport layer”的新命名。 |
| HTTPUpgrade | `network=httpupgrade`，`path` `host` `headers` `acceptProxyProtocol` | HTTPUpgrade 文档 | 官方还明确建议优先转向 XHTTP。 |
| XHTTP | `network=xhttp` | XHTTP 文档与 Transport Configuration | 官方文档把它列为主线 transport，但专页把细节外链到官方文章。 |
| TLS | `serverName` `verifyPeerCertByName` `rejectUnknownSni` `alpn` `minVersion` `maxVersion` `cipherSuites` `certificates` `disableSystemRoot` `enableSessionResumption` `fingerprint` `pinnedPeerCertSha256` `curvePreferences` `ech*` | TLS 文档 | 官方建议证书链使用 full chain；`allowInsecure` 已标记 deprecated。 |
| REALITY | `security=reality`，`target`/旧名 `dest`，`serverNames`，`privateKey`，`shortIds`，`serverName`，`fingerprint`，`password`（旧名 `publicKey`），`shortId`，`spiderX` | REALITY 文档 | 26.5.3 已包含更新后的字段命名与别名说明。 |
| VLESS flow | 空 flow / `xtls-rprx-vision` | VLESS 官方文档 | 官方把 Vision 放在 VLESS 主文档里，不再需要单独“XTLS 开关式 UX”。 |
| VMess | 官方文档列出 `id` `security` `level` `experiments`，未见 `alterId` | VMess 官方文档 | 说明面板继续暴露 `alterId` 已明显滞后。 |
| Shadowsocks | 官方推荐 `2022-blake3-aes-128-gcm` / `2022-blake3-aes-256-gcm` / `2022-blake3-chacha20-poly1305`，并兼容 `aes-256-gcm` `aes-128-gcm` `chacha20-poly1305` `xchacha20-poly1305` `none` | Shadowsocks 官方文档 | v-ui 当前仅暴露老三项。 |
| Sniffing | `enabled` `destOverride=["http","tls","quic","fakedns"]` `metadataOnly` `domainsExcluded` `ipsExcluded` `routeOnly` | Inbound Proxy / SniffingObject 文档 | v-ui 只露出 `enabled` `metadataOnly` `routeOnly`。 |
| Mux / XUDP | `enabled` `concurrency` `xudpConcurrency` `xudpProxyUDP443` | Outbound Proxy (Mux, XUDP) 文档 | v-ui 无结构化表单。 |
| RoutingObject | `domainStrategy` `rules` `balancers` | Routing 文档 | v-ui 结构化管理只覆盖 `rules` 的一小部分字段。 |
| RuleObject | `domain` `ip` `port` `sourcePort` `localPort` `network` `sourceIP` `localIP` `user` `vlessRoute` `inboundTag` `protocol` `attrs` `process` `outboundTag` `balancerTag` `ruleTag` `webhook` | Routing 文档 | v-ui 当前缺大量字段。 |

## 3. 已淘汰 / 不建议继续暴露的参数

| 参数名 | v-ui 位置 | 26.5.3 状态 | 建议 |
|---|---|---|---|
| `xtls` 开关 | `web/html/xui/form/tls_settings.html`, `web/assets/js/model/xray.js` | 仍有 VLESS Vision / REALITY 相关语义，但官方主线已切到 `tls` / `reality` + `flow=xtls-rprx-vision`，不再适合继续以“独立 xtls 安全面”对用户建模 | 第一轮先保留兼容；UI 标注 `deprecated`，后续隐藏为兼容模式。 |
| `vmess.alterId` | `web/html/xui/form/protocol/vmess.html`, `web/assets/js/model/xray.js` | 当前官方 VMess 文档已不再列出 `alterId` | 不应继续主推；第一轮隐藏或标注 `deprecated`，保留旧配置兼容。 |
| 旧 `tcp` 名称 | `web/html/xui/form/stream/stream_settings.html`, `web/assets/js/model/xray.js` | 官方已把旧 TCP transport layer 改名为 `RAW`，并使用 `network=raw` / `rawSettings` | 第一轮文档标注“旧命名”；后续 UI 改名为 RAW，同时保留旧配置读取兼容。 |
| 旧 TCP HTTP camouflage 表单 | `web/html/xui/form/stream/stream_tcp.html`, `web/assets/js/model/xray.js` | 官方仍兼容 RAW 的 `header.type=http`，但这套玩法已经是旧 transport 时代遗留 | 保留兼容，不建议继续扩展示意；后续放到高级/兼容区。 |
| 旧 `http` 传输表单 | `web/html/xui/form/stream/stream_http.html`, `web/assets/js/model/xray.js` | 当前 transport 主线已出现 `xhttp` / `httpupgrade`；旧 `httpSettings` 不是新主线 | 保留兼容，但不要把它当“现代 HTTP transport”继续扩展。 |
| `quic` 结构化主暴露 | `web/html/xui/form/stream/stream_quic.html` | 核心仍支持，但官方 transport 主线中已不属于最优先 UX | 保留兼容，不作为第一轮补齐重点。 |
| `kcp` / `mkcp` 结构化主暴露 | `web/html/xui/form/stream/stream_kcp.html` | 核心仍支持，但不属于当前推荐主线 | 保留兼容，不新增复杂参数。 |

补充说明：

- TLS 官方文档已把 `allowInsecure` 标成 deprecated，推荐改用 `pinnedPeerCertSha256`。v-ui 结构化表单当前并未暴露该开关，所以这里主要是后续 raw JSON 文档约束，而不是现有 UI 的直接问题。
- VLESS `flow` 仍有效，但“`xtls` 开关 + flow 下拉”的 UI 建模已经落后于当前文档表述。

## 4. 面板缺失的新参数

这里仅列 26.5.3 官方文档明确存在、且值得进入 v-ui 后续结构化表单的缺口。优先级按“实际节点常用度 + 对当前 UX 的影响”评估。

| 参数名 | 26.5.3 支持情况 | v-ui 是否缺失 | 建议优先级 |
|---|---|---|---|
| `security=reality` / `realitySettings` | 官方完整支持 | 缺失 | P0 |
| REALITY `serverNames` `privateKey` `shortIds` | 官方完整支持 | 缺失 | P0 |
| REALITY 客户端面 `fingerprint` `password`(旧名 `publicKey`) `shortId` `spiderX` | 官方完整支持 | 缺失 | P0 |
| `network=xhttp` | 官方 transport 主线 | 缺失 | P0 |
| `network=httpupgrade` | 官方 transport 主线 | 缺失 | P1 |
| TLS `alpn` | 官方完整支持 | 缺失 | P1 |
| TLS `fingerprint` / uTLS | 官方完整支持 | 缺失 | P1 |
| TLS `minVersion` `maxVersion` | 官方完整支持 | 缺失 | P1 |
| Sniffing `destOverride` 可编辑 | 官方完整支持 | 缺失 | P1 |
| Sniffing `domainsExcluded` `ipsExcluded` | 官方完整支持 | 缺失 | P1 |
| Sniffing `fakedns` in `destOverride` | 官方支持 | 缺失 | P1 |
| Mux `xudpConcurrency` `xudpProxyUDP443` | 官方完整支持 | 缺失结构化表单 | P1 |
| 出站协议 `dns` | 官方支持 | 缺失 | P1 |
| 出站协议 `http` | 官方支持 | 缺失 | P1 |
| 出站协议 `loopback` | 官方支持 | 缺失 | P2 |
| 入站 / 出站 `wireguard` | 官方支持 | 缺失 | P2 |
| Shadowsocks 2022 方法 | 官方推荐 | 缺失 | P1 |
| Routing `sourcePort` `localPort` `sourceIP` `localIP` `user` `attrs` `process` `balancerTag` `ruleTag` `webhook` | 官方支持 | 缺失 | P1 |
| Routing `balancers` | 官方支持 | 缺失 | P2 |
| DNS / FakeDNS 结构化编辑 | 官方支持 | 缺失 | P2 |

不列入首轮缺口的项：

- `packetEncoding`：在本轮审阅的 26.5.3 官方配置文档页中未找到对应结构化配置入口，不作为当前面板审计的主结论。
- `domainMatcher`：在当前官方 Routing 文档页中未找到该字段，不应在本轮报告中强行列为 26.5.3 面板缺口。

## 5. 兼容风险

1. 直接删除旧字段会影响已有节点。
   `xtls`、`alterId`、旧 `tcp/http/quic/kcp` 表单已经可能存量存在于数据库中，不能直接清理数据。

2. “核心支持”不等于“面板已支持”。
   v-ui 现在大量依赖 `xrayTemplateConfig`、出站 `settings` / `streamSettings` / `mux` 原始 JSON 文本框承载高级能力。这些能力只能算“透传可写”，不能算结构化支持。

3. 第一轮不应直接改生成逻辑。
   尤其是 `streamSettings`：如果把 `tcp` 强改成 `raw`、把 `xtls` 强改成 `tls/reality`，极易影响现有分享链接、旧节点编辑和老库数据回写。

4. REALITY / XHTTP 需要整套 UX，而不是只补几个字段。
   这两块一旦只做半套字段，很容易生成“看起来能填、实际上不能用”的配置。

5. Routing 不能只看前端。
   当前 `web/service/xray.go` 结构化 routing 组装只覆盖 `field` 规则的少数键。即使前端加字段，服务层不跟进也会丢失。

## 6. 后续改造阶段建议

### Phase A

只加标注和文档，不改生成逻辑。

- 把 `xtls`、`alterId`、旧 `tcp/http/quic/kcp` 标为兼容项或 deprecated。
- 在 README / docs 中明确“结构化支持矩阵”和“原始 JSON 逃生口”的边界。

### Phase B

补齐 P0 参数，不动大面积老字段。

- REALITY 结构化表单
- XHTTP 结构化表单
- VLESS Vision / REALITY 现代化建模

### Phase C

补齐 P1 参数，并开始隐藏 deprecated。

- TLS `alpn` / `fingerprint` / version
- Sniffing 完整字段
- SS 2022 方法
- Mux XUDP 参数
- Routing 高价值字段
- UI 隐藏 `alterId`、`xtls` 等旧入口，但保留读取兼容

### Phase D

做真实节点矩阵测试。

- VLESS + REALITY
- VLESS + XHTTP
- VLESS + TLS
- VMess 兼容旧节点
- Shadowsocks 2022
- Routing / Sniffing / Mux 联合验证

## 结论

当前 v-ui 的结构化协议表单，整体仍停留在“旧 TCP/WS/HTTP/QUIC/gRPC + TLS/XTLS”这一代；而 Xray 26.5.3 的官方主线已经明显转向：

- `RAW` 取代旧 `tcp` transport 命名
- `XHTTP` / `HTTPUpgrade` 成为更现代的 HTTP 类 transport
- `REALITY` 成为必须补齐的安全面
- `Sniffing`、`Mux`、`Routing` 的结构化字段都比 v-ui 现在完整得多

推荐第一轮改造只做两件事：

1. 先标注和冻结旧字段，不要马上删除兼容项。
2. 只补 P0 主线能力：`REALITY` 和 `XHTTP`，不要同时扩展成“全协议大翻修”。
