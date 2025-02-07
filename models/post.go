package models

import (
	"time"

	"gorm.io/gorm"
)

// Table definitions
type Post struct {
	gorm.Model
	Title   string `json:"title" gorm:"not null"`
	Content string `json:"content" gorm:"not null"`
	UserID  uint   `json:"-"`
	User    User   `json:"author" gorm:"foreignKey:UserID"`
	Tags    []Tag  `json:"tags" gorm:"many2many:post_tags"`
}

type Tag struct {
	gorm.Model
	Name  string `json:"name" gorm:"unique;not null"`
	Posts []Post `json:"posts" gorm:"many2many:post_tags"`
}

type PostTag struct {
	PostID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
	CreatedAt time.Time
}

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=3"`
	Content string `json:"content" binding:"required,min=10"`
	TagIds  []uint `json:"tag_ids" binding:"omitempty,dive,gt=0"`
}

type UpdatePostRequest struct {
	Title   string `json:"title" binding:"omitempty,min=3"`
	Content string `json:"content" binding:"omitempty,min=10"`
	TagIds  []uint `json:"tag_ids" binding:"omitempty,dive,gt=0"`
}

type TagResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type PostResponse struct {
	ID        uint          `json:"id"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Author    UserResponse  `json:"author"`
	Tags      []TagResponse `json:"tags"`
}

func (p *Post) ToResponse() PostResponse {
	tags := make([]TagResponse, len(p.Tags))
	for i, tag := range p.Tags {
		tags[i] = tag.ToResponse()
	}

	return PostResponse{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		Author:    p.User.ToResponse(),
		Tags:      tags,
	}
}

func (t *Tag) ToResponse() TagResponse {
	return TagResponse{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}
}

