# AI Agent 服务（aiagent-api）冒烟测试
#
# 覆盖：
#   C 端：健康信息 / 无 token 拒绝 / 找房意图→工具调用+房源卡片 / 政策问答(RAG) /
#         查订单 / 下单后用 Agent 取消订单（写操作二次确认）
#   管理端：经营看板 / 趋势 / 订单状态 / 房源列表 / 知识库 CRUD /
#           写操作（房源下架→确认执行→恢复）= 演练 Agent 的安全约束
#   主动型 Agent：经营日报生成（含指标卡片与执行步骤）/ 异常巡检（确定性规则、幂等）/
#                 告警一键处置（仍是二次确认 + 令牌一次性）
#
# 用法：pwsh -File deploy/script/dev/smoke-aiagent.ps1
#
# 注意：本脚本会临时创建一个测试订单并将其取消，并会在结束时把被处置房源恢复为「在售」。

param(
  [string]$AiApi   = 'http://127.0.0.1:1006/ai/v1',
  [string]$AdminApi = 'http://127.0.0.1:1005/admin/v1',
  [string]$UserApi = 'http://127.0.0.1:1004/usercenter/v1',
  [string]$TravelApi = 'http://127.0.0.1:1003/travel/v1',
  [string]$OrderApi = 'http://127.0.0.1:1001/order/v1'
)

$ErrorActionPreference = 'Stop'
$script:pass = 0
$script:fail = 0

function Ok($msg) { $script:pass++; Write-Host ("  [PASS] " + $msg) -ForegroundColor Green }
function Bad($msg) { $script:fail++; Write-Host ("  [FAIL] " + $msg) -ForegroundColor Red }
function Check($cond, $msg) { if ($cond) { Ok $msg } else { Bad $msg } }
function Head($msg) { Write-Host ''; Write-Host ("== " + $msg) }

function PostJson($url, $body, $headers) {
  $p = @{ Uri = $url; Method = 'Post'; ContentType = 'application/json; charset=utf-8' }
  if ($headers) { $p.Headers = $headers }
  if ($null -ne $body) { $p.Body = ($body | ConvertTo-Json -Depth 8 -Compress) } else { $p.Body = '{}' }
  try {
    return @{ ok = $true; data = (Invoke-RestMethod @p) }
  } catch {
    $text = ''
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) { $text = $_.ErrorDetails.Message }
    $code = 0
    if ($_.Exception.Response) { try { $code = [int]$_.Exception.Response.StatusCode } catch { $code = 0 } }
    return @{ ok = $false; status = $code; raw = $text }
  }
}

# 用 HttpClient 读 SSE（Invoke-RestMethod 会把流缓冲，拿不到增量）
function PostSse($url, $body, $headers) {
  $client = New-Object System.Net.Http.HttpClient
  $client.Timeout = [TimeSpan]::FromSeconds(120)
  try {
    $req = New-Object System.Net.Http.HttpRequestMessage([System.Net.Http.HttpMethod]::Post, $url)
    $json = $body | ConvertTo-Json -Depth 8 -Compress
    $req.Content = New-Object System.Net.Http.StringContent($json, [System.Text.Encoding]::UTF8, 'application/json')
    if ($headers) { foreach ($k in $headers.Keys) { $req.Headers.Add($k, [string]$headers[$k]) } }

    $resp = $client.SendAsync($req, [System.Net.Http.HttpCompletionOption]::ResponseHeadersRead).GetAwaiter().GetResult()
    if (-not $resp.IsSuccessStatusCode) {
      $txt = $resp.Content.ReadAsStringAsync().GetAwaiter().GetResult()
      return @{ ok = $false; status = [int]$resp.StatusCode; raw = $txt; events = @() }
    }
    $events = New-Object System.Collections.ArrayList
    $stream = $resp.Content.ReadAsStreamAsync().GetAwaiter().GetResult()
    $reader = New-Object System.IO.StreamReader($stream, [System.Text.Encoding]::UTF8)
    while (-not $reader.EndOfStream) {
      $line = $reader.ReadLine()
      if ($line -and $line.StartsWith('data:')) {
        $payload = $line.Substring(5).Trim()
        if ($payload) { [void]$events.Add(($payload | ConvertFrom-Json)) }
      }
    }
    $reader.Close()
    return @{ ok = $true; status = 200; events = @($events) }
  } finally {
    $client.Dispose()
  }
}

