# Xray 26.5.3 Parameter Mapping

## Scope

本文件只建立三层映射：

- 当前 UI
- 当前 JSON
- Xray 26.5.3 官方名称 / 官方主线概念

本轮不修改：

- 数据库
- API
- Xray config 生成
- 运行逻辑

## Sources

- 仓库内置内核：`./bin/xray-linux-amd64 version`
  - `Xray 26.5.3`
- Project X / Xray 官方文档
  - Inbound
  - Outbound
  - Transport
  - TLS
  - REALITY
  - Routing
  - VMess / VLESS / Trojan / Shadowsocks

## Status Legend

- `KEEP`
  - 当前字段应继续保留
- `DEPRECATED`
  - 当前字段仍需兼容，但应标 Deprecated / Compatibility
- `REMOVE`
  - 后续应从主 UI 移除，不代表马上删除数据库数据
- `NEW`
  - 26.5.3 官方主线存在，但当前 UI 缺失

## 1. Protocol Mapping

### 1.1 VMess

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `id` | `settings.clients[0].id` | `clients[].id` | KEEP | 核心客户端标识。 |
| `额外 ID (Deprecated)` | `settings.clients[0].alterId` | `alterId` | DEPRECATED | 官方主线已不再推荐；仍需兼容旧节点。 |
| `禁用不安全加密` | `settings.disableInsecureEncryption` | `disableInsecureEncryption` | KEEP | 当前 JSON 名称与现有配置一致。 |
| 无 | 无结构化 UI | `clients[].security` / `experiments` 等 | NEW | 当前 UI 未结构化暴露。 |

### 1.2 VLESS

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `id` | `settings.clients[0].id` | `clients[].id` | KEEP | 核心客户端标识。 |
| `flow`（仅 XTLS 开关下显示） | `settings.clients[0].flow` | `clients[].flow` | DEPRECATED | 现有实现能写入 `xtls-rprx-vision`，但 UI 建模仍旧。 |
| 无 UI，仅模型默认 | `settings.decryption` | `decryption` | KEEP | 当前默认 `none`，建议保留默认。 |
| `fallbacks.name` | `settings.fallbacks[].name` | `fallbacks[].name` | KEEP | 兼容字段。 |
| `fallbacks.alpn` | `settings.fallbacks[].alpn` | `fallbacks[].alpn` | KEEP | 兼容字段。 |
| `fallbacks.path` | `settings.fallbacks[].path` | `fallbacks[].path` | KEEP | 兼容字段。 |
| `fallbacks.dest` | `settings.fallbacks[].dest` | `fallbacks[].dest` | KEEP | 兼容字段。 |
| `fallbacks.xver` | `settings.fallbacks[].xver` | `fallbacks[].xver` | KEEP | 兼容字段。 |
| 无 | 无结构化 UI | `security=reality` + `realitySettings` | NEW | 26.5.3 主线缺口。 |
| 无现代 Vision 入口 | 间接通过 `flow` | `xtls-rprx-vision` | NEW | 当前是 partial，不是现代 UI。 |

### 1.3 Trojan

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `密码` | `settings.clients[0].password` | `clients[].password` | KEEP | 核心字段。 |
| `fallbacks.name` | `settings.fallbacks[].name` | `fallbacks[].name` | KEEP | 兼容字段。 |
| `fallbacks.alpn` | `settings.fallbacks[].alpn` | `fallbacks[].alpn` | KEEP | 兼容字段。 |
| `fallbacks.path` | `settings.fallbacks[].path` | `fallbacks[].path` | KEEP | 兼容字段。 |
| `fallbacks.dest` | `settings.fallbacks[].dest` | `fallbacks[].dest` | KEEP | 兼容字段。 |
| `fallbacks.xver` | `settings.fallbacks[].xver` | `fallbacks[].xver` | KEEP | 兼容字段。 |
| 无 | 无结构化 UI | REALITY 相关安全面 | NEW | 26.5.3 主线缺口。 |

### 1.4 Shadowsocks

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `加密` | `settings.method` | `method` | KEEP | 当前老三项仍可兼容。 |
| `密码` | `settings.password` | `password` | KEEP | 核心字段。 |
| `网络` | `settings.network` | `network` | KEEP | 当前 `tcp` / `udp` / `tcp,udp`。 |
| 无 | 无结构化 UI | `2022-blake3-aes-128-gcm` 等 SS 2022 方法 | NEW | 26.5.3 官方推荐能力缺失。 |
| 无 | 无结构化 UI | `xchacha20-poly1305` / `none` 等 | NEW | 当前方法集不完整。 |

