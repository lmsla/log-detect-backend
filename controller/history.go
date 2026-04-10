package controller

import (
	"net/http"
	"strconv"
	"log-detect/services"

	"github.com/gin-gonic/gin"
)


// @Summary Get History Data
// @Tags History
// @Accept  json
// @Produce  json
// @Param logname path string true "Logname"
// @Param hours query int false "Time range in hours (1, 3, 6). Omit or set 0 for full day." Enums(0, 1, 3, 6)
// @Success 200 {object} string
// @Router /History/GetData/{logname} [GET]
func GetHistoryData(c *gin.Context) {

	logname := c.Param("logname")
	if logname == "" {
		c.JSON(http.StatusBadRequest, "Missing logname parameter")
		return
	}

	// 解析 hours 查詢參數，預設為 0（全天）
	hours := 0
	if hoursStr := c.Query("hours"); hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h >= 0 {
			hours = h
		}
	}

	res := services.DataDealing(logname, hours)
	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Body)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}



// @Summary Get History Data
// @Tags History
// @Accept  json
// @Produce  json
// @Success 200 {object} string
// @Router /History/GetLognameData [GET]


// @Summary Get Get Logname in History
// @Tags History
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /History/GetLognameData [get]
func GetLognameData(c *gin.Context) {

	res := services.GetLognameData()

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}