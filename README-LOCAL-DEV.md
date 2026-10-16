# 本机全栈开发指南（民宿预订 · 前后端分离）

> 在原始 go-zero-looklook 微服务之上做的一次改造：
> **后端**保留民宿预订业务链路（travel / order / payment / usercenter + Kafka + asynq），只补接口与数据；
> **前端**分两套：
> - `web/`：Vue 3 + Vite + TypeScript + Pinia + TailwindCSS 的移动端 H5（仿携程/美团的简洁版）；
> - `admin/`：同技术栈的**运营管理后台**（桌面布局），后端配套独立服务 `app/admin/cmd/api`（端口 1005）。
>
> 界面风格：两套前端共用同一套淡色高级灰令牌（无蓝色），统一 200~300ms 动效；H5 居中 480px 显示。

---

## 0. 目录速览

| 路径 | 说明 |
| --- | --- |
| `web/` | 新增前端项目（Vue3 + Vite + TS，移动端 H5） |
| `admin/` | 新增前端项目（Vue3 + Vite + TS，运营管理后台） |
| `app/admin/cmd/api/` | 新增后端服务 admin-api（管理后台专用，1005） |
| `app/aiagent/cmd/api/` | 新增后端服务 aiagent-api（AI Agent 层，1006） |
| `web/src/components/ai/`、`admin/src/components/ai/` | 两端助手的对话 UI（悬浮球/抽屉/整页 + 卡片 + 二次确认） |
| `deploy/local-dev/*.yaml` | 本机跑 Go 服务用的配置（`mysql/redis/kafka` 指向宿主机端口） |
| `deploy/script/dev/gen-local-config.mjs` | 由 `app/**/etc/*.yaml` 生成本机配置，**不要手改 `deploy/local-dev`** |
| `deploy/script/dev/db-init.ps1` | 建库 + 建表 + 导入演示数据（可重复执行） |
| `deploy/script/dev/run-local.ps1` | 一键构建/启动/停止/查看后端服务 |
| `deploy/script/dev/smoke-admin.ps1` | 管理端接口冒烟（39 项断言，回归必跑） |
| `deploy/script/dev/smoke-aiagent.ps1` | AI Agent 冒烟（88 项断言，回归必跑） |
| `deploy/script/dev/bench-order.ps1` | 并发下单压测：检验「同一房源同一晚不超卖」 |
| `deploy/script/dev/shot-readme.mjs` | 无头 Chrome 抓界面截图，输出到 `doc/screenshots/`（README 用） |
| `doc/screenshots/` | README 里引用的界面截图 |
| `deploy/script/seed/main.mjs` | 生成演示数据 SQL |
| `deploy/sql/looklook_seed_travel.sql` | 演示数据（导入 `looklook_travel`） |
| `deploy/sql/looklook_seed_usercenter.sql` | 演示用户（导入 `looklook_usercenter`） |
| `deploy/sql/looklook_seed_admin.sql` | 管理员表 + `user.status` 列 + 演示管理员（导入 `looklook_usercenter`） |
| `deploy/sql/looklook_aiagent.sql` | AI Agent 库（7 张表，导入 `looklook_aiagent`） |
| `deploy/sql/looklook_seed_aiagent.sql` | AI 知识库 44 条 + `ai-agent` 服务账号 |

端口约定：

| 服务 | 端口 |
| --- | --- |
| order-api / order-rpc | 1001 / 2001 |
| payment-api / payment-rpc | 1002 / 2002 |
| travel-api / travel-rpc | 1003 / 2003 |
| usercenter-api / usercenter-rpc | 1004 / 2004 |
| **admin-api** | **1005**（Prometheus **4012**） |
| **aiagent-api** | **1006**（Prometheus **4013**） |
| H5 前端 dev server | 5173 |
| **管理后台 dev server** | **5174** |
| MySQL（容器 → 宿主） | 3306 → **33069**，root / `PXDN93VRKUm8TeE7` |
| Redis（容器 → 宿主） | 6379 → **36410**，密码 `G62m50oigInC30sf` |
| Kafka（容器 → 宿主） | 9092 → **9094** |

---

## 1. 启动中间件

```bash
docker compose -f docker-compose-env.yml up -d
```

必需的是 `mysql` / `redis` / `kafka` / `asynqmon`，其余（jaeger、prometheus、grafana、elasticsearch、kibana、go-stash、filebeat）只影响链路追踪与监控，可按需启动：

