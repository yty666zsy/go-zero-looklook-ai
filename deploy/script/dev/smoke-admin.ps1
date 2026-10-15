# 后台管理端（admin-api）冒烟测试
#
# 覆盖：登录 / 看板 / 民宿上下架+改价 / 订单列表详情+强制取消 / 用户禁用启用 / 评价删除
# 并且会顺带验证「禁用用户后 C 端登录被拒」这条联动。
#
# 用法：pwsh -File deploy/script/dev/smoke-admin.ps1

param(
  [string]$AdminApi = 'http://127.0.0.1:1005/admin/v1',
  [string]$UserApi = 'http://127.0.0.1:1004/usercenter/v1',
  [string]$OrderApi = 'http://127.0.0.1:1001/order/v1'
)

$ErrorActionPreference = 'Stop'
$script:pass = 0
$script:fail = 0

function Ok($msg) { $script:pass++; Write-Host ("  [PASS] " + $msg) -ForegroundColor Green }
function Bad($msg) { $script:fail++; Write-Host ("  [FAIL] " + $msg) -ForegroundColor Red }
function Check($cond, $msg) { if ($cond) { Ok $msg } else { Bad $msg } }
function Head($msg) { Write-Host ''; Write-Host ("== " + $msg) }
function Safe($fn) { try { & $fn } catch { '' } }

function PostJson($url, $body, $headers) {
  $p = @{ Uri = $url; Method = 'Post'; ContentType = 'application/json' }
  if ($headers) { $p.Headers = $headers }
  if ($null -ne $body) { $p.Body = ($body | ConvertTo-Json -Depth 6) }
  else { $p.Body = '{}' }
  try {
    return @{ ok = $true; data = (Invoke-RestMethod @p) }
  } catch {
    $resp = $_.Exception.Response
    $text = ''
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
      $text = $_.ErrorDetails.Message
    } elseif ($resp -and $resp.Content) {
      try { $text = $resp.Content.ReadAsStringAsync().GetAwaiter().GetResult() } catch {}
    }
    $code = 0
    if ($resp) { try { $code = [int]$resp.StatusCode } catch { $code = 0 } }
    return @{ ok = $false; status = $code; raw = $text }
  }
}

# ---------------------------------------------------------------- 登录
Head '1. 管理员登录'
$r = PostJson "$AdminApi/admin/login" @{ username = 'admin'; password = 'admin123' } $null
Check ($r.ok -and $r.data.code -eq 200) "admin/admin123 登录成功"
Check ($r.data.data.admin.role -eq 'super') "返回角色 super"
$token = $r.data.data.accessToken
$AH = @{ Authorization = "Bearer $token" }

$r = PostJson "$AdminApi/admin/login" @{ username = 'admin'; password = 'wrong' } $null
Check ((-not $r.ok) -or $r.data.code -ne 200) "错误密码被拒绝"

$r = PostJson "$AdminApi/dashboard/overview" @{} $null
Check ((-not $r.ok) -or $r.data.code -ne 200) "无 token 访问被拒绝(401)"

Head '2. 管理员自身'
$r = PostJson "$AdminApi/admin/profile" @{} $AH
Check ($r.ok -and $r.data.data.admin.username -eq 'admin') "profile 返回 admin"

# ---------------------------------------------------------------- 看板
Head '3. 数据看板'
$r = PostJson "$AdminApi/dashboard/overview" @{} $AH
$ov = $r.data.data
Check ($r.ok -and $ov.userCount -gt 0 -and $ov.homestayCount -gt 0) ("overview: 用户={0} 民宿={1} 店铺={2} 评价={3} 订单={4} GMV={5}" -f $ov.userCount, $ov.homestayCount, $ov.businessCount, $ov.commentCount, $ov.totalOrderCount, $ov.totalGmv)

$r = PostJson "$AdminApi/dashboard/trend" @{ days = 7 } $AH
Check ($r.ok -and $r.data.data.list.Count -eq 7) ("trend 返回 7 天（末日订单={0}）" -f $r.data.data.list[6].orderCount)

$r = PostJson "$AdminApi/dashboard/city" @{} $AH
Check ($r.ok -and $r.data.data.list.Count -gt 0) ("city 返回 {0} 个城市" -f $r.data.data.list.Count)

