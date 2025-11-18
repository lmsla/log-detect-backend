-- ======================================================
-- Log Detect - 添加 Elasticsearch 監控權限
-- ======================================================
-- 使用方式: mysql -u root -p logdetect < scripts/add_elasticsearch_permissions.sql
--
-- 此腳本會:
-- 1. 添加 4 個 elasticsearch 權限
-- 2. 將權限分配給 admin 角色
-- 3. 驗證權限設定
-- ======================================================

USE logdetect;

-- ============================================
-- 步驟 1: 添加 Elasticsearch 權限
-- ============================================

INSERT INTO permissions (name, resource, action, description, created_at, updated_at)
VALUES
  ('elasticsearch:create', 'elasticsearch', 'create', 'Create Elasticsearch monitors', NOW(), NOW()),
  ('elasticsearch:read', 'elasticsearch', 'read', 'Read Elasticsearch monitors', NOW(), NOW()),
  ('elasticsearch:update', 'elasticsearch', 'update', 'Update Elasticsearch monitors', NOW(), NOW()),
  ('elasticsearch:delete', 'elasticsearch', 'delete', 'Delete Elasticsearch monitors', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  updated_at = NOW();

SELECT '✅ Elasticsearch 權限已添加/更新' AS Status;

-- ============================================
-- 步驟 2: 獲取角色和權限 ID
-- ============================================

-- 查看 admin 角色 ID
SELECT id, name, description FROM roles WHERE name = 'admin';

-- 查看 elasticsearch 權限 ID
SELECT id, name, resource, action FROM permissions WHERE resource = 'elasticsearch';

-- ============================================
-- 步驟 3: 將權限分配給 admin 角色
-- ============================================

-- 為 admin 角色分配所有 elasticsearch 權限
INSERT INTO role_permissions (role_id, permission_id)
SELECT
  r.id AS role_id,
  p.id AS permission_id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
  AND p.resource = 'elasticsearch'
ON DUPLICATE KEY UPDATE
  role_id = role_id;  -- 如果已存在則不做任何事

SELECT '✅ 權限已分配給 admin 角色' AS Status;

-- ============================================
-- 步驟 4: 驗證權限設定
-- ============================================

-- 查看所有 elasticsearch 權限
SELECT
  p.id,
  p.name,
  p.resource,
  p.action,
  p.description
FROM permissions p
WHERE p.resource = 'elasticsearch'
ORDER BY p.action;

-- 查看 admin 角色的所有權限（按資源分組）
SELECT
  p.resource,
  COUNT(*) AS permission_count,
  GROUP_CONCAT(p.action ORDER BY p.action) AS actions
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.name = 'admin'
GROUP BY p.resource
ORDER BY p.resource;

-- 驗證 admin 用戶的 elasticsearch 權限
SELECT
  u.id AS user_id,
  u.username,
  u.email,
  r.name AS role_name,
  p.name AS permission_name,
  p.resource,
  p.action
FROM users u
JOIN roles r ON u.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE u.username = 'admin'
  AND p.resource = 'elasticsearch'
ORDER BY p.action;

-- ============================================
-- 步驟 5: 統計摘要
-- ============================================

SELECT '📊 權限統計摘要' AS '';

-- 各資源的權限數量
SELECT
  resource,
  COUNT(*) AS total_permissions
FROM permissions
GROUP BY resource
ORDER BY resource;

-- 各角色的權限數量
SELECT
  r.name AS role_name,
  COUNT(DISTINCT p.id) AS total_permissions,
  COUNT(DISTINCT p.resource) AS total_resources
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
LEFT JOIN permissions p ON rp.permission_id = p.id
GROUP BY r.id, r.name
ORDER BY r.name;

-- ============================================
-- 完成
-- ============================================

SELECT '🎉 Elasticsearch 權限設定完成！' AS '';
SELECT '💡 請重新登入以獲取新的 JWT token' AS '';
