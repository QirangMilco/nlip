package user

import (
    "time"
)

type User struct {
    ID           string    `json:"id"`
    Username     string    `json:"username"`
    PasswordHash string    `json:"-"`
    IsAdmin      bool      `json:"isAdmin"`
    CreatedAt    time.Time `json:"createdAt"`
}

type LoginRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Password string `json:"password" validate:"required,min=6,max=50"`
}

type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Password string `json:"password" validate:"required,min=6,max=50"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  *User  `json:"user"`
} 