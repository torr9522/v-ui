# Z-UI Alignment Review

## 1. z-ui 实际域名证书逻辑

### 1.1 主线结论

重新审计 `/tmp/v-ui-init/z-ui-source` 后，z-ui 当前主线的域名证书逻辑是：

- 证书集中存放在 `/etc/x-ui/certs/<name>/`
- 通过 `x-ui.sh` 做证书管理
- 面板 HTTPS 通过写入 `webCertFile` / `webKeyFile` 启用
- 域名证书申请使用 `acme.sh standalone HTTP-01`
- 续期使用 `x-ui cert renew` + `systemd timer`
- 面板侧主要提供“证书发现/导入”能力，不是完整 ACME/DNS provider 面板

对应证据：

- `knowledge-base/07-certificate-system.md`
- `README.md`
- `x-ui.sh`
- `web/service/certificate.go`
- `web/controller/certificate.go`
- `web/web.go`

### 1.2 z-ui 是否真的做 Cloudflare DNS-01

主线结论：`否`。

当前 z-ui 主线没有看到以下内容：

- Cloudflare Token 配置页面
- Cloudflare Zone 管理页面
- DNS provider 管理体系
- 面板侧自动 TXT 创建/删除流程
- 主线 `x-ui.sh` 的 Cloudflare DNS-01 申请命令

审计结果显示，z-ui 主线证书申请入口在 `x-ui.sh:884` 的 `cmd_cert_issue_acme()`，明确执行的是：

```bash
"${acme}" --issue --standalone -d "${domain}"
```

这就是 `HTTP-01 standalone`，不是 `DNS-01`。

### 1.3 z-ui 是否存在 Cloudflare 相关痕迹

有，但不是主线实现。

在 `x-ui_en.sh` 中存在旧的：

- `ssl_cert_issue()`
- `ssl_cert_issue_by_cloudflare`

这更像旧英文脚本/遗留脚本能力，不是当前 z-ui 主线架构。主线 `x-ui.sh`、README、knowledge-base、panel web/controller/service 中都没有形成完整 Cloudflare DNS-01 产品链路。

因此应判断为：

- `x-ui_en.sh` 里有历史/旁支 Cloudflare 痕迹
- z-ui 当前主线不是 Cloudflare DNS provider 驱动的证书系统

### 1.4 z-ui 的证书申请入口在哪里

主要入口在 CLI：

- `x-ui.sh:884` `cmd_cert_issue_acme()`

功能流程：

1. 输入域名
2. 检查域名 A 记录是否解析到当前服务器 IP
3. 检查 80 端口是否可用
4. 安装或复用 `acme.sh`
5. 执行 `acme.sh --issue --standalone -d <domain>`
6. 执行 `--install-cert --fullchain-file --key-file`
7. 写入证书元数据
8. 安装续期 timer

### 1.5 z-ui 的域名保存在哪里

z-ui 没有看到“独立的 panel domain setting 页面/字段主线”。

现有域名相关存储位置主要是：

- 证书元数据：`/etc/x-ui/certs/<name>/meta.json`
- 面板 HTTPS 指向：数据库 `settings.webCertFile` / `settings.webKeyFile`

也就是说：

- 证书对应域名保存在 `meta.json`
- 面板是否启用 HTTPS 由 `webCertFile` / `webKeyFile` 决定
- 不是一个复杂的“域名 + DNS provider”数据模型

### 1.6 z-ui 是否有 Cloudflare Token 页面

没有在 z-ui 主线 panel 中发现。

搜索范围：

- `web/html`
- `web/controller`
- `web/service`

没有发现 Cloudflare 配置页、Token 页、Zone 页、DNS 测试页。

### 1.7 z-ui 是否有 DNS provider 管理

没有。

未发现：

- provider 列表
- provider 抽象层
- DNS 凭据设置体系
- provider 切换 UI

### 1.8 z-ui 是否有自动 TXT 记录管理

没有发现主线实现。

主线 `x-ui.sh` 做的是：

- `HTTP-01 standalone`

不是：

- `_acme-challenge` TXT 自动创建
- TXT 传播轮询
- TXT 清理

### 1.9 z-ui 的 HTTPS 启用逻辑是什么

z-ui 的面板 HTTPS 启用逻辑非常直接：

