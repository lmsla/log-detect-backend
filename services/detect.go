package services

import (
	// "context"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"log-detect/models"
	"time"

	// "github.com/elastic/go-elasticsearch/v8"
)

func Detect(execute_time time.Time, indexID int, index string, field string, period string, unit int, receiver []string, subject string, logname string, device_group string) {
	timenow := execute_time.Format("2006-01-02 15:04:05")
	timenow_es := execute_time.Format("2006-01-02T15:04") + ":00.000+08:00" // ISO 8601 格式用於 ES 查詢
	// var cronjob string
	var time3_str string
	date_time := execute_time.Format("2006-01-02")
	hour_time := execute_time.Format("15:04")

	if period == "minutes" {
		time3 := execute_time.Add(time.Minute * -time.Duration(unit))
		time3_str = time3.Format("2006-01-02T15:04") + ":00.000+08:00"
	} else if period == "hours" {
		time3 := execute_time.Add(time.Hour * -time.Duration(unit))
		time3_str = time3.Format("2006-01-02T15:04") + ":00.000+08:00"
	}

	// 取得該 Index 對應的 ES 客戶端
	manager := GetESConnectionManager()
	esClient, err := manager.GetClientForIndex(indexID)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to get ES client for index %d: %s", indexID, err.Error()))
		log.Logrecord_no_rotate("WARNING", fmt.Sprintf("Falling back to default ES client for index %d", indexID))
		// Fallback 到預設客戶端
		esClient = manager.GetDefaultClient()
		if esClient == nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Default ES client is nil, cannot execute detect for index %d", indexID))
			return
		}
	}

	var result_list []string
	result := SearchRequestWithClient(esClient, index, field, time3_str, timenow_es)
	// fmt.Println("資料搜尋結果:",result)

	for i := range result.Aggregations.Num2.Buckets {
		// fmt.Println("host", result.Aggregations.Num2.Buckets[i].Key)
		// fmt.Println("doc_count", result.Aggregations.Num2.Buckets[i].DocCount)
		result_list = append(result_list, result.Aggregations.Num2.Buckets[i].Key)
	}

	fmt.Println("執行時間:", timenow)
	fmt.Println("檢查起始時間:", time3_str)

	deviceslist, err := GetDevicesDataByGroupName(device_group)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("get devices data error: %s", err.Error()))
	}
	var device_list []string
	var origin_list []entities.Device
	var new_list []entities.Device

	// 資產管理清單未建立，自動把從 ES 撈出的設備加入群組中
	// 只要同群組中不存在，就允許建立；同名設備可存在於不同群組
	if len(deviceslist) == 0 {
		for _, device := range result_list {
			var existCount int64
			if err := global.Mysql.Model(&entities.Device{}).
				Where("device_group = ? AND name = ?", device_group, device).
				Count(&existCount).Error; err != nil {
				log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Check device existence failed for '%s': %s", device, err.Error()))
				continue
			}
			if existCount > 0 {
				log.Logrecord_no_rotate("INFO", fmt.Sprintf(
					"Auto-discovery skipped: device '%s' already exists in current group '%s'",
					device, device_group))
				continue
			}
			newDevice := entities.Device{
				Common:      models.Common{},
				DeviceGroup: device_group,
				Name:        device,
				HAGroup:     lookupHAGroupFromYML(device_group, device),
			}
			origin_list = append(origin_list, newDevice)
			device_list = append(device_list, device)
		}

		if len(origin_list) > 0 {
			CreateDevice(origin_list)
		}

	} else {
		// db 中的 device list
		for _, device := range deviceslist {
			device_list = append(device_list, device.Name)
		}
	}

	/// added: 搜尋結果中新增的 device ; removed: 搜尋結果中缺失的 device
	added, removed, intersection := ListCompare(device_list, result_list)
	fmt.Println("新增的設備:", added)

	// 將偵測到的新設備寫入 devices table 中
	// 只檢查同一個 device_group 是否已存在，允許同名設備出現在不同群組
	if len(added) != 0 {
		for _, device := range added {
			var existCount int64
			if err := global.Mysql.Model(&entities.Device{}).
				Where("device_group = ? AND name = ?", device_group, device).
				Count(&existCount).Error; err != nil {
				log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Check device existence failed for '%s': %s", device, err.Error()))
				continue
			}
			if existCount > 0 {
				log.Logrecord_no_rotate("INFO", fmt.Sprintf(
					"Auto-discovery skipped: device '%s' already exists in current group '%s'",
					device, device_group))
				continue
			}
			newDevice := entities.Device{
				Common:      models.Common{},
				DeviceGroup: device_group,
				Name:        device,
				HAGroup:     lookupHAGroupFromYML(device_group, device),
			}
			new_list = append(new_list, newDevice)
		}
		if len(new_list) > 0 {
			CreateDevice(new_list)
		}
	}

	// 清除重複裝置：保留每組 (name, device_group) 中 id 最小的那筆（最早建立），刪除其餘重複
	var dupIDs []int
	dupRows, err := global.Mysql.Raw(`
		SELECT d.id
		FROM devices d
		INNER JOIN (
			SELECT name, device_group, MIN(id) AS keep_id
			FROM devices
			GROUP BY name, device_group
			HAVING COUNT(*) > 1
		) AS dup ON d.name = dup.name AND d.device_group = dup.device_group
		WHERE d.id != dup.keep_id
	`).Rows()
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("error querying duplicate devices: %s", err.Error()))
		return
	}
	defer dupRows.Close()
	for dupRows.Next() {
		var id int
		if err := dupRows.Scan(&id); err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("error scanning duplicate device id: %s", err.Error()))
			return
		}
		dupIDs = append(dupIDs, id)
	}

	if len(dupIDs) > 0 {
		if err := global.Mysql.Where("id IN ?", dupIDs).Delete(&entities.Device{}).Error; err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("error deleting duplicate devices: %s", err.Error()))
			return
		}
	}

	// === HA Group 過濾 ===
	trulyRemoved, standbyDevices := filterHAGroups(device_group, removed, intersection)

	fmt.Println("真正失聯的設備: ", trulyRemoved)
	fmt.Println("HA 待命的設備: ", standbyDevices)

	cc := []string{""}
	bcc := []string{""}
	if len(trulyRemoved) > 0 {
		// 取得 HA 群組資訊用於郵件
		haInfo := getHAGroupInfo(device_group, trulyRemoved)
		Mail4WithHA(receiver, cc, bcc, subject, logname, trulyRemoved, haInfo)
		mailHistory := entities.MailHistory{
			Date:    date_time,
			Time:    hour_time,
			Logname: logname,
			Sended:  true,
		}
		CreateMailHistory(mailHistory)
	}

	// 紀錄在線設備到歷史記錄中
	for _, device := range intersection {
		historyData := entities.History{
			Logname:      logname,
			DeviceGroup:  device_group,
			Name:         device,
			Status:       "online",
			Lost:         "false",
			LostNum:      0,
			Date:         date_time,
			Time:         hour_time,
			DateTime:     timenow,
			Timestamp:    execute_time.Unix(),
			Period:       period,
			Unit:         unit,
			ResponseTime: 100,
			DataCount:    1,
		}

		// === Feature Toggle: History ===
		if global.EnvConfig.Features.TimescaleDB {
			if global.BatchWriter != nil {
				if err := global.BatchWriter.AddHistory(historyData); err != nil {
					log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to add history to batch: %s", err.Error()))
				}
			}
		}
	}

	// 紀錄真正失聯設備到 history table 中
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Processing %d truly offline devices", len(trulyRemoved)))
	for _, device := range trulyRemoved {
		historyData := entities.History{
			Logname:      logname,
			DeviceGroup:  device_group,
			Name:         device,
			Status:       "offline",
			Lost:         "true",
			LostNum:      1,
			Date:         date_time,
			Time:         hour_time,
			DateTime:     timenow,
			Timestamp:    execute_time.Unix(),
			Period:       period,
			Unit:         unit,
			ResponseTime: 0,
			DataCount:    0,
			ErrorMsg:     "Device not found in logs",
			ErrorCode:    "DEVICE_OFFLINE",
		}

		// === Feature Toggle: History ===
		if global.EnvConfig.Features.TimescaleDB {
			if global.BatchWriter != nil {
				if err := global.BatchWriter.AddHistory(historyData); err != nil {
					log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to add history to batch for device %s: %s", device, err.Error()))
				}
			}
		}
	}

	// 紀錄 HA 待命設備到 history table 中
	for _, device := range standbyDevices {
		historyData := entities.History{
			Logname:      logname,
			DeviceGroup:  device_group,
			Name:         device,
			Status:       "standby",
			Lost:         "false",
			LostNum:      0,
			Date:         date_time,
			Time:         hour_time,
			DateTime:     timenow,
			Timestamp:    execute_time.Unix(),
			Period:       period,
			Unit:         unit,
			ResponseTime: 0,
			DataCount:    0,
			ErrorMsg:     "HA standby - partner device is online",
			ErrorCode:    "HA_STANDBY",
		}

		// === Feature Toggle: History ===
		if global.EnvConfig.Features.TimescaleDB {
			if global.BatchWriter != nil {
				if err := global.BatchWriter.AddHistory(historyData); err != nil {
					log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to add history to batch for standby device %s: %s", device, err.Error()))
				}
			}
		}
	}

}

