# Changelog

## Unreleased / Reality Progress Baseline

### Summary

- add minimal REALITY config support for `VLESS + TCP + REALITY`
- add panel-private `realityShare` metadata for share export
- add minimal REALITY `vless://` share link export
- keep `realityShare` out of final `config.json`
- keep Custom Share Address compatible with REALITY export
- freeze this work as a development baseline only

### Notes

- this baseline does not create a new stable release
- current stable release remains `v1.0.2-stable`
- development freeze tag: `v1.0.2-reality-baseline`

## v0.1.0

第一个稳定独立开发基线。

### Summary

- 完成 `c-ui` 从旧仓库切出并建立独立 GitHub 仓库
- 完成 Outbound 数据结构、初始化、只读 API、CRUD API 与前端 CRUD
- 完成 Routing 数据结构、初始化、只读 API、CRUD API 与前端 CRUD
- 完成 AI Traffic Split Template
- 完成 `managedOutboundsRouting` 开关与托管 Xray 配置生成
- 完成 Outbound / Routing 保存后的 `ValidateManagedXrayConfig` 与 `RestartXray` 接入
- 完成源码安装模式 bootstrap 与 `CGO_ENABLED=1` 构建链路
- 完成两台 Debian 11 VPS 真机矩阵验证

### Major Commits

- `31d9737` Prepare c-ui standalone repository
- `c9f07c7` Bootstrap file command for source install
- `8ba1b51` Bootstrap source install dependencies automatically
- `fdaea64` Build source install binary with CGO enabled
- `f7b9d03` Fix routing enabled persistence
- `ef92b52` Add AI routing template
- `9cc3358` Apply managed config after outbound routing changes
- `460fe7d` Add managed outbound routing config
- `0d92284` Add routing CRUD
- `11c301a` Add routing model and readonly page
- `819781b` Add outbound frontend CRUD
- `4160b6a` Add outbound CRUD API
- `160c872` Add outbound readonly page
- `5b78232` Add outbound list API
- `0d69c19` Add outbound model and initialization
- `5bc9e8b` Document runtime freeze strategy

### Important Fixes

- 修复 `routing_rules.enabled=false` 持久化缺陷，确保禁用规则不会进入最终托管配置
- 修复源码安装模式缺少预编译 `x-ui` 时直接失败的问题
- 修复 `go-sqlite3` 需要 CGO 时的构建链路，统一使用 `CGO_ENABLED=1`
- 补齐 bootstrap 依赖自动安装：Go、gcc、git、curl、tar、unzip、file
- 修复托管配置保存链路，确保 `TestConfig` 失败时回滚数据库修改

### Release Focus

- 独立仓库
- 独立安装入口
- 独立运行资产
- 稳定开发基线
