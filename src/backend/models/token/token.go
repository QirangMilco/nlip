package token

import (
	"time"
	"nlip/models/user"
)

type Token struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Token       string    `json:"token"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
}

type CreateTokenRequest struct {
	Description string `json:"description" validate:"required,max=100"`
	ExpiryDays  *int    `json:"expiry_days" validate:"required,min=1,max=3650"`
}

type CreateTokenResponse struct {
	Token string `json:"token"`
	TokenInfo *Token `json:"token_info"`
}

type TokenLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Token string `json:"token" validate:"required"`
}

type TokenLoginResponse struct {
	JWTToken string `json:"jwt_token"`
	User     *user.User  `json:"user"`
}

type ListTokensResponse struct {
	Tokens []*Token `json:"tokens"`
	MaxItems int `json:"max_items"`
}
