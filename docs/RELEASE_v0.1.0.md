# C-UI v0.1.0 Stable

## 项目定位

- 项目名称：`C-UI`
- Runtime 兼容：`x-ui`
- 仓库状态：独立仓库 `torr9522/c-ui`

## 本版本新增

- ✓ Outbounds
- ✓ Routing
- ✓ AI Traffic Split
- ✓ managedOutboundsRouting
- ✓ Source Bootstrap Install
- ✓ CGO Build
- ✓ 自动安装：
  - Go
  - gcc
  - git
  - curl
  - tar
  - unzip
  - file

## 已验证

两台 Debian 11 VPS 已完成完整运行矩阵并通过：

- `5.226.49.140`
- `43.198.88.130`

验证覆盖：

- Outbound CRUD
- Routing CRUD
- AI Template 去重
- managedOutboundsRouting 开关
- ValidateManagedXrayConfig
- RestartXray
- Source Bootstrap Install

## Runtime Freeze

继续保持以下兼容面不变：

- `x-ui.service`
- `/usr/local/x-ui`
- `/usr/bin/x-ui`
- `/xui`

这是兼容性设计，不是遗留错误。

## 后续路线图

`v0.2` 目标：

- Runtime API
- Hot Reload
- Observatory
- Route Test
- AI Rule Library
- 更多 Outbound 类型
