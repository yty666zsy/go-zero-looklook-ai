# go-zero-looklook · 民宿预订 + AI Agent（二次开发版）

> 基于 **[go-zero](https://github.com/zeromicro/go-zero)** 微服务框架，在开源项目 **[go-zero-looklook](https://github.com/Mikaelemmmm/go-zero-looklook)**
> 之上做的**全栈二次开发**：保留原有民宿预订业务，补齐「前后端分离 + 运营后台 + AI Agent 层 + 高并发防超卖」。

**新手直接跳到 [5 分钟跑起来](#5-分钟跑起来windows)** —— 不用配置任何密钥，克隆下来就能跑通全部功能。

---

## 一、这个版本比上游多了什么

| 能力 | 说明 | 位置 |
| --- | --- | --- |
| **前后端分离** | 新增移动端 H5「小栖民宿」与管理后台两个 Vue3 工程 | `web/`（5173）、`admin/`（5174） |
| **AI Agent 层** | C 端悬浮球助手 + 管理端 Copilot + 主动型「AI 值守」（经营日报 / 异常巡检 / 一键处置） | `app/aiagent/`（1006） |
| **运营管理后台** | 看板、订单、房源、用户、评价管理，独立服务与管理端鉴权 | `app/admin/`（1005） |
| **高并发防超卖** | 房源粒度 Redis 分布式锁 + 逐日占用日历（数据库唯一键兜底）+ 压测脚本 | `pkg/booking/`、`deploy/script/dev/bench-order.ps1` |
| **一键本机开发** | 建库、生成配置、起服务、冒烟、压测全部脚本化 | `deploy/script/dev/` |

几个能直接演示的点：

- AI 助手可以**真的调工具**：查订单、找房源、取消订单（写操作走「二次确认 + 一次性令牌 + 审计」）。
- 管理端「AI 值守」每天生成经营日报，异常巡检（超卖/待支付超时/低价/高取消率…）自动产出告警并能一键处置。
- 并发压测：100 个请求抢同一套房源的同一晚 → **只成功 1 单，0 超卖**（修复前是 40 单全成功）。

---

## 二、技术栈

- 后端：**go-zero 1.7**、gRPC、go-queue(Kafka)、asynq(延时任务)、MySQL 8、Redis、Prometheus/Jaeger（可选）
- 前端：Vue 3 + Vite 5 + TypeScript + Tailwind CSS（两端独立工程）
- 本机脚本：PowerShell 7（`pwsh`）+ Node.js 18+

---

## 三、5 分钟跑起来（Windows）

### 0. 先装好这四样

| 需要 | 版本 | 检查命令 |
| --- | --- | --- |
| Docker Desktop | 运行中即可 | `docker ps` |
| Go | **1.22+** | `go version` |
| Node.js | **18+** | `node -v` |
| PowerShell | **7+**（Windows 自带的 5.1 不行） | `pwsh -v` |

> 没装 pwsh 7：`winget install Microsoft.PowerShell`，或者从 https://aka.ms/powershell 下载。

### 1. 拉代码

```powershell
git clone https://github.com/yty666zsy/go-zero-looklook-ai.git
cd go-zero-looklook-ai
```

### 2. 起中间件（MySQL / Redis / Kafka）

```powershell
docker compose -f docker-compose-env.yml up -d mysql redis kafka asynqmon
```

首次会拉镜像，耐心等 1~2 分钟。（`jaeger` / `prometheus` / `grafana` / `elasticsearch` 是监控与日志，不影响功能，需要再起。）

### 3. 初始化数据库（建库 + 建表 + 演示数据）

```powershell
pwsh -File deploy/script/dev/db-init.ps1
```

看到 `done` 就好；脚本可重复执行。

### 4. 生成本机配置

```powershell
node deploy/script/dev/gen-local-config.mjs
```

它把 `app/**/etc/*.yaml` 里的容器地址换成宿主机端口，输出到 `deploy/local-dev/`（该目录不会进仓库）。

### 5. 启动后端（13 个进程）

```powershell
pwsh -File deploy/script/dev/run-local.ps1
```

```text
pwsh -File deploy/script/dev/run-local.ps1 -Status   # 看状态
pwsh -File deploy/script/dev/run-local.ps1 -Stop     # 全停
pwsh -File deploy/script/dev/run-local.ps1 -Only aiagent-api   # 只重启某个
```

日志在 `deploy/local-dev/logs/*.log`。

### 6. 启动前端（两个新终端）

```powershell
cd web    ; npm install ; npm run dev     # C 端 H5      → http://localhost:5173
cd admin  ; npm install ; npm run dev     # 运营管理后台  → http://localhost:5174
```

### 7. 登录体验

| 端 | 账号 | 密码 | 看什么 |
| --- | --- | --- | --- |
| C 端 H5（5173） | `13800000001` | `123456` | 右下角悬浮球「小栖」，试：`帮我看看我的订单` / `上海有什么房源` / `帮我取消订单` |
| 管理后台（5174） | `admin` | `admin123` | 左侧「AI 值守」看日报与告警，右下角 Copilot |

> 不想手点？见下面「四、自测」一条命令跑完 88 项断言。

---

## 四、自测 / 压测（可选，但推荐）

后端起来之后：

```powershell
pwsh -File deploy/script/dev/smoke-admin.ps1      # 管理端 39 项断言
pwsh -File deploy/script/dev/smoke-aiagent.ps1    # AI Agent 88 项断言
pwsh -File deploy/script/dev/bench-order.ps1      # 50 并发抢同一晚，验防超卖
```

`bench-order.ps1` 的关键输出：

```text
总请求 50  成功 1  拒绝 49
延迟 P50=45ms  P95=162ms
结论：PASS —— 只有 1 单成功（容量 1 晚），无超卖
```

---

## 五、AI 用真模型（可选）

**默认不需要任何密钥**：`aiagent-api` 会自动使用本地「规则规划器」（mock）模式，ReAct 工具调用、RAG 问答、日报巡检全部可演示。

想接真实大模型（DeepSeek / 通义千问 / Ollama / vLLM 等 OpenAI 兼容服务）：

```powershell
$env:AI_AGENT_API_KEY = "sk-你的key"
node deploy/script/dev/gen-local-config.mjs        # 密钥只写入本机配置（已 gitignore）
pwsh -File deploy/script/dev/run-local.ps1 -Only aiagent-api
```

验证：`POST http://127.0.0.1:1006/ai/v1/health`，返回 `"mock": false` 即已接上真模型。
换服务商就改 `app/aiagent/cmd/api/etc/aiagent.yaml` 的 `BaseUrl` / `Model`（仓库里 `ApiKey` 恒为空）。

> ⚠️ 提示：`deploy/local-dev/`、`.env` 都已在 `.gitignore` 里。**密钥只放环境变量或本机配置，别提交。**

---

## 六、目录结构

```text
app/                     后端微服务（每个服务 cmd/{api,rpc,mq} + model）
  usercenter/  travel/  payment/  order/     ← 上游原有业务
  admin/      运营管理后台接口（1005）
  aiagent/    AI Agent 层（1006）：agent 运行时 / tools / RAG / insight 巡检
  mqueue/     asynq 延时任务与定时任务
pkg/                     公共包（xerr、ctxdata、booking 防超卖原语…）
web/                     C 端 H5（Vue3 + Vite）
admin/                   运营管理后台（Vue3 + Vite）
deploy/
  sql/                   建表 + 演示数据 SQL
  script/dev/            本机开发脚本（db-init / run-local / smoke / bench / 生成配置）
  local-dev/             本机运行配置与日志（自动生成，不进仓库）
doc/                     上游原版文档（中文 / 英文）
README-LOCAL-DEV.md      完整本机开发指南（端口、数据模型、接口清单、FAQ）
```

端口速查：`order 1001/2001`、`payment 1002/2002`、`travel 1003/2003`、`usercenter 1004/2004`、
`admin-api 1005`、`aiagent-api 1006`；C 端 `5173`、管理端 `5174`；MySQL `33069`、Redis `36410`、Kafka `9094`。

---

## 七、二次开发说明与致谢

- **框架**：本项目后端基于 [go-zero](https://github.com/zeromicro/go-zero) 构建（`goctl` 1.7.3）。
- **上游项目**：[Mikaelemmmm/go-zero-looklook](https://github.com/Mikaelemmmm/go-zero-looklook)（MIT License，Copyright (c) 2022 Mikael）。
  本仓库是它的二次开发版本，**保留原 MIT 协议与版权声明**（见 [LICENSE](LICENSE)）；业务代码（民宿预订）与上游一致，
  新增部分为前后端分离改造、运营后台、AI Agent 层与高并发防超卖。
- 上游原版的英文/中文说明仍保留在 [README-cn.md](README-cn.md) 与 `doc/` 目录，可作为 go-zero 微服务结构的参考读物。

遇到问题先看 **[README-LOCAL-DEV.md](README-LOCAL-DEV.md)**（第 9 节防超卖设计、第 10 节常见问题）。

---

English: A full-stack homestay-booking demo built with **go-zero** (secondary development on top of
[go-zero-looklook](https://github.com/Mikaelemmmm/go-zero-looklook)): Go microservices + Vue3 H5 + admin console +
an AI Agent layer (tool-calling assistant, daily report, anomaly patrol) + oversell-proof booking
(Redis lock + per-night occupancy calendar). Quick start: `docker compose -f docker-compose-env.yml up -d mysql redis kafka asynqmon`
→ `pwsh -File deploy/script/dev/db-init.ps1` → `node deploy/script/dev/gen-local-config.mjs` → `pwsh -File deploy/script/dev/run-local.ps1`.
No API key required (falls back to a local rule planner).
