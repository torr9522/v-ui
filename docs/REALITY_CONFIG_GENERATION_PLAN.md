# REALITY Config Generation Plan

## Scope

本文件只做 `REALITY Support Phase 3` 设计审计。

本轮不做：

- Go 代码修改
- 前端修改
- 数据库修改
- 放开 REALITY 保存
- 分享链接 / 订阅 / 二维码接入

目标是明确：

- 当前入站生成链路
- REALITY 应如何进入最终 Xray 配置
- 哪些字段应进入服务端配置
- 哪些字段只属于客户端导出
- 第一轮编码最小改动范围

## 1. 当前生成链路

### Frontend -> HTTP

当前新增/编辑入站的前端提交入口：

- `web/html/xui/inbounds.html`

关键行为：

- `settings: inbound.settings.toString()`
- `streamSettings: inbound.stream.toString()`
- `sniffing: inbound.sniffing.toString()`

结论：

- `streamSettings` 由前端直接序列化成 JSON 字符串提交
- 目前没有单独的 REALITY API 结构

### HTTP -> DB

当前绑定入口：

- `web/controller/inbound.go`

关键行为：

- `POST /xui/inbound/add`
- `POST /xui/inbound/update/:id`
- `ShouldBind(inbound)` 直接把 `streamSettings` 绑定到 `model.Inbound.StreamSettings`

服务层：

- `web/service/inbound.go`

当前行为：

- `normalizeLimit(inbound)`
- `normalizeProtocolSettings(inbound)`
- `checkPortExist(...)`
- 直接 `Save(inbound)`

结论：

- 当前后端对 `streamSettings` 没有结构化校验
- 旧的 `VLESS flow` 兼容清理发生在 `settings.clients[].flow`
- REALITY 第一轮最合适的校验落点是 `web/service/inbound.go`

### DB -> Xray Config

当前生成入口：

- `database/model/model.go`
- `web/service/xray.go`

关键行为：

- `model.Inbound.GenXrayInboundConfig()` 直接把
  - `Settings`
  - `StreamSettings`
  - `Sniffing`
  作为原始 JSON 放入 `xray.InboundConfig`
- `web/service/xray.go` 的 `getLegacyXrayConfig()` / `getManagedXrayConfigForDB()` 直接 `append(inbound.GenXrayInboundConfig())`

结论：

- 当前入站 `streamSettings` 基本是“原样透传”
- `web/service/xray.go` 不是 REALITY 字段转换主战场
- REALITY 第一轮编码更像“保存前校验 + 透传前裁剪”

## 2. 当前 TLS / XTLS 处理方式

当前前端模型：

- `web/assets/js/model/xray.js`

当前行为：

- `security=tls` 时输出 `tlsSettings`
- `security=xtls` 时输出 `xtlsSettings`
- `security=reality` 时前端已能输出 `realitySettings`

当前后端行为：

- 不会单独翻译 `tlsSettings / xtlsSettings`
- Xray 最终拿到的是数据库中保存的 `streamSettings` 原始 JSON

结论：

- REALITY 第一轮必须保持这套模式不变
- 只在 `security=reality` 时增加校验和必要清理
- 不要顺手改 `tls` / `xtls` 的任何字段命名和输出策略

## 3. REALITY 官方字段映射

官方依据：

- Project X REALITY
  - https://xtls.github.io/config/transports/reality.html
- Project X VLESS inbound
  - https://xtls.github.io/config/inbounds/vless.html

基于官方文档，`streamSettings.security = reality` 时的结构应为：

```json
{
  "network": "tcp",
  "security": "reality",
  "realitySettings": {
    "target": "example.com:443",
    "serverNames": ["example.com"],
    "privateKey": "...",
    "shortIds": ["0123456789abcdef"],
    "show": false,
    "maxTimeDiff": 0
  }
}
```

说明：

- 官方文档当前主字段名是 `target`
- 官方说明 `dest` 是旧称，`target` 与 `dest` 互为 alias
- 为避免第一轮牵连前端字段大改，v-ui 可继续接受前端 `dest`
- 但服务端生成建议标准化输出为 `target`

### 字段映射表

