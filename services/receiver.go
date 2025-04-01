package services

import (
	// "fmt"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/models"
	"log-detect/log"
)



func GetAllReceivers() models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Receiver{}

	err := global.Mysql.Find(&res.Body).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Get All Receivers error: %s", err.Error()))
		res.Msg = fmt.Sprintf("Get All Receivers error: %s", err.Error())
		return res
	}
	res.Success = true
	res.Msg = "Get All Receivers Success"
	return res
}


// 新增receiver
func CreateReceiver(receiver entities.Receiver) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Receiver{}
	fmt.Println(receiver.Name)
	// result := global.Mysql.Where("name = ?", receiver.Name).First(&entities.Receiver{}) 
	// if result.RowsAffected > 0 {
	// 	res.Msg = "receiver Name already existed"
	// 	return res
	// }

	err := global.Mysql.Create(&receiver).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Create receiver Fail: %s", err.Error()))
		res.Msg = "Create receiver Fail"
		return res
	}
	res.Success = true
	res.Body = receiver
	res.Msg = "Create receiver Success"
	// global.Mysql.Where("name = ?", receiver.Name).First(&res.Body)
	return res
}


func UpdateReceiver(receiver entities.Receiver) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = []entities.Receiver{}


	err := global.Mysql.Select("*").Where("id = ?", receiver.ID).Updates(&receiver).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Update receiver Fail: %s", err.Error()))
		res.Msg = "Update receiver Fail"
		return res
	}

	res.Success = true
	res.Body = receiver
	res.Msg = "Update receiver Success"
	// global.Mysql.Where("id = ?", receiver.ID).First(&res.Body)

	return res

}


func DeleteReceiver(id int) models.Response {

	res := models.Response{}
	res.Success = false
	res.Body = nil

	result := global.Mysql.Where("id = ?", id).First(&entities.Receiver{})
	if result.RowsAffected == 0 {
		res.Msg = "Receiver ID does not exist"
		return res
	}


	err := global.Mysql.Where("id = ?", id).Delete(&entities.Receiver{}).Error
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Error when deleting receiver: %s", err.Error()))
		res.Msg = fmt.Sprintf("Error when deleting receiver: %s", err)
		return res
	}

	res.Success = true
	res.Msg = "Delete Success"

	return res

}