package db

import (
	"database/sql"
	"fmt"
	"nlip/utils/logger"
)

// WithTransaction 执行事务
func WithTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		logger.Error("开始事务失败: %v", err)
		return fmt.Errorf("开始事务失败: %w", err)
	}

	logger.Debug("事务开始")

	defer func() {
		if p := recover(); p != nil {
			logger.Error("事务发生panic: %v", p)
			tx.Rollback()
			panic(p) // 重新抛出panic
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("回滚事务失败: %v (原始错误: %v)", rbErr, err)
			return fmt.Errorf("回滚事务失败: %v (原始错误: %w)", rbErr, err)
		}
		logger.Warning("事务回滚: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("提交事务失败: %v", err)
		return fmt.Errorf("提交事务失败: %w", err)
	}

	logger.Debug("事务提交成功")
	return nil
} 