package spaces

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"nlip/config"
	"nlip/models/space"
	"nlip/utils/db"
	"nlip/utils/email"
	"nlip/utils/id"
	"nlip/utils/logger"
	"time"

	"github.com/gofiber/fiber/v2"
)

func getMessage(emailEnabled bool) string {
	if emailEnabled {
		return "生成邀请链接成功，已发送邮件"
	}
	return "生成邀请链接成功，请手动将邀请链接发送给协作者"
}

// getSpace 获取空间信息
// @Summary 获取单个空间信息
// @Description 通过空间ID获取空间详细信息，优先从locals获取，不存在时查询数据库
// @Tags 空间
// @Param spaceID path string true "空间ID"
// @Success 200 {object} space.Space "获取空间信息成功"
// @Failure 404 {object} string "空间不存在"
// @Failure 500 {object} string "服务器内部错误"
func getSpace(c *fiber.Ctx, spaceID string) (space.Space, error) {
	if c.Locals("space") != nil {
		localSpace := c.Locals("space").(space.Space)
		if localSpace.ID == spaceID {
			return localSpace, nil
		}
	}
	logger.Debug("获取空间信息: spaceID=%s", spaceID)
	var s space.Space
	var collaboratorsJSON sql.NullString
	if spaceID != "" {
		err := config.DB.QueryRow(`
			SELECT id, name, type, owner_id, collaborators
			FROM nlip_spaces WHERE id = ?
		`, spaceID).Scan(&s.ID, &s.Name, &s.Type, &s.OwnerID, &collaboratorsJSON)
		if err == sql.ErrNoRows {
			logger.Warning("尝试获取不存在的空间信息: %s", spaceID)
			return space.Space{}, c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    fiber.StatusNotFound,
				"message": "空间不存在",
				"data":    nil,
			})
		} else if err != nil {
			logger.Error("获取空间信息失败: %v", err)
			return space.Space{}, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    fiber.StatusInternalServerError,
				"message": "获取空间信息失败",
				"data":    nil,
			})
		}

		collaboratorsMap := make(map[string]string)
		if collaboratorsJSON.Valid {
			if err := json.Unmarshal([]byte(collaboratorsJSON.String), &collaboratorsMap); err != nil {
				logger.Error("解析协作者列表失败: %v", err)
				return space.Space{}, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    fiber.StatusInternalServerError,
					"message": "解析协作者数据失败",
					"data":    nil,
				})
			}
		}

		if s.Type == "private" {
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
	}
	return s, nil
}