```bash
docker compose -f docker-compose-env.yml up -d mysql redis kafka asynqmon
```

> 本机改动说明：
> - redis 容器名改为 `looklook-redis`（避免与宿主机已有 `redis` 容器重名），宿主端口用 `36410`（`36379` 落在 Windows 保留端口段内会 bind 失败）。
> - grafana 宿主端口 `3001 → 3002`。
> - mysql 增加 `MYSQL_ROOT_HOST: '%'`，让宿主机上的 Go 服务能用 root 直连。

---

## 2. 初始化数据库

```powershell
pwsh -File deploy/script/dev/db-init.ps1
```

脚本会：建 4 个库 → 保证 `root@'%'` 可连 → 导入建表 SQL → 导入演示数据。可重复执行（种子里的 schema 变更用存储过程做了幂等）。

只想建表不带数据：`pwsh -File deploy/script/dev/db-init.ps1 -SkipSeed`

**演示账号**：手机号 `13800000001` ~ `13800000008`，密码统一 `123456`。

管理后台管理员（`looklook_seed_admin.sql` 一起导入）：

| 账号 | 密码 | 角色 |
| --- | --- | --- |
| `admin` | `admin123` | super（超级管理员） |
| `operator` | `operator123` | operator（运营） |

想改演示数据，编辑 `deploy/script/seed/main.mjs` 后：

```bash
node deploy/script/seed/main.mjs
pwsh -File deploy/script/dev/db-init.ps1
```

---

## 3. 启动后端服务

```powershell
pwsh -File deploy/script/dev/run-local.ps1              # 构建并启动全部 13 个进程
pwsh -File deploy/script/dev/run-local.ps1 -Status      # 查看运行状态
pwsh -File deploy/script/dev/run-local.ps1 -Stop        # 全部停止
pwsh -File deploy/script/dev/run-local.ps1 -Only travel-api,travel-rpc   # 只重启某几个
```

- 产物在 `.local-bin/`，日志在 `deploy/local-dev/logs/<service>.log`。
- 启动顺序：usercenter → travel → payment → order → order-mq → mqueue-scheduler → mqueue-job，最后 admin-api 与 aiagent-api（两者都不依赖 RPC，谁先谁后都行）。
- `admin-api`（1005）是管理后台专用服务：直连 4 个业务库跑原生 SQL，不走 RPC，所以改它不需要动现有服务。
- `aiagent-api`（1006）是 AI Agent 层：同样不走 RPC，以 HTTP 调用上面这些服务的既有接口。

管理端接口冒烟（需后端已启动）：

```powershell
pwsh -File deploy/script/dev/smoke-admin.ps1
```

覆盖登录鉴权、看板 4 接口、民宿上下架/改价、下单 → 强制取消 → C 端同步、用户禁用 → C 端登录被拒、评价删除与计数回写等 39 项断言，跑完会打印 PASS/FAIL 汇总。

AI Agent 冒烟（需后端已启动）：

```powershell
pwsh -File deploy/script/dev/smoke-aiagent.ps1
```

55 → 88 项断言（原 55 项 + 主动型 Agent 33 项），覆盖：健康信息、未登录 401、C 端找房/政策问答/查订单/下单后由 Agent 取消（写操作二次确认 + 令牌一次性）、管理端看板/趋势/订单状态/房源列表、房源下架待确认 → 确认执行 → 恢复现场、取消写操作（reject 路径）、知识库 CRUD 与索引热更新、场景隔离（C 端 token 不能读管理端会话）、**经营日报生成（步骤/卡片/幂等）、异常巡检（确定性规则、重复巡检幂等、自动消除）、告警一键处置（取消不改数据、令牌不可复用、确认后才落地、收尾恢复）**。

如果改了 `app/**/etc/*.yaml`（或想换端口/地址），先重新生成本机配置：

```bash
node deploy/script/dev/gen-local-config.mjs
```

也可以手工单跑：

```bash
go run app/travel/cmd/api/travel.go -f deploy/local-dev/travel-api.yaml
```

---

## 4. 启动前端

### 4.1 H5（C 端）

```bash
cd web
npm install
npm run dev        # http://localhost:5173
```

- 开发期 `vite.config.ts` 用 proxy 把 `/travel` `/order` `/payment` `/usercenter` `/ai` 分别转发到 1003 / 1001 / 1002 / 1004 / 1006，**不依赖 nginx 网关**。
- 其他命令：`npm run typecheck`、`npm run build`（产物 `web/dist`）。

