package clips

import (
	"database/sql"
	"fmt"
	"nlip/config"
	"nlip/models/clip"
	"nlip/models/space"
	"nlip/utils/db"
	"nlip/utils/id"
	"nlip/utils/logger"
	"nlip/utils/storage"
	"nlip/utils/validator"
	"path/filepath"
	"time"

	"mime/multipart"

	"nlip/tasks/cleaner"

	"github.com/gofiber/fiber/v2"
)

// 添加常量定义
const (
	// 空间类型
	SpaceTypePublic  = "public"
	SpaceTypePrivate = "private"

	PublicSpaceID = "public-space"
	GuestUserID   = "guest"

	// 错误消息
	ErrNoPermission     = "没有权限执行此操作"
	ErrSpaceNotFound    = "空间不存在"
	ErrClipNotFound     = "剪贴板内容不存在"
	ErrInvalidRequest   = "无效的请求数据"
	ErrFileUploadFailed = "文件上传失败"
)

// 添加 SQL 查询常量
const (
	// 基础查询语句
	selectClipWithCreatorSQL = `
        SELECT 
            c.id, 
            c.clip_id, 
            c.space_id, 
            c.content_type, 
            c.content, 
            c.file_path, 
            c.created_at,
            c.updated_at,
            u.id as creator_id,
            CASE WHEN u.id = 'guest' THEN '游客' ELSE u.username END as creator_username
        FROM nlip_clipboard_items c
        LEFT JOIN nlip_users u ON c.creator_id = u.id
    `
)

// 添加权限检查辅助函数
func checkClipPermission(tx *sql.Tx, spaceID, spaceType, clipID, userID string, isAdmin bool) (string, error) {
	var creatorID string

	// 只查询剪贴板项目的创建者
	if clipID != "" {
		err := db.QueryRowTx(tx, `
			SELECT creator_id 
			FROM nlip_clipboard_items 
			WHERE clip_id = ? AND space_id = ?
		`, clipID, spaceID).Scan(&creatorID)

		if err == sql.ErrNoRows {
			return "", fiber.NewError(fiber.StatusNotFound, ErrClipNotFound)
		} else if err != nil {
			return "", err
		}

		// 在公共空间中，只有管理员和创建者可以修改剪贴板项目
		if spaceType == SpaceTypePublic && !isAdmin && userID != creatorID {
			return "", fiber.NewError(fiber.StatusForbidden, ErrNoPermission)
		}
	}

	return creatorID, nil
}

// 添加文件处理辅助函数
func handleFileUpload(file *multipart.FileHeader) ([]byte, string, error) {
	// 检查文件大小
	if file.Size > config.AppConfig.MaxFileSize {
		return nil, "", fiber.NewError(fiber.StatusBadRequest, "文件大小超过限制")
	}

	// 验证文件名和类型
	if !validator.ValidateFileName(file.Filename) {
		return nil, "", fiber.NewError(fiber.StatusBadRequest, "文件名不合法")
	}

	contentType := file.Header.Get("Content-Type")
	if !validator.ValidateFileType(file.Filename, contentType) {
		return nil, "", fiber.NewError(fiber.StatusBadRequest, "不支持的文件类型")
	}

	// 读取文件数据
	src, err := file.Open()
	if err != nil {
		return nil, "", err
	}
	defer src.Close()

	data := make([]byte, file.Size)
	if _, err := src.Read(data); err != nil {
		return nil, "", err
	}

	return data, contentType, nil
}

// scanClip 辅助函数：扫描剪贴板数据
func scanClip(rows *sql.Rows) (*clip.Clip, error) {
	var cl clip.Clip
	var creatorID, creatorUsername, content sql.NullString
	var filePath sql.NullString

	err := rows.Scan(
		&cl.ID,
		&cl.ClipID,
		&cl.SpaceID,
		&cl.ContentType,
		&content,
		&filePath,
		&cl.CreatedAt,
		&cl.UpdatedAt,
		&creatorID,
		&creatorUsername,
	)

	if err != nil {
		return nil, err
	}

	if content.Valid {
		cl.Content = content.String
	}

	if filePath.Valid {
		cl.FilePath = filePath.String
	}

	if creatorID.Valid && creatorUsername.Valid {
		cl.Creator = &clip.Creator{
			ID:       creatorID.String,
			Username: creatorUsername.String,
		}
	}

	return &cl, nil
}

