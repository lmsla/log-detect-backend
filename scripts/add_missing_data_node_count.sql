-- ======================================================
-- 添加缺少的 data_node_count 欄位
-- ======================================================
-- 使用方式：
-- psql -U logdetect -d monitoring -f scripts/add_missing_data_node_count.sql
-- 或使用 superuser:
-- psql -U postgres -d monitoring -f scripts/add_missing_data_node_count.sql
-- ======================================================

\echo '=== 添加缺少的 data_node_count 欄位 ==='
\echo ''

-- 檢查當前表結構
\echo '📌 當前 es_metrics 欄位列表:'
SELECT column_name, data_type, column_default
FROM information_schema.columns
WHERE table_name = 'es_metrics'
ORDER BY ordinal_position;

\echo ''

-- 添加 data_node_count 欄位
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'data_node_count'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN data_node_count INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 data_node_count 欄位';
    ELSE
        RAISE NOTICE '⏭️  data_node_count 欄位已存在';
    END IF;
END $$;

\echo ''
\echo '📌 驗證：檢查 data_node_count 欄位'
SELECT column_name, data_type, column_default
FROM information_schema.columns
WHERE table_name = 'es_metrics' AND column_name = 'data_node_count';

\echo ''
\echo '📌 總欄位數（應該是 23）:'
SELECT COUNT(*) as total_columns
FROM information_schema.columns
WHERE table_name = 'es_metrics';

\echo ''
\echo '🎉 完成！請重啟應用程式以重新初始化 prepared statement'
