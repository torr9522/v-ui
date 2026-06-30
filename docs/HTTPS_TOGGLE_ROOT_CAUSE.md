# HTTPS Toggle Root Cause

## 1. 外部资料搜索摘要

本轮重点看了几类资料：

- Go `net/http` 的 `Server.Shutdown` 语义
- Go `crypto/tls` 的 `tls.NewListener` 语义
- `307 Temporary Redirect` 的请求方法保留语义
- Go `http.Client` 的 `CheckRedirect` / `ErrUseLastResponse` 语义

结论摘要：

1. `Server.Shutdown` 的职责是优雅关闭服务端并停止接收新连接，正确使用时不需要依赖 `systemd restart` 才能切换 HTTP/TLS。
2. `tls.NewListener` 只是把一个普通 `net.Listener` 包装成 TLS listener；是否启用 TLS，取决于当前这次启动时是否给 listener 包了一层 TLS。
3. `307 Temporary Redirect` 会保留原始请求方法；浏览器通常会继续把 `POST` 跳到新地址，但不带 `-L` 的 `curl` 不会自动跟随。
4. `http.Client` 如果把 `CheckRedirect` 设成 `http.ErrUseLastResponse`，拿到 `307` 也会被当成“收到了一个合法 HTTP 响应”。

和 v-ui 的关系：

- 这说明“启用 HTTPS 后，HTTP 请求先被 307 到 HTTPS”是完全符合实现预期的。
- 真正危险点在于：
  - 如果后续 API 仍然走 `http://127.0.0.1:<port>`，请求可能根本到不了 controller。
  - 如果 readiness probe 只判断“拿到任意 HTTP 响应就算成功”，那么 HTTPS listener 返回的 `307` 也会被误判成“HTTP 已恢复”。

适合 v-ui 的方案：

- 基于当前进程内重建 listener 的方案仍然可用。
- 更重要的是：
  - cert toggle 的调用端必须感知当前 scheme。
  - `waitForHTTPReady` 必须判定“这是一个真正的明文 HTTP listener”，而不是“HTTPS listener 给了一个 307”。

不适合直接照搬的方案：

- 仅靠 `systemd restart x-ui` 解决一切。
  - 它能更粗暴地重启进程，但不能自动修复“调用端还在走 HTTP”或“HTTP readiness 判定过宽”的问题。

参考资料：

- Go `net/http` package docs: https://pkg.go.dev/net/http
- Go `crypto/tls` package docs: https://pkg.go.dev/crypto/tls
- HTTP 307 redirect semantics: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/307

## 2. 本地代码调用链

### 2.1 disableHttps 调用链

`POST /xui/cert/disableHttps`

↓

`web/controller/cert.go`

- `disableHTTPS()` 调用 `a.certService.DisableHTTPS()`
- 成功后再调用 `GetStatus()` 并返回 `success=true`

关键位置：

- [web/controller/cert.go](/tmp/v-ui-init/v-ui/web/controller/cert.go:117)

↓

`web/service/cert.go`

- `DisableHTTPS()`
  - 读取当前 setting snapshot
  - `SetCertFile("")`
  - `SetKeyFile("")`
  - 如果手动证书文件还存在，则把状态改成：
    - `webCertStatus=uploaded`
    - `webCertMode=manual`
    - `webCertProvider=manual`
  - 调用 `restart(defaultRestartDelay)`
  - 然后调用 `waitForHTTPReady(port, defaultListenerWaitTimeout)`

关键位置：

- [web/service/cert.go](/tmp/v-ui-init/v-ui/web/service/cert.go:340)
- [web/service/cert.go](/tmp/v-ui-init/v-ui/web/service/cert.go:351)
- [web/service/cert.go](/tmp/v-ui-init/v-ui/web/service/cert.go:381)
- [web/service/cert.go](/tmp/v-ui-init/v-ui/web/service/cert.go:386)

↓

`web/service/panel.go`

- `RestartPanel(delay)` 不是 `systemd restart`
- 它只是当前进程内异步 `sleep(delay)` 后给自己发 `SIGHUP`

关键位置：

- [web/service/panel.go](/tmp/v-ui-init/v-ui/web/service/panel.go:13)

↓

`main.go`

- 主进程收到 `SIGHUP` 后：
  - `server.Stop()`
  - `server = web.NewServer()`
  - `server.Start()`

这说明：

- 协议切换逻辑确实会重新构建 `Server`
- 不是单纯 reload 某个 in-memory flag

