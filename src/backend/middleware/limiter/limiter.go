package limiter

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
    "strings"
    "time"
)

// New 创建一个新的速率限制中间件
func New() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        60,                // 每个IP每分钟最多60个请求
        Expiration: 1 * time.Minute,   // 计数器重置时间
        KeyGenerator: func(c *fiber.Ctx) string {
            // 根据路径设置不同的限制
            path := c.Path()
            switch {
            case strings.Contains(path, "/auth/login"):
                c.Locals("limit", 5)  // 登录接口限制
            case strings.Contains(path, "/clips/upload"):
                c.Locals("limit", 10) // 上传接口限制
            default:
                c.Locals("limit", 60) // 其他接口限制
            }
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return c.Status(429).JSON(fiber.Map{
                "code": 429,
                "message": "请求过于频繁，请稍后再试",
                "data": nil,
            })
        },
    })
} 