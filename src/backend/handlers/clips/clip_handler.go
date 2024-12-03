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
	"mime/multipart"
)

// 添加常量定义
const (
	// 空间类型
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
func checkClipPermission(tx *sql.Tx, spaceID, clipID, userID string, isAdmin bool) (string, error) {
	var creatorID string

	if spaceID == PublicSpaceID {
		err := tx.QueryRow(`
			SELECT creator_id 
			FROM nlip_clipboard_items 
			WHERE clip_id = ? AND space_id = ?
		`, clipID, spaceID).Scan(&creatorID)

		if err == sql.ErrNoRows {
			return "", fiber.NewError(fiber.StatusNotFound, ErrClipNotFound)
		} else if err != nil {
			return "", err
		}

		// 公共空间权限检查
		if !isAdmin && (userID != creatorID || creatorID == GuestUserID) {
			return "", fiber.NewError(fiber.StatusForbidden, ErrNoPermission)
		}
	} else {
		// 其他空间权限检查
		var spaceOwnerID string
		err := tx.QueryRow("SELECT owner_id FROM nlip_spaces WHERE id = ?", spaceID).Scan(&spaceOwnerID)
		if err == sql.ErrNoRows {
			return "", fiber.NewError(fiber.StatusNotFound, ErrSpaceNotFound)
		} else if err != nil {
			return "", err
		}

		if !isAdmin && spaceOwnerID != userID {
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
	var creatorID, creatorUsername sql.NullString
	var filePath sql.NullString

	err := rows.Scan(
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
		// 验证creator字段必须为guest
		creator := form.Value["creator"]
		if len(creator) == 0 || creator[0] != GuestUserID {
			return fiber.NewError(fiber.StatusBadRequest, "游客上传必须设置creator为guest")
		}

		if files := form.File["file"]; len(files) > 0 {
			file := files[0]
			logger.Debug("处理公共空间文件上传: %s, 大小: %d bytes", file.Filename, file.Size)

			data, contentType, err := handleFileUpload(file)
			if err != nil {
				logger.Error("处理文件上传失败: %v", err)
				return err
			}

			// 生成剪贴板ID
			fullID, clipID := id.GenerateClipID(PublicSpaceID)

			// 创建剪贴板内容
			cl := clip.Clip{
				ID:          fullID,
				ClipID:      clipID,
				SpaceID:     PublicSpaceID,
				ContentType: contentType,
				Creator: &clip.Creator{
					ID:       GuestUserID,
					Username: "游客",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// 保存文件
			fileName := fmt.Sprintf("%s%s", cl.ClipID, filepath.Ext(file.Filename))
			filePath, err := storage.SaveFile(data, fileName)
			if err != nil {
				logger.Error("保存文件失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, ErrFileUploadFailed)
			}
			cl.FilePath = filePath

			// 插入数据库
			_, err = config.DB.Exec(`
				INSERT INTO nlip_clipboard_items 
				(id, clip_id, space_id, content_type, file_path, creator_id, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, cl.ID, cl.ClipID, cl.SpaceID, cl.ContentType, cl.FilePath, GuestUserID, cl.CreatedAt, cl.UpdatedAt)

			if err != nil {
				if cl.FilePath != "" {
					if err := storage.DeleteFile(cl.FilePath); err != nil {
						logger.Error("删除失败的上传文件失败: %v", err)
					}
				}
				logger.Error("保存剪贴板内容失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "保存剪贴板内容失败")
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"code":    fiber.StatusCreated,
				"message": "上传成功",
				"data": clip.ClipResponse{
					Clip: &cl,
				},
			})
		}
	}

	// 处理文本内容上传
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrInvalidRequest)
	}

	// 从表单数据中获取 creator
	creator := c.FormValue("creator", "")
	if creator == "" {
		return fiber.NewError(fiber.StatusBadRequest, "游客上传必须设置creator为guest")
	}
	if creator != GuestUserID {
		return fiber.NewError(fiber.StatusBadRequest, "creator必须为guest")
	}

	// 强制设置为公共空间
	req.SpaceID = PublicSpaceID

	// 生成剪贴板ID
	fullID, clipID := id.GenerateClipID(req.SpaceID)

	// 创建剪贴板内容
	cl := clip.Clip{
		ID:          fullID,
		ClipID:      clipID,
		SpaceID:     req.SpaceID,
		ContentType: req.ContentType,
		Content:     req.Content,
		Creator: &clip.Creator{
			ID:       GuestUserID,
			Username: "游客",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 插入数据库
	_, err := config.DB.Exec(`
		INSERT INTO nlip_clipboard_items 
		(id, clip_id, space_id, content_type, content, creator_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, cl.ID, cl.ClipID, cl.SpaceID, cl.ContentType, cl.Content, GuestUserID, cl.CreatedAt, cl.UpdatedAt)

	if err != nil {
		logger.Error("保存剪贴板内容失败: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "保存剪贴板内容失败")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"code":    fiber.StatusCreated,
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

	// 处理文件上传
	if form, err := c.MultipartForm(); err == nil && form.File != nil {
		if files := form.File["file"]; len(files) > 0 {
			file := files[0]
			logger.Debug("处理文件上传: %s, 大小: %d bytes", file.Filename, file.Size)

			data, contentType, err := handleFileUpload(file)
			if err != nil {
				logger.Error("处理文件上传失败: %v", err)
				return err
			}

			req.File = data
			req.FileName = file.Filename
			req.ContentType = contentType
		}
	}

	// 解析其他表单数据
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrInvalidRequest)
	}

	return db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 检查空间权限
		if req.SpaceID != PublicSpaceID {
			_, err := checkClipPermission(tx, req.SpaceID, "", userID, c.Locals("isAdmin").(bool))
			if err != nil {
				return err
			}
		}

		// 生成剪贴板ID
		fullID, clipID := id.GenerateClipID(req.SpaceID)

		// 创建剪贴板内容
		cl := clip.Clip{
			ID:          fullID,
			ClipID:      clipID,
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

		// 处理文件上传
		if req.File != nil {
			fileName := fmt.Sprintf("%s%s", cl.ClipID, filepath.Ext(req.FileName))
			filePath, err := storage.SaveFile(req.File, fileName)
			if err != nil {
				logger.Error("保存文件失败: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, ErrFileUploadFailed)
			}
			cl.FilePath = filePath
		}

		// 插入数据库
		_, err := tx.Exec(`
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

		logger.Info("用户 %s 成功上传内容到空间 %s", userID, req.SpaceID)
		return c.JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "上传成功",
			"data": clip.ClipResponse{
				Clip: &cl,
			},
		})
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

	return db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 权限检查
		if spaceID != PublicSpaceID {
			_, err := checkClipPermission(tx, spaceID, "", userID, isAdmin)
			if err != nil {
				return err
			}
		}

		// 使用通用查询语句
		rows, err := tx.Query(
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
	})
}

// HandleGetClip 获取单个剪贴板内容
func HandleGetClip(c *fiber.Ctx) error {
	spaceID := c.Params("spaceId")
	clipID := c.Params("id")
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)
	isDownload := c.Query("download") == "true"

	return db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 权限检查
		if spaceID != PublicSpaceID {
			_, err := checkClipPermission(tx, spaceID, clipID, userID, isAdmin)
			if err != nil {
				return err
			}
		}

		// 查询剪贴板内容
		row := tx.QueryRow(
			selectClipWithCreatorSQL+
				"WHERE c.clip_id = ? AND c.space_id = ?",
			clipID, spaceID,
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
func HandleDeleteClip(c *fiber.Ctx) error {
	spaceID := c.Params("spaceId")
	clipID := c.Params("id")
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	logger.Debug("处理删除剪贴板内容请求: spaceID=%s, clipID=%s", spaceID, clipID)

	return db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 权限检查
		_, err := checkClipPermission(tx, spaceID, clipID, userID, isAdmin)
		if err != nil {
			return err
		}

		// 获取文件路径
		var filePath string
		err = tx.QueryRow("SELECT file_path FROM nlip_clipboard_items WHERE clip_id = ? AND space_id = ?",
			clipID, spaceID).Scan(&filePath)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "查询剪贴板内容失败")
		}

		// 删除数据库记录
		_, err = tx.Exec("DELETE FROM nlip_clipboard_items WHERE clip_id = ? AND space_id = ?", clipID, spaceID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "删除剪贴板内容失败")
		}

		// 删除关联文件
		if filePath != "" {
			if err := storage.DeleteFile(filePath); err != nil {
				logger.Error("删除文件失败: %v", err)
			}
		}

		logger.Info("用户 %s 删除了剪贴板内容: spaceID=%s, clipID=%s", userID, spaceID, clipID)
		return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
			"code": fiber.StatusNoContent,
		})
	})
}

// HandleUpdateClip 处理更新剪贴板内容
func HandleUpdateClip(c *fiber.Ctx) error {
	spaceID := c.Params("spaceId")
	clipID := c.Params("id")
	userID := c.Locals("userId").(string)
	isAdmin := c.Locals("isAdmin").(bool)

	var req clip.UpdateClipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrInvalidRequest)
	}

	return db.WithTransaction(config.DB, func(tx *sql.Tx) error {
		// 权限检查
		_, err := checkClipPermission(tx, spaceID, clipID, userID, isAdmin)
		if err != nil {
			return err
		}

		// 更新内容
		result, err := tx.Exec(`
			UPDATE nlip_clipboard_items 
			SET content = ?, updated_at = ?
			WHERE clip_id = ? AND space_id = ?
		`, req.Content, time.Now(), clipID, spaceID)

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
			clipID, spaceID,
		)

		cl, err := scanSingleClip(row)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "获取更新后的剪贴板内容失败")
		}

		logger.Info("用户 %s 更新了剪贴板内容: spaceID=%s, clipID=%s", userID, spaceID, clipID)
		return c.JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "更新成功",
			"data": clip.ClipResponse{
				Clip: cl,
			},
		})
	})
}
