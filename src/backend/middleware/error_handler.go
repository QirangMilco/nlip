package middleware

import (
	"github.com/gofiber/fiber/v2"
	"nlip/utils/logger"
)

// CustomErrorHandler 处理所有的错误响应
func CustomErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "服务器内部错误"

	// 处理NlipError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// 记录错误日志
	if code >= 500 {
		logger.Error("服务器错误: %d - %s - %s %s", 
			code, message, c.Method(), c.Path())
	} else {
		logger.Warning("客户端错误: %d - %s - %s %s", 
			code, message, c.Method(), c.Path())
	}

	// 返回统一的响应格式
	return c.Status(code).JSON(fiber.Map{
		"code": code,
		"message": message,
		"data": nil,
	})
} 