package models




type Index struct {
	Common
	ID int `gorm:"primaryKey;index" json:"id" form:"id"`
	// TargetID    int    `gorm:"index" json:"target_id" form:"target_id"`
	Pattern     string `gorm:"type:varchar(50)" json:"pattern" form:"pattern"`
	DeviceGroup string `gorm:"type:varchar(50)" json:"device_group" form:"device_group"`
	Logname     string `gorm:"type:varchar(50)" json:"logname" form:"logname"`
	Period      string `gorm:"type:varchar(50)" json:"period" form:"period"`
	Unit        int    `type:"int" json:"unit" form:"unit"`
	Field       string `gorm:"type:varchar(50)" json:"field" form:"field"`
}