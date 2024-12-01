package spaces

import (
    "database/sql"
    "github.com/gofiber/fiber/v2"
    "nlip/config"
    "nlip/models/space"
    "nlip/utils/logger"
    "nlip/utils/db"
    "nlip/utils/id"
    "time"
)

// HandleListSpaces 处理获取空间列表
func HandleListSpaces(c *fiber.Ctx) error {
    userID := c.Locals("userId").(string)
    isAdmin := c.Locals("isAdmin").(bool)

    logger.Debug("获取空间列表: userID=%s, isAdmin=%v", userID, isAdmin)

    var rows *sql.Rows
    var err error

    // 管理员可以看到所有空间，普通用户只能看到公共空间和自己的私有空间
    if isAdmin {
        rows, err = db.QueryRows(config.DB, `
            SELECT id, name, type, owner_id, max_items, retention_days, created_at, updated_at
            FROM nlip_spaces
            ORDER BY created_at DESC
        `)
    } else {
        rows, err = db.QueryRows(config.DB, `
            SELECT id, name, type, owner_id, max_items, retention_days, created_at, updated_at
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
        err := rows.Scan(
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
            logger.Error("读取空间数据失败: %v", err)
            return fiber.NewError(fiber.StatusInternalServerError, "读取空间数据失败")
        }
        spaces = append(spaces, s)
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
    }

    // 插入数据库
    err := db.WithTransaction(config.DB, func(tx *sql.Tx) error {
        _, err := db.ExecTx(tx, `
            INSERT INTO nlip_spaces (id, name, type, owner_id, max_items, retention_days, created_at, updated_at) 
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `, s.ID, s.Name, s.Type, s.OwnerID, s.MaxItems, s.RetentionDays, s.CreatedAt, s.UpdatedAt)
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

    // 更新数据库
    err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
        _, err := db.ExecTx(tx, `
            UPDATE nlip_spaces 
            SET name = ?, max_items = ?, retention_days = ?
            WHERE id = ?
        `, s.Name, s.MaxItems, s.RetentionDays, s.ID)
        return err
    })

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
        "code": fiber.StatusOK,
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
        _, err := db.ExecTx(tx, "DELETE FROM nlip_clipboard_items WHERE space_id = ?", spaceID)
        if err != nil {
            return err
        }

        // 删除空间
        _, err = db.ExecTx(tx, "DELETE FROM nlip_spaces WHERE id = ?", spaceID)
        return err
    })

    if err != nil {
        logger.Error("删除空间失败: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "删除空间失败")
    }

    logger.Info("用户 %s 删除了空间: id=%s", userID, spaceID)
    return c.JSON(fiber.Map{
        "code": fiber.StatusOK,
        "data": nil,
        "message": "删除空间成功",
    })
} 