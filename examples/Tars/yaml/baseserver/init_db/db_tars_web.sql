-- MySQL dump 10.13  Distrib 5.7.29, for Linux (x86_64)
--
-- Host: 127.0.0.1    Database: db_tars_web
-- ------------------------------------------------------
-- Server version	5.6.47

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `t_bm_case`
--

DROP TABLE IF EXISTS `t_bm_case`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_bm_case` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `servant` varchar(128) NOT NULL,
  `fn` varchar(64) NOT NULL,
  `des` varchar(256) DEFAULT '',
  `in_values` text,
  `endpoints` text,
  `links` int(11) DEFAULT NULL,
  `speed` int(11) DEFAULT NULL,
  `duration` int(11) DEFAULT NULL,
  `posttime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `is_deleted` int(11) NOT NULL DEFAULT '0',
  `status` int(11) NOT NULL DEFAULT '0',
  `results` text,
  PRIMARY KEY (`id`),
  KEY `t_bm_case_servant_fn` (`servant`,`fn`)
) ENGINE=InnoDB CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_kafka_queue`
--

DROP TABLE IF EXISTS `t_kafka_queue`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_kafka_queue` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `topic` varchar(16) NOT NULL DEFAULT '',
  `partition` int(4) NOT NULL DEFAULT '0',
  `offset` int(11) NOT NULL DEFAULT '0',
  `task_no` varchar(64) NOT NULL DEFAULT '' COMMENT '任务ID',
  `status` varchar(16) NOT NULL DEFAULT 'waiting' COMMENT '任务状态',
  `message` varchar(256) DEFAULT '',
  PRIMARY KEY (`id`,`task_no`,`status`)
) ENGINE=InnoDB CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_patch_task`
--

DROP TABLE IF EXISTS `t_patch_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_patch_task` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `server` varchar(50) DEFAULT NULL,
  `tgz` text,
  `task_id` varchar(64) DEFAULT NULL COMMENT '任务ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_tars_files`
--

DROP TABLE IF EXISTS `t_tars_files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_tars_files` (
  `f_id` int(11) NOT NULL AUTO_INCREMENT,
  `application` varchar(64) NOT NULL DEFAULT '' COMMENT '应用名',
  `server_name` varchar(128) NOT NULL DEFAULT '' COMMENT '服务名',
  `file_name` varchar(64) NOT NULL DEFAULT '' COMMENT '文件名',
  `posttime` datetime DEFAULT NULL COMMENT '更新时间',
  `context` text COMMENT '解析后的JSON对象',
  `benchmark_context` text,
  PRIMARY KEY (`server_name`,`file_name`),
  UNIQUE KEY `f_id` (`f_id`) USING BTREE
) ENGINE=InnoDB CHARSET=utf8 COMMENT='接口测试tars文件表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2020-06-20 17:02:14
