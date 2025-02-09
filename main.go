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

	db.AutoMigrate(&models.User{}, &models.Post{}, &models.Tag{}, &models.PostTag{})

	authController := controllers.NewAuthController(db)
	userController := controllers.NewUserController(db)
	postController := controllers.NewPostController(db)
	sysController := controllers.NewSysController(db)


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
			 protected.POST("/users", userController.CreateUser)

			 // Tag routes
			 protected.POST("/tags", postController.CreateTag)

			 // Without DB routes
			 protected.POST("/send", controllers.CreateUserWithoutDB)
			 protected.GET("/get", controllers.GetUserWithoutDB)

			 // Post Routes
			 protected.POST("/post", postController.CreatePost)
			 protected.GET("/post", postController.GetPosts)
			 protected.GET("/posts/:id", postController.GetPost)
			 protected.PUT("/posts/:id", postController.UpdatePost)
			 protected.DELETE("/posts/:id", postController.DeletePost)

			 // SYS Routes
			 protected.POST("/directory", sysController.CreateDirectory)
			 protected.POST("/file", sysController.CreateFile)
			 protected.POST("/file/read", sysController.ReadFile)

			 protected.PUT("/file/rename", sysController.RenameFile)

			 protected.POST("/file/upload", sysController.UploadFile)

			 protected.GET("/file/download", sysController.DownloadFile)

		 }
	}

	r.Run(":8080")
}