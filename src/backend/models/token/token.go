package token

import (
	"time"
	"nlip/models/user"
)

type Token struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Token       string    `json:"token"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	LastUsedAt  *time.Time `json:"lastUsedAt"`
}

type CreateTokenRequest struct {
	Description string `json:"description" validate:"required,max=100"`
	ExpiryDays  *int    `json:"expiryDays"`
}

type CreateTokenResponse struct {
	Token string `json:"token"`
	TokenInfo *Token `json:"tokenInfo"`
}

type TokenLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Token string `json:"token" validate:"required"`
}

type TokenLoginResponse struct {
	JWTToken string `json:"jwtToken"`
	User     *user.User  `json:"user"`
}

type ListTokensResponse struct {
	Tokens []*Token `json:"tokens"`
	MaxItems int `json:"maxItems"`
}
