// es_log_simulator.go - 持續向 ES 寫入模擬日誌，讓 log-detect 能偵測到設備連線狀態
// 使用方式: go run scripts/es_log_simulator.go
//
// 環境變數（選填）:
//   ES_URL        ES 位址          (預設: http://localhost:9200)
//   ES_USER       ES 帳號          (選填)
//   ES_PASSWORD   ES 密碼          (選填)
//   ES_INDEX      寫入的索引       (預設: logstash-log_detect)
//   ES_FIELD      host 欄位名稱    (預設: host.keyword)
//   INTERVAL_SEC  寫入間隔（秒）   (預設: 30)
//
// 模擬場景：
//   - 共 10 台設備（3 群組），依不同機率間歇性失聯
//   - 每個 INTERVAL_SEC 寫一輪（只有「上線中」的設備才寫文件）
//   - 每 20 輪隨機切換設備的上下線狀態

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json" 
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)
 
// ---- 設定 ----

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// ---- 模擬設備清單 ----

type Device struct {
	Name    string
	Group   string
	Offline bool // 是否目前失聯（不寫 log）
}

var devices = []*Device{
	{Name: "fw-01", Group: "firewall"},
	{Name: "fw-02", Group: "firewall"},
	{Name: "sw-core-01", Group: "switch"},
	{Name: "sw-core-02", Group: "switch"},
	{Name: "sw-access-01", Group: "switch"},
	{Name: "sw-access-02", Group: "switch"},
	{Name: "app-server-01", Group: "server"},
	{Name: "app-server-02", Group: "server"},
	{Name: "db-server-01", Group: "server"},
	{Name: "db-server-02", Group: "server"},
}

// ---- ES 連線 ----

func newESClient() (*elasticsearch.Client, error) {
	esURL := getEnv("ES_URL", "http://localhost:9200")
	user := getEnv("ES_USER", "")
	password := getEnv("ES_PASSWORD", "")

	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	if user != "" {
		cfg.Username = user
		cfg.Password = password
	}

	return elasticsearch.NewClient(cfg)
}

// ---- 寫入單筆 log 文件 ----

func writeLog(client *elasticsearch.Client, index, fieldName, deviceName string, t time.Time) error {
	doc := map[string]interface{}{
		"@timestamp": t.Format("2006-01-02T15:04:05.000+08:00"),
		"message":    fmt.Sprintf("heartbeat log from %s", deviceName),
		"level":      "INFO",
	}

	// 去掉 .keyword 後綴取得實際欄位名
	// ES dynamic mapping 會自動為 text 欄位建立 .keyword sub-field（keyword 型別）
	// 例如：寫入 host="fw-01" → ES 自動建立 host.keyword 供 terms 聚合使用
	rawField := fieldName
	if len(rawField) > 8 && rawField[len(rawField)-8:] == ".keyword" {
		rawField = rawField[:len(rawField)-8]
	}
	doc[rawField] = deviceName

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

// ---- 隨機切換設備上下線狀態 ----

func randomizeOfflineStatus() {
	for _, d := range devices {
		// 每次重新隨機，約 15% 機率失聯
		d.Offline = rand.Float64() < 0.15
	}
}

// ---- main ----

func main() {
	rand.Seed(time.Now().UnixNano())

	esIndex := getEnv("ES_INDEX", "logstash-log_detect")
	esField := getEnv("ES_FIELD", "host.keyword")
	intervalSec := 30
	if v := getEnv("INTERVAL_SEC", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			intervalSec = n
		}
	}

	fmt.Println("=== Log Detect ES 日誌模擬器 ===")
	fmt.Printf("  ES URL    : %s\n", getEnv("ES_URL", "http://localhost:9200"))
	fmt.Printf("  Index     : %s\n", esIndex)
	fmt.Printf("  Field     : %s\n", esField)
	fmt.Printf("  Interval  : %d 秒\n", intervalSec)
	fmt.Println()
	fmt.Println("設備清單：")
	for _, d := range devices {
		fmt.Printf("  [%s] %s\n", d.Group, d.Name)
	}
	fmt.Println()

	client, err := newESClient()
	if err != nil {
		log.Fatalf("❌ 建立 ES 客戶端失敗: %v", err)
	}

	// 測試連線
	res, err := client.Info()
	if err != nil {
		log.Fatalf("❌ 連接 ES 失敗: %v", err)
	}
	res.Body.Close()
	fmt.Println("✅ ES 連接成功，開始寫入模擬日誌...")
	fmt.Println("（按 Ctrl+C 停止）")
	fmt.Println()

	round := 0
	for {
		round++
		now := time.Now()

		// 每 20 輪重新隨機設備狀態，製造失聯/恢復的變化
		if round%20 == 1 {
			randomizeOfflineStatus()
			fmt.Printf("[%s] 🔄 重新隨機設備狀態（第 %d 輪）\n",
				now.Format("15:04:05"), round)
		}

		onlineCount := 0
		offlineCount := 0

		for _, d := range devices {
			if d.Offline {
				offlineCount++
				continue // 失聯設備不寫 log
			}
			if err := writeLog(client, esIndex, esField, d.Name, now); err != nil {
				fmt.Printf("  ⚠️  寫入失敗 [%s]: %v\n", d.Name, err)
			} else {
				onlineCount++
			}
		}

		fmt.Printf("[%s] ✅ 寫入 %d 台設備 log（失聯 %d 台）\n",
			now.Format("15:04:05"), onlineCount, offlineCount)

		if offlineCount > 0 {
			fmt.Print("  失聯設備：")
			for _, d := range devices {
				if d.Offline {
					fmt.Printf("%s ", d.Name)
				}
			}
			fmt.Println()
		}

		time.Sleep(time.Duration(intervalSec) * time.Second)
	}
}
