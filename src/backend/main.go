package main

import (
	"context"
	"log"
	"nlip/config"
	"nlip/middleware"
	"nlip/middleware/compress"
	"nlip/middleware/cors"
	"nlip/middleware/limiter"
	"nlip/middleware/logger"
	"nlip/middleware/recover"
	"nlip/routes"
	"nlip/tasks/cleaner"
	appLogger "nlip/utils/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
	"path/filepath"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 验证配置
	if err := config.ValidateConfig(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 初始化数据库
	if err := config.InitDatabase(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer config.CloseDatabase()

	// 初始化应用
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.CustomErrorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10MB
	})

	// 全局中间件
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(compress.New())
	app.Use(limiter.New())

	// 添加静态文件服务
	// 1. 直接服务dist目录
	distPath := "./static/dist"

	app.Static("/", distPath)

	// 2. 对于SPA应用,所有未匹配的路由重定向到index.html
	app.Get("/*", func(c *fiber.Ctx) error {
		// 如果请求的是API路由,跳过
		if len(c.Path()) >= 4 && c.Path()[:4] == "/api" {
			return c.Next()
		}
		// 其他路由返回index.html
		return c.SendFile(filepath.Join(distPath, "index.html"))
	})

	// API路由组
	api := app.Group("/api")
	routes.SetupRoutes(api)

	// 启动清理任务
	cleaner.StartCleanupTask()

	// 记录启动日志
	appLogger.Info("服务器启动在端口 :3000")

	// 优雅关闭处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		appLogger.Info("接收到关闭信号，开始优雅关闭")

		// 设置关闭超时
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 关闭服务器
		if err := app.ShutdownWithContext(ctx); err != nil {
			appLogger.Error("服务器关闭失败: %v", err)
		}

		// 关闭日志
		appLogger.Close()
	}()

	// 启动服务器
	if err := app.Listen(":3000"); err != nil {
		appLogger.Error("服务器启动失败: %v", err)
		log.Fatal(err)
	}

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Download-Options", "noopen")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("X-DNS-Prefetch-Control", "off")
		return c.Next()
	})
}
