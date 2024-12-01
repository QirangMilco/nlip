package admin

import (
	"github.com/gofiber/fiber/v2"
	"nlip/config"
	"nlip/utils/logger"
)

// ServerSettings 服务器设置结构
type ServerSettings struct {
	FileTypes struct {
		AllowList []string `json:"allow_list"`
		DenyList  []string `json:"deny_list"`
	} `json:"file_types"`
	Upload struct {
		MaxSize int64 `json:"max_size"`
	} `json:"upload"`
	Space struct {
		DefaultMaxItems      int `json:"default_max_items"`
		DefaultRetentionDays int `json:"default_retention_days"`
	} `json:"space"`
	Security struct {
		TokenExpiry string `json:"token_expiry"`
	} `json:"security"`
}

// HandleGetSettings 获取当前服务器设置
func HandleGetSettings(c *fiber.Ctx) error {
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return fiber.NewError(fiber.StatusForbidden, "需要管理员权限")
	}

	settings := ServerSettings{
		FileTypes: struct {
			AllowList []string `json:"allow_list"`
			DenyList  []string `json:"deny_list"`
		}{
			AllowList: config.AppConfig.FileTypes.AllowList,
			DenyList:  config.AppConfig.FileTypes.DenyList,
		},
		Upload: struct {
			MaxSize int64 `json:"max_size"`
		}{
			MaxSize: config.AppConfig.MaxFileSize,
		},
		Space: struct {
			DefaultMaxItems      int `json:"default_max_items"`
			DefaultRetentionDays int `json:"default_retention_days"`
		}{
			DefaultMaxItems:      20,
			DefaultRetentionDays: 7,
		},
		Security: struct {
			TokenExpiry string `json:"token_expiry"`
		}{
			TokenExpiry: config.AppConfig.TokenExpiry.String(),
		},
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"data":    settings,
		"message": "获取设置成功",
	})
}

// HandleUpdateSettings 更新服务器设置
func HandleUpdateSettings(c *fiber.Ctx) error {
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return fiber.NewError(fiber.StatusForbidden, "需要管理员权限")
	}

	var settings ServerSettings
	if err := c.BodyParser(&settings); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	updates := make(map[string]interface{})

	// 更新文件类型设置
	if len(settings.FileTypes.AllowList) > 0 || len(settings.FileTypes.DenyList) > 0 {
		updates["file_types"] = map[string]interface{}{
			"allow_list": settings.FileTypes.AllowList,
			"deny_list":  settings.FileTypes.DenyList,
		}
	}

	// 更新上传设置
	if settings.Upload.MaxSize > 0 {
		updates["max_file_size"] = settings.Upload.MaxSize
	}

	// 更新空间默认设置
	if settings.Space.DefaultMaxItems > 0 || settings.Space.DefaultRetentionDays > 0 {
		updates["space_defaults"] = map[string]interface{}{
			"max_items":      settings.Space.DefaultMaxItems,
			"retention_days": settings.Space.DefaultRetentionDays,
		}
	}

	// 更新安全设置
	if settings.Security.TokenExpiry != "" {
		updates["token_expiry"] = settings.Security.TokenExpiry
	}

	if err := config.UpdateConfig(updates); err != nil {
		logger.Error("更新配置失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "更新配置失败")
	}

	logger.Info("管理员更新了服务器设置")
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "更新设置成功",
	})
} 