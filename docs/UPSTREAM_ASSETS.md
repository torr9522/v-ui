# Upstream Assets

## 目标

`c-ui` 现在以独立仓库形式维护，安装脚本、管理脚本和运行必需资产不再依赖旧 `n-ui` 仓库的 raw 链接。

## 已纳入仓库的关键资产

| 文件 | 大小 | 原因 |
| --- | ---: | --- |
| `bin/xray-linux-amd64` | `47,076,583` bytes | amd64 运行所需的 Xray 主程序 |
| `bin/xray-linux-arm64` | `34,078,846` bytes | arm64 运行所需的 Xray 主程序 |
| `bin/geoip.dat` | `18,663,777` bytes | Xray 路由与地理规则所需数据 |
| `bin/geosite.dat` | `10,641,174` bytes | Xray 路由与地理规则所需数据 |
| `bin/config.json` | `1,187` bytes | 仓库内默认运行配置基线 |
| `x-ui.service` | tracked | 安装与 systemd 启动所需 |
| `x-ui.sh` / `x-ui_en.sh` | tracked | 面板管理入口脚本 |
| `install.sh` / `install_en.sh` | tracked | 安装入口脚本 |
| `xui-portlimit-sync.sh` | tracked | 端口限制同步脚本 |
| `xui-portlimit-sync.service` | tracked | 端口限制同步服务单元 |
| `xui-portlimit-sync.timer` | tracked | 端口限制同步定时器 |
| `scripts/bbr.sh` | tracked | 管理脚本内调用的辅助脚本 |
| `scripts/acme_install.sh` | tracked | 管理脚本内调用的辅助脚本 |

## 安装与运行方式调整

- `install.sh` 默认 raw 地址已改为 `torr9522/c-ui`
- 直接执行 raw 安装命令时，脚本会从 `torr9522/c-ui` 下载源码归档并在目标机器自举编译
- 如果用户先 clone 仓库再执行安装，脚本会直接使用仓库内的源码和 `bin/` 资产
- `sync_default_xray_assets()` 会优先使用仓库已携带的 `bin/` 资产，不再强制依赖外部 release 包

## 仍保留的外部链接

以下链接仍然保留，原因是它们属于官方或系统级依赖，而不是旧仓库依赖：

- `https://go.dev/dl/...`
  - Go 官方工具链下载地址，用于源码安装模式自动安装 Go
- `apt-get` / `yum` / `dnf`
  - 系统包管理器，用于安装 gcc、git、curl、tar、unzip、file 等基础依赖
- `https://api64.ipify.org`、`https://ipv4.icanhazip.com`、`https://ifconfig.me/ip`
  - 安装完成后显示公网面板地址时使用的公网 IP 探测服务
- `https://github.com/torr9522/c-ui/releases/download/c-ui-assets/...`
  - 作为 release 资产回退地址保留在安装脚本与版本更新逻辑中，但不再依赖旧仓库
- `https://github.com/XTLS/Xray-core`
  - 仅作为致谢与依赖来源说明

## 不纳入仓库的内容

以下内容不会进入 `c-ui` 仓库：

- `.git.n-ui-backup/`
- 任意 `*.db`、`*.db-shm`、`*.db-wal`
- 任意 `*.log`
- 服务器备份包
- 测试报告压缩包
- 根目录临时构建出的 `x-ui`
- 根目录重复的 `xray-linux-*`

## 结论

当前仓库已经完成对旧 raw 依赖的切断。后续用户只需要拿 `torr9522/c-ui` 这个仓库，就可以继续安装、运行和二次开发。
