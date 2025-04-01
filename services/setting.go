package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	// "log-detect/models"
	"gorm.io/gorm/clause"
)

func GetServerModule() []entities.Module {
	db := global.Mysql
	modules := []entities.Module{}
	err := db.Model(entities.Module{}).Where(&entities.Module{
		Disabled: false}).Find(&modules).Error
	if err != nil {
		// global.Logger.Error(
		// 	err.Error(),
		// 	zap.String(global.LogEvent.Tag.Service, global.LogEvent.API.Enviroment),
		// )
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get Server Module error: %s", err.Error()))

	}
	return modules
}

func GetServerMenu() []entities.MainMenu {
	db := global.Mysql
	menus := []entities.MainMenu{}
	// var err error

	var err = db.Model(entities.MainMenu{}).Preload(clause.Associations).Find(&menus).Error

	// switch role {
	// case global.EnvConfig.SSO.AdminRole:

	// 	err = db.Model(entities.MainMenu{}).Preload(clause.Associations).Find(&menus).Error
	// default:

	// 	err = db.Model(entities.MainMenu{}).Preload(clause.Associations).Not(
	// 		&entities.MainMenu{
	// 			OnlyAdmin: true}).Find(&menus).Error
	// }
	if err != nil {
		// global.Logger.Error(
		// 	err.Error(),
		// 	zap.String(global.LogEvent.Tag.Service, global.LogEvent.API.Enviroment),
		// )
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get Server Menu error: %s", err.Error()))
	}
	for _, data := range menus {
		fmt.Println(data.Icon)
	}
	fmt.Println(menus)
	return menus
}
