package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string        `json:"name"`
	Email        string        `json:"email" gorm:"unique"`
	Password     string        `json:"-"` // "-" means it won't appear in JSON
	Balance      float64       `json:"balance" gorm:"default:0"`
	Posts        []Post        `json:"posts"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	gorm.Model
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Type        string  `json:"type" gorm:"type:enum('credit','debit')"` // credit or debit
	UserID      uint    `json:"-"`
	User        User    `json:"user" gorm:"foreignKey:UserID"`
}

// Request/Response structures
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	Balance  float64 `json:"balance" binding:"required,min=0"`
}

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"required"`
	Type        string  `json:"type" binding:"required,oneof=credit debit"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type TransactionResponse struct {
	ID          uint          `json:"id"`
	Amount      float64       `json:"amount"`
	Description string        `json:"description"`
	Type        string        `json:"type"`
	CreatedAt   time.Time     `json:"created_at"`
	User        UserResponse  `json:"user"`
}

type UpdateTransactionRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"required"`
	Type        string  `json:"type" binding:"required,oneof=credit debit"`
}

// User methods
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// Helper method to create UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Balance:   u.Balance,
		CreatedAt: u.CreatedAt,
	}
}

func (t *Transaction) ToResponse() TransactionResponse {
	return TransactionResponse{
		ID:          t.ID,
		Amount:      t.Amount,
		Description: t.Description,
		Type:        t.Type,
		CreatedAt:   t.CreatedAt,
		User:        t.User.ToResponse(),
	}
}