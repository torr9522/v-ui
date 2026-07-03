# REALITY Data Model

## Scope

本轮只补齐前端 `REALITY` 数据模型。

明确不包含：

- 数据库字段
- 后端 API 保存
- Xray 配置生成
- 分享链接 / 订阅 / 二维码
- REALITY 真正可运行支持

当前 `REALITY` 仍保持：

- Experimental
- 不可作为可运行配置提交
- 不影响现有 `TLS / XTLS` 保存链路

## Model Location

- `web/assets/js/model/xray.js`

当前前端模型类：

- `RealityStreamSettings`

该模型只用于承载 UI 与后续阶段参数映射，不代表本轮已经接入保存和生成。

## Field Mapping

以下字段按 `Xray-core 26.5.3` 的 `realitySettings` / `VLESS flow` 需求整理：

| 字段 | 默认值 | 作用 | 对应 Xray 26.5.3 |
|---|---|---|---|
| `dest` | `""` | REALITY 回落目标，通常是 `host:port` | `realitySettings.dest` |
| `serverNames` | `[]` | 允许的 SNI 列表 | `realitySettings.serverNames` |
| `privateKey` | `""` | 服务端私钥 | `realitySettings.privateKey` |
| `shortIds` | `[]` | 服务端允许的 short ID 列表 | `realitySettings.shortIds` |
| `show` | `false` | 调试/显示开关 | `realitySettings.show` |
| `spiderX` | `""` | 客户端伪装路径占位 | `realitySettings.spiderX` |
| `maxTimeDiff` | `0` | 允许时间偏差 | `realitySettings.maxTimeDiff` |
| `fingerprint` | `"chrome"` | uTLS 指纹占位 | `realitySettings.fingerprint` |
| `publicKey` | `""` | 客户端导出占位字段 | 客户端 REALITY `publicKey` |
| `shortId` | `""` | 客户端单个 short ID 占位 | 客户端 REALITY `shortId` |
| `flow` | `""` | Vision 流控占位，后续用于 `xtls-rprx-vision` | `settings.clients[].flow` |

## Compatibility Boundary

本轮保持以下边界不变：

- `StreamSettings.toJson()` 仍不输出 `realitySettings`
- 现有 `TLS` 保存逻辑不变
- 现有 `XTLS` 兼容逻辑不变
- 现有 `VMess / VLESS / Trojan / Shadowsocks` 行为不变

这意味着：

- 旧配置仍按原逻辑工作
- `REALITY` 只是更完整的前端模型
- 不会把未完成的 `REALITY` 数据误送进当前生成链路

## Next Phase Split

下一轮如果进入 `API 保存`，优先处理：

- `dest`
- `serverNames`
- `privateKey`
- `shortIds`
- `show`
- `maxTimeDiff`
- `fingerprint`
- `spiderX`
- `flow`

后续进入 `Config Generator` 时，再真正接入：

- `stream.security = "reality"`
- `stream.realitySettings.*`
- `VLESS flow = xtls-rprx-vision`
- 客户端导出所需的 `publicKey / shortId / fingerprint / spiderX`

## Notes

- `publicKey` 当前仅作为后续导出占位，不参与本轮保存。
- `flow` 进入模型只是为 `Vision` 做前端字段对齐，不代表本轮启用 `REALITY + Vision`。
- 本文件对应的是模型准备阶段，不是功能完成阶段。
