package recover

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// New 创建一个新的恢复中间件
func New() fiber.Handler {
	return recover.New()
} 