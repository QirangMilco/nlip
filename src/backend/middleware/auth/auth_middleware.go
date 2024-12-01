package auth

import (
	"strings"
	"github.com/gofiber/fiber/v2"
	"nlip/utils/jwt"
	"nlip/utils/logger"
)

// AuthMiddleware 验证JWT令牌的中间件
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取Authorization头
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Warning("请求缺少认证令牌: %s %s", c.Method(), c.Path())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": fiber.StatusUnauthorized,
				"message": "未提供认证令牌",
				"data": nil,
			})
		}

		// 检查Bearer前缀
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warning("认证令牌格式错误: %s", authHeader)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": fiber.StatusUnauthorized,
				"message": "认证令牌格式错误",
				"data": nil,
			})
		}

		// 验证令牌
		claims, err := jwt.ValidateToken(parts[1])
		if err != nil {
			logger.Warning("无效的认证令牌: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": fiber.StatusUnauthorized,
				"message": "无效的认证令牌",
				"data": nil,
			})
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