// filterHAGroups 根據 HA 群組過濾失聯裝置
// 回傳 trulyRemoved（真正失聯，需告警）和 standbyDevices（HA 待命，不告警）
func filterHAGroups(deviceGroup string, removed []string, online []string) (trulyRemoved []string, standbyDevices []string) {
	if len(removed) == 0 {
		return nil, nil
	}

	// 合併 removed + online 名稱，一次查詢所有裝置的 ha_group
	allNames := append(append([]string{}, removed...), online...)
	var devices []entities.Device
	if err := global.Mysql.Where("device_group = ? AND name IN (?)", deviceGroup, allNames).Find(&devices).Error; err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to query HA groups: %s", err.Error()))
		// 查詢失敗時退回原始行為：所有 removed 都視為真正失聯
		return removed, nil
	}

	// 建立 name → ha_group 映射
	nameToHA := make(map[string]string)
	for _, d := range devices {
		nameToHA[d.Name] = d.HAGroup
	}

	// 建立 online 裝置的 ha_group 集合
	onlineHAGroups := make(map[string]bool)
	for _, name := range online {
		haGroup := nameToHA[name]
		if haGroup != "" {
			onlineHAGroups[haGroup] = true
		}
	}

	// 分類 removed 裝置
	for _, name := range removed {
		haGroup := nameToHA[name]
		if haGroup == "" {
			// 獨立裝置 → 真正失聯
			trulyRemoved = append(trulyRemoved, name)
		} else if onlineHAGroups[haGroup] {
			// HA 群組中有成員在線 → 待命
			standbyDevices = append(standbyDevices, name)
		} else {
			// HA 群組全部失聯 → 真正失聯
			trulyRemoved = append(trulyRemoved, name)
		}
	}

	return trulyRemoved, standbyDevices
}

