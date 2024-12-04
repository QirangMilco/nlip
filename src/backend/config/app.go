package config

import (
    "os"
    "path/filepath"
    "time"
    "nlip/utils/logger"
    "encoding/json"
    "strings"
    "sync"
)

type Config struct {
    AppEnv      string
    JWTSecret   string
    TokenExpiry time.Duration
    UploadDir   string
    MaxFileSize int64
    ServerPort  string
    FileUpload struct {
        MaxSize       int64    `json:"max_size"`
        AllowedTypes  []string `json:"allowed_types"`
    }
    FileTypes struct {
        AllowList []string `json:"allow_list"`
        DenyList  []string `json:"deny_list"`
    } `json:"file_types"`
    Space struct {
        DefaultMaxItems          int `json:"default_max_items"`      // 空间默认最大条目数
        DefaultRetentionDays    int `json:"default_retention_days"` // 空间默认保留天数
        MaxItemsLimit          int `json:"max_items_limit"`        // 允许设置的空间最大条目数上限
        MaxRetentionDaysLimit  int `json:"max_retention_days_limit"` // 允许设置的空间保留天数上限
    } `json:"space"`
}

var (
    AppConfig   Config
    configMutex sync.Mutex
)

// 添加默认支持的文件类型
var defaultAllowedExtensions = []string{
    // 文本文件
    "txt", "md", "json", "xml", "csv", "log",
    "html", "htm", "css", "js", "ts", "yaml", "yml",
    // 图片文件
    "jpg", "jpeg", "png", "gif", "bmp", "webp", "svg",
}

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

    // 设置文件上传配置
    AppConfig.FileUpload = struct {
        MaxSize      int64    `json:"max_size"`
        AllowedTypes []string `json:"allowed_types"`
    }{
        MaxSize: 10 * 1024 * 1024, // 10MB
        AllowedTypes: []string{
            "image/*",
            "text/*",
            "application/pdf",
        },
    }

    // 设置默认的文件类型配置
    AppConfig.FileTypes.AllowList = defaultAllowedExtensions
    AppConfig.FileTypes.DenyList = []string{}

    // 设置空间相关默认配置
    AppConfig.Space = struct {
        DefaultMaxItems         int `json:"default_max_items"`
        DefaultRetentionDays    int `json:"default_retention_days"`
        MaxItemsLimit          int `json:"max_items_limit"`
        MaxRetentionDaysLimit  int `json:"max_retention_days_limit"`
    }{
        DefaultMaxItems:         20,    // 默认最大条目数
        DefaultRetentionDays:    7,     // 默认保留7天
        MaxItemsLimit:          100,   // 最大允许100条
        MaxRetentionDaysLimit:  30,    // 最大允许保留30天
    }

    // 从配置文件加载自定义文件类型设置
    if configFile := getEnv("CONFIG_FILE", "config.dev.json"); configFile != "" {
        if data, err := os.ReadFile(configFile); err == nil {
            var fileConfig struct {
                FileTypes struct {
                    AllowList []string `json:"allow_list"`
                    DenyList  []string `json:"deny_list"`
                } `json:"file_types"`
            }
            
            if err := json.Unmarshal(data, &fileConfig); err == nil {
                // 如果配置文件中指定了白名单，则覆盖默认配置
                if len(fileConfig.FileTypes.AllowList) > 0 {
                    AppConfig.FileTypes.AllowList = fileConfig.FileTypes.AllowList
                }
                // 添加黑名单
                AppConfig.FileTypes.DenyList = fileConfig.FileTypes.DenyList
            }
        }
    }

    // 从配置文件加载自定义空间设置
    if configFile := getEnv("CONFIG_FILE", "config.dev.json"); configFile != "" {
        if data, err := os.ReadFile(configFile); err == nil {
            var spaceConfig struct {
                Space struct {
                    DefaultMaxItems         int `json:"default_max_items"`
                    DefaultRetentionDays    int `json:"default_retention_days"`
                    MaxItemsLimit          int `json:"max_items_limit"`
                    MaxRetentionDaysLimit  int `json:"max_retention_days_limit"`
                } `json:"space"`
            }
            
            if err := json.Unmarshal(data, &spaceConfig); err == nil {
                // 先更新上限值
                if spaceConfig.Space.MaxItemsLimit > 0 {
                    AppConfig.Space.MaxItemsLimit = spaceConfig.Space.MaxItemsLimit
                }
                if spaceConfig.Space.MaxRetentionDaysLimit > 0 {
                    AppConfig.Space.MaxRetentionDaysLimit = spaceConfig.Space.MaxRetentionDaysLimit
                }

                // 更新默认值，确保不超过上限
                if spaceConfig.Space.DefaultMaxItems > 0 {
                    if spaceConfig.Space.DefaultMaxItems > AppConfig.Space.MaxItemsLimit {
                        AppConfig.Space.DefaultMaxItems = AppConfig.Space.MaxItemsLimit
                    } else {
                        AppConfig.Space.DefaultMaxItems = spaceConfig.Space.DefaultMaxItems
                    }
                }
                if spaceConfig.Space.DefaultRetentionDays > 0 {
                    if spaceConfig.Space.DefaultRetentionDays > AppConfig.Space.MaxRetentionDaysLimit {
                        AppConfig.Space.DefaultRetentionDays = AppConfig.Space.MaxRetentionDaysLimit
                    } else {
                        AppConfig.Space.DefaultRetentionDays = spaceConfig.Space.DefaultRetentionDays
                    }
                }
            }
        }
    }

    // 最后确保默认值不超过上限
    if AppConfig.Space.DefaultMaxItems > AppConfig.Space.MaxItemsLimit {
        AppConfig.Space.DefaultMaxItems = AppConfig.Space.MaxItemsLimit
    }
    if AppConfig.Space.DefaultRetentionDays > AppConfig.Space.MaxRetentionDaysLimit {
        AppConfig.Space.DefaultRetentionDays = AppConfig.Space.MaxRetentionDaysLimit
    }

    // 规范化文件扩展名格式
    for i, ext := range AppConfig.FileTypes.AllowList {
        AppConfig.FileTypes.AllowList[i] = strings.ToLower(strings.TrimPrefix(ext, "."))
    }
    for i, ext := range AppConfig.FileTypes.DenyList {
        AppConfig.FileTypes.DenyList[i] = strings.ToLower(strings.TrimPrefix(ext, "."))
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

// 添加更新配置的函数
func UpdateConfig(updates map[string]interface{}) error {
    // 使用互斥锁保护配置更新
    configMutex.Lock()
    defer configMutex.Unlock()

    if fileTypes, ok := updates["file_types"].(map[string]interface{}); ok {
        if allowList, ok := fileTypes["allow_list"].([]string); ok {
            AppConfig.FileTypes.AllowList = allowList
        }
        if denyList, ok := fileTypes["deny_list"].([]string); ok {
            AppConfig.FileTypes.DenyList = denyList
        }
    }

    if spaceSettings, ok := updates["space_defaults"].(map[string]interface{}); ok {
        // 先处理上限值的更新
        if maxItemsLimit, ok := spaceSettings["max_items_limit"].(int); ok {
            AppConfig.Space.MaxItemsLimit = maxItemsLimit
        }
        if maxRetentionDaysLimit, ok := spaceSettings["max_retention_days_limit"].(int); ok {
            AppConfig.Space.MaxRetentionDaysLimit = maxRetentionDaysLimit
        }

        // 再处理默认值的更新，确保不超过上限
        if maxItems, ok := spaceSettings["max_items"].(int); ok {
            if maxItems <= AppConfig.Space.MaxItemsLimit {
                AppConfig.Space.DefaultMaxItems = maxItems
            } else {
                AppConfig.Space.DefaultMaxItems = AppConfig.Space.MaxItemsLimit
            }
        }
        if retentionDays, ok := spaceSettings["retention_days"].(int); ok {
            if retentionDays <= AppConfig.Space.MaxRetentionDaysLimit {
                AppConfig.Space.DefaultRetentionDays = retentionDays
            } else {
                AppConfig.Space.DefaultRetentionDays = AppConfig.Space.MaxRetentionDaysLimit
            }
        }
    }

    // 保存更新后的配置到文件
    return SaveConfig()
}

// 添加保存配置的函数
func SaveConfig() error {
    configData, err := json.MarshalIndent(AppConfig, "", "    ")
    if err != nil {
        return err
    }

    configFile := getEnv("CONFIG_FILE", "config.dev.json")
    return os.WriteFile(configFile, configData, 0644)
} 