-- MySQL dump 10.13  Distrib 5.6.26, for Linux (x86_64)
--
-- Host: localhost    Database: db_tars
-- ------------------------------------------------------
-- Server version	5.6.26-log

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
-- Table structure for table `t_adapter_conf`
--

DROP TABLE IF EXISTS `t_adapter_conf`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_adapter_conf` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `application` varchar(50) DEFAULT '',
  `server_name` varchar(128) DEFAULT '',
  `node_name` varchar(50) DEFAULT '',
  `adapter_name` varchar(100) DEFAULT '',
  `registry_timestamp` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `thread_num` int(11) DEFAULT '1',
  `endpoint` varchar(128) DEFAULT '',
  `max_connections` int(11) DEFAULT '1000',
  `allow_ip` varchar(255) NOT NULL DEFAULT '',
  `servant` varchar(128) DEFAULT '',
  `queuecap` int(11) DEFAULT NULL,
  `queuetimeout` int(11) DEFAULT NULL,
  `posttime` datetime DEFAULT NULL,
  `lastuser` varchar(30) DEFAULT NULL,
  `protocol` varchar(64) DEFAULT 'tars',
  `handlegroup` varchar(64) DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `application` (`application`,`server_name`,`node_name`,`adapter_name`),
  KEY `adapter_conf_endpoint_index` (`endpoint`),
  KEY `index_regtime_1` (`registry_timestamp`)
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_ats_cases`
--

DROP TABLE IF EXISTS `t_ats_cases`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_ats_cases` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `casename` varchar(20) DEFAULT NULL,
  `retvalue` text,
  `paramvalue` text,
  `interfaceid` int(11) DEFAULT NULL,
  `posttime` datetime DEFAULT NULL,
  `lastuser` varchar(30) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_ats_interfaces`
--

DROP TABLE IF EXISTS `t_ats_interfaces`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_ats_interfaces` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `objname` varchar(150) DEFAULT NULL,
  `funcname` varchar(150) DEFAULT NULL,
  `retype` text,
  `paramtype` text,
  `outparamtype` text,
  `interfaceid` int(11) DEFAULT NULL,
  `postime` datetime DEFAULT NULL,
  `lastuser` varchar(30) DEFAULT NULL,
  `request_charset` varchar(16) NOT NULL,
  `response_charset` varchar(16) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `objname` (`objname`,`funcname`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_config_files`
--

DROP TABLE IF EXISTS `t_config_files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_config_files` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `server_name` varchar(128) DEFAULT '',
  `set_name` varchar(16) NOT NULL DEFAULT '',
  `set_area` varchar(16) NOT NULL DEFAULT '',
  `set_group` varchar(16) NOT NULL DEFAULT '',
  `host` varchar(50) NOT NULL DEFAULT '',
  `filename` varchar(128) DEFAULT NULL,
  `config` longtext,
  `posttime` datetime DEFAULT NULL,
  `lastuser` varchar(50) DEFAULT NULL,
  `level` int(11) DEFAULT '2',
  `config_flag` int(10) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `application` (`server_name`,`filename`,`host`,`level`,`set_name`,`set_area`,`set_group`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_config_history_files`
--

DROP TABLE IF EXISTS `t_config_history_files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_config_history_files` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `configid` int(11) DEFAULT NULL,
  `reason` varchar(128) DEFAULT '',
  `reason_select` varchar(20) NOT NULL DEFAULT '',
  `content` longtext,
  `posttime` datetime DEFAULT NULL,
  `lastuser` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_config_references`
--

DROP TABLE IF EXISTS `t_config_references`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_config_references` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `config_id` int(11) DEFAULT NULL,
  `reference_id` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `config_id` (`config_id`,`reference_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_group_priority`
--

DROP TABLE IF EXISTS `t_group_priority`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_group_priority` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) DEFAULT '',
  `group_list` text,
  `list_order` int(11) DEFAULT '0',
  `station` varchar(128) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_machine_tars_info`
--

DROP TABLE IF EXISTS `t_machine_tars_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_machine_tars_info` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `application` varchar(100) NOT NULL DEFAULT '',
  `server_name` varchar(100) NOT NULL DEFAULT '',
  `app_server_name` varchar(50) NOT NULL DEFAULT '',
  `node_name` varchar(50) NOT NULL DEFAULT '',
  `location` varchar(255) NOT NULL DEFAULT '',
  `machine_type` varchar(50) NOT NULL DEFAULT '',
  `update_time` datetime DEFAULT NULL,
  `update_person` varchar(64) NOT NULL DEFAULT '',
  PRIMARY KEY (`application`,`server_name`,`node_name`),
  UNIQUE KEY `id` (`id`),
  KEY `tmachine_i_2` (`node_name`,`server_name`)
) ENGINE=InnoDB AUTO_INCREMENT=23 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_node_info`
--