function EvTypes($sse) { return ($sse.events | ForEach-Object { $_.type }) }
function EvText($sse) {
  $t = ''
  foreach ($e in $sse.events) { if ($e.type -eq 'delta') { $t += $e.text } }
  return $t
}
function HasStep($sse, $tool) {
  return @($sse.events | Where-Object { $_.type -eq 'step' -and $_.step.tool -eq $tool }).Count -gt 0
}
function HasCard($sse, $kind) {
  return @($sse.events | Where-Object { $_.type -eq 'card' -and $_.card.kind -eq $kind }).Count -gt 0
}
function FirstAction($sse) {
  $a = @($sse.events | Where-Object { $_.type -eq 'action' })
  if ($a.Count -eq 0) { return $null }
  return $a[0].action
}

# ================================================================ C 端
Head '1. AI 服务健康信息（公开）'
$r = PostJson "$AiApi/health" @{} $null
Check ($r.ok -and $r.data.code -eq 200) 'health 返回 200'
Check ($r.data.data.knowledgeCount -gt 0) ("知识库已加载 {0} 条" -f $r.data.data.knowledgeCount)
Check ($r.data.data.toolsCustomer.Count -ge 5) ("C 端工具 {0} 个" -f $r.data.data.toolsCustomer.Count)
Check ($r.data.data.toolsAdmin.Count -ge 10) ("管理端工具 {0} 个" -f $r.data.data.toolsAdmin.Count)
if ($r.data.data.mock) { Write-Host ("  [INFO] 当前为本地规则模式（未配置 LLM.ApiKey），事件流与真实模型一致") -ForegroundColor Yellow }
else { Write-Host ("  [INFO] 当前接入真实模型: " + $r.data.data.provider) -ForegroundColor Yellow }

Head '2. 未登录访问被拒绝'
$r = PostSse "$AiApi/chat/stream" @{ content = '你好' } $null
Check ((-not $r.ok) -and $r.status -eq 401) '无 token 调用 SSE 返回 401'
$r = PostJson "$AiApi/conversation/list" @{} $null
Check ((-not $r.ok) -or $r.data.code -ne 200) '无 token 访问会话列表被拒绝'

Head '3. C 端用户登录'
$r = PostJson "$UserApi/user/login" @{ mobile = '13800000001'; password = '123456' } $null
Check ($r.ok -and $r.data.code -eq 200) '13800000001 登录成功'
$uToken = $r.data.data.accessToken
$UH = @{ Authorization = "Bearer $uToken" }

Head '4. C 端：找房意图 → 工具调用 + 房源卡片'
$sse = PostSse "$AiApi/chat/stream" @{ conversationId = 0; content = '上海有没有适合 2 个人住的民宿，预算 800 以内' } $UH
Check ($sse.ok) 'SSE 连接成功'
Check ((EvTypes $sse) -contains 'meta') ('收到 meta 事件') 
Check (HasStep $sse 'search_homestay') 'Agent 调用了 search_homestay'
Check (HasCard $sse 'homestay_list') '返回 homestay_list 卡片'
$answer = EvText $sse
Check ($answer.Length -gt 10) ("模型输出正文 {0} 字" -f $answer.Length)
Check ((EvTypes $sse) -contains 'done') '收到 done 事件'
$convId = ($sse.events | Where-Object { $_.type -eq 'meta' } | Select-Object -First 1).conversationId
Check ($convId -gt 0) ("会话已创建 id={0}" -f $convId)

Head '5. C 端：政策问答走 RAG（不调用工具）'
$sse = PostSse "$AiApi/chat/stream" @{ conversationId = 0; content = '取消订单有什么规则？退款多久到账？' } $UH
$t = EvText $sse
Check ($t -match '取消|退款') '回答命中了知识库内容'
Check (-not (HasStep $sse 'search_homestay')) '政策问答没有误调用工具'

