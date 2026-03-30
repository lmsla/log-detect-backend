// ha_simulator.go - 模擬含 HA Group 的設備日誌，用於測試 log-detect 的 HA 過濾邏輯
//
// 使用方式:
//   go run scripts/ha_simulator.go
//
// 情境說明（依序循環）:
//   1. 全部上線          → 無告警
//   2. HA 主機失聯        → fw-01-primary 下線，fw-01-standby 仍在 → log-detect 應歸類為 standby，不告警
//   3. HA 群組全滅        → fw-01-primary + fw-01-standby 全部下線 → log-detect 應告警
//   4. 獨立設備失聯       → fw-02 下線（無 HA group）→ 應立即告警
//   5. 多台同時失聯       → fw-02 + waf-01 同時下線 → 應各自告警
//
// 設備與 devices.yml 對應（device_group: forti / waf）:
//   fw-01-primary  ha_group: fw-cluster-01
//   fw-01-standby  ha_group: fw-cluster-01
//   fw-02          ha_group: (無)
//   waf-01         ha_group: (無)
//   waf-02         ha_group: (無)

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ============================================================
// ★ 設定區 — 依實際環境修改以下變數
// ============================================================

const (
	esURL      = "https://10.99.1.213:9200" // ES 位址
	esUser     = "elastic"                  // ES 帳號（無驗證請留空）
	esPassword = "a12345678"                // ES 密碼（無驗證請留空）
	esIndex    = "logstash-forti-group"           // 寫入的索引名稱
	esField    = "host.keyword"             // log-detect 用來聚合設備名稱的欄位

	intervalSec = 60  // 每輪寫入間隔（秒），建議與 log-detect cron 週期一致
	scenarioSec = 180 // 每個情境持續時間（秒），建議 > intervalSec * 2
)

// ============================================================

// ---- 設備定義 ----

type SimDevice struct {
	Name    string
	Group   string // ES index 內不用到，僅供顯示
	HAGroup string // 僅供顯示，實際 HA 邏輯由 log-detect 側處理
	Offline bool
}

// 對應 devices.yml 的設備清單
var simDevices = []*SimDevice{
	{Name: "fw-01-primary", Group: "forti", HAGroup: "fw-cluster-01"},
	{Name: "fw-01-standby", Group: "forti", HAGroup: "fw-cluster-01"},
	{Name: "fw-02", Group: "forti", HAGroup: ""},
	{Name: "waf-01", Group: "waf", HAGroup: ""},
	{Name: "waf-02", Group: "waf", HAGroup: ""},
}

// ---- 測試情境定義 ----

type Scenario struct {
	ID          int
	Name        string
	Description string
	Expected    string
	OfflineSet  map[string]bool // 此情境中應失聯的設備名稱
}

var scenarios = []Scenario{
	{
		ID:          1,
		Name:        "全部上線",
		Description: "所有設備正常發送 heartbeat log",
		Expected:    "✅ 無告警",
		OfflineSet:  map[string]bool{},
	},
	{
		ID:          2,
		Name:        "HA 主機失聯（備機仍在線）",
		Description: "fw-01-primary 下線，fw-01-standby 仍在線（同屬 fw-cluster-01）",
		Expected:    "✅ log-detect 應將 fw-01-primary 歸類為 HA standby，不發告警",
		OfflineSet:  map[string]bool{"fw-01-primary": true},
	},
	{
		ID:          3,
		Name:        "HA 群組全滅",
		Description: "fw-01-primary 與 fw-01-standby 同時下線，整個 fw-cluster-01 無存活成員",
		Expected:    "🚨 log-detect 應對 fw-01-primary 和 fw-01-standby 都發送告警",
		OfflineSet:  map[string]bool{"fw-01-primary": true, "fw-01-standby": true},
	},
	{
		ID:          4,
		Name:        "獨立設備失聯",
		Description: "fw-02 下線（無 HA group，無法依靠備機）",
		Expected:    "🚨 log-detect 應對 fw-02 立即發送告警",
		OfflineSet:  map[string]bool{"fw-02": true},
	},
	{
		ID:          5,
		Name:        "多台獨立設備同時失聯",
		Description: "fw-02 與 waf-01 同時下線",
		Expected:    "🚨 log-detect 應分別對 fw-02 和 waf-01 發送告警",
		OfflineSet:  map[string]bool{"fw-02": true, "waf-01": true},
	},
}

// ---- ES 連線 ----

func newHAESClient() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	if esUser != "" {
		cfg.Username = esUser
		cfg.Password = esPassword
	}

	return elasticsearch.NewClient(cfg)
}

