# 🔧 修復：PostgreSQL 權限錯誤

## 錯誤訊息

```
ERROR:  must be owner of table es_metrics
SQL state: 42501
```

## 問題原因

當前用戶（`logdetect`）不是 `es_metrics` 表的擁有者，無法執行 `ALTER TABLE` 操作。

---

## 🚀 快速修復（3 種方法）

### 方法 1: 使用 postgres 超級用戶執行腳本（推薦）

```bash
cd /Users/chen/Downloads/01BiMap/03MyDevs/log-detect/log-detect-backend

# 使用 postgres 超級用戶執行
psql -U postgres -d monitoring -f scripts/fix_es_metrics_with_superuser.sql
```

**如果提示輸入密碼**:
- 預設密碼通常是 `postgres` 或你安裝時設定的密碼
- 如果忘記密碼，參考下方「重置密碼」章節

---

### 方法 2: 使用 sudo 執行（本地開發環境）

如果 PostgreSQL 是本地安裝且使用 peer authentication：

```bash
# 方式 A: 切換到 postgres 用戶
sudo -u postgres psql -d monitoring -f scripts/fix_es_metrics_with_superuser.sql

# 方式 B: 直接執行
sudo -u postgres psql monitoring << 'EOF'
-- 添加欄位
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS total_indices INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS total_documents BIGINT DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS total_size_bytes BIGINT DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS active_shards INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS relocating_shards INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS unassigned_shards INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS query_latency BIGINT DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS indexing_rate DECIMAL(10,2) DEFAULT 0.00;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS search_rate DECIMAL(10,2) DEFAULT 0.00;

-- 授予權限
GRANT ALL PRIVILEGES ON TABLE es_metrics TO logdetect;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO logdetect;

-- 驗證
SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'es_metrics';
EOF
```

---

### 方法 3: 手動步驟（逐步執行）

#### 步驟 1: 連接為超級用戶

```bash
# 連接到資料庫
psql -U postgres -d monitoring

# 或使用 sudo
sudo -u postgres psql monitoring
```

#### 步驟 2: 檢查表擁有者

```sql
-- 查看表擁有者
SELECT schemaname, tablename, tableowner
FROM pg_tables
WHERE tablename = 'es_metrics';
```

**可能的結果**:
- `tableowner = postgres` → 表由 postgres 創建
- `tableowner = logdetect` → 表由 logdetect 創建（不應該有權限問題）

#### 步驟 3: 添加欄位

```sql
-- 添加所有缺少的欄位
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS total_indices INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS total_documents BIGINT DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS total_size_bytes BIGINT DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS active_shards INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS relocating_shards INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS unassigned_shards INTEGER DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS query_latency BIGINT DEFAULT 0;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS indexing_rate DECIMAL(10,2) DEFAULT 0.00;
ALTER TABLE es_metrics ADD COLUMN IF NOT EXISTS search_rate DECIMAL(10,2) DEFAULT 0.00;
```

#### 步驟 4: 授予權限

```sql
-- 授予 logdetect 用戶完整權限
GRANT ALL PRIVILEGES ON TABLE es_metrics TO logdetect;
GRANT ALL PRIVILEGES ON TABLE es_alert_history TO logdetect;

-- 為未來的表授權
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO logdetect;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO logdetect;

-- 設置預設權限（新建的表自動授權）
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL PRIVILEGES ON TABLES TO logdetect;
```

#### 步驟 5: 驗證

```sql
-- 檢查欄位數量（應該是 23）
SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'es_metrics';

-- 檢查權限
SELECT grantee, privilege_type
FROM information_schema.role_table_grants
WHERE table_name = 'es_metrics' AND grantee = 'logdetect';

-- 退出
\q
```

#### 步驟 6: 測試 logdetect 用戶

```bash
# 使用 logdetect 用戶連接
psql -U logdetect -d monitoring

# 測試查詢
SELECT COUNT(*) FROM es_metrics;

# 測試寫入（應該成功）
-- 會由應用程式自動寫入
```

---

## 🔐 如果忘記 postgres 密碼

### macOS (Homebrew 安裝)