// 添加新的扫描函数用于单行结果
func scanSingleClip(row *sql.Row) (*clip.Clip, error) {
	var cl clip.Clip
	var creatorID, creatorUsername sql.NullString
	var filePath sql.NullString

	err := row.Scan(
		&cl.ID,
		&cl.ClipID,
		&cl.SpaceID,
		&cl.ContentType,
		&cl.Content,
		&filePath,
		&cl.CreatedAt,
		&cl.UpdatedAt,
		&creatorID,
		&creatorUsername,
	)

	if err != nil {
		return nil, err
	}

	if filePath.Valid {
		cl.FilePath = filePath.String
	}

	if creatorID.Valid && creatorUsername.Valid {
		cl.Creator = &clip.Creator{
			ID:       creatorID.String,
			Username: creatorUsername.String,
		}
	}

	return &cl, nil
}

// HandleUploadClip 处理上传剪贴板内容
// @Summary 上传Clip
// @Description 上传剪贴板内容
// @Tags 剪贴板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body clip.UploadClipRequest true "上传Clip请求参数"
// @Success 200 {object} clip.ClipResponse "上传成功"
// @Failure 400 {object} string "请求参数错误"
// @Failure 401 {object} string "未授权"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/clips [post]
func HandleUploadClip(c *fiber.Ctx) error {
	var req clip.UploadClipRequest
	userID := GuestUserID
	username := "游客"
	isAdmin := false
	if c.Locals("userId") != nil {
		userID = c.Locals("userId").(string)
		username = c.Locals("username").(string)
		isAdmin = c.Locals("isAdmin").(bool)
	}
	
	s := c.Locals("space").(space.Space)

	// 解析表单数据
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrInvalidRequest)
	}

	if req.SpaceID != s.ID {
		return fiber.NewError(fiber.StatusBadRequest, "空间ID不匹配，req.SpaceID="+req.SpaceID+", s.ID="+s.ID)
	}

	var err error

	// 非公共空间需要检查权限
	if s.Type != SpaceTypePublic {
		err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
			_, err := checkClipPermission(tx, req.SpaceID, s.Type, "", userID, isAdmin)
			return err
		})
		if err != nil {
			return err
		}
	}

	// 处理文件上传
	var fileData []byte
	var fileName string
	var contentType string
	if form, err := c.MultipartForm(); err == nil && form.File != nil {
		if files := form.File["file"]; len(files) > 0 {
			file := files[0]
			logger.Debug("处理文件上传: %s, 大小: %d bytes", file.Filename, file.Size)

			data, cType, err := handleFileUpload(file)
			if err != nil {
				logger.Error("处理文件上传失败: %v", err)
				return err
			}

			fileData = data
			fileName = file.Filename
			contentType = cType
		}
	}

	// 生成剪贴板ID（事务外进行）
	fullID, clipID := id.GenerateClipID(req.SpaceID)

	// 准备剪贴板内容
	cl := clip.Clip{
		ID:          fullID,
		ClipID:      clipID,
		SpaceID:     req.SpaceID,
		ContentType: contentType,
		Content:     req.Content,
		Creator: &clip.Creator{
			ID:       userID,
			Username: username,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	var uploadedClip *clip.Clip

	// 执行数据库事务
	err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 处理文件上传
		if fileData != nil {
			fileName := fmt.Sprintf("%s%s", cl.ClipID, filepath.Ext(fileName))
			filePath, err := storage.SaveFile(fileData, fileName)
			if err != nil {
				logger.Error("保存文件失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, ErrFileUploadFailed)
			}
			cl.FilePath = filePath
		}

		// 插入数据库
		_, err = tx.Exec(`
			INSERT INTO nlip_clipboard_items 
			(id, clip_id, space_id, content_type, content, file_path, creator_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, cl.ID, cl.ClipID, cl.SpaceID, cl.ContentType, cl.Content, cl.FilePath, userID, cl.CreatedAt, cl.UpdatedAt)

		if err != nil {
			if cl.FilePath != "" {
				if err := storage.DeleteFile(cl.FilePath); err != nil {
					logger.Error("删除失败的上传文件失败: %v", err)
				}
			}
			logger.Error("保存剪贴板内容失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "保存剪贴板内容失败")
		}

		uploadedClip = &cl
		return nil
	})

	if err != nil {
		return err
	}

	// 创建一个 channel 用于等待清理操作完成
	cleanupDone := make(chan error)

	// 执行清理操作
	go func() {
		// 添加延迟，确保上传事务完全提交
		time.Sleep(100 * time.Millisecond)

		err := cleaner.CleanSpaceOverflow(req.SpaceID)
		if err != nil {
			logger.Error("清理空间超量内容失败: %v", err)
			cleanupDone <- err
			return
		}
		cleanupDone <- nil
	}()

	// 等待清理操作完成
	if cleanupErr := <-cleanupDone; cleanupErr != nil {
		// 如果清理失败，记录日志但仍然返回上传成功
		logger.Error("清理空间超量内容失败，但上传已成功: %v", cleanupErr)
	}

	logger.Info("用户 %s 成功上传内容到空间 %s，清理操作已完成", userID, req.SpaceID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "上传成功",
		"data": clip.ClipResponse{
			Clip: uploadedClip,
		},
	})
}

// HandleListClips 获取剪贴板内容列表
// @Summary 获取Clip列表
// @Description 获取当前用户的Clip列表
// @Tags 剪贴板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} clip.ListClipsResponse "获取成功"
// @Failure 401 {object} string "未授权"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/clips [get]
func HandleListClips(c *fiber.Ctx) error {
	s := c.Locals("space").(space.Space)

	return db.WithTransaction(config.DB, func(tx *sql.Tx) error {

		// 使用通用查询语句
		rows, err := tx.Query(
			selectClipWithCreatorSQL+
				"WHERE c.space_id = ? ORDER BY c.created_at DESC",
			s.ID,
		)

		if err != nil {
			logger.Error("获取剪贴板内容失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "获取剪贴板内容失败")
		}
		defer rows.Close()

		var clips []clip.Clip
		for rows.Next() {
			cl, err := scanClip(rows)
			if err != nil {
				logger.Error("读取剪贴板数据失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "读取剪贴板数据失败")
			}
			clips = append(clips, *cl)
		}

		return c.JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "获取成功",
			"data": clip.ListClipsResponse{
				Clips: clips,
			},
		})
	})
}


// HandleGetLastClip 获取最近修改的剪贴板
// @Summary 获取最近修改的Clip
// @Description 获取最近修改的Clip
// @Tags 剪贴板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} clip.ClipResponse "获取成功"
// @Failure 401 {object} string "未授权"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/clips/last [get]
func HandleGetLastClip(c *fiber.Ctx) error {
	s := c.Locals("space").(space.Space)
	isDownload := c.Query("download") == "true"
	return db.WithTransaction(config.DB, func(tx *sql.Tx) error {

		// 使用通用查询语句
		row := tx.QueryRow(
			selectClipWithCreatorSQL+
				"WHERE c.space_id = ? ORDER BY c.updated_at DESC LIMIT 1",
			s.ID,
		)

		cl, err := scanSingleClip(row)
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, ErrClipNotFound)
		} else if err != nil {
			logger.Error("获取剪贴板内容失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "获取剪贴板内容失败")
		}

		if cl.FilePath != "" && isDownload {
			data, err := storage.GetFile(cl.FilePath)
			if err != nil {
				logger.Error("读取文件失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "读取文件失败")
			}

			c.Set("Content-Type", cl.ContentType)
			c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`,
				filepath.Base(cl.FilePath)))

			return c.Send(data)
		}

		return c.JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "获取成功",
			"data": clip.ClipResponse{
				Clip: cl,
			},
		})
	})
}