| v-ui 前端字段 | 第一轮服务端生成 | Xray 26.5.3 官方位置 | 结论 |
|---|---|---|---|
| `stream.security` | `reality` | `streamSettings.security` | 必须 |
| `stream.network=tcp` | 保持 `tcp` | 文档主线称 `RAW` | 第一轮保留现有 `tcp` 输出，避免破坏旧链路 |
| `stream.reality.dest` | 映射为 `target` | `realitySettings.target` | 必填 |
| `stream.reality.serverNames[]` | 原样输出 | `realitySettings.serverNames` | 必填 |
| `stream.reality.privateKey` | 原样输出 | `realitySettings.privateKey` | 必填 |
| `stream.reality.shortIds[]` | 原样输出 | `realitySettings.shortIds` | 必填 |
| `stream.reality.show` | 原样输出 | `realitySettings.show` | 选填，服务端字段 |
| `stream.reality.maxTimeDiff` | 原样输出 | `realitySettings.maxTimeDiff` | 选填，服务端字段 |
| `stream.reality.publicKey` | 不进入服务端生成 | 客户端 `password`（旧名 `publicKey`） | 导出字段，忽略 |
| `stream.reality.shortId` | 不进入服务端生成 | 客户端 `shortId` | 导出字段，忽略 |
| `stream.reality.fingerprint` | 不进入服务端生成 | 客户端 `fingerprint` | 导出字段，忽略 |
| `stream.reality.spiderX` | 不进入服务端生成 | 客户端 `spiderX` | 导出字段，忽略 |
| `stream.reality.flow` | 不进入 `realitySettings` | 不属于 `streamSettings` | 见 Vision 关系 |

### 当前阶段不进入第一轮的官方字段

这些字段在官方文档存在，但不建议在第一轮接入：

- `xver`
- `minClientVer`
- `maxClientVer`
- `mldsa65Seed`
- `mldsa65Verify`
- `limitFallbackUpload`
- `limitFallbackDownload`

原因：

- 当前前端无结构化输入
- 当前目标是最小 REALITY 可生成，不是一次覆盖全部高级参数

## 4. REALITY 与 Vision Flow 的关系

官方依据：

- `VLESS UserObject.flow` 位于 `settings.clients[].flow`
- 官方示例与文档说明 `xtls-rprx-vision` 是 `VLESS` 用户字段，不属于 `streamSettings.realitySettings`

当前仓库现状：

- 前端 VLESS flow 在 `inbound.settings.vlesses[0].flow`
- `web/service/inbound.go` / `database/db.go` 已有旧 flow 清理逻辑

设计结论：

- `flow` 的权威来源必须继续是 `settings.clients[].flow`
- 第一轮生成代码必须忽略 `stream.reality.flow`
- 若 UI 为兼容保留 `stream.reality.flow` 占位，最多只作为前端提示，不应参与后端生成

### 是否强制 `xtls-rprx-vision`

建议：

- 第一轮不在后端强制改写 flow
- 但当 `security=reality` 且 `protocol=vless` 时，建议校验为：
  - 允许空值通过
  - 若非空，推荐只接受 `xtls-rprx-vision`

原因：

- 避免破坏现有兼容数据
- 保持第一轮范围最小
- 真正的“自动默认 Vision”更适合在后续 UI/交互轮次处理

## 5. 必填校验规则

REALITY 第一轮建议只支持：

- 协议：`VLESS`
- 传输：`tcp`（UI 显示 `RAW`）
- 安全：`reality`

### 基础组合校验

必须满足：

- `protocol == vless`
- `streamSettings.security == reality`
- `streamSettings.network == tcp`

不满足时：

- 后端直接拒绝保存

### REALITY 字段必填

#### `target`

来源：

- 当前前端字段 `dest`

规则：

- trim 后不能为空
- 格式建议为 `host:port`

#### `serverNames`

规则：

- 数组不能为空
- 至少一个有效条目
- 第一轮建议至少包含一个非空域名，避免先引入“空 SNI”高级模式

说明：

- 官方允许 `""` 作为无 SNI 接入特例
- v-ui 第一轮不建议开放该特例，避免导出链路和用户理解复杂化

#### `privateKey`

规则：

- trim 后不能为空

#### `shortIds`

规则：

- 数组不能为空
- 每项 trim 后：
  - 可以为空字符串 `""` 吗？
    - 官方允许
    - 第一轮建议默认不允许空字符串，降低配置复杂度
  - 非空时必须满足：
    - 十六进制字符
    - 长度不超过 16
    - 长度为偶数

