-- =============================================================
-- looklook 演示用户 / demo users（密码统一 123456）
-- 由 deploy/script/seed/main.mjs 生成，请勿手改
-- 前置：先执行 looklook_usercenter.sql
-- =============================================================

SET NAMES utf8mb4;

TRUNCATE TABLE `user`;
TRUNCATE TABLE `user_auth`;
INSERT INTO `user` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`version`,`mobile`,`password`,`nickname`,`sex`,`avatar`,`info`) VALUES
(1,NOW(),NOW(),NOW(),0,0,'13800000001','e10adc3949ba59abbe56e057f20f883e','林小满',0,'https://picsum.photos/seed/avatar1/200/200','喜欢到处走走，住遍好看的房子。'),
(2,NOW(),NOW(),NOW(),0,0,'13800000002','e10adc3949ba59abbe56e057f20f883e','周予安',1,'https://picsum.photos/seed/avatar2/200/200','喜欢到处走走，住遍好看的房子。'),
(3,NOW(),NOW(),NOW(),0,0,'13800000003','e10adc3949ba59abbe56e057f20f883e','陈屿',0,'https://picsum.photos/seed/avatar3/200/200','喜欢到处走走，住遍好看的房子。'),
(4,NOW(),NOW(),NOW(),0,0,'13800000004','e10adc3949ba59abbe56e057f20f883e','苏晚',1,'https://picsum.photos/seed/avatar4/200/200','喜欢到处走走，住遍好看的房子。'),
(5,NOW(),NOW(),NOW(),0,0,'13800000005','e10adc3949ba59abbe56e057f20f883e','沈知白',0,'https://picsum.photos/seed/avatar5/200/200','喜欢到处走走，住遍好看的房子。'),
(6,NOW(),NOW(),NOW(),0,0,'13800000006','e10adc3949ba59abbe56e057f20f883e','许一诺',1,'https://picsum.photos/seed/avatar6/200/200','喜欢到处走走，住遍好看的房子。'),
(7,NOW(),NOW(),NOW(),0,0,'13800000007','e10adc3949ba59abbe56e057f20f883e','顾清和',0,'https://picsum.photos/seed/avatar7/200/200','喜欢到处走走，住遍好看的房子。'),
(8,NOW(),NOW(),NOW(),0,0,'13800000008','e10adc3949ba59abbe56e057f20f883e','江眠',1,'https://picsum.photos/seed/avatar8/200/200','喜欢到处走走，住遍好看的房子。');

INSERT INTO `user_auth` (`id`,`create_time`,`update_time`,`delete_time`,`del_state`,`version`,`user_id`,`auth_key`,`auth_type`) VALUES
(1,NOW(),NOW(),NOW(),0,0,1,'13800000001','system'),
(2,NOW(),NOW(),NOW(),0,0,2,'13800000002','system'),
(3,NOW(),NOW(),NOW(),0,0,3,'13800000003','system'),
(4,NOW(),NOW(),NOW(),0,0,4,'13800000004','system'),
(5,NOW(),NOW(),NOW(),0,0,5,'13800000005','system'),
(6,NOW(),NOW(),NOW(),0,0,6,'13800000006','system'),
(7,NOW(),NOW(),NOW(),0,0,7,'13800000007','system'),
(8,NOW(),NOW(),NOW(),0,0,8,'13800000008','system');
