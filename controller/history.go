package controller

import (
	"net/http"
	// "strconv"
	// "log-detect/handler"
	// "log-detect/models"
	// "log-detect/entities"
	"log-detect/services"

	"github.com/gin-gonic/gin"
)


// @Summary Get History Data
// @Tags History
// @Accept  json
// @Produce  json
// @Param logname path string true "Logname"
// @Success 200 {object} string
// @Router /History/GetData/{logname} [GET]
func GetHistoryData(c *gin.Context) {


	logname := c.Param("logname")
	if logname == "" {
		c.JSON(http.StatusBadRequest, "Missing logname parameter")
		return
	}


	res := services.DataDealing(logname)
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