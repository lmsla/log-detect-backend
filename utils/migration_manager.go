package utils

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationManager 管理資料庫 migration
type MigrationManager struct {
	mysqlMigrate      *migrate.Migrate
	timescaleDBMigrate *migrate.Migrate
}

// NewMigrationManager 建立 Migration 管理器
func NewMigrationManager(mysqlDB *sql.DB, timescaleDB *sql.DB, migrationPath string) (*MigrationManager, error) {
	manager := &MigrationManager{}

	// 設定 MySQL migrations
	if mysqlDB != nil {
		mysqlDriver, err := mysql.WithInstance(mysqlDB, &mysql.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create MySQL driver: %w", err)
		}

		mysqlMigrate, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s/mysql", migrationPath),
			"mysql",
			mysqlDriver,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create MySQL migration instance: %w", err)
		}

		manager.mysqlMigrate = mysqlMigrate
		log.Println("✅ MySQL migration manager initialized")
	}

	// 設定 TimescaleDB migrations (使用 PostgreSQL driver)
	if timescaleDB != nil {
		pgDriver, err := postgres.WithInstance(timescaleDB, &postgres.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL driver: %w", err)
		}

		timescaleDBMigrate, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s/timescaledb", migrationPath),
			"postgres",
			pgDriver,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create TimescaleDB migration instance: %w", err)
		}

		manager.timescaleDBMigrate = timescaleDBMigrate
		log.Println("✅ TimescaleDB migration manager initialized")
	}

	return manager, nil
}

// MigrateUp 執行所有 pending migrations (up)
func (m *MigrationManager) MigrateUp() error {
	var errors []error

	// Migrate MySQL
	if m.mysqlMigrate != nil {
		log.Println("📊 Running MySQL migrations...")
		if err := m.mysqlMigrate.Up(); err != nil && err != migrate.ErrNoChange {
			errors = append(errors, fmt.Errorf("MySQL migration failed: %w", err))
		} else if err == migrate.ErrNoChange {
			log.Println("✅ MySQL schema is up to date")
		} else {
			log.Println("✅ MySQL migrations completed successfully")
		}
	}

	// Migrate TimescaleDB
	if m.timescaleDBMigrate != nil {
		log.Println("📊 Running TimescaleDB migrations...")
		if err := m.timescaleDBMigrate.Up(); err != nil && err != migrate.ErrNoChange {
			errors = append(errors, fmt.Errorf("TimescaleDB migration failed: %w", err))
		} else if err == migrate.ErrNoChange {
			log.Println("✅ TimescaleDB schema is up to date")
		} else {
			log.Println("✅ TimescaleDB migrations completed successfully")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("migration errors: %v", errors)
	}

	return nil
}

// MigrateDown 回滾一個 migration
func (m *MigrationManager) MigrateDown() error {
	var errors []error

	// Rollback MySQL
	if m.mysqlMigrate != nil {
		log.Println("⏪ Rolling back MySQL migration...")
		if err := m.mysqlMigrate.Steps(-1); err != nil && err != migrate.ErrNoChange {
			errors = append(errors, fmt.Errorf("MySQL rollback failed: %w", err))
		} else {
			log.Println("✅ MySQL migration rolled back successfully")
		}
	}

	// Rollback TimescaleDB
	if m.timescaleDBMigrate != nil {
		log.Println("⏪ Rolling back TimescaleDB migration...")
		if err := m.timescaleDBMigrate.Steps(-1); err != nil && err != migrate.ErrNoChange {
			errors = append(errors, fmt.Errorf("TimescaleDB rollback failed: %w", err))
		} else {
			log.Println("✅ TimescaleDB migration rolled back successfully")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("rollback errors: %v", errors)
	}

	return nil
}

// MigrateTo 遷移到特定版本
func (m *MigrationManager) MigrateTo(version uint) error {
	var errors []error

	// MySQL
	if m.mysqlMigrate != nil {
		log.Printf("📊 Migrating MySQL to version %d...\n", version)
		if err := m.mysqlMigrate.Migrate(version); err != nil && err != migrate.ErrNoChange {
			errors = append(errors, fmt.Errorf("MySQL migration to version %d failed: %w", version, err))
		} else {
			log.Printf("✅ MySQL migrated to version %d\n", version)
		}
	}

	// TimescaleDB
	if m.timescaleDBMigrate != nil {
		log.Printf("📊 Migrating TimescaleDB to version %d...\n", version)
		if err := m.timescaleDBMigrate.Migrate(version); err != nil && err != migrate.ErrNoChange {
			errors = append(errors, fmt.Errorf("TimescaleDB migration to version %d failed: %w", version, err))
		} else {
			log.Printf("✅ TimescaleDB migrated to version %d\n", version)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("migration to version %d errors: %v", version, errors)
	}

	return nil
}

// GetVersion 取得目前的 migration 版本
func (m *MigrationManager) GetVersion() (mysqlVersion uint, mysqlDirty bool, timescaleVersion uint, timescaleDirty bool, err error) {
	// MySQL version
	if m.mysqlMigrate != nil {
		mysqlVersion, mysqlDirty, err = m.mysqlMigrate.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return 0, false, 0, false, fmt.Errorf("failed to get MySQL version: %w", err)
		}
	}

	// TimescaleDB version
	if m.timescaleDBMigrate != nil {
		timescaleVersion, timescaleDirty, err = m.timescaleDBMigrate.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return 0, false, 0, false, fmt.Errorf("failed to get TimescaleDB version: %w", err)
		}
	}

	return mysqlVersion, mysqlDirty, timescaleVersion, timescaleDirty, nil
}

// Force 強制設定版本（用於修復 dirty 狀態）
func (m *MigrationManager) Force(version int) error {
	var errors []error

	// MySQL
	if m.mysqlMigrate != nil {
		log.Printf("⚠️  Forcing MySQL to version %d...\n", version)
		if err := m.mysqlMigrate.Force(version); err != nil {
			errors = append(errors, fmt.Errorf("MySQL force failed: %w", err))
		} else {
			log.Println("✅ MySQL version forced successfully")
		}
	}

	// TimescaleDB
	if m.timescaleDBMigrate != nil {
		log.Printf("⚠️  Forcing TimescaleDB to version %d...\n", version)
		if err := m.timescaleDBMigrate.Force(version); err != nil {
			errors = append(errors, fmt.Errorf("TimescaleDB force failed: %w", err))
		} else {
			log.Println("✅ TimescaleDB version forced successfully")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("force errors: %v", errors)
	}

	return nil
}

// Close 關閉 migration 實例
func (m *MigrationManager) Close() error {
	var errors []error

	if m.mysqlMigrate != nil {
		if srcErr, dbErr := m.mysqlMigrate.Close(); srcErr != nil || dbErr != nil {
			errors = append(errors, fmt.Errorf("MySQL close error - src: %v, db: %v", srcErr, dbErr))
		}
	}

	if m.timescaleDBMigrate != nil {
		if srcErr, dbErr := m.timescaleDBMigrate.Close(); srcErr != nil || dbErr != nil {
			errors = append(errors, fmt.Errorf("TimescaleDB close error - src: %v, db: %v", srcErr, dbErr))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("close errors: %v", errors)
	}

	return nil
}
