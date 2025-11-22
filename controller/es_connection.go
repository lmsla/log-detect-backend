package controller

import (
	"net/http"
	"strconv"

	"log-detect/entities"
	"log-detect/services"

	"github.com/gin-gonic/gin"
)

// @Summary Get All ES Connections
// @Tags ESConnection
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /ESConnection/GetAll [get]
func GetAllESConnections(c *gin.Context) {
	res := services.GetAllESConnections()

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Get ES Connection
// @Tags ESConnection
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} models.Response
// @Router /ESConnection/Get/{id} [get]
func GetESConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.GetESConnection(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Create ES Connection
// @Tags ESConnection
// @Accept  json
// @Produce  json
// @Param ESConnection body entities.ESConnection true "es_connection"
// @Success 200 {object} models.Response
// @Router /ESConnection/Create [post]
func CreateESConnection(c *gin.Context) {
	body := new(entities.ESConnection)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.CreateESConnection(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Update ES Connection
// @Tags ESConnection
// @Accept  json
// @Produce  json
// @Param ESConnection body entities.ESConnection true "es_connection"
// @Success 200 {object} models.Response
// @Router /ESConnection/Update [put]
func UpdateESConnection(c *gin.Context) {
	body := new(entities.ESConnection)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.UpdateESConnection(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Delete ES Connection
// @Tags ESConnection
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} string
// @Router /ESConnection/Delete/{id} [delete]
func DeleteESConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.DeleteESConnection(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Msg)
}

// @Summary Test ES Connection
// @Tags ESConnection
// @Accept  json
// @Produce  json
// @Param ESConnection body entities.ESConnection true "es_connection"
// @Success 200 {object} models.Response
// @Router /ESConnection/Test [post]
func TestESConnection(c *gin.Context) {
	body := new(entities.ESConnection)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.TestESConnection(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}

// @Summary Set Default ES Connection
// @Tags ESConnection
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} string
// @Router /ESConnection/SetDefault/{id} [put]
func SetDefaultESConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.SetDefaultESConnection(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Msg)
}

// @Summary Reload ES Connection
// @Tags ESConnection
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} string
// @Router /ESConnection/Reload/{id} [put]
func ReloadESConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.ReloadESConnection(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Msg)
}

// @Summary Reload All ES Connections
// @Tags ESConnection
// @Accept  json
// @Produce  json
// @Success 200 {object} string
// @Router /ESConnection/ReloadAll [put]
func ReloadAllESConnections(c *gin.Context) {
	res := services.ReloadAllESConnections()

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Msg)
}
