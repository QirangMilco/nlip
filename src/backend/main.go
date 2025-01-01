package main

import (
	"context"
	"embed"
	"io/fs"
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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed static/dist/*
var embedDistFiles embed.FS

func main() {
	// 加载配置
	config.LoadConfig()

	// 验证配置
	if err := config.ValidateConfig(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	appLogger.SetAppEnv(config.AppConfig.AppEnv)

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

	// API路由组 - 移到静态文件处理之前
	api := app.Group("/api")
	routes.SetupRoutes(api)

	// 静态文件服务
	distFS, err := fs.Sub(embedDistFiles, "static/dist")
	if err != nil {
		log.Fatalf("无法加载嵌入的静态文件: %v", err)
	}

	// 使用 filesystem 中间件服务静态文件
	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(distFS),
		Browse:     false,
		Next: func(c *fiber.Ctx) bool {
			// 只有当路径不是静态文件时才跳过
			return !(c.Path() == "/favicon.ico" || strings.HasPrefix(c.Path(), "/assets/"))
		},
	}))

	// SPA 路由处理放在最后
	app.Use("/*", func(c *fiber.Ctx) error {
		// 读取嵌入的 index.html
		indexContent, err := embedDistFiles.ReadFile("static/dist/index.html")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("无法读取 index.html")
		}
		
		c.Set("Content-Type", "text/html")
		return c.Send(indexContent)
	})

	// 启动清理任务
	cleaner.StartCleanupTask()

	// 记录启动日志
	appLogger.Info("服务器启动在端口 :%s", config.AppConfig.ServerPort)

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
	if err := app.Listen(":" + config.AppConfig.ServerPort); err != nil {
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
