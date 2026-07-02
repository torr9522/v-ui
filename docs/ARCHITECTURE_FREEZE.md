# v-ui v1.0.1 Stable Architecture Freeze

## 1. 项目定位

- 项目名：`v-ui`
- 运行层：`x-ui`

`v-ui` 作为项目与仓库代号存在，运行层继续保持 `x-ui` 兼容形态。这是主架构决策，不是过渡遗漏。

保持兼容的范围包括：

- `x-ui.service`
- `/usr/local/x-ui`
- `/etc/x-ui`
- `/xui`
- 既有 API 路径
- 既有数据库与 setting key

这样做的原因：

- 降低升级和迁移风险
- 保持既有部署与脚本兼容
- 把研发资源优先放在功能与稳定性，而不是运行层改名

## 2. 设计原则

优先级固定如下：

1. 稳定
2. 兼容
3. 可维护
4. 功能

冻结后必须遵守：

- 不为了新功能破坏已有运行逻辑
- 不为了“看起来更现代”重写稳定模块
- 不在没有审计、快照、验证前直接改主线

## 3. 证书模块冻结

当前主线采用的是：

- `Panel Certificate`

不是：

- `z-ui` 风格 `Certificate Store`
- `/etc/x-ui/certs/<name>` 证书仓
- 多证书资产平台

当前证书模块只围绕“当前活动面板证书”工作，来源限定为：

- `Manual`
- `ACME HTTP-01`

选择这个方案的原因：

- 满足面板 HTTPS 与入站 TLS 复用
- 逻辑简单，状态单一
- 避免把证书模块扩展成独立资产系统

## 4. ACME

官方主线方案固定为：

- `HTTP-01`

不进入主线的方向：

- 多 DNS Provider
- `Cloudflare`
- DNS 平台化
- Provider 扩展系统

原因：

- HTTP-01 已覆盖当前主线需求
- DNS 类能力会显著放大平台复杂度
- 会把项目重心从面板管理偏移到证书平台

## 5. HTTPS

当前最终方案由以下能力组成：

- `Startup Self-Heal`
- `Enable`
- `Disable`
- `Wait Listener`
- `Rollback`

这套方案已经作为面板 HTTPS 主线冻结。后续允许做缺陷修复，不再允许推倒重做。

## 6. Modern UI

`Modern UI` 的设计思路允许借鉴 `3x-ui` 的交互经验，例如：

- 分组表单
- 标签页结构
- 字段联动
- 默认值
- 说明文案

但明确禁止照搬：

- React 架构
- 数据库设计
- API 语义
- 整套后端组织方式

原则是：

- 学习交互，不复制系统
- 改 UI，不改运行主架构

## 7. 协议支持

现阶段主线协议与能力基线如下：

- `VMess`
- `VLESS`
- `Trojan`
- `Shadowsocks`
- `SOCKS`
- `HTTP`
- `Mixed`
- `Tunnel`
- `Dokodemo-door`

当前主线已冻结的是现有兼容与生成方式，不在本轮强行扩展协议矩阵。

说明：

- `TLS`、`XTLS`、现有传输项按当前兼容方案保留
- `Reality`
- `XHTTP`
- `Vision`

以上属于下一阶段能力，不属于 `v1.0.1-stable` 架构冻结范围。

## 8. 未来允许开发

后续允许进入主线规划的方向：

- `REALITY`
- `XHTTP`
- `Routing`
- `Outbound`
- `Runtime API`
- `Subscription`
- `UI`

前提：

- 不破坏 `x-ui` 运行层兼容
- 不绕开审计、快照、真机验证、UAT

## 9. 未来禁止开发

以下方向永久不进入 `v-ui` 主线：

- `Cloudflare DNS` 平台
- 多证书平台
- 多 Provider 系统
- 租户证书系统
- 证书仓
- Node 平台
- Host 平台
- `3x-ui` 整套后台
- 任何脱离 `x-ui` 主架构的大改

判断标准：

- 如果一个方案要求重做运行层、证书层、平台层，它就不属于 `v-ui` 主线

## 10. Release Freeze

`v1.0.1-stable` 正式成为长期开发基线。

后续开发规则：

- 所有新功能必须从 `feature/` 分支开展
- `v-ui` 分支只保留稳定基线与必要合并
- 禁止直接在 `v-ui` 稳定分支上进行日常功能开发

允许直接进入稳定分支的内容只包括：

- Bug Fix
- Security Fix
- Compatibility Fix

其余工作必须先在功能分支完成，再经过验证后合并。
