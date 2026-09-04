-- 电气车间隐患闭环系统：示例枚举数据（可选手动执行；服务启动时表为空也会自动写入，等价于此文件）
-- 用法：mysql -u hazard -p -h 127.0.0.1 hazard_system < db/seed.sql

INSERT INTO responsible_units (name, person, sort, status, created_at, updated_at) VALUES
  ('电气车间',   '张三', 1, 1, NOW(), NOW()),
  ('动力车间',   '李四', 2, 1, NOW(), NOW()),
  ('自动化班组', '王五', 3, 1, NOW(), NOW());

-- 隐患类型：每行一个「大类(major)+小类(minor)」组合，无父级引用
INSERT INTO hazard_types (major, minor, created_at, updated_at) VALUES
  ('电气设备', '线路老化',     NOW(), NOW()),
  ('电气设备', '接线不规范',   NOW(), NOW()),
  ('电气设备', '绝缘破损',     NOW(), NOW()),
  ('安全防护', '警示标识缺失', NOW(), NOW()),
  ('安全防护', '防护罩缺失',   NOW(), NOW());
