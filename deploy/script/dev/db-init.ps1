# 初始化 looklook 的本机数据库（Windows / PowerShell）
#
# 会做四件事：
#   1) 等待 mysql 容器可连接
#   2) 建库 looklook_travel / looklook_order / looklook_payment / looklook_usercenter
#   3) 保证宿主机可以用 root 连接（root@'%'）
#   4) 导入 deploy/sql/*.sql 建表脚本 + deploy/sql/looklook_seed_*.sql 演示数据
#
# 用法：
#   pwsh -File deploy/script/dev/db-init.ps1            # 全量：建库 + 建表 + 种子数据
#   pwsh -File deploy/script/dev/db-init.ps1 -SkipSeed  # 只建库 + 建表

param(
  [switch]$SkipSeed,
  [string]$Container = 'mysql'
)

$ErrorActionPreference = 'Stop'
$root = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path

$dbs = @('looklook_travel', 'looklook_order', 'looklook_payment', 'looklook_usercenter', 'looklook_aiagent')

function Invoke-Sql {
  param([string]$Database, [string]$Sql)
  # 在容器内执行，避免在命令行暴露密码（用容器自带的 $MYSQL_ROOT_PASSWORD）
  $execArgs = @()
  if ($Database) { $execArgs += @('-e', "DB=$Database") }
  $Sql | docker exec -i @execArgs $Container bash -lc 'mysql --default-character-set=utf8mb4 -uroot -p"$MYSQL_ROOT_PASSWORD" ${DB:+"$DB"}' 2>&1 |
    Where-Object { $_ -notmatch 'Using a password on the command line' }
}

Write-Host '1/4 waiting for mysql ...'
docker exec $Container bash -lc 'until mysqladmin ping -uroot -p"$MYSQL_ROOT_PASSWORD" --silent >/dev/null 2>&1; do sleep 1; done'
Write-Host '    mysql is ready'

Write-Host '2/4 creating databases ...'
$create = ($dbs | ForEach-Object { "CREATE DATABASE IF NOT EXISTS ``$_`` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;" }) -join ' '
$create += " CREATE USER IF NOT EXISTS 'root'@'%' IDENTIFIED WITH mysql_native_password BY '" + `
  (docker exec $Container bash -lc 'printf %s "$MYSQL_ROOT_PASSWORD"') + "';" + `
  " GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' WITH GRANT OPTION; FLUSH PRIVILEGES;"
Invoke-Sql -Sql $create | Out-Null
Write-Host "    ok: $($dbs -join ', ')  (root@'%' ensured)"

Write-Host '3/4 importing schema ...'
foreach ($d in $dbs) {
  $file = Join-Path $root "deploy/sql/$d.sql"
  if (-not (Test-Path $file)) { throw "missing $file" }
  Invoke-Sql -Database $d -Sql (Get-Content -Raw $file) | Out-Null
  Write-Host "    $d <- deploy/sql/$d.sql"
}

if ($SkipSeed) {
  Write-Host '4/4 seed skipped (-SkipSeed)'
  exit 0
}

Write-Host '4/4 importing demo seed data ...'
Invoke-Sql -Database 'looklook_usercenter' -Sql (Get-Content -Raw (Join-Path $root 'deploy/sql/looklook_seed_usercenter.sql')) | Out-Null
Write-Host '    looklook_usercenter <- deploy/sql/looklook_seed_usercenter.sql'
Invoke-Sql -Database 'looklook_travel' -Sql (Get-Content -Raw (Join-Path $root 'deploy/sql/looklook_seed_travel.sql')) | Out-Null
Write-Host '    looklook_travel     <- deploy/sql/looklook_seed_travel.sql'
Invoke-Sql -Database 'looklook_usercenter' -Sql (Get-Content -Raw (Join-Path $root 'deploy/sql/looklook_seed_admin.sql')) | Out-Null
Write-Host '    looklook_usercenter <- deploy/sql/looklook_seed_admin.sql'
Invoke-Sql -Database 'looklook_aiagent' -Sql (Get-Content -Raw (Join-Path $root 'deploy/sql/looklook_seed_aiagent.sql')) | Out-Null
Write-Host '    looklook_aiagent    <- deploy/sql/looklook_seed_aiagent.sql'

# 直接用 SQL 灌数据绕过了应用，go-zero 的模型缓存（Redis 里的 cache:* ）会留下旧值，
# 所以这里顺手清掉；只删 cache:* ，不动 asynq:* 的任务。
Write-Host 'clearing go-zero model cache ...'
$n = docker exec looklook-redis redis-cli -a G62m50oigInC30sf --no-auth-warning EVAL "local ks=redis.call('keys','cache:*') local n=#ks if n>0 then redis.call('del',unpack(ks)) end return n" 0 2>$null
Write-Host "    cleared $n cache key(s)"

Write-Host 'done. demo users: 13800000001 ~ 13800000008 / password 123456'
