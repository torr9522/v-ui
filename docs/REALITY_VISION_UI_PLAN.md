# REALITY / Vision UI Plan

## Scope

本文件只做 `REALITY / Vision` 的前端字段骨架设计与参数映射审计。

本轮不做：

- 数据库修改
- 后端 API 修改
- Xray 配置生成修改
- 入站保存逻辑修改
- REALITY 真正生效

目标是为下一轮编码建立边界，避免把现代参数继续硬塞进当前旧 `tls / xtls` 结构。

## Current v-ui Status

### Current Security State

当前仓库已经具备：

- `TLS` 正常 UI
- `XTLS` 兼容模式 UI
- `REALITY` 仅占位选项

当前 `REALITY` 的实际状态：

- `web/html/xui/inbound_modal.html`
  - `securityOptions()` 已显示 `REALITY`
  - 选择后只弹提示：
    - `REALITY support is planned for next phase`
  - 不会写入真实 `security=reality`
- `web/html/xui/form/tls_settings.html`
  - 显示 `REALITY is planned for next phase`
  - 没有任何 `realitySettings` 字段表单

结论：

- 当前只是 UI 入口占位
- 没有 REALITY 模型
- 没有 REALITY 保存
- 没有 REALITY 生成

### Current VLESS / Vision State

当前 `Vision` 相关能力只保留最小兼容：

- `web/html/xui/form/protocol/vless.html`
  - `inbound.settings.vlesses[0].flow`
  - 字段标题已调整为 `Vision Flow / Flow`
- `web/assets/js/model/xray.js`
  - `FLOW_CONTROL.VISION = "xtls-rprx-vision"`
  - `Inbound.VLESSSettings.VLESS` 当前只有：
    - `id`
    - `flow`

结论：

- 当前只能编辑 `flow`
- 还没有现代 `REALITY + Vision` 的结构化安全面
- `Vision` 仍依赖旧 `XTLS / flow` 入口暴露

### Current Legacy Compatibility Behavior

当前仓库已经内置旧 `flow` 兼容处理：

- `database/db.go`
- `web/service/inbound.go`

会清理旧值：

- `xtls-rprx-direct`
- `xtls-rprx-origin`

并保留：

- `xtls-rprx-vision`

这说明现阶段的正确策略应当是：

- 保持旧 `XTLS` 配置可回显、可编辑
- 不自动把旧 `XTLS` 节点迁移成 `REALITY`
- 让 `Vision` 先作为兼容字段继续存在

## Xray 26.5.3 Official REALITY Parameters

参考官方文档：

- REALITY
  - https://xtls.github.io/config/transports/reality.html
- VLESS
  - https://xtls.github.io/config/inbounds/vless.html

说明：

- `REALITY` 是传输安全层配置
- `Vision` 是 VLESS `flow` 的现代主线写法
- `VLESS + TCP/RAW + TLS/REALITY + flow=xtls-rprx-vision` 是当前主线组合

### Server-Side REALITY Fields

按官方配置，服务端入站侧核心字段包括：

| 字段 | 官方位置 | 作用 | UI 建议 |
|---|---|---|---|
| `dest` | `realitySettings.dest` | REALITY 回落目标 | Basic Advanced |
| `serverNames` | `realitySettings.serverNames[]` | 服务端允许的 SNI 列表 | Main |
| `privateKey` | `realitySettings.privateKey` | 服务端私钥 | Main |
| `shortIds` | `realitySettings.shortIds[]` | 服务端允许的 short ID 列表 | Main |
| `maxTimeDiff` | `realitySettings.maxTimeDiff` | 允许的时间偏差 | Advanced |
| `show` | `realitySettings.show` | 调试/展示相关开关 | Advanced |

### Client-Side REALITY Fields

客户端连接参数侧核心字段包括：

| 字段 | 客户端位置 | 作用 | UI 建议 |
|---|---|---|---|
| `serverName` | TLS / REALITY client setting | 客户端访问域名 | Main |
| `publicKey` | REALITY client setting | 服务端公钥 | Main |
| `shortId` | REALITY client setting | 单个 short ID | Main |
| `fingerprint` | REALITY client setting | uTLS 指纹 | Main |
| `spiderX` | REALITY client setting | 爬行路径伪装 | Advanced |
| `flow` | `settings.clients[0].flow` | `xtls-rprx-vision` | Main |

说明：

- 面板新增入站时主要录入的是服务端字段
- 分享链接或客户端导出时需要派生客户端字段
- 因此 `publicKey` 虽不是入站配置直接输入的服务端私有字段，但必须纳入 UI 设计，因为它决定后续分享链接和客户端导入体验

## Current v-ui to REALITY / Vision Mapping

### Already Directly Mappable

这些字段可以直接沿用当前 UI/模型习惯：

| 目标字段 | 当前可复用来源 | 说明 |
|---|---|---|
| `flow` | `inbound.settings.vlesses[0].flow` | 已存在，可继续承载 `xtls-rprx-vision` |
| `serverName` | 现有 `TLS` 区域 `SNI / 域名` | 后续可在 REALITY 共享显示表达 |
| `dest` | 当前 VLESS fallback 已有同名概念 | 名称用户熟悉，但语义不同，需单独说明 |

这些字段可以在前端模型层先增加壳体，但不能直接复用旧保存路径：

| 目标字段 | 原因 |
|---|---|
| `serverNames[]` | 当前没有数组模型 |
| `privateKey` | 当前没有 REALITY 安全面模型 |
| `shortIds[]` | 当前没有数组模型 |
| `publicKey` | 当前没有客户端导出模型 |
| `shortId` | 当前没有客户端导出模型 |
| `fingerprint` | 当前 TLS UI 也还未结构化 |
| `spiderX` | 当前没有高级表单承载 |
| `show` | 当前没有 REALITY Advanced 区 |
| `maxTimeDiff` | 当前没有 REALITY Advanced 区 |

