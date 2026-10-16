/**
 * 用本机 Chrome（无头 + CDP）给 README 抓真实界面截图。
 *
 * 前置：后端 13 个服务已启动（run-local.ps1），web(5173) / admin(5174) 已 npm run dev。
 * 用法：node deploy/script/dev/shot-readme.mjs
 * 产物：doc/screenshots/*.png
 *
 * 之所以不用 puppeteer：本机 Node 22 自带 WebSocket，直接用 CDP 最省依赖。
 */
import { spawn } from 'node:child_process'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const ROOT = path.resolve(__dirname, '../../..')
const OUT = path.join(ROOT, 'doc/screenshots')
const PORT = Number(process.env.CDP_PORT || 9333)

const CANDIDATES = [
  process.env.CHROME_PATH,
  'C:/Program Files/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  'C:/Program Files/Microsoft/Edge/Application/msedge.exe',
].filter(Boolean)
const CHROME = CANDIDATES.find((p) => fs.existsSync(p))
if (!CHROME) throw new Error('没找到 Chrome/Edge，可用 CHROME_PATH 指定')

const WEB = 'http://localhost:5173'
const ADMIN = 'http://localhost:5174'

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

// ---------------------------------------------------------------- CDP 客户端
async function httpJson(url, method = 'GET') {
  const res = await fetch(url, { method })
  if (!res.ok) throw new Error(`${method} ${url} -> ${res.status}`)
  return res.json()
}

class Cdp {
  constructor(ws) {
    this.ws = ws
    this.seq = 0
    this.waiting = new Map()
    ws.onmessage = (ev) => {
      const msg = JSON.parse(ev.data)
      if (msg.id && this.waiting.has(msg.id)) {
        const { resolve, reject } = this.waiting.get(msg.id)
        this.waiting.delete(msg.id)
        msg.error ? reject(new Error(JSON.stringify(msg.error))) : resolve(msg.result)
      }
    }
  }
  send(method, params = {}) {
    const id = ++this.seq
    return new Promise((resolve, reject) => {
      this.waiting.set(id, { resolve, reject })
      this.ws.send(JSON.stringify({ id, method, params }))
      setTimeout(() => {
        if (this.waiting.has(id)) {
          this.waiting.delete(id)
          reject(new Error(`CDP timeout: ${method}`))
        }
      }, 60000)
    })
  }
  async js(expression, awaitPromise = false) {
    const r = await this.send('Runtime.evaluate', {
      expression,
      awaitPromise,
      returnByValue: true,
    })
    if (r.exceptionDetails) {
      throw new Error('JS 执行失败: ' + (r.exceptionDetails.exception?.description || r.exceptionDetails.text))
    }
    return r.result?.value
  }
  async goto(url, settle = 2600) {
    await this.send('Page.navigate', { url })
    for (let i = 0; i < 60; i++) {
      await sleep(250)
      try {
        if ((await this.js('document.readyState')) === 'complete') break
      } catch {}
    }
    await sleep(settle)
  }
  async shot(name) {
    const { data } = await this.send('Page.captureScreenshot', { format: 'png' })
    const file = path.join(OUT, name)
    fs.writeFileSync(file, Buffer.from(data, 'base64'))
    const kb = Math.round(fs.statSync(file).size / 1024)
    console.log(`  ✔ ${name}  ${kb} KB`)
  }
}

async function newPage({ width, height, mobile, scale = 2 }) {
  const target = await httpJson(`http://127.0.0.1:${PORT}/json/new?about:blank`, 'PUT')
  const ws = new WebSocket(target.webSocketDebuggerUrl)
  await new Promise((resolve, reject) => {
    ws.onopen = resolve
    ws.onerror = (e) => reject(new Error('WS 连接失败'))
  })
  const cdp = new Cdp(ws)
  await cdp.send('Page.enable')
  await cdp.send('Runtime.enable')
  await cdp.send('Emulation.setDeviceMetricsOverride', {
    width,
    height,
    deviceScaleFactor: scale,
    mobile: !!mobile,
  })
  cdp.targetId = target.id
  return cdp
}

// ---------------------------------------------------------------- 交互小工具
const loginWeb = (cdp, mobile = '13800000001', password = '123456') =>
  cdp.js(
    `(async () => {
      const r = await fetch('/usercenter/v1/user/login', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mobile: '${mobile}', password: '${password}' }),
      });
      const j = await r.json();
      if (j.code !== 200) return 'login failed: ' + j.msg;
      localStorage.setItem('looklook_token', j.data.accessToken);
      return 'ok';
    })()`,
    true,
  )

const loginAdmin = (cdp, username = 'admin', password = 'admin123') =>
  cdp.js(
    `(async () => {
      const r = await fetch('/admin/v1/admin/login', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: '${username}', password: '${password}' }),
      });
      const j = await r.json();
      if (j.code !== 200) return 'login failed: ' + j.msg;
      localStorage.setItem('looklook_admin_token', j.data.accessToken);
      return 'ok';
    })()`,
    true,
  )

/** 给 textarea/input 赋值并触发 Vue 的 v-model，再按 Enter 发送 */
const typeAndSend = (cdp, placeholderPrefix, text) =>
  cdp.js(`(() => {
    const ta = document.querySelector('textarea[placeholder^="${placeholderPrefix}"]');
    if (!ta) return 'no textarea';
    const setter = Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, 'value').set;
    setter.call(ta, ${JSON.stringify(text)});
    ta.dispatchEvent(new Event('input', { bubbles: true }));
    ta.focus();
    ta.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    return 'sent';
  })()`)

