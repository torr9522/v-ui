# Download Chain Audit

## 1. 审计范围

本次只审计安装依赖、下载入口、更新入口与相关说明文档，不修改任何运行逻辑。

审计目标：

- 确认 `install.sh`、`install_en.sh`、`x-ui.sh`、`x-ui_en.sh`、`server.go`、`README` 不会从 `v-ui` 仓库以外误拉旧文件
- 确认不存在 `c-ui` / `n-ui` / `z-ui` / `3x-ui` / `bin456789` / `MHSanaei` 作为安装依赖
- 确认 release / raw / source tar 下载链路固定在 `torr9522/v-ui`
- 列出仍然保留的外部依赖，并判断是否可接受

## 2. 安装入口

当前公开安装入口为：

```bash
apt update && apt install -y curl
bash <(curl -Ls https://raw.githubusercontent.com/torr9522/v-ui/v-ui/install.sh)
```

安装链路判断结果：

- raw 入口：固定到 `torr9522/v-ui`
- 源码归档：固定到 `https://github.com/torr9522/v-ui/archive/refs/heads/v-ui.tar.gz`
- release fallback：固定到 `https://github.com/torr9522/v-ui/releases/download/v-ui-assets`
- 默认不会拉取 `c-ui` / `n-ui` / `z-ui` / `3x-ui`

结论：

- 主安装入口已经独立

## 3. 脚本下载链路

| 文件 | 下载目标 | 地址 | 分类 | 结论 |
|---|---|---|---|---|
| `install.sh` | 源码归档 | `https://github.com/torr9522/v-ui/archive/refs/heads/v-ui.tar.gz` | 主仓库 | 正常 |
| `install.sh` | release asset fallback | `https://github.com/torr9522/v-ui/releases/download/v-ui-assets/x-ui-linux-<arch>.tar.gz` | 主仓库 | 正常 |
| `install.sh` | Xray/geo fallback zip | `https://github.com/torr9522/v-ui/releases/download/v-ui-assets/xray-linux-<arch>.zip` | 主仓库 | 正常 |
| `install.sh` | 管理脚本 fallback | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/x-ui.sh` | 主仓库 | 正常 |
| `install.sh` | 端口限制同步脚本 | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/xui-portlimit-sync.sh` 等 | 主仓库 | 正常 |
| `install.sh` | 证书续期 timer/service | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/x-ui-cert-renew.service` 等 | 主仓库 | 正常 |
| `install.sh` | Go toolchain | `https://go.dev/dl/go1.22.7.linux-<arch>.tar.gz` | 官方依赖 | 可接受 |
| `install.sh` | 公网 IP 探测 | `https://api64.ipify.org` / `https://ipv4.icanhazip.com` / `https://ifconfig.me/ip` | 公共服务 | 可接受 |
| `install_en.sh` | 委托中文安装器 | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/install.sh` | 主仓库 | 正常 |
| `x-ui.sh` | 重装/更新入口 | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/install.sh` | 主仓库 | 正常 |
| `x-ui.sh` | BBR 脚本 | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/scripts/bbr.sh` | 主仓库 | 正常 |
| `x-ui.sh` | acme 安装脚本 | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/scripts/acme_install.sh` | 主仓库 | 正常 |
| `x-ui.sh` | geo 数据更新 | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/bin/geoip.dat` / `geosite.dat` | 主仓库 | 正常 |
| `x-ui.sh` | shell 自更新 | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/x-ui.sh` | 主仓库 | 正常 |
| `x-ui_en.sh` | 重装/更新入口 | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/install_en.sh` | 主仓库 | 正常 |
| `x-ui_en.sh` | shell 自更新 | `https://raw.githubusercontent.com/torr9522/v-ui/v-ui/x-ui_en.sh` | 主仓库 | 正常 |
| `x-ui_en.sh` | acme / BBR / geo | 同 `x-ui.sh`，均指向 `torr9522/v-ui` | 主仓库 | 正常 |
| `web/service/server.go` | Xray 更新 zip | `https://github.com/torr9522/v-ui/releases/download/v-ui-assets/xray-linux-<arch>.zip` | 主仓库 | 正常 |
| `scripts/acme_install.sh` | acme.sh 官方安装脚本 | `https://raw.githubusercontent.com/acmesh-official/acme.sh/$BRANCH/acme.sh` | 官方依赖 | 可接受 |
| `scripts/bbr.sh` | Ubuntu kernel PPA / RPM kernel 下载 | `https://kernel.ubuntu.com/...` 等 | 外部依赖 | 可接受，且非默认安装链路 |
| `README.md` / `README_EN.md` | 安装说明 | `raw.githubusercontent.com/torr9522/v-ui/v-ui/install.sh` | 主仓库 | 正常 |
| `x-ui.service` / `x-ui-cert-renew.*` | 无下载行为 | 无 | 本地文件 | 正常 |

## 4. 是否存在旧项目残留

### 4.1 安装与运行链路

结论：

- 未发现 `c-ui`
- 未发现 `n-ui`
- 未发现 `z-ui`
- 未发现 `3x-ui`
- 未发现 `bin456789`
- 未发现 `MHSanaei/3x-ui`

作为实际安装依赖、更新依赖或运行时下载目标。

也未发现以下风险：

- 误拉 `c-ui` source tar
- 误拉 `n-ui` raw 脚本
- 误拉旧 release asset
- 误拉 `c-ui-source-c-ui`
- 误拉 `vc-ui-local`

