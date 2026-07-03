# Custom Share Address Audit

## Scope

本文件只审查“自定义分享地址覆盖”是否适合做成纯前端临时覆盖能力。

本轮不做：

- 数据库修改
- Xray 配置修改
- Xray 重启
- 订阅逻辑改造
- REALITY 导出接入

目标能力限定为：

- 仅在“复制分享链接”或“生成自定义二维码”时临时替换连接地址
- 不影响原始链接
- 不影响原始二维码
- 不影响运行配置

## 1. 当前分享链路

### 1.1 入站列表页面位置

当前入站列表页面：

- `web/html/xui/inbounds.html`

相关动作：

- 列表菜单中的 `二维码`
- 列表中的 `详情`

### 1.2 入站详情弹窗位置

入站详情弹窗：

- `web/html/xui/inbound_info_modal.html`
- `web/html/xui/component/inbound_info.html`

当前行为：

- 弹窗 `OK` 按钮就是“复制链接”
- 复制内容直接来自：
  - `this.dbInbound.genLink()`

说明：

- 当前详情弹窗没有单独的“分享参数编辑区”
- 也没有“二维码 / 自定义地址”分支逻辑

### 1.3 分享链接生成主链

当前分享链接主链如下：

1. `web/html/xui/inbounds.html`
   - `showQrcode(dbInbound)`
   - `showInfo(dbInbound)`
2. `web/assets/js/model/models.js`
   - `DBInbound.genLink()`
3. `web/assets/js/model/xray.js`
   - `Inbound.genLink(address, remark)`
   - 再按协议分发到：
     - `genVmessLink`
     - `genVLESSLink`
     - `genTrojanLink`
     - `genSSLink`

结论：

- 原始链接生成完全在前端
- 没有调用后端“分享链接 API”
- 这非常适合做前端最小覆盖

## 2. 当前二维码链路

二维码弹窗：

- `web/html/common/qrcode_modal.html`

当前行为：

- `qrModal.show(title, content, okText, copyText)`
- `content` 用于二维码显示
- `copyText` 用于点击按钮复制
- 未传 `copyText` 时默认复制 `content`

从入站列表触发二维码时：

- `web/html/xui/inbounds.html`
  - `const link = dbInbound.genLink();`
  - `qrModal.show('二维码', link);`

结论：

- 二维码与复制文本当前都默认共用同一条原始链接
- 但 `qrModal.show()` 已经天然支持“二维码内容”和“复制内容”分离
- 这意味着后续若做“原始二维码 + 自定义二维码”或“原始链接 + 覆盖链接”都不需要后端支持

## 3. 当前订阅链路

本轮扫描结果：

- 仓库里没有发现明确的订阅页面
- 没有发现独立的订阅生成 UI
- 没有发现与入站分享导出绑定的 `subscription / subscribe` 主链实现

实际结论：

- 当前代码库中的“单节点分享链接”与“订阅”是分离的
- 至少在当前仓库层面，没有证据表明 `dbInbound.genLink()` 被订阅功能复用

因此：

- Custom Share Address Override 第一阶段只做前端单节点分享覆盖
- 不会天然波及订阅

## 4. 各协议生成位置

各协议分享链接生成都在：

- `web/assets/js/model/xray.js`

具体位置：

- `genVmessLink(address, remark)`
- `genVLESSLink(address, remark)`
- `genSSLink(address, remark)`
- `genTrojanLink(address, remark)`
- `genLink(address, remark)`

中间桥接层：

- `web/assets/js/model/models.js`
  - `DBInbound.genLink()`

## 5. VMess 格式确认

当前 `VMess` 仍然是传统格式：

- `vmess://` + `Base64(JSON)`

证据：

- `web/assets/js/model/xray.js`
  - `return 'vmess://' + base64(JSON.stringify(obj, null, 2));`

其中对象至少包含：

- `add`
- `port`
- `id`
- `aid`
- `net`
- `type`
- `host`
- `path`
- `tls`

结论：

- 当前不是 URI 风格的 `vmess://user@host:port?...`
- 后续实现必须保持传统 Base64 JSON 格式不变

## 6. 当前分享地址来源

### 6.1 默认地址来源

默认地址来源在：

- `web/assets/js/model/models.js`
  - `DBInbound.address`

逻辑：

- 先取 `location.hostname`
- 如果 `listen` 非空且不等于 `0.0.0.0`
  - 则改用 `listen`

结论：

- 当前分享地址不是来自后端请求 Host
- 也不是来自面板设置的 `webDomain`
- 而是前端页面当前访问主机名，或特定 `listen`

### 6.2 TLS / XTLS 对地址的二次覆盖

协议生成函数中还存在二次覆盖：

- `VMess`
  - `tls` 打开且 `stream.tls.server` 非空时
  - `address = this.stream.tls.server`
- `VLESS`
  - `tls` 打开且 `stream.tls.server` 非空时
  - `address = this.stream.tls.server`
  - 并写入 `sni`
- `Trojan`
  - `tls` / `xtls` 打开且 `stream.tls.server` 非空时
  - `address = this.stream.tls.server`
  - 并写入 `sni`

说明：

- 当前实际“最终导出地址”并不总是 `DBInbound.address`
- TLS 节点通常会优先导出 `tls serverName`

这对 Custom Share Address Override 的含义是：

- 若只在 `DBInbound.address` 层覆盖，TLS 场景仍可能被 `stream.tls.server` 抢回
- 因此下一轮不能简单替换 `DBInbound.address`
- 必须在“最终导出前”传入一个更高优先级的临时 `overrideAddress`

