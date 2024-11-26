package middleware

import (
	"github.com/gofiber/fiber/v2"
	"nlip/utils/logger"
)

// CustomErrorHandler 处理所有的错误响应
func CustomErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	// 检查是否是自定义错误
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// 记录错误日志
	if code >= 500 {
		logger.Error("服务器错误: %d - %s - %s %s", 
			code, err.Error(), c.Method(), c.Path())
	} else {
		logger.Warning("客户端错误: %d - %s - %s %s", 
			code, err.Error(), c.Method(), c.Path())
	}

	// 返回JSON格式的错误信息
	return c.Status(code).JSON(fiber.Map{
		"code":    code,
		"message": err.Error(),
		"data":    nil,
	})
} 