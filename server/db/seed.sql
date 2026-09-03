-- 电气车间隐患闭环系统：示例枚举数据（可选手动执行；服务启动时表为空也会自动写入，等价于此文件）
-- 用法：mysql -u hazard -p -h 127.0.0.1 hazard_system < db/seed.sql

INSERT INTO responsible_units (name, person, sort, status, created_at, updated_at) VALUES
  ('电气车间',   '张三', 1, 1, NOW(), NOW()),
  ('动力车间',   '李四', 2, 1, NOW(), NOW()),
  ('自动化班组', '王五', 3, 1, NOW(), NOW());

INSERT INTO hazard_types (parent_id, name, sort, status, created_at, updated_at) VALUES
  (0, '电气设备', 1, 1, NOW(), NOW()),
  (0, '安全防护', 2, 1, NOW(), NOW());

-- 小类（分类）通过 parent_id 关联上述大类（假设大类 id 为 1、2，实际以插入后 id 为准）
INSERT INTO hazard_types (parent_id, name, sort, status, created_at, updated_at) VALUES
  (1, '线路老化',     1, 1, NOW(), NOW()),
  (1, '接线不规范',   2, 1, NOW(), NOW()),
  (1, '绝缘破损',     3, 1, NOW(), NOW()),
  (2, '警示标识缺失', 1, 1, NOW(), NOW()),
  (2, '防护罩缺失',   2, 1, NOW(), NOW());