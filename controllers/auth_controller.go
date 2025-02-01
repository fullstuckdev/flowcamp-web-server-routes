package controllers

import(
	"golangapi/models"
	"golangapi/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

func (ac *AuthController) Register(c *gin.Context) {
    var user models.User
	
	// harus JSON
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

	// HASH Password
    if err := user.HashPassword(user.Password); err != nil {
        c.JSON(500, gin.H{"error": "Error hashing password"})
        return
    }

	// Create ke DB
    result := ac.DB.Create(&user)
    if result.Error != nil {
        c.JSON(400, gin.H{"error": "Error creating user"})
        return
    }

	// Generate Token
    token, err := utils.GenerateToken(user.ID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Error generating token"})
        return
    }

    c.JSON(201, gin.H{
        "message": "User registered successfully",
        "token":   token,
    })
}

func (ac *AuthController) Login(c *gin.Context) {
    var loginReq models.LoginRequest

	// harus bentuknya JSON
    if err := c.ShouldBindJSON(&loginReq); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    var user models.User

	// ngecheck datanya ada atau ga?
    if err := ac.DB.Where("email = ?", loginReq.Email).First(&user).Error; err != nil {
        c.JSON(401, gin.H{"error": "Invalid email or password"})
        return
    }

	// check password ke si JWT
    if err := user.CheckPassword(loginReq.Password); err != nil {
        c.JSON(401, gin.H{"error": "Invalid email or password"})
        return
    }


	// Generate Token
    token, err := utils.GenerateToken(user.ID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Error generating token"})
        return
    }

    c.JSON(200, gin.H{
        "message": "Login successful",
        "token":   token,
    })
}