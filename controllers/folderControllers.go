package controllers

import (
	"interactive-path/initializers"
	"interactive-path/models"

	"github.com/gin-gonic/gin"
)

func CreateFolder(c *gin.Context) {
	var body struct {
		Name string
	}

	c.Bind(&body)
	folder := models.Folder{Name: body.Name}

	result := initializers.DB.Create(&folder)

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"folder": folder,
	})
}
