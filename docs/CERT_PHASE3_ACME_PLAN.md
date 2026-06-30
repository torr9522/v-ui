# Certificate Phase 3 ACME Plan

## 1. acme.sh 还是 certbot

推荐选择：`acme.sh`

选择理由：

1. 仓库里已经存在 ACME 相关历史资产：
   - [x-ui.sh](/tmp/v-ui-init/v-ui/x-ui.sh)
   - [scripts/acme_install.sh](/tmp/v-ui-init/v-ui/scripts/acme_install.sh)
2. 现有项目历史更接近 shell-first ACME 集成，而不是 system package + certbot 风格。
3. `acme.sh` 更轻量：
   - 单脚本安装
   - 不强依赖 Python / snap
   - 更容易在当前 x-ui 风格安装环境中集成
4. 后续如果扩展 DNS-01、Cloudflare、renew hook，`acme.sh` 的参数和 hook 体系更贴近当前项目需求。
5. 当前 Phase 3 第一轮只做 HTTP-01，`acme.sh --issue --standalone` 足够直接。

不选 `certbot` 的主要原因：

1. 更重。
2. 对系统环境依赖更多。
3. 与当前仓库已有 `acme_install.sh` 方向不一致。
4. 后续做多 provider 扩展时，不如 `acme.sh` 灵活。

本阶段结论：

- Phase 3 使用 `acme.sh`
- CA 只支持 Let’s Encrypt
- 只设计 HTTP-01

## 2. HTTP-01 申请流程

推荐流程：

1. 用户在“域名证书”页面先设置 `webDomain`
2. 前端调用 `checkDomain`
3. 后端再次校验：
   - domain 非空
   - domain 格式合法
   - domain 解析命中当前服务器公网 IP
4. 检查 `acme.sh` 是否已安装
   - 未安装则自动安装
5. 检查 80 端口是否可用
6. 若 80 可用：
   - 使用 `acme.sh --set-default-ca --server letsencrypt`
   - 使用 `acme.sh --issue -d <domain> --standalone --httpport 80`
7. 申请成功后，执行 `acme.sh --install-cert`
8. 将证书安装到 v-ui 管理路径
9. 更新 setting 元数据
10. 可选地复用 Phase 2 的 `enableHttps`
11. 返回：
   - 申请结果
   - cert status
   - 是否已应用 HTTPS

推荐的第一轮产品行为：

- “申请证书”与“启用 HTTPS”拆成两个明确动作
- 但后端保留一个 `issueHttpAndEnable` 的组合能力设计空间

Phase 3 第一轮编码建议优先做：

- 先申请并落盘
- 再显式调用启用 HTTPS

这样更稳，回滚也更清晰。

## 3. 80 端口占用处理

HTTP-01 的核心风险就是 80 端口。

推荐设计：

### 3.1 前置检查

在执行 `acme.sh --issue --standalone --httpport 80` 之前：

- 检查 80 端口是否被占用
- 检查占用进程是什么

建议后端返回：

- `port80Available`
- `port80Owner`
- `canProceed`

### 3.2 Phase 3 第一轮策略

第一轮不要自动停服务，不要自动抢端口。

原因：

1. 风险太高
2. 可能影响用户已有网站/反代
3. 一旦失败，回滚复杂

第一轮建议行为：

- 如果 80 端口被占用，直接返回失败
- 明确提示：
  - 80 端口被谁占用
  - 需要临时释放 80 再重试

### 3.3 后续可扩展策略

后续版本再考虑：

- 临时停本面板或停反代再申请
- webroot 模式
- 反代 challenge 转发

但 Phase 3 第一轮不做。

## 4. 申请成功后的证书落盘路径

推荐统一落盘到受 v-ui 管理的固定路径：

- `/usr/local/x-ui/cert/acme-panel.crt`
- `/usr/local/x-ui/cert/acme-panel.key`
- 可选：
  - `/usr/local/x-ui/cert/acme-panel.ca.cer`
  - `/usr/local/x-ui/cert/acme-panel.fullchain.cer`

理由：

1. 不直接依赖 `~/.acme.sh/...` 运行态路径
2. 和 Phase 2 手动证书路径体系保持一致
3. 后续 `GetStatus()`、启用 HTTPS、备份逻辑都更统一
4. 权限控制更直接

权限建议：

- cert: `0644`
- key: `0600`

## 5. 如何写入 webCertFile / webKeyFile

推荐方式：

1. ACME 申请成功
2. `acme.sh --install-cert`
3. 安装到：
   - `/usr/local/x-ui/cert/acme-panel.crt`
   - `/usr/local/x-ui/cert/acme-panel.key`
4. 然后写 setting：
   - `webCertFile=/usr/local/x-ui/cert/acme-panel.crt`
   - `webKeyFile=/usr/local/x-ui/cert/acme-panel.key`
   - `webCertStatus=issued` 或 `uploaded` 的 ACME 对应状态
   - `webCertMode=acme`
   - `webCertProvider=letsencrypt`
   - `webCertIssuer=<issuer>`
   - `webCertExpireAt=<notAfter>`
   - `webCertAutoRenew=true/false` 先只设计，不一定第一轮启用