### 1.5 SOCKS

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `密码认证` | `settings.auth` | `auth` | KEEP | `password` / `noauth`。 |
| `用户名` | `settings.accounts[0].user` | `accounts[].user` | KEEP | 当前只编辑首个 account。 |
| `密码` | `settings.accounts[0].pass` | `accounts[].pass` | KEEP | 当前只编辑首个 account。 |
| `启用 udp` | `settings.udp` | `udp` | KEEP | 当前保留。 |
| `IP` | `settings.ip` | `ip` | KEEP | 当前保留。 |
| 多账户 UI | 现有 JSON 支持数组 | `accounts[]` | NEW | 当前 UI 未结构化支持多账户。 |

### 1.6 HTTP

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `密码认证` | 隐含 `accounts` 是否为空 | `accounts[]` | KEEP | 当前 `auth` 只是 UI 计算概念。 |
| `用户名` | `settings.accounts[0].user` | `accounts[].user` | KEEP | 当前只编辑首个 account。 |
| `密码` | `settings.accounts[0].pass` | `accounts[].pass` | KEEP | 当前只编辑首个 account。 |
| 多账户 UI | 现有 JSON 支持数组 | `accounts[]` | NEW | 当前 UI 未结构化支持多账户。 |

### 1.7 Mixed

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `密码认证` | `settings.auth` | `auth` | KEEP | 复用 SOCKS 建模。 |
| `用户名` | `settings.accounts[0].user` | `accounts[].user` | KEEP | 当前只编辑首个 account。 |
| `密码` | `settings.accounts[0].pass` | `accounts[].pass` | KEEP | 当前只编辑首个 account。 |
| `启用 udp` | `settings.udp` | `udp` | KEEP | 复用 SOCKS 建模。 |
| `IP` | `settings.ip` | `ip` | KEEP | 复用 SOCKS 建模。 |

### 1.8 Tunnel

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `目标地址` | `settings.address` | `address` | KEEP | 当前与 Dokodemo 同形态。 |
| `目标端口` | `settings.port` | `port` | KEEP | 当前与 Dokodemo 同形态。 |
| `网络` | `settings.network` | `network` | KEEP | `tcp` / `udp` / `tcp,udp`。 |

### 1.9 Dokodemo

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `目标地址` | `settings.address` | `address` | KEEP | 核心字段。 |
| `目标端口` | `settings.port` | `port` | KEEP | 核心字段。 |
| `网络` | `settings.network` | `network` | KEEP | `tcp` / `udp` / `tcp,udp`。 |
| 无 | 无结构化 UI | `followRedirect` 等 | NEW | 当前 UI 未暴露更完整字段。 |

## 2. Transport Mapping

### 2.1 RAW (TCP)

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `RAW` | `stream.network="tcp"` | `RAW` (`network=raw` 的现代概念) | KEEP | 当前 UI 已重命名为 RAW，但 JSON 仍写 `tcp` 兼容。 |
| `HTTP 伪装` | `tcpSettings.header.type="http"` | `rawSettings.header.type=http` | DEPRECATED | 兼容保留，不应继续主推。 |
| `请求版本` | `tcpSettings.header.request.version` | `request.version` | KEEP | 兼容字段。 |
| `请求方法` | `tcpSettings.header.request.method` | `request.method` | KEEP | 兼容字段。 |
| `请求路径` | `tcpSettings.header.request.path[]` | `request.path[]` | KEEP | 兼容字段。 |
| `请求头` | `tcpSettings.header.request.headers` | `request.headers` | KEEP | 兼容字段。 |
| `响应版本` | `tcpSettings.header.response.version` | `response.version` | KEEP | 兼容字段。 |
| `响应状态` | `tcpSettings.header.response.status` | `response.status` | KEEP | 兼容字段。 |
| `响应状态说明` | `tcpSettings.header.response.reason` | `response.reason` | KEEP | 兼容字段。 |
| `响应头` | `tcpSettings.header.response.headers` | `response.headers` | KEEP | 兼容字段。 |
| 无 | 无结构化 UI | `acceptProxyProtocol` | NEW | 26.5.3 RAW 缺失字段。 |

### 2.2 WebSocket

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `WebSocket` | `stream.network="ws"` | `WebSocket` | KEEP | 当前主线 transport。 |
| `路径` | `wsSettings.path` | `path` | KEEP | 核心字段。 |
| `请求头` | `wsSettings.headers` | `headers` | KEEP | 核心字段。 |
| 无 | 无结构化 UI | `acceptProxyProtocol` 等附加字段 | NEW | 当前 UI 未暴露。 |

### 2.3 gRPC

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `gRPC` | `stream.network="grpc"` | `gRPC` | KEEP | 当前主线 transport。 |
| `serviceName` | `grpcSettings.serviceName` | `serviceName` | KEEP | 当前唯一结构化字段。 |
| 无 | 无结构化 UI | 其它 gRPC transport 细节 | NEW | 当前 UI 未暴露。 |

