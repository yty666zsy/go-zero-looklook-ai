// 生成演示种子数据
//   deploy/sql/looklook_seed_travel.sql      -> 导入 looklook_travel
//   deploy/sql/looklook_seed_usercenter.sql  -> 导入 looklook_usercenter
// 用法: node deploy/script/seed/main.mjs
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const OUT_TRAVEL = path.resolve(__dirname, '../../sql/looklook_seed_travel.sql')
const OUT_USER = path.resolve(__dirname, '../../sql/looklook_seed_usercenter.sql')

const MD5_123456 = 'e10adc3949ba59abbe56e057f20f883e' // md5("123456")
const PICSUM = (seed, w = 900, h = 640) => `https://picsum.photos/seed/${seed}/${w}/${h}`

const cities = [
  { code: 'beijing', name: '北京', lng: 116.4074, lat: 39.9042, areas: ['东城区', '西城区', '朝阳区', '海淀区', '怀柔区'] },
  { code: 'shanghai', name: '上海', lng: 121.4737, lat: 31.2304, areas: ['黄浦区', '静安区', '徐汇区', '浦东新区', '青浦区'] },
  { code: 'hangzhou', name: '杭州', lng: 120.1551, lat: 30.2741, areas: ['西湖区', '上城区', '拱墅区', '滨江区', '桐庐县'] },
  { code: 'chengdu', name: '成都', lng: 104.0668, lat: 30.5728, areas: ['锦江区', '青羊区', '武侯区', '成华区', '都江堰市'] },
  { code: 'xiamen', name: '厦门', lng: 118.0894, lat: 24.4798, areas: ['思明区', '湖里区', '集美区', '同安区', '环岛路'] },
  { code: 'lijiang', name: '丽江', lng: 100.2299, lat: 26.8721, areas: ['古城区', '束河古镇', '白沙古镇', '玉龙县', '泸沽湖'] },
  { code: 'sanya', name: '三亚', lng: 109.5119, lat: 18.2528, areas: ['海棠湾', '亚龙湾', '大东海', '三亚湾', '崖州区'] },
  { code: 'dali', name: '大理', lng: 100.2676, lat: 25.6065, areas: ['大理古城', '洱海东岸', '双廊镇', '喜洲镇', '苍山脚下'] },
]

const brands = [
  '云栖', '栖溪', '栖岸', '山海间', '慢时光', '拾光', '半山', '寻野',
  '初见', '未央', '南栀', '屿见', '序章', '听风', '青苔', '白屿',
  '霁月', '玖月', '长亭', '拾贝', '空谷', '岑谷', '慢屿', '澄舍',
  '浮生', '木白', '砚山', '小隐', '停云', '青庐', '有间', '风吟',
  '陌上', '见山', '望野', '归屿', '简山', '柒月', '屿北', '云庐',
]

const roomTypes = ['山景大床房', '海景套房', '庭院小院', '设计师loft', '湖景复式', '露台亲子房']

const tagPool = [
  ['近地铁', '可做饭', '洗衣机'],
  ['免费停车', '含早餐', '宠物友好'],
  ['独栋小院', '私汤温泉', '观景露台'],
  ['近景区', '24h前台', '行李寄存'],
  ['设计感', '大落地窗', '投影仪'],
  ['海景房', '无边泳池', '免费接机'],
]

const foodInfos = ['含双早｜本地特色早餐', '含三餐｜管家私房菜', '含双早｜现做手冲咖啡', '不含餐｜周边餐饮丰富']

const infoTpl = (city, area, name) =>
  `「${name}」位于${city}${area}，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。` +
  `房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。`

const nicknames = ['林小满', '周予安', '陈屿', '苏晚', '沈知白', '许一诺', '顾清和', '江眠', '温简', '叶疏影', '程屿', '白鹿', '宋辞', '南栀', '陆时', '何夕']
const commentPool = [
  '位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。',
  '窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。',
  '房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。',
  '房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。',
  '性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。',
  '入住流程很顺畅，自助入住不用等，管家响应也很快。',
  '院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。',
  '洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。',
  '带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。',
  '早餐是现做的，分量足也好吃，比酒店自助舒服。',
  '空调给力，热水很快，洗澡体验很好。',
  '设计感很强，拍了很多照片，朋友都在问链接。',
  '停车方便，自驾过来完全不用愁，车位就在门口。',
  '房间采光太棒了，白天基本不用开灯，很通透。',
  '整体超出预期，服务态度好，退房也很快，没有套路。',
  '唯一小建议是希望能加个全身镜，其他都很满意。',
  '安静、干净、舒服，适合度假躺平，什么都不干也很好。',
  '前台小哥帮我们搬了行李，还送了水果，服务很nice。',
  '离景区很近，早起去看日出非常方便，省了打车钱。',
  '床垫软硬适中，枕头有高低两种可选，睡眠质量很好。',
]

