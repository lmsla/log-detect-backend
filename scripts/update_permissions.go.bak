package main

import (
	"fmt"
	"log"
	"log-detect/global"
	"log-detect/services"
	"os"
)

// UpdatePermissions 更新權限設定的獨立腳本
// 使用方式: go run scripts/update_permissions.go
func main() {
	fmt.Println("=== Log Detect 權限更新腳本 ===")
	fmt.Println()

	// 1. 初始化全局變量（模擬應用啟動）
	fmt.Println("📌 步驟 1: 連接資料庫...")
	if err := initDatabase(); err != nil {
		log.Fatalf("❌ 資料庫連接失敗: %v", err)
	}
	fmt.Println("✅ 資料庫連接成功")
	fmt.Println()

	// 2. 更新權限
	fmt.Println("📌 步驟 2: 更新權限定義...")
	authService := services.NewAuthService()
	if err := authService.CreateDefaultRolesAndPermissions(); err != nil {
		log.Fatalf("❌ 權限更新失敗: %v", err)
	}
	fmt.Println("✅ 權限定義已更新")
	fmt.Println()

	// 3. 驗證 elasticsearch 權限
	fmt.Println("📌 步驟 3: 驗證 elasticsearch 權限...")
	if err := verifyElasticsearchPermissions(); err != nil {
		log.Fatalf("❌ 驗證失敗: %v", err)
	}
	fmt.Println("✅ elasticsearch 權限驗證通過")
	fmt.Println()

	// 4. 顯示權限摘要
	fmt.Println("📌 步驟 4: 顯示權限摘要...")
	showPermissionSummary()
	fmt.Println()

	fmt.Println("🎉 權限更新完成！")
	fmt.Println()
	fmt.Println("💡 建議:")
	fmt.Println("   1. 重新登入以獲取新的 token")
	fmt.Println("   2. 測試 elasticsearch API 端點")
	fmt.Println("   3. 如果仍有問題，檢查 user 的 role_id 是否正確")
}

// initDatabase 初始化資料庫連接
func initDatabase() error {
	// 這裡需要根據實際項目的初始化方式調整
	// 可以參考 main.go 中的資料庫初始化代碼

	// 如果使用環境變量
	dbHost := os.Getenv("MYSQL_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	fmt.Printf("   連接到: %s\n", dbHost)

	// TODO: 調用實際的資料庫初始化函數
	// 例如: return database.InitMySQL()

	return fmt.Errorf("請在此實作資料庫初始化邏輯")
}

// verifyElasticsearchPermissions 驗證 elasticsearch 權限
func verifyElasticsearchPermissions() error {
	var count int64

	// 檢查 elasticsearch 權限是否存在
	result := global.Mysql.Table("permissions").
		Where("resource = ?", "elasticsearch").
		Count(&count)

	if result.Error != nil {
		return result.Error
	}

	if count != 4 {
		return fmt.Errorf("elasticsearch 權限數量不正確，預期 4 個，實際 %d 個", count)
	}

	fmt.Printf("   找到 %d 個 elasticsearch 權限\n", count)

	// 檢查 admin 角色是否有 elasticsearch 權限
	var adminPermCount int64
	result = global.Mysql.Table("role_permissions").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Where("roles.name = ? AND permissions.resource = ?", "admin", "elasticsearch").
		Count(&adminPermCount)

	if result.Error != nil {
		return result.Error
	}

	if adminPermCount != 4 {
		return fmt.Errorf("admin 角色的 elasticsearch 權限數量不正確，預期 4 個，實際 %d 個", adminPermCount)
	}

	fmt.Printf("   admin 角色已分配 %d 個 elasticsearch 權限\n", adminPermCount)

	return nil
}

// showPermissionSummary 顯示權限摘要
func showPermissionSummary() error {
	type PermissionSummary struct {
		Resource string
		Count    int64
	}

	var summaries []PermissionSummary

	result := global.Mysql.Table("permissions").
		Select("resource, COUNT(*) as count").
		Group("resource").
		Order("resource").
		Scan(&summaries)

	if result.Error != nil {
		return result.Error
	}

	fmt.Println("   資源權限統計:")
	fmt.Println("   ┌─────────────────┬────────┐")
	fmt.Println("   │ Resource        │ Count  │")
	fmt.Println("   ├─────────────────┼────────┤")

	for _, summary := range summaries {
		fmt.Printf("   │ %-15s │ %6d │\n", summary.Resource, summary.Count)
	}

	fmt.Println("   └─────────────────┴────────┘")

	return nil
}
