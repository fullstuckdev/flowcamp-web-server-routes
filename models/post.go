package models

import (
	"time"
	"gorm.io/gorm"
)

// Table 
type Tag struct {
	gorm.Model
	Name string `json:"name" gorm:"unique"`
	Posts []Post `json:"posts" gorm:"many2many:post_tags"`
}

// Table
type Post struct {
	gorm.Model
	Title string `json:"title"`
	Content string `json:"content"`
	UserId uint `json:"-"`
	User User `json:"author" gorm:"foreignKey:UserId"`
	Tags []Tag `json:"tags" gorm:"many2many:post_tags"`
}

// Table
type PostTag struct {
	PostID uint `gorm:"primaryKey"`
	TagID uint `gorm:"primaryKey"`
}

// Request data
type CreatePostRequest struct {
	Title string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	TagIds []uint `json:"tag_ids"`
}

type UserResponse struct {
	ID uint `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

type TagResponse struct {
	ID uint `json:"id"`
	Name string `json:"name"`
}

type PostResponse struct {
	ID uint `json:"id"`
	Title string `json:"title"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Author UserResponse `json:"author"`
	Tags []TagResponse `json:"tags"`
}


type UpdatePostRequest struct {
	Title   string `json:"title" binding:"omitempty,min=3"`
	Content string `json:"content" binding:"omitempty,min=10"`
	TagIds  []uint `json:"tag_ids" binding:"omitempty,dive,gt=0"`
}

