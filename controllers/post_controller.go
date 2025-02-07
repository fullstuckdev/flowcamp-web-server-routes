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

	// Reload the transaction with user data
	if err := pc.DB.Preload("User").First(&transaction, transaction.ID).Error; err != nil {
		c.JSON(400, gin.H{"error": "Error loading created data"})
		return
	}

	c.JSON(201, gin.H{"data": transaction.ToResponse()})
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

// Update Post
func (pc *PostController) UpdatePost(c *gin.Context) {
	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var post models.Post
	// Check if post exists and belongs to user
	if err := pc.DB.Where("id = ? AND user_id = ?", c.Param("id"), userId).First(&post).Error; err != nil {
		c.JSON(404, gin.H{"error": "Post not found or unauthorized"})
		return
	}

	tx := pc.DB.Begin()

	// Update basic post info
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}

	if err := tx.Save(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Update tags if provided
	if len(req.TagIds) > 0 {
		var tags []models.Tag
		if err := tx.Find(&tags, req.TagIds).Error; err != nil {
			tx.Rollback()
			c.JSON(400, gin.H{"error": "Invalid tag IDs"})
			return
		}

		if len(tags) != len(req.TagIds) {
			tx.Rollback()
			c.JSON(400, gin.H{"error": "Some tags were not found"})
			return
		}

		// Replace existing tags
		if err := tx.Model(&post).Association("Tags").Replace(&tags); err != nil {
			tx.Rollback()
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
	}

	tx.Commit()

	// Reload post with associations
	if err := pc.DB.Preload("User").Preload("Tags").First(&post, post.ID).Error; err != nil {
		c.JSON(400, gin.H{"error": "Error loading updated post"})
		return
	}

	c.JSON(200, gin.H{"data": post.ToResponse()})
}

// Delete Post
func (pc *PostController) DeletePost(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var post models.Post
	// Check if post exists and belongs to user
	if err := pc.DB.Where("id = ? AND user_id = ?", c.Param("id"), userId).First(&post).Error; err != nil {
		c.JSON(404, gin.H{"error": "Post not found or unauthorized"})
		return
	}

	tx := pc.DB.Begin()

	// Clear associations first
	if err := tx.Model(&post).Association("Tags").Clear(); err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Delete the post
	if err := tx.Delete(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(200, gin.H{"data": "Post deleted successfully"})
}

// Update Transaction
func (pc *PostController) UpdateTransaction(c *gin.Context) {
	var req models.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	tx := pc.DB.Begin()

	var transaction models.Transaction
	if err := tx.Where("id = ? AND user_id = ?", c.Param("id"), userId).First(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"error": "Transaction not found or unauthorized"})
		return
	}

	// Get user for balance update
	var user models.User
	if err := tx.First(&user, userId).Error; err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Store original values
	currentBalance := user.Balance
	oldAmount := transaction.Amount
	oldType := transaction.Type

	// Calculate balance after reverting the old transaction
	var balanceAfterRevert float64
	if oldType == "credit" {
		balanceAfterRevert = currentBalance - oldAmount
	} else { // if it was a debit
		balanceAfterRevert = currentBalance + oldAmount
	}

	// Validate and calculate new balance
	var newBalance float64
	if req.Type == "debit" {
		if balanceAfterRevert < req.Amount {
			tx.Rollback()
			c.JSON(400, gin.H{"error": "Insufficient balance"})
			return
		}
		newBalance = balanceAfterRevert - req.Amount
	} else { // credit
		newBalance = balanceAfterRevert + req.Amount
	}

	// Update user's balance
	user.Balance = newBalance

	// Update transaction
	transaction.Amount = req.Amount
	transaction.Description = req.Description
	transaction.Type = req.Type

	// Save changes
	if err := tx.Save(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	// Reload the transaction with user data
	if err := pc.DB.Preload("User").First(&transaction, transaction.ID).Error; err != nil {
		c.JSON(400, gin.H{"error": "Error loading updated data"})
		return
	}

	c.JSON(200, gin.H{"data": transaction.ToResponse()})
}

// Delete Transaction
func (pc *PostController) DeleteTransaction(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	tx := pc.DB.Begin()

	var transaction models.Transaction
	// Use Unscoped() to ignore soft delete
	if err := tx.Unscoped().Where("id = ? AND user_id = ?", c.Param("id"), userId).First(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"error": "Transaction not found or unauthorized"})
		return
	}

	// Get user for balance update
	var user models.User
	if err := tx.First(&user, userId).Error; err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Revert transaction from balance
	if transaction.Type == "credit" {
		user.Balance -= transaction.Amount
	} else {
		user.Balance += transaction.Amount
	}

	// Save updated balance
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Perform hard delete using Unscoped()
	if err := tx.Unscoped().Delete(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(200, gin.H{
		"data": "Transaction permanently deleted",
		"transaction_id": transaction.ID,
		"updated_balance": user.Balance,
	})
}

// GetUserTransactions retrieves all transactions for the current user
func (pc *PostController) GetUserTransactions(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var transactions []models.Transaction
	if err := pc.DB.Where("user_id = ?", userId).Order("created_at desc").Find(&transactions).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Load user data for each transaction
	for i := range transactions {
		if err := pc.DB.Preload("User").First(&transactions[i], transactions[i].ID).Error; err != nil {
			c.JSON(400, gin.H{"error": "Error loading transaction data"})
			return
		}
	}

	response := make([]models.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		response[i] = tx.ToResponse()
	}

	c.JSON(200, gin.H{
		"data": response,
		"count": len(response),
	})
}

// GetTransaction retrieves a specific transaction by ID
func (pc *PostController) GetTransaction(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var transaction models.Transaction
	if err := pc.DB.Preload("User").Where("id = ? AND user_id = ?", c.Param("id"), userId).First(&transaction).Error; err != nil {
		c.JSON(404, gin.H{"error": "Transaction not found or unauthorized"})
		return
	}

	c.JSON(200, gin.H{"data": transaction.ToResponse()})
}