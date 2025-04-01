package entities

import (
	"log-detect/models"
)

type Target struct {
	models.Common
	ID      int    `gorm:"primaryKey;index" json:"id" form:"id"`
	Subject string `gorm:"type:varchar(50)" json:"subject" form:"subject"`
	To      to     `gorm:"serializer:json"  json:"to" form:"to"`
	Enable  bool
	// Indices []Index `gorm:"foreignKey:TargetID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`
	Indices []Index `gorm:"many2many:indices_targets;foreignKey:ID;reference:ID;"`
}

type Receiver struct {
	models.Common
	ID   int  `gorm:"primaryKey;index" json:"id" form:"id"`
	Name Name `gorm:"serializer:json"`
}

type IndicesTargets struct {
	// models.Common
	TargetID int `gorm:"primaryKey" form:"target_id"`
	IndexID  int `gorm:"primaryKey" form:"index_id"`
}

type Name []string
type to []string

type Index struct {
	models.Common
	ID int `gorm:"primaryKey;index" json:"id" form:"id"`
	// TargetID    int    `gorm:"index" json:"target_id" form:"target_id"`
	Targets     []Target `gorm:"many2many:indices_targets;"`
	Pattern     string   `gorm:"type:varchar(50)" json:"pattern" form:"pattern"`
	DeviceGroup string   `gorm:"type:varchar(50)" json:"device_group" form:"device_group"`
	Logname     string   `gorm:"type:varchar(50)" json:"logname" form:"logname"`
	Period      string   `gorm:"type:varchar(50)" json:"period" form:"period"`
	Unit        int      `type:"int" json:"unit" form:"unit"`
	Field       string   `gorm:"type:varchar(50)" json:"field" form:"field"`
	Enable      bool
}

type Device struct {
	models.Common
	ID          int    `gorm:"primaryKey;index" json:"id" form:"id"`
	DeviceGroup string `gorm:"type:varchar(50)" json:"device_group" form:"device_group"`
	Name        string `gorm:"type:varchar(50)" json:"name" form:"name"`
}

type CronList struct {
	models.Common
	EntryID  int `gorm:"index" json:"entry_id" form:"entry_id"`
	TargetID int `gorm:"index" json:"target_id" form:"target_id"`
	IndexID  int `gorm:"index" json:"index_id" form:"index_id"`
}

type Table_counts struct {
	DeviceGroup  string `json:"device_group" form:"device_group"`
	DevicesCount int64  `json:"devices_count" form:"devices_count"`
}

type GroupName struct {
	DeviceGroup string `json:"device_group" form:"device_group"`
}

type History struct {
	models.Common
	Logname     string `gorm:"type:varchar(50)" json:"logname" form:"logname"`
	DeviceGroup string `gorm:"type:varchar(50)" json:"device_group" form:"device_group"`
	Name        string `gorm:"type:varchar(50)" json:"name" form:"name"`
	Lost        string `gorm:"type:varchar(50)" json:"lost" form:"lost"`
	LostNum     int    `type:"int" json:"lost_num" form:"lost_num"`
	Date        string `gorm:"type:varchar(50)" json:"date" form:"date"`
	Time        string `gorm:"type:varchar(50)" json:"time" form:"time"`
	DateTime    string `gorm:"type:varchar(50)" json:"date_time" form:"date_time"`
	Period      string `gorm:"type:varchar(50)" json:"period" form:"period"`
	Unit        int    `type:"int" json:"unit" form:"unit"`
}

type Logname struct {
	Logname string `json:"logname" form:"logname"`
}

type HistoryData struct {
	Name string `json:"name" form:"name"`
	Time string `json:"time" form:"time"`
	Lost string `json:"lost" form:"lost"`
}

type LognameCheck struct {
	Name string `json:"name" form:"name"`
	Lost string `json:"lost" form:"lost"`
}

type MailHistory struct {
	models.Common
	Date    string `gorm:"type:varchar(50)" json:"date" form:"date"`
	Time    string `gorm:"type:varchar(50)" json:"time" form:"time"`
	Logname string `gorm:"type:varchar(50)" json:"logname" form:"logname"`
	Sended  bool   `gorm:"type:boolean" json:"sended" form:"sended"`
}