## 7. 原始复制链接与原始二维码是否共用同一函数

结论：

- 是，共用同一条前端生成函数链

详情弹窗复制：

- `dbInbound.genLink()`

二维码：

- `const link = dbInbound.genLink()`
- `qrModal.show('二维码', link)`

这意味着：

- 原始复制链接和原始二维码现在是一致的
- 如果要新增“自定义分享地址”，最稳妥的做法是新增“第二条临时导出路径”
- 不要直接替换现有 `genLink()` 的默认行为

## 8. embed / dist / assets / 模板缓存情况

当前前端资源来源：

- `web/html/*`
- `web/assets/*`

运行方式：

- `web/web.go`
  - 生产模式使用 Go `embed`
  - 调试模式直接读取本地 `web/html` 与 `web/assets`

结论：

- 没有单独的前端 `dist/` 构建产物
- 没有 React/Vite/Webpack 打包层
- 模板和 JS 改动点就是实际生效点
- 生产部署时需要重新编译二进制，单独替换 HTML 文件通常不会生效

## 9. 是否适合纯前端实现

结论：

- 适合

原因：

1. 当前链接生成完全在前端
2. 当前二维码内容也完全在前端
3. 当前没有后端分享链接 API 依赖
4. 当前 `Inbound.genLink(address, remark)` 已有“地址参数”入口
5. 功能目标只是临时导出覆盖，不涉及运行配置

## 10. 最小实现方案

建议第一轮只做“详情弹窗内的临时覆盖”，不碰保存结构。

### 10.1 建议交互

在入站详情弹窗增加一块“Custom Share Address”：

- 原始分享地址
- 自定义分享地址输入框
- `复制原始链接`
- `复制自定义链接`
- `原始二维码`
- `自定义二维码`

原则：

- 原始链路保留不变
- 自定义链路只在按钮点击时临时生成

### 10.2 建议实现边界

不要改：

- `DBInbound.address`
- `DBInbound.genLink()` 默认行为
- 订阅
- Xray 配置

建议新增：

- 在 `DBInbound` 或 `Inbound` 层增加可选参数：
  - `genLink(addressOverride, remarkOverride)`
  - 或新增辅助函数，不覆盖原函数默认路径

更稳妥的方式是：

- 保留原 `dbInbound.genLink()`
- 在详情弹窗中：
  - `const inbound = dbInbound.toInbound()`
  - `inbound.genLink(customAddress, dbInbound.remark)`

这样能保证：

- 原始分享链接继续完全不变
- 自定义覆盖只在弹窗内临时生效

### 10.3 TLS / Host / path / flow 保护要求

下一轮实现时必须确保：

- 只覆盖“连接地址”
- 不改：
  - `host`
  - `path`
  - `serviceName`
  - `flow`
  - `publicKey`
  - `shortId`
  - `spiderX`
  - `sni`
  - `serverName`

实现含义：

- 自定义分享地址应只替代“最终链接里的 host / add / authority 部分”
- 不能顺手把 TLS SNI 或 REALITY 参数一并改掉

## 11. 风险点

### 风险 1

TLS 节点当前会把地址自动改成 `stream.tls.server`。

如果下一轮直接改默认地址来源：

- 自定义覆盖可能不生效
- 或者反过来错误覆盖 `sni`

### 风险 2

VMess 必须继续保持传统 Base64 JSON。

如果下一轮为了统一协议实现而改成 URI 风格：

- 会直接破坏现有客户端兼容性

### 风险 3

Shadowsocks 当前只支持 plain SS 标准分享链接。

如果下一轮把自定义地址覆盖做成“所有协议统一增强”：

- 不能误把 TLS/stream 版 SS 当作完整可导出节点

### 风险 4

REALITY 当前仍是 UI skeleton。

如果下一轮把 custom share override 顺手接到 REALITY：

- 会造成“UI 看起来支持 REALITY 导出，但后端与生成层还没落地”的错觉

## 12. REALITY 边界

当前 `REALITY` 仍然只是 UI skeleton：

- 已有表单骨架
- 保存被阻止
- 没有 config generation
- 没有分享链接导出

因此：

- Custom Share Address Override 第一阶段不应接入 REALITY 导出
- 文档、代码和 UI 都应明确把 REALITY 排除在第一阶段之外

## 13. 下一轮编码建议

建议只改这些文件：

- `web/html/xui/inbound_info_modal.html`
  - 增加自定义地址输入与原始/自定义复制按钮
- `web/html/common/qrcode_modal.html`
  - 复用现有 `content / copyText` 分离能力
- `web/assets/js/model/models.js`
  - 保持原 `genLink()` 不变
  - 如有必要只加辅助调用入口
- `web/assets/js/model/xray.js`
  - 仅在协议生成层增加“临时地址覆盖优先级”处理
  - 不改协议参数本体

建议不要改：

- `web/html/xui/inbounds.html`
  - 第一轮不必把列表二维码动作一起复杂化
- 后端任意控制器或 service
- 数据库模型

## 14. 结论

本轮审查结论：

1. 当前单节点分享链路完全在前端
2. 当前二维码链路也完全在前端
3. 原始复制链接和原始二维码共用同一条 `genLink()` 生成链
4. 当前 `VMess` 明确仍是传统 `vmess://Base64(JSON)`
5. 当前仓库未发现需要同步修改的订阅主链
6. Custom Share Address Override 第一阶段非常适合做成前端最小实现
7. REALITY 当前仍是 UI skeleton，下一轮不应把该功能接入 REALITY 导出
