# Changelog v1.0

## v1.0 Stable Freeze

当前稳定版本：

- `v-ui v1.0 Stable`
- Runtime 兼容层：`x-ui`

## 开发过程

### 1. 基线导入

- `1daa767` Import v-ui source baseline

建立 `v-ui` 仓库基线，保留 `x-ui` 运行层兼容面。

### 2. 证书能力起步

- `2ce3944` Add certificate management skeleton
- `e9c7354` Add manual certificate upload and HTTPS toggle

建立域名/证书页、手动证书上传、HTTPS 启用/禁用基础链路。

### 3. Runtime 兼容冻结

- `08955c1` Restore x-ui branding for v-ui runtime

明确保持：

- `x-ui.service`
- `/usr/local/x-ui`
- `/usr/bin/x-ui`
- `/xui`

### 4. HTTPS Toggle 修复

- `0211ea6` Wait for panel listener after HTTPS toggle
- `67320c0` Fix HTTPS disable readiness detection
- `364f4e2` Document certificate phase 2 completion

修复 HTTPS 切换等待时机、HTTP ready 误判与前端协议切换问题，完成 Phase 2。

### 5. ACME HTTP-01 主线

- `22304a6` Plan certificate phase 3 ACME HTTP-01
- `c879fd7` Add ACME HTTP-01 certificate issue
- `1bc444c` Fix acme.sh executable discovery
- `35c234f` Use ACME fullchain for panel certificate

完成：

- ACME HTTP-01 骨架
- `acme.sh` 路径发现
- Production / Staging 签发
- Fullchain 落盘与面板 HTTPS 使用

### 6. 启动健壮性

- `79e9828` Add HTTPS startup self-heal

当数据库仍保留 HTTPS 设置但证书文件已缺失时，自动回退 HTTP，避免启动失败。

### 7. 偏航与回归主线

- `7d2816c` Add Cloudflare API integration
- `5be8c63` Add Cloudflare DNS-01 certificate issue
- `1b83d73` Revert "Add Cloudflare API integration"
- `878ef88` Revert "Add Cloudflare DNS-01 certificate issue"
- `386e447` Document z-ui and v-ui certificate alignment

短暂探索 Cloudflare / DNS-01 后，明确其偏离 `z-ui` 主线，已完整回滚，主线重新聚焦：

- 手动证书
- HTTPS enable / disable
- HTTP-01
- Fullchain
- Startup Self-Heal
- Renew Timer

### 8. 续期链路

- `1c95a9b` Add certificate renewal timer

增加：

- `x-ui-cert-renew.service`
- `x-ui-cert-renew.timer`
- `x-ui cert-renew`
- 手动证书自动跳过

### 9. 证书模块冻结

- `f3a064e` Freeze certificate module

将证书模块收口，唯一保留问题为 `IssueHttp Idempotent`。

### 10. 安装文档修正

- `a42fe78` Document curl prerequisite for installer

补充 raw `curl` 安装入口前提：宿主机必须预装 `curl`。

## v1.0 最终能力

- 入站管理
- 出站管理
- Routing
- AI Routing Template
- Managed Routing
- HTTPS Panel
- 手动证书
- HTTP-01
- ACME Production / Staging
- Fullchain
- Startup Self-Heal
- Renew Timer

## 真机与 UAT 结果

- `5.226.49.140`：`PASS`
- `43.198.88.130`：`PASS`
- `139.180.135.210`：`PASS`

真人 UAT：

- 安装：`PASS`
- 登录：`PASS`
- 申请证书：`PASS`
- 新增 TLS 节点：`PASS`
- 客户端连接：`PASS`
- 修改：`PASS`
- 删除：`PASS`

## 已知问题

- `Issue001` `IssueHttp Idempotent` `P2`

## 稳定版结论

`v-ui v1.0` 正式冻结。

后续进入维护模式：

- 允许 Bug Fix
- 允许 Security Fix
- 允许 Compatibility Fix
- 不再接受证书系统扩展

下一阶段应回到 `v-ui` 主功能开发。
