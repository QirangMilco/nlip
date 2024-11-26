package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "context"
    "time"
    "github.com/gofiber/fiber/v2"
    "nlip/config"
    "nlip/routes"
    "nlip/middleware"
    "nlip/middleware/logger"
    "nlip/middleware/recover"
    "nlip/middleware/limiter"
    "nlip/middleware/cors"
    "nlip/middleware/compress"
    "nlip/tasks/cleaner"
    appLogger "nlip/utils/logger"
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
        BodyLimit:   10 * 1024 * 1024, // 10MB
    })

    // 全局中间件
    app.Use(logger.New())
    app.Use(recover.New())
    app.Use(cors.New())
    app.Use(compress.New())
    app.Use(limiter.New())

    // 设置路由
    routes.SetupRoutes(app)

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
} 