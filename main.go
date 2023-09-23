package main

import (
	"interactive-path/controllers"
	"interactive-path/initializers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}

func main() {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"} // Add your frontend URL
	r.Use(cors.New(config))

	r.POST("/projects", controllers.CreateProject)
	r.GET("/projects", controllers.GetProjects)
	r.GET("/projects/:id", controllers.GetSingleProject)
	r.PUT("/projects/:id", controllers.UpdateProject)
	r.DELETE("/projects/:id", controllers.DeleteProject)

	r.POST("/files", controllers.CreateProject)
	r.GET("/files", controllers.GetProjects)
	r.GET("/files/:id", controllers.GetSingleProject)
	r.PUT("/files/:id", controllers.UpdateProject)
	r.DELETE("/files/:id", controllers.DeleteProject)

	r.POST("/folders", controllers.CreateFolder)
	// r.POST("/folders/:id/files", controllers.CreateFile)
	r.Run() // listen and serve on 0.0.0.0:8080
}
