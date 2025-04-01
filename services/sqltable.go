package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/global"
)


func CreateTable() {
	global.Mysql.Exec("USE logdetect")
	err := global.Mysql.AutoMigrate(&entities.Device{},&entities.Receiver{},&entities.Index{},&entities.Target{})
	if err != nil {
		fmt.Println("error")
	}
}