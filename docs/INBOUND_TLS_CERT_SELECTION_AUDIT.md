# Inbound TLS Certificate Selection Audit

## 1. z-ui 实现方式

### 1.1 前端行为

`z-ui` 确实有“入站 TLS 证书可选择”的能力，但它的实际行为不是“打开 TLS 后自动弹窗”，而是：

- 在入站弹窗打开时，预加载证书列表与系统待导入证书
- 当 `security=tls` 时，在 TLS 表单区域直接显示“本地证书”下拉
- 如果本地托管证书为空，但发现系统证书，则显示“系统证书”导入区
- 用户选择本地证书后，自动把 `certFile` / `keyFile` 写入当前入站
- 如果当前 `serverName` 为空，还会顺便把证书 `domain` 填到 `serverName`

关键前端文件：

- `z-ui`: `web/html/xui/inbound_modal.html`
- `z-ui`: `web/html/xui/form/tls_settings.html`

关键实现点：

- `inModal.show()` 时调用：
  - `loadCertificates()`
  - `loadDiscoveredCertificates()`
- TLS 区域显示条件：`inbound.stream.security === 'tls'`
- 证书选择联动：`applyCertificate(certFile)`
- 系统证书导入联动：`importDiscoveredCertificate(candidate)`

### 1.2 后端 API

`z-ui` 提供了专门的证书 API：

- `GET /api/certificates`
  - 返回托管证书列表
- `GET /api/certificates/discover`
  - 扫描系统证书来源
- `POST /api/certificates/import-discovered`
  - 把系统发现的证书导入到托管目录

关键文件：

- `z-ui`: `web/controller/certificate.go`
- `z-ui`: `web/service/certificate.go`
- `z-ui`: `web/entity/certificate.go`

### 1.3 证书来源

`z-ui` 的证书选择能力建立在“证书仓目录”之上，而不是直接读取当前面板 setting：

- 托管证书目录：`/etc/x-ui/certs/<name>/`
  - `fullchain.pem`
  - `privkey.pem`
  - `meta.json`
- 系统发现来源：
  - `/etc/letsencrypt/live`
  - `/root/.acme.sh`
  - `/home/*/.acme.sh`

`GET /api/certificates` 只列 `/etc/x-ui/certs` 下的托管证书。  
`GET /api/certificates/discover` 扫描外部来源并允许导入。

### 1.4 最终写入方式

`z-ui` 选择证书后，前端直接把：

- `streamSettings.tlsSettings.certificates[0].certificateFile`
- `streamSettings.tlsSettings.certificates[0].keyFile`

写到入站模型中，然后随入站保存一起提交。

后端入站保存本身不额外替换这两个字段，而是：

- 走协议模块 `Migrate`
- 走协议模块 `Validate`
- 走 `validateInboundRuntime`

也就是说，证书路径由前端填充，后端负责验证是否能用于实际运行。

### 1.5 校验强度

`z-ui` 的证书相关校验明显比 `v-ui` 完整：

- 导入已发现证书时：
  - 校验路径是否在允许扫描根目录内
  - 校验 `cert` / `key` 是否能组成有效 `tls.X509KeyPair`
- 入站保存前：
  - 协议层会校验部分 TLS / Reality 结构
  - `saveInboundWithClients()` 会跑 `validateInboundRuntime()`
  - 错误会在保存阶段暴露，而不是等 Xray 重启后才发现

### 1.6 Reality 处理

`z-ui` 的 `security` 是统一选择器：

- `none`
- `tls`
- `reality`

Reality 与 TLS 是分支逻辑：

- `tls` 时显示证书下拉 / 证书路径
- `reality` 时显示 `dest` / `privateKey` / `publicKey` / `shortIds` 等
- Reality 不使用 `certFile` / `keyFile`

## 2. v-ui 当前实现

### 2.1 前端现状

`v-ui` 当前入站 TLS 表单只有：

- TLS 开关
- XTLS 开关
- `serverName`
- 手工输入 `certFile`
- 手工输入 `keyFile`
- 或手工输入 PEM 内容

没有：

- 证书下拉
- 证书列表 API 调用
- 系统证书导入区
- 自动填充 `certFile` / `keyFile`

关键文件：

- `v-ui`: `web/html/xui/inbound_modal.html`
- `v-ui`: `web/html/xui/form/tls_settings.html`

### 2.2 后端现状

`v-ui` 当前证书后端只有“面板证书”相关 API：

- `/xui/cert/status`
- `/xui/cert/setDomain`
- `/xui/cert/checkDomain`
- `/xui/cert/upload`
- `/xui/cert/issueHttp`
- `/xui/cert/enableHttps`
- `/xui/cert/disableHttps`