### 4.2 文档残留

仍存在旧项目文字残留，但不属于实际下载链路：

- `docs/UPSTREAM_ASSETS.md`
  - 仍写着 `torr9522/c-ui`
  - 仍写着 `c-ui-assets`
- `docs/ARCHITECTURE.md`
  - 仍写着 raw 安装从 `torr9522/c-ui` 拉源码归档
- `CHANGELOG.md`
  - 有历史 `c-ui` 切仓记录
- 多份审计文档中存在 `z-ui` / `3x-ui` / `n-ui` 作为来源或对比说明

结论：

- 这些不是运行阻断
- 但其中 `docs/UPSTREAM_ASSETS.md` 和 `docs/ARCHITECTURE.md` 已经与当前真实安装链路不一致

## 5. 是否存在拉取旧文件风险

### 5.1 branch / raw / source tar

- branch：固定 `v-ui`
- raw：固定 `torr9522/v-ui/v-ui`
- source archive：固定 `torr9522/v-ui/archive/refs/heads/v-ui.tar.gz`

结论：

- 不存在 branch 漂到旧仓库的风险

### 5.2 release / assets

- `install.sh` 的 fallback release 固定为 `v-ui-assets`
- `server.go` 的 Xray 更新 fallback 也固定为 `v-ui-assets`
- 未发现 `c-ui-assets`
- 未发现旧项目 release asset 名称

结论：

- 不存在误拉旧 release asset 的风险

### 5.3 仓库内置 bin 资产

`install.sh` 的行为是：

1. 若源码目录中已有 `bin/xray-linux-<arch>`、`bin/geoip.dat`、`bin/geosite.dat`
2. 优先直接使用仓库内置资产
3. 只有缺失时，才回退到 `v-ui-assets/xray-linux-<arch>.zip`

结论：

- 默认优先使用当前仓库资产
- 缺失时也只会从 `torr9522/v-ui` 自己的 release 补齐

## 6. 可接受外链

以下外链可以保留：

- Go 官方
  - `https://go.dev/dl/...`
  - 用于源码安装模式自动安装 Go
- acme.sh 官方
  - `https://raw.githubusercontent.com/acmesh-official/acme.sh/...`
  - 这是证书脚本官方来源
- OS package manager
  - `apt-get` / `yum` / `dnf`
  - 属于系统依赖安装
- 公网 IP 探测服务
  - `api64.ipify.org`
  - `ipv4.icanhazip.com`
  - `ifconfig.me/ip`
  - 仅用于安装完成后显示公网面板地址
- BBR 外部脚本链路
  - Ubuntu kernel PPA / kernel 下载源
  - 仅在用户手动运行 `x-ui bbr` 时触发

## 7. 分类结果

### A. 必须改：旧项目残留

未发现真实安装链路中的 A 类问题。

结论：

- `P0 = none`

### B. 可保留：官方依赖

- `go.dev`
- `acmesh-official/acme.sh`
- OS 包管理器
- 公网 IP 探测服务
- BBR 依赖的系统/内核下载源

### C. 仅文档来源说明

- `README` / `README_EN` 的致谢中 `vaxilu/x-ui`、`XTLS/Xray-core`
- 各类 `z-ui` / `3x-ui` 审计文档
- `BACKUP_LOG` 中历史 `n-ui` 记录

这些不构成安装链路依赖。

### D. 需要确认

- `docs/UPSTREAM_ASSETS.md`
- `docs/ARCHITECTURE.md`

这两份文档不是安装入口，但内容已经与当前主线事实不一致，后续应单独修正文档。

## 8. 其他发现

### 8.1 中英文脚本不一致

`x-ui_en.sh` 与 `x-ui.sh` 并不完全一致：

- `x-ui_en.sh` 仍保留更旧的证书申请菜单
- 仍暴露 `Cloudflare DNS API mode` 的旧交互说明
- 未同步当前中文脚本里的最小 `cert-renew` / `cert-renew-status` 体系

结论：

- 这是脚本一致性问题
- 不是当前公开安装下载链路阻断
- 但若后续继续维护英文脚本，应单独审计与同步

### 8.2 文档与事实不一致

以下文档已不反映当前事实：

- `docs/UPSTREAM_ASSETS.md`
- `docs/ARCHITECTURE.md`

它们仍使用 `c-ui` 视角描述仓库与安装链路。

结论：

- 这是文档问题，不是运行问题

## 9. 阻断项

### P0

- 无

### P1

- 无真实安装链路 P1

### P2

- `docs/UPSTREAM_ASSETS.md` 仍写 `torr9522/c-ui` / `c-ui-assets`
- `docs/ARCHITECTURE.md` 仍写 raw 安装从 `torr9522/c-ui` 拉源码归档
- `x-ui_en.sh` 与 `x-ui.sh` 在证书相关菜单与交互说明上不一致

## 10. 结论

结论如下：

- 可以认定当前安装链路已经独立
- 未发现会误拉旧项目文件的真实下载路径
- 未发现 `c-ui` / `n-ui` / `z-ui` / `3x-ui` / `bin456789` / `MHSanaei` 作为安装依赖
- 可以继续进入“协议老旧参数清理 + 新参数显示”

但建议先记住两点：

1. 运行链路已独立，问题主要在旧文档与英文脚本一致性
2. 如果下一轮要碰用户可见文档，优先修正 `docs/UPSTREAM_ASSETS.md` 与 `docs/ARCHITECTURE.md`
