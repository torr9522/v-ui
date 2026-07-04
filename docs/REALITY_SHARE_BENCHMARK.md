# REALITY Share / Export Benchmark

## Scope

本文件只做 `REALITY Share / Export Benchmark Review`。

本轮不做：

- 代码修改
- DB 修改
- API 修改
- Xray config generator 修改
- 分享链接功能上线
- 订阅 / 二维码接入

目标是先回答三件事：

1. Project X / Xray 官方对 REALITY 与 VLESS URI 到底定义了什么
2. 3x-ui / v2rayN / NekoBox 等主流实现实际上导出了哪些参数
3. 在不破坏 `v-ui` 当前架构的前提下，下一轮应该怎么接入

## 1. Project X 官方方案

### 1.1 官方 JSON 配置是明确的

本轮核对的官方资料：

- Project X REALITY transport:
  - `https://xtls.github.io/en/config/transports/reality.html`
- Project X VLESS inbound:
  - `https://xtls.github.io/en/config/inbounds/vless.html`

官方文档明确区分了两类字段：

服务端 `streamSettings.realitySettings`：

- `target` / 旧名 `dest`
- `serverNames`
- `privateKey`
- `shortIds`
- `show`
- `maxTimeDiff`

客户端 REALITY 字段：

- `serverName`
- `fingerprint`
- `password`（旧名 `publicKey`）
- `shortId`
- `spiderX`

结论：

- `privateKey` 是服务端字段，绝不能进入分享链接
- `publicKey/password` 是客户端字段，必须导出
- `shortIds` 是服务端候选集合，导出时应变成单个 `shortId`
- `serverNames` 是服务端候选集合，导出时应变成单个 `serverName` / `sni`

### 1.2 官方对 `flow` 的归属也是明确的

VLESS 官方文档把 `flow` 放在用户对象，即：

- `settings.clients[].flow`

而不是：

- `streamSettings.realitySettings`

结论：

- `xtls-rprx-vision` 属于 VLESS client/user
- 不应把 Vision flow 塞进 REALITY server settings

### 1.3 官方没有“配置文档级别”的稳定 URI 规范页

本轮没有找到像 JSON 配置文档那样稳定、独立的“REALITY share URI 规范页”。

能公开核对到的“URI 参数标准”主要来自 Xray 官方 GitHub Discussion：

- `VLESS URI Standard`
  - `https://github.com/XTLS/Xray-core/discussions/716`

该讨论列出的 VLESS URL 查询参数包含：

- `security`
- `type`
- `flow`
- `sni`
- `fp`
- `pbk`
- `sid`
- `spx`

结论：

- REALITY 分享 URI 更接近“官方维护中的互操作标准”
- 设计上应以官方文档定义字段语义，以该讨论和客户端实现确认 URI 参数命名

## 2. 3x-ui 实现

### 2.1 定位到的源码

本轮实际定位到的 3x-ui 关键源码：

- `internal/sub/service.go`
  - `genVlessLink(...)`
  - `genTrojanLink(...)`
  - `applyShareRealityParams(...)`
- `internal/sub/json_service.go`
  - `realityData(...)`
- `internal/sub/service_sharelink_test.go`
- `internal/sub/service_flow_test.go`
- `internal/util/link/outbound.go`
- `internal/util/link/outbound_helpers_test.go`

### 2.2 3x-ui 导出 REALITY 的做法

`genVlessLink(...)` 在 `security == reality` 时调用 `applyShareRealityParams(...)`。

实际导出行为：

- `security=reality`
- 从 `serverNames[]` 里选一个导出为 `sni`
- 从 `shortIds[]` 里选一个导出为 `sid`
- 从 `realitySettings.settings.publicKey` 导出 `pbk`
- 从 `realitySettings.settings.fingerprint` 导出 `fp`
- 从 `realitySettings.settings.spiderX` 派生出每客户端稳定的 `spx`
- 当 `client.Flow` 存在且规则允许时，导出 `flow`

3x-ui 的测试还明确锁定了：

- `pbk` 与 `sid` 不能串位
- `tcp + reality` 必须保留 `flow=xtls-rprx-vision`
- `xhttp + reality` 是否导出 `flow` 取决于它自己的 VLESS encryption 条件

### 2.3 值得借鉴的点

- `flow` 继续从 client settings 取，不从 stream settings 取
- `serverNames[]` / `shortIds[]` 导出时都应降为单值
- `spx` 若实现，最好是“同一客户端稳定、不同客户端不同”

### 2.4 不应照搬的点

3x-ui 当前把客户端导出相关字段挂在：