它暴露的是“当前面板证书状态”，不是“可复用证书列表”。

关键文件：

- `v-ui`: `web/controller/cert.go`
- `v-ui`: `web/service/cert.go`
- `v-ui`: `web/entity/cert.go`

### 2.3 v-ui 当前证书来源

`v-ui` 没有 `z-ui` 那种 `/etc/x-ui/certs` 托管证书仓目录。

当前稳定主线只有两类面板证书落盘路径：

- 手动上传：
  - `/usr/local/x-ui/cert/panel.crt`
  - `/usr/local/x-ui/cert/panel.key`
- ACME HTTP-01：
  - `/usr/local/x-ui/cert/acme-panel.crt`
  - `/usr/local/x-ui/cert/acme-panel.key`

当前状态 API 通过 setting 与模式推导展示路径：

- `webCertMode=manual` -> 面板手动证书路径
- `webCertMode=acme_http` -> 面板 ACME 证书路径

### 2.4 入站保存现状

`v-ui` 当前 `InboundService` 对 TLS 证书路径没有 z-ui 那种保存前的运行时校验链路：

- 没有证书列表抽象
- 没有证书路径存在性校验
- 没有 keypair 匹配校验
- 没有 `validateInboundRuntime()`

现状更接近：

- 前端把 `streamSettings` 原样拼出来
- 后端做协议/限速/端口等基础处理
- 保存数据库
- 后续靠 Xray 重启时暴露问题

### 2.5 Reality 现状

当前 `v-ui v1.0 Stable` 入站 UI 没有 `z-ui` 那种 `security=none/tls/reality` 统一模式，也没有对应 Reality 表单。

因此本轮“TLS 证书自动选择”审计不需要考虑 Reality 兼容分支迁移。  
结论是：

- `z-ui`：TLS / Reality 都在同一套安全层 UI 下
- `v-ui`：当前只讨论 TLS 自动填路径，不涉及 Reality

## 3. 对比表

| 项目 | z-ui | v-ui | 差距 | 是否可照搬 |
|---|---|---|---|---|
| TLS 开关 UI | `security` 统一选择 `none/tls/reality` | `tls` / `xtls` 布尔开关 | 交互模型不同 | 部分可照搬 |
| 证书选择 UI | 有“本地证书”下拉；无证书时有“系统证书”导入区 | 只有手工路径输入 | 缺证书选择层 | 可照搬思路 |
| 证书列表 API | `GET /api/certificates` | 无 | 缺后端列表接口 | 需要新做 |
| 系统证书扫描 API | `GET /api/certificates/discover` | 无 | 缺扫描/导入体系 | 不建议直接照搬 |
| 当前面板证书复用 | 依赖证书先进入 `/etc/x-ui/certs` | 只有当前 panel/manual/acme 路径状态 | 数据模型不同 | 不能直接照搬 |
| 证书目录扫描 | 有 `/etc/letsencrypt/live`、`~/.acme.sh` 扫描 | 无 | 完整缺失 | 不建议第一轮引入 |
| cert/key 校验 | 导入时校验 keypair；保存前跑 runtime validate | 目前没有对应链路 | 校验不足 | 第一轮不必全搬 |
| 入站保存结构 | 前端写 `certificateFile/keyFile`，后端模块验证 | 前端写 `certFile/keyFile`，后端基本直存 | 保存链路更弱 | 前端写入部分可照搬 |
| Reality 是否受影响 | Reality 单独分支，不用证书 | 当前稳定版无 Reality UI | 架构不同 | 不需要迁移 |
| TLS 是否受影响 | TLS 下拉选择 + 自动填入 | TLS 仅手填路径 | 体验差距明显 | 建议补齐 |

## 4. 结论

### 4.1 z-ui 是否有这个功能

有。  
但准确说法是：

- `z-ui` 有“TLS 打开后可选择已托管证书”的功能
- 如果没有已托管证书，还能显示系统发现证书并导入
- 它不是“自动弹窗”，而是“TLS 表单内联下拉/导入区”

### 4.2 v-ui 缺什么

`v-ui` 当前主要缺三层：

1. 可复用证书列表抽象
2. 入站 TLS 表单联动
3. 保存前的证书路径校验链路

### 4.3 哪些能直接照搬

可直接照搬的不是“完整证书仓系统”，而是这些最小交互：

- 打开入站弹窗时加载证书列表
- TLS 打开后显示“本地证书”选择控件
- 选择证书后自动填 `certFile` / `keyFile`
- 若 `serverName` 为空，则自动填证书域名
- 保留手工路径输入作为兜底

