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

// @Summary Get Receiver
// @Tags Receiver
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /Receiver/GetAll [get]
func GetAllReceivers(c *gin.Context) {

	res := services.GetAllReceivers()

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}



// @Summary Create Receiver
// @Tags Receiver
// @Accept  json
// @Produce  json
// @Param Receiver body entities.Receiver true "receiver"
// @Success 200 {object} models.Response
// @Router /Receiver/Create [post]
func CreateReceiver(c *gin.Context) {

	// body := new(models.Instance)
	body := new(entities.Receiver)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.CreateReceiver(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}


// @Summary Update Receiver
// @Tags Receiver
// @Accept  json
// @Produce  json
// @Param Instacne body entities.Receiver true "receiver"
// @Success 200 {object} string
// @Router /Receiver/Update [put]
func UpdateReceiver(c *gin.Context) {

	body := new(entities.Receiver)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.UpdateReceiver(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}

// @Summary Delete Receiver
// @Tags Receiver
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} string
// @Router /Receiver/Delete/{id} [delete]
func DeleteReceiver(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.DeleteReceiver(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Msg)
}