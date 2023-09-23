package main

import (
	"interactive-path/initializers"
	"interactive-path/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}

func main() {
	initializers.DB.AutoMigrate(&models.Project{}, &models.Folder{}, &models.File{})
}
