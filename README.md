# v-ui Project
简体中文 | [ENGLISH](./README_EN.md)

`v-ui` 是当前开发代号；实际运行层、服务名、命令名、安装路径和面板显示名继续保持 `x-ui`。当前代码基线包含出站管理、路由管理、AI 分流模板，以及基于 `managedOutboundsRouting` 的托管 Xray 配置能力。

## Current Stable

- `v1.0`
- Status: `Ready For Production`
- Status: `Ready For Node TLS`

## 核心功能

- 出站管理：只读列表、CRUD、系统项保护、被路由引用时禁止删除
- 路由管理：只读列表、CRUD、排序、启停控制
- AI 分流模板：一键生成 OpenAI、Claude、Gemini、Grok、Perplexity、Poe、Cursor、GitHub Copilot、HuggingFace 规则
- 托管配置：`managedOutboundsRouting` 开关控制是否由 `x-ui` 接管 `outbounds` 和 `routing.rules`
- 安装自举：源码安装模式下，如果仓库中没有预编译 `x-ui`，`install.sh` 会自动安装 Go、gcc、git、tar、curl、unzip、file，并使用 `CGO_ENABLED=1` 构建；但 raw `curl` 安装入口本身要求宿主机先具备 `curl`

## Runtime Freeze

为了保持与既有部署兼容，运行层名称保持不变：

- 二进制仍然叫 `x-ui`
- systemd 服务仍然叫 `x-ui.service`
- 安装目录仍然是 `/usr/local/x-ui`
- Web 路由前缀仍然是 `/xui`

这是一项兼容性决策，不是命名遗漏。详见 [docs/ADR-001-runtime-freeze.md](./docs/ADR-001-runtime-freeze.md)。

补充说明：

- 项目目录、Git 分支、文档中的项目代号可以是 `v-ui`
- 面板标题、侧边栏显示、运行层名称一律保持 `x-ui`

## 安装

纯净 Debian / Ubuntu 机器请先安装 `curl`：

```bash
apt update && apt install -y curl
```

正式安装命令：

```bash
bash <(curl -Ls https://raw.githubusercontent.com/torr9522/v-ui/v-ui/install.sh)
```

English installer:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/torr9522/v-ui/v-ui/install_en.sh)
```

说明：

- 如果你是直接 `git clone` 本仓库再执行 `install.sh`，脚本会优先使用本地源码和仓库内自带的 `bin/` 资源
- 如果你是通过上面的 raw 命令直接安装，脚本会从 `torr9522/v-ui` 拉取同仓库源码归档并在目标机器上自举构建
- raw `curl` 安装入口不是“零依赖一键启动”，它要求宿主机预先安装 `curl`
- 不再依赖旧 `n-ui` 仓库的 raw 链接

## 仓库内置资产

本仓库已经纳入运行和二开所需的关键资产：

- `bin/xray-linux-amd64`
- `bin/xray-linux-arm64`
- `bin/geoip.dat`
- `bin/geosite.dat`
- `x-ui.service`
- `x-ui.sh`
- `x-ui_en.sh`
- `xui-portlimit-sync.sh`
- `xui-portlimit-sync.service`
- `xui-portlimit-sync.timer`
- `scripts/bbr.sh`
- `scripts/acme_install.sh`

资产清单、大小、保留外链与原因见 [docs/UPSTREAM_ASSETS.md](./docs/UPSTREAM_ASSETS.md)。

## 验证状态

已完成两台 Debian 11 测试机真机验证：

- `5.226.49.140`
- `43.198.88.130`

验证范围包括：

- 源码安装自举
- `managedOutboundsRouting` 开关
- Outbounds CRUD
- Routing CRUD
- AI 分流模板去重
- `ValidateManagedXrayConfig`
- `RestartXray`

详细记录见：

- [docs/LOCAL_FEATURE_MATRIX.md](./docs/LOCAL_FEATURE_MATRIX.md)
- [docs/LOCAL_RUNTIME_MATRIX.md](./docs/LOCAL_RUNTIME_MATRIX.md)

## 开发文档

- [BASELINE.md](./BASELINE.md)
- [docs/DEVELOPMENT_PLAN.md](./docs/DEVELOPMENT_PLAN.md)
- [docs/ADR-001-runtime-freeze.md](./docs/ADR-001-runtime-freeze.md)
- [docs/UPSTREAM_ASSETS.md](./docs/UPSTREAM_ASSETS.md)

## 致谢

- [vaxilu/x-ui](https://github.com/vaxilu/x-ui)
- [XTLS/Xray-core](https://github.com/XTLS/Xray-core)
- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)

## Stargazers

[![Stargazers over time](https://starchart.cc/torr9522/c-ui.svg)](https://starchart.cc/torr9522/c-ui)