### Fields That Require Generator Support

以下字段即使前端先做出来，也不能在本轮真正生效：

- `stream.security = "reality"`
- `stream.realitySettings.dest`
- `stream.realitySettings.serverNames`
- `stream.realitySettings.privateKey`
- `stream.realitySettings.shortIds`
- `stream.realitySettings.maxTimeDiff`
- `stream.realitySettings.show`
- 客户端导出的：
  - `publicKey`
  - `shortId`
  - `fingerprint`
  - `spiderX`

原因：

- 当前 `web/assets/js/model/xray.js` 没有 `realitySettings`
- 当前保存链路没有 REALITY JSON 结构
- 当前 Xray config 生成层没有 REALITY 输出
- 当前分享链接逻辑没有 REALITY 参数导出

结论：

- 下一轮第一步只能先补模型壳层和 disabled UI
- 真正可用必须等生成层和导出层一起补

## Recommended Frontend Skeleton

### Security Layout

`Security` 区建议明确拆成三层：

1. `None`
2. `TLS`
3. `REALITY`
4. `XTLS Legacy`

策略：

- `REALITY` 作为现代主线
- `XTLS` 留在 Compatibility
- 不再把 `Vision` 只依附在 `XTLS` 名义下显示

### VLESS / Vision Layout

对于 `VLESS`，建议把 `flow` 从旧 `XTLS` 语义中解耦：

- 字段标题：
  - `Vision Flow / Flow`
- 默认选项：
  - empty
  - `xtls-rprx-vision`
- 说明文字：
  - `Vision is the modern VLESS flow. Legacy XTLS configs remain editable for compatibility.`

这样可以同时满足：

- 旧 `XTLS` 回显不丢
- 未来 `REALITY + Vision` 有自然落点

### REALITY Main Fields

建议第一轮真正展示但默认 disabled / coming-soon 的字段：

| 分组 | 字段 | 展示方式 |
|---|---|---|
| Main | `serverNames` | tags / multi-line input |
| Main | `privateKey` | password-like input |
| Main | `shortIds` | tags / multi-line input |
| Main | `dest` | single input |
| Main | `publicKey` | readonly / coming-soon export field |
| Main | `shortId` | readonly / coming-soon client field |
| Main | `fingerprint` | select |
| Main | `flow` | existing Vision flow select |

说明：

- `publicKey` 更适合在“客户端参数 / 分享信息”区显示
- 但为了让用户理解 REALITY 成套参数，第一轮也可以先在 Security 区显示只读占位说明

### REALITY Advanced Fields

建议放到 `Advanced` 折叠区：

| 字段 | 原因 |
|---|---|
| `spiderX` | 普通用户不必第一屏看到 |
| `show` | 调试性质强 |
| `maxTimeDiff` | 低频高级项 |

### Disabled State Rules

在生成逻辑接入前，UI 层必须遵守：

- 可以显示字段
- 可以展示帮助文案
- 可以显示 coming-soon / disabled
- 不允许写入真实 `stream.security = reality`
- 不允许把占位字段混写到现有 `tlsSettings` 或 `xtlsSettings`
- 不允许污染旧配置保存数据

## Migration Strategy

### Do Not Auto-Migrate Legacy XTLS

禁止以下行为：

- 自动把 `XTLS` 节点转成 `REALITY`
- 自动清空旧 `flow`
- 自动替换旧 `security=xtls`

原因：

- 现有数据库里可能仍有老节点
- 用户需要继续编辑、查看和迁移旧配置
- 当前生成层尚未具备安全替换条件

### Keep Old Configs Editable

必须继续保证：

- 旧 `xtls` 节点能打开编辑
- 旧 `flow=xtls-rprx-vision` 能回显
- 旧 `http / kcp / quic` 等兼容项不被 REALITY UI 影响

### Future Transition Path

建议后续按这个顺序推进：

1. 先增加前端模型壳层
2. 再增加 REALITY UI disabled 字段
3. 再增加保存层与 `xray.js` 模型
4. 再增加 Xray config 生成
5. 最后增加分享链接 / 客户端导出字段

这样可以避免：

- UI 已经让用户能填
- 但保存和导出还没有实现
- 最终造成“看起来支持，实际上不可用”的二次回归

## First Coding Round Recommendation

第一轮真正编码时，建议只改以下文件：

- `web/assets/js/model/xray.js`
  - 增加 `RealitySettings` 前端模型
  - 不接入生成逻辑
- `web/html/xui/form/tls_settings.html`
  - 增加 `REALITY` 字段壳层
  - disabled / coming-soon
- `web/html/xui/form/protocol/vless.html`
  - 保留并整理 `Vision Flow / Flow`
  - 让它不再只依附旧 `XTLS`
- `web/html/xui/inbound_modal.html`
  - 增加安全模式联动
  - 保证 `REALITY` 占位不污染现有保存值

如需拆分模板，建议新增：

- `web/html/xui/form/reality_settings.html`

目的：

- 避免继续把现代字段堆进 `tls_settings.html`
- 给后续真实接入留出干净边界

## Conclusion

当前是否可以直接进入 REALITY 真正编码：

- 可以进入设计受控的第一轮编码
- 但不能直接跳到“保存 + 生成 + 分享链接”一起做

正确顺序应是：

1. 前端模型壳层
2. disabled / coming-soon REALITY 字段
3. 保存映射
4. Xray config 生成
5. 客户端导出与分享链接

在这个顺序下：

- 旧 `XTLS` 不会被破坏
- `Vision` 能先从语义上现代化
- `REALITY` 可以逐层落地，而不是一次性重写安全面
