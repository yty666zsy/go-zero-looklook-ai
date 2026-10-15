-- =============================================================
-- looklook AI Agent 服务（aiagent）建表脚本
-- 目标库：looklook_aiagent
-- 可重复执行（全部 CREATE TABLE IF NOT EXISTS）
--
-- 表设计说明：
--   ai_conversation  会话（C 端按 userId、管理端按 adminId 隔离）
--   ai_message       消息流水（含工具调用步骤与结构化卡片，便于前端回放）
--   ai_knowledge     知识库（政策/FAQ/运营手册，供 RAG 检索）
--   ai_action_token  写操作待确认令牌（Agent 不直接执行危险动作）
--   ai_tool_audit    工具调用审计（管理端写操作留痕）
--   ai_report        经营日报（无人值守生成，含指标卡片与执行步骤）
--   ai_alert         异常巡检告警（可一键处置，处置仍走二次确认）
-- =============================================================

SET NAMES utf8mb4;

-- ---------- 1) 会话 ----------
CREATE TABLE IF NOT EXISTS `ai_conversation` (
  `id`           bigint   NOT NULL AUTO_INCREMENT,
  `create_time`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `scene`        varchar(16) NOT NULL DEFAULT 'c' COMMENT '场景 c:C端用户 admin:运营管理端',
  `owner_id`     bigint   NOT NULL DEFAULT 0 COMMENT 'C端=userId，管理端=adminId',
  `title`        varchar(128) NOT NULL DEFAULT '' COMMENT '会话标题（取首条提问）',
  `msg_count`    int      NOT NULL DEFAULT 0 COMMENT '消息数',
  `last_msg`     varchar(255) NOT NULL DEFAULT '' COMMENT '最后一条消息摘要',
  `del_state`    tinyint  NOT NULL DEFAULT 0 COMMENT '0:正常 1:已删除',
  PRIMARY KEY (`id`),
  KEY `idx_owner` (`scene`, `owner_id`, `del_state`, `id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'AI 会话';

-- ---------- 2) 消息 ----------
CREATE TABLE IF NOT EXISTS `ai_message` (
  `id`              bigint   NOT NULL AUTO_INCREMENT,
  `create_time`     datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `conversation_id` bigint   NOT NULL DEFAULT 0,
  `scene`           varchar(16) NOT NULL DEFAULT 'c',
  `owner_id`        bigint   NOT NULL DEFAULT 0,
  `role`            varchar(16) NOT NULL DEFAULT '' COMMENT 'user/assistant',
  `content`         mediumtext,
  `steps`           json     DEFAULT NULL COMMENT '工具调用步骤，前端可回放思考过程',
  `cards`           json     DEFAULT NULL COMMENT '结构化卡片（房源/订单/图表）',
  `model`           varchar(64) NOT NULL DEFAULT '' COMMENT '实际使用的模型',
  `prompt_tokens`   int      NOT NULL DEFAULT 0,
  `completion_tokens` int    NOT NULL DEFAULT 0,
  `latency_ms`      int      NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_conv` (`conversation_id`, `id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'AI 消息';

-- ---------- 3) 知识库 ----------
CREATE TABLE IF NOT EXISTS `ai_knowledge` (
  `id`           bigint   NOT NULL AUTO_INCREMENT,
  `create_time`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `scene`        varchar(16) NOT NULL DEFAULT 'all' COMMENT '生效场景 all/c/admin',
  `category`     varchar(32) NOT NULL DEFAULT '' COMMENT '分类',
  `title`        varchar(128) NOT NULL DEFAULT '',
  `content`      text     NOT NULL,
  `keywords`     varchar(255) NOT NULL DEFAULT '' COMMENT '检索加权关键词，逗号分隔',
  `enabled`      tinyint  NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_title` (`title`),
  KEY `idx_scene` (`scene`, `enabled`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'AI 知识库（RAG 语料）';

-- ---------- 4) 待确认动作令牌 ----------
CREATE TABLE IF NOT EXISTS `ai_action_token` (
  `token`           varchar(64) NOT NULL,
  `create_time`     datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `scene`           varchar(16) NOT NULL DEFAULT '',
  `owner_id`        bigint   NOT NULL DEFAULT 0,
  `conversation_id` bigint   NOT NULL DEFAULT 0,
  `tool`            varchar(64) NOT NULL DEFAULT '',
  `args`            json     DEFAULT NULL,
  `summary`         varchar(255) NOT NULL DEFAULT '' COMMENT '给用户看的确认文案',
  `status`          tinyint  NOT NULL DEFAULT 0 COMMENT '0待确认 1已执行 2已取消 3已过期',
  `expire_time`     datetime NOT NULL,
  PRIMARY KEY (`token`),
  KEY `idx_owner` (`scene`, `owner_id`, `status`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'AI 写操作待确认令牌';

-- ---------- 5) 工具调用审计 ----------
CREATE TABLE IF NOT EXISTS `ai_tool_audit` (
  `id`              bigint   NOT NULL AUTO_INCREMENT,
  `create_time`     datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `conversation_id` bigint   NOT NULL DEFAULT 0,
  `scene`           varchar(16) NOT NULL DEFAULT '',
  `owner_id`        bigint   NOT NULL DEFAULT 0,
  `tool`            varchar(64) NOT NULL DEFAULT '',
  `args`            json     DEFAULT NULL,
  `result`          text,
  `status`          varchar(16) NOT NULL DEFAULT 'ok' COMMENT 'ok/fail/pending/rejected',
  `cost_ms`         int      NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_owner` (`scene`, `owner_id`, `id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'AI 工具调用审计';

-- ---------- 6) 经营日报（主动型 Agent 产物） ----------
-- 日报不是「让人去问」，而是服务按节奏自己跑完一组工具、给出结论与建议，
-- 所以它同时存了「结论正文」「指标卡片」「执行步骤」三部分：
--   content/cards  给运营看结论
--   steps          给运营看 Agent 是怎么得出来的（可追溯、可复现）
CREATE TABLE IF NOT EXISTS `ai_report` (
  `id`           bigint   NOT NULL AUTO_INCREMENT,
  `create_time`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `report_type`  varchar(16) NOT NULL DEFAULT 'daily' COMMENT 'daily 日报 / weekly 周报',
  `report_date`  varchar(16) NOT NULL DEFAULT '' COMMENT '归属日期 YYYY-MM-DD',
  `title`        varchar(128) NOT NULL DEFAULT '',
  `summary`      varchar(500) NOT NULL DEFAULT '' COMMENT '一句话摘要（列表用）',
  `content`      mediumtext COMMENT '正文结论与建议',
  `metrics`      json DEFAULT NULL COMMENT '核心指标',
  `cards`        json DEFAULT NULL COMMENT '结构化卡片（趋势/分布/订单），复用对话卡片渲染',
  `steps`        json DEFAULT NULL COMMENT 'Agent 执行步骤',
  `highlights`   json DEFAULT NULL COMMENT '关键结论列表',
  `model`        varchar(64) NOT NULL DEFAULT '' COMMENT '润色用的模型，mock 表示未润色',
  `generated_by` varchar(32) NOT NULL DEFAULT 'cron' COMMENT 'cron 定时 / manual 手动',
  `latency_ms`   int      NOT NULL DEFAULT 0,
  `del_state`    tinyint  NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_type_date` (`report_type`, `report_date`),
  KEY `idx_list` (`del_state`, `id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'AI 经营日报';

-- ---------- 7) 异常巡检告警（主动型 Agent 产物） ----------
-- 巡检规则是确定性的（不依赖大模型），每条告警带一个「可一键处置」的写工具，
-- 但处置动作仍然走 ai_action_token 二次确认，Agent 不会自行改业务数据。
--   status 0 待处理 / 1 已处置 / 2 已忽略 / 3 已消除（下次巡检发现异常自己消失了）
CREATE TABLE IF NOT EXISTS `ai_alert` (
  `id`            bigint   NOT NULL AUTO_INCREMENT,
  `create_time`   datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time`   datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `alert_key`     varchar(96) NOT NULL COMMENT '去重键：规则+业务对象',
  `rule`          varchar(48) NOT NULL DEFAULT '' COMMENT '规则标识',
  `level`         varchar(8)  NOT NULL DEFAULT 'warn' COMMENT 'info/warn/error',
  `category`      varchar(32) NOT NULL DEFAULT '' COMMENT 'order/homestay/user/traffic',
  `title`         varchar(160) NOT NULL DEFAULT '',
  `detail`        text,
  `suggestion`    varchar(500) NOT NULL DEFAULT '' COMMENT '给运营的处理建议',
  `action_tool`   varchar(48) NOT NULL DEFAULT '' COMMENT '可一键处置的写工具',
  `action_args`   json DEFAULT NULL COMMENT '写工具入参',
  `action_text`   varchar(255) NOT NULL DEFAULT '' COMMENT '处置动作的确认文案',
  `related_type`  varchar(24) NOT NULL DEFAULT '' COMMENT 'order/homestay/user',
  `related_id`    varchar(64) NOT NULL DEFAULT '' COMMENT '业务对象 id 或订单号',
  `status`        tinyint NOT NULL DEFAULT 0 COMMENT '0待处理 1已处置 2已忽略 3已消除',
  `hit_count`     int     NOT NULL DEFAULT 1 COMMENT '累计命中次数',
  `batch`         varchar(32) NOT NULL DEFAULT '' COMMENT '产生该告警的巡检批次',
  `last_hit_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_alert_key` (`alert_key`),
  KEY `idx_status` (`status`, `level`, `id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'AI 异常巡检告警';
