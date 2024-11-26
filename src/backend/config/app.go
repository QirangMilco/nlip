package config

import (
    "os"
    "path/filepath"
    "time"
    "nlip/utils/logger"
    "encoding/json"
)

type Config struct {
    AppEnv      string
    JWTSecret   string
    TokenExpiry time.Duration
    UploadDir   string
    MaxFileSize int64
    ServerPort  string
}

var AppConfig Config

// LoadConfig 加载配置
func LoadConfig() {
    // 获取当前工作目录
    workDir, err := os.Getwd()
    if err != nil {
        logger.Error("获取工作目录失败: %v", err)
        workDir = "."
    }

    // 设置默认配置
    AppConfig = Config{
        AppEnv:      getEnv("APP_ENV", "development"),
        JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
        TokenExpiry: 24 * time.Hour,
        UploadDir:   getEnv("UPLOAD_DIR", filepath.Join(workDir, "uploads")),
        MaxFileSize: 10 * 1024 * 1024, // 10MB
        ServerPort:  getEnv("PORT", "3000"),
    }

    // 根据环境加载不同配置
    switch AppConfig.AppEnv {
    case "development":
        loadDevConfig()
    case "production":
        loadProdConfig()
    case "test":
        loadTestConfig()
    }

    logger.Info("配置加载完成: env=%s, port=%s", AppConfig.AppEnv, AppConfig.ServerPort)
}

// 加载开发环境配置
func loadDevConfig() {
    // 从配置文件加载
    configFile := getEnv("CONFIG_FILE", "config.dev.json")
    if _, err := os.Stat(configFile); err == nil {
        file, err := os.ReadFile(configFile)
        if err != nil {
            logger.Error("读取配置文件失败: %v", err)
            return
        }

        var config struct {
            JWTSecret   string        `json:"jwt_secret"`
            TokenExpiry string        `json:"token_expiry"`
            UploadDir   string        `json:"upload_dir"`
            MaxFileSize int64         `json:"max_file_size"`
        }

        if err := json.Unmarshal(file, &config); err != nil {
            logger.Error("解析配置文件失败: %v", err)
            return
        }

        // 更新配置
        if config.JWTSecret != "" {
            AppConfig.JWTSecret = config.JWTSecret
        }
        if config.TokenExpiry != "" {
            if duration, err := time.ParseDuration(config.TokenExpiry); err == nil {
                AppConfig.TokenExpiry = duration
            }
        }
        if config.UploadDir != "" {
            AppConfig.UploadDir = config.UploadDir
        }
        if config.MaxFileSize > 0 {
            AppConfig.MaxFileSize = config.MaxFileSize
        }

        logger.Debug("从配置文件加载开发环境配置: %s", configFile)
    }
}

// 加载生产环境配置
func loadProdConfig() {
    // 生产环境优先使用环境变量
    if port := os.Getenv("PORT"); port != "" {
        AppConfig.ServerPort = port
    }
    if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
        AppConfig.JWTSecret = jwtSecret
    }
    if uploadDir := os.Getenv("UPLOAD_DIR"); uploadDir != "" {
        AppConfig.UploadDir = uploadDir
    }
}

// 加载测试环境配置
func loadTestConfig() {
    AppConfig.UploadDir = filepath.Join(os.TempDir(), "nlip-test-uploads")
    AppConfig.ServerPort = "0" // 随机端口
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
} 