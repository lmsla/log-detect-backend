package services

import (
	"database/sql"
	"fmt"
	"log"
	"log-detect/global"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations 執行資料庫 migrations
// 在程式啟動時自動呼叫，檢查並執行未執行過的 migration
func RunMigrations() error {
	log.Println("Starting database migrations...")

	// 執行 MySQL migrations
	if err := runMySQLMigrations(); err != nil {
		return fmt.Errorf("MySQL migration failed: %w", err)
	}

	// 執行 TimescaleDB migrations
	if err := runTimescaleDBMigrations(); err != nil {
		return fmt.Errorf("TimescaleDB migration failed: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// runMySQLMigrations 執行 MySQL migrations
func runMySQLMigrations() error {
	db, err := global.Mysql.DB()
	if err != nil {
		return fmt.Errorf("failed to get MySQL connection: %w", err)
	}

	return executeMigrations(db, "migrations/mysql", "mysql")
}

// runTimescaleDBMigrations 執行 TimescaleDB migrations
func runTimescaleDBMigrations() error {
	if global.TimescaleDB == nil {
		log.Println("TimescaleDB not configured, skipping migrations")
		return nil
	}

	return executeMigrations(global.TimescaleDB, "migrations/timescaledb", "timescaledb")
}

// executeMigrations 執行指定目錄的 migrations
func executeMigrations(db *sql.DB, migrationsDir string, dbName string) error {
	// 確保 schema_migrations 表存在
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 取得已執行的 migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 掃描 migration 檔案
	files, err := scanMigrationFiles(migrationsDir)
	if err != nil {
		// 如果目錄不存在，跳過
		if os.IsNotExist(err) {
			log.Printf("Migration directory %s not found, skipping", migrationsDir)
			return nil
		}
		return fmt.Errorf("failed to scan migration files: %w", err)
	}

	// 執行未執行的 migrations
	for _, file := range files {
		version := extractVersion(file)
		if _, ok := applied[version]; ok {
			continue // 已執行過，跳過
		}

		log.Printf("[%s] Applying migration: %s", dbName, file)

		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// 執行 SQL
		if err := executeSQL(db, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		// 記錄已執行
		if err := recordMigration(db, version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		log.Printf("[%s] Applied migration: %s", dbName, file)
	}

	return nil
}

// ensureMigrationsTable 確保 schema_migrations 表存在
func ensureMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations 取得已執行的 migrations
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// scanMigrationFiles 掃描 migration 檔案（只取 .up.sql）
func scanMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, name)
		}
	}

	// 按檔名排序（確保執行順序）
	sort.Strings(files)

	return files, nil
}

// extractVersion 從檔名提取版本號
// 例如 "001_initial_schema.up.sql" -> "001_initial_schema"
func extractVersion(filename string) string {
	return strings.TrimSuffix(filename, ".up.sql")
}

// executeSQL 執行 SQL（支援多條語句）
func executeSQL(db *sql.DB, sqlContent string) error {
	// 分割 SQL 語句（以 ; 分割，但要處理字串中的 ;）
	statements := splitSQL(sqlContent)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("SQL error: %w\nStatement: %s", err, truncateSQL(stmt))
		}
	}

	return nil
}

// splitSQL 分割 SQL 語句
func splitSQL(content string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	stringChar := rune(0)

	for _, ch := range content {
		if inString {
			current.WriteRune(ch)
			if ch == stringChar {
				inString = false
			}
		} else {
			if ch == '\'' || ch == '"' {
				inString = true
				stringChar = ch
				current.WriteRune(ch)
			} else if ch == ';' {
				statements = append(statements, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		}
	}

	// 處理最後一條語句（可能沒有 ;）
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}

// recordMigration 記錄已執行的 migration
func recordMigration(db *sql.DB, version string) error {
	_, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
	return err
}

// truncateSQL 截斷 SQL 用於錯誤訊息
func truncateSQL(sql string) string {
	if len(sql) > 200 {
		return sql[:200] + "..."
	}
	return sql
}
