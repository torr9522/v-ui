# Architecture

## Overview

`C-UI` 维持 `x-ui` Runtime 兼容层，同时在数据库、服务层、控制器层和前端页面上增加 Outbound、Routing、AI Template 与托管配置能力。

## Database

### Outbound

表模型：`database/model/model.go`

关键字段：

- `tag`
- `protocol`
- `settings`
- `streamSettings`
- `mux`
- `enabled`
- `remark`
- `sort`
- `isSystem`

系统初始化：

- `direct`
- `blocked`

规则：

- 系统项只初始化，不覆盖用户数据
- `direct` / `blocked` 受保护

### RoutingRule

表模型：`database/model/model.go`

关键字段：

- `type`
- `domain`
- `ip`
- `port`
- `protocol`
- `network`
- `source`
- `inboundTag`
- `outboundTag`
- `enabled`
- `remark`
- `sort`
- `isSystem`

规则：

- `outboundTag` 必须指向现有 outbound
- 禁用规则不会进入托管配置

## Services

### OutboundService

位置：`web/service/outbound.go`

能力：

- 查询列表
- CRUD
- 系统项保护
- Routing 引用计数保护
- 保存后按 `managedOutboundsRouting` 决定是否执行配置校验与重启

### RoutingService

位置：`web/service/routing.go`

能力：

- 查询列表
- CRUD
- 排序
- 规则校验
- AI Template 批量生成
- 保存后按 `managedOutboundsRouting` 决定是否执行配置校验与重启

### XrayService

位置：`web/service/xray.go`

能力：

- `GetXrayConfig()`
- 旧模板模式与托管模式切换
- 托管模式下构造 `outbounds` 与 `routing.rules`
- `ValidateManagedXrayConfig()`
- `ApplyManagedXrayIfEnabled()`

## Managed Config

开关：`managedOutboundsRouting`

行为：

- `false`
  - 完全沿用旧模板模式
- `true`
  - 使用旧模板作 base
  - 接管 `outbounds`
  - 接管 `routing.rules`
  - 保留既有 inbounds 逻辑

顺序约束：

- `direct` 必须第一个 outbound
- `blocked` 必须第二个 outbound
- `api -> api` 必须第一条 routing rule
- 用户 Outbound / Routing 按 `sort asc, id asc`

安全约束：

- `TestConfig` 失败时回滚 DB 修改
- 不写入不可运行的托管配置

## AI Template

位置：`web/service/routing.go`

能力：

- 根据用户选择的 `outboundTag` 批量创建 AI 平台规则
- 规则指向：
  - OpenAI
  - Anthropic
  - Google AI
  - xAI
  - Perplexity
  - Poe
  - Cursor
  - GitHub Copilot
  - HuggingFace

去重：

- 使用 routing fingerprint 避免重复创建

## Install Bootstrap

位置：`install.sh`

源码安装模式触发条件：

- 根目录不存在预编译 `./x-ui`

bootstrap 能力：

- 自动安装 Go
- 自动安装 gcc / build-essential
- 自动安装 git
- 自动安装 curl
- 自动安装 tar
- 自动安装 unzip
- 自动安装 file

构建方式：

- `CGO_ENABLED=1 go build -o x-ui .`

资产策略：

- 优先使用仓库内 `bin/` 资产
- raw 安装时从 `torr9522/c-ui` 拉取源码归档

## Request Flow

1. 前端调用 `/xui/outbound/*` 或 `/xui/routing/*`
2. controller 绑定请求并调用 service
3. service 校验数据并执行 DB 变更
4. 若启用 `managedOutboundsRouting`
5. 基于事务后数据生成托管配置
6. `ValidateManagedXrayConfig()`
7. 成功后 `RestartXray`
8. 失败则回滚 DB

## Baseline

`v0.1.0-stable` 是当前后续所有开发的稳定基线。
