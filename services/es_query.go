package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log-detect/global"
	"log-detect/log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)


// SearchRequestWithClient 使用指定的 ES 客戶端執行查詢（支援多連線）
func SearchRequestWithClient(esClient *elasticsearch.Client, index string, field string, timefrom string, timeto string) Search_Request {
	querybody := fmt.Sprintf(`{
		"aggs": {
		  "2": {
			"terms": {
			  "field": "%s",
			  "order": {
				"_count": "desc"
			  },
			  "size": 100,
			  "shard_size": 25
			}
		  }
		},
		"size": 0,
		"fields": [
		  {
			"field": "@timestamp",
			"format": "date_time"
		  }
		],
		"script_fields": {},
		"stored_fields": [
		  "*"
		],
		"runtime_mappings": {},
		"_source": {
		  "excludes": []
		},
		"query": {
		  "bool": {
			"must": [],
			"filter": [
			  {
				"range": {
				  "@timestamp": {
					"format": "strict_date_optional_time",
					"gte": "%s",
					"lte": "%s"
				  }
				}
			  }
			],
			"should": [],
			"must_not": []
		  }
		}
	  }`, field, timefrom, timeto)
	// fmt.Println(querybody)
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(querybody),
	}
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("es connect error: %s", err.Error()))

	}

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("read res.body error: %s", err.Error()))

	}
	var s Search_Request
	json.Unmarshal(resString, &s)
	// // log.Println(res)
	defer res.Body.Close()

	// fmt.Println(resString)
	return s
}

// SearchRequest 使用預設 ES 客戶端執行查詢（向後兼容）
// dsl 比實際時間再減八小時
func SearchRequest(index string, field string, timefrom string, timeto string) Search_Request {
	return SearchRequestWithClient(global.Elasticsearch, index, field, timefrom, timeto)
}


type Search_Request struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore interface{}   `json:"max_score"`
		Hits     []interface{} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		Num2 struct {
			DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
			SumOtherDocCount        int `json:"sum_other_doc_count"`
			Buckets                 []struct {
				Key      string `json:"key"`
				DocCount int    `json:"doc_count"`
			} `json:"buckets"`
		} `json:"2"`
	} `json:"aggregations"`
}