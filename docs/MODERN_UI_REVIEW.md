# Modern UI Review

本审计只关注 UI、交互、布局、说明、默认值、联动。

明确不做：

- 不修改数据库
- 不修改协议生成
- 不修改 Xray 配置
- 不照搬 3x-ui 后端 / 前端架构

参考对象：

- 当前 v-ui：`web/html/xui/*`、`web/assets/js/model/xray.js`
- 当前 3x-ui：`/tmp/v-ui-init/3x-ui-source/frontend/src/*`
- 当前 Xray 26.5.3 能力边界：已审计于 [XRAY_26_5_3_PROTOCOL_AUDIT.md](/tmp/v-ui-init/v-ui/docs/XRAY_26_5_3_PROTOCOL_AUDIT.md)

---

## 第一部分：Inbound UX

### v-ui 当前优点

1. 单页直出，理解成本低，打开就能看到所有字段。
2. 旧协议兼容面广，`vmess/vless/trojan/shadowsocks/socks/http/mixed/tunnel` 都能直接编辑。
3. 现有 TLS 面板证书自动填已经补上，至少解决了“还要手抄证书路径”的痛点。
4. 基础新增入站流程没有异步依赖，表单行为稳定。
5. 对老用户来说，字段位置变化少，迁移阻力低。

### v-ui 当前缺点

1. 表单是“整页平铺”，不是“按任务分组”。
   用户在“新增一个 VLESS TLS 节点”时，会同时看到流量、IP 限制、协议、传输、TLS、sniffing 等大量不相关字段。

2. 参数显隐规则弱。
   现在主要是协议级 `v-if` 和少量 TLS 显隐，缺少“协议 + transport + security”联合联动。

3. 命名停留在旧时代。
   仍然把 `tcp`、`http`、`xtls` 直接暴露给用户，而 Xray 26.5.3 的主线已经偏向 `RAW`、`XHTTP`、`REALITY`。

4. 缺少“路径引导”。
   用户不知道应该先选协议、再选 transport、再选 security，还是反过来。UI 没有给顺序。

5. 缺少“默认值解释”。
   比如 `sniffing`、`flow`、TLS 域名、HTTP path、WS headers，用户看见空框，不知道该不该填。

6. 缺少“能力裁剪”。
   Reality 完全不存在；即使后续补参数，如果还沿用当前平铺方式，复杂度会继续爆炸。

7. 客户端字段组织弱。
   `vmess/vless/trojan` 现在实质上只编辑第一个 client，缺少“用户是在建入站，还是在配客户端”的明确边界。

8. 错误预防不足。
   很多字段只能在保存失败或节点不可用后才暴露问题，表单内缺少即时提示。

9. 没有“高级模式隔离”。
   高风险、低频字段和高频字段混在一起，用户容易在第一次创建节点时就被淹没。

10. 缺少“结果预览心智”。
   用户不知道自己当前做的是 “VLESS + TCP + TLS” 还是 “Trojan + gRPC + TLS”，页面没有持续摘要。

### 3x-ui 为什么更好用

3x-ui 好用，不是因为字段多，而是因为它把“复杂参数”组织成了清晰路径：

1. 入站被拆成标签页。
   `Basic / Protocol / Stream / Security / Sniffing / Advanced`，用户一次只处理一个问题。

2. 默认值是“可用默认值”，不是空壳。
   新建 VLESS 时默认就是 `vless + tcp + none + sniffing 默认值`，不是让用户从空对象开始拼。

3. 显隐逻辑按能力而不是按文件写死。
   例如 `canEnableTls`、`canEnableReality`、`canEnableSniffing`、`hasSelectableTransport`，使 UI 自动收缩到当前协议可用的最小集合。

4. transport 与 security 有联动。
   选了不支持 TLS 的 transport，就不会继续展示 TLS 入口；支持 Reality 才会显示 Reality。

5. 文案是“用户任务导向”。
   例如 `SNI`、`shareAddrStrategy`、`deployTo`、`set default cert`、`scan target`，不是一堆原始字段名。

6. 高级参数折叠到单独区域。
   高级 JSON 编辑器存在，但不会把新手路径挤坏。

