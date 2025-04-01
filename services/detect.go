package services

import (
	// "context"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"log-detect/models"
	"time"
)

func Detect(execute_time time.Time, index string, field string, period string, unit int, receiver []string, subject string, logname string, device_group string) {
	timenow := execute_time.Format("2006-01-02T15:04") + ":00.000+08:00"
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

	var result_list []string
	result := SearchRequest(index, field, time3_str, timenow)
	// fmt.Println("資料搜尋結果:",result)

	for i := range result.Aggregations.Num2.Buckets {
		// fmt.Println("host", result.Aggregations.Num2.Buckets[i].Key)
		// fmt.Println("doc_count", result.Aggregations.Num2.Buckets[i].DocCount)
		result_list = append(result_list, result.Aggregations.Num2.Buckets[i].Key)
	}
	// fmt.Println("result.Aggregations.Num2.Buckets len",len(result.Aggregations.Num2.Buckets))

	fmt.Println("執行時間:", timenow)
	fmt.Println("檢查起始時間:", time3_str)

	deviceslist, err := GetDevicesDataByGroupName(device_group)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("get devices data error: %s",err.Error()))
	}
	var device_list []string
	var origin_list []entities.Device
	var new_list []entities.Device

	// 資產管理清單未建立，自動把從 ES 撈出的設備加入群組中
	if len(deviceslist) == 0 {
		// fmt.Println("no data")
		device_list = result_list

		for _, device := range result_list {
			newDevice := entities.Device{
				Common:      models.Common{},
				DeviceGroup: device_group,
				Name:        device,
			}
			origin_list = append(origin_list, newDevice)
		}

		CreateDevice(origin_list)

	} else {
		// db 中的 device list
		for _, device := range deviceslist {
			device_list = append(device_list, device.Name)
		}
	}

	/// added: 搜尋結果中新增的 device ; removed: 搜尋結果中缺失的 device
	added, removed,intersection := ListCompare(device_list, result_list)
	fmt.Println("新增的設備:", added)

	// 將偵測到的新設備寫入 devices table 中
	if len(added) != 0 {
		for _, device := range added {
			newDevice := entities.Device{
				Common:      models.Common{},
				DeviceGroup: device_group,
				Name:        device,
			}
			new_list = append(new_list, newDevice)
		}
		CreateDevice(new_list)
	}

	// 1. 找出要刪除的重複資料的 id
	rows, err := global.Mysql.Raw("SELECT MIN(id) as id FROM devices GROUP BY name, device_group HAVING COUNT(*) > 1").Rows()
	if err != nil {
		// 處理錯誤
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("error querying duplicate devices: %s",err.Error()))
		return
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
	}

	// 2. 刪除重複資料
	result_to := global.Mysql.Where("id IN (?)", ids).Delete(&entities.Device{})
	if result_to.Error != nil {
		// 處理錯誤
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("error deleting duplicate devices: %s",result_to.Error.Error()))
		return
	}

	fmt.Println("遺失的設備: ", removed)
	cc := []string{""}
	bcc := []string{""}
	if removed != nil {
		// SendEmail(receiver,subject,logname,removed)
		Mail4(receiver, cc, bcc, subject, logname, removed)
		mailHistory := entities.MailHistory {
			Date: date_time,
			Time: hour_time,
			Logname: logname,
			Sended: true,
		}
		CreateMailHistory(mailHistory)
	}
	// 紀錄檢查結果到 es 中
	for _, device := range intersection {

		historyData := entities.History{
			Logname: logname,
			DeviceGroup: device_group ,
			Name: device,
			Lost: "false",
			LostNum: 2,
			Date: date_time,
			Time: hour_time,
			DateTime: timenow,
			Period: period,
			Unit: unit,
		}
		Insert_HistoryData(historyData)
		CreateHistory(historyData)
		
	}

	// 紀錄缺失設備到 history table 中
	for _, device := range removed {

		historyData := entities.History{
			Logname: logname,
			DeviceGroup: device_group ,
			Name: device,
			Lost: "true",
			LostNum: 1,
			Date: date_time,
			Time: hour_time,
			DateTime: timenow,
			Period: period,
			Unit: unit,
		}
		Insert_HistoryData(historyData)
		CreateHistory(historyData)
		
	}

}
