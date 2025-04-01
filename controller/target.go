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



// @Summary Get Target
// @Tags Target
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Router /Target/GetAll [get]
func GetAllTargets(c *gin.Context) {

	res := services.GetAllTargets()

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}

	c.JSON(http.StatusOK, res.Body)
}




// @Summary Create Target
// @Tags Target
// @Accept  json
// @Produce  json
// @Param Target body entities.Target true "target"
// @Success 200 {object} models.Response
// @Router /Target/Create [post]
func CreateTarget(c *gin.Context) {

	// body := new(models.Instance)
	body := new(entities.Target)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.CreateTarget(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}


// @Summary Update Target
// @Tags Target
// @Accept  json
// @Produce  json
// @Param Target body entities.Target true "target"
// @Success 200 {object} string
// @Router /Target/Update [put]
func UpdateTarget(c *gin.Context) {

	body := new(entities.Target)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.UpdateTarget(*body)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Body)
}


// @Summary Delete Target
// @Tags Target
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} string
// @Router /Target/Delete/{id} [delete]
func DeleteTarget(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	res := services.DeleteTarget(id)

	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Msg)
}