-- ======================================================
-- TimescaleDB es_metrics 表結構檢查與修復腳本
-- ======================================================
-- 使用方式: psql -U logdetect -d monitoring -f scripts/check_and_fix_es_metrics_table.sql
--
-- 此腳本會:
-- 1. 檢查 es_metrics 表的當前結構
-- 2. 比對缺少的欄位
-- 3. 安全地添加缺少的欄位
-- ======================================================

\echo '=== TimescaleDB es_metrics 表結構檢查 ==='
\echo ''

-- ============================================
-- 步驟 1: 檢查當前表結構
-- ============================================

\echo '📌 步驟 1: 檢查當前表結構'
\echo ''

SELECT
    column_name,
    data_type,
    column_default,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'es_metrics'
ORDER BY ordinal_position;

\echo ''
\echo '--- 當前欄位數量 ---'
SELECT COUNT(*) as current_column_count
FROM information_schema.columns
WHERE table_name = 'es_metrics';

\echo ''

-- ============================================
-- 步驟 2: 檢查缺少的欄位
-- ============================================

\echo '📌 步驟 2: 檢查必要欄位是否存在'
\echo ''

-- 列出應該存在的所有欄位
WITH required_columns AS (
    SELECT unnest(ARRAY[
        'time',
        'monitor_id',
        'status',
        'cluster_name',
        'cluster_status',
        'response_time',
        'cpu_usage',
        'memory_usage',
        'disk_usage',
        'node_count',
        'data_node_count',
        'query_latency',
        'indexing_rate',
        'search_rate',
        'total_indices',
        'total_documents',
        'total_size_bytes',
        'active_shards',
        'relocating_shards',
        'unassigned_shards',
        'error_message',
        'warning_message',
        'metadata'
    ]) AS column_name
),
existing_columns AS (
    SELECT column_name
    FROM information_schema.columns
    WHERE table_name = 'es_metrics'
)
SELECT
    r.column_name,
    CASE
        WHEN e.column_name IS NOT NULL THEN '✅ 存在'
        ELSE '❌ 缺少'
    END AS status
FROM required_columns r
LEFT JOIN existing_columns e ON r.column_name = e.column_name
ORDER BY r.column_name;

\echo ''

-- ============================================
-- 步驟 3: 安全地添加缺少的欄位
-- ============================================

\echo '📌 步驟 3: 添加缺少的欄位'
\echo ''

-- 添加 total_indices（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'total_indices'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN total_indices INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 total_indices 欄位';
    ELSE
        RAISE NOTICE '⏭️  total_indices 欄位已存在';
    END IF;
END $$;

-- 添加 total_documents（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'total_documents'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN total_documents BIGINT DEFAULT 0;
        RAISE NOTICE '✅ 已添加 total_documents 欄位';
    ELSE
        RAISE NOTICE '⏭️  total_documents 欄位已存在';
    END IF;
END $$;

-- 添加 total_size_bytes（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'total_size_bytes'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN total_size_bytes BIGINT DEFAULT 0;
        RAISE NOTICE '✅ 已添加 total_size_bytes 欄位';
    ELSE
        RAISE NOTICE '⏭️  total_size_bytes 欄位已存在';
    END IF;
END $$;

-- 添加 active_shards（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'active_shards'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN active_shards INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 active_shards 欄位';
    ELSE
        RAISE NOTICE '⏭️  active_shards 欄位已存在';
    END IF;
END $$;

-- 添加 relocating_shards（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'relocating_shards'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN relocating_shards INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 relocating_shards 欄位';
    ELSE
        RAISE NOTICE '⏭️  relocating_shards 欄位已存在';
    END IF;
END $$;

-- 添加 unassigned_shards（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'unassigned_shards'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN unassigned_shards INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 unassigned_shards 欄位';
    ELSE
        RAISE NOTICE '⏭️  unassigned_shards 欄位已存在';
    END IF;
END $$;

-- 添加 query_latency（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'query_latency'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN query_latency BIGINT DEFAULT 0;
        RAISE NOTICE '✅ 已添加 query_latency 欄位';
    ELSE
        RAISE NOTICE '⏭️  query_latency 欄位已存在';
    END IF;
END $$;

-- 添加 indexing_rate（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'indexing_rate'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN indexing_rate DECIMAL(10,2) DEFAULT 0.00;
        RAISE NOTICE '✅ 已添加 indexing_rate 欄位';
    ELSE
        RAISE NOTICE '⏭️  indexing_rate 欄位已存在';
    END IF;
END $$;

-- 添加 search_rate（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'search_rate'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN search_rate DECIMAL(10,2) DEFAULT 0.00;
        RAISE NOTICE '✅ 已添加 search_rate 欄位';
    ELSE
        RAISE NOTICE '⏭️  search_rate 欄位已存在';
    END IF;
END $$;

\echo ''

-- ============================================
-- 步驟 4: 驗證最終結構
-- ============================================

\echo '📌 步驟 4: 驗證最終表結構'
\echo ''

SELECT
    column_name,
    data_type,
    COALESCE(column_default::text, 'NULL') as default_value
FROM information_schema.columns
WHERE table_name = 'es_metrics'
ORDER BY ordinal_position;

\echo ''
\echo '--- 最終欄位數量（應該是 23 個）---'
SELECT COUNT(*) as final_column_count
FROM information_schema.columns
WHERE table_name = 'es_metrics';

\echo ''

-- ============================================
-- 步驟 5: 檢查 Hypertable 狀態
-- ============================================

\echo '📌 步驟 5: 檢查 Hypertable 狀態'
\echo ''

SELECT
    hypertable_name,
    num_dimensions,
    num_chunks,
    compression_enabled,
    CASE WHEN table_bytes IS NOT NULL THEN pg_size_pretty(table_bytes) ELSE 'N/A' END as table_size
FROM timescaledb_information.hypertables
WHERE hypertable_name = 'es_metrics';

\echo ''

-- ============================================
-- 完成
-- ============================================

\echo '🎉 es_metrics 表結構檢查與修復完成！'
\echo ''
\echo '💡 下一步:'
\echo '   1. 測試 API: GET /api/v1/elasticsearch/statistics'
\echo '   2. 如果仍有問題，檢查應用日誌'
\echo '   3. 確保 BatchWriter 正確寫入資料'
\echo ''
