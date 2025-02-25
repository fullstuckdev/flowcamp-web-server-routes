package controllers

import (
    "golangapi/models"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "net/http"
)

type RelationController struct {
    DB *gorm.DB
}

func NewRelationController(db *gorm.DB) *RelationController {
    return &RelationController{DB: db}
}

// CreateProfile (One-to-One)
func (rc *RelationController) CreateProfile(c *gin.Context) {
    var profile models.Profile
    if err := c.ShouldBindJSON(&profile); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := rc.DB.Create(&profile).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": profile})
}

// CreatePost (One-to-Many)
func (rc *RelationController) CreatePost(c *gin.Context) {
    var post models.Post
    if err := c.ShouldBindJSON(&post); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := rc.DB.Create(&post).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": post})
}

// CreateGroup (Many-to-Many)
func (rc *RelationController) CreateGroup(c *gin.Context) {
    var group models.Group
    if err := c.ShouldBindJSON(&group); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := rc.DB.Create(&group).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"data": group})
}

// AddUserToGroup (Many-to-Many)
func (rc *RelationController) AddUserToGroup(c *gin.Context) {
    var request struct {
        UserID  uint `json:"user_id"`
        GroupID uint `json:"group_id"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    var group models.Group

    if err := rc.DB.First(&user, request.UserID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    if err := rc.DB.First(&group, request.GroupID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
        return
    }

    if err := rc.DB.Model(&group).Association("Users").Append(&user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User added to group successfully"})
}

// GetUserWithRelations mengambil user beserta semua relasinya
func (rc *RelationController) GetUserWithRelations(c *gin.Context) {
    var user models.User
    userID := c.Param("id")

    if err := rc.DB.Preload("Profile").
        Preload("Posts").
        Preload("Groups").
        First(&user, userID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": user})
}