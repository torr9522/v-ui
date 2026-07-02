# Protocol Modernization Plan

## Goal

Modern UI 第二阶段的目标不是重做内核适配层，而是：

- 在不破坏现有 `x-ui` 运行主架构的前提下
- 把当前 UI 参数表达更新到更接近 `Xray 26.5.3` 的主线用法

本计划只定义顺序、边界和禁止事项。

## Core Rules

必须遵守：

- 不直接删数据库字段
- 不直接删旧 JSON 字段
- 不在没有映射层的情况下硬改 `streamSettings`
- 不把现代化工作扩展成协议大重构

## Must Keep

以下参数必须保留兼容：

- `alterId`
- `xtls`
- `tcp` 旧 JSON 写法
- `http` 旧 transport 写法
- `kcp`
- `quic`
- `fallbacks`
- 现有 SOCKS / HTTP / Mixed 认证字段

原因：

- 这些字段已经可能存在于现有数据库
- 直接删除会破坏旧节点编辑与回显

## Should Hide

以下参数应从主路径隐藏到 Compatibility：

- `XTLS`
- `HTTP (Legacy)`
- `mKCP (Legacy)`
- `QUIC (Legacy)`
- RAW 的 `HTTP camouflage`

原则：

- 不删兼容
- 不放主路径

## Should Mark Deprecated

应明确标 Deprecated 的项：

- `alterId`
- `XTLS`
- 旧 `HTTP transport`
- `KCP`
- `QUIC`

UI 文案建议：

- `Deprecated`
- `Compatibility Only`
- `Legacy`

## Should Add

第二阶段真正应该补齐的是：

- `REALITY`
- `XHTTP`
- `Vision`
- TLS Advanced
  - `ALPN`
  - `Fingerprint`
  - `MinVersion`
  - `MaxVersion`
- Sniffing Advanced
  - `domainsExcluded`
  - `ipsExcluded`
- Mux Advanced
  - `enabled`
  - `concurrency`
  - `xudpConcurrency`
  - `xudpProxyUDP443`

## Phase Order

### Phase 2A

先做现代参数表达层，不改生成逻辑。

内容：

- 统一 transport 命名
  - `RAW`
  - `WebSocket`
  - `gRPC`
  - `HTTP (Legacy)`
  - `mKCP (Legacy)`
  - `QUIC (Legacy)`
- 统一 security 命名
  - `TLS`
  - `REALITY`
  - `Compatibility / XTLS`
- 在 UI 中明确 Deprecated / Compatibility 分区

目标：

- 先把用户看到的词汇和结构理顺
- 不立刻触碰 JSON 生成

### Phase 2B

补 `REALITY` 和 `Vision` 的结构化表单。

内容：

- `security=reality`
- `serverNames`
- `privateKey`
- `shortIds`
- `fingerprint`
- `password/publicKey`
- `shortId`
- `spiderX`
- `flow=xtls-rprx-vision`

说明：

- 这是 P0 主线
- 必须优先于“把所有旧字段都优化一遍”

### Phase 2C

补 `XHTTP` 和 `HTTPUpgrade`。

内容：

- 现代 HTTP transport 入口
- 与旧 `HTTP (Legacy)` 明确分离

### Phase 2D

补 TLS Advanced 与 Sniffing Advanced。

内容：

- `ALPN`
- `Fingerprint`
- `MinVersion`
- `MaxVersion`
- `domainsExcluded`
- `ipsExcluded`

### Phase 2E

补 Mux / Routing advanced 字段。

内容：

- Mux XUDP
- Routing advanced match fields

## First Coding Step

Modern UI Phase 2 的第一步不应该直接写 `REALITY` 生成逻辑。

第一步应该是：

### 建立前端现代参数壳层

也就是先在 UI 和模型层建立这两个选择面：

1. `Transport`
   - `RAW`
   - `WebSocket`
   - `gRPC`
   - `XHTTP`
   - `HTTP (Legacy)`
   - `mKCP (Legacy)`
   - `QUIC (Legacy)`

2. `Security`
   - `TLS`
   - `REALITY`
   - `Compatibility / XTLS`

要求：

- 先做显示层和参数组织层
- 保证旧配置仍可回显
- 先不改数据库
- 先不改 API
- 先不改生成逻辑

原因：

- 如果先写 `REALITY` 生成逻辑，现有 UI 结构会继续把现代参数塞进旧 `tls/xtls` 模型
- 先做参数壳层，后续新增字段才不会继续污染旧模型

## Never Delete

以下能力永远不能在“现代化”名义下直接删除：

- 旧节点可编辑性
- 旧数据库字段可读取性
- 旧分享链接基础能力
- 现有 `x-ui` 运行兼容层

## Conclusion

Modern UI 第二阶段应按这个顺序推进：

1. 参数壳层
2. REALITY / Vision
3. XHTTP
4. TLS Advanced / Sniffing
5. Mux / Routing Advanced

核心原则只有一句：

- 先把现代参数表达层建立起来，再碰真实生成逻辑