- `realitySettings.settings.publicKey`
- `realitySettings.settings.fingerprint`
- `realitySettings.settings.spiderX`

这不适合直接照搬到 `v-ui`，原因：

- `v-ui` 当前服务端 REALITY 保存链路已经按官方 server-side 结构最小化
- `v-ui` 当前目标是不改 DB / 不改 config generator 主线
- 直接复制 3x-ui 的 Reality 存储形状会把分享导出字段重新混进服务端配置层

结论：

- 3x-ui 的导出参数选择值得参考
- 但其内部存储结构不适合作为 `v-ui` 的直接模板

## 3. v2rayN 实现

### 3.1 定位到的源码

本轮实际定位到的 v2rayN 关键源码：

- `v2rayN/ServiceLib/Handler/Fmt/BaseFmt.cs`
- `v2rayN/ServiceLib/Handler/Fmt/VLESSFmt.cs`
- `v2rayN/ServiceLib/Handler/ConfigHandler.cs`
- `v2rayN/ServiceLib/Handler/Builder/NodeValidator.cs`
- `v2rayN/ServiceLib.Tests/Fmt/FmtHandlerTests.cs`

### 3.2 v2rayN 导出 / 解析 REALITY 的做法

`BaseFmt.ToUriQuery(...)` 明确处理：

- `security`
- `sni`
- `fp`
- `pbk`
- `sid`
- `spx`
- `alpn`
- `type`
- `host`
- `path`
- `serviceName`

`VLESSFmt.ToUri(...)` 还会补：

- `encryption=none`
- `flow`

`ResolveUriQuery(...)` / `VLESSFmt.Resolve(...)` 会反向解析这些字段。

额外值得注意的行为：

- `ConfigHandler.AddServerCommon(...)` 会在 `security=reality` 且 `fingerprint` 为空时，补默认指纹
- `NodeValidator` 对 REALITY 至少要求 `publicKey` 非空

结论：

- v2rayN 使用的 REALITY URI 参数集非常完整
- `pbk` / `sid` / `spx` / `fp` 是其一等公民，不是私有扩展

## 4. NekoBox 实现

### 4.1 定位到的源码

本轮实际定位到的 NekoBoxForAndroid 关键源码：

- `app/src/main/java/io/nekohasekai/sagernet/fmt/v2ray/V2RayFmt.kt`
- `app/src/main/java/io/nekohasekai/sagernet/fmt/v2ray/StandardV2RayBean.java`
- `app/src/main/res/xml/standard_v2ray_preferences.xml`

### 4.2 NekoBox 的 REALITY URI 行为

解析侧：

- 接受 `security=reality`
- 读取 `pbk`
- 读取 `sid`
- 读取 `fp`

导出侧：

- VLESS/Trojan 使用 URI 形式
- 当 `realityPubKey` 非空时，把 `security` 改写为 `reality`
- 导出 `pbk`
- 导出 `sid`
- `fp` 仅在已有值时导出
- VLESS 的 `flow` 继续沿用自身 VLESS 字段导出

本轮源码审查没有发现 NekoBox 对 `spx` 的明确解析/导出实现。

### 4.3 结论

- `pbk` / `sid` / `fp` 是 NekoBox 可公开确认的 REALITY URI 互通字段
- `spx` 不是所有客户端都明确实现的公共最小集

## 5. Shadowrocket

本轮没有拿到可公开核对的 Shadowrocket 源码或官方解析实现。

因此本轮只能得出保守结论：

- 公开世界里大量 Shadowrocket REALITY 分享示例沿用 `vless://...?...security=reality...`
- 常见参数仍是 `type` / `security` / `sni` / `pbk` / `sid` / `fp` / `flow`
- 但没有足够的源码级证据证明其对 `spx`、多 `serverNames` 选择策略等细节的真实处理

结论：

- Shadowrocket 可以作为“互通目标”参考
- 不能作为 `v-ui` 的规范来源
- 下一轮设计仍应以 Project X + v2rayN + NekoBox + 3x-ui 交集为准

## 6. v-ui 当前实现

### 6.1 当前分享链路

当前 `v-ui` 分享链路集中在：

- `web/assets/js/model/xray.js`

关键函数：

- `genVmessLink(...)`
- `genVLESSLink(...)`
- `genTrojanLink(...)`
- `genSSLink(...)`
- `genLink(...)`

现状：

- `VMess` 继续使用 `vmess:// + Base64(JSON)`
- `VLESS` / `Trojan` / `SS` 都是前端直接生成链接
- 当前没有后端 share-link API
- 当前订阅链路不与单节点 `genLink()` 绑定

### 6.2 当前 REALITY 状态

当前 `v-ui` 已完成：