DROP TABLE IF EXISTS `t_node_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_node_info` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `node_name` varchar(128) DEFAULT '',
  `node_obj` varchar(128) DEFAULT '',
  `endpoint_ip` varchar(50) DEFAULT '',
  `endpoint_port` int(11) DEFAULT '0',
  `data_dir` varchar(128) DEFAULT '',
  `load_avg1` float DEFAULT '0',
  `load_avg5` float DEFAULT '0',
  `load_avg15` float DEFAULT '0',
  `last_reg_time` datetime DEFAULT NULL,
  `last_heartbeat` datetime DEFAULT NULL,
  `setting_state` enum('active','inactive') DEFAULT 'inactive',
  `present_state` enum('active','inactive') DEFAULT 'inactive',
  `tars_version` varchar(128) NOT NULL DEFAULT '',
  `template_name` varchar(128) NOT NULL DEFAULT '',
  `modify_time` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `group_id` int(11) DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `node_name` (`node_name`),
  KEY `indx_node_info_1` (`last_heartbeat`)
) ENGINE=InnoDB AUTO_INCREMENT=68 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_profile_template`
--

DROP TABLE IF EXISTS `t_profile_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_profile_template` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `template_name` varchar(128) DEFAULT '',
  `parents_name` varchar(128) DEFAULT '',
  `profile` text NOT NULL,
  `posttime` datetime DEFAULT NULL,
  `lastuser` varchar(30) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `template_name` (`template_name`)
) ENGINE=InnoDB AUTO_INCREMENT=366 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_registry_info`
--

