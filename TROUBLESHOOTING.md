# 🔧 故障排除指南

## 常見問題

### 1. 表不存在錯誤

**錯誤信息：**
```
Error 1146 (42S02): Table 'logdetect.permissions' doesn't exist
Error 1146 (42S02): Table 'logdetect.roles' doesn't exist
```

**原因：**
應用啟動順序問題 - 認證系統初始化在表創建之前。

**解決方案：**

#### 自動解決 (推薦)
重新啟動應用，表會自動創建：
```bash
go run main.go
```

#### 手動解決
如果自動創建失敗，可以手動創建表：
```bash
go run create_tables.go
```

### 2. 數據庫連接失敗

**錯誤信息：**
```
SQL Database 連線失敗
```

**檢查項目：**
1. MySQL 服務是否運行
2. 數據庫連接配置是否正確 (`setting.yml`)
3. 數據庫用戶權限是否正確

**解決方案：**
```yaml
# 檢查 setting.yml
database:
  host: "10.99.1.133"  # 確保 IP 正確
  port: "3306"
  user: "runner"
  password: "1qaz2wsx"
  name: "logdetect"
```

### 3. JWT 認證失敗

**錯誤信息：**
```
Invalid or expired token
```

**檢查項目：**
1. JWT_SECRET 環境變數是否設置
2. Token 是否過期 (默認 24 小時)
3. Token 格式是否正確

**解決方案：**
```bash
# 設置環境變數
export JWT_SECRET="your-super-secret-key"

# 或者重新登錄獲取新 token
curl -X POST http://localhost:8006/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 4. 權限不足錯誤

**錯誤信息：**
```
Insufficient permissions
```

**檢查項目：**
1. 用戶角色是否正確
2. 角色權限是否正確配置
3. API 端點權限要求是否合理

**解決方案：**
檢查用戶角色和權限：
```bash
# 獲取用戶信息
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8006/api/v1/auth/profile
```

### 5. Elasticsearch 連接失敗

**錯誤信息：**
```
ES Cluster 連線失敗
```

**檢查項目：**
1. Elasticsearch 服務是否運行
2. 連接配置是否正確
3. TLS 配置是否正確

## 調試命令

### 檢查數據庫表
```sql
-- 連接到 MySQL
mysql -h 10.99.1.133 -u runner -p logdetect

-- 查看所有表
SHOW TABLES;

-- 查看用戶表結構
DESCRIBE users;

-- 查看角色表結構
DESCRIBE roles;

-- 查看權限表結構
DESCRIBE permissions;
```

### 檢查應用日誌
```bash
# 運行應用並查看日誌
go run main.go

# 或者運行測試腳本
./test_auth.sh
```

### 檢查環境變數
```bash
# 檢查 JWT 密鑰
echo $JWT_SECRET

# 如果未設置，使用默認值 (不安全)
export JWT_SECRET="your-production-secret-key"
```

## 快速修復腳本

創建 `fix_database.sh` 腳本：

```bash
#!/bin/bash
echo "🔧 Fixing database issues..."

# 停止現有應用
pkill -f "go run main.go" || true

# 創建表
go run create_tables.go

# 重新啟動應用
echo "🚀 Starting application..."
go run main.go
```

運行修復腳本：
```bash
chmod +x fix_database.sh
./fix_database.sh
```

## 聯絡支持

如果以上方法都無法解決問題，請提供：
1. 完整的錯誤信息
2. 應用啟動日誌
3. 數據庫狀態信息
4. 環境配置信息
