package clips

import (
	"database/sql"
	"fmt"
	"nlip/config"
	"nlip/models/clip"
	"nlip/utils/db"
	"nlip/utils/id"
	"nlip/utils/logger"
	"nlip/utils/storage"
	"nlip/utils/validator"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// 添加通用的查询语句常量
const (
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
            u.username as creator_username
        FROM nlip_clipboard_items c
        LEFT JOIN nlip_users u ON c.creator_id = u.id
    `

	selectClipWithSpaceAndCreatorSQL = `
        SELECT 
            c.id, 
            c.clip_id, 
            c.space_id, 
            c.content_type, 
            c.content, 
            c.file_path, 
            c.created_at,
            c.updated_at,
            s.owner_id as space_owner_id,
            u.id as creator_id,
            u.username as creator_username
        FROM nlip_clipboard_items c
        JOIN nlip_spaces s ON c.space_id = s.id
        LEFT JOIN nlip_users u ON c.creator_id = u.id
    `
)

// scanClip 辅助函数：扫描剪贴板数据
func scanClip(rows *sql.Rows) (*clip.Clip, error) {
	var cl clip.Clip
	var creatorID, creatorUsername sql.NullString

	err := rows.Scan(
		&cl.ID,
		&cl.ClipID,
		&cl.SpaceID,
		&cl.ContentType,
		&cl.Content,
		&cl.FilePath,
		&cl.CreatedAt,
		&cl.UpdatedAt,
		&creatorID,
		&creatorUsername,
	)

	if err != nil {
		return nil, err
	}

	if creatorID.Valid && creatorUsername.Valid {
		cl.Creator = &clip.Creator{
			ID:       creatorID.String,
			Username: creatorUsername.String,
		}
	}

	return &cl, nil
}

// scanClipWithSpaceOwner 辅助函数：扫描带空间所有者信息的剪贴板数据
func scanClipWithSpaceOwner(row *sql.Row) (*clip.Clip, string, error) {
	var cl clip.Clip
	var spaceOwnerID string
	var creatorID, creatorUsername sql.NullString

	err := row.Scan(
		&cl.ID,
		&cl.ClipID,
		&cl.SpaceID,
		&cl.ContentType,
		&cl.Content,
		&cl.FilePath,
		&cl.CreatedAt,
		&cl.UpdatedAt,
		&spaceOwnerID,
		&creatorID,
		&creatorUsername,
	)

	if err != nil {
		return nil, "", err
	}

	if creatorID.Valid && creatorUsername.Valid {
		cl.Creator = &clip.Creator{
			ID:       creatorID.String,
			Username: creatorUsername.String,
		}
	}

	return &cl, spaceOwnerID, nil
}

// HandleListPublicClips 处理获取公共空间剪贴板列表
func HandleListPublicClips(c *fiber.Ctx) error {
	logger.Debug("获取公共空间剪贴板列表")

	// 使用通用查询语句
	rows, err := config.DB.Query(
		selectClipWithCreatorSQL+
			"WHERE c.space_id = ? ORDER BY c.created_at DESC",
		"public-space",
	)

	if err != nil {
		logger.Error("获取公共空间剪贴板内容失败: %v", err)
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
}

// HandleUploadPublicClip 处理游客上传公共空间剪贴板内容
func HandleUploadPublicClip(c *fiber.Ctx) error {
	var req clip.UploadClipRequest

	// 处理multipart表单数据
	if form, err := c.MultipartForm(); err == nil {
		if files := form.File["file"]; len(files) > 0 {
			file := files[0]
			logger.Debug("处理公共空间文件上传: %s, 大小: %d bytes", file.Filename, file.Size)

			// 检查文件大小
			if file.Size > config.AppConfig.MaxFileSize {
				logger.Warning("文件大小超过限制: %s (%d bytes)", file.Filename, file.Size)
				return fiber.NewError(fiber.StatusBadRequest, "文件大小超过限制")
			}

			// 验证文件名和类型
			if !validator.ValidateFileName(file.Filename) {
				logger.Warning("文件名不合法: %s", file.Filename)
				return fiber.NewError(fiber.StatusBadRequest, "文件名不合法")
			}

			contentType := file.Header.Get("Content-Type")
			if !validator.ValidateFileType(file.Filename, contentType) {
				logger.Warning("不支持的文件类型: %s (%s)", file.Filename, contentType)
				return fiber.NewError(fiber.StatusBadRequest, "不支持的文件类型")
			}

			// 读取文件数据
			fileContent, err := file.Open()
			if err != nil {
				logger.Error("读取文件失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "读取文件失败")
			}
			defer fileContent.Close()

			data := make([]byte, file.Size)
			if _, err := fileContent.Read(data); err != nil {
				logger.Error("读取文件数据失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "读取文件失败")
			}

			req.File = data
			req.FileName = file.Filename
			req.ContentType = contentType
		}
	}

	// 解析其他表单数据
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析请求数据失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	// 强制设置为公共空间，并确保是游客上传
	req.SpaceID = "public-space"

	// 生成剪贴板ID
	fullID, clipID := id.GenerateClipID(req.SpaceID)

	// 创建剪贴板内容，明确指定为游客
	cl := clip.Clip{
		ID:          fullID,
		ClipID:      clipID,
		SpaceID:     req.SpaceID,
		ContentType: req.ContentType,
		Content:     req.Content,
		Creator: &clip.Creator{
			ID:       "guest",
			Username: "游客",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 如果有文件，保存到存储
	if req.File != nil {
		fileName := fmt.Sprintf("%s%s", cl.ClipID, filepath.Ext(req.FileName))
		filePath, err := storage.SaveFile(req.File, fileName)
		if err != nil {
			logger.Error("保存文件失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "保存文件失败")
		}
		cl.FilePath = filePath
	}

	// 插入数据库
	_, err := config.DB.Exec(`
        INSERT INTO nlip_clipboard_items 
        (id, clip_id, space_id, content_type, content, file_path, creator_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, cl.ID, cl.ClipID, cl.SpaceID, cl.ContentType, cl.Content, cl.FilePath, "guest", cl.CreatedAt, cl.UpdatedAt)

	if err != nil {
		if cl.FilePath != "" {
			if err := storage.DeleteFile(cl.FilePath); err != nil {
				logger.Error("删除失败的上传文件失败: %v", err)
			}
		}
		logger.Error("保存剪贴板内容失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "保存剪贴板内容失败")
	}

	logger.Info("游客成功上传内容到公共空间")
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "上传成功",
		"data": clip.ClipResponse{
			Clip: &cl,
		},
	})
}

