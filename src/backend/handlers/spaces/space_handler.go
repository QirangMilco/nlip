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

// HandleListSpaces 处理获取空间列表
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
            SELECT id, name, type, owner_id, max_items, retention_days, invited_users, created_at, updated_at
            FROM nlip_spaces 
            WHERE type = 'public'
            ORDER BY created_at DESC
        `)
	} else if isAdmin {
		// 管理员可以看到所有空间
		rows, err = db.QueryRows(config.DB, `
            SELECT id, name, type, owner_id, max_items, retention_days, invited_users, created_at, updated_at
            FROM nlip_spaces
            ORDER BY created_at DESC
        `)
	} else {
		// 已认证的普通用户可以看到公共空间和自己的私有空间
		rows, err = db.QueryRows(config.DB, `
            SELECT id, name, type, owner_id, max_items, retention_days, invited_users, created_at, updated_at
            FROM nlip_spaces 
            WHERE type = 'public' OR owner_id = ?
            ORDER BY created_at DESC
        `, userID)
	}

	if err != nil {
		logger.Error("获取空间列表失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取空间列表失败")
	}
	defer rows.Close()

	var spaces []space.Space
	for rows.Next() {
		var s space.Space
		var invitedUsersJSON string
		err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Type,
			&s.OwnerID,
			&s.MaxItems,
			&s.RetentionDays,
			&invitedUsersJSON,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			logger.Error("读取空间数据失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "读取空间数据失败")
		}

		s.InvitedUsers = []space.CollaboratorInfo{}

		// 解析协作者信息
		if invitedUsersJSON != "" {
			if err := json.Unmarshal([]byte(invitedUsersJSON), &s.InvitedUsers); err != nil {
				logger.Error("解析协作者列表失败: %v", err)
				continue
			}

			// 查询协作者详细信息
			for _, collaborator := range s.InvitedUsers {
				var username, email string
				err := config.DB.QueryRow(`
					SELECT username, email FROM nlip_users WHERE id = ?
				`, collaborator.UserID).Scan(&username, &email)

				if err != nil {
					logger.Warning("获取协作者信息失败: %v", err)
					continue
				}

				s.InvitedUsers = append(s.InvitedUsers, space.CollaboratorInfo{
					UserID:     collaborator.UserID,
					Username:   username,
					Email:      email,
					Permission: collaborator.Permission,
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

// HandleCreateSpace 处理创建空间
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
		InvitedUsers:  req.InvitedUsers,
	}

	// 将 InvitedUsers 转换为 JSON 字符串
	invitedUsersJSON, err := json.Marshal(s.InvitedUsers)
	if err != nil {
		logger.Error("序列化邀请用户数据失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "创建空间失败")
	}

	// 插入数据库
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		_, err = db.ExecTx(tx, `
            INSERT INTO nlip_spaces (id, name, type, owner_id, max_items, retention_days, invited_users, created_at, updated_at) 
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, s.ID, s.Name, s.Type, s.OwnerID, s.MaxItems, s.RetentionDays, string(invitedUsersJSON), s.CreatedAt, s.UpdatedAt)
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