// HandleGetClip 获取单个剪贴板内容
// @Summary 获取Clip详情
// @Description 根据ID获取单个Clip的详细信息
// @Tags 剪贴板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Clip ID"
// @Success 200 {object} clip.ClipResponse "获取成功"
// @Failure 401 {object} string "未授权"
// @Failure 404 {object} string "Clip不存在"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/clips/{id} [get]
func HandleGetClip(c *fiber.Ctx) error {
	s := c.Locals("space").(space.Space)
	clipID := c.Params("clipId")
	isDownload := c.Query("download") == "true"

	return db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 查询剪贴板内容
		row := tx.QueryRow(
			selectClipWithCreatorSQL+
				"WHERE c.clip_id = ? AND c.space_id = ?",
			clipID, s.ID,
		)

		cl, err := scanSingleClip(row)
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, ErrClipNotFound)
		} else if err != nil {
			logger.Error("获取剪贴板内容失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "获取剪贴板内容失败")
		}

		// 处理文件下载请求
		if cl.FilePath != "" && isDownload {
			data, err := storage.GetFile(cl.FilePath)
			if err != nil {
				logger.Error("读取文件失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "读取文件失败")
			}

			c.Set("Content-Type", cl.ContentType)
			c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`,
				filepath.Base(cl.FilePath)))

			return c.Send(data)
		}

		return c.JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "获取成功",
			"data": clip.ClipResponse{
				Clip: cl,
			},
		})
	})
}

// HandleDeleteClip 删除剪贴板内容
// @Summary 删除Clip
// @Description 删除指定的Clip
// @Tags 剪贴板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Clip ID"
// @Success 200 {object} string "删除成功"
// @Failure 401 {object} string "未授权"
// @Failure 403 {object} string "无权限删除"
// @Failure 404 {object} string "Clip不存在"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/clips/{id} [delete]
func HandleDeleteClip(c *fiber.Ctx) error {
	s := c.Locals("space").(space.Space)
	clipID := c.Params("clipId")
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	logger.Debug("处理删除剪贴板内容请求: spaceID=%s, clipID=%s", s.ID, clipID)

	err := db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 权限检查
		_, err := checkClipPermission(tx, s.ID, s.Type, clipID, userID, isAdmin)
		if err != nil {
			return err
		}

		// 获取文件路径
		var filePath string
		err = db.QueryRowTx(tx, "SELECT file_path FROM nlip_clipboard_items WHERE clip_id = ? AND space_id = ?",
			clipID, s.ID).Scan(&filePath)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "查询剪贴板内容失败")
		}

		// 删除数据库记录
		_, err = tx.Exec("DELETE FROM nlip_clipboard_items WHERE clip_id = ? AND space_id = ?", clipID, s.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "删除剪贴板内容失败")
		}

		// 删除关联文件
		if filePath != "" {
			if err := storage.DeleteFile(filePath); err != nil {
				logger.Error("删除文件失败: %v", err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	logger.Info("用户 %s 删除了剪贴板内容: spaceID=%s, clipID=%s", userID, s.ID, clipID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "删除成功",
		"data":    nil,
	})
}

// HandleUpdateClip 处理更新剪贴板内容
// @Summary 更新Clip
// @Description 更新指定Clip的信息
// @Tags 剪贴板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Clip ID"
// @Param request body clip.UpdateClipRequest true "更新Clip请求参数"
// @Success 200 {object} clip.ClipResponse "更新成功"
// @Failure 400 {object} string "请求参数错误"
// @Failure 401 {object} string "未授权"
// @Failure 403 {object} string "无权限修改"
// @Failure 404 {object} string "Clip不存在"
// @Failure 500 {object} string "服务器内部错误"
// @Router /api/v1/nlip/clips/{id} [put]
func HandleUpdateClip(c *fiber.Ctx) error {
	s := c.Locals("space").(space.Space)
	clipID := c.Params("clipId")
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	var req clip.UpdateClipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrInvalidRequest)
	}

	return db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 权限检查
		_, err := checkClipPermission(tx, s.ID, s.Type, clipID, userID, isAdmin)
		if err != nil {
			return err
		}

		// 更新内容
		result, err := tx.Exec(`
			UPDATE nlip_clipboard_items 
			SET content = ?, updated_at = ?
			WHERE clip_id = ? AND space_id = ?
		`, req.Content, time.Now(), clipID, s.ID)

		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "更新剪贴板内容失败")
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fiber.NewError(fiber.StatusNotFound, ErrClipNotFound)
		}

		// 查询更新后的完整剪贴板内容
		row := tx.QueryRow(
			selectClipWithCreatorSQL+
				"WHERE c.clip_id = ? AND c.space_id = ?",
			clipID, s.ID,
		)

		cl, err := scanSingleClip(row)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "获取更新后的剪贴板内容失败")
		}

		logger.Info("用户 %s 更新了剪贴板内容: spaceID=%s, clipID=%s", userID, s.ID, clipID)
		return c.JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "更新成功",
			"data": clip.ClipResponse{
				Clip: cl,
			},
		})
	})
}
