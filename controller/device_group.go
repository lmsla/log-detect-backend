package controller

import (
	"net/http"
	"strconv"
	"log-detect/entities"
	"log-detect/services"

	"github.com/gin-gonic/gin"
)

// @Summary Create Device Group
// @Tags DeviceGroup
// @Accept  json
// @Produce  json
// @Param DeviceGroup body entities.DeviceGroup true "device group"
// @Success 200 {object} models.Response
// @Router /DeviceGroup/Create [post]
func CreateDeviceGroup(c *gin.Context) {
	body := new(entities.DeviceGroup)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := services.CreateDeviceGroup(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Msg})
		return
	}
	c.JSON(http.StatusOK, res.Body)
}

// @Summary Update Device Group
// @Tags DeviceGroup
// @Accept  json
// @Produce  json
// @Param DeviceGroup body entities.DeviceGroup true "device group"
// @Success 200 {object} models.Response
// @Router /DeviceGroup/Update [put]
func UpdateDeviceGroup(c *gin.Context) {
	body := new(entities.DeviceGroup)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := services.UpdateDeviceGroup(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Msg})
		return
	}
	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get All Device Groups
// @Tags DeviceGroup
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /DeviceGroup/GetAll [get]
func GetAllDeviceGroups(c *gin.Context) {
	res := services.GetAllDeviceGroups()

	if !res.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Msg})
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get Device Group by ID
// @Tags DeviceGroup
// @Accept  json
// @Produce  json
// @Param id path int true "group id"
// @Success 200 {object} models.Response
// @Router /DeviceGroup/Get/{id} [get]
func GetDeviceGroupByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := services.GetDeviceGroupByID(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Msg})
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Delete Device Group
// @Tags DeviceGroup
// @Accept  json
// @Produce  json
// @Param id path int true "group id"
// @Success 200 {object} models.Response
// @Router /DeviceGroup/Delete/{id} [delete]
func DeleteDeviceGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := services.DeleteDeviceGroup(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": res.Msg})
}

// @Summary Move Devices to Group
// @Tags DeviceGroup
// @Accept  json
// @Produce  json
// @Param request body object true "move devices request"
// @Success 200 {object} models.Response
// @Router /DeviceGroup/MoveDevices [post]
func MoveDevicesToGroup(c *gin.Context) {
	var body struct {
		DeviceIDs       []int  `json:"device_ids" binding:"required"`
		TargetGroupName string `json:"target_group_name" binding:"required"`
	}

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := services.MoveDevicesToGroup(body.DeviceIDs, body.TargetGroupName)

	if !res.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Msg})
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Move All Devices Between Groups
// @Tags DeviceGroup
// @Accept  json
// @Produce  json
// @Param request body object true "move group devices request"
// @Success 200 {object} models.Response
// @Router /DeviceGroup/MoveGroupDevices [post]
func MoveGroupDevices(c *gin.Context) {
	var body struct {
		SourceGroupName string `json:"source_group_name" binding:"required"`
		TargetGroupName string `json:"target_group_name" binding:"required"`
	}

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := services.MoveGroupDevices(body.SourceGroupName, body.TargetGroupName)

	if !res.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Msg})
		return
	}

	c.JSON(http.StatusOK, res.Body)
}