### 4.4 哪些不能直接照搬

以下内容不适合直接搬到 `v-ui v1.0` 主线：

- `/etc/x-ui/certs` 全量证书仓
- letsencrypt / acme.sh 外部目录扫描
- discover + import-discovered 整套系统证书导入
- Reality 联动改造

原因：

- `v-ui` 当前证书架构只有“面板活动证书”
- 没有证书库存目录和 `meta.json` 体系
- 直接移植会把 `v-ui` 又带回复杂证书平台方向

## 5. v-ui 最小实现方案

按 `z-ui` 思路，但只做 `v-ui` 当前主线需要的最小版本：

### 5.1 后端 API

新增：

- `POST /xui/cert/listUsable`

返回“当前可复用的面板证书候选”，而不是完整证书仓。

建议字段：

```json
[
  {
    "name": "Panel Certificate",
    "certFile": "/usr/local/x-ui/cert/acme-panel.crt",
    "keyFile": "/usr/local/x-ui/cert/acme-panel.key",
    "source": "panel",
    "domain": "cshtps.527270.xyz",
    "issuer": "Let's Encrypt",
    "expireAt": 1790599698,
    "valid": true
  }
]
```

候选来源只限：

- 当前活动 panel 手动证书
- 当前活动 panel ACME HTTP-01 证书
- 可选：如果手动证书文件和 ACME 证书文件都存在，也可同时列出

不做：

- 系统目录 discover
- acme.sh 目录扫描
- 第三方证书仓

### 5.2 前端改造

在 `TLS` 打开后：

- 请求 `/xui/cert/listUsable`
- 如果返回 1 个证书：
  - 自动填 `certFile` / `keyFile`
- 如果返回多个证书：
  - 显示一个“本地证书”下拉
- 保留手工路径输入

写入目标保持现有结构：

- `streamSettings.tlsSettings.certificates[0].certificateFile`
- `streamSettings.tlsSettings.certificates[0].keyFile`

### 5.3 UI 行为建议

建议仿 `z-ui` 的内联模式，不做强制弹窗：

- TLS 打开后，直接在 TLS 表单区显示证书选择
- 如果只有一个候选，可直接自动填充
- 如果多个候选，再让用户显式选择

### 5.4 校验建议

第一轮只做轻量校验：

- 前端若选中证书后发现 `valid=false`，显示 warning
- 前端若接口返回空列表，提示“未检测到可复用证书，请手工填写路径”
- 后端先不引入完整 runtime validate

## 6. API 设计建议

建议新增 entity：

```json
{
  "name": "Panel Certificate",
  "certFile": "/usr/local/x-ui/cert/acme-panel.crt",
  "keyFile": "/usr/local/x-ui/cert/acme-panel.key",
  "source": "panel",
  "domain": "cshtps.527270.xyz",
  "issuer": "YE1",
  "expireAt": 1790599698,
  "valid": true
}
```

说明：

- `name`: UI 展示名称
- `source`: 仅需区分 `panel-manual` / `panel-acme`
- `valid`: 基于文件存在 + keypair 可解析

## 7. 前端改造点

下一轮如果编码，最小改动建议集中在：

- `web/html/xui/inbound_modal.html`
- `web/html/xui/form/tls_settings.html`

建议后端最小改动：

- `web/controller/cert.go`
- `web/entity/cert.go`
- `web/service/cert.go`

## 8. 风险

主要风险：

1. `v-ui` 当前没有证书库存概念  
   如果未来继续扩展，很容易滑向“复杂证书平台”

2. `v-ui` 当前无保存前 runtime validate  
   第一轮若只做 UI 自动填充，保存阶段仍可能接受无效路径

3. `z-ui` 与 `v-ui` 的 TLS 安全模型不同  
   `z-ui` 是 `security=tls/reality`，`v-ui` 当前是 `tls/xtls` 开关

## 9. 下一轮编码建议

只建议做最小实现：

1. 新增 `POST /xui/cert/listUsable`
2. TLS 打开后自动拉取候选证书
3. 单候选自动填充 `certFile` / `keyFile`
4. 多候选显示下拉
5. 保留手工输入

不要做：

- 多证书资产平台
- 证书仓库管理
- DNS provider
- Cloudflare
- 自动签发
- 入站级证书申请
- 每个入站独立 ACME

## 10. 封顶线

本能力只用于：

- 复用已经安装好的面板证书
- 帮用户自动填 TLS 路径

不把 `v-ui` 扩展成完整证书管理系统。
