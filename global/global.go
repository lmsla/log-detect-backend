package global

import (
	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
	"log-detect/structs"
	"github.com/robfig/cron/v3"
)

var (
	EnvConfig     *structs.EnviromentModel
	Elasticsearch *elasticsearch.Client
	TargetStruct  *structs.TargetStruct
	Mysql         *gorm.DB
	Crontab       *cron.Cron
)