7. REALITY 不是只给几个输入框。
   它配了扫描、候选查找、随机 shortIds、生成 keypair、结果解释，这就是“减少文档依赖”的核心。

8. sniffing 是完整结构，不是只有开关。
   `destOverride`、`metadataOnly`、`routeOnly`、`ipsExcluded`、`domainsExcluded` 都有明确入口。

9. TLS 表单有更完整的现代字段。
   `ALPN`、`uTLS fingerprint`、TLS version、curvePreferences、证书设置动作按钮，符合 26.5.3 语境。

10. 列表页提供即时摘要。
   Inbound 列表直接展示 `protocol + transport + L4 + TLS/Reality` 标签，用户不用点进详情才能知道节点结构。

### Xray 26.5.3 对 Inbound UI 的实际要求

从 26.5.3 角度看，Inbound UI 不是“再多加几个字段”就够，而是必须解决这些交互问题：

1. 现代 transport 需要正确命名。
   用户应该看到 `RAW / WebSocket / gRPC / HTTPUpgrade / XHTTP`，而不是继续把旧名 `tcp/http` 当主入口。

2. security 需要成为一级选择，而不是附属开关。
   `none / tls / reality` 应该是清晰的互斥决策。

3. Reality 需要动作型 UI。
   只给 `publicKey/privateKey/shortId` 输入框没有意义，必须给生成、扫描、应用结果。

4. sniffing 需要默认值可见。
   `destOverride=["http","tls","quic","fakedns"]` 这类默认行为，不能继续隐藏在模型里。

5. 高级 transport 字段必须隔离。
   否则 XHTTP / TLS / Reality 的细节一旦都平铺，表单会彻底失控。

### Inbound 最值得改的 Top 10

1. 把当前单块平铺表单拆成 `Basic / Protocol / Transport / Security / Sniffing / Advanced` 标签结构。
2. 把 `tls + xtls` 双开关改成单一 `Security` 选择器，第一阶段只做 `none / tls`，为后续 `reality` 留位置。
3. 把 transport 入口从旧命名改成现代命名显示，至少 UI 文案改成 `RAW`、`WebSocket`、`gRPC`。
4. 在表单顶部增加“当前组合摘要”，实时显示 `VLESS + RAW + TLS` 这类结果。
5. 给高频字段加默认值和 hint，而不是留空框。
6. 把 TLS 证书选择、SNI、证书来源、到期时间收成一个更紧凑区块。
7. 把 `sniffing` 从三个零散开关改成完整分组区块，至少显式展示 `destOverride`。
8. 把与 client 相关的字段明确为“首个客户端 / 客户端配置”，避免用户误以为一个入站只能有一个用户。
9. 增加协议能力联动，禁止无效组合在 UI 出现。
10. 增加 Advanced 区域，把低频复杂参数隔离，而不是继续平铺。

---

## 第二部分：Routing UX

### v-ui 当前优点

1. 路由功能已经从模板直改升级到了可管理页面。
2. 出站可选、启停、增删改、排序、AI 模板都已经有入口。
3. 基础字段覆盖 `domain/ip/port/protocol/network/source/inboundTag/outboundTag`，足够支撑常见分流。
4. 表格方式对“已有规则浏览”是直观的。

### v-ui 当前缺点

1. 规则表过宽。
   用户第一次进来就看到十几列，信息密度高但优先级混乱。

2. 新增规则仍然偏“字段仓库”。
   Domain、IP、Port、Protocol、Network、Source、InboundTag 被平铺成一堆文本框，用户必须自己理解这些字段的组合关系。

3. 输入方式不友好。
   多值字段依旧是“JSON array 或换行”，这对普通用户非常不友好。

4. 缺少“从源到目的”的可视化。
   用户想知道“什么流量走到哪个 outbound”，当前页面不能一眼回答。

5. 缺少规则摘要卡片。
   表格虽然完整，但不适合快速扫读，也不适合移动端。

6. 排序交互弱。
   只有上移、下移、手填 sort，没有拖拽，没有更强的视觉反馈。

7. 不区分基础规则和高级规则。
   很多用户只需要 `domain -> outboundTag`，但页面从第一屏就把全部字段平铺出来。

