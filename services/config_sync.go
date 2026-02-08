package services

import (
	"fmt"
	"log"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/structs"

	"gorm.io/gorm"
)

// SyncConfigToDB 將 config.yml 的擴充配置同步至資料庫
// 策略：以 YML 為唯一來源，全量覆蓋（upsert by business key）
func SyncConfigToDB() error {
	cfg := global.YMLConfig
	if cfg == nil {
		return fmt.Errorf("YMLConfig is nil, cannot sync")
	}

	tx := global.Mysql.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic during config sync, rolled back: %v", r)
		}
	}()

	// Step 1: 同步 ES Connections
	if err := syncESConnections(tx, cfg); err != nil {
		tx.Rollback()
		return fmt.Errorf("sync ES connections failed: %w", err)
	}

	// Step 2: 同步 Targets + Indices
	if err := syncTargets(tx, cfg); err != nil {
		tx.Rollback()
		return fmt.Errorf("sync targets failed: %w", err)
	}

	// Step 3: 同步 Devices
	if err := syncDevices(tx, cfg); err != nil {
		tx.Rollback()
		return fmt.Errorf("sync devices failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit config sync: %w", err)
	}

	log.Println("Config sync completed successfully")
	return nil
}

// syncESConnections 同步 ES 連線配置（以 name 為識別鍵）
func syncESConnections(tx *gorm.DB, cfg *structs.YMLConfig) error {
	if len(cfg.ESConnections) == 0 {
		log.Println("No ES connections in config.yml, skipping")
		return nil
	}

	for _, ymlConn := range cfg.ESConnections {
		var existing entities.ESConnection
		result := tx.Where("name = ?", ymlConn.Name).First(&existing)

		conn := entities.ESConnection{
			Name:        ymlConn.Name,
			Host:        ymlConn.Host,
			Port:        ymlConn.Port,
			Username:    ymlConn.Username,
			Password:    ymlConn.Password,
			EnableAuth:  ymlConn.EnableAuth,
			UseTLS:      ymlConn.UseTLS,
			IsDefault:   ymlConn.IsDefault,
			Description: ymlConn.Description,
		}

		if result.Error == gorm.ErrRecordNotFound {
			// 新增
			if err := tx.Create(&conn).Error; err != nil {
				return fmt.Errorf("create ES connection '%s' failed: %w", ymlConn.Name, err)
			}
			log.Printf("Created ES connection: %s", ymlConn.Name)
		} else if result.Error == nil {
			// 更新
			conn.ID = existing.ID
			if err := tx.Model(&existing).Updates(conn).Error; err != nil {
				return fmt.Errorf("update ES connection '%s' failed: %w", ymlConn.Name, err)
			}
			log.Printf("Updated ES connection: %s", ymlConn.Name)
		} else {
			return fmt.Errorf("query ES connection '%s' failed: %w", ymlConn.Name, result.Error)
		}
	}

	return nil
}

// syncTargets 同步監控目標與索引（以 subject 為識別鍵，indices 以 logname 為識別鍵）
func syncTargets(tx *gorm.DB, cfg *structs.YMLConfig) error {
	if len(cfg.Targets) == 0 {
		log.Println("No targets in config.yml, skipping")
		return nil
	}

	for _, ymlTarget := range cfg.Targets {
		var existing entities.Target
		result := tx.Preload("Indices").Where("subject = ?", ymlTarget.Subject).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			// 建立新 Target
			target := entities.Target{
				Subject: ymlTarget.Subject,
				To:      entities.To(ymlTarget.Receiver),
				Enable:  ymlTarget.Enable,
			}

			if err := tx.Create(&target).Error; err != nil {
				return fmt.Errorf("create target '%s' failed: %w", ymlTarget.Subject, err)
			}

			// 建立 Indices 並關聯
			if err := syncIndicesForTarget(tx, &target, ymlTarget.Indices); err != nil {
				return err
			}

			log.Printf("Created target: %s with %d indices", ymlTarget.Subject, len(ymlTarget.Indices))
		} else if result.Error == nil {
			// 更新現有 Target
			if err := tx.Model(&existing).Updates(map[string]interface{}{
				"to":     entities.To(ymlTarget.Receiver),
				"enable": ymlTarget.Enable,
			}).Error; err != nil {
				return fmt.Errorf("update target '%s' failed: %w", ymlTarget.Subject, err)
			}

			// 同步 Indices
			if err := syncIndicesForTarget(tx, &existing, ymlTarget.Indices); err != nil {
				return err
			}

			log.Printf("Updated target: %s", ymlTarget.Subject)
		} else {
			return fmt.Errorf("query target '%s' failed: %w", ymlTarget.Subject, result.Error)
		}
	}

	return nil
}

