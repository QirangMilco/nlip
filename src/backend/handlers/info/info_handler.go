package info

import (
    "database/sql"
    "github.com/gofiber/fiber/v2"
    "nlip/config"
    "nlip/utils/logger"
)

// HandleGetCurrentUserInfo 获取当前用户信息
func HandleGetCurrentUserInfo(c *fiber.Ctx) error {
    userID := c.Locals("userId").(string)
    
    // 从数据库查询完整的用户信息
    var username string
    var isAdmin bool
    var createdAt string
    
    err := config.DB.QueryRow(`
        SELECT username, is_admin, created_at 
        FROM nlip_users 
        WHERE id = ?
    `, userID).Scan(&username, &isAdmin, &createdAt)

    if err == sql.ErrNoRows {
        logger.Warning("用户不存在: %s", userID)
        return fiber.NewError(fiber.StatusNotFound, "用户不存在")
    } else if err != nil {
        logger.Error("查询用户信息失败: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "获取用户信息失败")
    }

    logger.Debug("获取用户信息成功: userID=%s, username=%s", userID, username)
    
    return c.JSON(fiber.Map{
        "code": 200,
        "data": fiber.Map{
            "id": userID,
            "username": username,
            "isAdmin": isAdmin,
            "createdAt": createdAt,
        },
        "message": "获取用户信息成功",
    })
} 