// lookupHAGroupFromYML 從已載入的 YMLConfig 查詢裝置的 ha_group 設定
// 用於自動發現裝置時，讓新建的 Device 記錄帶上正確的 ha_group，而非永遠空白
// 若 YMLConfig 未載入或找不到對應裝置，回傳空字串（獨立裝置行為）
func lookupHAGroupFromYML(deviceGroup, deviceName string) string {
	if global.YMLConfig == nil {
		return ""
	}
	for _, group := range global.YMLConfig.Devices {
		if group.DeviceGroup != deviceGroup {
			continue
		}
		for _, item := range group.Names {
			if item.Name == deviceName {
				return item.HAGroup
			}
		}
	}
	return ""
}

// getHAGroupInfo 取得失聯裝置的 HA 群組資訊（用於郵件顯示）
func getHAGroupInfo(deviceGroup string, trulyRemoved []string) map[string]string {
	haInfo := make(map[string]string)
	if len(trulyRemoved) == 0 {
		return haInfo
	}

	var devices []entities.Device
	if err := global.Mysql.Where("device_group = ? AND name IN (?)", deviceGroup, trulyRemoved).Find(&devices).Error; err != nil {
		return haInfo
	}

	for _, d := range devices {
		haInfo[d.Name] = d.HAGroup
	}
	return haInfo
}
