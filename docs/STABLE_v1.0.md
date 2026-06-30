# v-ui v1.0 Stable

## 一、版本

- `v-ui v1.0 Stable`

## 二、最终功能

- ✓ 入站管理
- ✓ 出站管理
- ✓ Routing
- ✓ AI Routing Template
- ✓ Managed Routing
- ✓ HTTPS Panel
- ✓ 手动证书
- ✓ HTTP-01
- ✓ ACME Production
- ✓ Fullchain
- ✓ Startup Self Heal
- ✓ Renew Timer

## 三、真机验证

以下服务器已完成当前稳定主线验证：

- `5.226.49.140`：`PASS`
- `43.198.88.130`：`PASS`
- `139.180.135.210`：`PASS`

说明：

- `5.226.49.140` 与 `43.198.88.130` 完成主功能与安装主线验证
- `139.180.135.210` 完成证书、HTTPS、ACME、续期与节点 TLS 主线验证

## 四、真人 UAT

`PASS`

覆盖范围：

- 安装
- 登录
- 申请证书
- 新增 TLS 节点
- 客户端连接
- 修改
- 删除

说明：

- `https://IP:panel_port` 的证书域名不匹配不计入本轮失败
- 本轮 UAT 核心目标是域名证书驱动的节点 TLS 可用性
- 最终 UAT 结论为 `Ready For Node TLS Usage`

## 五、最终状态

- `Ready For Production`
- `Ready For Node TLS`

## 六、Known Issues

仅保留：

- `Issue001`
- `IssueHttp Idempotent`
- Priority: `P2`

## 七、冻结原则

允许：

- Bug Fix
- Security Fix
- Compatibility Fix

禁止：

- 新增证书功能
- 新增 DNS Provider
- 新增 ACME Provider
- 新增复杂证书系统

## 结论

`v-ui v1.0 Stable` 进入冻结状态。

后续只接受稳定性、安全性、兼容性修复；主线工作应回到 `v-ui` 核心产品能力，而不是继续扩展证书系统。
