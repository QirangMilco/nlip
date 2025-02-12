package routes

import (
	"nlip/handlers/admin"
	authHandler "nlip/handlers/auth"
	"nlip/handlers/clips"
	"nlip/handlers/spaces"
	"nlip/handlers/ws"
	"nlip/middleware/auth"
	"nlip/middleware/validator"
	"nlip/models/clip"
	"nlip/models/space"
	"nlip/models/token"
	"nlip/models/user"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func SetupRoutes(router fiber.Router) {
	// 创建 v1 版本的路由组
	v1 := router.Group("/v1/nlip")
	setupV1Routes(v1)

	// 为未来的版本预留扩展点
	// v2 := router.Group("/v2/nlip")
	// setupV2Routes(v2)
}

// setupV1Routes 设置 v1 版本的所有路由
func setupV1Routes(api fiber.Router) {
	// 1. 完全公开的路由 - 不需要认证
	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "NLIP API 服务正在运行",
			"data": fiber.Map{
				"version": "1.0.0",
				"status":  "running",
			},
		})
	})

	// 2. 认证路由 - 不需要token
	authRoutes := api.Group("/auth")
	authRoutes.Post("/login", validator.ValidateBody(&user.LoginRequest{}), authHandler.HandleLogin)
	authRoutes.Post("/register", validator.ValidateBody(&user.RegisterRequest{}), authHandler.HandleRegister)
	authRoutes.Post("/token-login", validator.ValidateBody(&token.TokenLoginRequest{}), authHandler.HandleTokenLogin)

	// 3. 需要认证的路由组 - 需要token
	authenticated := api.Group("")
	authenticated.Use(auth.AuthMiddleware())

	// 用户相关路由
	authenticated.Get("/auth/me", authHandler.HandleGetCurrentUser)
	authenticated.Post("/auth/change-password",
		validator.ValidateBody(&user.ChangePasswordRequest{}),
		authHandler.HandleChangePassword)

	// 用户token相关路由
	tokenRoutes := authenticated.Group("/tokens")
	tokenRoutes.Post("/create", validator.ValidateBody(&token.CreateTokenRequest{}), authHandler.HandleCreateToken)
	tokenRoutes.Get("/list", authHandler.HandleListTokens)
	tokenRoutes.Delete("/:tokenId", authHandler.HandleRevokeToken)
	tokenRoutes.Get("/:tokenId", authHandler.HandleGetToken)

	// 空间路由 - 所有操作都需要验证
	spaceRoutes := authenticated.Group("/spaces")
	spaceRoutes.Get("/list", spaces.HandleListSpaces)
	spaceRoutes.Post("/create",
		validator.ValidateBody(&space.CreateSpaceRequest{}),
		spaces.HandleCreateSpace)
	spaceRoutes.Put("/:spaceId",
		validator.ValidateBody(&space.UpdateSpaceRequest{}),
		spaces.HandleUpdateSpace)
	spaceRoutes.Delete("/:spaceId", spaces.HandleDeleteSpace)
	spaceRoutes.Get("/:spaceId/stats", spaces.HandleSpaceStats)

	// 协作者相关路由
	spaceRoutes.Get("/:spaceId/collaborators", spaces.HandleListCollaborators)
	spaceRoutes.Post("/:spaceId/collaborators/invite",
		validator.ValidateBody(&space.InviteCollaboratorRequest{}),
		spaces.HandleInviteCollaborator)
	spaceRoutes.Post("/collaborators/verify-invite",
		validator.ValidateBody(&space.ValidateInviteRequest{}),
		spaces.HandleVerifyInviteToken)
	spaceRoutes.Post("/collaborators/accept-invite",
		validator.ValidateBody(&space.AcceptInviteRequest{}),
		spaces.HandleAcceptInvite)
	spaceRoutes.Delete("/:spaceId/collaborators/remove",
		validator.ValidateBody(&space.RemoveCollaboratorRequest{}),
		spaces.HandleRemoveCollaborator)
	spaceRoutes.Put("/:spaceId/collaborators/update-permissions",
		validator.ValidateBody(&space.UpdateCollaboratorPermissionsRequest{}),
		spaces.HandleUpdateCollaboratorPermissions)

	// 更新空间设置路由
	spaceRoutes.Put("/:spaceId/settings",
		validator.ValidateBody(&space.UpdateSpaceSettingsRequest{}),
		spaces.HandleUpdateSpaceSettings)

	// 剪贴板路由 - 所有操作都需要验证
	clipRoutes := spaceRoutes.Group("/:spaceId/clips")
	clipRoutes.Get("/list", clips.HandleListClips)
	clipRoutes.Get("/:clipId", clips.HandleGetClip)
	clipRoutes.Post("/upload",
		validator.ValidateBody(&clip.UploadClipRequest{}),
		clips.HandleUploadClip)
	clipRoutes.Put("/:clipId",
		validator.ValidateBody(&clip.UpdateClipRequest{}),
		clips.HandleUpdateClip)
	clipRoutes.Delete("/:clipId", clips.HandleDeleteClip)

	// WebSocket路由 - 需要验证
	authenticated.Get("/ws", websocket.New(ws.HandleWebSocket))

	// 管理员路由 - 需要验证
	adminRoutes := authenticated.Group("/admin")
	adminRoutes.Get("/settings", admin.HandleGetSettings)
	adminRoutes.Put("/settings", admin.HandleUpdateSettings)

	// 添加版本控制中间件
	api.Use(func(c *fiber.Ctx) error {
		c.Set("API-Version", "1.0.0")
		return c.Next()
	})
}
