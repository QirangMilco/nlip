package auth

import (
	"nlip/config"
	"nlip/models/token"
	"nlip/utils/logger"
	"nlip/utils/id"
	"time"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// HandleCreateToken 创建新的访问Token
// @Summary 创建Token
// @Description 为用户创建新的访问Token，可用于API认证
// @Tags Token
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body token.CreateTokenRequest true "创建Token请求参数"
// @Success 200 {object} token.CreateTokenResponse "创建成功"
// @Failure 400 {object} string "请求参数错误"
// @Failure 401 {object} string "未授权"
// @Failure 403 {object} string "达到最大Token数量限制"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/tokens [post]
func HandleCreateToken(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)

	// 检查当前用户的token数量
	var tokenCount int
	err := config.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM nlip_tokens 
		WHERE user_id = ?
	`, userID).Scan(&tokenCount)
	if err != nil {
		logger.Error("查询token数量失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "无法验证token数量")
	}

	// 检查是否超过最大限制
	if tokenCount >= config.AppConfig.Token.MaxItems {
		return fiber.NewError(fiber.StatusBadRequest, "已达到最大token数量限制")
	}

	var req token.CreateTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求格式")
	}

	var expiresAt *time.Time
	if req.ExpiryDays != nil && *req.ExpiryDays >= 1 && *req.ExpiryDays <= 3650 {
		exp := time.Now().AddDate(0, 0, *req.ExpiryDays)
		expiresAt = &exp
	} 

	// 生成token
	tokenStr := id.GenerateSecureToken()
	idStr := uuid.New().String()
	_, err = config.DB.Exec(`
		INSERT INTO nlip_tokens (id, user_id, token, description, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, idStr, userID, tokenStr, req.Description, expiresAt)

	var encryptedToken string
	if len(tokenStr) > 8 {
		encryptedToken = tokenStr[:4] + "****" + tokenStr[len(tokenStr)-4:]
	} else {
		encryptedToken = "****"
	}
	tokenInfo := token.Token{
		ID: idStr,
		Token: encryptedToken,
		Description: req.Description,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		LastUsedAt: nil,
	}

	if err != nil {
		logger.Error("创建token失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "创建token失败")
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "创建token成功",
		"data": token.CreateTokenResponse{
			Token: tokenStr,
			TokenInfo: &tokenInfo,
		},
	})
}

// HandleListTokens 获取Token列表
// @Summary 获取Token列表
// @Description 获取当前用户的所有Token列表
// @Tags Token
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} token.ListTokensResponse "获取成功"
// @Failure 401 {object} string "未授权"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/tokens [get]
func HandleListTokens(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)

	rows, err := config.DB.Query(`
		SELECT id, user_id, token, description, created_at, expires_at, last_used_at
		FROM nlip_tokens
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
		var fullToken string
		if err := rows.Scan(&t.ID, &t.UserID, &fullToken, &t.Description, &t.CreatedAt, &t.ExpiresAt, &t.LastUsedAt); err != nil {
			logger.Error("扫描token数据失败: %v", err)
			continue
		}
		// 处理token显示
		if len(fullToken) > 8 {
			t.Token = fullToken[:4] + "****" + fullToken[len(fullToken)-4:]
		} else {
			t.Token = "****" // 对于过短的token直接显示****
		}
		tokens = append(tokens, t)
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取token列表成功",
		"data":    fiber.Map{
			"tokens": tokens,
			"maxItems": config.AppConfig.Token.MaxItems,
		},
	})
}

// HandleRevokeToken 撤销Token
// @Summary 撤销Token
// @Description 撤销指定的Token，使其失效
// @Tags Token
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "要撤销的Token ID"
// @Success 200 {object} string "撤销成功"
// @Failure 401 {object} string "未授权"
// @Failure 404 {object} string "Token不存在"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/tokens/{id} [delete]
func HandleRevokeToken(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	tokenID := c.Params("tokenId")

	result, err := config.DB.Exec(`
		DELETE FROM nlip_tokens
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

func HandleGetToken(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	tokenID := c.Params("tokenId")

	var tokenStr string
	err := config.DB.QueryRow(`
		SELECT token 
		FROM nlip_tokens 
		WHERE id = ? AND user_id = ?
	`, tokenID, userID).Scan(&tokenStr)

	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "未找到该token")
		}
		logger.Error("获取token值失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取token值失败")
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取token值成功",
		"data": fiber.Map{
			"token": tokenStr,
		},
	})
}