// HandleGetPublicClip 获取公共空间的单个剪贴板内容
func HandleGetPublicClip(c *fiber.Ctx) error {
	clipID := c.Params("id")

	var cl clip.Clip
	var creatorID, creatorUsername sql.NullString

	err := config.DB.QueryRow(`
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
        WHERE c.clip_id = ? AND c.space_id = ?
    `, clipID, "public-space").Scan(
		&cl.ID,
		&cl.ClipID,
		&cl.SpaceID,
		&cl.ContentType,
		&cl.Content,
		&cl.FilePath,
		&cl.CreatedAt,
		&cl.UpdatedAt,
		&creatorID,
		&creatorUsername,
	)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "剪贴板内容不存在")
	} else if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "获取剪贴板内容失败")
	}

	// 设置创建者信息
	if creatorID.Valid && creatorUsername.Valid {
		cl.Creator = &clip.Creator{
			ID:       creatorID.String,
			Username: creatorUsername.String,
		}
	}

	// 如果是文件类型且请求下载
	if cl.FilePath != "" && c.Query("download") == "true" {
		data, err := storage.GetFile(cl.FilePath)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "读取文件失败")
		}

		c.Set("Content-Type", cl.ContentType)
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(cl.FilePath)))
		return c.Send(data)
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取成功",
		"data": clip.ClipResponse{
			Clip: &cl,
		},
	})
}