1. 选择一个证书目录
2. 将该目录下的 `fullchain.pem` / `privkey.pem` 写到 DB：
   - `webCertFile`
   - `webKeyFile`
3. 重启或 reload `x-ui`

证据：

- `x-ui.sh:856` `cmd_cert_set_panel_https()`
- `web/web.go` 启动时读取 `webCertFile` / `webKeyFile`
- 如果两个值存在，就 `tls.LoadX509KeyPair(...)` 并启动 TLS listener

### 1.10 z-ui 的续期逻辑是什么

z-ui 续期逻辑是：

- `x-ui cert renew`
- 只续期 30 天内到期的证书
- `acme.sh --renew -d <domain>`
- 再执行 `--install-cert --fullchain-file --key-file --reloadcmd ...`
- 由 `x-ui-cert-renew.timer` 每天执行

证据：

- `x-ui.sh:301` `install_cert_renew_timer()`
- `x-ui.sh:1010` `renew_one_certificate()`
- `x-ui.sh:1061` `cmd_cert_renew()`
- `x-ui.sh:1072` `cmd_cert_autorenew()`

## 2. v-ui 当前实现对齐情况

### 2.1 符合 z-ui 主线目标的内容

以下内容与 z-ui 主线方向一致，建议保留：

- 域名保存
  - `webDomain`
- 域名解析检查
  - `CheckDomain`
- 手动证书上传
  - 上传 cert/key
- HTTPS 启用/禁用
  - `webCertFile` / `webKeyFile`
- HTTP-01 自动申请
  - `acme.sh standalone`
- fullchain 落盘
  - `acme-panel.crt` 使用 fullchain
- HTTPS startup self-heal
  - 证书文件丢失时自动回退 HTTP

这些能力虽然实现方式比 z-ui 更面板化，但目标一致：

- 仍以 `x-ui` runtime 为核心
- 仍以本地证书文件和 `webCertFile/webKeyFile` 为核心
- 仍以 `acme.sh HTTP-01` 为主线

### 2.2 可能偏离 z-ui 主线的内容

以下内容明显超出了 z-ui 当前主线：

- Cloudflare API 页面
  - `web/html/xui/cloudflare.html`
- Cloudflare Controller
  - `web/controller/cloudflare.go`
- Cloudflare Service
  - `web/service/cloudflare.go`
- Cloudflare setting keys
  - `cloudflareApiToken`
  - `cloudflareZoneID`
  - `cloudflareAccountID`
  - `cloudflareEmail`
  - `cloudflareApiKey`
  - `cloudflareEnabled`
- Cloudflare DNS-01 申请入口
  - `POST /xui/cert/issueDnsCloudflare`
- 自动 TXT 创建/删除
  - `CreateTXTRecord`
  - `DeleteTXTRecord`
- DNS 传播检测
  - TXT polling

这套能力已经从“模仿 z-ui 域名证书逻辑”走向了“扩展一个 DNS provider 系统”。

## 3. 偏离项判断

### 3.1 是否确认 Cloudflare DNS-01 是偏离主线

确认。

原因：

1. z-ui 主线证书申请入口是 `HTTP-01 standalone`
2. z-ui 主线没有 Cloudflare Token 页面
3. z-ui 主线没有 DNS provider 管理体系
4. z-ui 主线没有自动 TXT 管理流程
5. v-ui 当前 Cloudflare 功能已经显著超出“照着 z-ui 搬/模仿/复刻/借鉴”

### 3.2 偏离程度

- `7d2816c Add Cloudflare API integration`
  - 中度偏离
  - 已引入单独配置页、API、settings、service
- `5be8c63 Add Cloudflare DNS-01 certificate issue`
  - 重度偏离
  - 已把证书主线扩展到 DNS provider 自动化

## 4. 处理方案

### 方案 A：保留 Cloudflare 代码，标记 experimental，并隐藏菜单/按钮

做法：

- 保留 service/controller/settings/test
- 从侧边栏隐藏 `Cloudflare`
- 从 `/xui/cert` 页面隐藏 `Cloudflare DNS-01` 按钮
- 不在 README/文档主线宣传
- 标注为 experimental / non-mainline

优点：

- 不丢已有代码
- 改动最小
- 后续若要单独孵化，可以继续

缺点：

- 主线仓库仍然背着一套偏离 z-ui 的 provider 体系
- settings / API / service 仍在
- 后续维护认知成本仍然存在

