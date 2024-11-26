package db

import (
	"database/sql"
	"fmt"
	"nlip/utils/logger"
)

// QueryRow 执行单行查询
func QueryRow(db *sql.DB, query string, args ...interface{}) *sql.Row {
	logger.Debug("执行单行查询: %s, 参数: %v", query, args)
	return db.QueryRow(query, args...)
}

// QueryRows 执行多行查询
func QueryRows(db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	logger.Debug("执行多行查询: %s, 参数: %v", query, args)
	rows, err := db.Query(query, args...)
	if err != nil {
		logger.Error("执行查询失败: %v", err)
		return nil, fmt.Errorf("执行查询失败: %w", err)
	}
	return rows, nil
}

// Exec 执行更新操作
func Exec(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	logger.Debug("执行更新操作: %s, 参数: %v", query, args)
	result, err := db.Exec(query, args...)
	if err != nil {
		logger.Error("执行更新失败: %v", err)
		return nil, fmt.Errorf("执行更新失败: %w", err)
	}

	if rowsAffected, err := result.RowsAffected(); err == nil {
		logger.Debug("更新影响行数: %d", rowsAffected)
	}

	return result, nil
}

// ExecTx 在事务中执行更新操作
func ExecTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	logger.Debug("在事务中执行更新操作: %s, 参数: %v", query, args)
	result, err := tx.Exec(query, args...)
	if err != nil {
		logger.Error("执行事务更新失败: %v", err)
		return nil, fmt.Errorf("执行事务更新失败: %w", err)
	}

	if rowsAffected, err := result.RowsAffected(); err == nil {
		logger.Debug("事务更新影响行数: %d", rowsAffected)
	}

	return result, nil
} 