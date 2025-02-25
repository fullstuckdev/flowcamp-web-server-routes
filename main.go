package main

import (
	"golangapi/config"
	"golangapi/controllers"
	"golangapi/models"
	"log"
	"golangapi/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error Loading ENV")
	}

	r := gin.Default()
	db := config.ConnectDatabase()

    db.AutoMigrate(&models.User{}, &models.Profile{}, &models.Post{}, &models.Group{})

	authController := controllers.NewAuthController(db)
	userController := controllers.NewUserController(db)
    relationController := controllers.NewRelationController(db)


	api := r.Group("/api")
	{
		auth := api.Group("/auth") 
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
		}

		 // Protected routes
		 protected := api.Group("/")
		 protected.Use(middleware.AuthMiddleware())
		 {
			 protected.GET("/users", userController.GetUsers)
			 protected.POST("/profiles", relationController.CreateProfile)
			 protected.POST("/posts", relationController.CreatePost)
			 protected.POST("/groups", relationController.CreateGroup)
			 protected.POST("/groups/add-user", relationController.AddUserToGroup)
			 protected.GET("/users/:id/relations", relationController.GetUserWithRelations)
		 }
	}

	r.Run(":8080")
}