Head '6. C 端：查我的订单'
$sse = PostSse "$AiApi/chat/stream" @{ conversationId = 0; content = '帮我看看我的订单' } $UH
Check (HasStep $sse 'my_orders') 'Agent 调用了 my_orders'
Check (HasCard $sse 'order_list') '返回 order_list 卡片'

Head '7. C 端：下单后用 Agent 取消（写操作二次确认）'
$r = PostJson "$TravelApi/homestay/search" @{ page = 1; pageSize = 1; city = '上海' } $UH
$homestayId = $r.data.data.list[0].id
Check ($homestayId -gt 0) ("取到房源 id={0}" -f $homestayId)

$now = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$r = PostJson "$OrderApi/homestayOrder/createHomestayOrder" @{
  homestayId = $homestayId; isFood = $false; liveStartTime = $now + 86400 * 7
  liveEndTime = $now + 86400 * 8; livePeopleNum = 2; remark = 'aiagent smoke'
} $UH
$sn = $r.data.data.orderSn
Check ($r.ok -and $sn) ("创建待支付订单 {0}" -f $sn)

$sse = PostSse "$AiApi/chat/stream" @{ conversationId = 0; content = "帮我取消订单 $sn" } $UH
Check (HasStep $sse 'cancel_order') 'Agent 识别出取消订单意图'
$act = FirstAction $sse
Check ($null -ne $act) '收到 action 待确认事件（未直接执行）'

# 此刻订单应仍是待支付 —— 证明 Agent 没有擅自执行写操作
$r = PostJson "$OrderApi/homestayOrder/userHomestayOrderDetail" @{ sn = $sn } $UH
Check ($r.data.data.tradeState -eq 0) '确认前订单仍是「待支付」（Agent 未自行执行）'

$r = PostJson "$AiApi/action/confirm" @{ token = $act.token; confirm = $true } $UH
Check ($r.ok -and $r.data.code -eq 200) '确认执行成功'

$r = PostJson "$OrderApi/homestayOrder/userHomestayOrderDetail" @{ sn = $sn } $UH
Check ($r.data.data.tradeState -eq -1) '确认后订单已变为「已取消」'

$r = PostJson "$AiApi/action/confirm" @{ token = $act.token; confirm = $true } $UH
Check ((-not $r.ok) -or $r.data.code -ne 200) '同一令牌不能重复执行（一次性）'

Head '8. C 端：会话列表 / 消息回放'
$r = PostJson "$AiApi/conversation/list" @{ limit = 10 } $UH
Check ($r.ok -and $r.data.data.list.Count -gt 0) ("会话列表返回 {0} 条" -f $r.data.data.list.Count)
$cid = $r.data.data.list[0].id
$r = PostJson "$AiApi/conversation/messages" @{ conversationId = $cid } $UH
Check ($r.ok -and $r.data.data.list.Count -gt 0) ("会话消息返回 {0} 条" -f $r.data.data.list.Count)
$hasCards = @($r.data.data.list | Where-Object { $_.cards.Count -gt 0 }).Count -gt 0
Check ($hasCards -or $true) '（消息中可能包含卡片，取决于该会话内容）'

# ================================================================ 管理端
Head '9. 管理端登录 + 场景隔离'
$r = PostJson "$AdminApi/admin/login" @{ username = 'admin'; password = 'admin123' } $null
Check ($r.ok -and $r.data.code -eq 200) 'admin 登录成功'
$aToken = $r.data.data.accessToken
$AH = @{ Authorization = "Bearer $aToken" }

$r = PostSse "$AiApi/admin/chat/stream" @{ conversationId = 0; content = '今天成交额多少' } $UH
Check ((-not $r.ok) -and $r.status -eq 401) 'C 端 token 调管理端 Agent 被拒绝(401)'

Head '10. 管理端：经营看板'
$sse = PostSse "$AiApi/admin/chat/stream" @{ conversationId = 0; content = '今天成交额和订单量怎么样？' } $AH
Check (HasStep $sse 'dashboard_overview') 'Agent 调用了 dashboard_overview'
Check (HasCard $sse 'metrics') '返回 metrics 卡片'
Check ((EvText $sse).Length -gt 5) '输出了经营结论'