### 2.4 HTTP (Legacy)

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `HTTP (Legacy)` | `stream.network="http"` | 旧 `HTTP/2 transport` | DEPRECATED | 保留兼容，不是现代 HTTP 主线。 |
| `路径` | `httpSettings.path` | `path` | KEEP | 兼容字段。 |
| `host` | `httpSettings.host[]` | `host[]` | KEEP | 兼容字段。 |

### 2.5 KCP

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `mKCP (Legacy)` | `stream.network="kcp"` | `mKCP` | DEPRECATED | 兼容保留。 |
| `伪装` | `kcpSettings.header.type` | `header.type` | KEEP | 兼容字段。 |
| `密码` | `kcpSettings.seed` | `seed` | KEEP | 兼容字段。 |
| `mtu` | `kcpSettings.mtu` | `mtu` | KEEP | 兼容字段。 |
| `tti` | `kcpSettings.tti` | `tti` | KEEP | 兼容字段。 |
| `uplink capacity` | `kcpSettings.uplinkCapacity` | `uplinkCapacity` | KEEP | 兼容字段。 |
| `downlink capacity` | `kcpSettings.downlinkCapacity` | `downlinkCapacity` | KEEP | 兼容字段。 |
| `congestion` | `kcpSettings.congestion` | `congestion` | KEEP | 兼容字段。 |
| `read buffer size` | `kcpSettings.readBufferSize` | `readBufferSize` | KEEP | 兼容字段。 |
| `write buffer size` | `kcpSettings.writeBufferSize` | `writeBufferSize` | KEEP | 兼容字段。 |

### 2.6 QUIC

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `QUIC (Legacy)` | `stream.network="quic"` | `QUIC` | DEPRECATED | 兼容保留。 |
| `加密` | `quicSettings.security` | `security` | KEEP | 兼容字段。 |
| `密码` | `quicSettings.key` | `key` | KEEP | 兼容字段。 |
| `伪装` | `quicSettings.header.type` | `header.type` | KEEP | 兼容字段。 |

### 2.7 Missing Modern Transport

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| 无 | 无 | `network=xhttp` | NEW | 必须补齐。 |
| 无 | 无 | `network=httpupgrade` | NEW | 建议补齐。 |

## 3. TLS / XTLS / REALITY Mapping

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `TLS` 开关 | `stream.security="tls"` | `security=tls` | KEEP | 当前主线安全面。 |
| `XTLS (Deprecated)` 开关 | `stream.security="xtls"` | 旧 `security=xtls` 建模 | DEPRECATED | 保留兼容，后续应隐藏到 Compatibility。 |
| `SNI / 域名` | `tlsSettings.serverName` / `xtlsSettings.serverName` | `serverName` | KEEP | 当前命名基本一致。 |
| `证书文件路径` | `certificates[0].certificateFile` | `certificates[].certificateFile` | KEEP | 当前主链路有效。 |
| `密钥文件路径` | `certificates[0].keyFile` | `certificates[].keyFile` | KEEP | 当前主链路有效。 |
| `公钥内容` | `certificates[0].certificate[]` | `certificates[].certificate[]` | KEEP | 当前保留。 |
| `密钥内容` | `certificates[0].key[]` | `certificates[].key[]` | KEEP | 当前保留。 |
| 无 | 无结构化 UI | `alpn` | NEW | TLS Advanced。 |
| 无 | 无结构化 UI | `fingerprint` / uTLS | NEW | TLS Advanced。 |
| 无 | 无结构化 UI | `minVersion` | NEW | TLS Advanced。 |
| 无 | 无结构化 UI | `maxVersion` | NEW | TLS Advanced。 |
| 无 | 无结构化 UI | `cipherSuites` | NEW | TLS Advanced。 |
| 无 | 无结构化 UI | `enableSessionResumption` | NEW | TLS Advanced。 |
| 无 | 无结构化 UI | `disableSystemRoot` / `pinnedPeerCertSha256` 等 | NEW | TLS Advanced。 |
| 无 | 无 | `security=reality` | NEW | 必须补齐。 |
| 无 | 无 | `realitySettings.serverNames` | NEW | 必须补齐。 |
| 无 | 无 | `realitySettings.privateKey` | NEW | 必须补齐。 |
| 无 | 无 | `realitySettings.shortIds` | NEW | 必须补齐。 |
| 无 | 无 | `serverName`（REALITY client view） | NEW | 必须补齐。 |
| 无 | 无 | `fingerprint`（REALITY client view） | NEW | 必须补齐。 |
| 无 | 无 | `password` / 旧名 `publicKey` | NEW | 必须补齐。 |
| 无 | 无 | `shortId` | NEW | 必须补齐。 |
| 无 | 无 | `spiderX` | NEW | 必须补齐。 |

