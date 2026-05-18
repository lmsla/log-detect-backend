package global

import (
	"database/sql"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"log-detect/structs"
	"sync"
)

var (
	EnvConfig     *structs.EnviromentModel
	Elasticsearch *elasticsearch.Client
	TargetStruct  *structs.TargetStruct
	YMLConfig     *structs.YMLConfig // 擴充格式的 config.yml（用於 YML-to-DB 同步）
	Mysql         *gorm.DB
	Crontab       *cron.Cron

	// TimescaleDB 相關
	TimescaleDB *sql.DB         // TimescaleDB 原生連接
	BatchWriter BatchWriterType // 批量寫入服務
	ConfigMu    sync.RWMutex
)

func GetYMLConfig() *structs.YMLConfig {
	ConfigMu.RLock()
	defer ConfigMu.RUnlock()
	return YMLConfig
}

func SetYMLConfig(cfg *structs.YMLConfig) {
	ConfigMu.Lock()
	defer ConfigMu.Unlock()
	YMLConfig = cfg
}

// BatchWriterType 將在 services/batch_writer.go 中定義
type BatchWriterType interface {
	AddHistory(history any) error
	Stop()
}
