# Custom Share Address Override

## 功能目标

在入站详情弹窗和二维码弹窗里提供“自定义分享地址”能力，只对临时复制的分享链接和临时生成的二维码生效，不改实际入站配置。

## 实现方式

- 纯前端临时覆盖
- 仅替换分享地址中的连接地址部分
- 原始分享链接和原始二维码保持不变
- 留空时回退到原始链接和原始二维码

## 支持范围

- 支持：域名
- 支持：IPv4
- 不支持：IPv6

当输入 IPv6 或带方括号的 IPv6 时，前端会提示：

`当前自定义分享地址暂不支持 IPv6，请使用域名或 IPv4`

## 协议约束

- VMess 仍保持 `vmess://` + Base64(JSON)
- VLESS / Trojan / Shadowsocks 只做地址覆盖
- 不改 SNI / Host / Path / Flow / Query / Remark

## 明确不影响

- 不改 DB
- 不改 Xray
- 不重启服务
- 不影响订阅
- 不影响原始分享链接
- 不影响原始二维码

## UAT 结果

- `hk.example.com`：PASS
- `1.2.3.4`：PASS
- IPv6 warning：PASS
- 原始链接/二维码不变：PASS
- 未写 DB：PASS
- 未触发 Xray 重启：PASS

## 结论

Custom Share Address Override 已按第一阶段目标完成。
