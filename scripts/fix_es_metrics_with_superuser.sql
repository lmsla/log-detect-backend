-- ======================================================
-- 使用 superuser 權限修復 es_metrics 表結構
-- ======================================================
-- 使用方式 1: 使用 postgres 超級用戶
-- psql -U postgres -d monitoring -f scripts/fix_es_metrics_with_superuser.sql
--
-- 使用方式 2: 使用 sudo (如果是本地安裝)
-- sudo -u postgres psql -d monitoring -f scripts/fix_es_metrics_with_superuser.sql
-- ======================================================

\echo '=== 使用超級用戶權限修復 es_metrics 表 ==='
\echo ''

-- ============================================
-- 步驟 1: 檢查當前連接用戶
-- ============================================

\echo '📌 當前連接用戶:'
SELECT current_user, session_user;

\echo ''

-- ============================================
-- 步驟 2: 檢查表擁有者
-- ============================================

\echo '📌 檢查 es_metrics 表的擁有者:'
SELECT
    schemaname,
    tablename,
    tableowner
FROM pg_tables
WHERE tablename = 'es_metrics';

\echo ''

-- ============================================
-- 步驟 3: 添加缺少的欄位
-- ============================================

\echo '📌 開始添加缺少的欄位...'
\echo ''

-- 添加 total_indices
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'total_indices'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN total_indices INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 total_indices';
    ELSE
        RAISE NOTICE '⏭️  total_indices 已存在';
    END IF;
END $$;

-- 添加 total_documents
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'total_documents'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN total_documents BIGINT DEFAULT 0;
        RAISE NOTICE '✅ 已添加 total_documents';
    ELSE
        RAISE NOTICE '⏭️  total_documents 已存在';
    END IF;
END $$;

-- 添加 total_size_bytes
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'total_size_bytes'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN total_size_bytes BIGINT DEFAULT 0;
        RAISE NOTICE '✅ 已添加 total_size_bytes';
    ELSE
        RAISE NOTICE '⏭️  total_size_bytes 已存在';
    END IF;
END $$;

-- 添加 active_shards
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'active_shards'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN active_shards INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 active_shards';
    ELSE
        RAISE NOTICE '⏭️  active_shards 已存在';
    END IF;
END $$;

-- 添加 relocating_shards
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'relocating_shards'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN relocating_shards INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 relocating_shards';
    ELSE
        RAISE NOTICE '⏭️  relocating_shards 已存在';
    END IF;
END $$;

-- 添加 unassigned_shards
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'unassigned_shards'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN unassigned_shards INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 unassigned_shards';
    ELSE
        RAISE NOTICE '⏭️  unassigned_shards 已存在';
    END IF;
END $$;

-- 添加 query_latency
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'query_latency'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN query_latency BIGINT DEFAULT 0;
        RAISE NOTICE '✅ 已添加 query_latency';
    ELSE
        RAISE NOTICE '⏭️  query_latency 已存在';
    END IF;
END $$;

-- 添加 indexing_rate
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'indexing_rate'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN indexing_rate DECIMAL(10,2) DEFAULT 0.00;
        RAISE NOTICE '✅ 已添加 indexing_rate';
    ELSE
        RAISE NOTICE '⏭️  indexing_rate 已存在';
    END IF;
END $$;

-- 添加 search_rate
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'es_metrics' AND column_name = 'search_rate'
    ) THEN
        ALTER TABLE es_metrics ADD COLUMN search_rate DECIMAL(10,2) DEFAULT 0.00;
        RAISE NOTICE '✅ 已添加 search_rate';
    ELSE
        RAISE NOTICE '⏭️  search_rate 已存在';
    END IF;
END $$;

\echo ''

-- ============================================
-- 步驟 4: 授予 logdetect 用戶權限
-- ============================================

\echo '📌 授予 logdetect 用戶權限...'

-- 授予表的所有權限
GRANT ALL PRIVILEGES ON TABLE es_metrics TO logdetect;
GRANT ALL PRIVILEGES ON TABLE es_alert_history TO logdetect;

-- 授予 schema 中所有表的權限（為未來的表做準備）
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO logdetect;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO logdetect;

\echo '✅ 權限已授予'
\echo ''

-- ============================================
-- 步驟 5: 驗證結果
-- ============================================

\echo '📌 驗證表結構:'

SELECT COUNT(*) as total_columns
FROM information_schema.columns
WHERE table_name = 'es_metrics';

\echo ''
\echo '📌 檢查關鍵欄位:'

SELECT
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'es_metrics' AND column_name = 'total_indices') THEN '✅ total_indices' ELSE '❌ total_indices' END,
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'es_metrics' AND column_name = 'total_documents') THEN '✅ total_documents' ELSE '❌ total_documents' END,
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'es_metrics' AND column_name = 'total_size_bytes') THEN '✅ total_size_bytes' ELSE '❌ total_size_bytes' END,
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'es_metrics' AND column_name = 'active_shards') THEN '✅ active_shards' ELSE '❌ active_shards' END;

\echo ''
\echo '🎉 修復完成！'
\echo ''
\echo '💡 驗證步驟:'
\echo '   1. 使用 logdetect 用戶測試連接:'
\echo '      psql -U logdetect -d monitoring -c "SELECT COUNT(*) FROM es_metrics;"'
\echo ''
\echo '   2. 測試 API:'
\echo '      curl http://localhost:8006/api/v1/elasticsearch/statistics'
\echo ''