#### 页面

| 路由 | 页面 |
| --- | --- |
| `/` | 首页：搜索条、日期、热门城市、热门标签、优选民宿、猜你喜欢 |
| `/search` | 列表：关键词、日期、筛选（价格/人数）、排序、滚动加载 |
| `/detail/:id` | 详情：图片轮播、评分维度、房源介绍、真实评价、房东 |
| `/confirm/:id` | 确认订单：入住日期、人数、餐食开关、费用明细 |
| `/pay/:sn` | 收银台：倒计时、支付方式、模拟支付 |
| `/orders`、`/orders/:sn` | 订单列表（按状态筛选、取消）与订单详情 |
| `/mine`、`/login` | 个人中心与登录/注册 |

#### 智能助手「小栖」

右下角悬浮球 → 全屏对话面板（所有页面可用，登录页除外）。支持找房、查订单、取消订单（二次确认）、政策问答；助手回答里会带可点击的房源/订单卡片。实现见 `web/src/components/ai/` 与 `web/src/stores/ai.ts`。

### 4.2 运营管理后台

```bash
cd admin
npm install
npm run dev        # http://localhost:5174
```

- proxy：`/admin/v1` → `http://127.0.0.1:1005`（admin-api），`/ai/v1` → `http://127.0.0.1:1006`（aiagent-api）。
- 独立 JWT（`looklook_admin_token`，与 C 端不通用），登录后用顶部头像菜单「修改密码」。
- 其他命令：`npm run typecheck`、`npm run build`（产物 `admin/dist`）。

#### 页面

| 路由 | 页面 |
| --- | --- |
| `/login` | 登录（可点演示账号一键填入） |
| `/` | 数据看板：4 个指标卡（含环比）、7/14/30 天下单趋势、城市分布、订单状态占比、最新订单 |
| `/homestays` | 民宿管理：城市/状态筛选、关键词搜索、上下架、改价、详情弹窗 |
| `/orders` | 订单管理：状态 Tab、按订单号/手机号搜索、详情、强制取消 |
| `/users` | 用户管理：搜索、禁用/启用（禁用后 C 端无法登录） |
| `/comments` | 评价管理：搜索、三维评分明细、删除 |
| `/agent` | **AI 助手**：整页版运营 Copilot（左侧工具清单/历史会话，右侧对话） |
| `/knowledge` | **知识库**：AI 助手回答依据的增删改查（场景隔离、启停、索引热更新） |

> 图表为手写 SVG（`LineChart.vue` / `BarList.vue`），没有引入图表库；颜色沿用同一套「淡灰、无蓝」令牌。
> 运营 Copilot 另有一个全局抽屉：任意页面点右上角「AI 助手」按钮即可唤起，与 `/agent` 整页共享同一份会话状态。

---

## 5. AI Agent 层

新增一个**独立的 go-zero 服务** `app/aiagent/cmd/api`（端口 **1006**，前缀 `ai/v1`，Prometheus 4013），
在**不改动民宿预订业务**的前提下，给 C 端与管理端各加一套智能助手。前端是 `web/`（小栖）与 `admin/`（运营 Copilot）。

### 5.1 设计原则

1. **工具即 API**：Agent 不重复实现业务逻辑，而是以 HTTP 调用既有接口（travel-api 1003 / order-api 1001 / admin-api 1005），并**透传调用方自己的 JWT**，权限模型与普通请求完全一致（越权不可能发生）。
2. **双场景隔离**：C 端路由挂 `JwtAuth`，管理端路由挂 `AdminJwtAuth`；知识库用 `scene`（`all` / `c` / `admin`）隔离，避免运营内部口径泄露给 C 端。
3. **写操作必须二次确认**：`tool.Write=true` 的工具不会直接执行，而是生成一次性 `ai_action_token`（TTL 300s）并推 `action` 事件，前端确认后才真正调用；所有工具调用写入 `ai_tool_audit`。
4. **零外部依赖的 RAG**：中文 bigram 分词 + BM25（标题 ×3 / 关键词 ×2 加权）内存索引，不引入向量库；知识库改动后 `ReloadKnowledge` 热更新。
5. **LLM 双模**：`LLM.Provider = mock | openai`。默认 `mock`（本地规则规划器：意图识别 → 生成 tool_calls → 读结果 → 播报），**无 API Key 也能完整跑通 ReAct**；配上 `openai` + `ApiKey` 即切真实模型（OpenAI 兼容的流式 + function calling）。
6. **主动型 Agent**：除了「用户提问才回答」，服务自己也会按节奏干活——定时跑**经营日报**与**异常巡检**（见 5.6）。主动任务同样只能调用只读工具，涉及写操作只会生成待确认动作，**不会绕过人工确认**。