// 简单可复现随机数
let _s = 20241014
const rnd = () => ((_s = (_s * 1103515245 + 12345) & 0x7fffffff) / 0x7fffffff)
const pick = (a) => a[Math.floor(rnd() * a.length)]
const int = (min, max) => min + Math.floor(rnd() * (max - min + 1))

const q = (s) => `'${String(s).replace(/'/g, "''")}'`

const lines = []
lines.push('-- =============================================================')
lines.push('-- looklook 民宿演示数据 / demo seed data for `looklook_travel`')
lines.push('-- 由 deploy/script/seed/main.mjs 生成，请勿手改')
lines.push('-- 前置：先执行 looklook_travel.sql 建表，再导入本文件')
lines.push('-- 可重复执行（schema 变更为幂等）')
lines.push('-- =============================================================')
lines.push('')
lines.push('SET NAMES utf8mb4;')
lines.push('')

// ---------- travel: schema delta ----------
// 用存储过程做「列/索引不存在才添加」，保证种子脚本可以重复执行
lines.push('-- ---------- schema delta: homestay / homestay_comment 增加展示字段 ----------')
lines.push('-- 下面的辅助存储过程保证重复执行不会报 Duplicate column / Duplicate key 错误')
lines.push('DROP PROCEDURE IF EXISTS `__looklook_add_col`;')
lines.push('DROP PROCEDURE IF EXISTS `__looklook_add_idx`;')
lines.push('')
lines.push('DELIMITER $$')
lines.push('CREATE PROCEDURE `__looklook_add_col`(IN p_table VARCHAR(64), IN p_col VARCHAR(64), IN p_ddl TEXT)')
lines.push('BEGIN')
lines.push("  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = p_table AND COLUMN_NAME = p_col) THEN")
lines.push("    SET @__sql = CONCAT('ALTER TABLE `', p_table, '` ADD COLUMN ', p_ddl);")
lines.push('    PREPARE __stmt FROM @__sql; EXECUTE __stmt; DEALLOCATE PREPARE __stmt;')
lines.push('  END IF;')
lines.push('END$$')
lines.push('')
lines.push('CREATE PROCEDURE `__looklook_add_idx`(IN p_table VARCHAR(64), IN p_idx VARCHAR(64), IN p_ddl TEXT)')
lines.push('BEGIN')
lines.push("  IF NOT EXISTS (SELECT 1 FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = p_table AND INDEX_NAME = p_idx) THEN")
lines.push("    SET @__sql = CONCAT('ALTER TABLE `', p_table, '` ADD ', p_ddl);")
lines.push('    PREPARE __stmt FROM @__sql; EXECUTE __stmt; DEALLOCATE PREPARE __stmt;')
lines.push('  END IF;')
lines.push('END$$')
lines.push('DELIMITER ;')
lines.push('')
const col = (t, c, ddl) => `CALL \`__looklook_add_col\`(${q(t)}, ${q(c)}, ${q(ddl)});`
const idx = (t, n, ddl) => `CALL \`__looklook_add_idx\`(${q(t)}, ${q(n)}, ${q(ddl)});`
lines.push(col('homestay', 'city', "`city` varchar(32) NOT NULL DEFAULT '' COMMENT '城市' AFTER `user_id`"))
lines.push(col('homestay', 'address', "`address` varchar(128) NOT NULL DEFAULT '' COMMENT '地址' AFTER `city`"))
lines.push(col('homestay', 'tags', "`tags` varchar(128) NOT NULL DEFAULT '' COMMENT '标签，逗号分隔' AFTER `address`"))
lines.push(col('homestay', 'longitude', "`longitude` double NOT NULL DEFAULT '0' COMMENT '经度' AFTER `tags`"))
lines.push(col('homestay', 'latitude', "`latitude` double NOT NULL DEFAULT '0' COMMENT '纬度' AFTER `longitude`"))
lines.push(col('homestay', 'score', "`score` double(2,1) NOT NULL DEFAULT '5.0' COMMENT '评分' AFTER `latitude`"))
lines.push(col('homestay', 'comment_count', "`comment_count` int NOT NULL DEFAULT '0' COMMENT '评价数' AFTER `score`"))
lines.push(col('homestay_comment', 'nickname', "`nickname` varchar(64) NOT NULL DEFAULT '' COMMENT '昵称冗余' AFTER `user_id`"))
lines.push(col('homestay_comment', 'avatar', "`avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '头像冗余' AFTER `nickname`"))
lines.push('')
lines.push('-- ---------- 索引 ----------')
lines.push(idx('homestay', 'idx_city', 'KEY `idx_city` (`city`,`row_state`,`del_state`) USING BTREE'))
lines.push(idx('homestay_comment', 'idx_homestay', 'KEY `idx_homestay` (`homestay_id`,`del_state`,`id`) USING BTREE'))
lines.push('')
lines.push('DROP PROCEDURE `__looklook_add_col`;')
lines.push('DROP PROCEDURE `__looklook_add_idx`;')
lines.push('')

