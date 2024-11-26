package config

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "time"
    "nlip/utils/logger"
    "fmt"
    "os"
    "path/filepath"
)

var DB *sql.DB

func InitDatabase() error {
    logger.Info("初始化数据库")
    
    // 确保数据目录存在
    dataDir := "./data"
    if err := os.MkdirAll(dataDir, 0755); err != nil {
        logger.Error("创建数据目录失败: %v", err)
        return err
    }

    dbPath := filepath.Join(dataDir, "nlip.db")
    var err error
    DB, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        logger.Error("打开数据库失败: %v", err)
        return err
    }

    // 配置连接池
    logger.Debug("配置数据库连接池")
    DB.SetMaxOpenConns(25)                // 最大打开连接数
    DB.SetMaxIdleConns(10)                // 最大空闲连接数
    DB.SetConnMaxLifetime(5 * time.Minute) // 连接最大生命周期
    DB.SetConnMaxIdleTime(3 * time.Minute) // 空闲连接最大生命周期

    // 创建表
    if err := createTables(); err != nil {
        logger.Error("创建数据库表失败: %v", err)
        return err
    }

    // 验证连接
    if err := DB.Ping(); err != nil {
        logger.Error("数据库连接测试失败: %v", err)
        return err
    }

    logger.Info("数据库初始化完成")
    return nil
}

func createTables() error {
    logger.Debug("开始创建数据库表")

    // 用户表
    logger.Debug("创建用户表")
    _, err := DB.Exec(`
        CREATE TABLE IF NOT EXISTS nlip_users (
            id VARCHAR(36) PRIMARY KEY,
            username VARCHAR(50) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            is_admin BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        logger.Error("创建用户表失败: %v", err)
        return err
    }

    // 空间表
    logger.Debug("创建空间表")
    _, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS nlip_spaces (
            id VARCHAR(36) PRIMARY KEY,
            name VARCHAR(50) NOT NULL,
            type VARCHAR(10) NOT NULL,
            owner_id VARCHAR(36),
            max_items INT DEFAULT 20,
            retention_days INT DEFAULT 7,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (owner_id) REFERENCES nlip_users(id)
        )
    `)
    if err != nil {
        logger.Error("创建空间表失败: %v", err)
        return err
    }

    // 剪贴板内容表
    logger.Debug("创建剪贴板内容表")
    _, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS nlip_clipboard_items (
            id VARCHAR(36) PRIMARY KEY,
            space_id VARCHAR(36) NOT NULL,
            content_type VARCHAR(50) NOT NULL,
            content TEXT,
            file_path VARCHAR(255),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (space_id) REFERENCES nlip_spaces(id)
        )
    `)
    if err != nil {
        logger.Error("创建剪贴板内容表失败: %v", err)
        return err
    }

    // 创建索引
    logger.Debug("创建数据库索引")
    indexes := []struct {
        name  string
        table string
        cols  string
    }{
        {"idx_users_username", "nlip_users", "username"},
        {"idx_spaces_owner", "nlip_spaces", "owner_id"},
        {"idx_spaces_type", "nlip_spaces", "type"},
        {"idx_clips_space", "nlip_clipboard_items", "space_id"},
        {"idx_clips_created", "nlip_clipboard_items", "created_at"},
    }

    for _, idx := range indexes {
        _, err := DB.Exec(fmt.Sprintf(`
            CREATE INDEX IF NOT EXISTS %s ON %s (%s)
        `, idx.name, idx.table, idx.cols))
        if err != nil {
            logger.Error("创建索引 %s 失败: %v", idx.name, err)
            return err
        }
    }

    logger.Info("数据库表和索引创建完成")
    return nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() {
    if DB != nil {
        logger.Info("关闭数据库连接")
        if err := DB.Close(); err != nil {
            logger.Error("关闭数据库连接失败: %v", err)
        }
    }
} 