### 5.2 接口

| 接口 | 鉴权 | 说明 |
| --- | --- | --- |
| `POST /ai/v1/health` | 公开 | 当前模式（mock / 模型名）、知识库条数、两端工具名 |
| `POST /ai/v1/chat/stream` | C 端 JWT | C 端流式对话（SSE） |
| `POST /ai/v1/admin/chat/stream` | 管理端 JWT | 管理端流式对话（SSE） |
| `POST /ai/v1/conversation/list` `…/messages` `…/delete` | C 端 JWT | 会话列表 / 消息回放 / 删除 |
| `POST /ai/v1/action/confirm` | C 端 JWT | 确认或取消写操作 |
| `POST /ai/v1/admin/conversation/{list,messages,delete}` | 管理端 JWT | 同上（管理端） |
| `POST /ai/v1/admin/action/confirm` | 管理端 JWT | 同上（管理端） |
| `POST /ai/v1/admin/knowledge/{list,save,delete}` | 管理端 JWT | 知识库 CRUD（保存后立即热更新） |
| `POST /ai/v1/admin/report/{list,detail,generate}` | 管理端 JWT | 经营日报列表 / 详情 / 手动生成（`force=false` 时当天已有则直接复用） |
| `POST /ai/v1/admin/patrol/run` | 管理端 JWT | 立即跑一次异常巡检（定时任务之外的「手动触发」） |
| `POST /ai/v1/admin/alert/list` | 管理端 JWT | 告警列表（`status` -1 全部 / 0 待处理 / 1 已处置 / 2 已忽略 / 3 已消除） |
| `POST /ai/v1/admin/alert/action` | 管理端 JWT | 生成「一键处置」的待确认令牌（不执行，只生成） |
| `POST /ai/v1/admin/alert/status` | 管理端 JWT | 标记告警：已处置 / 已忽略 / 重新打开 |

**SSE 协议**：每行 `data: {json}`，`type ∈ ping | meta | step | delta | retract | card | action | done | error`
（`retract` 用来撤回模型在调用工具前吐出的过渡语，如「I'll look up that order first.」——它已经流式展示过，
前端按 `text.length` 裁掉；正文增量仍然逐字下发，保留打字机效果）。
前端用 `fetch` + `ReadableStream` 解析（`EventSource` 不支持 POST 与自定义头）。
`StepItem = {index,type,tool,args,status,summary,costMs}`，卡片类型有
`homestay_list` / `order_list` / `metrics` / `trend` / `admin_orders` / `admin_users` / `admin_homestays`。

### 5.3 工具清单

- **C 端（5）**：`search_homestay`、`homestay_detail`、`my_orders`、`order_detail`、`cancel_order`(写)
- **管理端（12）**：`dashboard_overview`、`dashboard_trend`、`dashboard_city`、`order_stat`、`order_list`、`order_detail`、`user_list`、`homestay_list`，
  以及写操作 `force_cancel_order`、`update_user_status`、`update_homestay_state`、`update_homestay_price`

### 5.4 数据表（库 `looklook_aiagent`）

`ai_conversation` / `ai_message` / `ai_knowledge` / `ai_action_token` / `ai_tool_audit` / `ai_report` / `ai_alert`，
建表与种子数据见 `deploy/sql/looklook_aiagent.sql`、`deploy/sql/looklook_seed_aiagent.sql`（44 条知识库 + `ai-agent` 服务账号）。

### 5.5 换成真实大模型（DeepSeek / 通义千问 / Ollama…）

`app/aiagent/cmd/api/etc/aiagent.yaml` 里已经按 **DeepSeek**（OpenAI 兼容协议）填好地址与模型，
**但 `ApiKey` 恒为空**——密钥不进仓库，只从环境变量或本机配置注入：

```yaml
LLM:
  Provider: openai                       # mock = 本地规则规划器；openai = 真实模型
  BaseUrl: https://api.deepseek.com/v1
  ApiKey: ""                             # 留空：走环境变量；都没有则自动降级 mock
  Model: deepseek-chat                   # 也可用 deepseek-reasoner
  Temperature: 0.3
  TimeoutMs: 60000
  MaxTokens: 1024
```