// HandleUploadClip 处理上传剪贴板内容
func HandleUploadClip(c *fiber.Ctx) error {
	var req clip.UploadClipRequest
	userID := c.Locals("userId").(string)
	username := c.Locals("username").(string)

	// 处理multipart表单数据
	if form, err := c.MultipartForm(); err == nil {
		if files := form.File["file"]; len(files) > 0 {
			file := files[0]
			logger.Debug("处理文件上传: %s, 大小: %d bytes", file.Filename, file.Size)

			// 检查文件大小
			if file.Size > config.AppConfig.MaxFileSize {
				logger.Warning("文件大小超过限制: %s (%d bytes)", file.Filename, file.Size)
				return fiber.NewError(fiber.StatusBadRequest, "文件大小超过限制")
			}

			// 验证文件名
			if !validator.ValidateFileName(file.Filename) {
				logger.Warning("文件名不合法: %s", file.Filename)
				return fiber.NewError(fiber.StatusBadRequest, "文件名不合法")
			}

			// 验证文件类型
			contentType := file.Header.Get("Content-Type")
			if !validator.ValidateFileType(file.Filename, contentType) {
				logger.Warning("不支持的文件类型: %s (%s)", file.Filename, contentType)
				return fiber.NewError(fiber.StatusBadRequest, "不支持的文件类型")
			}

			// 读取文件数据
			fileContent, err := file.Open()
			if err != nil {
				logger.Error("读取文件失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "读取文件失败")
			}
			defer fileContent.Close()

			data := make([]byte, file.Size)
			if _, err := fileContent.Read(data); err != nil {
				logger.Error("读取文件数据失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "读取文件失败")
			}

			req.File = data
			req.FileName = file.Filename
			req.ContentType = contentType
			logger.Debug("文件验证通过: %s", file.Filename)
		}
	}

	// 解析其他表单数据
	if err := c.BodyParser(&req); err != nil {
		logger.Warning("解析请求数据失败: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	// 检查空间是否存在
	var spaceOwnerID string
	err := config.DB.QueryRow("SELECT owner_id FROM nlip_spaces WHERE id = ?", req.SpaceID).Scan(&spaceOwnerID)
	if err == sql.ErrNoRows {
		logger.Warning("空间不存在: %s", req.SpaceID)
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		logger.Error("查询空间失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "查询空间失败")
	}

	// 修改权限检查逻辑：允许上传到公共空间
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin && spaceOwnerID != userID && req.SpaceID != "public-space" {
		logger.Warning("用户 %s 尝试上传到无权限的空间 %s", userID, req.SpaceID)
		return fiber.NewError(fiber.StatusForbidden, "没有权限上传到此空间")
	}

	// 生成剪贴板ID
	fullID, clipID := id.GenerateClipID(req.SpaceID)

	// 创建剪贴板内容
	cl := clip.Clip{
		ID:          fullID, // 使用完整ID作为主键
		ClipID:      clipID, // 新增的空间内唯一ID
		SpaceID:     req.SpaceID,
		ContentType: req.ContentType,
		Content:     req.Content,
		Creator: &clip.Creator{
			ID:       userID,
			Username: username,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 如果有文件，保存到存储
	if req.File != nil {
			fileName := fmt.Sprintf("%s%s", cl.ClipID, filepath.Ext(req.FileName)) // 使用clipID而不是完整ID
			filePath, err := storage.SaveFile(req.File, fileName)
			if err != nil {
				logger.Error("保存文件失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "保存文件失败")
			}
			cl.FilePath = filePath
			logger.Debug("文件已保存: %s", filePath)
	}

	// 插入数据库
	_, err = config.DB.Exec(`
        INSERT INTO nlip_clipboard_items 
        (id, clip_id, space_id, content_type, content, file_path, creator_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, cl.ID, cl.ClipID, cl.SpaceID, cl.ContentType, cl.Content, cl.FilePath, userID, cl.CreatedAt, cl.UpdatedAt)

	if err != nil {
		// 如果数据库插入失败，删除已上传的文件
		if cl.FilePath != "" {
			if err := storage.DeleteFile(cl.FilePath); err != nil {
				logger.Error("删除失败的上传文件失败: %v", err)
			}
		}
		logger.Error("保存剪贴板内容失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "保存剪贴板内容失败")
	}

	logger.Info("用户 %s 成功上传内容到空间 %s", userID, req.SpaceID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "上传成功",
		"data": clip.ClipResponse{
			Clip: &cl,
		},
	})
}

// HandleListClips 获取剪贴板内容列表
func HandleListClips(c *fiber.Ctx) error {
	spaceID := c.Params("spaceId")
	if spaceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "必须指定空间ID")
	}

	// 检查空间访问权限
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	var spaceType, spaceOwnerID string
	err := config.DB.QueryRow("SELECT type, owner_id FROM nlip_spaces WHERE id = ?", spaceID).Scan(&spaceType, &spaceOwnerID)
	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "空间不存在")
	} else if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "查询空间失败")
	}

	if spaceType != "public" && !isAdmin && spaceOwnerID != userID {
		return fiber.NewError(fiber.StatusForbidden, "没有权限访问此空间")
	}

	// 使用通用查询语句
	rows, err := config.DB.Query(
		selectClipWithCreatorSQL+
			"WHERE c.space_id = ? ORDER BY c.created_at DESC",
		spaceID,
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
}

// HandleGetClip 获取单个剪贴板内容
func HandleGetClip(c *fiber.Ctx) error {
	spaceID := c.Params("spaceId")
	clipID := c.Params("id")

	// 验证剪贴板是否属于指定空间
	var cl clip.Clip
	var spaceType, spaceOwnerID string
	var creatorID, creatorUsername sql.NullString

	err := config.DB.QueryRow(`
        SELECT 
            c.id, 
            c.clip_id, 
            c.space_id, 
            c.content_type, 
            c.content, 
            c.file_path, 
            c.created_at,
            c.updated_at,
            s.type as space_type, 
            s.owner_id as space_owner_id,
            u.id as creator_id,
            u.username as creator_username
        FROM nlip_clipboard_items c
        JOIN nlip_spaces s ON c.space_id = s.id
        LEFT JOIN nlip_users u ON c.creator_id = u.id
        WHERE c.clip_id = ? AND c.space_id = ?
    `, clipID, spaceID).Scan(
		&cl.ID,
		&cl.ClipID,
		&cl.SpaceID,
		&cl.ContentType,
		&cl.Content,
		&cl.FilePath,
		&cl.CreatedAt,
		&cl.UpdatedAt,
		&spaceType,
		&spaceOwnerID,
		&creatorID,
		&creatorUsername,
	)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "剪贴板内容不存在")
	} else if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "获取剪贴板内容失败")
	}

	// 检查访问权限
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	if spaceType != "public" && !isAdmin && spaceOwnerID != userID {
		return fiber.NewError(fiber.StatusForbidden, "没有权限访问此内容")
	}

	// 如果有创建者信息，添加到响应中
	if creatorID.Valid && creatorUsername.Valid {
		cl.Creator = &clip.Creator{
			ID:       creatorID.String,
			Username: creatorUsername.String,
		}
	}

	// 如果是文件类型，需要读取文件内容
	if cl.FilePath != "" && c.Query("download") == "true" {
		data, err := storage.GetFile(cl.FilePath)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "读取文件失败")
		}

		// 设置文件下载响应头
		c.Set("Content-Type", cl.ContentType)
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(cl.FilePath)))
		return c.Send(data)
	}

	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "获取成功",
		"data": clip.ClipResponse{
			Clip: &cl,
		},
	})
}

// HandleDeleteClip 删除剪贴板内容
func HandleDeleteClip(c *fiber.Ctx) error {
	spaceID := c.Params("spaceId")
	clipID := c.Params("id")

	// 验证剪贴板是否属于指定空间
	var filePath string
	var spaceOwnerID string
	err := config.DB.QueryRow(`
        SELECT c.file_path, s.owner_id
        FROM nlip_clipboard_items c
        JOIN nlip_spaces s ON c.space_id = s.id
        WHERE c.clip_id = ? AND c.space_id = ?
    `, clipID, spaceID).Scan(&filePath, &spaceOwnerID)

	if err == sql.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "剪贴板内容不存在")
	} else if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "查询剪贴板内容失败")
	}

	// 检查删除权限
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	if !isAdmin && spaceOwnerID != userID {
		return fiber.NewError(fiber.StatusForbidden, "没有权限删除此内容")
	}

	// 开始事务
	tx, err := config.DB.Begin()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "开始事务失败")
	}

	// 删除数据库记录
	_, err = tx.Exec("DELETE FROM nlip_clipboard_items WHERE clip_id = ? AND space_id = ?", clipID, spaceID)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Error("事务回滚失败: %v", rollbackErr)
		}
		return fiber.NewError(fiber.StatusInternalServerError, "删除剪贴板内容失败")
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "提交事务失败")
	}

	// 删除文件（如果存在）
	if filePath != "" {
		if err := storage.DeleteFile(filePath); err != nil {
			// 这里我们只记录错误，不影响API响应
			fmt.Printf("删除文件失败: %v\n", err)
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleUpdateClip 处理更新剪贴板内容
func HandleUpdateClip(c *fiber.Ctx) error {
	spaceID := c.Params("spaceId")
	clipID := c.Params("id")
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	logger.Debug("处理更新剪贴板内容请求: spaceID=%s, clipID=%s", spaceID, clipID)

	// 添加请求解析
	var req clip.UpdateClipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "无效的请求数据")
	}

	// 使用事务进行更新
	err := db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 验证剪贴板是否存在并获取当前信息
		_, spaceOwnerID, err := scanClipWithSpaceOwner(tx.QueryRow(
			selectClipWithSpaceAndCreatorSQL+
				"WHERE c.clip_id = ? AND c.space_id = ?",
			clipID, spaceID,
		))

		if err == sql.ErrNoRows {
			logger.Warning("尝试更新不存在的剪贴板内容: spaceID=%s, clipID=%s", spaceID, clipID)
			return fiber.NewError(fiber.StatusNotFound, "剪贴板内容不存在")
		} else if err != nil {
			logger.Error("获取剪贴板内容失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "获取剪贴板内容失败")
		}

		// 检查权限
		if !isAdmin && spaceOwnerID != userID {
			logger.Warning("用户 %s 尝试更新无权限的剪贴板内容: spaceID=%s, clipID=%s", userID, spaceID, clipID)
			return fiber.NewError(fiber.StatusForbidden, "没有权限修改此内容")
		}

		// 更新内容
		_, err = db.ExecTx(tx, `
			UPDATE nlip_clipboard_items 
			SET content = ?
			WHERE clip_id = ? AND space_id = ?
		`, req.Content, clipID, spaceID)

		if err != nil {
			logger.Error("更新剪贴板内容失败: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "更新剪贴板内容失败")
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 重新获取更新后的内容
	var cl clip.Clip
	var creatorID, creatorUsername sql.NullString

	err = config.DB.QueryRow(
		selectClipWithCreatorSQL+
			"WHERE c.clip_id = ? AND c.space_id = ?",
		clipID, spaceID,
	).Scan(
		&cl.ID,
		&cl.ClipID,
		&cl.SpaceID,
		&cl.ContentType,
		&cl.Content,
		&cl.FilePath,
		&cl.CreatedAt,
		&cl.UpdatedAt,
		&creatorID,
		&creatorUsername,
	)

	if err != nil {
		logger.Error("获取更新后的剪贴板内容失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "获取更新后的内容失败")
	}

	// 添加创建者信息
	if creatorID.Valid && creatorUsername.Valid {
		cl.Creator = &clip.Creator{
			ID:       creatorID.String,
			Username: creatorUsername.String,
		}
	}

	logger.Info("用户 %s 更新了剪贴板内容: spaceID=%s, clipID=%s", userID, spaceID, clipID)
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "更新成功",
		"data": clip.ClipResponse{
			Clip: &cl,
		},
	})
}
