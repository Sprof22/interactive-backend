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
	initializers.MigrateDB()

}

func main() {

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"}, // Include "Authorization"
	}))
	r.POST("/createUser", controllers.CreateUser)
	r.POST("/get_github_access_token", controllers.GetGithubAccessTokenGin)
	// r.POST("/exchangeToken", controllers.ExchangeToken)
	r.PUT("/updateUser", controllers.UpdateUser)
	r.POST("/createProject", controllers.CreateProject)
	r.GET("/fetchProject", controllers.FetchUserProjects)
	r.GET("/projects", controllers.FetchAllProjects)
	r.DELETE("/projects/:id", controllers.DeleteProject)
	r.Run() // listen and serve on 0.0.0.0:8080
}
