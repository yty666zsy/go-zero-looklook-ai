-- =============================================================
-- looklook 后台管理端：管理员表 + user.status 字段
-- 目标库：looklook_usercenter
-- 可重复执行（幂等）
--
-- 演示账号（密码为 md5）：
--   admin    / admin123     超级管理员
--   operator / operator123  运营
-- =============================================================

SET NAMES utf8mb4;

-- ---------- 1) user 表补 status（1:正常 0:禁用） ----------
SET @has_status := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user' AND COLUMN_NAME = 'status'
);
SET @ddl := IF(@has_status = 0,
  'ALTER TABLE `user` ADD COLUMN `status` tinyint NOT NULL DEFAULT 1 COMMENT ''1:正常 0:禁用'' AFTER `info`',
  'DO 0');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ---------- 2) 管理员表（和 C 端 user 表隔离） ----------
CREATE TABLE IF NOT EXISTS `admin_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `delete_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
  `del_state` tinyint NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '版本号',
  `username` varchar(64) NOT NULL DEFAULT '' COMMENT '登录名',
  `password` varchar(64) NOT NULL DEFAULT '' COMMENT 'md5 密码',
  `nickname` varchar(64) NOT NULL DEFAULT '' COMMENT '昵称',
  `role` varchar(32) NOT NULL DEFAULT 'operator' COMMENT '角色 super/operator',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '1:正常 0:禁用',
  `last_login_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '后台管理员';

-- ---------- 3) 演示管理员 ----------
INSERT INTO `admin_user`
  (`id`, `create_time`, `update_time`, `delete_time`, `del_state`, `version`,
   `username`, `password`, `nickname`, `role`, `status`, `last_login_time`)
VALUES
  (1, NOW(), NOW(), '1970-01-01 00:00:00', 0, 0,
   'admin', '0192023a7bbd73250516f069df18b500', '超级管理员', 'super', 1, '1970-01-01 00:00:00'),
  (2, NOW(), NOW(), '1970-01-01 00:00:00', 0, 0,
   'operator', '2407bd807d6ca01d1bcd766c730cec9a', '运营小助手', 'operator', 1, '1970-01-01 00:00:00')
ON DUPLICATE KEY UPDATE
  `password` = VALUES(`password`),
  `nickname` = VALUES(`nickname`),
  `role`     = VALUES(`role`),
  `status`   = VALUES(`status`);