8. 规则创建缺少模板化引导。
   除 AI 模板外，没有“常见规则”快速起步入口。

9. outboundTag 是唯一强约束字段，但 UI 没有把它塑造成“目标动作”。
   用户经常先纠结 domain 怎么写，而不是先明确这条规则要去哪里。

10. 字段命名偏技术内部。
   `inboundTag`、`protocol`、`source`、`network`、`sort` 对普通用户都需要二次解释。

### 3x-ui 为什么更好用

1. Routing 被拆成多个视角。
   不是只有规则表，还包括基础路由、规则卡片、测试器、导入导出。

2. 规则展示优先表达“流向”。
   `Inbound -> Outbound/Balancer` 是卡片核心，而不是十几个并列列头。

3. 条件被做成 chips。
   domain/ip/protocol/user/sourcePort 等不会占一整列，而是收纳成可扫读的条件块。

4. 规则列表支持卡片化和拖拽。
   排序更自然，顺序的“重要性”更明显。

5. 编辑弹窗更像“构造规则”而不是“填数据库字段”。
   先有 enable、源条件、目标 outbound/balancer，再补细项。

6. 多值条件统一成逗号输入、标签或选择器。
   用户不需要懂 JSON array。

7. inboundTag 有 remark 映射。
   不是直接裸露 tag，而是尽量带出人类可读信息。

8. 有 Route Tester。
   这非常重要，它让用户不必靠猜测验证 routing。

9. 基础路由预设与高级规则分开。
   常见场景可以少点几次完成。

10. import/export 存在，但不占主路径。
   专业入口有，普通入口不被污染。

### Xray 26.5.3 对 Routing UI 的实际要求

26.5.3 的 routing 字段已经远不止 v-ui 当前这套，因此 UI 更需要“分层”：

1. 基础用户只需要可理解的常见条件。
2. 高级用户才需要 `user/sourcePort/attrs/balancerTag` 等。
3. 新字段不能继续直接堆进同一个表单，否则复杂度会失控。
4. 规则结果必须可读，否则新增字段只会让规则更像 JSON 编辑器。

### Routing 最值得改的 Top 10

1. 列表默认改成“规则卡片/流向卡片”，而不是超宽表格优先。
2. 在卡片头部直接展示 `Inbound -> Outbound/Balancer`。
3. 多值条件改成 tags / token 输入，取消“JSON array 或换行”认知负担。
4. 规则表单拆成“来源条件 / 目标动作 / 高级条件”三组。
5. 提供拖拽排序，而不是只靠上移下移和 sort。
6. 给常用条件提供预设，如 `geoip:*`、`geosite:*`、AI 域名模板。
7. outboundTag 选择放到更靠前位置，让用户先明确“这条规则要去哪”。
8. 为 `domain/ip/inboundTag/protocol` 提供 chips 预览，减少保存后回表确认成本。
9. 增加 Routing 测试器或至少增加命中预览思路。
10. 把高级字段如 `user/sourcePort/attrs` 放进折叠区，不污染基础路径。

---

## 第三部分：3x-ui 值得借鉴

这里借鉴的是交互原则，不是源码实现。

### 值得借鉴的设计

1. 能力驱动显隐。
   UI 根据 `protocol + network + security` 动态收缩，而不是把所有字段都摆出来。

2. 默认值工厂。
   新建不同协议时，直接给用户一个可工作的初始状态。

3. 分标签页。
   把“创建节点”拆成多个可理解任务。

4. 条件摘要。
   列表页直接显示协议、transport、TLS/Reality、client 数等高价值摘要。

5. 高级参数隔离。
   高风险参数收进 Advanced，不影响主流程。

6. 动作型 UI。
   REALITY 的扫描、生成、随机化等，不让用户自己手工拼。

7. 多值输入标准化。
   tags、多选、select all / clear all，比 textarea + JSON 友好得多。

8. 列表与编辑视角分离。
   列表负责扫读，弹窗负责编辑，不让一个表承担两个目标。

9. 更强的校验前置。
   很多错误在表单内就能发现，而不是提交后才报。

10. 移动端/窄屏考虑。
   Routing 卡片化就是一个例子。