const clickByText = (cdp, text, tag = 'button') =>
  cdp.js(`(() => {
    const el = [...document.querySelectorAll('${tag}')].find((b) => (b.innerText || '').includes(${JSON.stringify(text)}));
    if (!el) return 'not found: ${text}';
    el.click();
    return 'clicked';
  })()`)

/** 等 AI 停止输出：sendingSelector 存在 = 还在流式输出 */
async function waitAiDone(cdp, sendingSelector, timeoutMs = 90000) {
  const start = Date.now()
  while (Date.now() - start < timeoutMs) {
    await sleep(600)
    const sending = await cdp.js(`!!document.querySelector(${JSON.stringify(sendingSelector)})`)
    if (!sending) break
  }
  await sleep(1200)
}

// ---------------------------------------------------------------- 主流程
const chrome = spawn(
  CHROME,
  [
    '--headless=new',
    `--remote-debugging-port=${PORT}`,
    `--user-data-dir=${path.join(os.tmpdir(), 'looklook-shot-profile')}`,
    '--no-first-run',
    '--no-default-browser-check',
    '--disable-gpu',
    '--hide-scrollbars',
    '--force-color-profile=srgb',
    'about:blank',
  ],
  { stdio: 'ignore' },
)

let ok = false
try {
  for (let i = 0; i < 60; i++) {
    try {
      await httpJson(`http://127.0.0.1:${PORT}/json/version`)
      break
    } catch {
      await sleep(300)
    }
  }
  fs.mkdirSync(OUT, { recursive: true })
  console.log(`Chrome: ${CHROME}\n输出目录: ${OUT}\n`)

  // ---------- C 端 H5（手机视口 430x932）----------
  console.log('C 端 H5')
  const web = await newPage({ width: 430, height: 932, mobile: true })
  await web.goto(WEB)
  console.log('  登录:', await loginWeb(web))
  await web.goto(`${WEB}/`)
  await web.shot('web-home.png')

  await web.goto(`${WEB}/search?city=上海`)
  await web.shot('web-search.png')

  // 详情页（带日期区间，展示日历/价格）
  await web.goto(`${WEB}/`)
  let detailId = await web.js(`(() => {
    const a = document.querySelector('a[href*="/detail/"]');
    if (!a) return null;
    const m = a.getAttribute('href').match(/\\/detail\\/(\\d+)/);
    return m ? m[1] : null;
  })()`)
  if (!detailId) {
    // 首页卡片是 router.push，没有 <a>；直接问后端要一个 id
    detailId = await web.js(
      `(async () => {
        const r = await fetch('/travel/v1/homestay/homestayList', {
          method: 'POST', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ page: 1, pageSize: 1 }),
        });
        const j = await r.json();
        return j?.data?.list?.[0]?.id ?? null;
      })()`,
      true,
    )
  }
  if (detailId) {
    await web.goto(`${WEB}/detail/${detailId}`, 3400)
    await web.shot('web-detail.png')
  } else {
    console.log('  跳过详情页：拿不到房源 id')
  }

  // 小栖对话
  await web.goto(`${WEB}/`)
  await web.js(`document.querySelector('button[aria-label="打开智能助手"]')?.click()`)
  await sleep(900)
  await web.js(`document.querySelector('button[aria-label="去登录"]')?.click()`)
  await sleep(1200)
  console.log('  提问:', await typeAndSend(web, '问点什么', '帮我看看我的订单，有没有快到期的？'))
  await waitAiDone(web, 'button[aria-label="停止"]')
  await web.shot('web-ai-chat.png')

  // 写操作二次确认卡片（下单页 / 取消订单确认）
  console.log('  提问:', await typeAndSend(web, '问点什么', '帮我取消最近的一笔订单'))
  await waitAiDone(web, 'button[aria-label="停止"]')
  await web.shot('web-ai-confirm.png')

  // ---------- 管理端（桌面视口 1440x900）----------
  console.log('管理端')
  const admin = await newPage({ width: 1440, height: 900, mobile: false, scale: 1 })
  await admin.goto(ADMIN)
  console.log('  登录:', await loginAdmin(admin))
  await admin.goto(`${ADMIN}/`, 3200)
  await admin.shot('admin-dashboard.png')

  // AI 值守：先点「立即巡检」把告警跑出来，再截图
  await admin.goto(`${ADMIN}/monitor`, 2800)
  console.log('  巡检:', await clickByText(admin, '立即巡检'))
  await sleep(9000)
  await admin.shot('admin-monitor.png')
  console.log('  日报:', await clickByText(admin, '经营日报'))
  await sleep(900)
  const gen = await clickByText(admin, '生成日报')
  if (gen === 'clicked') await sleep(12000)
  else await sleep(2500)
  await admin.shot('admin-report.png')

  // 运营 Copilot：任意页面唤起
  await admin.goto(`${ADMIN}/`, 2600)
  await admin.js(`[...document.querySelectorAll('button')].find(b => (b.getAttribute('title')||'').includes('唤起 AI 助手'))?.click()`)
  await sleep(1000)
  console.log('  提问:', await typeAndSend(admin, '问点什么', '最近一周订单量和取消率怎么样？'))
  await waitAiDone(admin, 'button[title="停止"]')
  await admin.shot('admin-copilot.png')

  ok = true
} finally {
  chrome.kill()
  if (!ok) console.error('截图流程中断')
}
console.log('\n完成，图片在 doc/screenshots/')