两种注入方式（任选其一）：

```powershell
# 方式 1：环境变量（推荐）。生成脚本会把 key 写进「本机配置」，仓库里依然是空
$env:AI_AGENT_API_KEY = "sk-你的key"                   # 也认 DEEPSEEK_API_KEY / OPENAI_API_KEY
node deploy/script/dev/gen-local-config.mjs
pwsh -File deploy/script/dev/run-local.ps1 -Only aiagent-api

# 方式 2：直接写 app/**/etc/aiagent.yaml（自己本地改，别提交）
```

`POST /ai/v1/health` 会返回 `provider/model/mock`，`mock: false` 即已走真实模型。
换其他 OpenAI 兼容网关（通义千问 / vLLM / Ollama）只改 `BaseUrl` + `Model` 即可。

> ⚠️ `deploy/local-dev/`（含注入后的密钥）与 `.env` 已在 `.gitignore` 里，别把密钥写进 `app/**/etc/*.yaml` 再提交。
> 想退回零依赖演示：`ApiKey` 留空即可（日报、巡检仍完整可用）。

> SSE 依赖响应不缓冲：`aiagent.yaml` 里 `Timeout: 0` 必须保持（关掉 rest 默认 3s 超时），
> nginx 网关的 `/ai/` location 也必须 `proxy_buffering off`。

### 5.6 主动型 Agent：经营日报 + 异常巡检

前面几节都是「人问 → 答」。这一节是**没人问，服务自己干**：运营打开「AI 值守」页（`/monitor`）就能看到
今天有什么异常、以及一份写好的日报，还能对异常点「一键处置」。

**调度内置在 aiagent-api 里**（goroutine + ticker，不额外引入分布式调度），开关与阈值都在配置里：

```yaml
Cron:
  Enabled: true
  RunOnStart: true          # 启动后先跑一次，本地演示不用等整点
  PatrolIntervalMin: 30     # 巡检间隔
  ReportHour: 9             # 每天 9 点生成日报
  ServiceUid: 3             # 内部服务账号 ai-agent 的 adminId
  MaxAlertsPerRule: 3       # 单条规则单次最多几条告警，避免刷屏
  # 巡检规则阈值
  PendingPayMin: 30
  IdleHomestayDays: 14
  LowScore: 4.5
  MinCommentCount: 5
  HighCancelRate: 0.6
  MinOrderForCancel: 3
```

**经营日报 = 一次无人值守的 Agent 会话**：按固定顺序跑 `dashboard_overview` → `dashboard_trend` → `order_stat` → `dashboard_city`，
把每个工具的结论拼成「核心指标 / 趋势 / 状态分布 / 城市分布 / 异常与风险 / 行动建议」六段正文，
同时把**结构化卡片**（指标、趋势、分布，和对话里的一模一样）与**执行步骤**（工具名、耗时、结论）一起落库 `ai_report`。
所以日报不只是文字，还能看到「Agent 是怎么得出来的」。唯一键 `(report_type, report_date)` 保证一天一份，重跑即覆盖。

若配了真实大模型（`Provider: openai`），会在确定性数据之上再让模型润色一版行文（提示词明确要求「只能用给定数据」）；
`mock` 模式保持确定性文案，日报依然完整可用。

**异常巡检是确定性规则，不依赖大模型**（告警必须可解释、可复现），7 条规则：

| 规则 | 判定 | 可一键处置 |
| --- | --- | --- |
| `pending_pay_timeout` | 待支付超过 `PendingPayMin` 分钟，一直占着房源 | 强制取消订单 |
| `overdue_unused` | 离店超 24h 仍是「未使用」，民宿主漏核销或用户没入住 | 人工核实 |
| `oversold` | 同一房源两笔有效订单的入住区间重叠（超售） | 人工介入 |
| `idle_homestay` | 近 `IdleHomestayDays` 天零有效订单的在售房源（聚合为一条） | 参考降价/下架 |
| `low_score` | 评价数够且评分低于 `LowScore` | 下架整改 |
| `price_above_market` | 售价高于门市价（数据异常） | 改回门市价 |
| `high_cancel_user` | 订单数够且取消率高于 `HighCancelRate` | 禁用账号 |

