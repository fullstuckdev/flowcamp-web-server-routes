package controllers

import (
	"fmt"
	"golangapi/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PostController struct {
	DB *gorm.DB
}

func NewPostController(db *gorm.DB) *PostController {
	return &PostController{DB: db}
}

func (pc *PostController) CreateTag(c *gin.Context) {
	var tag models.Tag

	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := pc.DB.Create(&tag).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
	}

	c.JSON(201, gin.H{"data": models.TagResponse{
		ID: tag.ID,
		Name: tag.Name,
	}})
}

func (pc *PostController) CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userId.(uint),
	}

	tx := pc.DB.Begin()

	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// pengecekan tag
	if len(req.TagIds) > 0 {
		var tags []models.Tag

		if err := tx.Find(&tags, req.TagIds).Error; err != nil {
			tx.Rollback()
			c.JSON(400, gin.H{"error": "invalid tag IDs"})
			return
		}

		if len(tags) != len(req.TagIds) {
			tx.Rollback()
			c.JSON(400, gin.H{"error": "Beberapa tag tidak ditemukan..."})
			return
		}

		if err := tx.Model(&post).Association("Tags").Append(&tags); err != nil {
			tx.Rollback()
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
	}

	tx.Commit()

	if err := pc.DB.Preload("User").Preload("Tags").First(&post, post.ID).Error; err != nil {
		c.JSON(400, gin.H{"error": "Error loading post data"})
		return
	} 

	c.JSON(201, gin.H{"data": post.ToResponse()})
}

func (pc *PostController) GetPosts(c *gin.Context) {
	var posts []models.Post

	if err := pc.DB.Preload("User").Preload("Tags").Find(&posts).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	response := make([]models.PostResponse, len(posts)) 
	for i, post := range posts {
		response[i] = models.PostResponse{
			ID: post.ID,
			Title: post.Title,
			Content: post.Content,
			CreatedAt: post.CreatedAt,
			Author: models.UserResponse{
				ID: post.User.ID,
				Name: post.User.Name,
				Email: post.User.Email,
			},
			Tags: make([]models.TagResponse, len(post.Tags)),
		}

		for j, tag := range post.Tags {
			response[i].Tags[j] = models.TagResponse{
				ID: tag.ID,
				Name: tag.Name,
			}
		}
	}

	c.JSON(200, gin.H{"data": response})
}

func (pc *PostController) GetPost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	if err := pc.DB.Preload("User").Preload("Tags").First(&post, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(200, gin.H{"data": post})
}

// Simulate database error after certain amount
func simulateError(amount float64) error {
	if amount == 999.99 {
		return fmt.Errorf("simulated database error")
	}
	return nil
}

func (pc *PostController) CreateTransactionWithTx(c *gin.Context) {
	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	transaction.UserID = userId.(uint)

	tx := pc.DB.Begin()

	var user models.User
	if err := tx.First(&user, userId).Error; err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	if transaction.Type == "credit" {
		user.Balance += transaction.Amount
	} else if transaction.Type == "debit" {
		if user.Balance < transaction.Amount {
			tx.Rollback()
			c.JSON(400, gin.H{"error": "Insufficient balance"})
			return
		}
		user.Balance -= transaction.Amount
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := simulateError(transaction.Amount); err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{
			"error": err.Error(),
			"status": "All changes rolled back. Transaction cancelled and balance unchanged.",
			"current_balance": user.Balance - transaction.Amount,
		})
		return
	}

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(201, gin.H{"data": transaction})
}

func (pc *PostController) CreateTransactionWithoutTx(c *gin.Context) {
	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	transaction.UserID = userId.(uint)

	// Get user for balance update
	var user models.User
	if err := pc.DB.First(&user, userId).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	originalBalance := user.Balance

	// Update balance based on transaction type
	if transaction.Type == "credit" {
		user.Balance += transaction.Amount
	} else if transaction.Type == "debit" {
		if user.Balance < transaction.Amount {
			c.JSON(400, gin.H{"error": "Insufficient balance"})
			return
		}
		user.Balance -= transaction.Amount
	}

	// Save transaction
	if err := pc.DB.Create(&transaction).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Simulate error after transaction is created
	if err := simulateError(transaction.Amount); err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
			"status": "Transaction was created but balance update failed. DATA INCONSISTENCY!",
			"transaction_id": transaction.ID,
			"original_balance": originalBalance,
			"failed_balance_update": user.Balance,
			"warning": "Database is now in an inconsistent state!",
		})
		return
	}

	// Update user balance
	if err := pc.DB.Save(&user).Error; err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
			"status": "Transaction was created but balance update failed. DATA INCONSISTENCY!",
			"transaction_id": transaction.ID,
			"warning": "Database is now in an inconsistent state!",
		})
		return
	}

	c.JSON(201, gin.H{"data": transaction})
}