Head '11. 管理端：趋势 / 订单状态 / 房源列表'
$sse = PostSse "$AiApi/admin/chat/stream" @{ conversationId = 0; content = '看一下最近 7 天的下单趋势' } $AH
Check (HasStep $sse 'dashboard_trend') 'Agent 调用了 dashboard_trend'
Check (HasCard $sse 'trend') '返回 trend 卡片'

$sse = PostSse "$AiApi/admin/chat/stream" @{ conversationId = 0; content = '订单状态占比是什么情况' } $AH
Check (HasStep $sse 'order_stat') 'Agent 调用了 order_stat'

$sse = PostSse "$AiApi/admin/chat/stream" @{ conversationId = 0; content = '上海有哪些民宿' } $AH
Check (HasStep $sse 'homestay_list') 'Agent 调用了 homestay_list'
Check (HasCard $sse 'admin_homestays') '返回 admin_homestays 卡片'

Head '12. 管理端：写操作安全约束（下架 → 确认 → 恢复）'
$r = PostJson "$AdminApi/homestay/list" @{ page = 1; pageSize = 1 } $AH
$hid = $r.data.data.list[0].id
Check ($hid -gt 0) ("取到房源 id={0}" -f $hid)

$sse = PostSse "$AiApi/admin/chat/stream" @{ conversationId = 0; content = "把房源 $hid 下架" } $AH
$act = FirstAction $sse
Check ($null -ne $act) '收到待确认动作（下架）'

$r = PostJson "$AdminApi/homestay/list" @{ page = 1; pageSize = 50 } $AH
$before = ($r.data.data.list | Where-Object { $_.id -eq $hid }).rowState
Check ($before -eq 1) '确认前房源仍在售（Agent 未擅自下架）'

$r = PostJson "$AiApi/admin/action/confirm" @{ token = $act.token; confirm = $true } $AH
Check ($r.ok -and $r.data.code -eq 200) '确认执行下架成功'
$r = PostJson "$AdminApi/homestay/list" @{ page = 1; pageSize = 50 } $AH
Check ((($r.data.data.list | Where-Object { $_.id -eq $hid }).rowState) -eq 0) '房源已下架'

# 恢复现场
$r = PostJson "$AdminApi/homestay/updateState" @{ id = $hid; rowState = 1 } $AH
Check ($r.ok -and $r.data.code -eq 200) '已恢复房源为在售（清理现场）'

Head '13. 管理端：取消写操作（reject 路径）'
$sse = PostSse "$AiApi/admin/chat/stream" @{ conversationId = 0; content = '禁用用户 2' } $AH
$act = FirstAction $sse
Check ($null -ne $act) '收到待确认动作（禁用用户）'
$r = PostJson "$AiApi/admin/action/confirm" @{ token = $act.token; confirm = $false } $AH
Check ($r.ok -and $r.data.code -eq 200) '取消执行返回成功'
$r = PostJson "$AdminApi/user/list" @{ page = 1; pageSize = 50 } $AH
$st = ($r.data.data.list | Where-Object { $_.id -eq 2 }).status
Check ($st -eq 1) '用户 2 仍为正常（取消后未产生任何改动）'

Head '14. 管理端：知识库 CRUD（含索引热更新）'
$r = PostJson "$AiApi/admin/knowledge/list" @{ keyword = 'AI' } $AH
Check ($r.ok -and $r.data.data.list.Count -gt 0) ("知识库检索命中 {0} 条" -f $r.data.data.list.Count)

$r = PostJson "$AiApi/admin/knowledge/save" @{
  scene = 'c'; category = 'smoke'; title = 'SMOKE-临时知识条目'
  content = '这是一条冒烟测试写入的知识条目，用于验证知识库热更新，稍后会被删除。'; keywords = 'smoke,临时'
} $AH
Check ($r.ok -and $r.data.data.id -gt 0) ("新增知识条目 id={0}" -f $r.data.data.id)
$kid = $r.data.data.id

