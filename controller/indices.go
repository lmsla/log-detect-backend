package controller

import (
	"log-detect/entities"
	"log-detect/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary Get Indices
// @Tags Indices
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /Indices/GetAll [get]
func GetAllIndices(c *gin.Context) {

	res := services.GetAllIndices()

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Create Indices
// @Tags Indices
// @Accept  json
// @Produce  json
// @Param Indices body entities.Index true "device"
// @Success 200 {object} models.Response
// @Router /Indices/Create [post]
func CreateIndices(c *gin.Context) {

	body := new(entities.Index)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.CreateIndices(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}

// @Summary Update Indices
// @Tags Indices
// @Accept  json
// @Produce  json
// @Param Indices body entities.Index true "device"
// @Success 200 {object} string
// @Router /Indices/Update [put]
func UpdateIndices(c *gin.Context) {

	body := new(entities.Index)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.UpdateIndices(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}

// @Summary Delete Indices
// @Tags Indices
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} string
// @Router /Indices/Delete/{id} [delete]
func DeleteIndices(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.DeleteIndice(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Msg)
}

// @Summary Get Indices by Target ID
// @Tags Indices
// @Accept  json
// @Produce  json
// @Param id path int true "target id"
// @Success 200 {object} entities.Index
// @Router /Indices/GetIndicesByTargetID/{id} [get]
func GetIndicesByTargetID(c *gin.Context) {

	TargetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "target ID should be int")
		// handler.WriteErrorLog(c, "instance ID should be integer")
		return
	}

	inventory, err := services.GetIndicesByTargetID(TargetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, "error to get inventory details")
		// handler.WriteErrorLog(c, "error to get inventory details")
		return
	}
	c.JSON(http.StatusOK, inventory)

}

// @Summary Get Logname
// @Tags Indices
// @Accept  json
// @Produce  json
// @Success 200 {object} string
// @Router /Indices/GetLogname [GET]
func GetLogname(c *gin.Context) {

	res := services.GetLogname()
	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Body)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get Indices Data
// @Tags Indices
// @Accept  json
// @Produce  json
// @Param logname path string true "Logname"
// @Success 200 {object} string
// @Router /Indices/GetIndicesByLogname/{logname} [GET]
func GetIndiceData(c *gin.Context) {

	logname := c.Param("logname")
	if logname == "" {
		c.JSON(http.StatusBadRequest, "Missing logname parameter")
		return
	}

	res := services.GetIndicesData(logname)
	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Body)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}
