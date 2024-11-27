package jwt

import (
    "time"
    "github.com/golang-jwt/jwt/v4"
    "nlip/config"
    "nlip/models/user"
    "nlip/utils/logger"
)

type Claims struct {
    UserID   string `json:"userId"`
    Username string `json:"username"`
    IsAdmin  bool   `json:"isAdmin"`
    jwt.RegisteredClaims
}

type UserClaims struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	IsAdmin      bool   `json:"isAdmin"`
	NeedChangePwd bool  `json:"needChangePwd"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(user *user.User) (string, error) {
    claims := UserClaims{
        UserID:   user.ID,
        Username: user.Username,
        IsAdmin:  user.IsAdmin,
        NeedChangePwd: user.NeedChangePwd,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AppConfig.TokenExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
    if err != nil {
        logger.Error("生成JWT令牌失败: %v", err)
        return "", err
    }

    logger.Debug("生成JWT令牌: userID=%s, username=%s, isAdmin=%v", 
        user.ID, user.Username, user.IsAdmin)
    return tokenString, nil
}

// ValidateToken 验证JWT令牌
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.AppConfig.JWTSecret), nil
    })

    if err != nil {
        logger.Warning("JWT令牌验证失败: %v", err)
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        logger.Debug("JWT令牌验证成功: userID=%s, username=%s", claims.UserID, claims.Username)
        return claims, nil
    }

    logger.Warning("JWT令牌签名无效")
    return nil, jwt.ErrSignatureInvalid
} 