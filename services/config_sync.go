package services

import (
	"fmt"
	"log"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/structs"
	"time"

	"gorm.io/gorm"
)

// SyncConfigToDB 將 config.yml 的擴充配置同步至資料庫
// 策略：以 YML 為唯一來源，全量覆蓋（upsert + 刪除 YML 中已移除的記錄）
// 執行順序：
//  1. ES Connections upsert（先建立連線供後續 Indices 參照）
//  2. Device Groups upsert（先建立群組，devices 才能正確關聯）
//  3. Targets + Indices upsert + 刪除
//  4. Devices upsert + 刪除
//  5. ES Connections 刪除（必須在 Indices 清理後才執行，避免 FK 牽連）
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

	// Step 1: ES Connections upsert（先建立連線供後續 Indices 參照）
	if err := syncESConnectionsUpsert(tx, cfg); err != nil {
		tx.Rollback()
		return fmt.Errorf("sync ES connections (upsert) failed: %w", err)
	}

	// Step 2: Device Groups upsert（先確保群組存在，syncDevices 才能正確寫入）
	if err := syncDeviceGroups(tx, cfg); err != nil {
		tx.Rollback()
		return fmt.Errorf("sync device groups failed: %w", err)
	}

	// Step 3: 同步 Targets + Indices（含刪除已移除記錄）
	if err := syncTargets(tx, cfg); err != nil {
		tx.Rollback()
		return fmt.Errorf("sync targets failed: %w", err)
	}

	// Step 4: 同步 Devices（含刪除已移除記錄）
	if err := syncDevices(tx, cfg); err != nil {
		tx.Rollback()
		return fmt.Errorf("sync devices failed: %w", err)
	}

	// Step 5: ES Connections 刪除（在 Indices 清理完成後才安全執行）
	if err := syncESConnectionsDelete(tx, cfg); err != nil {
		tx.Rollback()
		return fmt.Errorf("sync ES connections (delete) failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit config sync: %w", err)
	}

	log.Println("Config sync completed successfully")
	return nil
}

// syncDeviceGroups 同步 device_groups 表（以 name 為識別鍵）
// 策略：只做 upsert，不刪除 —— YML 移除的群組可能仍有手動建立的裝置，
//
//	刪除需透過 API 手動確認，避免誤刪資料
func syncDeviceGroups(tx *gorm.DB, cfg *structs.YMLConfig) error {
	if len(cfg.Devices) == 0 {
		log.Println("No device groups in config, skipping")
		return nil
	}

	for _, ymlGroup := range cfg.Devices {
		if ymlGroup.DeviceGroup == "" {
			continue
		}

		var existing entities.DeviceGroup
		result := tx.Where("name = ?", ymlGroup.DeviceGroup).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			newGroup := entities.DeviceGroup{
				Name:        ymlGroup.DeviceGroup,
				Description: "",
			}
			if err := tx.Create(&newGroup).Error; err != nil {
				return fmt.Errorf("create device group '%s' failed: %w", ymlGroup.DeviceGroup, err)
			}
			log.Printf("Created device group: %s", ymlGroup.DeviceGroup)
		} else if result.Error != nil {
			return fmt.Errorf("query device group '%s' failed: %w", ymlGroup.DeviceGroup, result.Error)
		}
		// 已存在則跳過（名稱是唯一鍵，不需要更新）
	}

	return nil
}