关键位置：

- [main.go](/tmp/v-ui-init/v-ui/main.go:52)
- [main.go](/tmp/v-ui-init/v-ui/main.go:58)

↓

`web/web.go`

- `Start()` 每次都重新读取：
  - `webCertFile`
  - `webKeyFile`
  - `webListen`
  - `webPort`
- 如果 `certFile != "" || keyFile != ""`
  - 加载证书
  - `listener = network.NewAutoHttpsListener(listener)`
  - `listener = tls.NewListener(listener, c)`
  - 日志打印 `web server run https`
- 否则：
  - 普通 TCP listener
  - 日志打印 `web server run http`

关键位置：

- [web/web.go](/tmp/v-ui-init/v-ui/web/web.go:324)
- [web/web.go](/tmp/v-ui-init/v-ui/web/web.go:345)
- [web/web.go](/tmp/v-ui-init/v-ui/web/web.go:354)
- [web/web.go](/tmp/v-ui-init/v-ui/web/web.go:357)

↓

`web/web.go` Stop

- `Server.Stop()` 会执行：
  - `s.httpServer.Shutdown(s.ctx)`
  - `s.listener.Close()`

关键位置：

- [web/web.go](/tmp/v-ui-init/v-ui/web/web.go:377)
- [web/web.go](/tmp/v-ui-init/v-ui/web/web.go:385)

### 2.2 谁决定最终是 HTTP 还是 HTTPS

最终是否启用 TLS，只由 `web/web.go` 在 `Start()` 时读取到的：

- `webCertFile`
- `webKeyFile`

来决定。

不是 Gin middleware 决定的，也不是 `systemd` 决定的。

### 2.3 307 是从哪里来的

307 不是 Gin 强制 HTTPS middleware。

307 来自：

- `network.NewAutoHttpsListener(listener)`
- `AutoHttpsConn.readRequest()`

它会在 TLS listener 收到“明文 HTTP 请求”时，直接回：

- `http.StatusTemporaryRedirect`
- `Location: https://<host><uri>`

关键位置：

- [web/network/auto_https_listener.go](/tmp/v-ui-init/v-ui/web/network/auto_https_listener.go:9)
- [web/network/autp_https_conn.go](/tmp/v-ui-init/v-ui/web/network/autp_https_conn.go:27)
- [web/network/autp_https_conn.go](/tmp/v-ui-init/v-ui/web/network/autp_https_conn.go:43)

## 3. 远程只读证据

目标机：`139.180.135.210`

只读检查结果：

1. 运行态只有一个 `x-ui` 主进程和一个监听 socket

- `ps` 显示单个主进程：
  - `/usr/local/x-ui/x-ui`
- `ss -lntp` 显示单个监听：
  - `*:36235 users:(("x-ui",pid=10281,fd=7))`

这说明：

- 没有“旧 TLS 进程没退，新 HTTP 进程又起来了”的双 listener 现象
- 也没有 zombie / 端口泄漏迹象

2. 当前数据库状态仍是 HTTPS 模式

`/etc/x-ui/x-ui.db`:

- `webDomain=cshtps.527270.xyz`
- `webCertFile=/usr/local/x-ui/cert/panel.crt`
- `webKeyFile=/usr/local/x-ui/cert/panel.key`
- `webCertStatus=enabled`
- `webCertMode=manual`

3. 当前协议表现

- `curl -v http://127.0.0.1:36235/`
  - 返回 `307 Temporary Redirect`
  - `Location: https://127.0.0.1:36235/`
- `curl -vk https://127.0.0.1:36235/`
  - TLS 正常握手
  - 返回 `HTTP/1.1 404 Not Found`

这说明：

- 当前 listener 明确仍是 HTTPS
- 307 正是 Auto HTTPS listener 的行为

4. 日志证据

`journalctl -u x-ui` 只出现了一次 cert toggle 相关 listener 切换：

- `INFO - web server run https on [::]:36235`

但没有任何后续：

- `web server run http on [::]:36235`

也没有第二次 `SIGHUP` stop/start 对应日志。

这说明：

- 在本次 Clean Install 失败复现里，`disableHttps` 对应的“重建成 HTTP listener”根本没有发生。

## 4. 根因判断

### 4.1 排除项

A. `disableHttps` 没有清空 `webCertFile/webKeyFile`

- 从代码看，`DisableHTTPS()` 一开始就会调用：
  - `SetCertFile("")`
  - `SetKeyFile("")`