## 4. Sniffing Mapping

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `sniffing` | `sniffing.enabled` | `enabled` | KEEP | 当前已支持。 |
| `destOverride` | `sniffing.destOverride[]` | `destOverride[]` | KEEP | 当前已支持 `http` / `tls` / `quic` / `fakedns`。 |
| `metadataOnly` | `sniffing.metadataOnly` | `metadataOnly` | KEEP | 当前已支持。 |
| `routeOnly` | `sniffing.routeOnly` | `routeOnly` | KEEP | 当前已支持。 |
| 无 | 无结构化 UI | `domainsExcluded[]` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `ipsExcluded[]` | NEW | 当前缺失。 |
| 无独立字段 | 仅通过 `destOverride` 间接 | `fakedns` | KEEP | 当前以 `destOverride` 选项存在。 |

## 5. Mux Mapping

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| 出站 `Mux` 文本框 | `outbound.mux` 原始 JSON | `mux` | KEEP | 当前是 JSON 透传，不是结构化表单。 |
| 无结构化 UI | 需用户手写 JSON | `enabled` | NEW | 缺少结构化控件。 |
| 无结构化 UI | 需用户手写 JSON | `concurrency` | NEW | 缺少结构化控件。 |
| 无结构化 UI | 需用户手写 JSON | `xudpConcurrency` | NEW | 缺少结构化控件。 |
| 无结构化 UI | 需用户手写 JSON | `xudpProxyUDP443` | NEW | 缺少结构化控件。 |
| 无结构化 UI | 未见当前官方主文档要求 | `packetEncoding` | NEW | 当前 UI 未覆盖，不列为首轮必须。 |

## 6. Routing Mapping

| 当前 UI | JSON | 26.5.3 官方名称 | 状态 | 说明 |
|---|---|---|---|---|
| `type` | `rule.type` | `type` | KEEP | 当前默认 `field`。 |
| `domain` | `rule.domain[]` | `domain[]` | KEEP | 当前结构化支持。 |
| `ip` | `rule.ip[]` | `ip[]` | KEEP | 当前结构化支持。 |
| `port` | `rule.port` | `port` | KEEP | 当前结构化支持。 |
| `protocol` | `rule.protocol[]` | `protocol[]` | KEEP | 当前结构化支持。 |
| `network` | `rule.network` | `network` | KEEP | 当前结构化支持。 |
| `source` | `rule.source[]` | `source[]` | KEEP | 当前结构化支持。 |
| `inboundTag` | `rule.inboundTag[]` | `inboundTag[]` | KEEP | 当前结构化支持。 |
| `outboundTag` | `rule.outboundTag` | `outboundTag` | KEEP | 当前结构化支持。 |
| 无 | 无结构化 UI | `domainStrategy` | NEW | 当前页面未建模。 |
| 无 | 无结构化 UI | `sourcePort` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `localPort` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `sourceIP` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `localIP` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `user` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `attrs` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `process` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `balancerTag` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `ruleTag` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `webhook` | NEW | 当前缺失。 |
| 无 | 无结构化 UI | `balancers[]` | NEW | 当前缺失。 |

## 7. Modernization Matrix

### P0

必须优先进入现代化主线的能力：

- `REALITY`
- `XHTTP`
- `Vision`

说明：

- `REALITY` 是 26.5.3 主线安全面缺口
- `XHTTP` 是现代 HTTP transport 缺口
- `Vision` 当前只是通过旧 `XTLS` 开关 partial 支持，不是现代 UI

### P1

建议第二层补齐：

- TLS Advanced
  - `ALPN`
  - `Fingerprint`
  - `MinVersion`
  - `MaxVersion`
- Sniffing
  - `domainsExcluded`
  - `ipsExcluded`
  - 更清晰的 `destOverride`
- Mux
  - `enabled`
  - `concurrency`
  - `xudpConcurrency`
  - `xudpProxyUDP443`
- Routing advanced fields

### P2

保留兼容，不应直接删除：

- `XTLS`
- `alterId`
- `HTTP transport`
- `KCP`
- `QUIC`

## 8. Mapping Conclusion

必须保留的兼容参数：

- `alterId`
- `xtls`
- 旧 `tcp/http/kcp/quic` transport JSON
- `fallbacks`
- SOCKS / HTTP / Mixed 的认证字段

应该隐藏到 Compatibility 区的参数：

- `XTLS`
- 旧 `HTTP transport`
- `KCP`
- `QUIC`
- RAW 的 `HTTP header camouflage`

应该标 `Deprecated` 的参数：

- `alterId`
- `XTLS`
- 旧 `HTTP transport`
- `KCP`
- `QUIC`

应该新增的现代参数：

- `REALITY`
- `XHTTP`
- `Vision` 的现代 UI 表达
- TLS Advanced
- Sniffing Advanced
- Mux Advanced
- Routing Advanced
