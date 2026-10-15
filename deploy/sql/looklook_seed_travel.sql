-- =============================================================
-- looklook 民宿演示数据 / demo seed data for `looklook_travel`
-- 由 deploy/script/seed/main.mjs 生成，请勿手改
-- 前置：先执行 looklook_travel.sql 建表，再导入本文件
-- 可重复执行（schema 变更为幂等）
-- =============================================================

SET NAMES utf8mb4;

-- ---------- schema delta: homestay / homestay_comment 增加展示字段 ----------
-- 下面的辅助存储过程保证重复执行不会报 Duplicate column / Duplicate key 错误
DROP PROCEDURE IF EXISTS `__looklook_add_col`;
DROP PROCEDURE IF EXISTS `__looklook_add_idx`;

DELIMITER $$
CREATE PROCEDURE `__looklook_add_col`(IN p_table VARCHAR(64), IN p_col VARCHAR(64), IN p_ddl TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = p_table AND COLUMN_NAME = p_col) THEN
    SET @__sql = CONCAT('ALTER TABLE `', p_table, '` ADD COLUMN ', p_ddl);
    PREPARE __stmt FROM @__sql; EXECUTE __stmt; DEALLOCATE PREPARE __stmt;
  END IF;
END$$

CREATE PROCEDURE `__looklook_add_idx`(IN p_table VARCHAR(64), IN p_idx VARCHAR(64), IN p_ddl TEXT)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = p_table AND INDEX_NAME = p_idx) THEN
    SET @__sql = CONCAT('ALTER TABLE `', p_table, '` ADD ', p_ddl);
    PREPARE __stmt FROM @__sql; EXECUTE __stmt; DEALLOCATE PREPARE __stmt;
  END IF;
END$$
DELIMITER ;