// HandleUpdateSpace 处理更新空间
func HandleUpdateSpace(c *fiber.Ctx) error {
	spaceID := c.Params("id")
	var req space.UpdateSpaceRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析更新空间请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	logger.Debug("处理更新空间请求: id=%s, userID=%s", spaceID, userID)

	// 检查空间是否存在并获取当前信息
	var s space.Space
	err := config.DB.QueryRow(`
        SELECT id, name, type, owner_id, max_items, retention_days, created_at, updated_at
        FROM nlip_spaces WHERE id = ?
    `, spaceID).Scan(
		&s.ID,
		&s.Name,
		&s.Type,
		&s.OwnerID,
		&s.MaxItems,
		&s.RetentionDays,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		logger.Warning("尝试更新不存在的空间: %s", spaceID)
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("获取空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取空间信息失败")
	}

	// 检查权限
	if !isAdmin && s.OwnerID != userID {
		logger.Warning("用户 %s 尝试更新无权限的空间 %s", userID, spaceID)
		return fiber.NewError(fiber.StatusForbidden, "没有权限修改此空间")
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

	if req.InvitedUsers != nil {
		invitedUsersJSON, jsonErr := json.Marshal(req.InvitedUsers)
		if jsonErr != nil {
			logger.Error("序列化邀请用户数据失败: %v", jsonErr)
			return fiber.NewError(fiber.StatusInternalServerError, "更新空间失败")
		}

		err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
			var execErr error
			_, execErr = db.ExecTx(tx, `
                UPDATE nlip_spaces 
                SET name = ?, max_items = ?, retention_days = ?, invited_users = ?, updated_at = ?
                WHERE id = ?
            `, s.Name, s.MaxItems, s.RetentionDays, string(invitedUsersJSON), time.Now(), s.ID)
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
    `, spaceID).Scan(
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
		"data": space.SpaceResponse{
			Space: &s,
		},
	})
}

// HandleDeleteSpace 删除空间
func HandleDeleteSpace(c *fiber.Ctx) error {
	spaceID := c.Params("id")
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	logger.Debug("处理删除空间请求: id=%s, userID=%s", spaceID, userID)

	// 检查空间是否存在并获取所有者信息
	var ownerID string
	var spaceType string
	err := config.DB.QueryRow("SELECT owner_id, type FROM nlip_spaces WHERE id = ?", spaceID).Scan(&ownerID, &spaceType)

	if err == sql.ErrNoRows {
		logger.Warning("尝试删除不存在的空间: %s", spaceID)
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("获取空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取空间信息失败")
	}

	// 检查权限
	if !isAdmin && ownerID != userID {
		logger.Warning("用户 %s 尝试删除无权限的空间 %s", userID, spaceID)
		return fiber.NewError(fiber.StatusForbidden, "没有权限删除此空间")
	}

	// 开始事务
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 删除空间内的所有内容
		var execErr error
		_, execErr = db.ExecTx(tx, "DELETE FROM nlip_clipboard_items WHERE space_id = ?", spaceID)
		if execErr != nil {
			return execErr
		}

		// 删除空间
		_, execErr = db.ExecTx(tx, "DELETE FROM nlip_spaces WHERE id = ?", spaceID)
		return execErr
	})

	if err != nil {
		logger.Error("删除空间失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "删除空间失败")
	}

	logger.Info("用户 %s 删除了空间: id=%s", userID, spaceID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"data":    nil,
		"message": "删除空间成功",
	})
}

// HandleInviteCollaborator 处理邀请协作者
func HandleInviteCollaborator(c *fiber.Ctx) error {
	spaceID := c.Params("id")
	userID := c.Locals("userId").(string)

	var req space.InviteCollaboratorRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析邀请协作者请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	// 验证空间所有权
	var ownerID string
	var spaceName string
	err := config.DB.QueryRow("SELECT owner_id, name FROM nlip_spaces WHERE id = ?", spaceID).Scan(&ownerID, &spaceName)
	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("查询空间失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "服务器错误")
	}

	logger.Debug("空间所有者ID: %s, 用户ID: %s", ownerID, userID)
	if ownerID != userID {
		return fiber.NewError(fiber.StatusForbidden, "只有空间所有者可以邀请协作者")
	}

	// 生成唯一的令牌哈希值
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s_%s_%d", spaceID, userID, time.Now().UnixNano())))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	// 存储邀请信息到数据库
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		_, err := db.ExecTx(tx, `
			INSERT INTO nlip_invites (
				token_hash, 
				space_id, 
				created_by, 
				created_at,
				expires_at,
				permission
			) VALUES (?, ?, ?, datetime('now'), datetime('now', '+24 hours'), ?)
		`, tokenHash, spaceID, userID, req.Permission)
		return err
	})

	if err != nil {
		logger.Error("存储邀请令牌失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "生成邀请链接失败")
	}

	inviteLink := fmt.Sprintf("%s/invite/%s", config.AppConfig.FrontendURL, tokenHash)

	if config.AppConfig.Email.Enabled {
		// 发送邀请邮件（如果启用了邮件功能）
		if err = email.SendInviteEmail(req.Email, spaceName, inviteLink); err != nil {
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
func HandleVerifyInviteToken(c *fiber.Ctx) error {
	var req space.ValidateInviteRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析验证邀请令牌请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)

	// 从数据库中查询邀请信息
	var spaceID, createdBy, permission string
	err := config.DB.QueryRow(`
		SELECT space_id, created_by, permission
			FROM nlip_invites 
			WHERE token_hash = ? 
				AND used_at IS NULL 
				AND expires_at > datetime('now')
	`, req.Token).Scan(&spaceID, &createdBy, &permission)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusBadRequest, "无效的邀请链接或已过期")
	} else if err != nil {
		logger.Error("验证邀请令牌失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "服务器错误")
	}

	// 检查空间是否存在并获取协作者信息
	var spaceName, inviterName string
	var invitedUsersJSON string
	err = config.DB.QueryRow(`
		SELECT s.name, u.username, COALESCE(s.invited_users, '{}')
			FROM nlip_spaces s 
			JOIN nlip_users u ON u.id = ?
			WHERE s.id = ?
	`, createdBy, spaceID).Scan(&spaceName, &inviterName, &invitedUsersJSON)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("获取空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取空间信息失败")
	}

	// 检查用户是否已经是协作者
	var invitedUsers = make(map[string]string)
	if invitedUsersJSON != "" && invitedUsersJSON != "{}" {
		if err := json.Unmarshal([]byte(invitedUsersJSON), &invitedUsers); err != nil {
			logger.Error("解析协作者列表失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "服务器错误")
		}
	}

	isCollaborator := false
	currentPermission := ""
	if existingPermission, exists := invitedUsers[userID]; exists {
		isCollaborator = true
		currentPermission = existingPermission
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "验证成功",
		"data": space.VerifyInviteTokenResponse{
			SpaceID:           spaceID,
			SpaceName:         spaceName,
			InviterName:       inviterName,
			Permission:        permission,
			IsCollaborator:    isCollaborator,
			CurrentPermission: currentPermission,
		},
	})
}

// HandleAcceptInvite 处理接受邀请
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
			AND expires_at > datetime('now')
	`, req.Token).Scan(&spaceID, &permission)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusBadRequest, "邀请链接无效或已过期")
	} else if err != nil {
		logger.Error("验证邀请令牌失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "服务器错误")
	}

	// 检查空间是否存在
	var exists bool
	err = config.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM nlip_spaces WHERE id = ?)
	`, spaceID).Scan(&exists)

	if err != nil {
		logger.Error("检查空间存在性失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "服务器错误")
	}

	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	}

	// 检查用户是否已经是协作者
	var invitedUsersJSON string
	err = config.DB.QueryRow(`
		SELECT COALESCE(invited_users, '{}') 
		FROM nlip_spaces 
		WHERE id = ?
	`, spaceID).Scan(&invitedUsersJSON)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("获取空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "服务器错误")
	}

	// 初始化 invitedUsers map
	invitedUsers := make(map[string]string)

	// 检查用户是否已经是协作者
	if invitedUsersJSON != "" && invitedUsersJSON != "{}" {
		if err := json.Unmarshal([]byte(invitedUsersJSON), &invitedUsers); err != nil {
			logger.Error("解析协作者列表失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "服务器错误")
		}
	}

	if _, exists := invitedUsers[userID]; exists {
		return fiber.NewError(fiber.StatusBadRequest, "您已经是该空间的协作者")
	}

	// 开启事务
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 标记令牌为已使用
		_, err = db.ExecTx(tx, `
			UPDATE nlip_invites 
			SET used_at = datetime('now'),
				used_by = ?
			WHERE token_hash = ?
		`, userID, req.Token)
		if err != nil {
			return err
		}

		// 更新协作者列表
		if invitedUsers == nil {
			invitedUsers = make(map[string]string)
		}
		invitedUsers[userID] = permission
		newInvitedUsersJSON, err := json.Marshal(invitedUsers)
		if err != nil {
			return err
		}

		_, err = db.ExecTx(tx, `
			UPDATE nlip_spaces 
			SET invited_users = ?, 
				updated_at = datetime('now')
			WHERE id = ?
		`, string(newInvitedUsersJSON), spaceID)

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
	})
}

// HandleRemoveCollaborator 处理删除协作者
func HandleRemoveCollaborator(c *fiber.Ctx) error {
	spaceID := c.Params("id")
	var req space.RemoveCollaboratorRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析删除协作者请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	// 检查空间是否存在并获取当前信息
	var s space.Space
	var invitedUsersJSON string
	err := config.DB.QueryRow(`
		SELECT id, name, type, owner_id, invited_users
		FROM nlip_spaces WHERE id = ?
	`, spaceID).Scan(&s.ID, &s.Name, &s.Type, &s.OwnerID, &invitedUsersJSON)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("获取空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取空间信息失败")
	}

	// 检查权限
	if !isAdmin && s.OwnerID != userID {
		return fiber.NewError(fiber.StatusForbidden, "没有权限删除协作者")
	}

	// 解析现有的协作者列表
	invitedUsers := make(map[string]string)
	if invitedUsersJSON != "" {
		if err := json.Unmarshal([]byte(invitedUsersJSON), &invitedUsers); err != nil {
			logger.Error("解析协作者列表失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "处理协作者数据失败")
		}
	}

	// 删除协作者
	delete(invitedUsers, req.CollaboratorID)

	// 序列化更新后的协作者列表
	newInvitedUsersJSON, err := json.Marshal(invitedUsers)
	if err != nil {
		logger.Error("序列化协作者列表失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "处理协作者数据失败")
	}

	// 更新数据库
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		_, err := db.ExecTx(tx, `
			UPDATE nlip_spaces 
			SET invited_users = ?
			WHERE id = ?
		`, string(newInvitedUsersJSON), spaceID)
		return err
	})

	if err != nil {
		logger.Error("更新协作者列表失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "删除协作者失败")
	}

	logger.Info("用户 %s 删除了协作者 %s 从空间 %s", userID, req.CollaboratorID, spaceID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "删除协作者成功",
	})
}

// HandleUpdateCollaboratorPermissions 处理更新协作者权限
func HandleUpdateCollaboratorPermissions(c *fiber.Ctx) error {
	spaceID := c.Params("id")
	var req space.UpdateCollaboratorPermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析更新协作者权限请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	// 检查空间是否存在并获取当前信息
	var s space.Space
	var invitedUsersJSON string
	err := config.DB.QueryRow(`
		SELECT id, name, type, owner_id, invited_users
		FROM nlip_spaces WHERE id = ?
	`, spaceID).Scan(&s.ID, &s.Name, &s.Type, &s.OwnerID, &invitedUsersJSON)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("获取空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取空间信息失败")
	}

	// 检查权限
	if !isAdmin && s.OwnerID != userID {
		return fiber.NewError(fiber.StatusForbidden, "没有权限更新协作者权限")
	}

	// 解析现有的协作者列表
	invitedUsers := make(map[string]string)
	if invitedUsersJSON != "" {
		if err := json.Unmarshal([]byte(invitedUsersJSON), &invitedUsers); err != nil {
			logger.Error("解析协作者列表失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "处理协作者数据失败")
		}
	}

	// 检查协作者是否存在
	if _, exists := invitedUsers[req.CollaboratorID]; !exists {
		return fiber.NewError(fiber.StatusBadRequest, "协作者不存在")
	}

	// 更新协作者权限
	invitedUsers[req.CollaboratorID] = req.Permission

	// 序列化更新后的协作者列表
	newInvitedUsersJSON, err := json.Marshal(invitedUsers)
	if err != nil {
		logger.Error("序列化协作者列表失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "处理协作者数据失败")
	}

	// 更新数据库
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		_, err := db.ExecTx(tx, `
			UPDATE nlip_spaces 
			SET invited_users = ?
			WHERE id = ?
		`, string(newInvitedUsersJSON), spaceID)
		return err
	})

	if err != nil {
		logger.Error("更新协作者权限失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "更新协作者权限失败")
	}

	logger.Info("用户 %s 更新了协作者 %s 的权限在空间 %s", userID, req.CollaboratorID, spaceID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "更新协作者权限成功",
	})
}

// HandleUpdateSpaceSettings 处理更新空间设置
func HandleUpdateSpaceSettings(c *fiber.Ctx) error {
	spaceID := c.Params("id")
	var req space.UpdateSpaceSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析更新空间设置请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	logger.Debug("处理更新空间设置请求: id=%s, userID=%s", spaceID, userID)

	// 检查空间是否存在并获取当前信息
	var s space.Space
	err := config.DB.QueryRow(`
		SELECT id, name, type, owner_id, max_items, retention_days, invited_users
		FROM nlip_spaces WHERE id = ?
	`, spaceID).Scan(&s.ID, &s.Name, &s.Type, &s.OwnerID, &s.MaxItems, &s.RetentionDays, &s.InvitedUsers)

	if err == sql.ErrNoRows {
		logger.Warning("尝试更新不存在的空间设置: %s", spaceID)
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("获取空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取空间信息失败")
	}

	// 检查权限
	if !isAdmin && s.OwnerID != userID {
		logger.Warning("用户 %s 尝试更新无权限的空间设置 %s", userID, spaceID)
		return fiber.NewError(fiber.StatusForbidden, "没有权限修改此空间设置")
	}

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
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
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
		"data": space.SpaceResponse{
			Space: &s,
		},
	})
}

// HandlePublicSpaceStats 处理获取公共空间统计信息
func HandlePublicSpaceStats(c *fiber.Ctx) error {
	// 获取公共空间的剪贴板数量
	var clipCount int
	err := config.DB.QueryRow(`
		SELECT COUNT(*) FROM nlip_clipboard_items 
		WHERE space_id IN (SELECT id FROM nlip_spaces WHERE type = 'public')
	`).Scan(&clipCount)

	if err != nil {
		logger.Error("获取公共空间统计信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取统计信息失败")
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取公共空间统计信息成功",
		"data": fiber.Map{
			"clipCount": clipCount,
		},
	})
}

// HandleSpaceStats 处理获取空间统计信息
func HandleSpaceStats(c *fiber.Ctx) error {
	spaceID := c.Params("id")
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	// 检查空间是否存在并获取基本信息
	var s space.Space
	err := config.DB.QueryRow(`
		SELECT id, name, type, owner_id
		FROM nlip_spaces WHERE id = ?
	`, spaceID).Scan(&s.ID, &s.Name, &s.Type, &s.OwnerID)

	if err == sql.ErrNoRows {
		logger.Warning("尝试获取不存在的空间统计信息: %s", spaceID)
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("获取空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取统计信息失败")
	}

	// 检查权限
	if s.Type != "public" && !isAdmin && s.OwnerID != userID {
		logger.Warning("用户 %s 尝试获取无权限的空间统计信息 %s", userID, spaceID)
		return fiber.NewError(fiber.StatusForbidden, "没有权限查看此空间统计信息")
	}

	// 获取剪贴板数量
	var clipCount int
	err = config.DB.QueryRow(`
		SELECT COUNT(*) FROM nlip_clipboard_items 
		WHERE space_id = ?
	`, spaceID).Scan(&clipCount)

	if err != nil {
		logger.Error("获取空间统计信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取统计信息失败")
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取空间统计信息成功",
		"data": fiber.Map{
			"clipCount": clipCount,
		},
	})
}

// HandleListCollaborators 处理获取空间协作者列表
func HandleListCollaborators(c *fiber.Ctx) error {
	spaceID := c.Params("id")
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	// 检查空间是否存在并获取当前信息
	var s space.Space
	var invitedUsersJSON string
	err := config.DB.QueryRow(`
		SELECT id, name, type, owner_id, invited_users
		FROM nlip_spaces WHERE id = ?
	`, spaceID).Scan(&s.ID, &s.Name, &s.Type, &s.OwnerID, &invitedUsersJSON)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("获取空间信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取空间信息失败")
	}

	// 检查权限
	if s.Type != "public" && !isAdmin && s.OwnerID != userID {
		return fiber.NewError(fiber.StatusForbidden, "没有权限查看此空间的协作者")
	}

	// 解析协作者列表
	invitedUsers := make(map[string]string)
	if invitedUsersJSON != "" {
		if err := json.Unmarshal([]byte(invitedUsersJSON), &invitedUsers); err != nil {
			logger.Error("解析协作者列表失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "处理协作者数据失败")
		}
	}

	// 获取协作者详细信息
	var collaborators []space.CollaboratorInfo
	for collaboratorID, permission := range invitedUsers {
		var username string
		err := config.DB.QueryRow(`
			SELECT username FROM nlip_users WHERE id = ?
		`, collaboratorID).Scan(&username)

		if err != nil {
			logger.Warning("获取协作者 %s 的用户信息失败: %v", collaboratorID, err)
			continue
		}

		collaborators = append(collaborators, space.CollaboratorInfo{
			UserID:     collaboratorID,
			Username:   username,
			Permission: permission,
		})
	}

	logger.Info("用户 %s 获取了空间 %s 的协作者列表", userID, spaceID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取协作者列表成功",
		"data": space.ListCollaboratorsResponse{
			Collaborators: collaborators,
		},
	})
}
