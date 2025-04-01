package entities

import (
	"log-detect/models"
)



type MainMenu struct {
	ID         int    `json:"id"`
	RouterPath string `json:"router_path"`
	Icon       string `json:"icon"`
	Title      string `json:"title"`
	Sort       int    `json:"sort"`
	Module     string `json:"module"`
	OnlyAdmin  bool   `json:"-"`
	// SubMenu []SubMenu `json:"children,omitempty" gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}


type Module struct {
	models.Common
	ID         int    `json:"id" gorm:"primaryKey"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	LicenseKey string `json:"-" `
	Disabled   bool   `json:"disabled"`
}
