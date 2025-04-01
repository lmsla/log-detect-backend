package controller

import (
	"net/http"
	"strconv"
	// "log-detect/handler"
	// "log-detect/models"
	"log-detect/entities"
	"log-detect/services"

	"github.com/gin-gonic/gin"
)



// @Summary Create Device
// @Tags Device
// @Accept  json
// @Produce  json
// @Param Device body []entities.Device true "device"
// @Success 200 {object} models.Response
// @Router /Device/Create [post]
func CreateDevice(c *gin.Context) {

	body := new([]entities.Device)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.CreateDevice(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}



// @Summary Update Device
// @Tags Device
// @Accept  json
// @Produce  json
// @Param Device body entities.Device true "device"
// @Success 200 {object} string
// @Router /Receiver/Update [put]
func UpdateDevice(c *gin.Context) {

	body := new(entities.Device)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.UpdateDevice(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}


// @Summary Get Device
// @Tags Device
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /Device/GetAll [get]
func GetAllDevices(c *gin.Context) {

	res := services.GetAllDevices()

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}


// @Summary Delete Device
// @Tags Device
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} string
// @Router /Device/Delete/{id} [delete]
func DeleteDevice(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.DeleteDevice(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Msg)
}


// @Summary Count Device
// @Tags Device
// @Accept  json
// @Produce  json
// @Success 200 {object} string
// @Router /Device/count [GET]
func GetTableCounts(c *gin.Context) {

	res := services.CountDevices()
	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Body)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}


// @Summary Get Group
// @Tags Device
// @Accept  json
// @Produce  json
// @Success 200 {object} string
// @Router /Device/GetGroup [GET]
func GetDeviceGroup(c *gin.Context) {

	res := services.GetDeviceGroup()
	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Body)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}