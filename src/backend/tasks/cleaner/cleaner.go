package cleaner

import (
	"database/sql"
	"fmt"
	"nlip/config"
	"nlip/utils/db"
	"nlip/utils/logger"
	"nlip/utils/storage"
	"sync"
	"time"
)

var (
	isRunning    bool
	runningMutex sync.Mutex
)

func runWithLock(task func() error) error {
	runningMutex.Lock()
	if isRunning {
		runningMutex.Unlock()
		return fmt.Errorf("清理任务正在执行中")
	}
	isRunning = true
	runningMutex.Unlock()

	defer func() {
		runningMutex.Lock()
		isRunning = false
		runningMutex.Unlock()
	}()

	return task()
}

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
	const batchSize = 100

	// 先查询空间信息，不放在事务中
	spaces, err := getSpacesInfo()
	if err != nil {
		return err
	}

	for _, space := range spaces {
		// 分批处理每个空间的数据
		offset := 0
		for {
			err := db.WithTransaction(config.DB, func(tx *sql.Tx) error {
				// 设置更短的超时时间
				_, err := tx.Exec("PRAGMA busy_timeout = 2000")
				if err != nil {
					return err
				}

				// 只处理一批数据
				result, err := db.ExecTx(tx, `
					DELETE FROM nlip_clipboard_items 
					WHERE id IN (
						SELECT id FROM nlip_clipboard_items 
						WHERE space_id = ? AND created_at < ?
						LIMIT ?
					)
				`, space.ID, space.ExpireTime, batchSize)
				if err != nil {
					return err
				}

				count, _ := result.RowsAffected()
				if count == 0 {
					return sql.ErrNoRows // 用于跳出循环
				}

				return nil
			})

			if err == sql.ErrNoRows {
				break // 没有更多数据需要处理
			}
			if err != nil {
				logger.Error("处理空间 %s 的第 %d 批数据失败: %v",
					space.ID, offset/batchSize, err)
				// 继续处理下一个空间
				break
			}

			offset += batchSize
			// 添加短暂延迟，让其他操作有机会获取锁
			time.Sleep(10 * time.Millisecond)
		}
	}
	return nil
}

// 抽取查询空间信息的函数
type spaceInfo struct {
	ID         string
	ExpireTime time.Time
}