- REALITY 前端模型
- REALITY 表单 UI
- REALITY 最小保存
- REALITY 最小 config generation

当前最小支持组合：

- `VLESS + TCP(RAW) + REALITY + xtls-rprx-vision`

### 6.3 当前导出缺口

当前 `genVLESSLink(...)`：

- 会导出 `type`
- 会导出 `security`
- `tls` 时会导出 `sni`
- `xtls` 时会导出 `flow`
- 还没有 REALITY 导出逻辑

当前后端最小 REALITY 保存链路会保留：

- `target`
- `serverNames`
- `privateKey`
- `shortIds`
- `show`

不会保留：

- `publicKey`
- `shortId`
- `fingerprint`
- `spiderX`

这意味着：

- `v-ui` 下一轮若要做 REALITY 分享导出，不能简单“直接读 DB 现成字段”
- 至少要解决 `publicKey` 的导出来源

## 7. REALITY 分享链接最小公共参数

以下表格回答“最小公共参数到底有哪些、来自哪里、是否该导出”。

| 参数 | 角色 | 来源 | 是否导出 | 说明 |
|---|---|---|---|---|
| `address` | 客户端连接地址 | 面板分享地址 / listen / 自定义分享地址 | 是 | 这是 URI 的主机部分，不属于 `realitySettings`。 |
| `port` | 客户端连接端口 | `inbound.port` | 是 | URI authority 的一部分。 |
| `uuid` | VLESS 客户端身份 | `settings.clients[].id` | 是 | 只用于 VLESS。 |
| `security` | 传输安全类型 | `streamSettings.security` | 是 | REALITY 场景必须是 `reality`。 |
| `type` | 传输类型 | `streamSettings.network` | 是 | `v-ui` 当前最小支持只应导出 `tcp`。 |
| `flow` | VLESS Vision flow | `settings.clients[].flow` | 条件导出 | 仅从 client settings 取；对当前 `v-ui` 最小支持应导出 `xtls-rprx-vision`。 |
| `sni` | 客户端 REALITY serverName | 从 `serverNames[]` 选一项 | 是 | 建议用首项或明确策略选单项。 |
| `pbk` | 客户端 REALITY public key | 由服务端 `privateKey` 推导 | 是 | 必须导出；绝不能导出 `privateKey`。 |
| `sid` | 客户端 REALITY shortId | 从 `shortIds[]` 选一项 | 是 | 分享时只能是单值。 |
| `fp` | 客户端 uTLS 指纹 | 客户端导出字段 | 建议导出 | v2rayN / 3x-ui / NekoBox 都识别；建议显式给 `chrome`。 |
| `spx` | REALITY spiderX | 客户端导出字段 | 可选 | 3x-ui / v2rayN 明确支持；NekoBox 本轮未确认。 |
| `alpn` | TLS/REALITY ALPN | 客户端导出字段 | 非当前最小必需 | 当前 `v-ui` 最小 REALITY 未覆盖。 |
| `host` | WS/HTTP host | stream transport | 仅对应传输导出 | 当前 `v-ui` 最小 REALITY 不应导出。 |
| `path` | WS/HTTP/gRPC path/service | stream transport | 仅对应传输导出 | 当前 `v-ui` 最小 REALITY 不应导出。 |
| `serviceName` | gRPC service | stream transport | 仅 gRPC 导出 | 当前 `v-ui` 最小 REALITY 不应导出。 |
| `encryption` | VLESS encryption | VLESS setting | 是 | 对普通 VLESS URI 应继续为 `none`。 |
| `network` | 语义同 `type` | 无需单独来源 | 否 | URI 层面用 `type` 即可，不必额外导出 `network`。 |

## 8. 哪些来自 Server，哪些来自 Client，哪些不应导出

### 8.1 服务端保存字段

属于 server config，只能存服务端，不应原样导出：

- `target` / `dest`
- `serverNames[]`
- `privateKey`
- `shortIds[]`
- `show`
- `maxTimeDiff`

### 8.2 客户端导出字段

属于 share/export 客户端视角，应导出或按需导出：

- `address`
- `port`
- `uuid`
- `security`
- `type`
- `encryption`
- `flow`
- `sni`
- `pbk`
- `sid`
- `fp`
- `spx`

### 8.3 明确不应该导出的字段

- `privateKey`
- `target` / `dest`
- `serverNames[]` 整个数组
- `shortIds[]` 整个数组
- `show`
- `maxTimeDiff`

导出规则应为：

- `serverNames[]` -> 选一个 `sni`
- `shortIds[]` -> 选一个 `sid`
- `privateKey` -> 推导 `pbk`

