package clients

import (
	"github.com/elastic/go-elasticsearch/v8"
	"log-detect/global"
	"log-detect/log"
	"crypto/tls"
	"fmt"
	"net/http"
)

var ES *elasticsearch.Client

func SetElkClient() {
	var err error
	cfg := elasticsearch.Config{
		Addresses: global.EnvConfig.ES.URL,
		Username:  global.EnvConfig.ES.SourceAccount,
		Password:  global.EnvConfig.ES.SourcePassword,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	ES, err = elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("ES Cluster 連線失敗")
		log.Logrecord("Elasticsearch ", "ES Cluster 連線失敗")
		
		panic(err) // 連線失敗
		
	}

	res, err := ES.Info()
	if err != nil {
		log.Logrecord("Elasticsearch ", fmt.Sprintf("Error getting response: %s", err))
		//   log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()
	// fmt.Println(res)
	// log.SetFlags(0)

}