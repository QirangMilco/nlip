package cleaner

import (
    "database/sql"
    "fmt"
    "time"
    "nlip/config"
    "nlip/utils/storage"
    "nlip/utils/logger"
    "nlip/utils/db"
)

// StartCleanupTask 启动清理任务
func StartCleanupTask() {
    logger.Info("启动清理任务")
    go func() {
        // 启动时先执行一次清理
        if err := cleanExpiredItems(); err != nil {
            logger.Error("清理过期内容失败: %v", err)
        }
        if err := cleanOverflowItems(); err != nil {
            logger.Error("清理超量内容失败: %v", err)
        }

        // 设置定时器
        ticker := time.NewTicker(1 * time.Hour)
        defer ticker.Stop()

        logger.Info("清理任务定时器已设置，间隔: 1小时")
        for range ticker.C {
            logger.Debug("开始执行定时清理任务")
            if err := cleanExpiredItems(); err != nil {
                logger.Error("清理过期内容失败: %v", err)
            }
            if err := cleanOverflowItems(); err != nil {
                logger.Error("清理超量内容失败: %v", err)
            }
            logger.Debug("定时清理任务完成")
        }
    }()
}

// cleanExpiredItems 清理过期的内容
func cleanExpiredItems() error {
    logger.Debug("开始清理过期内容")
    // 查询所有空间的保留天数
    rows, err := db.QueryRows(config.DB, `
        SELECT id, retention_days 
        FROM nlip_spaces
    `)
    if err != nil {
        return fmt.Errorf("查询空间信息失败: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var spaceID string
        var retentionDays int
        if err := rows.Scan(&spaceID, &retentionDays); err != nil {
            return fmt.Errorf("读取空间数据失败: %w", err)
        }

        // 计算过期时间
        expireTime := time.Now().AddDate(0, 0, -retentionDays)
        logger.Debug("处理空间 %s, 保留天数: %d, 过期时间: %v", spaceID, retentionDays, expireTime)

        err := db.WithTransaction(config.DB, func(tx *sql.Tx) error {
            // 查询需要删除的文件路径
            fileRows, err := tx.Query(`
                SELECT file_path 
                FROM nlip_clipboard_items 
                WHERE space_id = ? AND created_at < ? AND file_path IS NOT NULL
            `, spaceID, expireTime)
            if err != nil {
                return fmt.Errorf("查询过期文件失败: %w", err)
            }
            defer fileRows.Close()

            var filePaths []string
            for fileRows.Next() {
                var filePath string
                if err := fileRows.Scan(&filePath); err != nil {
                    return fmt.Errorf("读取文件路径失败: %w", err)
                }
                filePaths = append(filePaths, filePath)
            }

            // 删除过期的记录
            result, err := db.ExecTx(tx, `
                DELETE FROM nlip_clipboard_items 
                WHERE space_id = ? AND created_at < ?
            `, spaceID, expireTime)
            if err != nil {
                return fmt.Errorf("删除过期记录失败: %w", err)
            }

            // 记录清理数量
            if count, err := result.RowsAffected(); err == nil {
                if count > 0 {
                    logger.Info("空间 %s 清理了 %d 条过期内容", spaceID, count)
                } else {
                    logger.Debug("空间 %s 没有过期内容需要清理", spaceID)
                }
            }

            // 删除文件
            for _, filePath := range filePaths {
                if err := storage.DeleteFile(filePath); err != nil {
                    logger.Error("删除文件失败 %s: %v", filePath, err)
                } else {
                    logger.Debug("成功删除文件: %s", filePath)
                }
            }

            return nil
        })

        if err != nil {
            return err
        }
    }

    logger.Debug("过期内容清理完成")
    return nil
}

// cleanOverflowItems 清理超出数量限制的内容
func cleanOverflowItems() error {
    logger.Debug("开始清理超量内容")
    // 查询所有空间的最大条目数
    rows, err := db.QueryRows(config.DB, `
        SELECT id, max_items 
        FROM nlip_spaces
    `)
    if err != nil {
        return fmt.Errorf("查询空间信息失败: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var spaceID string
        var maxItems int
        if err := rows.Scan(&spaceID, &maxItems); err != nil {
            return fmt.Errorf("读取空间数据失败: %w", err)
        }

        logger.Debug("处理空间 %s, 最大条目数: %d", spaceID, maxItems)

        err := db.WithTransaction(config.DB, func(tx *sql.Tx) error {
            // 查询需要删除的文件路径
            fileRows, err := tx.Query(`
                SELECT file_path 
                FROM nlip_clipboard_items 
                WHERE space_id = ? AND file_path IS NOT NULL
                ORDER BY created_at DESC
                LIMIT -1 OFFSET ?
            `, spaceID, maxItems)
            if err != nil {
                return fmt.Errorf("查询超量文件失败: %w", err)
            }
            defer fileRows.Close()

            var filePaths []string
            for fileRows.Next() {
                var filePath string
                if err := fileRows.Scan(&filePath); err != nil {
                    return fmt.Errorf("读取文件路径失败: %w", err)
                }
                filePaths = append(filePaths, filePath)
            }

            // 删除超出限制的记录
            result, err := db.ExecTx(tx, `
                DELETE FROM nlip_clipboard_items 
                WHERE id IN (
                    SELECT id 
                    FROM nlip_clipboard_items 
                    WHERE space_id = ? 
                    ORDER BY created_at DESC
                    LIMIT -1 OFFSET ?
                )
            `, spaceID, maxItems)
            if err != nil {
                return fmt.Errorf("删除超量记录失败: %w", err)
            }

            // 记录清理数量
            if count, err := result.RowsAffected(); err == nil {
                if count > 0 {
                    logger.Info("空间 %s 清理了 %d 条超量内容", spaceID, count)
                } else {
                    logger.Debug("空间 %s 没有超量内容需要清理", spaceID)
                }
            }

            // 删除文件
            for _, filePath := range filePaths {
                if err := storage.DeleteFile(filePath); err != nil {
                    logger.Error("删除文件失败 %s: %v", filePath, err)
                } else {
                    logger.Debug("成功删除文件: %s", filePath)
                }
            }

            return nil
        })

        if err != nil {
            return err
        }
    }

    logger.Debug("超量内容清理完成")
    return nil
}