package db

import (
	"database/sql"
	"fmt"
	"time"
	"nlip/utils/logger"
)

// Migration 数据库迁移记录
type Migration struct {
	Version   int64
	Name      string
	CreatedAt time.Time
}

// InitMigrationTable 初始化迁移表
func InitMigrationTable(db *sql.DB) error {
	logger.Info("初始化迁移表")
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS nlip_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		logger.Error("创建迁移表失败: %v", err)
		return err
	}
	return nil
}

// GetAppliedMigrations 获取已应用的迁移版本
func GetAppliedMigrations(db *sql.DB) (map[int64]bool, error) {
	logger.Debug("获取已应用的迁移版本")
	rows, err := db.Query("SELECT version FROM nlip_migrations")
	if err != nil {
		logger.Error("查询迁移版本失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			logger.Error("读取迁移版本失败: %v", err)
			return nil, err
		}
		applied[version] = true
		logger.Debug("已应用的迁移版本: %d", version)
	}
	return applied, nil
}

// RecordMigration 记录迁移执行
func RecordMigration(tx *sql.Tx, version int64, name string) error {
	logger.Info("记录迁移: 版本=%d, 名称=%s", version, name)
	_, err := tx.Exec(`
		INSERT INTO nlip_migrations (version, name) 
		VALUES (?, ?)
	`, version, name)
	if err != nil {
		logger.Error("记录迁移失败: %v", err)
	}
	return err
}

// RunMigration 执行迁移
func RunMigration(db *sql.DB, version int64, name string, up func(*sql.Tx) error) error {
	logger.Info("执行迁移: 版本=%d, 名称=%s", version, name)
	return WithTransaction(db, func(tx *sql.Tx) error {
		if err := up(tx); err != nil {
			logger.Error("迁移失败 %d-%s: %v", version, name, err)
			return fmt.Errorf("迁移失败 %d-%s: %w", version, name, err)
		}
		if err := RecordMigration(tx, version, name); err != nil {
			logger.Error("记录迁移失败 %d-%s: %v", version, name, err)
			return fmt.Errorf("记录迁移失败 %d-%s: %w", version, name, err)
		}
		logger.Info("迁移完成: 版本=%d, 名称=%s", version, name)
		return nil
	})
} 