package auth

import (
	"nlip/config"
	"nlip/models/user"
	"nlip/utils/jwt"
	"nlip/utils/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
		"SELECT id, username, password_hash, is_admin, created_at, need_change_pwd FROM nlip_users WHERE username = ?",
		req.Username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.NeedChangePwd)

	if err != nil {
		logger.Warning("用户名不存在: %s", req.Username)
		return fiber.NewError(fiber.StatusUnauthorized, "用户名不存在")
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

	logger.Info("用户登录成功: username=%s, id=%s, needChangePwd=%t", u.Username, u.ID, u.NeedChangePwd)
	return c.JSON(fiber.Map{
		"code": fiber.StatusOK,
		"message": "登录成功",
		"data": user.AuthResponse{
			Token:         token,
			User:          &u,
			NeedChangePwd: u.NeedChangePwd,
		},
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
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"code": fiber.StatusCreated,
		"message": "注册成功",
		"data": user.AuthResponse{
			Token: token,
			User:  &u,
		},
	})
}

// HandleGetCurrentUser 获取当前登录用户信息
func HandleGetCurrentUser(c *fiber.Ctx) error {
	// 从context中获取用户信息
	userRaw := c.Locals("user")
	if userRaw == nil {
		logger.Warning("获取当前用户失败：用户未登录")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    fiber.StatusUnauthorized,
			"message": "用户未登录",
			"data":    nil,
		})
	}

	user, ok := userRaw.(*jwt.UserClaims)
	if !ok {
		logger.Error("获取当前用户失败：用户信息格式错误")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": "用户信息格式错误",
			"data":    nil,
		})
	}

	logger.Info("成功获取当前用户信息: userId=%s, username=%s", user.UserID, user.Username)
	return c.JSON(fiber.Map{
		"code": fiber.StatusOK,
		"message": "获取用户信息成功",
		"data": fiber.Map{
			"id":       user.UserID,
			"username": user.Username,
			"isAdmin":  user.IsAdmin,
		},
	})
}

// HandleChangePassword 处理修改密码请求
func HandleChangePassword(c *fiber.Ctx) error {
	var req user.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析修改密码请求失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	userID := c.Locals("userId").(string)

	// 获取用户信息
	var u user.User
	err := config.DB.QueryRow(`
        SELECT id, username, password_hash, is_admin, need_change_pwd 
        FROM nlip_users WHERE id = ?
    `, userID).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.IsAdmin, &u.NeedChangePwd)

	if err != nil {
		logger.Error("获取用户信息失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取用户信息失败")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)); err != nil {
		logger.Warning("旧密码验证失败: username=%s", u.Username)
		return fiber.NewError(fiber.StatusUnauthorized, "旧密码错误")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("生成密码哈希失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "密码加密失败")
	}

	// 更新密码和状态
	_, err = config.DB.Exec(`
        UPDATE nlip_users 
        SET password_hash = ?, need_change_pwd = FALSE 
        WHERE id = ?
    `, string(hashedPassword), userID)

	if err != nil {
		logger.Error("更新密码失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "更新密码失败")
	}

	logger.Info("用户 %s 成功修改密码", u.Username)
	return c.JSON(fiber.Map{
		"code": fiber.StatusOK,
		"message": "密码修改成功",
		"data": nil,
	})
}