// syncESConnectionsUpsert 新增或更新 YML 中定義的 ES 連線配置（以 name 為識別鍵）
func syncESConnectionsUpsert(tx *gorm.DB, cfg *structs.YMLConfig) error {
	if len(cfg.ESConnections) == 0 {
		log.Println("No ES connections in config.yml, skipping upsert")
		return nil
	}

	for _, ymlConn := range cfg.ESConnections {
		var existing entities.ESConnection
		result := tx.Where("name = ? AND (deleted_at IS NULL OR deleted_at = 0)", ymlConn.Name).First(&existing)

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
			// 新增（或曾被軟刪除後重新出現：建立新記錄）
			if err := tx.Create(&conn).Error; err != nil {
				return fmt.Errorf("create ES connection '%s' failed: %w", ymlConn.Name, err)
			}
			log.Printf("Created ES connection: %s", ymlConn.Name)
		} else if result.Error == nil {
			// 更新現有連線
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

// syncESConnectionsDelete 刪除 DB 中有但 YML 中已移除的 ES 連線配置
// 必須在 syncTargets（含 Indices 清理）之後執行，確保無 index 再參照這些連線
func syncESConnectionsDelete(tx *gorm.DB, cfg *structs.YMLConfig) error {
	// 載入所有有效的 ES 連線（未軟刪除）
	var allExisting []entities.ESConnection
	if err := tx.Where("deleted_at IS NULL OR deleted_at = 0").Find(&allExisting).Error; err != nil {
		return fmt.Errorf("load existing ES connections failed: %w", err)
	}

	// 建立 YML name 集合
	ymlNames := make(map[string]bool, len(cfg.ESConnections))
	for _, ymlConn := range cfg.ESConnections {
		ymlNames[ymlConn.Name] = true
	}

	for _, conn := range allExisting {
		if ymlNames[conn.Name] {
			continue
		}
		// 確認此連線已無任何 index 或 monitor 參照（避免遺留孤立 FK）
		var indexRef int64
		tx.Model(&entities.Index{}).Where("es_connection_id = ?", conn.ID).Count(&indexRef)
		var monitorRef int64
		tx.Model(&entities.ElasticsearchMonitor{}).Where("es_connection_id = ?", conn.ID).Count(&monitorRef)

		if indexRef > 0 || monitorRef > 0 {
			log.Printf("WARNING: ES connection '%s' not in YML but still referenced (%d indices, %d monitors), skipping delete",
				conn.Name, indexRef, monitorRef)
			continue
		}

		// 軟刪除（與 API 刪除保持一致）
		now := int(time.Now().Unix())
		if err := tx.Model(&conn).Update("deleted_at", now).Error; err != nil {
			return fmt.Errorf("soft-delete ES connection '%s' failed: %w", conn.Name, err)
		}
		log.Printf("Soft-deleted ES connection: %s (not in YML)", conn.Name)
	}

	return nil
}

// syncTargets 同步監控目標與索引（以 subject 為識別鍵，indices 以 logname 為識別鍵）
// 策略：全量覆蓋 — YML 有的 upsert，DB 有但 YML 沒有的刪除
func syncTargets(tx *gorm.DB, cfg *structs.YMLConfig) error {
	// 預先載入所有現有 targets（含 Indices），用於後續比對與刪除
	var allExisting []entities.Target
	if err := tx.Preload("Indices").Find(&allExisting).Error; err != nil {
		return fmt.Errorf("load existing targets failed: %w", err)
	}
	existingBySubject := make(map[string]*entities.Target, len(allExisting))
	for i := range allExisting {
		existingBySubject[allExisting[i].Subject] = &allExisting[i]
	}

	if len(cfg.Targets) == 0 {
		log.Println("WARNING: No targets defined in config.yml — all existing targets will be removed from DB")
	}

	ymlSubjects := make(map[string]bool)

	for _, ymlTarget := range cfg.Targets {
		ymlSubjects[ymlTarget.Subject] = true

		if existing, ok := existingBySubject[ymlTarget.Subject]; ok {
			// 更新現有 Target
			// 注意：必須用 struct-based Updates + Select，讓 GORM 套用 serializer:json tag
			// 若用 map[string]interface{} 方式更新，GORM 會繞過序列化器，導致 to 欄位寫入格式錯誤
			if err := tx.Model(existing).Select("To", "Enable").Updates(entities.Target{
				To:     entities.To(ymlTarget.Receiver),
				Enable: ymlTarget.Enable,
			}).Error; err != nil {
				return fmt.Errorf("update target '%s' failed: %w", ymlTarget.Subject, err)
			}
			// 同步 Indices（含刪除已移除的 index）
			if err := syncIndicesForTarget(tx, existing, ymlTarget.Indices); err != nil {
				return err
			}
			log.Printf("Updated target: %s", ymlTarget.Subject)
		} else {
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
		}
	}

	// 刪除 DB 中有但 YML 中沒有的 targets
	for subject, existing := range existingBySubject {
		if ymlSubjects[subject] {
			continue
		}
		// 先取得此 target 關聯的 index ID 清單
		linkedIndexIDs := make([]int, 0, len(existing.Indices))
		for _, idx := range existing.Indices {
			linkedIndexIDs = append(linkedIndexIDs, idx.ID)
		}
		// 清除 cron_lists
		if err := tx.Exec("DELETE FROM cron_lists WHERE target_id = ?", existing.ID).Error; err != nil {
			return fmt.Errorf("delete cron_lists for target '%s' failed: %w", subject, err)
		}
		// 清除 indices_targets 關聯
		if err := tx.Exec("DELETE FROM indices_targets WHERE target_id = ?", existing.ID).Error; err != nil {
			return fmt.Errorf("delete indices_targets for target '%s' failed: %w", subject, err)
		}
		// 刪除 target（YML 唯一來源，直接硬刪除）
		if err := tx.Delete(existing).Error; err != nil {
			return fmt.Errorf("delete target '%s' failed: %w", subject, err)
		}
		// 刪除已孤立的 index 記錄
		if len(linkedIndexIDs) > 0 {
			if err := tx.Where("id IN ?", linkedIndexIDs).Delete(&entities.Index{}).Error; err != nil {
				return fmt.Errorf("delete orphaned indices for target '%s' failed: %w", subject, err)
			}
		}
		log.Printf("Deleted target: %s (not in YML)", subject)
	}

	return nil
}

// syncIndicesForTarget 同步某個 Target 下的 Indices（以 logname 為識別鍵）
// 策略：全量覆蓋 — YML 有的 upsert，此 target 下 DB 有但 YML 沒有的解除關聯並刪除孤立記錄
func syncIndicesForTarget(tx *gorm.DB, target *entities.Target, ymlIndices []structs.YMLIndex) error {
	// 建立 YML logname 集合，用於比對哪些 index 需要保留
	ymlLognames := make(map[string]bool, len(ymlIndices))
	for _, ymlIdx := range ymlIndices {
		ymlLognames[ymlIdx.Logname] = true
	}

	// 刪除此 target 下，YML 中已移除的 indices
	for _, existingIdx := range target.Indices {
		if ymlLognames[existingIdx.Logname] {
			continue
		}
		// 移除 many2many 關聯
		if err := tx.Exec("DELETE FROM indices_targets WHERE target_id = ? AND index_id = ?",
			target.ID, existingIdx.ID).Error; err != nil {
			return fmt.Errorf("unlink index '%s' from target '%s' failed: %w", existingIdx.Logname, target.Subject, err)
		}
		// 若此 index 不再被任何 target 關聯，則一併刪除 index 記錄
		var refCount int64
		if err := tx.Raw("SELECT COUNT(*) FROM indices_targets WHERE index_id = ?", existingIdx.ID).Scan(&refCount).Error; err != nil {
			return fmt.Errorf("check index '%s' reference count failed: %w", existingIdx.Logname, err)
		}
		if refCount == 0 {
			if err := tx.Delete(&existingIdx).Error; err != nil {
				return fmt.Errorf("delete orphaned index '%s' failed: %w", existingIdx.Logname, err)
			}
			log.Printf("Deleted orphaned index: %s", existingIdx.Logname)
		}
		log.Printf("Unlinked index '%s' from target '%s' (not in YML)", existingIdx.Logname, target.Subject)
	}

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
// 策略：全量覆蓋 — YML 有的 upsert，同群組 DB 有但 YML 沒有的刪除
// 注意：只處理 YML 中有定義的 device_group；未出現的群組不受影響（保留自動發現的裝置）
func syncDevices(tx *gorm.DB, cfg *structs.YMLConfig) error {
	if len(cfg.Devices) == 0 {
		log.Println("No devices in config.yml, skipping")
		return nil
	}

	for _, ymlGroup := range cfg.Devices {
		// 建立此群組在 YML 中的 name 集合
		ymlNames := make(map[string]bool, len(ymlGroup.Names))
		for _, item := range ymlGroup.Names {
			ymlNames[item.Name] = true
		}

		// 載入 DB 中此群組的所有現有 devices
		var existingDevices []entities.Device
		if err := tx.Where("device_group = ?", ymlGroup.DeviceGroup).Find(&existingDevices).Error; err != nil {
			return fmt.Errorf("load existing devices for group '%s' failed: %w", ymlGroup.DeviceGroup, err)
		}

		// 刪除 DB 有但 YML 中已移除的 devices
		for _, dev := range existingDevices {
			if ymlNames[dev.Name] {
				continue
			}
			if err := tx.Delete(&dev).Error; err != nil {
				return fmt.Errorf("delete device '%s/%s' failed: %w", ymlGroup.DeviceGroup, dev.Name, err)
			}
			log.Printf("Deleted device: %s/%s (not in YML)", ymlGroup.DeviceGroup, dev.Name)
		}

		// Upsert：新增或更新 YML 中定義的 devices
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
