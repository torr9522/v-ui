# v-ui v1.0.2 Stable

## 版本说明

- `v-ui v1.0.2 Stable`
- 基于 `v1.0.1-stable`
- 运行层继续保持 `x-ui`

## 本版本完成内容

- 新增 `Custom Share Address Override`
- 自定义分享地址支持域名与 IPv4
- IPv6 自定义分享地址暂不支持，并提供明确 warning
- 修复 Trojan 分享链接参数不完整问题
- 修复 Shadowsocks 分享链接误导问题
- 完成 `Modern UI Phase 1`
- 完成入站 `TLS` 证书自动选择
- 移除 `x-ui.sh` / `x-ui_en.sh` 中 Cloudflare DNS 证书流程残留

## 稳定性确认

- `RC2` 审计结果：`P0=0`
- `RC2` 审计结果：`P1=0`
- 当前代码状态可进入 `v1.0.2-stable`

## 兼容性边界

- 运行层仍为 `x-ui`
- 不改 `DB`
- 不改 `API` 语义
- 不改 `Xray Config Generator` 主线
- 不恢复 Cloudflare DNS 证书流程

## 已知问题

- `IssueHttp Idempotent`
- Priority: `P2`

## 结论

`v-ui v1.0.2 Stable` 可以作为新的稳定发布点与后续开发基线。
