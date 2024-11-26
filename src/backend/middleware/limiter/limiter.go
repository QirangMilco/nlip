package limiter

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
    "time"
)

// New 创建一个新的速率限制中间件
func New() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        60,                // 每个IP每分钟最多60个请求
        Expiration: 1 * time.Minute,   // 计数器重置时间
        KeyGenerator: func(c *fiber.Ctx) string {
            // 使用IP作为限制键
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return fiber.NewError(fiber.StatusTooManyRequests, "请求过于频繁，请稍后再试")
        },
    })
} 