适用：

- 想先止血，不想立刻删代码

### 方案 B：回滚 7d2816c 和 5be8c63，删除 Cloudflare API / DNS-01

做法：

- 回滚 `7d2816c Add Cloudflare API integration`
- 回滚 `5be8c63 Add Cloudflare DNS-01 certificate issue`
- 回到 HTTP-01 + 手动证书 + HTTPS toggle + startup self-heal 主线

优点：

- 与 z-ui 主线最一致
- 架构最干净
- 不再引入 DNS provider 复杂度
- 认知边界清晰

缺点：

- 已写的 Cloudflare 相关代码全部撤回
- 需要一次完整清理和回归验证

适用：

- 明确要求“只复刻 z-ui 需要的域名逻辑”
- 不希望主线继续带着复杂 provider 设计

### 方案 C：保留 Cloudflare API service，但删除/隐藏 DNS-01 申请入口

做法：

- 保留 `CloudflareService`
- 保留 Token/Zone 测试 API
- 隐藏 `Cloudflare DNS-01` 证书申请入口
- 隐藏 Cloudflare 页面入口，或仅内部保留

优点：

- 比 B 的改动小
- 比 A 更收敛

缺点：

- 仍然保留一套与 z-ui 主线不一致的配置模型
- 后续别人会继续问“既然有 Cloudflare 页面，为什么不作为主线”

适用：

- 想保留部分基础设施，但暂时不暴露 DNS-01

## 5. 推荐方案

推荐：**方案 B**

原因：

1. 最符合“照着 z-ui 域名逻辑做”的目标
2. z-ui 主线实际没有 Cloudflare provider 产品化体系
3. v-ui 当前真正需要的主线已经足够完整：
   - 域名保存
   - 域名检查
   - 手动上传
   - HTTPS 启用/禁用
   - HTTP-01
   - fullchain
   - startup self-heal
4. 继续保留 Cloudflare API / DNS-01 会持续把项目推向复杂 DNS 系统

如果当前不想立刻删代码，次优方案是：

- **先执行方案 A**
- 但中期仍建议走向 **方案 B**

## 6. 如果要回滚，应该回滚哪些 commit

严格按偏离项回滚：

- `7d2816c Add Cloudflare API integration`
- `5be8c63 Add Cloudflare DNS-01 certificate issue`

保留以下与 z-ui 主线一致的 commit：

- `c879fd7 Add ACME HTTP-01 certificate issue`
- `1bc444c Fix acme.sh executable discovery`
- `35c234f Use ACME fullchain for panel certificate`
- `79e9828 Add HTTPS startup self-heal`

## 7. 如果要保留，哪些入口要隐藏

若不立即回滚，至少应隐藏：

- 侧边栏 `Cloudflare` 菜单
  - `web/html/xui/common_sider.html`
- 页面 `/xui/cloudflare`
  - `web/controller/xui.go`
- `Cloudflare DNS-01` 申请按钮
  - `web/html/xui/cert.html`
- `POST /xui/cert/issueDnsCloudflare`
  - 至少不在前端暴露

更进一步可隐藏：

- `/xui/cloudflare/status`
- `/xui/cloudflare/save`
- `/xui/cloudflare/test`
- `/xui/cloudflare/zones`
- `/xui/cloudflare/records`

## 8. 下一步最小操作建议

最小建议分两步：

1. 立即暂停 Phase 4，不再测试 Cloudflare，不再继续 DNS-01
2. 决策：
   - 若严格对齐 z-ui：回滚 `7d2816c` 和 `5be8c63`
   - 若想先稳住：隐藏 Cloudflare 菜单、页面、DNS-01 按钮和入口

在完成上述收敛前，不建议继续：

- Cloudflare staging 测试
- Cloudflare production 测试
- 其它 DNS provider
- 续期自动化基于 DNS provider 的扩展

## 9. 审计结论

- z-ui 主线证书逻辑：`证书目录 + panel HTTPS 证书路径 + acme.sh HTTP-01 + renew timer`
- z-ui 主线不是 Cloudflare DNS provider 系统
- v-ui 的 Cloudflare API / DNS-01 已偏离“照着 z-ui 做”的当前目标
- 现阶段应暂停 Phase 4
- 推荐回到 HTTP-01 主线，Cloudflare 不应继续扩大
