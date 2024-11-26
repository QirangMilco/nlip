package routes

import (
	"github.com/gofiber/fiber/v2"
	authHandler "nlip/handlers/auth"
	"nlip/handlers/spaces"
	"nlip/handlers/clips"
	"nlip/handlers/info"
	"nlip/middleware/auth"
	"nlip/middleware/validator"
	"nlip/models/user"
	"nlip/models/space"
	"nlip/models/clip"
)

func SetupRoutes(app *fiber.App) {
	// 添加根路径处理
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "success",
			"message": "NLIP API 服务正在运行",
			"version": "1.0.0",
		})
	})

	// API路由组
	api := app.Group("/api/v1/nlip")

	// 认证路由
	authRoutes := api.Group("/auth")
	authRoutes.Post("/login", validator.ValidateBody(&user.LoginRequest{}), authHandler.HandleLogin)
	authRoutes.Post("/register", validator.ValidateBody(&user.RegisterRequest{}), authHandler.HandleRegister)
	authRoutes.Get("/me", authHandler.HandleGetCurrentUser)

	// 需要认证的路由组
	authenticated := api.Use(auth.AuthMiddleware())

	// 信息路由
	infoRoutes := authenticated.Group("/info")
	infoRoutes.Get("/me", info.HandleGetCurrentUserInfo)

	// 空间路由
	spaceRoutes := authenticated.Group("/spaces")
	spaceRoutes.Get("/list", spaces.HandleListSpaces)
	spaceRoutes.Post("/create", validator.ValidateBody(&space.CreateSpaceRequest{}), spaces.HandleCreateSpace)
	spaceRoutes.Put("/:id", validator.ValidateBody(&space.UpdateSpaceRequest{}), spaces.HandleUpdateSpace)
	spaceRoutes.Delete("/:id", spaces.HandleDeleteSpace)

	// 剪贴板路由
	clipRoutes := authenticated.Group("/clips")
	clipRoutes.Post("/upload", validator.ValidateBody(&clip.UploadClipRequest{}), clips.HandleUploadClip)
	clipRoutes.Get("/list", clips.HandleListClips)
	clipRoutes.Get("/:id", clips.HandleGetClip)
	clipRoutes.Delete("/:id", clips.HandleDeleteClip)
} 