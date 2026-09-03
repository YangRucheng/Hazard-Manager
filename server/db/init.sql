-- 电气车间隐患闭环系统：数据库初始化（建库 + 建用户 + 授权）
-- 用法：mysql -u root < db/init.sql
-- 注意：生产环境请修改密码并与 .env 中 DB_PASSWORD 保持一致。

CREATE DATABASE IF NOT EXISTS hazard_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 应用账号（本地开发默认密码；生产请替换）
CREATE USER IF NOT EXISTS 'hazard'@'localhost' IDENTIFIED BY 'hazard_dev_password';
CREATE USER IF NOT EXISTS 'hazard'@'127.0.0.1' IDENTIFIED BY 'hazard_dev_password';

GRANT ALL PRIVILEGES ON hazard_system.* TO 'hazard'@'localhost';
GRANT ALL PRIVILEGES ON hazard_system.* TO 'hazard'@'127.0.0.1';
FLUSH PRIVILEGES;