// HandleListSpaces 获取空间列表
// @Summary 获取空间列表
// @Description 获取当前用户有权限访问的空间列表，包括自己创建的和协作的空间
// @Tags 空间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} space.ListSpacesResponse "获取空间列表成功"
// @Failure 401 {object} string "未授权"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/list [get]
func HandleListSpaces(c *fiber.Ctx) error {
	// 尝试获取用户信息，如果不存在则表示未认证
	userID, ok := c.Locals("userId").(string)
	isAdmin, _ := c.Locals("isAdmin").(bool)

	logger.Debug("获取空间列表: authenticated=%v, userID=%s, isAdmin=%v", ok, userID, isAdmin)

	var rows *sql.Rows
	var err error

	// 根据认证状态决定查询逻辑
	if !ok {
		// 未认证用户只能看到公共空间
		rows, err = db.QueryRows(config.DB, `
            SELECT id, name, type, owner_id, max_items, retention_days, collaborators, created_at, updated_at
            FROM nlip_spaces 
            WHERE type = 'public'
            ORDER BY created_at DESC
        `)
	} else if isAdmin {
		// 管理员可以看到所有空间
		rows, err = db.QueryRows(config.DB, `
            SELECT id, name, type, owner_id, max_items, retention_days, collaborators, created_at, updated_at
            FROM nlip_spaces
            ORDER BY created_at DESC
        `)
	} else {
		// 已认证的普通用户可以看到:
		// 1. 公共空间
		// 2. 自己创建的私有空间
		// 3. 作为协作者的空间
		rows, err = db.QueryRows(config.DB, `
            SELECT id, name, type, owner_id, max_items, retention_days, collaborators, created_at, updated_at
            FROM nlip_spaces 
            WHERE type = 'public' 
                OR owner_id = ?
                OR collaborators LIKE ?
            ORDER BY created_at DESC
        `, userID, "%"+userID+"%")
	}

	if err != nil {
		logger.Error("获取空间列表失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取空间列表失败")
	}
	defer rows.Close()

	var spaces []space.Space
	for rows.Next() {
		var s space.Space
		var collaboratorsJSON sql.NullString
		err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Type,
			&s.OwnerID,
			&s.MaxItems,
			&s.RetentionDays,
			&collaboratorsJSON,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			logger.Error("读取空间数据失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "读取空间数据失败")
		}

		s.Collaborators = []space.CollaboratorInfo{}

		collaboratorsMap := make(map[string]string)

		if collaboratorsJSON.Valid && collaboratorsJSON.String != "" {
			if err := json.Unmarshal([]byte(collaboratorsJSON.String), &collaboratorsMap); err != nil {
				logger.Error("解析协作者列表失败: %v", err)
				continue
			}

			// 查询协作者详细信息
			for collaborator, permission := range collaboratorsMap {
				var username string
				err := config.DB.QueryRow(`
					SELECT username FROM nlip_users WHERE id = ?
				`, collaborator).Scan(&username)

				if err != nil {
					logger.Warning("获取协作者信息失败: %v", err)
					continue
				}

				s.Collaborators = append(s.Collaborators, space.CollaboratorInfo{
					ID:         collaborator,
					Username:   username,
					Permission: permission,
				})
			}
		}

		spaces = append(spaces, s)
	}

	if err = rows.Err(); err != nil {
		logger.Error("遍历空间数据时发生错误: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "读取空间数据失败")
	}

	logger.Info("用户 %s 获取了 %d 个空间的列表", userID, len(spaces))
	return c.JSON(fiber.Map{
		"code": fiber.StatusOK,
		"data": space.ListSpacesResponse{
			Spaces: spaces,
		},
		"message": "获取空间列表成功",
	})
}

// HandleCreateSpace 创建空间
// @Summary 创建新空间
// @Description 创建一个新的空间，普通用户只能创建私有空间，管理员可以创建公共空间
// @Tags 空间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body space.CreateSpaceRequest true "创建空间请求参数"
// @Success 201 {object} space.SpaceResponse "创建空间成功"
// @Failure 400 {object} string "请求参数错误"
// @Failure 403 {object} string "没有权限创建公共空间"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/create [post]
func HandleCreateSpace(c *fiber.Ctx) error {
	var req space.CreateSpaceRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析创建空间请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	// 设置默认值
	if req.Type == "" {
		req.Type = "private" // 默认创建私有空间
	}

	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	logger.Debug("处理创建空间请求: name=%s, type=%s, userID=%s", req.Name, req.Type, userID)

	// 只有管理员可以创建公共空间
	if req.Type == "public" && !isAdmin {
		logger.Warning("非管理员用户 %s 尝试创建公共空间", userID)
		return fiber.NewError(fiber.StatusForbidden, "没有权限创建公共空间")
	}

	now := time.Now()
	// 创建空间
	s := space.Space{
		ID:            id.GenerateSpaceID(),
		Name:          req.Name,
		Type:          req.Type,
		OwnerID:       userID,
		MaxItems:      req.MaxItems,
		RetentionDays: req.RetentionDays,
		CreatedAt:     now,
		UpdatedAt:     now,
		Collaborators: req.Collaborators,
	}

	// 将 Collaborators 转换为 JSON 字符串
	collaboratorsJSON, err := json.Marshal(s.Collaborators)
	if err != nil {
		logger.Error("序列化邀请用户数据失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "创建空间失败")
	}

	// 插入数据库
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		_, err = db.ExecTx(tx, `
            INSERT INTO nlip_spaces (id, name, type, owner_id, max_items, retention_days, collaborators, created_at, updated_at) 
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, s.ID, s.Name, s.Type, s.OwnerID, s.MaxItems, s.RetentionDays, string(collaboratorsJSON), s.CreatedAt, s.UpdatedAt)
		return err
	})

	if err != nil {
		logger.Error("创建空间失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "创建空间失败")
	}

	logger.Info("用户 %s 创建了新空间: id=%s, name=%s, type=%s", userID, s.ID, s.Name, s.Type)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"code": fiber.StatusCreated,
		"data": space.SpaceResponse{
			Space: &s,
		},
		"message": "创建空间成功",
	})
}

// HandleUpdateSpace 更新空间信息
// @Summary 更新空间
// @Description 更新空间信息，普通用户只能更新自己拥有的私有空间，管理员可以更新所有空间
// @Tags 空间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body space.UpdateSpaceRequest true "更新空间请求参数"
// @Success 200 {object} space.Space "更新成功，返回更新后的空间信息"
// @Failure 400 {object} string "请求参数错误"
// @Failure 403 {object} string "没有权限更新该空间"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/{id} [put]
func HandleUpdateSpace(c *fiber.Ctx) error {
	var req space.UpdateSpaceRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析更新空间请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}
	s := c.Locals("space").(space.Space)
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	logger.Debug("处理更新空间请求: id=%s, userID=%s", s.ID, userID)

	if s.Type == "public" && !isAdmin {
		logger.Warning("非管理员用户 %s 尝试更新公共空间", userID)
		return fiber.NewError(fiber.StatusForbidden, "没有权限更新公共空间")
	}

	// 获取服务器设置
	serverSettings := config.AppConfig.Space

	// 更新字段并检查是否超过服务器设置
	if req.Name != "" {
		s.Name = req.Name
	}
	if req.MaxItems > 0 {
		if req.MaxItems > serverSettings.MaxItemsLimit {
			s.MaxItems = serverSettings.MaxItemsLimit
		} else {
			s.MaxItems = req.MaxItems
		}
	}
	if req.RetentionDays > 0 {
		if req.RetentionDays > serverSettings.MaxRetentionDaysLimit {
			s.RetentionDays = serverSettings.MaxRetentionDaysLimit
		} else {
			s.RetentionDays = req.RetentionDays
		}
	}

	var err error

	if req.Collaborators != nil {
		collaboratorsJSON, jsonErr := json.Marshal(req.Collaborators)
		if jsonErr != nil {
			logger.Error("序列化邀请用户数据失败: %v", jsonErr)
			return fiber.NewError(fiber.StatusInternalServerError, "更新空间失败")
		}

		err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
			var execErr error
			_, execErr = db.ExecTx(tx, `
                UPDATE nlip_spaces 
                SET name = ?, max_items = ?, retention_days = ?, collaborators = ?, updated_at = ?
                WHERE id = ?
            `, s.Name, s.MaxItems, s.RetentionDays, string(collaboratorsJSON), time.Now(), s.ID)
			return execErr
		})
	} else {
		err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
			var execErr error
			_, execErr = db.ExecTx(tx, `
                UPDATE nlip_spaces 
                SET name = ?, max_items = ?, retention_days = ?, updated_at = ?
                WHERE id = ?
            `, s.Name, s.MaxItems, s.RetentionDays, time.Now(), s.ID)
			return execErr
		})
	}

	if err != nil {
		logger.Error("更新空间失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "更新空间失败")
	}

	// 重新获取更新后的空间信息（包括新的 updated_at）
	err = config.DB.QueryRow(`
        SELECT id, name, type, owner_id, max_items, retention_days, created_at, updated_at
        FROM nlip_spaces WHERE id = ?
    `, s.ID).Scan(
		&s.ID,
		&s.Name,
		&s.Type,
		&s.OwnerID,
		&s.MaxItems,
		&s.RetentionDays,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		logger.Error("获取更新后的空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取更新后的空间信息失败")
	}

	logger.Info("用户 %s 更新了空间: id=%s, name=%s", userID, s.ID, s.Name)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "更新空间成功",
		"data":    s,
	})
}

// HandleDeleteSpace 删除空间
// @Summary 删除空间
// @Description 删除空间及其所有内容，普通用户只能删除自己拥有的私有空间，管理员可以删除所有空间
// @Tags 空间
// @Security BearerAuth
// @Success 200 {object} string "删除成功"
// @Failure 403 {object} string "没有权限删除该空间"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/{id} [delete]
func HandleDeleteSpace(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	s := c.Locals("space").(space.Space)
	isAdmin := c.Locals("isAdmin").(bool)

	logger.Debug("处理删除空间请求: id=%s, userID=%s", s.ID, userID)

	if s.Type == "public" && !isAdmin {
		logger.Warning("非管理员用户 %s 尝试删除公共空间", userID)
		return fiber.NewError(fiber.StatusForbidden, "没有权限删除公共空间")
	}

	// 开始事务
	err := db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 删除空间内的所有内容
		var execErr error
		_, execErr = db.ExecTx(tx, "DELETE FROM nlip_clipboard_items WHERE space_id = ?", s.ID)
		if execErr != nil {
			return execErr
		}

		// 删除空间
		_, execErr = db.ExecTx(tx, "DELETE FROM nlip_spaces WHERE id = ?", s.ID)
		return execErr
	})

	if err != nil {
		logger.Error("删除空间失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "删除空间失败")
	}

	logger.Info("用户 %s 删除了空间: id=%s", userID, s.ID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"data":    nil,
		"message": "删除空间成功",
	})
}

// HandleInviteCollaborator 邀请协作者
// @Summary 邀请协作者
// @Description 生成邀请链接并可选发送邀请邮件，只有空间所有者可以邀请协作者
// @Tags 空间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body space.InviteCollaboratorRequest true "邀请协作者请求参数"
// @Success 200 {object} space.InviteCollaboratorResponse "邀请成功"
// @Failure 400 {object} string "请求参数错误"
// @Failure 403 {object} string "没有权限邀请协作者"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/{id}/invite [post]
func HandleInviteCollaborator(c *fiber.Ctx) error {
	var req space.InviteCollaboratorRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析邀请协作者请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	s := c.Locals("space").(space.Space)
	userID := c.Locals("userId").(string)

	logger.Debug("空间所有者ID: %s, 用户ID: %s", s.OwnerID, userID)
	if s.OwnerID != userID {
		return fiber.NewError(fiber.StatusForbidden, "只有空间所有者可以邀请协作者")
	}

	// 生成唯一的令牌哈希值
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s_%s_%d", s.ID, userID, time.Now().UnixNano())))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	// 存储邀请信息到数据库
	err := db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		_, err := db.ExecTx(tx, `
			INSERT INTO nlip_invites (
				token_hash, 
				space_id, 
				created_by, 
				created_at,
				expires_at,
				permission
			) VALUES (?, ?, ?, strftime('%s', 'now'), datetime('now', '+24 hours'), ?)
		`, tokenHash, s.ID, userID, req.Permission)
		return err
	})

	if err != nil {
		logger.Error("存储邀请令牌失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "生成邀请链接失败")
	}

	inviteLink := fmt.Sprintf("%s/invite/%s", config.AppConfig.FrontendURL, tokenHash)

	if config.AppConfig.Email.Enabled {
		// 发送邀请邮件（如果启用了邮件功能）
		if err = email.SendInviteEmail(req.Email, s.Name, inviteLink); err != nil {
			logger.Error("发送邀请邮件失败: %v", err)
			return c.JSON(fiber.Map{
				"code":    fiber.StatusOK,
				"message": "生成邀请链接成功，但邮件发送失败",
				"data": space.InviteCollaboratorResponse{
					InviteLink: inviteLink,
				},
				"warning": "邮件发送失败，请手动将邀请链接发送给协作者",
			})
		}
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": getMessage(config.AppConfig.Email.Enabled),
		"data": space.InviteCollaboratorResponse{
			InviteLink: inviteLink,
		},
	})
}

// HandleVerifyInviteToken 验证邀请令牌
// @Summary 验证邀请令牌
// @Description 验证邀请令牌的有效性，返回空间信息和邀请者信息
// @Tags 空间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body space.ValidateInviteRequest true "验证邀请令牌请求参数"
// @Success 200 {object} space.VerifyInviteTokenResponse "验证成功"
// @Failure 400 {object} string "无效的邀请链接或已过期"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/invite/verify [post]
func HandleVerifyInviteToken(c *fiber.Ctx) error {
	var req space.ValidateInviteRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析验证邀请令牌请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	logger.Debug("开始获取用户ID")
	userID := c.Locals("userId").(string)
	logger.Debug("验证邀请令牌请求: userID=%s, token=%s", userID, req.Token)

	// 从数据库中查询邀请信息和邀请者用户名
	var spaceID, createdBy, permission, inviterName string
	err := config.DB.QueryRow(`
		SELECT i.space_id, i.created_by, i.permission, u.username
		FROM nlip_invites i
		JOIN nlip_users u ON i.created_by = u.id
		WHERE i.token_hash = ? 
			AND i.used_at IS NULL 
			AND i.expires_at > strftime('%s', 'now')
	`, req.Token).Scan(&spaceID, &createdBy, &permission, &inviterName)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusBadRequest, "无效的邀请链接或已过期")
	} else if err != nil {
		logger.Error("验证邀请令牌失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "服务器错误")
	}

	logger.Debug("验证邀请令牌成功: spaceID=%s, createdBy=%s, permission=%s", spaceID, createdBy, permission)

	s, err := getSpace(c, spaceID)

	if err != nil {
		logger.Error("获取空间失败: %s", err.Error())
		return err
	}

	isCollaborator := false
	if s.CollaboratorsMap != nil && s.CollaboratorsMap[userID] != "" {
		isCollaborator = true
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "验证成功",
		"data": space.VerifyInviteTokenResponse{
			SpaceID:        s.ID,
			SpaceName:      s.Name,
			InviterName:    inviterName,
			Permission:     permission,
			IsCollaborator: isCollaborator,
		},
	})
}

// HandleAcceptInvite 接受邀请
// @Summary 接受邀请
// @Description 接受空间邀请并成为协作者
// @Tags 空间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body space.AcceptInviteRequest true "接受邀请请求参数"
// @Success 200 {object} string "成功加入空间"
// @Failure 400 {object} string "无效的请求数据或用户已加入空间"
// @Failure 403 {object} string "空间所有者不能接受邀请"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/invite/accept [post]
func HandleAcceptInvite(c *fiber.Ctx) error {
	var req space.AcceptInviteRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析接受邀请请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)

	// 检查令牌是否存在且未过期
	var spaceID, permission string
	err := config.DB.QueryRow(`
		SELECT space_id, permission 
		FROM nlip_invites 
		WHERE token_hash = ? 
			AND used_at IS NULL 
			AND expires_at > strftime('%s', 'now')
	`, req.Token).Scan(&spaceID, &permission)

	if err == sql.ErrNoRows {
		logger.Error("邀请链接无效或已过期")
		return fiber.NewError(fiber.StatusBadRequest, "邀请链接无效或已过期")
	} else if err != nil {
		logger.Error("验证邀请令牌失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "服务器错误")
	}

	s, err := getSpace(c, spaceID)

	if err != nil {
		logger.Error("获取空间失败: %s", err.Error())
		return err
	}

	if s.OwnerID == userID {
		logger.Error("空间所有者不能接受邀请")
		return fiber.NewError(fiber.StatusForbidden, "空间所有者不能接受邀请")
	}

	if s.CollaboratorsMap != nil && s.CollaboratorsMap[userID] != "" {
		logger.Error("用户已加入空间")
		return fiber.NewError(fiber.StatusBadRequest, "用户已加入空间")
	}

	// 开启事务
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 标记令牌为已使用
		_, err = db.ExecTx(tx, `
			UPDATE nlip_invites 
			SET used_at = strftime('%s', 'now'),
				used_by = ?
			WHERE token_hash = ?
		`, userID, req.Token)
		if err != nil {
			logger.Error("标记邀请令牌为已使用失败: %v", err)
			return err
		}

		// 更新协作者列表
		if s.CollaboratorsMap == nil {
			s.CollaboratorsMap = make(map[string]string)
		}
		s.CollaboratorsMap[userID] = permission
		newCollaboratorsJSON, err := json.Marshal(s.CollaboratorsMap)
		if err != nil {
			logger.Error("序列化协作者列表失败: %v", err)
			return err
		}

		_, err = db.ExecTx(tx, `
			UPDATE nlip_spaces 
			SET collaborators = ?, 
				updated_at = strftime('%s', 'now')
			WHERE id = ?
		`, string(newCollaboratorsJSON), spaceID)

		return err
	})

	if err != nil {
		logger.Error("添加协作者失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "接受邀请失败")
	}

	logger.Info("用户 %s 成功加入空间 %s", userID, spaceID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "成功加入空间",
		"data":    nil,
	})
}

// HandleRemoveCollaborator 删除协作者
// @Summary 删除协作者
// @Description 从空间中删除协作者，只有空间所有者可以执行此操作
// @Tags 空间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body space.RemoveCollaboratorRequest true "删除协作者请求参数"
// @Success 200 {object} string "删除协作者成功"
// @Failure 400 {object} string "协作者不存在"
// @Failure 403 {object} string "没有权限删除协作者"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/{id}/collaborators [delete]
func HandleRemoveCollaborator(c *fiber.Ctx) error {
	var req space.RemoveCollaboratorRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析删除协作者请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)
	s := c.Locals("space").(space.Space)

	if userID != s.OwnerID {
		return fiber.NewError(fiber.StatusForbidden, "只有空间所有者可以删除协作者")
	}

	if s.CollaboratorsMap == nil || s.CollaboratorsMap[req.CollaboratorID] == "" {
		return fiber.NewError(fiber.StatusBadRequest, "协作者不存在")
	}

	// 删除协作者
	delete(s.CollaboratorsMap, req.CollaboratorID)

	// 序列化更新后的协作者列表
	newCollaboratorsJSON, err := json.Marshal(s.CollaboratorsMap)
	if err != nil {
		logger.Error("序列化协作者列表失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "处理协作者数据失败")
	}

	// 更新数据库
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		_, err := db.ExecTx(tx, `
			UPDATE nlip_spaces 
			SET collaborators = ?
			WHERE id = ?
		`, string(newCollaboratorsJSON), s.ID)
		return err
	})

	if err != nil {
		logger.Error("更新协作者列表失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "删除协作者失败")
	}

	logger.Info("用户 %s 删除了协作者 %s 从空间 %s", userID, req.CollaboratorID, s.ID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "删除协作者成功",
		"data":    nil,
	})
}

// HandleUpdateCollaboratorPermissions 更新协作者权限
// @Summary 更新协作者权限
// @Description 更新指定协作者的权限，只有空间所有者可以执行此操作
// @Tags 空间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body space.UpdateCollaboratorPermissionsRequest true "更新协作者权限请求参数"
// @Success 200 {object} string "更新协作者权限成功"
// @Failure 400 {object} string "协作者不存在"
// @Failure 403 {object} string "没有权限更新协作者权限"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/{id}/collaborators/permissions [put]
func HandleUpdateCollaboratorPermissions(c *fiber.Ctx) error {
	var req space.UpdateCollaboratorPermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析更新协作者权限请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)

	s := c.Locals("space").(space.Space)

	if userID != s.OwnerID {
		logger.Warning("用户 %s 尝试更新协作者权限: spaceID=%s, collaboratorID=%s", userID, s.ID, req.CollaboratorID)
		return fiber.NewError(fiber.StatusForbidden, "只有空间所有者可以更新协作者权限")
	}

	if s.CollaboratorsMap == nil || s.CollaboratorsMap[req.CollaboratorID] == "" {
		return fiber.NewError(fiber.StatusBadRequest, "协作者不存在")
	}

	// 更新协作者权限
	s.CollaboratorsMap[req.CollaboratorID] = req.Permission

	// 序列化更新后的协作者列表
	newCollaboratorsJSON, err := json.Marshal(s.CollaboratorsMap)
	if err != nil {
		logger.Error("序列化协作者列表失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "处理协作者数据失败")
	}

	// 更新数据库
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		_, err := db.ExecTx(tx, `
			UPDATE nlip_spaces 
			SET collaborators = ?
			WHERE id = ?
		`, string(newCollaboratorsJSON), s.ID)
		return err
	})

	if err != nil {
		logger.Error("更新协作者权限失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "更新协作者权限失败")
	}

	logger.Info("用户 %s 更新了协作者 %s 的权限在空间 %s", userID, req.CollaboratorID, s.ID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "更新协作者权限成功",
		"data":    nil,
	})
}

// HandleUpdateSpaceSettings 更新空间设置
// @Summary 更新空间设置
// @Description 更新空间设置，普通用户只能更新自己拥有的私有空间，管理员可以更新所有空间
// @Tags 空间
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body space.UpdateSpaceSettingsRequest true "更新空间设置请求参数"
// @Success 200 {object} space.Space "更新成功，返回更新后的空间信息"
// @Failure 400 {object} string "请求参数错误"
// @Failure 403 {object} string "没有权限更新该空间设置"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/{id}/settings [put]
func HandleUpdateSpaceSettings(c *fiber.Ctx) error {
	var req space.UpdateSpaceSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析更新空间设置请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)
	s := c.Locals("space").(space.Space)

	if s.Type == "public" && !isAdmin {
		logger.Warning("非管理员用户 %s 尝试更新公共空间设置: spaceID=%s", userID, s.ID)
		return fiber.NewError(fiber.StatusForbidden, "没有权限更新公共空间设置")
	}

	if s.Type == "private" && userID != s.OwnerID {
		logger.Warning("用户 %s 尝试更新私有空间设置: spaceID=%s", userID, s.ID)
		return fiber.NewError(fiber.StatusForbidden, "只有空间所有者可以更新空间设置")
	}

	logger.Debug("处理更新空间设置请求: id=%s, userID=%s", s.ID, userID)

	// 更新字段
	if req.Name != "" {
		s.Name = req.Name
	}
	if req.MaxItems > 0 {
		s.MaxItems = req.MaxItems
	}
	if req.RetentionDays > 0 {
		s.RetentionDays = req.RetentionDays
	}
	if req.Visibility != "" {
		s.Type = req.Visibility
	}

	// 更新数据库
	err := db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		_, err := db.ExecTx(tx, `
			UPDATE nlip_spaces 
			SET name = ?, max_items = ?, retention_days = ?, type = ?, updated_at = ?
			WHERE id = ?
		`, s.Name, s.MaxItems, s.RetentionDays, s.Type, time.Now(), s.ID)
		return err
	})

	if err != nil {
		logger.Error("更新空间设置失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "更新空间设置失败")
	}

	logger.Info("用户 %s 更新了空间设置: id=%s, name=%s", userID, s.ID, s.Name)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "更新空间设置成功",
		"data":    s,
	})
}

// HandleSpaceStats 获取空间统计信息
// @Summary 获取空间统计信息
// @Description 获取空间的统计信息，包括剪贴板数量和所有者信息
// @Tags 空间
// @Produce json
// @Security BearerAuth
// @Success 200 {object} space.SpaceStatsResponse "获取统计信息成功"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/{id}/stats [get]
func HandleSpaceStats(c *fiber.Ctx) error {
	s := c.Locals("space").(space.Space)

	// 获取剪贴板数量
	var clipCount int
	var ownerUsername string
	err := config.DB.QueryRow(`
		SELECT COUNT(*) FROM nlip_clipboard_items 
		WHERE space_id = ?
	`, s.ID).Scan(&clipCount)

	if err != nil {
		logger.Error("获取空间统计信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取统计信息失败")
	}

	if s.Type == "private" && s.OwnerID != "" {
		err = config.DB.QueryRow(`
			SELECT username FROM nlip_users WHERE id = ?
		`, s.OwnerID).Scan(&ownerUsername)

		if err != nil {
			logger.Error("获取空间所有者用户名失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "获取空间所有者用户名失败")
		}
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取空间统计信息成功",
		"data": space.SpaceStatsResponse{
			ClipsCount:     clipCount,
			OwnerUsername:  ownerUsername,
		},
	})
}

// HandleListCollaborators 获取空间协作者列表
// @Summary 获取空间协作者列表
// @Description 获取指定空间的协作者列表，公共空间不支持此功能
// @Tags 空间
// @Produce json
// @Security BearerAuth
// @Success 200 {object} space.ListCollaboratorsResponse "获取协作者列表成功"
// @Failure 400 {object} string "公共空间不支持协作者功能"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/spaces/{id}/collaborators [get]
func HandleListCollaborators(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	s := c.Locals("space").(space.Space)

	if s.Type == "public" {
		return fiber.NewError(fiber.StatusBadRequest, "公共空间不支持协作者功能")
	}

	if s.Collaborators != nil {
		for i := range s.Collaborators {
			var username string
			err := config.DB.QueryRow(`
				SELECT username FROM nlip_users WHERE id = ?
			`, s.Collaborators[i].ID).Scan(&username)

			if err != nil {
				logger.Warning("获取协作者 %s 的用户信息失败: %v", s.Collaborators[i].ID, err)
				continue
			}
			s.Collaborators[i].Username = username
		}
	}

	logger.Info("用户 %s 获取了空间 %s 的协作者列表", userID, s.ID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取协作者列表成功",
		"data":    s.Collaborators,
	})
}
