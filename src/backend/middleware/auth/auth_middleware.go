package auth

import (
	"strings"
	"github.com/gofiber/fiber/v2"
	"nlip/utils/jwt"
	"nlip/utils/logger"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取Authorization头
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Warning("请求缺少认证令牌: %s %s", c.Method(), c.Path())
			return fiber.NewError(fiber.StatusUnauthorized, "未提供认证令牌")
		}

		// 检查Bearer前缀
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warning("认证令牌格式错误: %s", authHeader)
			return fiber.NewError(fiber.StatusUnauthorized, "认证令牌格式错误")
		}

		// 验证令牌
		claims, err := jwt.ValidateToken(parts[1])
		if err != nil {
			logger.Warning("无效的认证令牌: %v", err)
			return fiber.NewError(fiber.StatusUnauthorized, "无效的认证令牌")
		}

		// 将用户信息存储在上下文中
		c.Locals("userId", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("isAdmin", claims.IsAdmin)

		logger.Debug("用户认证成功: userID=%s, username=%s, isAdmin=%v", 
			claims.UserID, claims.Username, claims.IsAdmin)

		return c.Next()
	}
} 