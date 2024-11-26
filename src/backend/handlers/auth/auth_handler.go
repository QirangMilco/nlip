package auth

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "nlip/config"
    "nlip/models/user"
    "nlip/utils/jwt"
    "nlip/utils/logger"
)

// HandleLogin 处理登录请求
func HandleLogin(c *fiber.Ctx) error {
    var req user.LoginRequest
    if err := c.BodyParser(&req); err != nil {
        logger.Warning("解析登录请求失败: %v", err)
        return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
    }

    logger.Debug("处理登录请求: username=%s", req.Username)

    // 查找用户
    var u user.User
    err := config.DB.QueryRow(
        "SELECT id, username, password_hash, is_admin, created_at FROM nlip_users WHERE username = ?",
        req.Username,
    ).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt)

    if err != nil {
        logger.Warning("用户名不存在: %s", req.Username)
        return fiber.NewError(fiber.StatusUnauthorized, "用户名或密码错误")
    }

    // 验证密码
    if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
        logger.Warning("密码验证失败: username=%s", req.Username)
        return fiber.NewError(fiber.StatusUnauthorized, "用户名或密码错误")
    }

    // 生成令牌
    token, err := jwt.GenerateToken(&u)
    if err != nil {
        logger.Error("生成令牌失败: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "生成令牌失败")
    }

    logger.Info("用户登录成功: username=%s, id=%s", u.Username, u.ID)
    return c.JSON(user.AuthResponse{
        Token: token,
        User:  &u,
    })
}

// HandleRegister 处理注册请求
func HandleRegister(c *fiber.Ctx) error {
    var req user.RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        logger.Warning("解析注册请求失败: %v", err)
        return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
    }

    logger.Debug("处理注册请求: username=%s", req.Username)

    // 检查用户名是否已存在
    var exists bool
    err := config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM nlip_users WHERE username = ?)", req.Username).Scan(&exists)
    if err != nil {
        logger.Error("检查用户名存在性失败: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "数据库查询错误")
    }
    if exists {
        logger.Warning("用户名已存在: %s", req.Username)
        return fiber.NewError(fiber.StatusConflict, "用户名已存在")
    }

    // 生成密码哈希
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        logger.Error("生成密码哈希失败: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "密码加密失败")
    }

    // 创建用户
    u := user.User{
        ID:           uuid.New().String(),
        Username:     req.Username,
        PasswordHash: string(hashedPassword),
        IsAdmin:      false,
    }

    // 插入数据库
    _, err = config.DB.Exec(
        "INSERT INTO nlip_users (id, username, password_hash, is_admin) VALUES (?, ?, ?, ?)",
        u.ID, u.Username, u.PasswordHash, u.IsAdmin,
    )
    if err != nil {
        logger.Error("创建用户失败: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "创建用户失败")
    }

    // 生成令牌
    token, err := jwt.GenerateToken(&u)
    if err != nil {
        logger.Error("生成令牌失败: %v", err)
        return fiber.NewError(fiber.StatusInternalServerError, "生成令牌失败")
    }

    logger.Info("用户注册成功: username=%s, id=%s", u.Username, u.ID)
    return c.Status(fiber.StatusCreated).JSON(user.AuthResponse{
        Token: token,
        User:  &u,
    })
}

// HandleGetCurrentUser 获取当前登录用户信息
func HandleGetCurrentUser(c *fiber.Ctx) error {
    // 从context中获取用户信息(之前在auth中间件中设置)
    user := c.Locals("user").(*jwt.UserClaims)
    
    return c.JSON(fiber.Map{
        "code": 200,
        "data": fiber.Map{
            "id": user.UserID,
            "username": user.Username,
            "isAdmin": user.IsAdmin,
        },
        "message": "获取用户信息成功",
    })
} 