DROP TABLE IF EXISTS `t_registry_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_registry_info` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `locator_id` varchar(128) NOT NULL DEFAULT '',
  `servant` varchar(128) NOT NULL DEFAULT '',
  `endpoint` varchar(128) NOT NULL DEFAULT '',
  `last_heartbeat` datetime DEFAULT NULL,
  `present_state` enum('active','inactive') DEFAULT 'inactive',
  `tars_version` varchar(128) NOT NULL DEFAULT '',
  `modify_time` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `enable_group` char(1) DEFAULT 'N',
  PRIMARY KEY (`id`),
  UNIQUE KEY `locator_id` (`locator_id`,`servant`)
) ENGINE=InnoDB AUTO_INCREMENT=456487 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_server_conf`
--

DROP TABLE IF EXISTS `t_server_conf`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_server_conf` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `application` varchar(128) DEFAULT '',
  `server_name` varchar(128) DEFAULT '',
  `node_group` varchar(50) NOT NULL DEFAULT '',
  `node_name` varchar(50) NOT NULL DEFAULT '',
  `registry_timestamp` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `base_path` varchar(128) DEFAULT '',
  `exe_path` varchar(128) NOT NULL DEFAULT '',
  `template_name` varchar(128) NOT NULL DEFAULT '',
  `bak_flag` int(11) NOT NULL DEFAULT '0',
  `setting_state` enum('active','inactive') NOT NULL DEFAULT 'inactive',
  `present_state` enum('active','inactive','activating','deactivating','destroyed') NOT NULL DEFAULT 'inactive',
  `process_id` int(11) NOT NULL DEFAULT '0',
  `patch_version` varchar(128) NOT NULL DEFAULT '',
  `patch_time` datetime DEFAULT NULL,
  `patch_user` varchar(128) NOT NULL DEFAULT '',
  `tars_version` varchar(128) NOT NULL DEFAULT '',
  `posttime` datetime DEFAULT NULL,
  `lastuser` varchar(30) DEFAULT NULL,
  `server_type` enum('tars_cpp','not_tars','tars_java', 'tars_nodejs','tars_php','tars_go') NOT NULL DEFAULT 'tars_cpp',
  `start_script_path` varchar(128) DEFAULT NULL,
  `stop_script_path` varchar(128) DEFAULT NULL,
  `monitor_script_path` varchar(128) DEFAULT NULL,
  `enable_group` char(1) DEFAULT 'N',
  `enable_set` char(1) NOT NULL DEFAULT 'N',
  `set_name` varchar(16) DEFAULT NULL,
  `set_area` varchar(16) DEFAULT NULL,
  `set_group` varchar(64) DEFAULT NULL,
  `ip_group_name` varchar(64) DEFAULT NULL,
  `profile` text,
  `config_center_port` int(11) NOT NULL DEFAULT '0',
  `async_thread_num` int(11) DEFAULT '3',
  `server_important_type` enum('0','1','2','3','4','5') DEFAULT '0',
  `remote_log_reserve_time` varchar(32) NOT NULL DEFAULT '65',
  `remote_log_compress_time` varchar(32) NOT NULL DEFAULT '2',
  `remote_log_type` int(1) NOT NULL DEFAULT '0',
  `grid_flag` varchar(16) NOT NULL DEFAULT 'NORMAL',
  PRIMARY KEY (`id`),
  UNIQUE KEY `application` (`application`,`server_name`,`node_name`),
  KEY `node_name` (`node_name`),
  KEY `index_i_3` (`setting_state`,`server_type`,`application`,`server_name`,`node_name`),
  KEY `index_regtime` (`registry_timestamp`)
) ENGINE=InnoDB AUTO_INCREMENT=23 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_server_group_relation`
--

DROP TABLE IF EXISTS `t_server_group_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_server_group_relation` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `application` varchar(90) NOT NULL DEFAULT '',
  `server_group` varchar(50) DEFAULT '',
  `server_name` varchar(50) DEFAULT '',
  `create_time` datetime DEFAULT NULL,
  `creator` varchar(30) DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `f_unique` (`application`,`server_group`,`server_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_server_group_rule`
--

DROP TABLE IF EXISTS `t_server_group_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_server_group_rule` (
  `group_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `ip_order` enum('allow_denny','denny_allow') NOT NULL DEFAULT 'denny_allow',
  `allow_ip_rule` text,
  `denny_ip_rule` text,
  `lastuser` varchar(50) DEFAULT NULL,
  `modify_time` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `group_name` varchar(128) DEFAULT '',
  `group_name_cn` varchar(128) DEFAULT '',
  PRIMARY KEY (`group_id`),
  UNIQUE KEY `group_name_index` (`group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_server_patchs`
--

DROP TABLE IF EXISTS `t_server_patchs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_server_patchs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `server` varchar(50) DEFAULT NULL,
  `version` varchar(1000) DEFAULT '',
  `tgz` text,
  `update_text` varchar(255) DEFAULT NULL,
  `reason_select` varchar(255) DEFAULT NULL,
  `document_complate` varchar(30) DEFAULT NULL,
  `is_server_group` tinyint(2) NOT NULL DEFAULT '0',
  `publish` tinyint(3) DEFAULT NULL,
  `publish_time` datetime DEFAULT NULL,
  `publish_user` varchar(30) DEFAULT NULL,
  `upload_time` datetime DEFAULT NULL,
  `upload_user` varchar(30) DEFAULT NULL,
  `posttime` datetime DEFAULT NULL,
  `lastuser` varchar(30) DEFAULT NULL,
  `is_release_version` enum('true','false') DEFAULT 'true',
  `package_type` tinyint(4) DEFAULT '0',
  `group_id` varchar(64) NOT NULL DEFAULT '',
  `default_version` tinyint(4) DEFAULT '0',
  `md5` varchar(40) DEFAULT NULL,
  `svn_version` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `server_patchs_server_index` (`server`),
  KEY `index_patchs_i1` (`server`),
  KEY `index_i_2` (`tgz`(50))
) ENGINE=InnoDB AUTO_INCREMENT=52 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_task`
--

DROP TABLE IF EXISTS `t_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_task` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `task_no` varchar(40) DEFAULT NULL,
  `serial` tinyint(1) DEFAULT NULL,
  `user_name` varchar(20) DEFAULT NULL,
  `create_time` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `f_task` (`task_no`)
) ENGINE=InnoDB AUTO_INCREMENT=139 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_task_item`
--

DROP TABLE IF EXISTS `t_task_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_task_item` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `task_no` varchar(40) DEFAULT NULL,
  `item_no` varchar(40) DEFAULT NULL,
  `application` varchar(30) DEFAULT NULL,
  `server_name` varchar(50) DEFAULT NULL,
  `node_name` varchar(50) DEFAULT NULL,
  `command` varchar(20) DEFAULT NULL,
  `parameters` text,
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `set_name` varchar(20) DEFAULT NULL,
  `log` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `f_uniq` (`item_no`,`task_no`),
  KEY `f_task_no` (`task_no`),
  KEY `f_index` (`application`,`server_name`,`command`)
) ENGINE=InnoDB AUTO_INCREMENT=142 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `t_web_release_conf`
--

DROP TABLE IF EXISTS `t_web_release_conf`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `t_web_release_conf` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `server` varchar(100) NOT NULL DEFAULT '',
  `path` varchar(200) NOT NULL DEFAULT '',
  `server_dir` varchar(200) NOT NULL DEFAULT '',
  `is_server_group` tinyint(2) NOT NULL DEFAULT '0',
  `enable_batch` tinyint(2) NOT NULL DEFAULT '0',
  `user` varchar(200) NOT NULL DEFAULT '*',
  `posttime` datetime DEFAULT NULL,
  `lastuser` varchar(60) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `server` (`server`,`is_server_group`),
  KEY `web_release_conf_server_index` (`server`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2016-11-24 14:28:02
