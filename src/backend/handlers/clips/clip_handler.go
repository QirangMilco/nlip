package clips

import (
	"database/sql"
	"fmt"
	"nlip/config"
	"nlip/models/clip"
	"nlip/utils/id"
	"nlip/utils/logger"
	"nlip/utils/storage"
	"nlip/utils/validator"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

// HandleUploadClip 处理上传剪贴板内容
func HandleUploadClip(c *fiber.Ctx) error {
	var req clip.UploadClipRequest
	userID := c.Locals("userId").(string)

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

	// 检查权限
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin && spaceOwnerID != userID {
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
        INSERT INTO nlip_clipboard_items (id, clip_id, space_id, content_type, content, file_path) 
        VALUES (?, ?, ?, ?, ?, ?)
    `, cl.ID, cl.ClipID, cl.SpaceID, cl.ContentType, cl.Content, cl.FilePath)

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
	return c.Status(fiber.StatusCreated).JSON(clip.ClipResponse{
		Clip: &cl,
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

	// 查询剪贴板内容
	rows, err := config.DB.Query(`
        SELECT id, clip_id, space_id, content_type, content, file_path, created_at 
        FROM nlip_clipboard_items 
        WHERE space_id = ? 
        ORDER BY created_at DESC
    `, spaceID)

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "获取剪贴板内容失败")
	}
	defer rows.Close()

	var clips []clip.Clip
	for rows.Next() {
		var cl clip.Clip
		err := rows.Scan(&cl.ID, &cl.ClipID, &cl.SpaceID, &cl.ContentType, &cl.Content, &cl.FilePath, &cl.CreatedAt)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "读取剪贴板数据失败")
		}
		clips = append(clips, cl)
	}

	return c.JSON(clip.ListClipsResponse{
		Clips: clips,
	})
}

// HandleGetClip 获取单个剪贴板内容
func HandleGetClip(c *fiber.Ctx) error {
	spaceID := c.Params("spaceId")
	clipID := c.Params("id")

	// 验证剪贴板是否属于指定空间
	var cl clip.Clip
	var spaceType, spaceOwnerID string
	err := config.DB.QueryRow(`
        SELECT c.id, c.clip_id, c.space_id, c.content_type, c.content, c.file_path, c.created_at,
               s.type as space_type, s.owner_id as space_owner_id
        FROM nlip_clipboard_items c
        JOIN nlip_spaces s ON c.space_id = s.id
        WHERE c.clip_id = ? AND c.space_id = ?
    `, clipID, spaceID).Scan(
		&cl.ID, &cl.ClipID, &cl.SpaceID, &cl.ContentType, &cl.Content, &cl.FilePath, &cl.CreatedAt,
		&spaceType, &spaceOwnerID,
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

	return c.JSON(clip.ClipResponse{
		Clip: &cl,
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
