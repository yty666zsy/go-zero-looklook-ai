/*
 Navicat MySQL Data Transfer

 Source Server         : looklook
 Source Server Type    : MySQL
 Source Server Version : 80028
 Source Host           : 127.0.0.1:33069
 Source Schema         : looklook_order

 Target Server Type    : MySQL
 Target Server Version : 80028
 File Encoding         : 65001

 Date: 10/03/2022 17:15:38
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for homestay_order
-- ----------------------------
DROP TABLE IF EXISTS `homestay_order`;
CREATE TABLE `homestay_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `delete_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_state` tinyint NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '版本号',
  `sn` char(25) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '订单号',
  `user_id` bigint NOT NULL DEFAULT '0' COMMENT '下单用户id',
  `homestay_id` bigint NOT NULL DEFAULT '0' COMMENT '民宿id',
  `title` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '标题',
  `sub_title` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '副标题',
  `cover` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '封面',
  `info` varchar(4069) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '介绍',
  `people_num` tinyint(1) NOT NULL DEFAULT '0' COMMENT '容纳人的数量',
  `row_type` tinyint(1) NOT NULL DEFAULT '0' COMMENT '售卖类型0：按房间出售 1:按人次出售',
  `need_food` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0:不需要餐食 1:需要参数',
  `food_info` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '餐食标准',
  `food_price` bigint NOT NULL COMMENT '餐食价格(分)',
  `homestay_price` bigint NOT NULL COMMENT '民宿价格(分)',
  `market_homestay_price` bigint NOT NULL DEFAULT '0' COMMENT '民宿市场价格(分)',
  `homestay_business_id` bigint NOT NULL DEFAULT '0' COMMENT '店铺id',
  `homestay_user_id` bigint NOT NULL DEFAULT '0' COMMENT '店铺房东id',
  `live_start_date` date NOT NULL COMMENT '开始入住日期',
  `live_end_date` date NOT NULL COMMENT '结束入住日期',
  `live_people_num` tinyint NOT NULL DEFAULT '0' COMMENT '实际入住人数',
  `trade_state` tinyint(1) NOT NULL DEFAULT '0' COMMENT '-1: 已取消 0:待支付 1:未使用 2:已使用  3:已退款 4:已过期',
  `trade_code` char(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '确认码',
  `remark` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户下单备注',
  `order_total_price` bigint NOT NULL DEFAULT '0' COMMENT '订单总价格（餐食总价格+民宿总价格）(分)',
  `food_total_price` bigint NOT NULL DEFAULT '0' COMMENT '餐食总价格(分)',
  `homestay_total_price` bigint NOT NULL DEFAULT '0' COMMENT '民宿总价格(分)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_sn` (`sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='每一间民宿';

SET FOREIGN_KEY_CHECKS = 1;

-- ----------------------------
-- 房源占用日历（高并发防超卖）
-- 逐日一行，(homestay_id, occupy_date) 唯一：民宿是整套房源，同一房源同一晚只能卖一次。
-- 下单：在房源粒度分布式锁里占位（和订单同一个事务），任一晚冲突就整单失败；
-- 释放：订单取消 / 退款 / 过期时删掉对应行，日期重新可订。
-- 唯一键是最终兜底：即使 Redis 锁失效或被绕过，数据库层面也不会超卖。
-- ----------------------------
DROP TABLE IF EXISTS `homestay_calendar`;
CREATE TABLE `homestay_calendar` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `homestay_id` bigint NOT NULL COMMENT '房源id',
  `occupy_date` date NOT NULL COMMENT '占用的那一晚',
  `order_sn` varchar(32) NOT NULL DEFAULT '' COMMENT '占用它的订单号',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_homestay_date` (`homestay_id`,`occupy_date`),
  KEY `idx_sn` (`order_sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='房源日历占用（防超卖）';

-- 回填：把现存的有效订单（待支付/未使用/已使用）展开成逐日占用，离店当天不占用。
-- 已有历史冲突订单用 INSERT IGNORE 跳过（保留先来的那一笔）。
INSERT IGNORE INTO `homestay_calendar` (`homestay_id`,`occupy_date`,`order_sn`)
SELECT `homestay_id`, `occupy_date`, `order_sn` FROM (
  WITH RECURSIVE `days` AS (
    SELECT `homestay_id` AS hid, `sn` AS osn, DATE(`live_start_date`) AS `d`, DATE(`live_end_date`) AS `e`
      FROM `homestay_order`
     WHERE `del_state` = 0 AND `trade_state` IN (0,1,2) AND `live_end_date` > `live_start_date`
    UNION ALL
    SELECT `hid`, `osn`, `d` + INTERVAL 1 DAY, `e` FROM `days` WHERE `d` + INTERVAL 1 DAY < `e`
  )
  SELECT `hid` AS `homestay_id`, `d` AS `occupy_date`, `osn` AS `order_sn` FROM `days`
) AS `occupied`;