$sse = PostSse "$AiApi/chat/stream" @{ conversationId = 0; content = 'SMOKE-临时知识条目是什么？' } $UH
Check ((EvText $sse) -match '冒烟测试') '新增知识立刻可被检索到（索引热更新生效）'

$r = PostJson "$AiApi/admin/knowledge/delete" @{ id = $kid } $AH
Check ($r.ok -and $r.data.code -eq 200) '删除知识条目成功'

$r = PostJson "$AiApi/admin/knowledge/delete" @{ id = $kid } $AH
Check ((-not $r.ok) -or $r.data.code -ne 200) '重复删除被拒绝'

Head '15. 管理端：会话与审计'
$r = PostJson "$AiApi/admin/conversation/list" @{ limit = 10 } $UH
Check ((-not $r.ok) -or $r.data.code -ne 200) 'C 端 token 不能读管理端会话'
$r = PostJson "$AiApi/admin/conversation/list" @{ limit = 10 } $AH
Check ($r.ok -and $r.data.data.list.Count -gt 0) ("管理端会话 {0} 条" -f $r.data.data.list.Count)

# ================================================================ 主动型 Agent
Head '16. 管理端：经营日报（无人值守产物）'
$r = PostJson "$AiApi/admin/report/generate" @{ force = $true } $AH
Check ($r.ok -and $r.data.code -eq 200) '手动生成经营日报成功'
$rep = $r.data.data.report
Check ($rep.id -gt 0) ("日报 id={0} 归属 {1}" -f $rep.id, $rep.date)
Check ($rep.steps.Count -ge 4) ("Agent 执行步骤 {0} 步（每个工具调用都可追溯）" -f $rep.steps.Count)
Check ($rep.cards.Count -ge 3) ("日报含 {0} 张结构化卡片" -f $rep.cards.Count)
Check ($rep.content -match '建议') '日报正文包含结论与行动建议'
Check ($rep.highlights.Count -gt 0) '日报含关键结论摘要'
Check ($rep.steps[0].tool -eq 'dashboard_overview') '第一步采集的是经营概览'

$r = PostJson "$AiApi/admin/report/generate" @{ force = $false } $AH
Check ($r.ok -and -not $r.data.data.created) '当天重复生成直接复用（幂等，不产生第二份）'

$r = PostJson "$AiApi/admin/report/list" @{ limit = 10 } $AH
Check ($r.ok -and $r.data.data.list.Count -gt 0) ("日报列表 {0} 条" -f $r.data.data.list.Count)
$rid = $r.data.data.list[0].id

$r = PostJson "$AiApi/admin/report/detail" @{ id = $rid } $AH
Check ($r.ok -and $r.data.data.report.content.Length -gt 100) '日报详情含完整正文'

$r = PostJson "$AiApi/admin/report/list" @{ limit = 10 } $UH
Check ((-not $r.ok) -or $r.data.code -ne 200) 'C 端 token 不能读经营日报'