$r = PostJson "$AdminApi/dashboard/orderState" @{} $AH
$states = ($r.data.data.list | ForEach-Object { "$($_.tradeState):$($_.count)" }) -join ' '
Check ($r.ok) ("orderState: " + $states)

# ---------------------------------------------------------------- 民宿
Head '4. 民宿管理'
$r = PostJson "$AdminApi/homestay/list" @{ page = 1; pageSize = 5 } $AH
$total = $r.data.data.total
Check ($r.ok -and $total -gt 0) ("列表 total=$total")
$first = $r.data.data.list[0]

$r = PostJson "$AdminApi/homestay/list" @{ page = 1; pageSize = 5; keyword = '大理' } $AH
Check ($r.ok -and $r.data.data.total -gt 0) ("关键字「大理」命中 {0} 条" -f $r.data.data.total)

$r = PostJson "$AdminApi/homestay/detail" @{ id = $first.id } $AH
Check ($r.ok -and $r.data.data.homestay.id -eq $first.id) ("详情 #$($first.id) $($r.data.data.homestay.title)")

$r = PostJson "$AdminApi/homestay/cityOptions" @{} $AH
Check ($r.ok) ("城市选项 " + (($r.data.data.list | ForEach-Object { $_.city }) -join ','))

$old = $first.rowState
$target = if ($old -eq 1) { 0 } else { 1 }
$r = PostJson "$AdminApi/homestay/updateState" @{ id = $first.id; rowState = $target } $AH
Check ($r.ok -and $r.data.data.rowState -eq $target) ("上架状态 $old -> $target")
$r = PostJson "$AdminApi/homestay/detail" @{ id = $first.id } $AH
Check ($r.data.data.homestay.rowState -eq $target) "详情已同步新状态"
$r = PostJson "$AdminApi/homestay/updateState" @{ id = $first.id; rowState = $old } $AH
Check ($r.ok -and $r.data.data.rowState -eq $old) ("状态还原为 $old")

$oldPrice = $first.homestayPrice
$newPrice = $oldPrice + 11
$r = PostJson "$AdminApi/homestay/updatePrice" @{ id = $first.id; homestayPrice = $newPrice; marketHomestayPrice = $newPrice + 100 } $AH
Check ($r.ok) ("改价 $oldPrice -> $newPrice")
$r = PostJson "$AdminApi/homestay/detail" @{ id = $first.id } $AH
Check ([math]::Abs($r.data.data.homestay.homestayPrice - $newPrice) -lt 0.01) "详情价格已更新"
$r = PostJson "$AdminApi/homestay/updatePrice" @{ id = $first.id; homestayPrice = 900; marketHomestayPrice = 800 } $AH
Check ((-not $r.ok) -or $r.data.code -ne 200) "售价高于市场价被拒绝"
$r = PostJson "$AdminApi/homestay/updatePrice" @{ id = $first.id; homestayPrice = $oldPrice; marketHomestayPrice = $first.marketHomestayPrice } $AH
Check ($r.ok) "价格还原"

# ---------------------------------------------------------------- 订单
Head '5. 订单管理'
$r = PostJson "$OrderApi/homestayOrder/createHomestayOrder" @{} $null
Check ((-not $r.ok) -or $r.data.code -ne 200) "C 端下单需要登录"

$login = PostJson "$UserApi/user/login" @{ mobile = '13800000001'; password = '123456' } $null
Check ($login.ok -and $login.data.code -eq 200) "C 端用户 13800000001 登录成功"
$UH = @{ Authorization = "Bearer " + $login.data.data.accessToken }

$now = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$create = PostJson "$OrderApi/homestayOrder/createHomestayOrder" @{
  homestayId = $first.id; isFood = $false
  liveStartTime = $now + 86400; liveEndTime = $now + 172800
  livePeopleNum = 2; remark = 'admin smoke'
} $UH
Check ($create.ok -and $create.data.code -eq 200) "C 端下单成功"
$sn = $create.data.data.orderSn

$r = PostJson "$AdminApi/order/detail" @{ sn = $sn } $AH
$od = Safe { $r.data.data.order }
Check ($r.ok -and $od -and $od.tradeState -eq 0) ("管理端订单详情: state={0} 金额={1} 手机={2}" -f $od.tradeState, $od.orderTotalPrice, $od.userMobile)