说明：

- 若未来需要兼容空 `shortId` 模式，再作为高级选项开放

## 6. Optional 字段分类

这里需要明确区分“服务端 optional”和“客户端导出 optional”。

### 服务端 optional

- `show`
- `maxTimeDiff`

### 客户端导出 required / optional

- `fingerprint`
  - 客户端侧必需，默认可用 `chrome`
  - 不应进入服务端 `realitySettings`
- `spiderX`
  - 客户端侧可选
  - 不应进入服务端 `realitySettings`
- `publicKey`
  - 客户端侧必需
  - 官方新名称是 `password`，旧称 `publicKey`
  - 当前阶段只保留前端占位，不进入服务端生成
- `shortId`
  - 客户端侧通常需要与服务端 `shortIds` 之一匹配
  - 当前阶段只保留前端占位，不进入服务端生成

## 7. 不影响旧配置的策略

第一轮编码必须坚持以下策略：

### 只处理 `security=reality`

- `tls` 分支不变
- `xtls` 分支不变
- 非 REALITY 入站完全不走新校验

### 不改现有字段名体系

- 前端仍允许用 `dest`
- 后端保存/生成时再标准化成 `target`
- 不在这一轮同时改 `tcp -> raw`

### 不写入客户端导出字段

生成服务端 config 时，必须显式忽略：

- `publicKey`
- `shortId`
- `fingerprint`
- `spiderX`
- `flow`（以 `settings.clients[].flow` 为准）

### 不改数据库结构

- `streamSettings` 仍然只是文本 JSON
- 不新增列
- 不迁移旧数据

## 8. 第一轮编码建议

建议第一轮只改这些文件：

### 必改

- `web/service/inbound.go`
  - 新增 REALITY 校验与规范化
  - 只在 `protocol=vless && security=reality` 时生效
- `web/html/xui/inbound_modal.html`
  - 放开“前端保存阻止”，改为依赖后端校验
  - 或在切换到 REALITY 后只保留前端提示，不再直接 return

### 可能改

- `web/assets/js/model/xray.js`
  - 若需要把 `dest` 与 `target` 的内部表达统一，可在这里做最小兼容
  - 若不做统一，可不改

### 尽量不要改

- `web/service/xray.go`
  - 当前不是主战场
  - 除非要新增一次“最终透传前过滤”
- `database/model/model.go`
  - 当前原样透传已足够
- 数据库层
  - 本阶段不需要

## 9. 真机测试矩阵

第一轮 REALITY config generation 编码完成后，建议真机矩阵如下：

### REALITY 正向

1. `VLESS + RAW(TCP) + REALITY`
2. `flow=xtls-rprx-vision`
3. `target/serverNames/privateKey/shortIds` 全部有效
4. 保存成功
5. Xray 重启成功
6. 面板可重新打开并回显

### REALITY 失败校验

1. 缺 `target`
2. 缺 `serverNames`
3. 缺 `privateKey`
4. 缺 `shortIds`
5. `shortId` 长度为奇数
6. `shortId` 非 hex
7. `protocol != vless`
8. `network != tcp`

预期：

- 保存被拒绝
- 返回明确错误
- 不写入坏配置
- 不触发 Xray 进入坏状态

### 回归

1. `VLESS + TLS`
2. `VLESS + XTLS`
3. `VMess + WS + TLS`
4. `Trojan + RAW + TLS`
5. TLS 面板证书自动选择
6. 分享链接原逻辑

预期：

- 行为与 `v1.0.2-stable` 一致

## 10. 结论

可以进入 REALITY config generation 编码，但建议只做最小第一轮：

1. 只支持 `VLESS + RAW(TCP) + REALITY`
2. 后端在 `web/service/inbound.go` 做组合校验与字段校验
3. 服务端生成只输出：
   - `security=reality`
   - `realitySettings.target/serverNames/privateKey/shortIds/show/maxTimeDiff`
4. `flow` 继续来自 `settings.clients[].flow`
5. `publicKey / shortId / fingerprint / spiderX` 继续留在导出阶段，不进入服务端生成

这条路径对现有 TLS / XTLS 影响最小，也最符合当前仓库“streamSettings 原样透传”的架构。