// ---------- travel: truncate ----------
lines.push('TRUNCATE TABLE `homestay`;')
lines.push('TRUNCATE TABLE `homestay_activity`;')
lines.push('TRUNCATE TABLE `homestay_business`;')
lines.push('TRUNCATE TABLE `homestay_comment`;')
lines.push('')

// ---------- usercenter ----------
const userLines = []
userLines.push('-- =============================================================')
userLines.push('-- looklook 演示用户 / demo users（密码统一 123456）')
userLines.push('-- 由 deploy/script/seed/main.mjs 生成，请勿手改')
userLines.push('-- 前置：先执行 looklook_usercenter.sql')
userLines.push('-- =============================================================')
userLines.push('')
userLines.push('SET NAMES utf8mb4;')
userLines.push('')
userLines.push('TRUNCATE TABLE `user`;')
userLines.push('TRUNCATE TABLE `user_auth`;')
const users = []
for (let i = 1; i <= 8; i++) {
  const mobile = '1380000000' + i
  const nickname = nicknames[i - 1]
  users.push({ id: i, mobile, nickname })
}
userLines.push('INSERT INTO `user` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`version`,`mobile`,`password`,`nickname`,`sex`,`avatar`,`info`) VALUES')
userLines.push(
  users
    .map(
      (u, i) =>
        `(${u.id},NOW(),NOW(),NOW(),0,0,${q(u.mobile)},${q(MD5_123456)},${q(u.nickname)},${i % 2},` +
        `${q(PICSUM('avatar' + u.id, 200, 200))},${q('喜欢到处走走，住遍好看的房子。')})`,
    )
    .join(',\n') + ';',
)
userLines.push('')
userLines.push('INSERT INTO `user_auth` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`version`,`user_id`,`auth_key`,`auth_type`) VALUES')
userLines.push(users.map((u) => `(${u.id},NOW(),NOW(),NOW(),0,0,${u.id},${q(u.mobile)},'system')`).join(',\n') + ';')
userLines.push('')

// ---------- travel business ----------
const businesses = []
cities.forEach((c, i) => {
  businesses.push({
    id: i + 1,
    title: `${c.name}·${pick(brands)}民宿`,
    userId: i + 1,
    city: c.name,
    star: (4.5 + rnd() * 0.5).toFixed(1),
    tags: pick(tagPool).slice(0, 2).join(','),
  })
})

lines.push('-- ---------- 店铺 ----------')
lines.push('INSERT INTO `homestay_business` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`version`,`title`,`user_id`,`info`,`boss_info`,`license_fron`,`license_back`,`row_state`,`star`,`tags`,`cover`,`header_img`) VALUES')
lines.push(
  businesses
    .map(
      (b) =>
        `(${b.id},NOW(),NOW(),NOW(),0,0,${q(b.title)},${b.userId},` +
        `${q(`专注${b.city}本地精品民宿运营，房源均为自营管理，提供 7x24 小时管家服务。`)},` +
        `${q(`${b.city}本地房东，熟悉周边吃喝玩乐，乐意分享私藏路线。`)},` +
        `${q(PICSUM('license' + b.id, 600, 400))},${q(PICSUM('license' + b.id + 'b', 600, 400))},1,${b.star},${q(b.tags)},` +
        `${q(PICSUM('biz' + b.id))},${q(PICSUM('bizh' + b.id, 1200, 500))})`,
    )
    .join(',\n') + ';',
)
lines.push('')

// ---------- homestays ----------
const homestays = []
let hsId = 0
cities.forEach((c, ci) => {
  const biz = businesses[ci]
  for (let k = 0; k < 5; k++) {
    hsId++
    const name = `${pick(brands)}·${roomTypes[k % roomTypes.length]}`
    const biz = businesses[ci]
    const price = int(188, 1588) * 100 + int(0, 9) * 10 // 分（188.0 ~ 1588.9 元/晚）
    const hasFood = rnd() < 0.55
    homestays.push({
      id: hsId,
      title: name,
      subTitle: `${c.name}${c.areas[k]} · 步行可达景区 · ${k % 2 === 0 ? '独立露台' : '大落地窗'}`,
      banner: [PICSUM(`hs${hsId}-1`), PICSUM(`hs${hsId}-2`), PICSUM(`hs${hsId}-3`)].join(','),
      info: infoTpl(c.name, c.areas[k], name),
      peopleNum: int(2, 6),
      bizId: biz.id,
      userId: biz.userId,
      city: c.name,
      address: `${c.name}${c.areas[k]}${int(1, 199)}号`,
      tags: pick(tagPool).join(','),
      lng: (c.lng + (rnd() - 0.5) * 0.1).toFixed(6),
      lat: (c.lat + (rnd() - 0.5) * 0.1).toFixed(6),
      score: (4.4 + rnd() * 0.6).toFixed(1),
      foodInfo: hasFood ? foodInfos[k % 3] : foodInfos[3],
      foodPrice: hasFood ? int(68, 168) * 100 : 0,
      price,
      marketPrice: Math.round(price * (1.2 + rnd() * 0.3)),
      tagsFull: pick(tagPool).slice(0, 3).join(','),
    })
  }
})

