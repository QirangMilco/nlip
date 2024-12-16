package auth

import (
	"database/sql"
	"encoding/json"
	"nlip/config"
	"nlip/models/space"
	"nlip/utils/jwt"
	"nlip/utils/logger"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func isGuest(authHeader string, userID string) bool {
	return authHeader == "" && userID == ""
}

func isSpaceRoute(path string) bool {
	return strings.Contains(path, "/spaces")
}

func isViewable(method string, path string) bool {
	if method == "GET" {
		return true
	}
	if method == "POST" && !strings.Contains(path, "/collaborators/invite") {
		return true
	}
	return false
}

// 根据给定路由和权限判断是否可以操作
func isOperable(method string, path string, permission string) bool {
	if permission == "view" {
		return isViewable(method, path)
	}
	if permission == "edit" {
		if method == "DELETE" {
			return strings.Contains(path, "/clips")
		}
		if strings.Contains(path, "/settings") || strings.Contains(path, "/collaborators") {
			return false
		}
		return true
	}
	return false
}

func handleGuestAccess(c *fiber.Ctx, path string, s *space.Space, userID string) (sql.NullString, bool, error) {
	if !isSpaceRoute(path) {
		return sql.NullString{}, false, nil
	}
	if strings.Contains(path, "/spaces/list") {
		if isGuest(c.Get("Authorization"), userID) {
			logger.Debug("获取空间列表无需验证token")
			return sql.NullString{}, true, nil
		}
		logger.Debug("获取空间列表无需获取空间ID")
		return sql.NullString{}, false, nil
	}

	if strings.Contains(path, "/spaces/create") {
		logger.Debug("创建空间无需获取空间ID")
		return sql.NullString{}, false, nil
	}

	if strings.Contains(path, "/spaces/verify-invite") || strings.Contains(path, "/spaces/collaborators") {
		logger.Debug("验证邀请无需获取空间ID")
		return sql.NullString{}, false, nil
	}

	//通过正则化匹配从path中获取spaceID
	re := regexp.MustCompile(`/spaces/([^/]+)`)
	spaceID := re.FindStringSubmatch(path)[1]
	logger.Debug("获取空间信息: spaceID=%s, path=%s", spaceID, path)
	var collaboratorsJSON sql.NullString
	if spaceID != "" {
		err := config.DB.QueryRow(`
			SELECT id, name, type, owner_id, collaborators
			FROM nlip_spaces WHERE id = ?
		`, spaceID).Scan(&s.ID, &s.Name, &s.Type, &s.OwnerID, &collaboratorsJSON)
		if err == sql.ErrNoRows {
			logger.Warning("尝试获取不存在的空间信息: %s", spaceID)
			return sql.NullString{}, false, fiber.NewError(fiber.StatusNotFound, "空间不存在")
		} else if err != nil {
			logger.Error("获取空间信息失败: %v", err)
			return sql.NullString{}, false, fiber.NewError(fiber.StatusInternalServerError, "获取空间信息失败")
		}

		if s.Type == "public"  {
			if strings.Contains(path, "collaborators") {
				logger.Error("公共空间不存在协作者")
				return sql.NullString{}, false, fiber.NewError(fiber.StatusNotFound, "公共空间不存在协作者")
			}
			if isGuest(c.Get("Authorization"), userID) {
				method := c.Method()
				if method == "GET" || (method == "POST" && !strings.Contains(path, "/collaborators/invite")) {
					logger.Debug("游客访问公共空间，跳过token验证")
					c.Locals("space", *s)
					return sql.NullString{}, true, nil
				}
				logger.Error("游客没有权限编辑公共空间")
				return sql.NullString{}, false, fiber.NewError(fiber.StatusForbidden, "游客没有权限编辑公共空间")
			}
		}
	}

	return collaboratorsJSON, false, nil
}

func handleTokenValidation(c *fiber.Ctx, authHeader string) error {
	if authHeader == "" {
		logger.Warning("请求缺少认证令牌: %s %s", c.Method(), c.Path())
		return fiber.NewError(fiber.StatusUnauthorized, "未提供认证令牌")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		logger.Warning("认证令牌格式错误: %s", authHeader)
		return fiber.NewError(fiber.StatusUnauthorized, "认证令牌格式错误")
	}

	claims, err := jwt.ValidateToken(parts[1])
	if err != nil {
		logger.Warning("无效的认证令牌: %v", err)
		return fiber.NewError(fiber.StatusUnauthorized, "无效的认证令牌")
	}

	c.Locals("userId", claims.UserID)
	c.Locals("username", claims.Username)
	c.Locals("isAdmin", claims.IsAdmin)

	logger.Debug("用户认证成功: userID=%s, username=%s, isAdmin=%v",
		claims.UserID, claims.Username, claims.IsAdmin)

	return nil
}

func handleSpaceAccess(c *fiber.Ctx, path string, userID string, s *space.Space, collaboratorsJSON sql.NullString) error {
	if !isSpaceRoute(path) {
		return nil
	}

	if strings.Contains(path, "/spaces/list") || strings.Contains(path, "/spaces/create") {
		logger.Debug("无需验证权限")
		return nil
	}

	if strings.Contains(path, "/spaces/verify-invite") || strings.Contains(path, "/spaces/collaborators") {
		logger.Debug("验证邀请无需验证权限")
		return nil
	}

	if s.Type == "public" {
		logger.Debug("公共空间，跳过权限验证")
		c.Locals("space", *s)
		return nil
	}

	collaboratorsMap := make(map[string]string)
	if collaboratorsJSON.Valid {
		if err := json.Unmarshal([]byte(collaboratorsJSON.String), &collaboratorsMap); err != nil {
			logger.Error("解析协作者列表失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "解析协作者数据失败")
		}

		var collaborators []space.CollaboratorInfo
		for userID, permission := range collaboratorsMap {
			collaborators = append(collaborators, space.CollaboratorInfo{
				ID:         userID,
				Permission: permission,
			})
		}
		s.Collaborators = collaborators
		s.CollaboratorsMap = collaboratorsMap
	}

	c.Locals("space", *s)

	logger.Debug("userID=%s, spaceOwnerID=%s", userID, s.OwnerID)

	if userID == s.OwnerID {
		logger.Debug("用户是空间所有者，跳过权限验证")
		return nil
	}

	permission := collaboratorsMap[userID]

	if !isOperable(c.Method(), path, permission) {
		logger.Error("没有权限操作")
		return fiber.NewError(fiber.StatusForbidden, "没有权限操作")
	}

	return nil
}

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		authHeader := c.Get("Authorization")
		var userID string
		if c.Locals("userId") == nil {
			userID = ""
		} else {
			userID = c.Locals("userId").(string)
		}

		var s space.Space

		collaboratorsJSON, guestAccess, err := handleGuestAccess(c, path, &s, userID)
		if err != nil {
			return err
		}
		if guestAccess {
			return c.Next()
		}

		if err := handleTokenValidation(c, authHeader); err != nil {
			return err
		}

		if c.Locals("userId") == nil {
			userID = ""
		} else {
			userID = c.Locals("userId").(string)
		}

		logger.Debug("开始验证权限")

		if err := handleSpaceAccess(c, path, userID, &s, collaboratorsJSON); err != nil {
			return err
		}

		logger.Debug("验证权限成功")

		return c.Next()
	}
}
