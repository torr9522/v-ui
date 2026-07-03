# REALITY Phase 2

## Scope

本轮只完成 `REALITY` 的前端保存/读取承载。

包含：

- `StreamSettings.fromJson()` 读取 `security=reality`
- `StreamSettings.fromJson()` 读取 `realitySettings`
- `StreamSettings.toJson()` 在 `security=reality` 时带出 `realitySettings`
- 新建/编辑弹窗中的 REALITY 字段前端回显
- 页面内部切换 tab 或前端对象复制时保留 REALITY 字段

不包含：

- 数据库结构修改
- 后端 API 语义修改
- Xray 配置生成
- REALITY 真正提交运行
- 分享链接 / 订阅 / 二维码

## Current Boundary

当前仍保持以下限制：

- 点击保存时继续拦截 REALITY
- 固定提示：
  - `REALITY config generation is not enabled yet.`
- `TLS / XTLS` 原有保存逻辑不变
- `VLESS + TLS` 证书自动选择逻辑不变

## Serialization Rule

当前前端序列化规则：

- `security=tls`
  - 输出 `tlsSettings`
- `security=xtls`
  - 输出 `xtlsSettings`
- `security=reality`
  - 输出 `realitySettings`

这只用于前端模型承载和回显，不表示后端已支持 REALITY 生成。

## Compatibility

旧配置没有 `realitySettings` 时：

- `fromJson()` 使用默认模型
- 不报错
- 不影响非 REALITY 入站编辑

已有未来 REALITY 配置样例时：

- 可以被前端读取
- 可以在表单里回显
- 仍不能被提交为可运行配置

## Next Step

下一轮如果进入设计，应先做：

1. `realitySettings` 到后端保存结构的边界设计
2. `security=reality` 的 Xray 配置输出设计
3. `VLESS flow=xtls-rprx-vision` 与 REALITY 的生成关系梳理
4. `publicKey / shortId / fingerprint / spiderX` 的导出边界设计