落库用唯一键 `alert_key`（规则 + 业务对象）幂等：重复命中只累加 `hit_count`，**不会刷屏**；
下次巡检发现异常自己消失了，则自动置为「已消除」。已被人工处置/忽略的不会自动重开（可手动「重开」），
所以「处理过的问题」不会反复弹回来。

**一键处置仍然走二次确认**：`alert/action` 只把告警里预置的工具 + 参数变成一次性令牌（TTL 300s），
前端弹确认框；用户点确认才调 `/ai/v1/admin/action/confirm` 真正执行——和对话式 Agent 是同一条安全通道，
工具调用同样落 `ai_tool_audit`。无人值守调用则以 **`role=service` 的内部账号 `ai-agent`** 身份签发 JWT，
admin-api 侧看到的是一条可追溯的实名调用，而不是匿名特权。

> 前端页面：管理后台左侧「AI 值守」（`admin/src/views/Monitor.vue`），两个 Tab——异常告警 / 经营日报。
> 告警按等级着色、可展开明细、可一键处置；日报支持「立即生成」与「强制重算当天」，并展示 Agent 执行时间线。

---

## 6. 本次新增/修改过的接口

**全部为 POST，响应 `{code:200,msg:"OK",data}`，业务错误 HTTP 400，鉴权失败 401，JWT 走 `Authorization: Bearer <token>`。**

### travel-api（1003）
| 接口 | 说明 |
| --- | --- |
| `POST /travel/v1/homestay/home` | 首页聚合：城市（实时房源数）、轮播、热门标签、优选、猜你喜欢 |
| `POST /travel/v1/homestay/search` | 搜索：关键词 / 城市 / 价格区间 / 人数 / 排序 / 分页 |
| `POST /travel/v1/homestay/homestayDetail` | 详情聚合：图片、店铺、三维评分统计（用 `JSON_EXTRACT` 实时算） |
| `POST /travel/v1/homestayComment/commentList` | 评价列表（真实实现，含昵称/头像） |

### order-api（1001）
| 接口 | 说明 |
| --- | --- |
| `POST /order/v1/homestayOrder/cancelHomestayOrder` | 取消订单：校验归属，仅「待支付」可取消 |
| `POST /order/v1/homestayOrder/userHomestayOrderList` | 列表项增加了 `id`（分页游标用） |

### payment-api（1002）
| 接口 | 说明 |
| --- | --- |
| `POST /payment/v1/thirdPayment/mockPay` | **开发环境模拟支付**：`Mode != dev` 直接拒绝；校验订单归属 → 创建流水 → 置为支付成功 → 通过 Kafka 推进订单状态，并做一次同步兜底 |

> 说明：`mockPay` 只用于本地联调，跳过真实微信支付。生产请走原有的 `thirdPaymentWxPay`。
> 若要接入真实微信支付，把 `app/payment/cmd/api/etc/payment.yaml` 里的 `Mode` 改成非 `dev` 即可让 `mockPay` 自动失效。

### admin-api（1005，管理后台专用）
前缀 `admin/v1`，**全部 POST**。管理员表在 `looklook_usercenter`，独立 JWT（`looklook_admin_token`），与 C 端账号体系完全隔离。

| 分组 | 接口 |
| --- | --- |
| 认证 | `/admin/login`、`/admin/profile`、`/admin/changePassword` |
| 看板 | `/dashboard/overview`（今日/累计指标 + 环比）、`/dashboard/trend`（7/14/30 天下单数与成交额）、`/dashboard/city`（城市分布）、`/dashboard/orderState`（订单状态占比） |
| 民宿 | `/homestay/list`、`/homestay/detail`、`/homestay/updateState`（上下架）、`/homestay/updatePrice`（改价，超过市场价会被拒绝）、`/homestay/cityOptions` |
| 订单 | `/order/list`、`/order/detail`、`/order/forceCancel`（已支付订单会同步把流水置为已退款并通知 C 端）、`/order/stat` |
| 用户 | `/user/list`、`/user/updateStatus`（禁用后 C 端无法登录） |
| 评价 | `/comment/list`、`/comment/delete`（删除后回写民宿评价数） |

实现要点：
- 不引入 RPC，`internal/dao` 直连 4 个业务库跑原生 SQL；写操作后统一 `FlushCache`（只清 `cache:*`，不动 `asynq:*` 队列）。
- 金额出入参在 API 层用「元」，库里仍是「分」；`homestay_order.live_start_date/live_end_date` 是 `datetime`，查询时用 `unix_timestamp()`。
- 为让「禁用用户」真正生效，`usercenter` 侧同步改了两处：`userModel_gen.go` 增加 `status` 字段、`loginLogic.go` 校验 `status != 1` 时拒绝登录。