$r = PostJson "$AdminApi/order/list" @{ page = 1; pageSize = 5; tradeState = 0 } $AH
Check ($r.ok -and ($r.data.data.list | Where-Object { $_.sn -eq $sn })) "订单列表能查到该单"

$r = PostJson "$AdminApi/order/stat" @{} $AH
Check ($r.ok -and $r.data.data.all -gt 0) ("订单统计 all={0} 待支付={1} 未使用/已使用={2} 已取消={3}" -f $r.data.data.all, $r.data.data.waitPay, ($r.data.data.waitUse + $r.data.data.used), $r.data.data.cancel)

$r = PostJson "$AdminApi/order/forceCancel" @{ sn = $sn; reason = '冒烟测试取消' } $AH
Check ($r.ok -and $r.data.data.tradeState -eq -1) "强制取消成功"
$r = PostJson "$AdminApi/order/forceCancel" @{ sn = $sn } $AH
Check ((-not $r.ok) -or $r.data.code -ne 200) "重复取消被拒绝"

$r = PostJson "$OrderApi/homestayOrder/userHomestayOrderDetail" @{ sn = $sn } $UH
Check ($r.ok -and $r.data.data.tradeState -eq -1) "C 端订单详情同步为已取消"

$r = PostJson "$AdminApi/order/detail" @{ sn = 'NOT_EXIST_SN' } $AH
Check ((-not $r.ok) -or $r.data.code -ne 200) "不存在的订单号被拒绝"

# ---------------------------------------------------------------- 用户
Head '6. 用户管理'
$r = PostJson "$AdminApi/user/list" @{ page = 1; pageSize = 5 } $AH
$u0 = Safe { $r.data.data.list[0] }
Check ($r.ok -and $r.data.data.total -gt 0) ("用户列表 total={0}，首个={1}（下单 {2} 笔）" -f $r.data.data.total, $u0.nickname, $u0.orderCount)

$uid = Safe { $login.data.data.userId }
if (-not $uid) {
  $r2 = PostJson "$AdminApi/user/list" @{ page = 1; pageSize = 50; keyword = '13800000001' } $AH
  $uid = Safe { $r2.data.data.list[0].id }
}
$r = PostJson "$AdminApi/user/updateStatus" @{ id = $uid; status = 0 } $AH
Check ($r.ok -and $r.data.data.status -eq 0) "禁用用户 #$uid"
$r = PostJson "$UserApi/user/login" @{ mobile = '13800000001'; password = '123456' } $null
Check ((-not $r.ok) -or $r.data.code -ne 200) "被禁用后 C 端登录失败（联动 OK）"
$r = PostJson "$AdminApi/user/updateStatus" @{ id = $uid; status = 1 } $AH
Check ($r.ok -and $r.data.data.status -eq 1) "恢复用户 #$uid"
$r = PostJson "$UserApi/user/login" @{ mobile = '13800000001'; password = '123456' } $null
Check ($r.ok -and $r.data.code -eq 200) "恢复后 C 端可登录"

# ---------------------------------------------------------------- 评价
Head '7. 评价管理'
$r = PostJson "$AdminApi/comment/list" @{ page = 1; pageSize = 5 } $AH
Check ($r.ok -and $r.data.data.total -gt 0) ("评价列表 total={0}" -f $r.data.data.total)
$c = Safe { $r.data.data.list[0] }

$r = PostJson "$AdminApi/homestay/detail" @{ id = $c.homestayId } $AH
$beforeCount = $r.data.data.homestay.commentCount

$r = PostJson "$AdminApi/comment/delete" @{ id = $c.id } $AH
Check ($r.ok) ("删除评价 #$($c.id)（民宿 #$($c.homestayId) $($c.homestayTitle)）")

$r = PostJson "$AdminApi/homestay/detail" @{ id = $c.homestayId } $AH
Check ($r.data.data.homestay.commentCount -eq ($beforeCount - 1)) ("评价数 {0} -> {1}（已回写）" -f $beforeCount, $r.data.data.homestay.commentCount)

$r = PostJson "$AdminApi/comment/delete" @{ id = $c.id } $AH
Check ((-not $r.ok) -or $r.data.code -ne 200) "重复删除被拒绝"

Write-Host ''
Write-Host ("==== PASS {0} / FAIL {1} ====" -f $script:pass, $script:fail)
if ($script:fail -gt 0) { exit 1 }