### 3x-ui 好用的根本原因

不是字段更多，而是：

- 它把复杂性藏在联动和默认值里。
- 它把现代 Xray 参数做成“选择路径”，不是“字段集合”。
- 它让用户先完成任务，再接触细节。

这正是 v-ui Modern UI 第二阶段应该学习的点。

---

## 第四部分：不能借鉴

### 绝不能直接照搬的内容

1. 3x-ui 的前端架构本身。
   它现在是独立 `frontend/src` React + Zod + hooks 体系，v-ui 当前是嵌入式模板页。直接照搬会演变成重写项目。

2. 3x-ui 的数据库模型。
   本轮明确不改 DB，也不应为了 UI 审美而改存储。

3. 3x-ui 的 API 组织方式。
   v-ui 当前控制器和页面交互方式不同，不应为了模仿 UI 感受而引入同样的 API 面。

4. 3x-ui 的全协议覆盖野心。
   它覆盖了更多现代协议和 transport，v-ui 现在不应该为了“看起来现代”而一口气扩张协议面。

5. 3x-ui 的高级编辑体系整体。
   Advanced JSON、Host、Node、Balancer、Tester 等是一整套系统，不能拆一半过来。

6. 3x-ui 的组件层级和状态管理方案。
   对 v-ui 来说，这会把“Modern UI 改造”变成“前端框架迁移”。

### 本轮可借鉴但不能复制的抽象

1. “能力判断函数”这个思路可以借鉴。
2. “默认值工厂”这个思路可以借鉴。
3. “表单分层”这个思路可以借鉴。
4. “摘要优先、细节折叠”这个思路可以借鉴。

---

## 第五部分：Modern UI Roadmap

这里的 roadmap 只允许：

- UI
- 交互
- 布局
- 说明
- 默认值
- 联动

不允许改 DB / Xray 生成逻辑。

### P0

必须先做，且不依赖新协议实现。

1. Inbound 弹窗拆分为标签页：
   `Basic / Protocol / Transport / Security / Sniffing / Advanced`

2. Inbound 顶部增加实时摘要：
   例如 `VLESS + RAW + TLS`

3. 用能力联动隐藏无效项：
   至少做到 TLS、证书区、sniffing、transport 子表单按当前选择收缩。

4. Routing 列表新增卡片视图：
   默认优先卡片，不替换表格，只是把主要入口改成更易读视图。

5. Routing 表单按“来源条件 / 目标 / 高级条件”重组。

6. 多值输入从 textarea 优先切到 tags/token 形式。

### P1

在不改生成逻辑的前提下继续补体验。

1. TLS 区块现代化文案：
   统一 `Security` 概念，弱化旧 `xtls` 暴露。

2. sniffing 区块补全基础可见项：
   至少把 `destOverride` 变成可见多选。

3. Routing 增加规则摘要 chips 和更清晰的 outbound 目标展示。

4. Inbound 列表增加 transport / TLS / Reality 摘要标签。

5. 为常见 Routing 条件提供 preset 辅助。

### P2

等 P0/P1 稳定后再做。

1. Advanced 折叠区或高级模式切换。
2. Routing tester 入口。
3. 更好的移动端布局。
4. 更丰富的 inline validation 和空值说明。

---

## 结论

### Inbound

v-ui 当前最大问题不是“字段不够”，而是：

- 没有分层
- 没有路径
- 没有摘要
- 没有能力联动

3x-ui 当前最值得学的是：

- 把复杂性转成步骤
- 把参数转成选择路径
- 把高级项藏到不打扰主流程的位置

### Routing

v-ui 当前已经“能用”，但离“好用”还差很远。

最关键不是继续往规则里加字段，而是：

- 让规则更像“流向说明”
- 让多值输入更自然
- 让排序和浏览更轻松

### Modern UI 第一阶段建议

只做这三件事：

1. 重构 Inbound 弹窗为分标签页结构。
2. 新增 Routing 卡片视图与规则分组表单。
3. 为现有字段补默认值、说明和联动，不新增任何协议生成能力。

这三件事能显著提升现代感，而且不会把本轮带偏成协议层重写。
