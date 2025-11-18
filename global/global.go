package global

import (
	"database/sql"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"log-detect/structs"
)

var (
	EnvConfig     *structs.EnviromentModel
	Elasticsearch *elasticsearch.Client
	TargetStruct  *structs.TargetStruct
	Mysql         *gorm.DB
	Crontab       *cron.Cron

	// TimescaleDB 相關
	TimescaleDB *sql.DB         // TimescaleDB 原生連接
	BatchWriter BatchWriterType // 批量寫入服務
)

// BatchWriterType 將在 services/batch_writer.go 中定義
type BatchWriterType interface {
	AddHistory(history any) error
	Stop()
}
