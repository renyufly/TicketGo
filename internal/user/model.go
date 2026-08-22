package user

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateInput struct {
	Email    string `json:"email" binding:"required,email,max=320"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Role     string `json:"role" binding:"omitempty,oneof=customer admin"`
}
type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=320"`
	Password string `json:"password" binding:"required,max=72"`
}