- 所以实现上不是“忘了清空”

结论：不是主因。

B. setting 清空了，但 web server 没重新读取

- `main.go` 的 `SIGHUP` 路径会 `server.Stop()` 后 `web.NewServer()` 再 `Start()`
- `web/web.go` 的 `Start()` 每次都会重新读 `webCertFile/webKeyFile`

结论：不是主因。

C. `RestartPanel` 只是异步 `SIGHUP`，不会重建 listener

- 它确实只是异步 `SIGHUP`
- 但 `main.go` 收到 `SIGHUP` 后会真实重建 `Server`
- `enableHttps` 已经成功证明这条路径能把 HTTP 切到 HTTPS

结论：不是主因。

D. 旧 TLS listener 没有 `Shutdown/Close`

- `Server.Stop()` 同时调用：
  - `httpServer.Shutdown(...)`
  - `listener.Close()`
- 远端也没有出现两个 listener 并存

结论：不是主因。

F. `disableHttps` 之后实际启动失败，旧进程还活着

- 当前远端是单进程、单监听
- 没有 systemd 反复拉起失败

结论：不是主因。

G. `systemd active` 状态误导

- 当前活动 listener、DB、协议探测三者一致
- 不是 `systemd active` 假阳性

结论：不是主因。

### 4.2 命中项

E. 存在强制 HTTPS redirect 逻辑

- 命中。
- 不是 Gin middleware，而是 `AutoHttpsConn` 在 HTTPS listener 上主动把明文 HTTP 请求回 307。

H. 其它原因

命中，而且是本次 P0 的核心。

核心子问题有两个：

#### H1. disable 请求在 HTTPS 模式下可能根本没命中 controller

在开启 HTTPS 后，端口行为已经变成：

- 明文 HTTP 请求先收到 307
- 真正 API 处理只在 HTTPS 下发生

如果调用端还继续请求：

- `http://127.0.0.1:<port>/xui/cert/disableHttps`

那么请求可能只拿到 307，而没有到达：

- `web/controller/cert.go:disableHTTPS()`

远端只读证据支持这一点：

- 当前 DB 仍然是 `enabled/manual`
- 日志里没有 disable 触发的第二次 stop/start

这说明在失败复现里，`disableHttps` 很可能根本没真正执行。

#### H2. `waitForHTTPReady()` 的成功条件过宽

当前实现：

- 访问 `http://127.0.0.1:<port>/`
- 只要 `client.Get(url)` 能拿到任意 HTTP 响应，就返回成功

关键代码：

- [web/service/cert.go](/tmp/v-ui-init/v-ui/web/service/cert.go:548)
- [web/service/cert.go](/tmp/v-ui-init/v-ui/web/service/cert.go:556)
- [web/service/cert.go](/tmp/v-ui-init/v-ui/web/service/cert.go:564)

问题在于：

- 如果端口仍然是 HTTPS listener
- `AutoHttpsConn` 会对明文 HTTP 返回 `307`
- 由于 `CheckRedirect` 被设成 `http.ErrUseLastResponse`
- `client.Get()` 会把这个 `307` 当成一个合法响应返回

于是 `waitForHTTPReady()` 会把：

- “HTTPS 端口返回了 307”

误判成：

- “HTTP listener 已恢复”

这会导致即使真实切换失败，也可能向上层报告成功。

### 4.3 最终根因分类

本次 P0 最准确的分类是：

- `E + H`

更具体地说：

1. 存在 HTTPS listener 自带的 307 跳转逻辑。
2. cert toggle 的调用端如果还走 `http://`，可能根本打不到 `disableHttps` controller。
3. `waitForHTTPReady()` 又把 307 误判成“HTTP ready”。

所以“disableHttps 返回成功但实际上没回到 HTTP”并不是单一的 listener 关闭失败，而是：

- 调用路径和探针语义都不够严格。

## 5. 方案对比

### 方案 1：enable/disable 后直接走 `systemd restart x-ui`

修改文件：

- `web/service/cert.go`
- 可能需要新增一个安全执行 systemctl 的 service helper

做法：

- `enableHttps/disableHttps` 不再发进程内 `SIGHUP`
- 直接执行 `systemctl restart x-ui`
- 重启后再做 readiness 检查

优点：

- 路径直观
- 和用户手工重启语义一致
- 避免当前进程内 reload 逻辑的时序问题

风险：

- 业务进程内直接调 systemd，耦合更重
- 需要 root + systemctl 可用
- 单元测试更难做
- 会改变现有 x-ui 的重启模型