### 顺带修掉的两个存量问题
1. `third_payment.pay_time` 是 `NOT NULL datetime`，未支付时零值 `time.Time` 会被驱动写成 `0000-00-00`，MySQL 8 严格模式下插入失败 → 在 model 的 `Insert` 里用 epoch 兜底。
2. `GetPaymentSuccessRefundByOrderSn` 原本拿 `trade_state` 去比 `pay_status` 的整数常量，永远查不到数据（订单详情的支付时间恒为 0）→ 改为按 `pay_status` 过滤。

---

## 7. 数据模型增量

`homestay` 增加展示字段（`deploy/sql/looklook_seed_travel.sql` 里用幂等存储过程添加）：

`city` / `address` / `tags` / `longitude` / `latitude` / `score` / `comment_count`

`homestay_comment` 增加：`nickname` / `avatar`

索引：`homestay(city,row_state,del_state)`、`homestay_comment(homestay_id,del_state,id)`

管理端相关（`deploy/sql/looklook_seed_admin.sql`，导入 `looklook_usercenter`）：

- `user.status`：`1` 正常 / `0` 禁用（默认 `1`，幂等添加）。
- `admin_user`：`username` / `password`(md5) / `nickname` / `role` / `status` / 时间戳，含 2 个演示管理员。

明细数据：8 城市、8 店铺、40 间民宿、300+ 条评价；图片统一用 `https://picsum.photos/seed/<seed>/<w>/<h>`，前端 `SmartImage` 加载失败会回退到本地灰色渐变占位。

AI Agent 相关（`deploy/sql/looklook_aiagent.sql`，导入 `looklook_aiagent`，7 张表）：

- `ai_report`：经营日报（正文 + 指标卡片 + 执行步骤 + 关键结论），唯一键 `(report_type, report_date)`，一天一份。
- `ai_alert`：异常巡检告警（等级/规则/明细/建议/可处置工具与参数），唯一键 `alert_key` 幂等，`status` 0 待处理 / 1 已处置 / 2 已忽略 / 3 已消除。

高并发防超卖（`deploy/sql/looklook_order.sql`，导入 `looklook_order`）：

- `homestay_calendar`：房源占用日历，逐日一行，唯一键 `(homestay_id, occupy_date)`；建表脚本会把现存有效订单（待支付/未使用/已使用）用递归 CTE 回填进去。详见第 9 节。

---

## 8. 状态约定

订单状态：`-1` 已取消 / `0` 待支付 / `1` 未使用(已支付) / `2` 已使用 / `3` 已退款 / `4` 已过期。

金额单位：**数据库与 RPC 用「分」，API 层用 `tool.Fen2Yuan/Yuan2Fen` 转成「元」**，前端拿到的都是元。

`homestay_comment.star` 是 JSON 数组字符串 `[卫生,位置,服务]`。

---

## 9. 高并发防超卖：房源占用日历 + 分布式锁

**问题**：原本下单链路是「RPC 查房源 → 直接 insert 订单」，没有任何并发控制。同一房源同一晚，N 个请求同时下单会**全部成功**
（`bench-order.ps1 -Concurrency 50` 修复前实测：50 并发 **成功 40 单**，超卖 39 单）。

**方案**（`pkg/booking`，两层保护，原则是「宁可下单失败，不可卖重」）：

1. **房源粒度 Redis 分布式锁** `lock:booking:homestay:{id}`：`SETNX` + TTL 8s + Lua 比对 token 释放；
   自旋等待 1.5s，抢不到返回「当前预订火爆，请稍后重试」。作用是把同一房源的并发下单串行化。
2. **房源占用日历 `looklook_order.homestay_calendar`**：逐日一行，唯一键 `(homestay_id, occupy_date)`；
   占位和订单写在**同一个事务**里，任一晚冲突就整单回滚，返回「所选日期已被预订（10-15、10-16），请更换日期」。
   唯一键是最终兜底：**即使 Redis 挂了或锁被绕过，数据库层面也不会超卖**。

释放：订单 **取消 / 退款 / 过期** 时删掉对应日历行（`order-rpc` 的 `UpdateHomestayOrderTradeState`；管理端强制取消在
`admin-api` 的 `ForceCancelOrder`），日期立刻重新可订。存量有效订单在建表脚本里已用递归 CTE 回填。