建议新增但不一定第一轮全部实现的 setting：

- `webCertManagedBy=acme`
- `webCertAcmeDomain=<domain>`
- `webCertRenewHookEnabled=true`

## 6. 如何复用 Phase 2 enableHttps

推荐完全复用 Phase 2 的 `enableHttps`

原因：

1. Phase 2 已经完成：
   - listener wait
   - HTTP/HTTPS 切换
   - rollback
   - frontend 协议跳转
2. 不应在 Phase 3 再复制一套“启用 HTTPS”逻辑
3. ACME 的职责应只到：
   - 申请证书
   - 落盘
   - 写 metadata / active cert path

推荐调用链：

`issueHttp`

↓

acme.sh issue/install-cert

↓

写 cert file / key file / ACME metadata

↓

调用 Phase 2 `EnableHTTPS()`

↓

返回最终状态

## 7. 失败回滚方案

必须区分三个阶段：

### 7.1 申请前失败

例如：

- domain 校验失败
- 80 端口占用
- acme.sh 安装失败

处理：

- 不写 setting
- 不动现有 panel cert
- 直接返回错误

### 7.2 申请中失败

例如：

- Let’s Encrypt challenge 失败
- `acme.sh --issue` 失败
- `install-cert` 失败

处理：

- 不覆盖当前 active `webCertFile/webKeyFile`
- 清理本次临时申请产生的无效目标文件
- 保留日志供排查

### 7.3 申请成功但启用 HTTPS 失败

这是最重要的回滚点。

处理：

1. 保留新申请成功的 cert/key 文件
2. 记录 ACME 结果
3. 调用 Phase 2 `EnableHTTPS()`
4. 如果 `EnableHTTPS()` 失败：
   - 依赖其现有 rollback 机制恢复旧 active setting
   - 返回：
     - certificate issued
     - https not applied

这样用户至少得到一个“证书已经申请成功，但面板切换失败”的明确状态，而不是全部丢失。

## 8. renew hook 设计

本轮不编码自动续期，但设计必须先定下来。

推荐设计：

### 8.1 acme.sh renew hook

后续使用 `acme.sh --install-cert` 时带 reload/restart hook。

建议 hook 行为：

1. 证书续期成功
2. 覆盖：
   - `/usr/local/x-ui/cert/acme-panel.crt`
   - `/usr/local/x-ui/cert/acme-panel.key`
3. 重新解析证书 metadata
4. 如果当前 active cert 正是这对 ACME 证书：
   - 调用 panel reload / `EnableHTTPS` 等价刷新路径
5. 写入 renew log

### 8.2 Phase 3 第一轮只做设计，不实现 timer

后续可以新增：

- `x-ui-cert-renew.service`
- `x-ui-cert-renew.timer`

但本轮不编码。

## 9. 安全风险

### 9.1 80 端口暴露风险

HTTP-01 要求 80 端口可达。

风险：

- 用户误以为只开面板端口即可
- 80 端口被其它服务占用

应对：

- 强制前置检查
- 明确告知占用进程和失败原因

### 9.2 证书覆盖风险

风险：

- 申请成功后覆盖当前可用 cert
- 若启用失败可能锁死面板

应对：

- 先写到 ACME 独立文件路径
- 再复用 Phase 2 `EnableHTTPS()`
- 失败靠 Phase 2 rollback 恢复旧 active setting

### 9.3 私钥权限风险

风险：

- key 文件权限过宽

应对：

- `0600`
- 仅 root 可读

### 9.4 日志泄漏风险

风险：

- `acme.sh` 日志可能包含敏感上下文

应对：

- 只记录必要摘要
- 避免把完整环境变量和 token（后续 DNS-01）直接写入 UI 日志

### 9.5 误操作风险

风险：

- 用户在错误 domain 或未解析时直接申请

应对：

- 强制先做 `checkDomain`
- 后端再次校验 matched

## 10. 第一轮编码建议

Phase 3 第一轮编码建议只做最小闭环：

1. setting 扩展
   - 最少补 ACME 模式所需 metadata
2. service 层新增 ACME HTTP-01 申请能力
   - 安装 `acme.sh`
   - issue
   - install-cert
3. controller 新增 HTTP-01 issue/status 接口
4. cert 页面新增：
   - “申请 Let’s Encrypt（HTTP-01）”按钮
   - 80 端口检查结果展示
   - 申请日志摘要展示
5. 成功后复用 Phase 2 `enableHttps`

第一轮不要做：

- Cloudflare DNS-01
- ZeroSSL
- 自动续期 timer
- renew hook 真正执行逻辑
- 多 provider
- 自动停占用 80 的服务

第一轮编码目标：

- 让一个满足条件的纯净 VPS 可以通过 HTTP-01 成功申请 Let’s Encrypt 证书
- 并安全切到 HTTPS