func getSpacesInfo() ([]spaceInfo, error) {
	rows, err := db.QueryRows(config.DB, `
		SELECT id, retention_days 
		FROM nlip_spaces
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spaces []spaceInfo
	for rows.Next() {
		var space spaceInfo
		var retentionDays int
		if err := rows.Scan(&space.ID, &retentionDays); err != nil {
			return nil, err
		}
		space.ExpireTime = time.Now().AddDate(0, 0, -retentionDays)
		spaces = append(spaces, space)
	}
	return spaces, nil
}

// cleanOverflowItems 修改建议
func cleanOverflowItems() error {
	logger.Debug("开始清理超量内容")

	// 先查询所有空间信息，不放在事务中
	spaces, err := getSpacesWithMaxItems()
	if err != nil {
		return fmt.Errorf("查询空间信息失败: %w", err)
	}

	for _, space := range spaces {
		// 对每个空间分批处理
		err := cleanSingleSpaceOverflow(space.ID, space.MaxItems)
		if err != nil {
			logger.Error("清理空间 %s 失败: %v", space.ID, err)
			continue
		}
	}

	logger.Debug("超量内容清理完成")
	return nil
}

// 新增函数：获取空间信息
type spaceMaxItems struct {
	ID       string
	MaxItems int
}

func getSpacesWithMaxItems() ([]spaceMaxItems, error) {
	rows, err := db.QueryRows(config.DB, `
		SELECT id, max_items 
		FROM nlip_spaces
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spaces []spaceMaxItems
	for rows.Next() {
		var space spaceMaxItems
		if err := rows.Scan(&space.ID, &space.MaxItems); err != nil {
			return nil, err
		}
		spaces = append(spaces, space)
	}
	return spaces, nil
}

// cleanSingleSpaceOverflow 修改建议
func cleanSingleSpaceOverflow(spaceID string, maxItems int) error {
	const batchSize = 50

	// 先检查当前条目数量
	var totalItems int
	err := config.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM nlip_clipboard_items 
		WHERE space_id = ?
	`, spaceID).Scan(&totalItems)
	if err != nil {
		return fmt.Errorf("查询条目总数失败: %w", err)
	}

	if totalItems <= maxItems {
		logger.Debug("空间 %s 当前条目数 %d 未超过限制 %d，无需清理",
			spaceID, totalItems, maxItems)
		return nil
	}

	needToDelete := totalItems - maxItems
	logger.Info("空间 %s 需要清理 %d 条记录（当前: %d, 最大: %d）",
		spaceID, needToDelete, totalItems, maxItems)

	totalCleaned := 0
	batchCount := 0

	for totalCleaned < needToDelete {
		batchCount++
		logger.Debug("处理第 %d 批数据", batchCount)

		// 计算本批次要删除的数量
		currentBatchSize := min(batchSize, needToDelete-totalCleaned)

		err = db.WithTransaction(config.DB, func(tx *sql.Tx) error {
			_, err := tx.Exec("PRAGMA busy_timeout = 1000")
			if err != nil {
				return fmt.Errorf("设置事务超时失败: %w", err)
			}

			// 先获取要删除的文件路径
			var filePaths []string
			rows, err := tx.Query(`
				SELECT file_path 
				FROM nlip_clipboard_items 
				WHERE space_id = ? 
				ORDER BY created_at ASC 
				LIMIT ?
			`, spaceID, currentBatchSize)
			if err != nil {
				return fmt.Errorf("查询文件路径失败: %w", err)
			}
			defer rows.Close()

			for rows.Next() {
				var filePath sql.NullString
				if err := rows.Scan(&filePath); err != nil {
					return fmt.Errorf("扫描文件路径失败: %w", err)
				}
				if filePath.Valid && filePath.String != "" {
					filePaths = append(filePaths, filePath.String)
				}
			}

			// 执行删除操作，限制删除数量
			result, err := tx.Exec(`
				DELETE FROM nlip_clipboard_items 
				WHERE id IN (
					SELECT id 
					FROM nlip_clipboard_items 
					WHERE space_id = ? 
					ORDER BY created_at ASC 
					LIMIT ?
				)
			`, spaceID, currentBatchSize)
			if err != nil {
				return fmt.Errorf("删除记录失败: %w", err)
			}

			count, _ := result.RowsAffected()
			if count > 0 {
				totalCleaned += int(count)
				logger.Debug("空间 %s 第 %d 批次清理了 %d 条记录", spaceID, batchCount, count)

				// 在事务外删除文件
				go func(paths []string) {
					for _, path := range paths {
						if err := storage.DeleteFile(path); err != nil {
							logger.Error("删除文件失败 %s: %v", path, err)
						}
					}
				}(filePaths)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("处理第 %d 批数据失败: %w", batchCount, err)
		}

		if totalCleaned >= needToDelete {
			break
		}

		time.Sleep(20 * time.Millisecond)
	}

	if totalCleaned > 0 {
		logger.Info("空间 %s 清理完成，共清理 %d 条超量内容，处理了 %d 个批次",
			spaceID, totalCleaned, batchCount)
	} else {
		logger.Debug("空间 %s 无需清理超量内容", spaceID)
	}

	return nil
}

// 新增函数：获取要删除的文件
func getFilesToDelete(spaceID string, offset, limit int) ([]string, error) {
	rows, err := db.QueryRows(config.DB, `
		SELECT file_path 
		FROM nlip_clipboard_items 
		WHERE space_id = ? AND file_path IS NOT NULL
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`, spaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filePaths []string
	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			return nil, err
		}
		if filePath != "" {
			filePaths = append(filePaths, filePath)
		}
	}
	return filePaths, nil
}

// CleanSpaceOverflow 清理指定空间超出数量限制的内容
func CleanSpaceOverflow(spaceID string) error {
	// 添加重试机制
	maxRetries := 3
	retryDelay := 100 * time.Millisecond

	// 先查询空间的 maxItems
	var maxItems int
	err := config.DB.QueryRow("SELECT max_items FROM nlip_spaces WHERE id = ?", spaceID).Scan(&maxItems)
	if err != nil {
		return fmt.Errorf("获取空间信息失败: %w", err)
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		err := runWithLock(func() error {
			return cleanSingleSpaceOverflow(spaceID, maxItems)
		})

		if err == nil {
			return nil
		}

		lastErr = err
		logger.Warning("清理空间超量内容重试 %d/%d: %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
		retryDelay *= 2 // 指数退避
	}

	return lastErr
}