lines.push('-- ---------- 民宿（40 间 / 8 城） ----------')
lines.push('INSERT INTO `homestay` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`version`,`title`,`sub_title`,`banner`,`info`,`people_num`,`homestay_business_id`,`user_id`,`city`,`address`,`tags`,`longitude`,`latitude`,`score`,`comment_count`,`row_state`,`row_type`,`food_info`,`food_price`,`homestay_price`,`market_homestay_price`) VALUES')
lines.push(
  homestays
    .map(
      (h) =>
        `(${h.id},NOW(),NOW(),NOW(),0,0,${q(h.title)},${q(h.subTitle)},${q(h.banner)},${q(h.info)},${h.peopleNum},${h.bizId},${h.userId},` +
        `${q(h.city)},${q(h.address)},${q(h.tagsFull)},${h.lng},${h.lat},${h.score},0,1,${h.id % 4 === 0 ? 1 : 0},${q(h.foodInfo)},${h.foodPrice},${h.price},${h.marketPrice})`,
    )
    .join(',\n') + ';',
)
lines.push('')

// ---------- comments ----------
const commentRows = []
let cmId = 0
const counts = {}
homestays.forEach((h) => {
  const n = int(3, 12)
  counts[h.id] = n
  for (let i = 0; i < n; i++) {
    cmId++
    const u = users[int(0, users.length - 1)]
    const d = int(1, 180)
    const s1 = int(9, 10) / 2, s2 = int(9, 10) / 2, s3 = int(9, 10) / 2
    commentRows.push({
      id: cmId,
      hsId: h.id,
      userId: u.id,
      nickname: u.nickname,
      avatar: PICSUM('avatar' + u.id, 200, 200),
      content: pick(commentPool),
      star: JSON.stringify([s1, s2, s3]),
      days: d,
    })
  }
})
lines.push(`-- ---------- 评价（${commentRows.length} 条） ----------`)
lines.push('INSERT INTO `homestay_comment` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`homestay_id`,`user_id`,`nickname`,`avatar`,`content`,`star`,`version`) VALUES')
lines.push(
  commentRows
    .map(
      (c) =>
        `(${c.id},DATE_SUB(NOW(), INTERVAL ${c.days} DAY),NOW(),NOW(),0,${c.hsId},${c.userId},${q(c.nickname)},${q(c.avatar)},${q(c.content)},${q(c.star)},0)`,
    )
    .join(',\n') + ';',
)
lines.push('')
lines.push('-- 同步评价数')
lines.push('UPDATE `homestay` SET `comment_count` = CASE `id`')
lines.push(Object.entries(counts).map(([id, n]) => `  WHEN ${id} THEN ${n}`).join('\n'))
lines.push('  ELSE `comment_count` END;')
lines.push('')

// ---------- activities ----------
const pref = homestays.filter((h) => h.id % 5 <= 1).map((h) => h.id)
lines.push('-- ---------- 优选民宿活动位 ----------')
lines.push('INSERT INTO `homestay_activity` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`row_type`,`data_id`,`row_status`,`version`) VALUES')
lines.push(
  pref.map((id, i) => `(${i + 1},NOW(),NOW(),NOW(),0,'preferredHomestay',${id},1,0)`).join(',\n') + ';',
)
lines.push('')
lines.push('-- 房东推荐位（供 goodBoss 使用）')
lines.push('INSERT INTO `homestay_activity` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`row_type`,`data_id`,`row_status`,`version`) VALUES')
lines.push(businesses.map((b, i) => `(${100 + i + 1},NOW(),NOW(),NOW(),0,'goodBusiness',${b.id},1,0)`).join(',\n') + ';')
lines.push('')

fs.mkdirSync(path.dirname(OUT_TRAVEL), { recursive: true })
fs.writeFileSync(OUT_TRAVEL, lines.join('\n'), 'utf8')
fs.writeFileSync(OUT_USER, userLines.join('\n'), 'utf8')
console.log(`generated ${OUT_TRAVEL}`)
console.log(`generated ${OUT_USER}`)
console.log(`homestays=${homestays.length} comments=${commentRows.length} businesses=${businesses.length} users=${users.length}`)
