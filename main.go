package main

import (
	"golangapi/config"
	"golangapi/controllers"
	"golangapi/middleware"
	"golangapi/models"
	"log"

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

	db.AutoMigrate(&models.User{}, &models.Transaction{}, &models.Post{}, &models.Tag{}, &models.PostTag{})

	authController := controllers.NewAuthController(db)
	userController := controllers.NewUserController(db)
	postController := controllers.NewPostController(db)

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
			// User routes
			protected.GET("/users", userController.GetUsers)
			protected.POST("/users", userController.CreateUser)

			// Tag routes
			protected.POST("/tags", postController.CreateTag)

			// Without DB routes
			protected.POST("/send", controllers.CreateUserWithoutDB)
			protected.GET("/get", controllers.GetUserWithoutDB)

			// Post Routes
			protected.POST("/posts", postController.CreatePost)
			protected.GET("/posts", postController.GetPosts)
			protected.GET("/posts/:id", postController.GetPost)

			// Transaction Routes
			transactions := protected.Group("/transactions")
			{
				// With Transaction (ACID)
				transactions.POST("/", postController.CreateTransactionWithTx)
				
				// Without Transaction (Demonstration purposes)
				transactions.POST("/no-tx", postController.CreateTransactionWithoutTx)
			
			}

		}
	}

	r.Run(":8080")
}