```bash
# 停止 PostgreSQL
brew services stop postgresql

# 編輯配置（臨時禁用密碼）
code /opt/homebrew/var/postgresql@14/pg_hba.conf
# 或
nano /opt/homebrew/var/postgresql@14/pg_hba.conf

# 將所有 md5 改為 trust
# 例如: local   all   all   md5 → local   all   all   trust

# 重啟 PostgreSQL
brew services start postgresql

# 連接並重設密碼
psql -U postgres
ALTER USER postgres PASSWORD 'new_password';
\q

# 恢復配置（改回 md5）
# 重啟 PostgreSQL
brew services restart postgresql
```

### Linux (Ubuntu/Debian)

```bash
# 切換到 postgres 用戶
sudo -u postgres psql

# 重設密碼
ALTER USER postgres PASSWORD 'new_password';
\q
```

### Docker

```bash
# 進入容器
docker exec -it timescaledb psql -U postgres

# 重設密碼
ALTER USER postgres PASSWORD 'new_password';
\q
```

---

## 🔍 故障排查

### 問題 1: 找不到 postgres 用戶

```bash
# 檢查 PostgreSQL 用戶列表
psql -U postgres -c "\du"

# 或
sudo -u postgres psql -c "\du"
```

### 問題 2: 仍然無法連接

檢查 `pg_hba.conf` 配置：

```bash
# 查找配置文件位置
psql -U postgres -c "SHOW hba_file;"

# 或
sudo -u postgres psql -c "SHOW hba_file;"

# 編輯配置
# macOS Homebrew: /opt/homebrew/var/postgresql@14/pg_hba.conf
# Linux: /etc/postgresql/14/main/pg_hba.conf
```

確保有類似這樣的配置：

```
# TYPE  DATABASE        USER            ADDRESS                 METHOD
local   all             postgres                                peer
local   all             all                                     md5
host    all             all             127.0.0.1/32            md5
host    all             all             ::1/128                 md5
```

### 問題 3: 權限仍然不足

```sql
-- 以 postgres 用戶執行
-- 1. 直接更改表擁有者
ALTER TABLE es_metrics OWNER TO logdetect;
ALTER TABLE es_alert_history OWNER TO logdetect;

-- 2. 或者授予 logdetect 超級用戶權限（不推薦用於生產環境）
ALTER USER logdetect WITH SUPERUSER;

-- 3. 檢查當前權限
\dp es_metrics
```

---

## 📋 驗證清單

完成修復後，請確認：

- [ ] 可以使用 postgres 用戶連接資料庫
  ```bash
  psql -U postgres -d monitoring -c "SELECT version();"
  ```

- [ ] es_metrics 表有 23 個欄位
  ```bash
  psql -U postgres -d monitoring -c "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'es_metrics';"
  ```

- [ ] logdetect 用戶有完整權限
  ```bash
  psql -U logdetect -d monitoring -c "SELECT COUNT(*) FROM es_metrics;"
  ```

- [ ] API 請求成功
  ```bash
  curl http://localhost:8006/api/v1/elasticsearch/statistics \
    -H "Authorization: Bearer YOUR_TOKEN"
  ```

---

## 💡 最佳實踐

### 1. 使用一致的用戶創建表

確保所有表都由 `logdetect` 用戶創建，或者都由 `postgres` 創建並授權。

### 2. 初始化腳本使用正確用戶

修改 `postgresql_install.sh`，使用 `logdetect` 用戶執行：

```bash
# 方式 1: 直接指定用戶
psql -U logdetect -d monitoring -f postgresql_install.sh

# 方式 2: 在腳本開頭添加
-- SET ROLE logdetect;
```

### 3. 設置預設權限

在初始化時設置預設權限：

```sql
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL PRIVILEGES ON TABLES TO logdetect;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL PRIVILEGES ON SEQUENCES TO logdetect;
```

---

## 🆘 仍然有問題？

提供以下資訊：

```bash
# 1. PostgreSQL 版本
psql --version

# 2. 當前用戶和權限
psql -U postgres -d monitoring -c "
  SELECT
    current_user,
    session_user,
    (SELECT tableowner FROM pg_tables WHERE tablename = 'es_metrics') as table_owner;
"

# 3. 完整錯誤訊息
# 包含 SQL state 和 context
```

---

**更新日期**: 2025-10-07
**相關腳本**: `scripts/fix_es_metrics_with_superuser.sql`