**压测/验证**：

```powershell
pwsh -File deploy/script/dev/bench-order.ps1                       # 50 并发抢同一晚
pwsh -File deploy/script/dev/bench-order.ps1 -Concurrency 100 -Nights 2
```

输出成功/拒绝数、拒绝原因分布、QPS 与 P50/P95/P99；容量按「1 晚/房源」算，**成功数 > 1 即 FAIL 并列超卖订单**，
脚本结束会自动取消测试订单（释放占用）。修复后结果：**1 成功 / 49 拒绝 / 0 超卖**。

**边界**：锁是房源粒度，同一房源的高并发下单会排队（1 晚场景实测 P95 仍在几十毫秒）；
民宿按「整套房源」计价，所以容量固定 1 晚/房源，暂不按人数扣减；搜索列表未按日历过滤（可能点进已订走的日期后被拒单），属后续可做项。

**顺带补的**：rpc 侧要给前端一句人话时用 `xerr.NewErrBiz("…")`，api 侧用 `xerr.BizMessage(err)` 取出来返回（`[biz]` 标记）；
没带标记的错误（db / panic / 超时）不会被透出，避免把底层细节暴露给前端。

---

## 10. 常见问题

- **下单报「所选日期已被预订（xx-xx），请更换日期」**：正常，是第 9 节的防超卖在起作用（锁 + 占用日历）。
  想看某套房的占用情况：`select * from looklook_order.homestay_calendar where homestay_id = ?`。
  老库补这张表时**不要**整份重灌 `looklook_order.sql`（脚本里是 DROP + CREATE，会清空订单），只执行文件末尾的 `homestay_calendar` 段落即可。
- **接口返回 `数据库繁忙,请稍后再试`**：看 `deploy/local-dev/logs/*.err.log`，通常是 MySQL/Redis 没起来，或 `root@'%'` 没建（跑一次 `db-init.ps1`）。
- **数据已重新导入，但接口还是返回旧值**：go-zero 的模型缓存（Redis 里的 `cache:*`）里还是旧数据。`db-init.ps1` 会自动清理；手工清理：`docker exec looklook-redis redis-cli -a G62m50oigInC30sf --no-auth-warning --scan --pattern 'cache:*'`，然后 `DEL` 掉。
- **改了 `deploy/sql/*.sql` 后**：直接 `docker exec` 里 `mysql` 灌数据即可，但记得清 `cache:*`。
- **Redis 容器起不来 / bind 失败**：宿主机端口换成保留段之外的（当前 `36410`）。
- **改了 proto/api 描述文件**：用 goctl 重新生成后再 `go build ./...`；本机配置不要重复生成覆盖了自定义项。
- **前端请求 404**：确认 `vite.config.ts` 的代理前缀与后端 `prefix` 一致（`/travel/v1`、`/order/v1`、`/payment/v1`、`/usercenter/v1`、管理端 `/admin/v1`）。
- **管理端登录 401 / 提示账号密码错误**：确认 `looklook_seed_admin.sql` 已导入（`admin` / `admin123`）；管理端 token 与 C 端不通用，两边要分别登录。
- **管理端改了数据，C 端看到的是旧值**：admin-api 写操作后会清 `cache:*`；如果是直接 `docker exec` 灌的 SQL，需要手动清（见上一条）。
- **`admin` 前端 5174 与 `web` 的 5173 互不影响**：两个 dev server 共用一套 Tailwind 令牌但是独立工程，改样式记得分别 `npm run build` 验证。
- **「AI 值守」页没有数据**：日报与告警都是服务自己跑出来的。配置里 `Cron.Enabled: true` 后，启动 20 秒左右会自动跑一次（`RunOnStart`）；
  也可以直接在页面上点「立即巡检」/「立即生成」。想看日志就翻 `deploy/local-dev/logs/aiagent-api.log`（搜 `insight cron` / `patrol done`）。
- **巡检告警感觉太多/太少**：阈值都在 `Cron` 段（`PendingPayMin`、`IdleHomestayDays`、`LowScore`、`HighCancelRate`…），
  `MaxAlertsPerRule` 控制单条规则单次最多产出几条；改完重新生成配置并重启 `aiagent-api` 即可。
- **一键处置点了没反应**：确认令牌 TTL 300 秒且只能用一次；过期或复用时后端会返回「该动作已处理过，请重新发起」，关掉弹窗重新点一次即可。
