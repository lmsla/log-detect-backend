package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Insert_HistoryDataWithClient 使用指定的 ES 客戶端插入歷史資料（支援多連線）
func Insert_HistoryDataWithClient(esClient *elasticsearch.Client, historyData entities.History) {

	// 將資料轉換為 JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(historyData); err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Error encoding data,%s", err.Error()))
		// log.Fatalf("Error encoding data: %s", err)
	}
	// nowtime := time.Now().Format("20060102")

	index := fmt.Sprintf("log-detect-history-%s", time.Now().Format("20060102"))

	// 構建索引請求
	req := esapi.IndexRequest{
		Index: index, // 替換為你的索引名稱
		// DocumentID: "1",  
		Body:    &buf,
		Refresh: "true", // 刷新索引，使數據立即可用
	}

	// 執行請求
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("es insert request error,%s", err.Error()))

	}
	defer res.Body.Close()

	// 打印請求結果
	if res.IsError() {
		// log.Printf("Error response: %s", res.String())
	} else {
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("es insert response error,%s", err.Error()))
		} else {
			// log.Logrecord_no_rotate("INFO", fmt.Sprintf("Document indexed successfully: %s", response["result"]))
		}
	}
	// return res.Body

}

// Insert_HistoryData 使用預設 ES 客戶端插入歷史資料（向後兼容）
func Insert_HistoryData(historyData entities.History) {
	Insert_HistoryDataWithClient(global.Elasticsearch, historyData)
}
