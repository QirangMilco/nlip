package auth

import (
	"nlip/config"
	"nlip/models/token"
	"nlip/utils/logger"
	"nlip/utils/id"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func HandleCreateToken(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)

	var req token.CreateTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求格式")
	}

	var expiresAt *time.Time
	if req.ExpiryDays != nil {
		exp := time.Now().AddDate(0, 0, *req.ExpiryDays)
		expiresAt = &exp
	} 

	// 生成token
	tokenStr := id.GenerateSecureToken()
	_, err := config.DB.Exec(`
		INSERT INTO tokens (id, user_id, token, description, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, uuid.New().String(), userID, tokenStr, req.Description, expiresAt)

	if err != nil {
		logger.Error("创建token失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "创建token失败")
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "创建token成功",
		"data": token.TokenResponse{
			Token: tokenStr,
		},
	})
}

func HandleListTokens(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)

	rows, err := config.DB.Query(`
		SELECT id, description, created_at, expires_at, last_used_at
		FROM tokens
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		logger.Error("获取token列表失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取token列表失败")
	}
	defer rows.Close()

	var tokens []token.Token
	for rows.Next() {
		var t token.Token
		if err := rows.Scan(&t.ID, &t.Description, &t.CreatedAt, &t.ExpiresAt, &t.LastUsedAt); err != nil {
			logger.Error("扫描token数据失败: %v", err)
			continue
		}
		tokens = append(tokens, t)
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取token列表成功",
		"data":    tokens,
	})
}

func HandleRevokeToken(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	tokenID := c.Params("id")

	result, err := config.DB.Exec(`
		DELETE FROM tokens
		WHERE id = ? AND user_id = ?
	`, tokenID, userID)

	if err != nil {
		logger.Error("删除token失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "删除token失败")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "未找到该token")
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "删除token成功",
		"data":    nil,
	})
}