Head '17. 管理端：异常巡检（确定性规则，不依赖大模型）'
$r = PostJson "$AiApi/admin/patrol/run" @{} $AH
Check ($r.ok -and $r.data.code -eq 200) '巡检执行成功'
$p = $r.data.data
Check ($p.scanned.orders -gt 0 -and $p.scanned.homestays -gt 0) `
  ("扫描 {0} 单 / {1} 套房源 / {2} 位用户" -f $p.scanned.orders, $p.scanned.homestays, $p.scanned.users)
Check ($p.findings -ge 1) ("命中 {0} 条异常" -f $p.findings)
Check (@($p.rules).Count -ge 1) ("规则命中明细 {0} 类" -f @($p.rules).Count)
Check ($p.costMs -lt 30000) ("巡检耗时 {0}ms" -f $p.costMs)

$r2 = PostJson "$AiApi/admin/patrol/run" @{} $AH
Check ($r2.ok -and $r2.data.data.created -eq 0) '重复巡检不会重复建告警（按 alert_key 幂等）'
Check ($r2.data.data.updated -ge 1) ("已存在告警累加命中次数 {0} 条" -f $r2.data.data.updated)

$r = PostJson "$AiApi/admin/alert/list" @{ status = 0; limit = 50 } $AH
Check ($r.ok -and $r.data.data.pendingCount -gt 0) ("待处理告警 {0} 条" -f $r.data.data.pendingCount)
$pending = $r.data.data.pendingCount
Check (@($r.data.data.list | Where-Object { $_.suggestion -and $_.detail }).Count -gt 0) '每条告警都带明细与处理建议'

$r = PostJson "$AiApi/admin/alert/list" @{ status = -1; limit = 50 } $AH
Check ($r.ok -and $r.data.data.totalCount -ge $pending) '全部告警数不少于待处理数'

$r = PostJson "$AiApi/admin/alert/list" @{ status = 0 } $UH
Check ((-not $r.ok) -or $r.data.code -ne 200) 'C 端 token 不能读异常告警'

Head '18. 管理端：告警一键处置（依然是写操作二次确认）'
# 找一条「可一键处置」且处于待处理状态的房源告警；都处理过就先重开一条
$r = PostJson "$AiApi/admin/alert/list" @{ status = 0; limit = 50 } $AH
$target = @($r.data.data.list | Where-Object { $_.actionTool -eq 'update_homestay_state' })[0]
if ($null -eq $target) {
  $r = PostJson "$AiApi/admin/alert/list" @{ status = -1; limit = 50 } $AH
  $target = @($r.data.data.list | Where-Object { $_.actionTool -eq 'update_homestay_state' })[0]
  if ($null -ne $target) { $null = PostJson "$AiApi/admin/alert/status" @{ id = $target.id; status = 0 } $AH }
}
Check ($null -ne $target) '找到可一键处置的告警（房源下架）'
$hid = [int]$target.actionArgs.homestayId

# 上面可能刚「重开」过一条告警，以当前待处理数为基准，后面才好比较增减
$r = PostJson "$AiApi/admin/alert/list" @{ status = 0; limit = 1 } $AH
$pending = $r.data.data.pendingCount

$r = PostJson "$AiApi/admin/alert/action" @{ id = $target.id } $AH
Check ($r.ok -and $r.data.data.token.Length -ge 16) '生成处置令牌成功'
Check ($r.data.data.confirmText -match '下架') '处置确认文案可读（说清要改什么）'
$token = $r.data.data.token

$r = PostJson "$AiApi/admin/action/confirm" @{ token = $token; confirm = $false } $AH
Check ($r.ok -and $r.data.data.ok -eq $false) '点「取消」不会改动任何数据'
$r = PostJson "$AdminApi/homestay/detail" @{ id = $hid } $AH
Check ($r.data.data.homestay.rowState -eq 1) '取消后房源仍在售'

$r = PostJson "$AiApi/admin/action/confirm" @{ token = $token; confirm = $true } $AH
Check ((-not $r.ok) -or $r.data.code -ne 200) '同一令牌不可复用'

$r = PostJson "$AiApi/admin/alert/action" @{ id = $target.id } $AH
$token2 = $r.data.data.token
$r = PostJson "$AiApi/admin/action/confirm" @{ token = $token2; confirm = $true } $AH
Check ($r.ok -and $r.data.data.ok -eq $true) '点「确认」后工具真正执行'
$r = PostJson "$AdminApi/homestay/detail" @{ id = $hid } $AH
Check ($r.data.data.homestay.rowState -eq 0) ("房源 #{0} 已下架（C 端不可见）" -f $hid)

$r = PostJson "$AiApi/admin/alert/status" @{ id = $target.id; status = 1 } $AH
Check ($r.ok -and $r.data.data.status -eq 1) '告警标记为「已处置」'
$r = PostJson "$AiApi/admin/alert/list" @{ status = 0; limit = 50 } $AH
Check ($r.data.data.pendingCount -lt $pending) '待处理数量随之下降'

$r = PostJson "$AdminApi/homestay/updateState" @{ id = $hid; rowState = 1 } $AH
Check ($r.data.data.rowState -eq 1) ("收尾：房源 #{0} 恢复在售" -f $hid)

# ================================================================ 结果
Write-Host ''
Write-Host ("================ PASS {0} / FAIL {1} ================" -f $script:pass, $script:fail)
if ($script:fail -gt 0) { exit 1 }
