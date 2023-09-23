package controllers

import (
	"interactive-path/initializers"
	"interactive-path/models"

	"github.com/gin-gonic/gin"
)

func CreateProject(c *gin.Context) {

	var body struct {
		Title       string
		Description string
	}

	c.Bind(&body)
	project := models.Project{Title: body.Title, Description: body.Description}

	result := initializers.DB.Create(&project)

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"project": project,
	})
}

func GetSingleProject(c *gin.Context) {

	id := c.Param("id")
	var project models.Project
	initializers.DB.First(&project, id)

	c.JSON(200, gin.H{
		"project": project,
	})
}

func GetProjects(c *gin.Context) {
	var project []models.Project
	result := initializers.DB.Order("id").Find(&project) // Order by ID ascending

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"project": project,
	})
}

func UpdateProject(c *gin.Context) {
	var body struct {
		Title       string
		Description string
	}

	id := c.Param("id")
	var project models.Project
	initializers.DB.First(&project, id)
	c.Bind(&body)

	initializers.DB.Model(&project).Updates(models.Project{Title: body.Title, Description: body.Description})

	c.JSON(200, gin.H{

		"project": project,
	})

}

func DeleteProject(c *gin.Context) {
	id := c.Param("id")
	var project models.Project
	initializers.DB.Delete(&models.Project{}, id)

	c.JSON(200, gin.H{
		"project": project,
	})
}