## 9. 推荐最终方案

### 9.1 Phase 1 只做 `v-ui` 当前已支持组合

下一轮 REALITY share/export 第一轮建议只覆盖：

- `protocol = vless`
- `type = tcp`
- `security = reality`
- `flow = xtls-rprx-vision`（若客户端有值）

不要一开始就做：

- Trojan REALITY
- WS REALITY
- gRPC REALITY
- XHTTP
- 订阅
- 二维码

原因：

- 当前 `v-ui` 运行层真正支持的最小 REALITY 就是这一条路径
- 先把最小公共互通链路做对，风险最低

### 9.2 URI 参数策略

第一轮建议导出：

- `vless://uuid@address:port`
- `encryption=none`
- `type=tcp`
- `security=reality`
- `flow=xtls-rprx-vision`（若存在）
- `sni=<serverNames[0]>`
- `pbk=<derived public key>`
- `sid=<shortIds[0]>`
- `fp=chrome`

第一轮建议暂不强制导出：

- `spx`
- `alpn`

原因：

- `pbk/sid/fp` 是主流公共交集
- `spx` 虽然被 3x-ui 与 v2rayN 明确支持，但 NekoBox 本轮未确认
- 当前 `v-ui` 也没有稳定的每客户端 `spx` 派生上下文

### 9.3 `publicKey` 的处理建议

这是 `v-ui` 当前最大架构问题。

因为当前保存链路只保留：

- `privateKey`

不保留：

- `publicKey`

所以第一轮编码前必须定一个推导策略：

方案 A：前端即时推导 `publicKey`

- 优点：不改 DB / 不改 API
- 风险：当前 `v-ui` 前端没有现成 X25519 推导工具，浏览器兼容性和实现风险高

方案 B：后端提供最小辅助推导层

- 只根据已保存的 `privateKey` 推导 `publicKey`
- 不改 DB
- 不改 Xray config generator
- 不改变入站保存语义

本轮推荐：

- 若坚持最小风险，优先选择“最小辅助推导层”
- 若后续确认前端能稳定引入极小的 X25519 推导工具，再考虑纯前端

### 9.4 为什么不照搬 3x-ui

因为 3x-ui 的 share/export 已经绑定了：

- 更复杂的 subscription server
- 更复杂的 REALITY 客户端字段存储结构
- `spx` 派生与多客户端策略
- XHTTP / 更多协议路径

而 `v-ui` 当前最合适的落点仍然是：

- 保持分享链在 `web/assets/js/model/xray.js`
- 只为 REALITY 的 VLESS 最小路径补导出
- 不碰 DB / 不碰 config generator

## 10. 为什么这样设计

推荐方案的核心原因：

1. 它与 `v-ui` 当前已上线的 REALITY 最小运行能力一致。
2. 它只接入主流客户端共有的 REALITY 参数，不发明私有 URI。
3. 它不需要复制 3x-ui 的 subscription/backend 架构。
4. 它不把服务端 `privateKey`、`target`、`shortIds[]` 原样暴露给客户端。
5. 它允许后续再扩展 `spx`、二维码、订阅，而不会推翻第一轮设计。

## 11. 下一轮编码建议

### 11.1 可以开始编码吗

可以，但建议只做“最小 REALITY VLESS share link”。

### 11.2 第一轮应该只改哪些文件

首选最小改动范围：

- `web/assets/js/model/xray.js`
  - 在现有 `genVLESSLink(...)` 内加入 REALITY 分支
- `tests/share_link_smoke.js`
  - 增加 REALITY share/export smoke cases

如果 `publicKey` 无法在前端稳定推导，再额外引入一层最小辅助：

- `web/service/server.go`
  - 仅做 `privateKey -> publicKey` 推导辅助
- 或单独的极小 Go utility/helper 文件

### 11.3 哪些地方绝不能改

- DB schema
- 现有后端 API 语义
- Xray config generator 主链
- VMess `vmess:// + Base64(JSON)` 逻辑
- 现有非 REALITY 分享链接行为
- 订阅链路
- 二维码链路
- REALITY 最小运行配置行为

## 12. 结论

结论如下：

- Project X 官方已经明确了 REALITY 的服务端字段与客户端字段边界
- 主流客户端公共交集至少包括 `security=reality`, `type`, `sni`, `pbk`, `sid`, `fp`
- `flow` 必须继续属于 VLESS client settings
- `privateKey` 绝不能导出
- `v-ui` 下一轮最合理的第一步，是只给当前已支持的 `VLESS + TCP + REALITY` 加最小 share link 导出
- `spx` 应作为后续增强项，而不是第一轮阻断项