// syncIndicesForTarget 同步某個 Target 下的 Indices（以 logname 為識別鍵）
func syncIndicesForTarget(tx *gorm.DB, target *entities.Target, ymlIndices []structs.YMLIndex) error {
	for _, ymlIdx := range ymlIndices {
		// 查詢 ES Connection ID（如果指定）
		var esConnID *int
		if ymlIdx.ESConnection != "" {
			var esConn entities.ESConnection
			if err := tx.Where("name = ?", ymlIdx.ESConnection).First(&esConn).Error; err != nil {
				log.Printf("Warning: ES connection '%s' not found for index '%s', skipping es_connection_id",
					ymlIdx.ESConnection, ymlIdx.Logname)
			} else {
				esConnID = &esConn.ID
			}
		}

		var existingIndex entities.Index
		result := tx.Where("logname = ?", ymlIdx.Logname).First(&existingIndex)

		idx := entities.Index{
			Pattern:        ymlIdx.Index,
			Logname:        ymlIdx.Logname,
			DeviceGroup:    ymlIdx.DeviceGroup,
			Period:         ymlIdx.Period,
			Unit:           ymlIdx.Unit,
			Field:          ymlIdx.Field,
			ESConnectionID: esConnID,
		}

		if result.Error == gorm.ErrRecordNotFound {
			// 新增 Index
			if err := tx.Create(&idx).Error; err != nil {
				return fmt.Errorf("create index '%s' failed: %w", ymlIdx.Logname, err)
			}
			// 建立 many2many 關聯
			if err := tx.Exec("INSERT IGNORE INTO indices_targets (target_id, index_id) VALUES (?, ?)",
				target.ID, idx.ID).Error; err != nil {
				return fmt.Errorf("link index '%s' to target '%s' failed: %w", ymlIdx.Logname, target.Subject, err)
			}
		} else if result.Error == nil {
			// 更新 Index
			if err := tx.Model(&existingIndex).Updates(map[string]interface{}{
				"pattern":          idx.Pattern,
				"device_group":     idx.DeviceGroup,
				"period":           idx.Period,
				"unit":             idx.Unit,
				"field":            idx.Field,
				"es_connection_id": idx.ESConnectionID,
			}).Error; err != nil {
				return fmt.Errorf("update index '%s' failed: %w", ymlIdx.Logname, err)
			}
			// 確保 many2many 關聯存在
			if err := tx.Exec("INSERT IGNORE INTO indices_targets (target_id, index_id) VALUES (?, ?)",
				target.ID, existingIndex.ID).Error; err != nil {
				return fmt.Errorf("link index '%s' to target '%s' failed: %w", ymlIdx.Logname, target.Subject, err)
			}
		} else {
			return fmt.Errorf("query index '%s' failed: %w", ymlIdx.Logname, result.Error)
		}
	}

	return nil
}

// syncDevices 同步裝置資料（以 device_group + name 為識別鍵）
func syncDevices(tx *gorm.DB, cfg *structs.YMLConfig) error {
	if len(cfg.Devices) == 0 {
		log.Println("No devices in config.yml, skipping")
		return nil
	}

	for _, ymlGroup := range cfg.Devices {
		for _, item := range ymlGroup.Names {
			var existing entities.Device
			result := tx.Where("device_group = ? AND name = ?", ymlGroup.DeviceGroup, item.Name).First(&existing)

			if result.Error == gorm.ErrRecordNotFound {
				device := entities.Device{
					DeviceGroup: ymlGroup.DeviceGroup,
					Name:        item.Name,
					HAGroup:     item.HAGroup,
				}
				if err := tx.Create(&device).Error; err != nil {
					return fmt.Errorf("create device '%s' in group '%s' failed: %w", item.Name, ymlGroup.DeviceGroup, err)
				}
				log.Printf("Created device: %s/%s (ha_group: %s)", ymlGroup.DeviceGroup, item.Name, item.HAGroup)
			} else if result.Error == nil {
				// 更新 ha_group（可能從空改為有值，或反之）
				if err := tx.Model(&existing).Update("ha_group", item.HAGroup).Error; err != nil {
					return fmt.Errorf("update device '%s' ha_group failed: %w", item.Name, err)
				}
			} else {
				return fmt.Errorf("query device '%s/%s' failed: %w", ymlGroup.DeviceGroup, item.Name, result.Error)
			}
		}
	}

	return nil
}
