// 把 app/**/etc/*.yaml 转成"本机开发"配置：把容器内的服务名换成宿主机映射端口。
// 这样不用改原配置（docker 部署仍然可用），也不要手工维护两份。
// 用法: node deploy/script/dev/gen-local-config.mjs
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const ROOT = path.resolve(__dirname, '../../..')
const APP = path.join(ROOT, 'app')
const OUT = path.join(ROOT, 'deploy/local-dev')

// docker-compose-env.yml 里映射到宿主机的端口
const REPLACEMENTS = [
  [/mysql:3306/g, '127.0.0.1:33069'],
  [/redis:6379/g, '127.0.0.1:36410'],
  [/kafka:9092/g, '127.0.0.1:9094'],
  [/jaeger:14268/g, '127.0.0.1:14268'],
]

// 服务名 -> 配置文件（顺序即推荐启动顺序）
const TARGETS = [
  ['usercenter-rpc', 'app/usercenter/cmd/rpc/etc/usercenter.yaml'],
  ['usercenter-api', 'app/usercenter/cmd/api/etc/usercenter.yaml'],
  ['travel-rpc', 'app/travel/cmd/rpc/etc/travel.yaml'],
  ['travel-api', 'app/travel/cmd/api/etc/travel.yaml'],
  ['payment-rpc', 'app/payment/cmd/rpc/etc/payment.yaml'],
  ['payment-api', 'app/payment/cmd/api/etc/payment.yaml'],
  ['order-rpc', 'app/order/cmd/rpc/etc/order.yaml'],
  ['order-api', 'app/order/cmd/api/etc/order.yaml'],
  ['order-mq', 'app/order/cmd/mq/etc/order.yaml'],
  ['admin-api', 'app/admin/cmd/api/etc/admin.yaml'],
  ['aiagent-api', 'app/aiagent/cmd/api/etc/aiagent.yaml'],
  ['mqueue-scheduler', 'app/mqueue/cmd/scheduler/etc/mqueue.yaml'],
  ['mqueue-job', 'app/mqueue/cmd/job/etc/mqueue.yaml'],
]

fs.mkdirSync(OUT, { recursive: true })

// 密钥只存在本机：源配置里的 ApiKey 留空，这里用环境变量补进来。
// 这样 deploy/local-dev/（已 gitignore）有真 key，仓库里的 etc/*.yaml 永远没有。
const AI_KEY = process.env.AI_AGENT_API_KEY || process.env.DEEPSEEK_API_KEY || ''

const written = []
for (const [name, rel] of TARGETS) {
  const src = path.join(APP, rel.replace(/^app\//, ''))
  if (!fs.existsSync(src)) {
    console.warn(`skip ${name}: ${src} not found`)
    continue
  }
  let content = fs.readFileSync(src, 'utf8')
  for (const [re, to] of REPLACEMENTS) {
    content = content.replace(re, to)
  }
  if (AI_KEY && /^\s*apiKey:\s*(""|'')?\s*$/im.test(content)) {
    content = content.replace(/^(\s*)apikey:\s*(""|'')?\s*$/im, `$1ApiKey: "${AI_KEY}"`)
  }
  const header = `# 本机开发配置（由 deploy/script/dev/gen-local-config.mjs 生成，请勿手改）\n# 源文件: ${rel}\n`
  fs.writeFileSync(path.join(OUT, `${name}.yaml`), header + content, 'utf8')
  written.push(name)
}

console.log(`generated ${written.length} local configs -> ${OUT}`)
console.log(written.join(', '))
if (AI_KEY) {
  console.log('已把 AI_AGENT_API_KEY 注入 aiagent-api.yaml（本机配置，不会进仓库）')
} else {
  console.log('未检测到 AI_AGENT_API_KEY：aiagent-api 会以 mock 模式启动（功能完整，无需密钥）')
}