CALL `__looklook_add_col`('homestay', 'city', '`city` varchar(32) NOT NULL DEFAULT '''' COMMENT ''城市'' AFTER `user_id`');
CALL `__looklook_add_col`('homestay', 'address', '`address` varchar(128) NOT NULL DEFAULT '''' COMMENT ''地址'' AFTER `city`');
CALL `__looklook_add_col`('homestay', 'tags', '`tags` varchar(128) NOT NULL DEFAULT '''' COMMENT ''标签，逗号分隔'' AFTER `address`');
CALL `__looklook_add_col`('homestay', 'longitude', '`longitude` double NOT NULL DEFAULT ''0'' COMMENT ''经度'' AFTER `tags`');
CALL `__looklook_add_col`('homestay', 'latitude', '`latitude` double NOT NULL DEFAULT ''0'' COMMENT ''纬度'' AFTER `longitude`');
CALL `__looklook_add_col`('homestay', 'score', '`score` double(2,1) NOT NULL DEFAULT ''5.0'' COMMENT ''评分'' AFTER `latitude`');
CALL `__looklook_add_col`('homestay', 'comment_count', '`comment_count` int NOT NULL DEFAULT ''0'' COMMENT ''评价数'' AFTER `score`');
CALL `__looklook_add_col`('homestay_comment', 'nickname', '`nickname` varchar(64) NOT NULL DEFAULT '''' COMMENT ''昵称冗余'' AFTER `user_id`');
CALL `__looklook_add_col`('homestay_comment', 'avatar', '`avatar` varchar(255) NOT NULL DEFAULT '''' COMMENT ''头像冗余'' AFTER `nickname`');

-- ---------- 索引 ----------
CALL `__looklook_add_idx`('homestay', 'idx_city', 'KEY `idx_city` (`city`,`row_state`,`del_state`) USING BTREE');
CALL `__looklook_add_idx`('homestay_comment', 'idx_homestay', 'KEY `idx_homestay` (`homestay_id`,`del_state`,`id`) USING BTREE');

DROP PROCEDURE `__looklook_add_col`;
DROP PROCEDURE `__looklook_add_idx`;

TRUNCATE TABLE `homestay`;
TRUNCATE TABLE `homestay_activity`;
TRUNCATE TABLE `homestay_business`;
TRUNCATE TABLE `homestay_comment`;

-- ---------- 店铺 ----------
INSERT INTO `homestay_business` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`version`,`title`,`user_id`,`info`,`boss_info`,`license_fron`,`license_back`,`row_state`,`star`,`tags`,`cover`,`header_img`) VALUES
(1,NOW(),NOW(),NOW(),0,0,'北京·山海间民宿',1,'专注北京本地精品民宿运营，房源均为自营管理，提供 7x24 小时管家服务。','北京本地房东，熟悉周边吃喝玩乐，乐意分享私藏路线。','https://picsum.photos/seed/license1/600/400','https://picsum.photos/seed/license1b/600/400',1,4.5,'免费停车,含早餐','https://picsum.photos/seed/biz1/900/640','https://picsum.photos/seed/bizh1/1200/500'),
(2,NOW(),NOW(),NOW(),0,0,'上海·听风民宿',2,'专注上海本地精品民宿运营，房源均为自营管理，提供 7x24 小时管家服务。','上海本地房东，熟悉周边吃喝玩乐，乐意分享私藏路线。','https://picsum.photos/seed/license2/600/400','https://picsum.photos/seed/license2b/600/400',1,4.8,'近地铁,可做饭','https://picsum.photos/seed/biz2/900/640','https://picsum.photos/seed/bizh2/1200/500'),
(3,NOW(),NOW(),NOW(),0,0,'杭州·小隐民宿',3,'专注杭州本地精品民宿运营，房源均为自营管理，提供 7x24 小时管家服务。','杭州本地房东，熟悉周边吃喝玩乐，乐意分享私藏路线。','https://picsum.photos/seed/license3/600/400','https://picsum.photos/seed/license3b/600/400',1,4.7,'海景房,无边泳池','https://picsum.photos/seed/biz3/900/640','https://picsum.photos/seed/bizh3/1200/500'),
(4,NOW(),NOW(),NOW(),0,0,'成都·停云民宿',4,'专注成都本地精品民宿运营，房源均为自营管理，提供 7x24 小时管家服务。','成都本地房东，熟悉周边吃喝玩乐，乐意分享私藏路线。','https://picsum.photos/seed/license4/600/400','https://picsum.photos/seed/license4b/600/400',1,4.5,'独栋小院,私汤温泉','https://picsum.photos/seed/biz4/900/640','https://picsum.photos/seed/bizh4/1200/500'),
(5,NOW(),NOW(),NOW(),0,0,'厦门·云栖民宿',5,'专注厦门本地精品民宿运营，房源均为自营管理，提供 7x24 小时管家服务。','厦门本地房东，熟悉周边吃喝玩乐，乐意分享私藏路线。','https://picsum.photos/seed/license5/600/400','https://picsum.photos/seed/license5b/600/400',1,5.0,'免费停车,含早餐','https://picsum.photos/seed/biz5/900/640','https://picsum.photos/seed/bizh5/1200/500'),
(6,NOW(),NOW(),NOW(),0,0,'丽江·屿北民宿',6,'专注丽江本地精品民宿运营，房源均为自营管理，提供 7x24 小时管家服务。','丽江本地房东，熟悉周边吃喝玩乐，乐意分享私藏路线。','https://picsum.photos/seed/license6/600/400','https://picsum.photos/seed/license6b/600/400',1,4.8,'免费停车,含早餐','https://picsum.photos/seed/biz6/900/640','https://picsum.photos/seed/bizh6/1200/500'),
(7,NOW(),NOW(),NOW(),0,0,'三亚·拾光民宿',7,'专注三亚本地精品民宿运营，房源均为自营管理，提供 7x24 小时管家服务。','三亚本地房东，熟悉周边吃喝玩乐，乐意分享私藏路线。','https://picsum.photos/seed/license7/600/400','https://picsum.photos/seed/license7b/600/400',1,5.0,'近景区,24h前台','https://picsum.photos/seed/biz7/900/640','https://picsum.photos/seed/bizh7/1200/500'),
(8,NOW(),NOW(),NOW(),0,0,'大理·砚山民宿',8,'专注大理本地精品民宿运营，房源均为自营管理，提供 7x24 小时管家服务。','大理本地房东，熟悉周边吃喝玩乐，乐意分享私藏路线。','https://picsum.photos/seed/license8/600/400','https://picsum.photos/seed/license8b/600/400',1,5.0,'近景区,24h前台','https://picsum.photos/seed/biz8/900/640','https://picsum.photos/seed/bizh8/1200/500');

-- ---------- 民宿（40 间 / 8 城） ----------
INSERT INTO `homestay` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`version`,`title`,`sub_title`,`banner`,`info`,`people_num`,`homestay_business_id`,`user_id`,`city`,`address`,`tags`,`longitude`,`latitude`,`score`,`comment_count`,`row_state`,`row_type`,`food_info`,`food_price`,`homestay_price`,`market_homestay_price`) VALUES
(1,NOW(),NOW(),NOW(),0,0,'青苔·山景大床房','北京东城区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs1-1/900/640,https://picsum.photos/seed/hs1-2/900/640,https://picsum.photos/seed/hs1-3/900/640','「青苔·山景大床房」位于北京东城区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',4,1,1,'北京','北京东城区19号','免费停车,含早餐,宠物友好',116.410935,39.857562,4.6,0,1,0,'含双早｜本地特色早餐',8100,106430,132615),
(2,NOW(),NOW(),NOW(),0,0,'听风·海景套房','北京西城区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs2-1/900/640,https://picsum.photos/seed/hs2-2/900/640,https://picsum.photos/seed/hs2-3/900/640','「听风·海景套房」位于北京西城区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',4,1,1,'北京','北京西城区198号','近景区,24h前台,行李寄存',116.422278,39.863653,4.9,0,1,0,'不含餐｜周边餐饮丰富',0,83730,105789),
(3,NOW(),NOW(),NOW(),0,0,'听风·庭院小院','北京朝阳区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs3-1/900/640,https://picsum.photos/seed/hs3-2/900/640,https://picsum.photos/seed/hs3-3/900/640','「听风·庭院小院」位于北京朝阳区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,1,1,'北京','北京朝阳区175号','近景区,24h前台,行李寄存',116.442915,39.921425,4.4,0,1,0,'含双早｜现做手冲咖啡',16200,40730,57780),
(4,NOW(),NOW(),NOW(),0,0,'望野·设计师loft','北京海淀区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs4-1/900/640,https://picsum.photos/seed/hs4-2/900/640,https://picsum.photos/seed/hs4-3/900/640','「望野·设计师loft」位于北京海淀区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,1,1,'北京','北京海淀区178号','独栋小院,私汤温泉,观景露台',116.378682,39.925534,4.9,0,1,1,'含双早｜本地特色早餐',8500,71980,91118),
(5,NOW(),NOW(),NOW(),0,0,'青苔·湖景复式','北京怀柔区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs5-1/900/640,https://picsum.photos/seed/hs5-2/900/640,https://picsum.photos/seed/hs5-3/900/640','「青苔·湖景复式」位于北京怀柔区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,1,1,'北京','北京怀柔区151号','近景区,24h前台,行李寄存',116.397714,39.880576,4.5,0,1,0,'不含餐｜周边餐饮丰富',0,133720,199218),
(6,NOW(),NOW(),NOW(),0,0,'白屿·山景大床房','上海黄浦区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs6-1/900/640,https://picsum.photos/seed/hs6-2/900/640,https://picsum.photos/seed/hs6-3/900/640','「白屿·山景大床房」位于上海黄浦区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',4,2,2,'上海','上海黄浦区44号','近景区,24h前台,行李寄存',121.515289,31.201646,4.5,0,1,0,'含双早｜本地特色早餐',12300,23410,32171),
(7,NOW(),NOW(),NOW(),0,0,'见山·海景套房','上海静安区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs7-1/900/640,https://picsum.photos/seed/hs7-2/900/640,https://picsum.photos/seed/hs7-3/900/640','「见山·海景套房」位于上海静安区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,2,2,'上海','上海静安区39号','设计感,大落地窗,投影仪',121.468468,31.216499,4.5,0,1,0,'含三餐｜管家私房菜',15800,137190,195400),
(8,NOW(),NOW(),NOW(),0,0,'青庐·庭院小院','上海徐汇区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs8-1/900/640,https://picsum.photos/seed/hs8-2/900/640,https://picsum.photos/seed/hs8-3/900/640','「青庐·庭院小院」位于上海徐汇区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',4,2,2,'上海','上海徐汇区187号','设计感,大落地窗,投影仪',121.518135,31.197907,4.9,0,1,1,'含双早｜现做手冲咖啡',12300,72000,101242),
(9,NOW(),NOW(),NOW(),0,0,'浮生·设计师loft','上海浦东新区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs9-1/900/640,https://picsum.photos/seed/hs9-2/900/640,https://picsum.photos/seed/hs9-3/900/640','「浮生·设计师loft」位于上海浦东新区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,2,2,'上海','上海浦东新区91号','近地铁,可做饭,洗衣机',121.514002,31.262427,4.4,0,1,0,'含双早｜本地特色早餐',14300,26030,35720),
(10,NOW(),NOW(),NOW(),0,0,'慢屿·湖景复式','上海青浦区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs10-1/900/640,https://picsum.photos/seed/hs10-2/900/640,https://picsum.photos/seed/hs10-3/900/640','「慢屿·湖景复式」位于上海青浦区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,2,2,'上海','上海青浦区122号','近景区,24h前台,行李寄存',121.498477,31.208046,4.8,0,1,0,'不含餐｜周边餐饮丰富',0,116030,158642),
(11,NOW(),NOW(),NOW(),0,0,'寻野·山景大床房','杭州西湖区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs11-1/900/640,https://picsum.photos/seed/hs11-2/900/640,https://picsum.photos/seed/hs11-3/900/640','「寻野·山景大床房」位于杭州西湖区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',4,3,3,'杭州','杭州西湖区101号','独栋小院,私汤温泉,观景露台',120.107751,30.322975,4.7,0,1,0,'含双早｜本地特色早餐',11100,72020,100251),
(12,NOW(),NOW(),NOW(),0,0,'木白·海景套房','杭州上城区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs12-1/900/640,https://picsum.photos/seed/hs12-2/900/640,https://picsum.photos/seed/hs12-3/900/640','「木白·海景套房」位于杭州上城区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,3,3,'杭州','杭州上城区138号','近景区,24h前台,行李寄存',120.186421,30.297760,4.7,0,1,1,'含三餐｜管家私房菜',13500,134360,195786),
(13,NOW(),NOW(),NOW(),0,0,'浮生·庭院小院','杭州拱墅区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs13-1/900/640,https://picsum.photos/seed/hs13-2/900/640,https://picsum.photos/seed/hs13-3/900/640','「浮生·庭院小院」位于杭州拱墅区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',6,3,3,'杭州','杭州拱墅区110号','近景区,24h前台,行李寄存',120.179255,30.266735,4.8,0,1,0,'含双早｜现做手冲咖啡',15100,125510,170283),
(14,NOW(),NOW(),NOW(),0,0,'小隐·设计师loft','杭州滨江区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs14-1/900/640,https://picsum.photos/seed/hs14-2/900/640,https://picsum.photos/seed/hs14-3/900/640','「小隐·设计师loft」位于杭州滨江区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',4,3,3,'杭州','杭州滨江区199号','海景房,无边泳池,免费接机',120.157421,30.229165,4.7,0,1,0,'不含餐｜周边餐饮丰富',0,108090,152452),
(15,NOW(),NOW(),NOW(),0,0,'砚山·湖景复式','杭州桐庐县 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs15-1/900/640,https://picsum.photos/seed/hs15-2/900/640,https://picsum.photos/seed/hs15-3/900/640','「砚山·湖景复式」位于杭州桐庐县，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,3,3,'杭州','杭州桐庐县64号','近景区,24h前台,行李寄存',120.123108,30.243987,4.8,0,1,0,'含三餐｜管家私房菜',8600,97880,129001),
(16,NOW(),NOW(),NOW(),0,0,'云栖·山景大床房','成都锦江区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs16-1/900/640,https://picsum.photos/seed/hs16-2/900/640,https://picsum.photos/seed/hs16-3/900/640','「云栖·山景大床房」位于成都锦江区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,4,4,'成都','成都锦江区68号','独栋小院,私汤温泉,观景露台',104.036322,30.538335,4.7,0,1,1,'含双早｜本地特色早餐',12600,56020,81811),
(17,NOW(),NOW(),NOW(),0,0,'屿北·海景套房','成都青羊区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs17-1/900/640,https://picsum.photos/seed/hs17-2/900/640,https://picsum.photos/seed/hs17-3/900/640','「屿北·海景套房」位于成都青羊区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',2,4,4,'成都','成都青羊区145号','独栋小院,私汤温泉,观景露台',104.055052,30.579634,4.7,0,1,0,'含三餐｜管家私房菜',16100,66400,93760),
(18,NOW(),NOW(),NOW(),0,0,'柒月·庭院小院','成都武侯区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs18-1/900/640,https://picsum.photos/seed/hs18-2/900/640,https://picsum.photos/seed/hs18-3/900/640','「柒月·庭院小院」位于成都武侯区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,4,4,'成都','成都武侯区125号','海景房,无边泳池,免费接机',104.027541,30.620802,4.6,0,1,0,'不含餐｜周边餐饮丰富',0,118060,153774),
(19,NOW(),NOW(),NOW(),0,0,'云庐·设计师loft','成都成华区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs19-1/900/640,https://picsum.photos/seed/hs19-2/900/640,https://picsum.photos/seed/hs19-3/900/640','「云庐·设计师loft」位于成都成华区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,4,4,'成都','成都成华区80号','独栋小院,私汤温泉,观景露台',104.024779,30.544464,4.4,0,1,0,'不含餐｜周边餐饮丰富',0,67720,93263),
(20,NOW(),NOW(),NOW(),0,0,'云庐·湖景复式','成都都江堰市 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs20-1/900/640,https://picsum.photos/seed/hs20-2/900/640,https://picsum.photos/seed/hs20-3/900/640','「云庐·湖景复式」位于成都都江堰市，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,4,4,'成都','成都都江堰市137号','海景房,无边泳池,免费接机',104.082204,30.539304,4.7,0,1,1,'含三餐｜管家私房菜',7000,83130,106103),
(21,NOW(),NOW(),NOW(),0,0,'寻野·山景大床房','厦门思明区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs21-1/900/640,https://picsum.photos/seed/hs21-2/900/640,https://picsum.photos/seed/hs21-3/900/640','「寻野·山景大床房」位于厦门思明区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',4,5,5,'厦门','厦门思明区72号','独栋小院,私汤温泉,观景露台',118.062086,24.433447,4.4,0,1,0,'不含餐｜周边餐饮丰富',0,154100,203336),
(22,NOW(),NOW(),NOW(),0,0,'拾贝·海景套房','厦门湖里区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs22-1/900/640,https://picsum.photos/seed/hs22-2/900/640,https://picsum.photos/seed/hs22-3/900/640','「拾贝·海景套房」位于厦门湖里区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,5,5,'厦门','厦门湖里区26号','免费停车,含早餐,宠物友好',118.079597,24.502740,4.6,0,1,0,'含三餐｜管家私房菜',7900,151820,209182),
(23,NOW(),NOW(),NOW(),0,0,'陌上·庭院小院','厦门集美区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs23-1/900/640,https://picsum.photos/seed/hs23-2/900/640,https://picsum.photos/seed/hs23-3/900/640','「陌上·庭院小院」位于厦门集美区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,5,5,'厦门','厦门集美区59号','设计感,大落地窗,投影仪',118.081433,24.456681,4.7,0,1,0,'含双早｜现做手冲咖啡',13200,30610,44556),
(24,NOW(),NOW(),NOW(),0,0,'木白·设计师loft','厦门同安区 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs24-1/900/640,https://picsum.photos/seed/hs24-2/900/640,https://picsum.photos/seed/hs24-3/900/640','「木白·设计师loft」位于厦门同安区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,5,5,'厦门','厦门同安区170号','独栋小院,私汤温泉,观景露台',118.080249,24.479574,4.8,0,1,1,'含双早｜本地特色早餐',7300,49830,66288),
(25,NOW(),NOW(),NOW(),0,0,'屿北·湖景复式','厦门环岛路 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs25-1/900/640,https://picsum.photos/seed/hs25-2/900/640,https://picsum.photos/seed/hs25-3/900/640','「屿北·湖景复式」位于厦门环岛路，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',6,5,5,'厦门','厦门环岛路63号','近景区,24h前台,行李寄存',118.090394,24.521574,4.5,0,1,0,'不含餐｜周边餐饮丰富',0,136800,168609),
(26,NOW(),NOW(),NOW(),0,0,'云庐·山景大床房','丽江古城区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs26-1/900/640,https://picsum.photos/seed/hs26-2/900/640,https://picsum.photos/seed/hs26-3/900/640','「云庐·山景大床房」位于丽江古城区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,6,6,'丽江','丽江古城区64号','设计感,大落地窗,投影仪',100.191704,26.855935,4.7,0,1,0,'不含餐｜周边餐饮丰富',0,27210,40204),
(27,NOW(),NOW(),NOW(),0,0,'岑谷·海景套房','丽江束河古镇 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs27-1/900/640,https://picsum.photos/seed/hs27-2/900/640,https://picsum.photos/seed/hs27-3/900/640','「岑谷·海景套房」位于丽江束河古镇，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,6,6,'丽江','丽江束河古镇169号','免费停车,含早餐,宠物友好',100.265957,26.830824,4.8,0,1,0,'不含餐｜周边餐饮丰富',0,78580,107194),
(28,NOW(),NOW(),NOW(),0,0,'柒月·庭院小院','丽江白沙古镇 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs28-1/900/640,https://picsum.photos/seed/hs28-2/900/640,https://picsum.photos/seed/hs28-3/900/640','「柒月·庭院小院」位于丽江白沙古镇，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,6,6,'丽江','丽江白沙古镇192号','设计感,大落地窗,投影仪',100.219980,26.874091,4.7,0,1,1,'不含餐｜周边餐饮丰富',0,95050,138466),
(29,NOW(),NOW(),NOW(),0,0,'归屿·设计师loft','丽江玉龙县 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs29-1/900/640,https://picsum.photos/seed/hs29-2/900/640,https://picsum.photos/seed/hs29-3/900/640','「归屿·设计师loft」位于丽江玉龙县，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,6,6,'丽江','丽江玉龙县197号','设计感,大落地窗,投影仪',100.229801,26.901992,4.6,0,1,0,'含双早｜本地特色早餐',8900,34790,51324),
(30,NOW(),NOW(),NOW(),0,0,'长亭·湖景复式','丽江泸沽湖 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs30-1/900/640,https://picsum.photos/seed/hs30-2/900/640,https://picsum.photos/seed/hs30-3/900/640','「长亭·湖景复式」位于丽江泸沽湖，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',2,6,6,'丽江','丽江泸沽湖49号','免费停车,含早餐,宠物友好',100.240418,26.867741,4.7,0,1,0,'含三餐｜管家私房菜',11500,119890,168122),
(31,NOW(),NOW(),NOW(),0,0,'白屿·山景大床房','三亚海棠湾 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs31-1/900/640,https://picsum.photos/seed/hs31-2/900/640,https://picsum.photos/seed/hs31-3/900/640','「白屿·山景大床房」位于三亚海棠湾，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,7,7,'三亚','三亚海棠湾22号','独栋小院,私汤温泉,观景露台',109.522149,18.241254,5.0,0,1,0,'不含餐｜周边餐饮丰富',0,121850,159694),
(32,NOW(),NOW(),NOW(),0,0,'停云·海景套房','三亚亚龙湾 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs32-1/900/640,https://picsum.photos/seed/hs32-2/900/640,https://picsum.photos/seed/hs32-3/900/640','「停云·海景套房」位于三亚亚龙湾，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,7,7,'三亚','三亚亚龙湾74号','独栋小院,私汤温泉,观景露台',109.506748,18.249972,4.6,0,1,1,'不含餐｜周边餐饮丰富',0,87060,118836),
(33,NOW(),NOW(),NOW(),0,0,'木白·庭院小院','三亚大东海 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs33-1/900/640,https://picsum.photos/seed/hs33-2/900/640,https://picsum.photos/seed/hs33-3/900/640','「木白·庭院小院」位于三亚大东海，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',3,7,7,'三亚','三亚大东海146号','免费停车,含早餐,宠物友好',109.523202,18.255648,4.5,0,1,0,'含双早｜现做手冲咖啡',14300,50600,71007),
(34,NOW(),NOW(),NOW(),0,0,'屿北·设计师loft','三亚三亚湾 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs34-1/900/640,https://picsum.photos/seed/hs34-2/900/640,https://picsum.photos/seed/hs34-3/900/640','「屿北·设计师loft」位于三亚三亚湾，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',6,7,7,'三亚','三亚三亚湾102号','免费停车,含早餐,宠物友好',109.547430,18.217107,4.7,0,1,0,'不含餐｜周边餐饮丰富',0,42180,55586),
(35,NOW(),NOW(),NOW(),0,0,'云栖·湖景复式','三亚崖州区 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs35-1/900/640,https://picsum.photos/seed/hs35-2/900/640,https://picsum.photos/seed/hs35-3/900/640','「云栖·湖景复式」位于三亚崖州区，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',6,7,7,'三亚','三亚崖州区22号','免费停车,含早餐,宠物友好',109.555356,18.263714,4.9,0,1,0,'不含餐｜周边餐饮丰富',0,104490,140695),
(36,NOW(),NOW(),NOW(),0,0,'白屿·山景大床房','大理大理古城 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs36-1/900/640,https://picsum.photos/seed/hs36-2/900/640,https://picsum.photos/seed/hs36-3/900/640','「白屿·山景大床房」位于大理大理古城，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,8,8,'大理','大理大理古城113号','近地铁,可做饭,洗衣机',100.298705,25.582791,4.7,0,1,1,'不含餐｜周边餐饮丰富',0,84210,122927),
(37,NOW(),NOW(),NOW(),0,0,'初见·海景套房','大理洱海东岸 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs37-1/900/640,https://picsum.photos/seed/hs37-2/900/640,https://picsum.photos/seed/hs37-3/900/640','「初见·海景套房」位于大理洱海东岸，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',5,8,8,'大理','大理洱海东岸178号','海景房,无边泳池,免费接机',100.252282,25.603171,4.5,0,1,0,'含三餐｜管家私房菜',7100,94400,140748),
(38,NOW(),NOW(),NOW(),0,0,'云庐·庭院小院','大理双廊镇 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs38-1/900/640,https://picsum.photos/seed/hs38-2/900/640,https://picsum.photos/seed/hs38-3/900/640','「云庐·庭院小院」位于大理双廊镇，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',6,8,8,'大理','大理双廊镇9号','近景区,24h前台,行李寄存',100.312197,25.581910,4.7,0,1,0,'含双早｜现做手冲咖啡',16300,132350,177254),
(39,NOW(),NOW(),NOW(),0,0,'空谷·设计师loft','大理喜洲镇 · 步行可达景区 · 大落地窗','https://picsum.photos/seed/hs39-1/900/640,https://picsum.photos/seed/hs39-2/900/640,https://picsum.photos/seed/hs39-3/900/640','「空谷·设计师loft」位于大理喜洲镇，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',6,8,8,'大理','大理喜洲镇56号','设计感,大落地窗,投影仪',100.251252,25.648596,5.0,0,1,0,'含双早｜本地特色早餐',11400,59770,75347),
(40,NOW(),NOW(),NOW(),0,0,'青庐·湖景复式','大理苍山脚下 · 步行可达景区 · 独立露台','https://picsum.photos/seed/hs40-1/900/640,https://picsum.photos/seed/hs40-2/900/640,https://picsum.photos/seed/hs40-3/900/640','「青庐·湖景复式」位于大理苍山脚下，步行可达核心商圈与景点。房间采用原木与微水泥的简约设计，配有大面积落地窗与独立露台，全天候自然采光。房间配备高品质床品、全套洗护、免费WiFi与投影设备；房东常驻本地，可提供行程规划与在地美食推荐。适合情侣出游、家庭亲子与朋友小聚。',2,8,8,'大理','大理苍山脚下77号','设计感,大落地窗,投影仪',100.225853,25.618263,4.5,0,1,1,'含三餐｜管家私房菜',11900,73370,104582);

-- ---------- 评价（307 条） ----------
INSERT INTO `homestay_comment` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`homestay_id`,`user_id`,`nickname`,`avatar`,`content`,`star`,`version`) VALUES
(1,DATE_SUB(NOW(), INTERVAL 1 DAY),NOW(),NOW(),0,1,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。','[5,5,4.5]',0),
(2,DATE_SUB(NOW(), INTERVAL 143 DAY),NOW(),NOW(),0,1,2,'周予安','https://picsum.photos/seed/avatar2/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,4.5,5]',0),
(3,DATE_SUB(NOW(), INTERVAL 10 DAY),NOW(),NOW(),0,1,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[4.5,5,4.5]',0),
(4,DATE_SUB(NOW(), INTERVAL 64 DAY),NOW(),NOW(),0,1,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[5,4.5,5]',0),
(5,DATE_SUB(NOW(), INTERVAL 30 DAY),NOW(),NOW(),0,2,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[5,4.5,4.5]',0),
(6,DATE_SUB(NOW(), INTERVAL 46 DAY),NOW(),NOW(),0,2,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[5,4.5,4.5]',0),
(7,DATE_SUB(NOW(), INTERVAL 70 DAY),NOW(),NOW(),0,2,8,'江眠','https://picsum.photos/seed/avatar8/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,4.5,5]',0),
(8,DATE_SUB(NOW(), INTERVAL 179 DAY),NOW(),NOW(),0,3,2,'周予安','https://picsum.photos/seed/avatar2/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[4.5,5,4.5]',0),
(9,DATE_SUB(NOW(), INTERVAL 54 DAY),NOW(),NOW(),0,3,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[5,4.5,4.5]',0),
(10,DATE_SUB(NOW(), INTERVAL 132 DAY),NOW(),NOW(),0,3,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[4.5,4.5,4.5]',0),
(11,DATE_SUB(NOW(), INTERVAL 32 DAY),NOW(),NOW(),0,3,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,5,5]',0),
(12,DATE_SUB(NOW(), INTERVAL 79 DAY),NOW(),NOW(),0,3,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[5,4.5,5]',0),
(13,DATE_SUB(NOW(), INTERVAL 63 DAY),NOW(),NOW(),0,3,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[5,5,4.5]',0),
(14,DATE_SUB(NOW(), INTERVAL 104 DAY),NOW(),NOW(),0,3,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','唯一小建议是希望能加个全身镜，其他都很满意。','[5,5,4.5]',0),
(15,DATE_SUB(NOW(), INTERVAL 135 DAY),NOW(),NOW(),0,3,8,'江眠','https://picsum.photos/seed/avatar8/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[4.5,4.5,4.5]',0),
(16,DATE_SUB(NOW(), INTERVAL 41 DAY),NOW(),NOW(),0,3,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[4.5,5,5]',0),
(17,DATE_SUB(NOW(), INTERVAL 13 DAY),NOW(),NOW(),0,4,2,'周予安','https://picsum.photos/seed/avatar2/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[5,5,5]',0),
(18,DATE_SUB(NOW(), INTERVAL 126 DAY),NOW(),NOW(),0,4,8,'江眠','https://picsum.photos/seed/avatar8/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[4.5,5,4.5]',0),
(19,DATE_SUB(NOW(), INTERVAL 28 DAY),NOW(),NOW(),0,4,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','空调给力，热水很快，洗澡体验很好。','[4.5,4.5,4.5]',0),
(20,DATE_SUB(NOW(), INTERVAL 169 DAY),NOW(),NOW(),0,4,8,'江眠','https://picsum.photos/seed/avatar8/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[4.5,5,5]',0),
(21,DATE_SUB(NOW(), INTERVAL 132 DAY),NOW(),NOW(),0,4,2,'周予安','https://picsum.photos/seed/avatar2/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,4.5,5]',0),
(22,DATE_SUB(NOW(), INTERVAL 53 DAY),NOW(),NOW(),0,4,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,5,4.5]',0),
(23,DATE_SUB(NOW(), INTERVAL 21 DAY),NOW(),NOW(),0,4,1,'林小满','https://picsum.photos/seed/avatar1/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[4.5,5,5]',0),
(24,DATE_SUB(NOW(), INTERVAL 122 DAY),NOW(),NOW(),0,4,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[5,5,4.5]',0),
(25,DATE_SUB(NOW(), INTERVAL 128 DAY),NOW(),NOW(),0,4,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[5,5,5]',0),
(26,DATE_SUB(NOW(), INTERVAL 23 DAY),NOW(),NOW(),0,5,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,5,4.5]',0),
(27,DATE_SUB(NOW(), INTERVAL 25 DAY),NOW(),NOW(),0,5,2,'周予安','https://picsum.photos/seed/avatar2/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[4.5,5,5]',0),
(28,DATE_SUB(NOW(), INTERVAL 153 DAY),NOW(),NOW(),0,5,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[5,5,4.5]',0),
(29,DATE_SUB(NOW(), INTERVAL 64 DAY),NOW(),NOW(),0,5,1,'林小满','https://picsum.photos/seed/avatar1/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,5,4.5]',0),
(30,DATE_SUB(NOW(), INTERVAL 166 DAY),NOW(),NOW(),0,5,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[5,5,5]',0),
(31,DATE_SUB(NOW(), INTERVAL 32 DAY),NOW(),NOW(),0,6,8,'江眠','https://picsum.photos/seed/avatar8/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[5,5,5]',0),
(32,DATE_SUB(NOW(), INTERVAL 127 DAY),NOW(),NOW(),0,6,8,'江眠','https://picsum.photos/seed/avatar8/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[4.5,4.5,4.5]',0),
(33,DATE_SUB(NOW(), INTERVAL 110 DAY),NOW(),NOW(),0,6,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','空调给力，热水很快，洗澡体验很好。','[5,5,4.5]',0),
(34,DATE_SUB(NOW(), INTERVAL 62 DAY),NOW(),NOW(),0,6,2,'周予安','https://picsum.photos/seed/avatar2/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[5,5,5]',0),
(35,DATE_SUB(NOW(), INTERVAL 162 DAY),NOW(),NOW(),0,6,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[4.5,5,5]',0),
(36,DATE_SUB(NOW(), INTERVAL 31 DAY),NOW(),NOW(),0,6,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','空调给力，热水很快，洗澡体验很好。','[5,5,5]',0),
(37,DATE_SUB(NOW(), INTERVAL 86 DAY),NOW(),NOW(),0,6,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,5,4.5]',0),
(38,DATE_SUB(NOW(), INTERVAL 67 DAY),NOW(),NOW(),0,6,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[5,5,5]',0),
(39,DATE_SUB(NOW(), INTERVAL 48 DAY),NOW(),NOW(),0,6,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','空调给力，热水很快，洗澡体验很好。','[5,4.5,5]',0),
(40,DATE_SUB(NOW(), INTERVAL 13 DAY),NOW(),NOW(),0,6,8,'江眠','https://picsum.photos/seed/avatar8/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[4.5,4.5,5]',0),
(41,DATE_SUB(NOW(), INTERVAL 71 DAY),NOW(),NOW(),0,6,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[5,4.5,5]',0),
(42,DATE_SUB(NOW(), INTERVAL 123 DAY),NOW(),NOW(),0,6,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','唯一小建议是希望能加个全身镜，其他都很满意。','[4.5,4.5,5]',0),
(43,DATE_SUB(NOW(), INTERVAL 10 DAY),NOW(),NOW(),0,7,1,'林小满','https://picsum.photos/seed/avatar1/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[5,5,5]',0),
(44,DATE_SUB(NOW(), INTERVAL 96 DAY),NOW(),NOW(),0,7,8,'江眠','https://picsum.photos/seed/avatar8/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[5,4.5,5]',0),
(45,DATE_SUB(NOW(), INTERVAL 154 DAY),NOW(),NOW(),0,7,2,'周予安','https://picsum.photos/seed/avatar2/200/200','空调给力，热水很快，洗澡体验很好。','[4.5,5,4.5]',0),
(46,DATE_SUB(NOW(), INTERVAL 127 DAY),NOW(),NOW(),0,7,1,'林小满','https://picsum.photos/seed/avatar1/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[5,4.5,5]',0),
(47,DATE_SUB(NOW(), INTERVAL 101 DAY),NOW(),NOW(),0,7,2,'周予安','https://picsum.photos/seed/avatar2/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[4.5,5,4.5]',0),
(48,DATE_SUB(NOW(), INTERVAL 63 DAY),NOW(),NOW(),0,7,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,4.5,4.5]',0),
(49,DATE_SUB(NOW(), INTERVAL 20 DAY),NOW(),NOW(),0,7,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[4.5,5,4.5]',0),
(50,DATE_SUB(NOW(), INTERVAL 14 DAY),NOW(),NOW(),0,8,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[4.5,4.5,5]',0),
(51,DATE_SUB(NOW(), INTERVAL 105 DAY),NOW(),NOW(),0,8,1,'林小满','https://picsum.photos/seed/avatar1/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[4.5,4.5,5]',0),
(52,DATE_SUB(NOW(), INTERVAL 96 DAY),NOW(),NOW(),0,8,2,'周予安','https://picsum.photos/seed/avatar2/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[4.5,5,4.5]',0),
(53,DATE_SUB(NOW(), INTERVAL 133 DAY),NOW(),NOW(),0,9,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[4.5,4.5,4.5]',0),
(54,DATE_SUB(NOW(), INTERVAL 169 DAY),NOW(),NOW(),0,9,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[5,5,4.5]',0),
(55,DATE_SUB(NOW(), INTERVAL 115 DAY),NOW(),NOW(),0,9,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[5,5,4.5]',0),
(56,DATE_SUB(NOW(), INTERVAL 157 DAY),NOW(),NOW(),0,9,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,5,4.5]',0),
(57,DATE_SUB(NOW(), INTERVAL 57 DAY),NOW(),NOW(),0,9,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[4.5,4.5,4.5]',0),
(58,DATE_SUB(NOW(), INTERVAL 168 DAY),NOW(),NOW(),0,9,2,'周予安','https://picsum.photos/seed/avatar2/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[5,5,4.5]',0),
(59,DATE_SUB(NOW(), INTERVAL 2 DAY),NOW(),NOW(),0,9,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[4.5,4.5,4.5]',0),
(60,DATE_SUB(NOW(), INTERVAL 70 DAY),NOW(),NOW(),0,9,8,'江眠','https://picsum.photos/seed/avatar8/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,4.5,4.5]',0),
(61,DATE_SUB(NOW(), INTERVAL 57 DAY),NOW(),NOW(),0,9,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[5,4.5,5]',0),
(62,DATE_SUB(NOW(), INTERVAL 93 DAY),NOW(),NOW(),0,9,1,'林小满','https://picsum.photos/seed/avatar1/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[4.5,4.5,4.5]',0),
(63,DATE_SUB(NOW(), INTERVAL 108 DAY),NOW(),NOW(),0,9,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,4.5,4.5]',0),
(64,DATE_SUB(NOW(), INTERVAL 15 DAY),NOW(),NOW(),0,10,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,4.5,5]',0),
(65,DATE_SUB(NOW(), INTERVAL 3 DAY),NOW(),NOW(),0,10,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,4.5,4.5]',0),
(66,DATE_SUB(NOW(), INTERVAL 113 DAY),NOW(),NOW(),0,10,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,5,5]',0),
(67,DATE_SUB(NOW(), INTERVAL 8 DAY),NOW(),NOW(),0,10,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。','[4.5,4.5,4.5]',0),
(68,DATE_SUB(NOW(), INTERVAL 133 DAY),NOW(),NOW(),0,10,2,'周予安','https://picsum.photos/seed/avatar2/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[4.5,4.5,4.5]',0),
(69,DATE_SUB(NOW(), INTERVAL 38 DAY),NOW(),NOW(),0,10,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,4.5,5]',0),
(70,DATE_SUB(NOW(), INTERVAL 96 DAY),NOW(),NOW(),0,10,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[4.5,5,5]',0),
(71,DATE_SUB(NOW(), INTERVAL 154 DAY),NOW(),NOW(),0,10,8,'江眠','https://picsum.photos/seed/avatar8/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,5,4.5]',0),
(72,DATE_SUB(NOW(), INTERVAL 139 DAY),NOW(),NOW(),0,10,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,5,5]',0),
(73,DATE_SUB(NOW(), INTERVAL 180 DAY),NOW(),NOW(),0,11,2,'周予安','https://picsum.photos/seed/avatar2/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[5,5,5]',0),
(74,DATE_SUB(NOW(), INTERVAL 35 DAY),NOW(),NOW(),0,11,8,'江眠','https://picsum.photos/seed/avatar8/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[4.5,5,4.5]',0),
(75,DATE_SUB(NOW(), INTERVAL 44 DAY),NOW(),NOW(),0,11,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[4.5,5,5]',0),
(76,DATE_SUB(NOW(), INTERVAL 107 DAY),NOW(),NOW(),0,11,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[4.5,5,5]',0),
(77,DATE_SUB(NOW(), INTERVAL 89 DAY),NOW(),NOW(),0,11,1,'林小满','https://picsum.photos/seed/avatar1/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,4.5,4.5]',0),
(78,DATE_SUB(NOW(), INTERVAL 98 DAY),NOW(),NOW(),0,11,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,5,4.5]',0),
(79,DATE_SUB(NOW(), INTERVAL 108 DAY),NOW(),NOW(),0,11,1,'林小满','https://picsum.photos/seed/avatar1/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,4.5,4.5]',0),
(80,DATE_SUB(NOW(), INTERVAL 36 DAY),NOW(),NOW(),0,12,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,4.5,5]',0),
(81,DATE_SUB(NOW(), INTERVAL 24 DAY),NOW(),NOW(),0,12,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[4.5,5,4.5]',0),
(82,DATE_SUB(NOW(), INTERVAL 180 DAY),NOW(),NOW(),0,12,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[4.5,4.5,5]',0),
(83,DATE_SUB(NOW(), INTERVAL 77 DAY),NOW(),NOW(),0,12,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[4.5,5,4.5]',0),
(84,DATE_SUB(NOW(), INTERVAL 99 DAY),NOW(),NOW(),0,12,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,4.5,5]',0),
(85,DATE_SUB(NOW(), INTERVAL 76 DAY),NOW(),NOW(),0,12,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','空调给力，热水很快，洗澡体验很好。','[5,4.5,5]',0),
(86,DATE_SUB(NOW(), INTERVAL 35 DAY),NOW(),NOW(),0,12,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[4.5,4.5,4.5]',0),
(87,DATE_SUB(NOW(), INTERVAL 47 DAY),NOW(),NOW(),0,12,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[4.5,4.5,5]',0),
(88,DATE_SUB(NOW(), INTERVAL 2 DAY),NOW(),NOW(),0,12,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[4.5,4.5,5]',0),
(89,DATE_SUB(NOW(), INTERVAL 114 DAY),NOW(),NOW(),0,12,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,4.5,5]',0),
(90,DATE_SUB(NOW(), INTERVAL 22 DAY),NOW(),NOW(),0,12,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,4.5,5]',0),
(91,DATE_SUB(NOW(), INTERVAL 52 DAY),NOW(),NOW(),0,13,8,'江眠','https://picsum.photos/seed/avatar8/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[5,5,4.5]',0),
(92,DATE_SUB(NOW(), INTERVAL 113 DAY),NOW(),NOW(),0,13,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[5,4.5,5]',0),
(93,DATE_SUB(NOW(), INTERVAL 127 DAY),NOW(),NOW(),0,13,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,5,4.5]',0),
(94,DATE_SUB(NOW(), INTERVAL 46 DAY),NOW(),NOW(),0,13,1,'林小满','https://picsum.photos/seed/avatar1/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,5,5]',0),
(95,DATE_SUB(NOW(), INTERVAL 165 DAY),NOW(),NOW(),0,14,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[5,5,4.5]',0),
(96,DATE_SUB(NOW(), INTERVAL 165 DAY),NOW(),NOW(),0,14,2,'周予安','https://picsum.photos/seed/avatar2/200/200','窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。','[4.5,5,4.5]',0),
(97,DATE_SUB(NOW(), INTERVAL 130 DAY),NOW(),NOW(),0,14,1,'林小满','https://picsum.photos/seed/avatar1/200/200','窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。','[5,4.5,4.5]',0),
(98,DATE_SUB(NOW(), INTERVAL 82 DAY),NOW(),NOW(),0,14,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,4.5,4.5]',0),
(99,DATE_SUB(NOW(), INTERVAL 61 DAY),NOW(),NOW(),0,14,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[4.5,5,5]',0),
(100,DATE_SUB(NOW(), INTERVAL 9 DAY),NOW(),NOW(),0,15,8,'江眠','https://picsum.photos/seed/avatar8/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[4.5,5,4.5]',0),
(101,DATE_SUB(NOW(), INTERVAL 178 DAY),NOW(),NOW(),0,15,2,'周予安','https://picsum.photos/seed/avatar2/200/200','空调给力，热水很快，洗澡体验很好。','[4.5,5,4.5]',0),
(102,DATE_SUB(NOW(), INTERVAL 30 DAY),NOW(),NOW(),0,15,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[5,4.5,5]',0),
(103,DATE_SUB(NOW(), INTERVAL 126 DAY),NOW(),NOW(),0,15,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[4.5,5,5]',0),
(104,DATE_SUB(NOW(), INTERVAL 7 DAY),NOW(),NOW(),0,15,2,'周予安','https://picsum.photos/seed/avatar2/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[4.5,5,4.5]',0),
(105,DATE_SUB(NOW(), INTERVAL 8 DAY),NOW(),NOW(),0,15,2,'周予安','https://picsum.photos/seed/avatar2/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[4.5,5,4.5]',0),
(106,DATE_SUB(NOW(), INTERVAL 125 DAY),NOW(),NOW(),0,15,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','唯一小建议是希望能加个全身镜，其他都很满意。','[5,4.5,5]',0),
(107,DATE_SUB(NOW(), INTERVAL 80 DAY),NOW(),NOW(),0,15,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','空调给力，热水很快，洗澡体验很好。','[4.5,4.5,5]',0),
(108,DATE_SUB(NOW(), INTERVAL 30 DAY),NOW(),NOW(),0,16,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,5,5]',0),
(109,DATE_SUB(NOW(), INTERVAL 79 DAY),NOW(),NOW(),0,16,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,4.5,5]',0),
(110,DATE_SUB(NOW(), INTERVAL 45 DAY),NOW(),NOW(),0,16,8,'江眠','https://picsum.photos/seed/avatar8/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,5,5]',0),
(111,DATE_SUB(NOW(), INTERVAL 120 DAY),NOW(),NOW(),0,16,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,4.5,5]',0),
(112,DATE_SUB(NOW(), INTERVAL 82 DAY),NOW(),NOW(),0,17,1,'林小满','https://picsum.photos/seed/avatar1/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,5,5]',0),
(113,DATE_SUB(NOW(), INTERVAL 48 DAY),NOW(),NOW(),0,17,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,5,5]',0),
(114,DATE_SUB(NOW(), INTERVAL 14 DAY),NOW(),NOW(),0,17,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[4.5,5,5]',0),
(115,DATE_SUB(NOW(), INTERVAL 120 DAY),NOW(),NOW(),0,17,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[4.5,4.5,5]',0),
(116,DATE_SUB(NOW(), INTERVAL 138 DAY),NOW(),NOW(),0,17,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,4.5,4.5]',0),
(117,DATE_SUB(NOW(), INTERVAL 47 DAY),NOW(),NOW(),0,17,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[4.5,4.5,4.5]',0),
(118,DATE_SUB(NOW(), INTERVAL 147 DAY),NOW(),NOW(),0,17,8,'江眠','https://picsum.photos/seed/avatar8/200/200','窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。','[4.5,5,5]',0),
(119,DATE_SUB(NOW(), INTERVAL 168 DAY),NOW(),NOW(),0,17,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[4.5,4.5,5]',0),
(120,DATE_SUB(NOW(), INTERVAL 86 DAY),NOW(),NOW(),0,17,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,4.5,4.5]',0),
(121,DATE_SUB(NOW(), INTERVAL 176 DAY),NOW(),NOW(),0,17,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[4.5,5,4.5]',0),
(122,DATE_SUB(NOW(), INTERVAL 1 DAY),NOW(),NOW(),0,17,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。','[4.5,5,4.5]',0),
(123,DATE_SUB(NOW(), INTERVAL 173 DAY),NOW(),NOW(),0,17,8,'江眠','https://picsum.photos/seed/avatar8/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[4.5,5,5]',0),
(124,DATE_SUB(NOW(), INTERVAL 138 DAY),NOW(),NOW(),0,18,1,'林小满','https://picsum.photos/seed/avatar1/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[4.5,5,5]',0),
(125,DATE_SUB(NOW(), INTERVAL 108 DAY),NOW(),NOW(),0,18,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,5,4.5]',0),
(126,DATE_SUB(NOW(), INTERVAL 172 DAY),NOW(),NOW(),0,18,2,'周予安','https://picsum.photos/seed/avatar2/200/200','窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。','[4.5,5,5]',0),
(127,DATE_SUB(NOW(), INTERVAL 61 DAY),NOW(),NOW(),0,18,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[5,5,4.5]',0),
(128,DATE_SUB(NOW(), INTERVAL 82 DAY),NOW(),NOW(),0,19,1,'林小满','https://picsum.photos/seed/avatar1/200/200','空调给力，热水很快，洗澡体验很好。','[4.5,4.5,5]',0),
(129,DATE_SUB(NOW(), INTERVAL 2 DAY),NOW(),NOW(),0,19,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,4.5,5]',0),
(130,DATE_SUB(NOW(), INTERVAL 9 DAY),NOW(),NOW(),0,19,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,5,4.5]',0),
(131,DATE_SUB(NOW(), INTERVAL 48 DAY),NOW(),NOW(),0,19,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,5,4.5]',0),
(132,DATE_SUB(NOW(), INTERVAL 125 DAY),NOW(),NOW(),0,19,2,'周予安','https://picsum.photos/seed/avatar2/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,4.5,5]',0),
(133,DATE_SUB(NOW(), INTERVAL 99 DAY),NOW(),NOW(),0,19,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,4.5,5]',0),
(134,DATE_SUB(NOW(), INTERVAL 155 DAY),NOW(),NOW(),0,20,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[4.5,4.5,5]',0),
(135,DATE_SUB(NOW(), INTERVAL 44 DAY),NOW(),NOW(),0,20,2,'周予安','https://picsum.photos/seed/avatar2/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,5,4.5]',0),
(136,DATE_SUB(NOW(), INTERVAL 174 DAY),NOW(),NOW(),0,20,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[5,5,4.5]',0),
(137,DATE_SUB(NOW(), INTERVAL 126 DAY),NOW(),NOW(),0,20,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[4.5,5,5]',0),
(138,DATE_SUB(NOW(), INTERVAL 9 DAY),NOW(),NOW(),0,20,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,4.5,4.5]',0),
(139,DATE_SUB(NOW(), INTERVAL 26 DAY),NOW(),NOW(),0,20,8,'江眠','https://picsum.photos/seed/avatar8/200/200','空调给力，热水很快，洗澡体验很好。','[5,5,5]',0),
(140,DATE_SUB(NOW(), INTERVAL 33 DAY),NOW(),NOW(),0,21,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,4.5,5]',0),
(141,DATE_SUB(NOW(), INTERVAL 9 DAY),NOW(),NOW(),0,21,2,'周予安','https://picsum.photos/seed/avatar2/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[5,4.5,4.5]',0),
(142,DATE_SUB(NOW(), INTERVAL 62 DAY),NOW(),NOW(),0,21,1,'林小满','https://picsum.photos/seed/avatar1/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,4.5,4.5]',0),
(143,DATE_SUB(NOW(), INTERVAL 100 DAY),NOW(),NOW(),0,22,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[4.5,4.5,4.5]',0),
(144,DATE_SUB(NOW(), INTERVAL 172 DAY),NOW(),NOW(),0,22,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[4.5,5,4.5]',0),
(145,DATE_SUB(NOW(), INTERVAL 132 DAY),NOW(),NOW(),0,22,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[4.5,5,4.5]',0),
(146,DATE_SUB(NOW(), INTERVAL 165 DAY),NOW(),NOW(),0,22,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[5,4.5,4.5]',0),
(147,DATE_SUB(NOW(), INTERVAL 21 DAY),NOW(),NOW(),0,22,1,'林小满','https://picsum.photos/seed/avatar1/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[5,5,5]',0),
(148,DATE_SUB(NOW(), INTERVAL 44 DAY),NOW(),NOW(),0,22,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,5,5]',0),
(149,DATE_SUB(NOW(), INTERVAL 58 DAY),NOW(),NOW(),0,22,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,4.5,4.5]',0),
(150,DATE_SUB(NOW(), INTERVAL 87 DAY),NOW(),NOW(),0,23,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,5,4.5]',0),
(151,DATE_SUB(NOW(), INTERVAL 56 DAY),NOW(),NOW(),0,23,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[5,5,4.5]',0),
(152,DATE_SUB(NOW(), INTERVAL 8 DAY),NOW(),NOW(),0,23,2,'周予安','https://picsum.photos/seed/avatar2/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[5,5,5]',0),
(153,DATE_SUB(NOW(), INTERVAL 2 DAY),NOW(),NOW(),0,23,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,4.5,5]',0),
(154,DATE_SUB(NOW(), INTERVAL 103 DAY),NOW(),NOW(),0,23,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,4.5,4.5]',0),
(155,DATE_SUB(NOW(), INTERVAL 157 DAY),NOW(),NOW(),0,23,1,'林小满','https://picsum.photos/seed/avatar1/200/200','空调给力，热水很快，洗澡体验很好。','[5,4.5,4.5]',0),
(156,DATE_SUB(NOW(), INTERVAL 138 DAY),NOW(),NOW(),0,23,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[4.5,5,5]',0),
(157,DATE_SUB(NOW(), INTERVAL 2 DAY),NOW(),NOW(),0,23,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[4.5,4.5,5]',0),
(158,DATE_SUB(NOW(), INTERVAL 53 DAY),NOW(),NOW(),0,23,1,'林小满','https://picsum.photos/seed/avatar1/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,5,5]',0),
(159,DATE_SUB(NOW(), INTERVAL 3 DAY),NOW(),NOW(),0,23,2,'周予安','https://picsum.photos/seed/avatar2/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,4.5,4.5]',0),
(160,DATE_SUB(NOW(), INTERVAL 62 DAY),NOW(),NOW(),0,23,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,4.5,5]',0),
(161,DATE_SUB(NOW(), INTERVAL 147 DAY),NOW(),NOW(),0,23,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[4.5,5,4.5]',0),
(162,DATE_SUB(NOW(), INTERVAL 103 DAY),NOW(),NOW(),0,24,2,'周予安','https://picsum.photos/seed/avatar2/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[4.5,5,5]',0),
(163,DATE_SUB(NOW(), INTERVAL 55 DAY),NOW(),NOW(),0,24,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[4.5,5,4.5]',0),
(164,DATE_SUB(NOW(), INTERVAL 168 DAY),NOW(),NOW(),0,24,2,'周予安','https://picsum.photos/seed/avatar2/200/200','空调给力，热水很快，洗澡体验很好。','[5,5,5]',0),
(165,DATE_SUB(NOW(), INTERVAL 62 DAY),NOW(),NOW(),0,25,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[4.5,4.5,5]',0),
(166,DATE_SUB(NOW(), INTERVAL 72 DAY),NOW(),NOW(),0,25,1,'林小满','https://picsum.photos/seed/avatar1/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[4.5,5,5]',0),
(167,DATE_SUB(NOW(), INTERVAL 69 DAY),NOW(),NOW(),0,25,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,4.5,4.5]',0),
(168,DATE_SUB(NOW(), INTERVAL 36 DAY),NOW(),NOW(),0,25,8,'江眠','https://picsum.photos/seed/avatar8/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[4.5,4.5,5]',0),
(169,DATE_SUB(NOW(), INTERVAL 127 DAY),NOW(),NOW(),0,25,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[5,4.5,5]',0),
(170,DATE_SUB(NOW(), INTERVAL 168 DAY),NOW(),NOW(),0,25,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[5,4.5,4.5]',0),
(171,DATE_SUB(NOW(), INTERVAL 78 DAY),NOW(),NOW(),0,25,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[5,4.5,5]',0),
(172,DATE_SUB(NOW(), INTERVAL 167 DAY),NOW(),NOW(),0,25,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。','[5,5,4.5]',0),
(173,DATE_SUB(NOW(), INTERVAL 35 DAY),NOW(),NOW(),0,25,1,'林小满','https://picsum.photos/seed/avatar1/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[4.5,5,4.5]',0),
(174,DATE_SUB(NOW(), INTERVAL 138 DAY),NOW(),NOW(),0,25,8,'江眠','https://picsum.photos/seed/avatar8/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[5,5,5]',0),
(175,DATE_SUB(NOW(), INTERVAL 48 DAY),NOW(),NOW(),0,25,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[4.5,5,4.5]',0),
(176,DATE_SUB(NOW(), INTERVAL 102 DAY),NOW(),NOW(),0,26,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','空调给力，热水很快，洗澡体验很好。','[4.5,5,4.5]',0),
(177,DATE_SUB(NOW(), INTERVAL 138 DAY),NOW(),NOW(),0,26,8,'江眠','https://picsum.photos/seed/avatar8/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,4.5,5]',0),
(178,DATE_SUB(NOW(), INTERVAL 1 DAY),NOW(),NOW(),0,26,1,'林小满','https://picsum.photos/seed/avatar1/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,5,4.5]',0),
(179,DATE_SUB(NOW(), INTERVAL 144 DAY),NOW(),NOW(),0,26,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[4.5,5,5]',0),
(180,DATE_SUB(NOW(), INTERVAL 143 DAY),NOW(),NOW(),0,26,2,'周予安','https://picsum.photos/seed/avatar2/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[5,5,4.5]',0),
(181,DATE_SUB(NOW(), INTERVAL 39 DAY),NOW(),NOW(),0,26,8,'江眠','https://picsum.photos/seed/avatar8/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[4.5,4.5,5]',0),
(182,DATE_SUB(NOW(), INTERVAL 72 DAY),NOW(),NOW(),0,26,8,'江眠','https://picsum.photos/seed/avatar8/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,5,5]',0),
(183,DATE_SUB(NOW(), INTERVAL 137 DAY),NOW(),NOW(),0,26,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[5,4.5,4.5]',0),
(184,DATE_SUB(NOW(), INTERVAL 157 DAY),NOW(),NOW(),0,26,8,'江眠','https://picsum.photos/seed/avatar8/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,4.5,4.5]',0),
(185,DATE_SUB(NOW(), INTERVAL 116 DAY),NOW(),NOW(),0,26,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[5,4.5,5]',0),
(186,DATE_SUB(NOW(), INTERVAL 167 DAY),NOW(),NOW(),0,27,8,'江眠','https://picsum.photos/seed/avatar8/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,5,5]',0),
(187,DATE_SUB(NOW(), INTERVAL 156 DAY),NOW(),NOW(),0,27,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[4.5,4.5,4.5]',0),
(188,DATE_SUB(NOW(), INTERVAL 71 DAY),NOW(),NOW(),0,27,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[4.5,4.5,4.5]',0),
(189,DATE_SUB(NOW(), INTERVAL 86 DAY),NOW(),NOW(),0,27,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[5,5,5]',0),
(190,DATE_SUB(NOW(), INTERVAL 16 DAY),NOW(),NOW(),0,27,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','唯一小建议是希望能加个全身镜，其他都很满意。','[4.5,4.5,4.5]',0),
(191,DATE_SUB(NOW(), INTERVAL 96 DAY),NOW(),NOW(),0,27,2,'周予安','https://picsum.photos/seed/avatar2/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[4.5,4.5,5]',0),
(192,DATE_SUB(NOW(), INTERVAL 166 DAY),NOW(),NOW(),0,27,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[4.5,4.5,5]',0),
(193,DATE_SUB(NOW(), INTERVAL 49 DAY),NOW(),NOW(),0,27,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[4.5,5,4.5]',0),
(194,DATE_SUB(NOW(), INTERVAL 26 DAY),NOW(),NOW(),0,27,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,5,5]',0),
(195,DATE_SUB(NOW(), INTERVAL 161 DAY),NOW(),NOW(),0,27,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,5,5]',0),
(196,DATE_SUB(NOW(), INTERVAL 12 DAY),NOW(),NOW(),0,27,1,'林小满','https://picsum.photos/seed/avatar1/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,5,4.5]',0),
(197,DATE_SUB(NOW(), INTERVAL 90 DAY),NOW(),NOW(),0,27,1,'林小满','https://picsum.photos/seed/avatar1/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,5,5]',0),
(198,DATE_SUB(NOW(), INTERVAL 82 DAY),NOW(),NOW(),0,28,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,4.5,4.5]',0),
(199,DATE_SUB(NOW(), INTERVAL 80 DAY),NOW(),NOW(),0,28,1,'林小满','https://picsum.photos/seed/avatar1/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,4.5,4.5]',0),
(200,DATE_SUB(NOW(), INTERVAL 103 DAY),NOW(),NOW(),0,28,1,'林小满','https://picsum.photos/seed/avatar1/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,5,5]',0),
(201,DATE_SUB(NOW(), INTERVAL 156 DAY),NOW(),NOW(),0,28,1,'林小满','https://picsum.photos/seed/avatar1/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,4.5,4.5]',0),
(202,DATE_SUB(NOW(), INTERVAL 34 DAY),NOW(),NOW(),0,28,1,'林小满','https://picsum.photos/seed/avatar1/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,5,5]',0),
(203,DATE_SUB(NOW(), INTERVAL 159 DAY),NOW(),NOW(),0,28,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[4.5,5,4.5]',0),
(204,DATE_SUB(NOW(), INTERVAL 6 DAY),NOW(),NOW(),0,28,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[5,4.5,5]',0),
(205,DATE_SUB(NOW(), INTERVAL 136 DAY),NOW(),NOW(),0,28,1,'林小满','https://picsum.photos/seed/avatar1/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,5,5]',0),
(206,DATE_SUB(NOW(), INTERVAL 166 DAY),NOW(),NOW(),0,28,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[4.5,5,4.5]',0),
(207,DATE_SUB(NOW(), INTERVAL 170 DAY),NOW(),NOW(),0,28,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[5,5,5]',0),
(208,DATE_SUB(NOW(), INTERVAL 41 DAY),NOW(),NOW(),0,28,1,'林小满','https://picsum.photos/seed/avatar1/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[4.5,4.5,4.5]',0),
(209,DATE_SUB(NOW(), INTERVAL 169 DAY),NOW(),NOW(),0,28,2,'周予安','https://picsum.photos/seed/avatar2/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[5,4.5,5]',0),
(210,DATE_SUB(NOW(), INTERVAL 20 DAY),NOW(),NOW(),0,29,1,'林小满','https://picsum.photos/seed/avatar1/200/200','窗外视野绝了，早上拉开窗帘就是大片，拍照随手出片，下次来还住这家。','[4.5,5,5]',0),
(211,DATE_SUB(NOW(), INTERVAL 153 DAY),NOW(),NOW(),0,29,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[4.5,4.5,4.5]',0),
(212,DATE_SUB(NOW(), INTERVAL 125 DAY),NOW(),NOW(),0,29,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[4.5,4.5,5]',0),
(213,DATE_SUB(NOW(), INTERVAL 152 DAY),NOW(),NOW(),0,29,1,'林小满','https://picsum.photos/seed/avatar1/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[4.5,4.5,4.5]',0),
(214,DATE_SUB(NOW(), INTERVAL 111 DAY),NOW(),NOW(),0,29,8,'江眠','https://picsum.photos/seed/avatar8/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,4.5,4.5]',0),
(215,DATE_SUB(NOW(), INTERVAL 131 DAY),NOW(),NOW(),0,30,2,'周予安','https://picsum.photos/seed/avatar2/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[4.5,5,5]',0),
(216,DATE_SUB(NOW(), INTERVAL 171 DAY),NOW(),NOW(),0,30,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,5,5]',0),
(217,DATE_SUB(NOW(), INTERVAL 26 DAY),NOW(),NOW(),0,30,8,'江眠','https://picsum.photos/seed/avatar8/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,5,5]',0),
(218,DATE_SUB(NOW(), INTERVAL 67 DAY),NOW(),NOW(),0,30,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[4.5,4.5,5]',0),
(219,DATE_SUB(NOW(), INTERVAL 40 DAY),NOW(),NOW(),0,30,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[4.5,4.5,4.5]',0),
(220,DATE_SUB(NOW(), INTERVAL 173 DAY),NOW(),NOW(),0,30,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[4.5,4.5,4.5]',0),
(221,DATE_SUB(NOW(), INTERVAL 96 DAY),NOW(),NOW(),0,30,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[4.5,5,4.5]',0),
(222,DATE_SUB(NOW(), INTERVAL 57 DAY),NOW(),NOW(),0,30,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,5,5]',0),
(223,DATE_SUB(NOW(), INTERVAL 47 DAY),NOW(),NOW(),0,30,2,'周予安','https://picsum.photos/seed/avatar2/200/200','空调给力，热水很快，洗澡体验很好。','[5,4.5,5]',0),
(224,DATE_SUB(NOW(), INTERVAL 157 DAY),NOW(),NOW(),0,30,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[4.5,4.5,5]',0),
(225,DATE_SUB(NOW(), INTERVAL 37 DAY),NOW(),NOW(),0,30,1,'林小满','https://picsum.photos/seed/avatar1/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,5,5]',0),
(226,DATE_SUB(NOW(), INTERVAL 168 DAY),NOW(),NOW(),0,30,1,'林小满','https://picsum.photos/seed/avatar1/200/200','空调给力，热水很快，洗澡体验很好。','[4.5,4.5,4.5]',0),
(227,DATE_SUB(NOW(), INTERVAL 120 DAY),NOW(),NOW(),0,31,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[4.5,5,4.5]',0),
(228,DATE_SUB(NOW(), INTERVAL 30 DAY),NOW(),NOW(),0,31,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,5,4.5]',0),
(229,DATE_SUB(NOW(), INTERVAL 57 DAY),NOW(),NOW(),0,31,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[5,5,4.5]',0),
(230,DATE_SUB(NOW(), INTERVAL 3 DAY),NOW(),NOW(),0,31,8,'江眠','https://picsum.photos/seed/avatar8/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[4.5,4.5,5]',0),
(231,DATE_SUB(NOW(), INTERVAL 19 DAY),NOW(),NOW(),0,31,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[4.5,4.5,5]',0),
(232,DATE_SUB(NOW(), INTERVAL 56 DAY),NOW(),NOW(),0,31,8,'江眠','https://picsum.photos/seed/avatar8/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[4.5,4.5,5]',0),
(233,DATE_SUB(NOW(), INTERVAL 68 DAY),NOW(),NOW(),0,31,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[5,5,4.5]',0),
(234,DATE_SUB(NOW(), INTERVAL 105 DAY),NOW(),NOW(),0,31,8,'江眠','https://picsum.photos/seed/avatar8/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,5,5]',0),
(235,DATE_SUB(NOW(), INTERVAL 99 DAY),NOW(),NOW(),0,31,1,'林小满','https://picsum.photos/seed/avatar1/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[4.5,4.5,5]',0),
(236,DATE_SUB(NOW(), INTERVAL 147 DAY),NOW(),NOW(),0,31,8,'江眠','https://picsum.photos/seed/avatar8/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,5,4.5]',0),
(237,DATE_SUB(NOW(), INTERVAL 124 DAY),NOW(),NOW(),0,32,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[4.5,4.5,4.5]',0),
(238,DATE_SUB(NOW(), INTERVAL 180 DAY),NOW(),NOW(),0,32,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[4.5,4.5,5]',0),
(239,DATE_SUB(NOW(), INTERVAL 123 DAY),NOW(),NOW(),0,32,8,'江眠','https://picsum.photos/seed/avatar8/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,5,5]',0),
(240,DATE_SUB(NOW(), INTERVAL 6 DAY),NOW(),NOW(),0,32,8,'江眠','https://picsum.photos/seed/avatar8/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[5,5,4.5]',0),
(241,DATE_SUB(NOW(), INTERVAL 170 DAY),NOW(),NOW(),0,32,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,4.5,5]',0),
(242,DATE_SUB(NOW(), INTERVAL 149 DAY),NOW(),NOW(),0,32,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[4.5,4.5,5]',0),
(243,DATE_SUB(NOW(), INTERVAL 175 DAY),NOW(),NOW(),0,32,1,'林小满','https://picsum.photos/seed/avatar1/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[5,5,5]',0),
(244,DATE_SUB(NOW(), INTERVAL 9 DAY),NOW(),NOW(),0,32,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[4.5,4.5,5]',0),
(245,DATE_SUB(NOW(), INTERVAL 178 DAY),NOW(),NOW(),0,32,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[5,4.5,4.5]',0),
(246,DATE_SUB(NOW(), INTERVAL 115 DAY),NOW(),NOW(),0,32,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,5,5]',0),
(247,DATE_SUB(NOW(), INTERVAL 173 DAY),NOW(),NOW(),0,32,2,'周予安','https://picsum.photos/seed/avatar2/200/200','离景区很近，早起去看日出非常方便，省了打车钱。','[4.5,5,5]',0),
(248,DATE_SUB(NOW(), INTERVAL 175 DAY),NOW(),NOW(),0,32,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,5,4.5]',0),
(249,DATE_SUB(NOW(), INTERVAL 164 DAY),NOW(),NOW(),0,33,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[5,4.5,4.5]',0),
(250,DATE_SUB(NOW(), INTERVAL 21 DAY),NOW(),NOW(),0,33,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','空调给力，热水很快，洗澡体验很好。','[5,4.5,5]',0),
(251,DATE_SUB(NOW(), INTERVAL 39 DAY),NOW(),NOW(),0,33,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[5,4.5,4.5]',0),
(252,DATE_SUB(NOW(), INTERVAL 121 DAY),NOW(),NOW(),0,33,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[4.5,4.5,5]',0),
(253,DATE_SUB(NOW(), INTERVAL 117 DAY),NOW(),NOW(),0,33,1,'林小满','https://picsum.photos/seed/avatar1/200/200','空调给力，热水很快，洗澡体验很好。','[5,5,4.5]',0),
(254,DATE_SUB(NOW(), INTERVAL 29 DAY),NOW(),NOW(),0,33,1,'林小满','https://picsum.photos/seed/avatar1/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,5,4.5]',0),
(255,DATE_SUB(NOW(), INTERVAL 19 DAY),NOW(),NOW(),0,33,8,'江眠','https://picsum.photos/seed/avatar8/200/200','唯一小建议是希望能加个全身镜，其他都很满意。','[4.5,4.5,5]',0),
(256,DATE_SUB(NOW(), INTERVAL 57 DAY),NOW(),NOW(),0,34,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[5,5,4.5]',0),
(257,DATE_SUB(NOW(), INTERVAL 61 DAY),NOW(),NOW(),0,34,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,4.5,4.5]',0),
(258,DATE_SUB(NOW(), INTERVAL 129 DAY),NOW(),NOW(),0,34,2,'周予安','https://picsum.photos/seed/avatar2/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[5,4.5,4.5]',0),
(259,DATE_SUB(NOW(), INTERVAL 16 DAY),NOW(),NOW(),0,34,2,'周予安','https://picsum.photos/seed/avatar2/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[4.5,4.5,4.5]',0),
(260,DATE_SUB(NOW(), INTERVAL 147 DAY),NOW(),NOW(),0,35,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[4.5,5,4.5]',0),
(261,DATE_SUB(NOW(), INTERVAL 3 DAY),NOW(),NOW(),0,35,1,'林小满','https://picsum.photos/seed/avatar1/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,4.5,5]',0),
(262,DATE_SUB(NOW(), INTERVAL 82 DAY),NOW(),NOW(),0,35,1,'林小满','https://picsum.photos/seed/avatar1/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[4.5,4.5,4.5]',0),
(263,DATE_SUB(NOW(), INTERVAL 88 DAY),NOW(),NOW(),0,35,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[5,4.5,5]',0),
(264,DATE_SUB(NOW(), INTERVAL 36 DAY),NOW(),NOW(),0,35,1,'林小满','https://picsum.photos/seed/avatar1/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,5,4.5]',0),
(265,DATE_SUB(NOW(), INTERVAL 101 DAY),NOW(),NOW(),0,35,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[5,4.5,4.5]',0),
(266,DATE_SUB(NOW(), INTERVAL 73 DAY),NOW(),NOW(),0,35,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','唯一小建议是希望能加个全身镜，其他都很满意。','[5,5,5]',0),
(267,DATE_SUB(NOW(), INTERVAL 79 DAY),NOW(),NOW(),0,35,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[4.5,4.5,5]',0),
(268,DATE_SUB(NOW(), INTERVAL 51 DAY),NOW(),NOW(),0,35,8,'江眠','https://picsum.photos/seed/avatar8/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[4.5,4.5,5]',0),
(269,DATE_SUB(NOW(), INTERVAL 9 DAY),NOW(),NOW(),0,36,8,'江眠','https://picsum.photos/seed/avatar8/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[5,4.5,4.5]',0),
(270,DATE_SUB(NOW(), INTERVAL 48 DAY),NOW(),NOW(),0,36,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,4.5,5]',0),
(271,DATE_SUB(NOW(), INTERVAL 14 DAY),NOW(),NOW(),0,36,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[5,5,4.5]',0),
(272,DATE_SUB(NOW(), INTERVAL 149 DAY),NOW(),NOW(),0,36,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,4.5,4.5]',0),
(273,DATE_SUB(NOW(), INTERVAL 51 DAY),NOW(),NOW(),0,36,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[4.5,4.5,4.5]',0),
(274,DATE_SUB(NOW(), INTERVAL 11 DAY),NOW(),NOW(),0,36,2,'周予安','https://picsum.photos/seed/avatar2/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[4.5,5,4.5]',0),
(275,DATE_SUB(NOW(), INTERVAL 125 DAY),NOW(),NOW(),0,37,8,'江眠','https://picsum.photos/seed/avatar8/200/200','空调给力，热水很快，洗澡体验很好。','[5,5,5]',0),
(276,DATE_SUB(NOW(), INTERVAL 106 DAY),NOW(),NOW(),0,37,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[5,4.5,4.5]',0),
(277,DATE_SUB(NOW(), INTERVAL 80 DAY),NOW(),NOW(),0,37,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[4.5,4.5,5]',0),
(278,DATE_SUB(NOW(), INTERVAL 101 DAY),NOW(),NOW(),0,37,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','前台小哥帮我们搬了行李，还送了水果，服务很nice。','[5,4.5,5]',0),
(279,DATE_SUB(NOW(), INTERVAL 159 DAY),NOW(),NOW(),0,37,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[4.5,4.5,4.5]',0),
(280,DATE_SUB(NOW(), INTERVAL 104 DAY),NOW(),NOW(),0,37,2,'周予安','https://picsum.photos/seed/avatar2/200/200','停车方便，自驾过来完全不用愁，车位就在门口。','[5,5,4.5]',0),
(281,DATE_SUB(NOW(), INTERVAL 19 DAY),NOW(),NOW(),0,37,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','安静、干净、舒服，适合度假躺平，什么都不干也很好。','[4.5,4.5,4.5]',0),
(282,DATE_SUB(NOW(), INTERVAL 131 DAY),NOW(),NOW(),0,37,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[5,4.5,5]',0),
(283,DATE_SUB(NOW(), INTERVAL 104 DAY),NOW(),NOW(),0,37,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','性价比很高，周边吃饭逛街都方便，步行几分钟就到地铁口。','[4.5,5,4.5]',0),
(284,DATE_SUB(NOW(), INTERVAL 134 DAY),NOW(),NOW(),0,38,8,'江眠','https://picsum.photos/seed/avatar8/200/200','整体超出预期，服务态度好，退房也很快，没有套路。','[4.5,5,4.5]',0),
(285,DATE_SUB(NOW(), INTERVAL 126 DAY),NOW(),NOW(),0,38,1,'林小满','https://picsum.photos/seed/avatar1/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[5,5,4.5]',0),
(286,DATE_SUB(NOW(), INTERVAL 32 DAY),NOW(),NOW(),0,38,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,5,5]',0),
(287,DATE_SUB(NOW(), INTERVAL 131 DAY),NOW(),NOW(),0,38,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[4.5,5,4.5]',0),
(288,DATE_SUB(NOW(), INTERVAL 48 DAY),NOW(),NOW(),0,38,2,'周予安','https://picsum.photos/seed/avatar2/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,5,4.5]',0),
(289,DATE_SUB(NOW(), INTERVAL 131 DAY),NOW(),NOW(),0,38,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','入住流程很顺畅，自助入住不用等，管家响应也很快。','[4.5,4.5,5]',0),
(290,DATE_SUB(NOW(), INTERVAL 25 DAY),NOW(),NOW(),0,38,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','设计感很强，拍了很多照片，朋友都在问链接。','[5,5,5]',0),
(291,DATE_SUB(NOW(), INTERVAL 138 DAY),NOW(),NOW(),0,38,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','洗漱用品都是好牌子，毛巾浴袍很新，细节做得到位。','[5,4.5,4.5]',0),
(292,DATE_SUB(NOW(), INTERVAL 179 DAY),NOW(),NOW(),0,38,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','空调给力，热水很快，洗澡体验很好。','[5,5,4.5]',0),
(293,DATE_SUB(NOW(), INTERVAL 23 DAY),NOW(),NOW(),0,38,1,'林小满','https://picsum.photos/seed/avatar1/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,4.5,5]',0),
(294,DATE_SUB(NOW(), INTERVAL 28 DAY),NOW(),NOW(),0,38,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','床垫软硬适中，枕头有高低两种可选，睡眠质量很好。','[5,5,4.5]',0),
(295,DATE_SUB(NOW(), INTERVAL 136 DAY),NOW(),NOW(),0,39,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[5,5,4.5]',0),
(296,DATE_SUB(NOW(), INTERVAL 47 DAY),NOW(),NOW(),0,39,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[4.5,4.5,4.5]',0),
(297,DATE_SUB(NOW(), INTERVAL 59 DAY),NOW(),NOW(),0,39,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','房间收拾得很干净，隔音也不错，晚上很安静，睡得很踏实。','[4.5,4.5,5]',0),
(298,DATE_SUB(NOW(), INTERVAL 120 DAY),NOW(),NOW(),0,39,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','房间采光太棒了，白天基本不用开灯，很通透。','[5,4.5,5]',0),
(299,DATE_SUB(NOW(), INTERVAL 65 DAY),NOW(),NOW(),0,39,7,'顾清和','https://picsum.photos/seed/avatar7/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[4.5,5,4.5]',0),
(300,DATE_SUB(NOW(), INTERVAL 37 DAY),NOW(),NOW(),0,39,6,'许一诺','https://picsum.photos/seed/avatar6/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[4.5,4.5,4.5]',0),
(301,DATE_SUB(NOW(), INTERVAL 53 DAY),NOW(),NOW(),0,39,3,'陈屿','https://picsum.photos/seed/avatar3/200/200','位置非常好找，房东提前发了详细定位，房间比图片还干净，床品很舒服，一觉睡到自然醒。','[5,5,5]',0),
(302,DATE_SUB(NOW(), INTERVAL 22 DAY),NOW(),NOW(),0,39,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,4.5,5]',0),
(303,DATE_SUB(NOW(), INTERVAL 4 DAY),NOW(),NOW(),0,40,5,'沈知白','https://picsum.photos/seed/avatar5/200/200','带娃来的，房间空间够大，房东还准备了儿童用品，很贴心。','[4.5,4.5,4.5]',0),
(304,DATE_SUB(NOW(), INTERVAL 64 DAY),NOW(),NOW(),0,40,1,'林小满','https://picsum.photos/seed/avatar1/200/200','院子打理得很有心思，下午在院子喝茶呆了一整个下午，太放松了。','[5,4.5,4.5]',0),
(305,DATE_SUB(NOW(), INTERVAL 46 DAY),NOW(),NOW(),0,40,1,'林小满','https://picsum.photos/seed/avatar1/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[4.5,4.5,4.5]',0),
(306,DATE_SUB(NOW(), INTERVAL 163 DAY),NOW(),NOW(),0,40,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','房东人特别热情，帮忙推荐了本地人常去的小馆子，比攻略靠谱多了。','[4.5,4.5,4.5]',0),
(307,DATE_SUB(NOW(), INTERVAL 18 DAY),NOW(),NOW(),0,40,4,'苏晚','https://picsum.photos/seed/avatar4/200/200','早餐是现做的，分量足也好吃，比酒店自助舒服。','[5,5,4.5]',0);

-- 同步评价数
UPDATE `homestay` SET `comment_count` = CASE `id`
  WHEN 1 THEN 4
  WHEN 2 THEN 3
  WHEN 3 THEN 9
  WHEN 4 THEN 9
  WHEN 5 THEN 5
  WHEN 6 THEN 12
  WHEN 7 THEN 7
  WHEN 8 THEN 3
  WHEN 9 THEN 11
  WHEN 10 THEN 9
  WHEN 11 THEN 7
  WHEN 12 THEN 11
  WHEN 13 THEN 4
  WHEN 14 THEN 5
  WHEN 15 THEN 8
  WHEN 16 THEN 4
  WHEN 17 THEN 12
  WHEN 18 THEN 4
  WHEN 19 THEN 6
  WHEN 20 THEN 6
  WHEN 21 THEN 3
  WHEN 22 THEN 7
  WHEN 23 THEN 12
  WHEN 24 THEN 3
  WHEN 25 THEN 11
  WHEN 26 THEN 10
  WHEN 27 THEN 12
  WHEN 28 THEN 12
  WHEN 29 THEN 5
  WHEN 30 THEN 12
  WHEN 31 THEN 10
  WHEN 32 THEN 12
  WHEN 33 THEN 7
  WHEN 34 THEN 4
  WHEN 35 THEN 9
  WHEN 36 THEN 6
  WHEN 37 THEN 9
  WHEN 38 THEN 11
  WHEN 39 THEN 8
  WHEN 40 THEN 5
  ELSE `comment_count` END;

-- ---------- 优选民宿活动位 ----------
INSERT INTO `homestay_activity` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`row_type`,`data_id`,`row_status`,`version`) VALUES
(1,NOW(),NOW(),NOW(),0,'preferredHomestay',1,1,0),
(2,NOW(),NOW(),NOW(),0,'preferredHomestay',5,1,0),
(3,NOW(),NOW(),NOW(),0,'preferredHomestay',6,1,0),
(4,NOW(),NOW(),NOW(),0,'preferredHomestay',10,1,0),
(5,NOW(),NOW(),NOW(),0,'preferredHomestay',11,1,0),
(6,NOW(),NOW(),NOW(),0,'preferredHomestay',15,1,0),
(7,NOW(),NOW(),NOW(),0,'preferredHomestay',16,1,0),
(8,NOW(),NOW(),NOW(),0,'preferredHomestay',20,1,0),
(9,NOW(),NOW(),NOW(),0,'preferredHomestay',21,1,0),
(10,NOW(),NOW(),NOW(),0,'preferredHomestay',25,1,0),
(11,NOW(),NOW(),NOW(),0,'preferredHomestay',26,1,0),
(12,NOW(),NOW(),NOW(),0,'preferredHomestay',30,1,0),
(13,NOW(),NOW(),NOW(),0,'preferredHomestay',31,1,0),
(14,NOW(),NOW(),NOW(),0,'preferredHomestay',35,1,0),
(15,NOW(),NOW(),NOW(),0,'preferredHomestay',36,1,0),
(16,NOW(),NOW(),NOW(),0,'preferredHomestay',40,1,0);

-- 房东推荐位（供 goodBoss 使用）
INSERT INTO `homestay_activity` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`row_type`,`data_id`,`row_status`,`version`) VALUES
(101,NOW(),NOW(),NOW(),0,'goodBusiness',1,1,0),
(102,NOW(),NOW(),NOW(),0,'goodBusiness',2,1,0),
(103,NOW(),NOW(),NOW(),0,'goodBusiness',3,1,0),
(104,NOW(),NOW(),NOW(),0,'goodBusiness',4,1,0),
(105,NOW(),NOW(),NOW(),0,'goodBusiness',5,1,0),
(106,NOW(),NOW(),NOW(),0,'goodBusiness',6,1,0),
(107,NOW(),NOW(),NOW(),0,'goodBusiness',7,1,0),
(108,NOW(),NOW(),NOW(),0,'goodBusiness',8,1,0);
