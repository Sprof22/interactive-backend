package initializers

import "interactive-path/models"

func MigrateDB() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Project{})
}
