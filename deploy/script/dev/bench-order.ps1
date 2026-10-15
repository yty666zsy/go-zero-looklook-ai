# 并发下单压测：验证「同一房源同一晚不会超卖」
#
# 用 N 个并发请求抢同一套房源的同一晚，统计成功数 / 拒绝数 / 延迟分位。
# 判定：房源容量 = 1 晚（民宿是整套房源），成功数必须 <= 1。
#   成功数 > 1  => 超卖，FAIL（打印超卖明细）
#   成功数 == 1 => PASS（并发保护生效）
#
# 用法：
#   pwsh -File deploy/script/dev/bench-order.ps1                       # 50 并发抢某房源
#   pwsh -File deploy/script/dev/bench-order.ps1 -Concurrency 100
#   pwsh -File deploy/script/dev/bench-order.ps1 -HomestayId 10 -StartDays 120 -Nights 2
#   pwsh -File deploy/script/dev/bench-order.ps1 -KeepOrders           # 不清理测试订单
#
# 说明：脚本会自己挑一个 90 天后的空闲日期窗口，跑完把成功下出来的订单取消掉（释放占用）。

param(
  [int]$Concurrency = 50,
  [long]$HomestayId = 0,
  [int]$StartDays = 90,
  [int]$Nights = 1,
  [long]$LivePeopleNum = 2,
  [string]$OrderApi = 'http://127.0.0.1:1001/order/v1',
  [string]$UserApi = 'http://127.0.0.1:1004/usercenter/v1',
  [string]$TravelApi = 'http://127.0.0.1:1003/travel/v1',
  [string]$Mobile = '13800000001',
  [string]$Password = '123456',
  [switch]$KeepOrders
)

$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

function PostJson($url, $body, $headers) {
  $p = @{ Uri = $url; Method = 'Post'; ContentType = 'application/json; charset=utf-8' }
  if ($headers) { $p.Headers = $headers }
  if ($null -ne $body) { $p.Body = ($body | ConvertTo-Json -Depth 6 -Compress) } else { $p.Body = '{}' }
  return Invoke-RestMethod @p
}

function Percentile($sorted, [double]$p) {
  if ($sorted.Count -eq 0) { return 0 }
  $i = [Math]::Floor(($sorted.Count - 1) * $p)
  return $sorted[$i]
}

Write-Host '== 1. 登录 + 选房源 =='
$u = PostJson "$UserApi/user/login" @{ mobile = $Mobile; password = $Password } $null
$token = $u.data.accessToken
if (-not $token) { throw '登录失败' }
$H = @{ Authorization = "Bearer $token" }

if ($HomestayId -le 0) {
  $s = PostJson "$TravelApi/homestay/search" @{ page = 1; pageSize = 1 } $H
  $HomestayId = [long]$s.data.list[0].id
}
$d = PostJson "$TravelApi/homestay/homestayDetail" @{ id = $HomestayId } $H
$title = $d.data.homestay.title
if (-not $title) { $title = $d.data.title }
Write-Host "    房源 #$HomestayId $title  capacity=1晚"

$start = [DateTime]::Today.AddDays($StartDays)
$liveStartTime = [long]([DateTimeOffset]$start).ToUnixTimeSeconds()
$liveEndTime = $liveStartTime + 86400 * $Nights
Write-Host "    入住 $($start.ToString('yyyy-MM-dd')) 共 $Nights 晚"

$payload = @{
  homestayId    = $HomestayId
  isFood        = $false
  liveStartTime = $liveStartTime
  liveEndTime   = $liveEndTime
  livePeopleNum = $LivePeopleNum
  remark        = 'bench-order'
}

Write-Host ''
Write-Host "== 2. $Concurrency 个并发同时下单 =="
$sw = [Diagnostics.Stopwatch]::StartNew()
$results = 1..$Concurrency | ForEach-Object -Parallel {
  $url = $using:OrderApi
  $tk = $using:token
  $body = ($using:payload | ConvertTo-Json -Depth 6 -Compress)
  $one = [Diagnostics.Stopwatch]::StartNew()
  try {
    $r = Invoke-RestMethod -Uri "$url/homestayOrder/createHomestayOrder" -Method Post `
      -Headers @{ Authorization = "Bearer $tk" } -ContentType 'application/json; charset=utf-8' -Body $body -TimeoutSec 30
    $one.Stop()
    [pscustomobject]@{ ok = $true; ms = $one.ElapsedMilliseconds; msg = ''; sn = $r.data.orderSn }
  } catch {
    $one.Stop()
    $m = $_.Exception.Message
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
      try { $m = ($_.ErrorDetails.Message | ConvertFrom-Json).msg } catch { $m = $_.ErrorDetails.Message }
    }
    [pscustomobject]@{ ok = $false; ms = $one.ElapsedMilliseconds; msg = $m; sn = '' }
  }
} -ThrottleLimit $Concurrency
$sw.Stop()

$okList = @($results | Where-Object { $_.ok })
$failList = @($results | Where-Object { -not $_.ok })
$lat = @($results | ForEach-Object { $_.ms } | Sort-Object)

Write-Host ''
Write-Host '== 3. 结果 =='
Write-Host ("    总请求 {0}  成功 {1}  拒绝 {2}" -f $results.Count, $okList.Count, $failList.Count)
Write-Host ("    总耗时 {0}ms   QPS≈{1}" -f $sw.ElapsedMilliseconds, [Math]::Round($results.Count / ($sw.ElapsedMilliseconds / 1000.0), 1))
Write-Host ("    延迟 P50={0}ms  P95={1}ms  P99={2}ms  max={3}ms" -f (Percentile $lat 0.5), (Percentile $lat 0.95), (Percentile $lat 0.99), $lat[-1])
if ($failList.Count -gt 0) {
  Write-Host '    拒绝原因分布：'
  $failList | Group-Object msg | Sort-Object Count -Descending | ForEach-Object {
    Write-Host ("      {0,4} × {1}" -f $_.Count, $_.Name)
  }
}

$oversell = [Math]::Max(0, $okList.Count - 1)
Write-Host ''
if ($okList.Count -eq 0) {
  Write-Host '结论：全部被拒绝 —— 该日期段本来就被占用（换个 -StartDays 再试）' -ForegroundColor Yellow
} elseif ($oversell -eq 0) {
  Write-Host ("结论：PASS —— 只有 1 单成功（容量 1 晚），无超卖，房间占用唯一" ) -ForegroundColor Green
} else {
  Write-Host ("结论：FAIL —— 超卖 $oversell 单！（同一房源同一晚卖出 {0} 次）" -f $okList.Count) -ForegroundColor Red
  $okList | ForEach-Object { Write-Host ("      超卖订单 $($_.sn)") }
}

if (-not $KeepOrders -and $okList.Count -gt 0) {
  Write-Host ''
  Write-Host '== 4. 清理测试订单（释放占用）=='
  $n = 0
  foreach ($o in $okList) {
    try {
      $null = PostJson "$OrderApi/homestayOrder/cancelHomestayOrder" @{ sn = $o.sn } $H
      $n++
    } catch { Write-Host ("      取消失败 {0}: {1}" -f $o.sn, $_.Exception.Message) }
  }
  Write-Host "    已取消 $n 单"
}

if ($oversell -gt 0) { exit 1 } else { exit 0 }