// ---- 寫入單筆 log ----

func writeHALog(client *elasticsearch.Client, index, fieldName, deviceName string, t time.Time) error {
	// 去掉 .keyword 後綴，取得實際寫入的欄位名稱
	rawField := fieldName
	if len(rawField) > 8 && rawField[len(rawField)-8:] == ".keyword" {
		rawField = rawField[:len(rawField)-8]
	}

	doc := map[string]interface{}{
		"@timestamp": t.Format("2006-01-02T15:04:05.000+08:00"),
		"message":    fmt.Sprintf("heartbeat from %s", deviceName),
		"level":      "INFO",
		rawField:     deviceName,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:   index,
		Body:    bytes.NewReader(data),
		Refresh: "false",
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES index error: %s", res.Status())
	}
	return nil
}

// ---- 套用情境 ----

func applyScenario(s Scenario) {
	for _, d := range simDevices {
		d.Offline = s.OfflineSet[d.Name]
	}
}

// ---- 顯示目前情境狀態 ----

func printScenarioHeader(s Scenario, scenarioSec int) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  情境 %d：%s\n", s.ID, s.Name)
	fmt.Printf("  說明：%s\n", s.Description)
	fmt.Printf("  預期：%s\n", s.Expected)
	fmt.Printf("  持續：%d 秒\n", scenarioSec)
	fmt.Println("  設備狀態：")
	for _, d := range simDevices {
		haTag := ""
		if d.HAGroup != "" {
			haTag = fmt.Sprintf(" [HA: %s]", d.HAGroup)
		}
		status := "🟢 在線"
		if d.Offline {
			status = "🔴 失聯（停止發 log）"
		}
		fmt.Printf("    %-20s %s%s\n", d.Name, status, haTag)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// ---- main ----

func main() {
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║       Log Detect HA Group 情境模擬器             ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Printf("  ES URL      : %s\n", esURL)
	fmt.Printf("  Index       : %s\n", esIndex)
	fmt.Printf("  Field       : %s\n", esField)
	fmt.Printf("  寫入間隔    : %d 秒\n", intervalSec)
	fmt.Printf("  情境持續    : %d 秒（每個情境跑完後自動切換）\n", scenarioSec)
	fmt.Println()
	fmt.Println("  提示：情境切換不影響 log-detect 的 cron 排程")
	fmt.Println("        log-detect 偵測到缺失的設備才會觸發 HA 判斷")
	fmt.Println("  按 Ctrl+C 停止")

	client, err := newHAESClient()
	if err != nil {
		log.Fatalf("❌ 建立 ES 客戶端失敗: %v", err)
	}

	// 測試連線
	res, err := client.Info()
	if err != nil {
		log.Fatalf("❌ 連接 ES 失敗: %v", err)
	}
	res.Body.Close()
	fmt.Println("\n✅ ES 連接成功，開始模擬...")

	scenarioIdx := 0
	round := 0
	scenarioStart := time.Now()

	// 套用第一個情境
	currentScenario := scenarios[scenarioIdx]
	applyScenario(currentScenario)
	printScenarioHeader(currentScenario, scenarioSec)

	for {
		now := time.Now()
		round++

		// 檢查是否該切換情境
		if time.Since(scenarioStart) >= time.Duration(scenarioSec)*time.Second {
			scenarioIdx = (scenarioIdx + 1) % len(scenarios)
			currentScenario = scenarios[scenarioIdx]
			applyScenario(currentScenario)
			scenarioStart = now
			printScenarioHeader(currentScenario, scenarioSec)
		}

		// 寫入這一輪所有在線設備的 log
		onlineCount := 0
		offlineCount := 0
		offlineNames := []string{}

		for _, d := range simDevices {
			if d.Offline {
				offlineCount++
				offlineNames = append(offlineNames, d.Name)
				continue
			}
			if err := writeHALog(client, esIndex, esField, d.Name, now); err != nil {
				fmt.Printf("  ⚠️  寫入失敗 [%s]: %v\n", d.Name, err)
			} else {
				onlineCount++
			}
		}

		// 計算本情境剩餘時間
		remaining := scenarioSec - int(time.Since(scenarioStart).Seconds())

		fmt.Printf("[%s] 情境%d | 寫入 %d 台 | 停止 %d 台",
			now.Format("15:04:05"), currentScenario.ID, onlineCount, offlineCount)
		if len(offlineNames) > 0 {
			fmt.Printf(" [%v]", offlineNames)
		}
		fmt.Printf(" | 剩餘 %ds\n", remaining)

		time.Sleep(time.Duration(intervalSec) * time.Second)
	}
}
