package config

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "time"
    "nlip/utils/logger"
    "fmt"
    "os"
    "path/filepath"
    "golang.org/x/crypto/bcrypt"
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
            need_change_pwd BOOLEAN DEFAULT FALSE,
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
            invited_users TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
            clip_id VARCHAR(15) NOT NULL,
            space_id VARCHAR(36) NOT NULL,
            content_type VARCHAR(50) NOT NULL,
            content TEXT,
            file_path VARCHAR(255),
            creator_id VARCHAR(36),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (space_id) REFERENCES nlip_spaces(id),
            FOREIGN KEY (creator_id) REFERENCES nlip_users(id),
            UNIQUE (space_id, clip_id)
        )
    `)
    if err != nil {
        logger.Error("创建剪贴板内容表失败: %v", err)
        return err
    }

    // 邀请表
    logger.Debug("创建邀请表")
    _, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS nlip_invites (
            token_hash VARCHAR(64) PRIMARY KEY,
            space_id VARCHAR(32) NOT NULL,
            permission VARCHAR(32) NOT NULL,
            created_by VARCHAR(32) NOT NULL,
            created_at DATETIME NOT NULL,
            expires_at DATETIME NOT NULL,
            used_at DATETIME,
            used_by VARCHAR(32),
            FOREIGN KEY (space_id) REFERENCES nlip_spaces(id),
            FOREIGN KEY (created_by) REFERENCES nlip_users(id),
            FOREIGN KEY (used_by) REFERENCES nlip_users(id)
        )
    `)
    if err != nil {
        logger.Error("创建邀请表失败: %v", err)
        return err
    }

    // 创建触发器，自动更新 updated_at 字段
    logger.Debug("创建更新时间触发器")
    _, err = DB.Exec(`
        CREATE TRIGGER IF NOT EXISTS update_spaces_timestamp 
        AFTER UPDATE ON nlip_spaces
        BEGIN
            UPDATE nlip_spaces 
            SET updated_at = CURRENT_TIMESTAMP 
            WHERE id = NEW.id;
        END;
    `)
    if err != nil {
        logger.Error("创建空间更新触发器失败: %v", err)
        return err
    }

    _, err = DB.Exec(`
        CREATE TRIGGER IF NOT EXISTS update_clips_timestamp 
        AFTER UPDATE ON nlip_clipboard_items
        BEGIN
            UPDATE nlip_clipboard_items 
            SET updated_at = CURRENT_TIMESTAMP 
            WHERE id = NEW.id;
        END;
    `)
    if err != nil {
        logger.Error("创建剪贴板更新触发器失败: %v", err)
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
        {"idx_spaces_timestamps", "nlip_spaces", "created_at, updated_at"},
        {"idx_clips_space", "nlip_clipboard_items", "space_id"},
        {"idx_clips_creator", "nlip_clipboard_items", "creator_id"},
        {"idx_clips_timestamps", "nlip_clipboard_items", "created_at, updated_at"},
        {"idx_invites_token", "nlip_invites", "token_hash"},
        {"idx_invites_space", "nlip_invites", "space_id"},
        {"idx_invites_expires", "nlip_invites", "expires_at"},
        {"idx_invited_users", "nlip_spaces", "(JSON_EXTRACT(invited_users, '$'))"},
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

    // 检查是否需要创建默认公共空间
    var count int
    err = DB.QueryRow("SELECT COUNT(*) FROM nlip_spaces WHERE type = 'public'").Scan(&count)
    if err != nil {
        logger.Error("检查公共空间失败: %v", err)
        return err
    }

    if count == 0 {
        logger.Info("创建默认公共空间")
        const defaultSpaceID = "public-space"
        _, err = DB.Exec(`
            INSERT INTO nlip_spaces (id, name, type, owner_id, max_items, retention_days) 
            VALUES (?, ?, ?, ?, ?, ?)
        `, defaultSpaceID, "公共空间", "public", "system", 20, 7)
        
        if err != nil {
            logger.Error("创建默认公共空间失败: %v", err)
            return err
        }
        logger.Info("默认公共空间创建成功: id=%s", defaultSpaceID)
    }

    // 检查是否需要创建管理员账号
    var adminCount int
    err = DB.QueryRow("SELECT COUNT(*) FROM nlip_users WHERE is_admin = TRUE").Scan(&adminCount)
    if err != nil {
        logger.Error("检查管理员账号失败: %v", err)
        return err
    }

    if adminCount == 0 {
        logger.Info("创建默认管理员账号")
        
        // 生成密码哈希
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte("nlip123"), bcrypt.DefaultCost)
        if err != nil {
            logger.Error("生成密码哈希失败: %v", err)
            return err
        }
        
        // 插入管理员账号
        _, err = DB.Exec(`
            INSERT INTO nlip_users (id, username, password_hash, is_admin, need_change_pwd) 
            VALUES (?, ?, ?, TRUE, TRUE)
        `, "admin-user", "admin", string(hashedPassword))
        
        if err != nil {
            logger.Error("创建管理员账号失败: %v", err)
            return err
        }
        logger.Info("默认管理员账号创建成功")
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