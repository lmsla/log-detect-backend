-- 添加告警閾值獨立欄位到 elasticsearch_monitors 表
-- 執行方式：
-- 方法 1（推薦）：使用存儲過程自動檢查欄位是否存在
-- mysql -u monitor -p config < add_threshold_fields.sql
--
-- 方法 2：手動逐條執行（如果欄位已存在會報錯，可忽略）

DELIMITER $$

-- 創建存儲過程來添加欄位（如果不存在）
DROP PROCEDURE IF EXISTS AddColumnIfNotExists$$

CREATE PROCEDURE AddColumnIfNotExists(
    IN tableName VARCHAR(128),
    IN columnName VARCHAR(128),
    IN columnDefinition VARCHAR(512)
)
BEGIN
    DECLARE column_count INT;

    -- 檢查欄位是否已存在
    SELECT COUNT(*) INTO column_count
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = tableName
      AND COLUMN_NAME = columnName;

    -- 如果欄位不存在，則添加
    IF column_count = 0 THEN
        SET @sql = CONCAT('ALTER TABLE `', tableName, '` ADD COLUMN `', columnName, '` ', columnDefinition);
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
        SELECT CONCAT('Added column: ', columnName) AS message;
    ELSE
        SELECT CONCAT('Column already exists: ', columnName) AS message;
    END IF;
END$$

DELIMITER ;

-- 添加獨立閾值欄位
CALL AddColumnIfNotExists('elasticsearch_monitors', 'cpu_usage_high', 'DECIMAL(5,2) COMMENT "CPU使用率-高閾值(%)"');
CALL AddColumnIfNotExists('elasticsearch_monitors', 'cpu_usage_critical', 'DECIMAL(5,2) COMMENT "CPU使用率-危險閾值(%)"');
CALL AddColumnIfNotExists('elasticsearch_monitors', 'memory_usage_high', 'DECIMAL(5,2) COMMENT "記憶體使用率-高閾值(%)"');
CALL AddColumnIfNotExists('elasticsearch_monitors', 'memory_usage_critical', 'DECIMAL(5,2) COMMENT "記憶體使用率-危險閾值(%)"');
CALL AddColumnIfNotExists('elasticsearch_monitors', 'disk_usage_high', 'DECIMAL(5,2) COMMENT "磁碟使用率-高閾值(%)"');
CALL AddColumnIfNotExists('elasticsearch_monitors', 'disk_usage_critical', 'DECIMAL(5,2) COMMENT "磁碟使用率-危險閾值(%)"');
CALL AddColumnIfNotExists('elasticsearch_monitors', 'response_time_high', 'BIGINT COMMENT "響應時間-高閾值(ms)"');
CALL AddColumnIfNotExists('elasticsearch_monitors', 'response_time_critical', 'BIGINT COMMENT "響應時間-危險閾值(ms)"');
CALL AddColumnIfNotExists('elasticsearch_monitors', 'unassigned_shards_threshold', 'INT COMMENT "未分配分片閾值"');

-- 清理存儲過程
DROP PROCEDURE IF EXISTS AddColumnIfNotExists;

-- 驗證欄位
SELECT
    COLUMN_NAME,
    COLUMN_TYPE,
    IS_NULLABLE,
    COLUMN_DEFAULT,
    COLUMN_COMMENT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'elasticsearch_monitors'
  AND COLUMN_NAME IN (
    'cpu_usage_high', 'cpu_usage_critical',
    'memory_usage_high', 'memory_usage_critical',
    'disk_usage_high', 'disk_usage_critical',
    'response_time_high', 'response_time_critical',
    'unassigned_shards_threshold'
  )
ORDER BY ORDINAL_POSITION;

-- 說明：
-- 這些獨立欄位使前端可以使用友好的表單控件（數字輸入框、滑桿等）
-- 而不需要手動編寫 JSON 格式
--
-- 優先級：
-- 1. 獨立欄位（如果設置）
-- 2. alert_threshold JSON 欄位（向後兼容）
-- 3. 預設值

-- 預設值說明：
-- 欄位為 NULL 時，系統會使用以下預設值：
-- - cpu_usage_high: 75.0%
-- - cpu_usage_critical: 85.0%
-- - memory_usage_high: 80.0%
-- - memory_usage_critical: 90.0%
-- - disk_usage_high: 85.0%
-- - disk_usage_critical: 95.0%
-- - response_time_high: 3000ms
-- - response_time_critical: 10000ms
-- - unassigned_shards_threshold: 1

-- 使用範例：

-- 1. 創建監控時設置閾值（推薦方式）
/*
INSERT INTO elasticsearch_monitors (
  name, host, port, check_type, interval,
  cpu_usage_high, cpu_usage_critical,
  memory_usage_high, memory_usage_critical,
  disk_usage_high, disk_usage_critical,
  response_time_high, response_time_critical,
  unassigned_shards_threshold
) VALUES (
  'Production ES', 'localhost', 9200, 'health,performance', 60,
  70.0, 80.0,  -- CPU 閾值
  75.0, 85.0,  -- 記憶體閾值
  80.0, 90.0,  -- 磁碟閾值
  2000, 5000,  -- 響應時間閾值（ms）
  2            -- 未分配分片閾值
);
*/

-- 2. 更新現有監控的閾值
/*
UPDATE elasticsearch_monitors
SET
  cpu_usage_high = 70.0,
  cpu_usage_critical = 80.0,
  memory_usage_high = 75.0,
  memory_usage_critical = 85.0
WHERE id = 1;
*/

-- 3. 使用預設閾值模板

-- 寬鬆模板（開發環境）
/*
UPDATE elasticsearch_monitors
SET
  cpu_usage_high = 85.0, cpu_usage_critical = 95.0,
  memory_usage_high = 85.0, memory_usage_critical = 95.0,
  disk_usage_high = 90.0, disk_usage_critical = 98.0,
  response_time_high = 5000, response_time_critical = 15000,
  unassigned_shards_threshold = 5
WHERE id = ?;
*/

-- 標準模板（一般生產環境）
/*
UPDATE elasticsearch_monitors
SET
  cpu_usage_high = 75.0, cpu_usage_critical = 85.0,
  memory_usage_high = 80.0, memory_usage_critical = 90.0,
  disk_usage_high = 85.0, disk_usage_critical = 95.0,
  response_time_high = 3000, response_time_critical = 10000,
  unassigned_shards_threshold = 1
WHERE id = ?;
*/

-- 嚴格模板（核心業務）
/*
UPDATE elasticsearch_monitors
SET
  cpu_usage_high = 60.0, cpu_usage_critical = 70.0,
  memory_usage_high = 70.0, memory_usage_critical = 80.0,
  disk_usage_high = 75.0, disk_usage_critical = 85.0,
  response_time_high = 1000, response_time_critical = 3000,
  unassigned_shards_threshold = 0
WHERE id = ?;
*/

-- 4. 清除閾值設置（使用預設值）
/*
UPDATE elasticsearch_monitors
SET
  cpu_usage_high = NULL,
  cpu_usage_critical = NULL,
  memory_usage_high = NULL,
  memory_usage_critical = NULL,
  disk_usage_high = NULL,
  disk_usage_critical = NULL,
  response_time_high = NULL,
  response_time_critical = NULL,
  unassigned_shards_threshold = NULL
WHERE id = ?;
*/
