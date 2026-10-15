# 本机一键启动 go-zero-looklook 的后端服务（Windows / PowerShell）
#
# 用法：
#   pwsh -File deploy/script/dev/run-local.ps1              # 构建并启动全部服务
#   pwsh -File deploy/script/dev/run-local.ps1 -Only travel-api,travel-rpc
#   pwsh -File deploy/script/dev/run-local.ps1 -Stop        # 停止全部服务
#   pwsh -File deploy/script/dev/run-local.ps1 -Status      # 查看状态
#
# 前置：
#   1) docker compose -f docker-compose-env.yml up -d   （mysql/redis/kafka 等）
#   2) 导入 deploy/sql/*.sql 与 deploy/sql/looklook_seed_*.sql
#   3) deploy/local-dev/*.yaml 由 deploy/script/dev/gen-local-config.mjs 生成

param(
  [string[]]$Only = @(),
  [switch]$Stop,
  [switch]$Status,
  [switch]$NoBuild
)

$ErrorActionPreference = 'Stop'
$root = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$binDir = Join-Path $root '.local-bin'
$logDir = Join-Path $root 'deploy/local-dev/logs'
$pidFile = Join-Path $root 'deploy/local-dev/logs/pids.json'

# 启动顺序很重要：被依赖的先起
$services = [ordered]@{
  'usercenter-rpc'   = 'app/usercenter/cmd/rpc/usercenter.go'
  'usercenter-api'   = 'app/usercenter/cmd/api/usercenter.go'
  'travel-rpc'       = 'app/travel/cmd/rpc/travel.go'
  'travel-api'       = 'app/travel/cmd/api/travel.go'
  'payment-rpc'      = 'app/payment/cmd/rpc/payment.go'
  'payment-api'      = 'app/payment/cmd/api/payment.go'
  'order-rpc'        = 'app/order/cmd/rpc/order.go'
  'order-api'        = 'app/order/cmd/api/order.go'
  'order-mq'         = 'app/order/cmd/mq/order.go'
  'admin-api'        = 'app/admin/cmd/api/admin.go'
  'aiagent-api'      = 'app/aiagent/cmd/api/aiagent.go'
  'mqueue-scheduler' = 'app/mqueue/cmd/scheduler/mqueue.go'
  'mqueue-job'       = 'app/mqueue/cmd/job/mqueue.go'
}

function Get-Targets {
  if ($Only.Count -eq 0) { return @($services.Keys) }
  return $Only
}

function Load-Pids {
  if (Test-Path $pidFile) {
    return (Get-Content -Raw $pidFile | ConvertFrom-Json -AsHashtable)
  }
  return @{}
}

function Save-Pids($map) {
  $map | ConvertTo-Json | Set-Content -Path $pidFile -Encoding utf8
}

if ($Stop) {
  $pids = Load-Pids
  $n = 0
  foreach ($k in $pids.Keys) {
    $p = $pids[$k]
    try {
      $proc = Get-Process -Id $p -ErrorAction Stop
      # go 服务会派生真实进程，直接整棵进程树杀
      & taskkill /PID $p /T /F 2>$null | Out-Null
      $n++
    } catch {}
  }
  Remove-Item $pidFile -ErrorAction SilentlyContinue
  Write-Host "stopped $n service(s)"
  exit 0
}

if ($Status) {
  $pids = Load-Pids
  foreach ($k in $services.Keys) {
    $p = $pids[$k]
    if ($p -and (Get-Process -Id $p -ErrorAction SilentlyContinue)) {
      Write-Host ("{0,-18} running (pid {1})" -f $k, $p)
    } else {
      Write-Host ("{0,-18} stopped" -f $k)
    }
  }
  exit 0
}

New-Item -ItemType Directory -Force -Path $binDir, $logDir | Out-Null

$targets = Get-Targets
# 校验参数
foreach ($t in $targets) {
  if (-not $services.Contains($t)) {
    throw "unknown service '$t'. available: $($services.Keys -join ', ')"
  }
}

Push-Location $root
try {
  if (-not $NoBuild) {
    Write-Host 'building...'
    foreach ($t in $targets) {
      $out = Join-Path $binDir "$t.exe"
      & go build -o $out $services[$t]
      if ($LASTEXITCODE -ne 0) { throw "build failed: $t" }
    }
  }

  $pids = Load-Pids
  foreach ($t in $targets) {
    $exe = Join-Path $binDir "$t.exe"
    if (-not (Test-Path $exe)) { throw "missing binary $exe (run without -NoBuild)" }

    # 先清掉旧的
    if ($pids[$t] -and (Get-Process -Id $pids[$t] -ErrorAction SilentlyContinue)) {
      & taskkill /PID $pids[$t] /T /F 2>$null | Out-Null
    }

    $log = Join-Path $logDir "$t.log"
    $errLog = Join-Path $logDir "$t.err.log"
    $conf = "deploy/local-dev/$t.yaml"
    $args = @('-f', $conf)
    $p = Start-Process -FilePath $exe -ArgumentList $args -WorkingDirectory $root `
      -RedirectStandardOutput $log -RedirectStandardError $errLog `
      -WindowStyle Hidden -PassThru
    $pids[$t] = $p.Id
    Write-Host ("started {0,-18} pid {1}  (log: {2})" -f $t, $p.Id, $log)
    Start-Sleep -Milliseconds 400
  }
  Save-Pids $pids
} finally {
  Pop-Location
}

Write-Host ''
Write-Host 'tip: pwsh -File deploy/script/dev/run-local.ps1 -Status'