回滚方式：

- 恢复 `RestartPanel` 调用路径即可

是否影响现有 x-ui：

- 影响较大

是否适合当前 v-ui：

- 能用，但不是最小修复。

### 方案 2：保持进程内 `SIGHUP`，修正 cert toggle 的探针与调用语义

修改文件：

- `web/service/cert.go`
- `web/html/xui/cert.html`
- 可能少量补 `web/controller/cert.go` 返回信息

做法：

1. `waitForHTTPReady()` 不能只判断“拿到任意 HTTP 响应”
   - 必须明确拒绝：
     - `307` 且 `Location` 指向 `https://...`
   - 或更严格地要求：
     - 返回的不是 TLS redirect
2. 前端在 `enableHttps` 成功后，后续 cert API 必须切到 `https://`
   - 或直接刷新到 `https://<host>:<port>/xui/cert`
3. 后端必要时返回更明确的状态，提示前端已切到 HTTPS

优点：

- 改动最小
- 不改现有重启模型
- 直接修正本次 P0 的两个真实问题

风险：

- 需要同时收紧后端探针和前端调用
- 如果未来还有非浏览器调用 cert API，也要注意 scheme 切换

回滚方式：

- 恢复 `waitForHTTPReady` 和 cert 页面的 scheme 处理即可

是否影响现有 x-ui：

- 影响最小

是否适合当前 v-ui：

- 最适合。

### 方案 3：让 `RestartPanel` 按场景区分“普通 reload”与“协议切换完整重启”

修改文件：

- `web/service/panel.go`
- `web/service/cert.go`
- 可能还要补一个 panel restart mode 枚举

做法：

- 普通 setting 仍走当前 `SIGHUP reload`
- `enableHttps/disableHttps` 走“完整进程重建”专用路径
  - 可以是显式退出交给 systemd 拉起
  - 也可以是更强语义的内部 restart

优点：

- 把协议切换与普通面板 reload 区分开
- 语义更清晰

风险：

- 比方案 2 复杂
- 需要重新定义 panel lifecycle
- 如果实现成自杀 + systemd 拉起，失败面会更大

回滚方式：

- 把 cert toggle 重新接回 `RestartPanel`

是否影响现有 x-ui：

- 中等

是否适合当前 v-ui：

- 可作为后续重构方案，但不是当前最小修复。

## 6. 推荐最小修复

推荐：`方案 2`

原因：

1. 从代码和远端证据看，`SIGHUP -> Stop -> NewServer -> Start` 本身是有效的。
   - `enableHttps` 已经证明这条路径能从 HTTP 切到 HTTPS。
2. 当前失败更集中在：
   - cert API 调用在协议切换后没有切换 scheme
   - `waitForHTTPReady()` 把 `307 https redirect` 当成了 HTTP 恢复
3. 因此没有必要先引入更重的 `systemd restart` 路线。

推荐的最小修复拆分：

### 后端

优先修：

- `web/service/cert.go`

具体应改：

1. `waitForHTTPReady()`
   - 如果响应是：
     - `307`
     - 且 `Location` 以 `https://` 开头
   - 必须视为“仍处于 HTTPS 模式”，不能返回成功
2. cert toggle 的 readiness 判断
   - 对 disable 之后的 HTTP ready，应该验证：
     - 不是 TLS handshake
     - 不是 HTTPS redirect
     - 是真正明文 HTTP 服务

### 前端

优先修：

- `web/html/xui/cert.html`

具体应改：

1. `enableHttps` 成功后，前端应更新 cert API 的调用 scheme
   - 最稳妥做法：直接跳转当前页面到 `https://.../xui/cert`
2. `disableHttps` 成功后，前端可再切回 `http://.../xui/cert`
   - 或在成功后刷新为相对 URL，由后端当前协议承接

## 7. 下一轮编码提示

第一轮修复只建议改这些文件：

- `web/service/cert.go`
- `web/html/xui/cert.html`

如需补更清晰的响应字段，可少量改：

- `web/controller/cert.go`

不建议第一轮就改：

- `web/service/panel.go`
- `main.go`
- `x-ui.service`
- `install.sh`

因为当前证据不支持“必须换成 systemd restart”这个结论。

第一轮编码目标应聚焦在两个点：

1. 后端不要把 HTTPS 307 误判成 HTTP ready
2. 前端在 enable 之后不要继续用 `http://` 调 cert API
