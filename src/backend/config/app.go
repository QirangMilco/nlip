package config

import (
	"encoding/json"
	"fmt"
	"nlip/utils/logger"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AppEnv      string        `json:"app_env"`
	JWTSecret   string        `json:"jwt_secret"`
	TokenExpiry time.Duration `json:"token_expiry"`
	UploadDir   string        `json:"upload_dir"`
	MaxFileSize int64         `json:"max_file_size"`
	ServerPort  string        `json:"server_port"`
	Domain      string        `json:"domain"`
	FrontendURL string        `json:"frontend_url"`

	FileUpload struct {
		MaxSize      int64    `json:"max_size"`
		AllowedTypes []string `json:"allowed_types"`
	} `json:"file_upload"`

	FileTypes struct {
		AllowList []string `json:"allow_list"`
		DenyList  []string `json:"deny_list"`
	} `json:"file_types"`

	Space struct {
		DefaultMaxItems       int `json:"default_max_items"`
		DefaultRetentionDays  int `json:"default_retention_days"`
		MaxItemsLimit         int `json:"max_items_limit"`
		MaxRetentionDaysLimit int `json:"max_retention_days_limit"`
	} `json:"space"`

	Email struct {
		Enabled  bool   `json:"enabled"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		From     string `json:"from"`
	} `json:"email"`

	Token struct {
		MaxItems        int `json:"max_items"`
		DefaultExpiryDays int `json:"default_expiry_days"`
		MaxExpiryDays   int `json:"max_expiry_days"`
	} `json:"token"`
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
	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		logger.Error("获取工作目录失败: %v", err)
		workDir = "."
	}

	// 设置基础默认配置
	AppConfig = Config{
		AppEnv:      getEnv("APP_ENV", "production"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		TokenExpiry: 24 * time.Hour,
		UploadDir:   getEnv("UPLOAD_DIR", filepath.Join(workDir, "uploads")),
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		ServerPort:  getEnv("PORT", "3000"),
		FileUpload: struct {
			MaxSize      int64    `json:"max_size"`
			AllowedTypes []string `json:"allowed_types"`
		}{
			MaxSize: 10 * 1024 * 1024,
			AllowedTypes: []string{
				"image/*",
				"text/*",
				"application/pdf",
			},
		},
		Space: struct {
			DefaultMaxItems       int `json:"default_max_items"`
			DefaultRetentionDays  int `json:"default_retention_days"`
			MaxItemsLimit         int `json:"max_items_limit"`
			MaxRetentionDaysLimit int `json:"max_retention_days_limit"`
		}{
			DefaultMaxItems:       20,
			DefaultRetentionDays:  7,
			MaxItemsLimit:         100,
			MaxRetentionDaysLimit: 30,
		},
		FileTypes: struct {
			AllowList []string `json:"allow_list"`
			DenyList  []string `json:"deny_list"`
		}{
			AllowList: defaultAllowedExtensions,
			DenyList:  []string{},
		},
		Email: struct {
			Enabled  bool   `json:"enabled"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Password string `json:"password"`
			From     string `json:"from"`
		}{
			Enabled: true,
		},
		Token: struct {
			MaxItems        int `json:"max_items"`
			DefaultExpiryDays int `json:"default_expiry_days"`
			MaxExpiryDays int `json:"max_expiry_days"`
		}{
			MaxItems:        10,
			DefaultExpiryDays: 7,
		},
	}

	// 根据环境加载配置
	switch AppConfig.AppEnv {
	case "development":
		loadDevConfig()
	case "production":
		loadProdConfig()
	case "test":
		loadTestConfig()
	default:
		loadProdConfig()
	}

	// 确保配置值在合理范围内
	validateAndAdjustConfig()

	// 设置域名相关配置
	setupDomainConfig()

	// 规范化文件类型列表
	normalizeFileTypes()

	logger.Info("配置加载完成: env=%s, port=%s, domain=%s",
		AppConfig.AppEnv,
		AppConfig.ServerPort,
		AppConfig.Domain,
	)
}

// validateAndAdjustConfig 验证并调整配置值
func validateAndAdjustConfig() {
	// 确保空间配置在合理范围内
	if AppConfig.Space.DefaultMaxItems > AppConfig.Space.MaxItemsLimit {
		AppConfig.Space.DefaultMaxItems = AppConfig.Space.MaxItemsLimit
	}
	if AppConfig.Space.DefaultRetentionDays > AppConfig.Space.MaxRetentionDaysLimit {
		AppConfig.Space.DefaultRetentionDays = AppConfig.Space.MaxRetentionDaysLimit
	}

	// 若设置最大过期天数，则检查默认过期天数是否超过最大过期天数
	if AppConfig.Token.MaxExpiryDays > 0 && AppConfig.Token.DefaultExpiryDays > AppConfig.Token.MaxExpiryDays {
		AppConfig.Token.DefaultExpiryDays = AppConfig.Token.MaxExpiryDays
	}
}

// setupDomainConfig 设置域名相关配置
func setupDomainConfig() {
	protocol := "http"
	if getEnv("HTTPS_ENABLED", "false") == "true" {
		protocol = "https"
	}

	domain := getEnv("DOMAIN", "localhost")
	port := AppConfig.ServerPort

	if (protocol == "http" && port == "80") || (protocol == "https" && port == "443") {
		AppConfig.Domain = fmt.Sprintf("%s://%s", protocol, domain)
	} else {
		AppConfig.Domain = fmt.Sprintf("%s://%s:%s", protocol, domain, port)
	}

	AppConfig.FrontendURL = getEnv("FRONTEND_URL", AppConfig.Domain)
}

// normalizeFileTypes 规范化文件类型列表
func normalizeFileTypes() {
	for i, ext := range AppConfig.FileTypes.AllowList {
		AppConfig.FileTypes.AllowList[i] = strings.ToLower(strings.TrimPrefix(ext, "."))
	}
	for i, ext := range AppConfig.FileTypes.DenyList {
		AppConfig.FileTypes.DenyList[i] = strings.ToLower(strings.TrimPrefix(ext, "."))
	}
}

// loadDevConfig 加载开发环境配置
func loadDevConfig() {
	// 默认使用 yaml 配置
	configFile := getEnv("CONFIG_FILE", "config.yaml")
	if _, err := os.Stat(configFile); err != nil {
		// 如果 yaml 不存在，尝试读取 json 配置
		configFile = "config.dev.json"
	}

	if _, err := os.Stat(configFile); err == nil {
		file, err := os.ReadFile(configFile)
		if err != nil {
			logger.Error("读取配置文件失败: %v", err)
			return
		}

		var config map[string]interface{}
		if strings.HasSuffix(configFile, ".yaml") || strings.HasSuffix(configFile, ".yml") {
			if err = yaml.Unmarshal(file, &config); err != nil {
				logger.Error("解析YAML配置文件失败: %v", err)
				return
			}
		} else {
			if err = json.Unmarshal(file, &config); err != nil {
				logger.Error("解析JSON配置文件失败: %v", err)
				return
			}
		}

		// 更新配置
		if jwtSecret, ok := config["jwt_secret"].(string); ok && jwtSecret != "" {
			AppConfig.JWTSecret = jwtSecret
		}
		if tokenExpiry, ok := config["token_expiry"].(string); ok && tokenExpiry != "" {
			if duration, err := time.ParseDuration(tokenExpiry); err == nil {
				AppConfig.TokenExpiry = duration
			}
		}
		if uploadDir, ok := config["upload_dir"].(string); ok && uploadDir != "" {
			AppConfig.UploadDir = uploadDir
		}
		if maxFileSize, ok := config["max_file_size"].(float64); ok && maxFileSize > 0 {
			AppConfig.MaxFileSize = int64(maxFileSize)
		}

		logger.Debug("从配置文件加载开发环境配置: %s", configFile)
	}
}

// 加载生产环境配置
func loadProdConfig() {
	// 先尝试从配置文件加载
	configFile := getEnv("CONFIG_FILE", "config.yaml")
	if _, err := os.Stat(configFile); err == nil {
		logger.Debug("从配置文件加载生产环境配置: %s", configFile)
		file, err := os.ReadFile(configFile)
		if err == nil {
			var config map[string]interface{}
			if strings.HasSuffix(configFile, ".yaml") || strings.HasSuffix(configFile, ".yml") {
				err = yaml.Unmarshal(file, &config)
			} else {
				err = json.Unmarshal(file, &config)
			}
			if err != nil {
				logger.Error("解析配置文件失败: %v", err)
			}
			// logger.Debug("解析配置文件成功: %v", config)
			// logger.Debug("before AppConfig: %v", AppConfig)
			metadata := &mapstructure.Metadata{}
			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				TagName:          "json",
				WeaklyTypedInput: true,
				Metadata:         metadata,
				Result:           &AppConfig,
			})
			if err != nil {
				logger.Error("创建解码器失败: %v", err)
			}
			err = decoder.Decode(config)
			if err != nil {
				logger.Error("解析配置文件失败: %v", err)
			}
			// logger.Debug("after AppConfig: %v", AppConfig)
			// logger.Debug("metadata: %v", metadata)
		}
	}

	// 环境变量优先级高于配置文件
	if port := os.Getenv("SERVER_PORT"); port != "" {
		AppConfig.ServerPort = port
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		AppConfig.JWTSecret = jwtSecret
	}
	if uploadDir := os.Getenv("UPLOAD_DIR"); uploadDir != "" {
		AppConfig.UploadDir = uploadDir
	}
	if maxFileSize := os.Getenv("MAX_FILE_SIZE"); maxFileSize != "" {
		if size, err := strconv.ParseInt(maxFileSize, 10, 64); err == nil {
			AppConfig.MaxFileSize = size
		}
	}
	if emailHost := os.Getenv("EMAIL_HOST"); emailHost != "" {
		AppConfig.Email.Host = emailHost
	}
	if emailPort := os.Getenv("EMAIL_PORT"); emailPort != "" {
		if port, err := strconv.Atoi(emailPort); err == nil {
			AppConfig.Email.Port = port
		}
	}
	if emailUser := os.Getenv("EMAIL_USERNAME"); emailUser != "" {
		AppConfig.Email.Username = emailUser
	}
	if emailPass := os.Getenv("EMAIL_PASSWORD"); emailPass != "" {
		AppConfig.Email.Password = emailPass
	}
	if emailFrom := os.Getenv("EMAIL_FROM"); emailFrom != "" {
		AppConfig.Email.From = emailFrom
	}
	if emailEnabled := os.Getenv("EMAIL_ENABLED"); emailEnabled != "" {
		AppConfig.Email.Enabled = emailEnabled == "true"
	}

	if tokenMaxItems := os.Getenv("TOKEN_MAX_ITEMS"); tokenMaxItems != "" {
		if items, err := strconv.Atoi(tokenMaxItems); err == nil {
			AppConfig.Token.MaxItems = items
		}
	}

	if tokenDefaultExpiryDays := os.Getenv("TOKEN_DEFAULT_EXPIRY_DAYS"); tokenDefaultExpiryDays != "" {
		if days, err := strconv.Atoi(tokenDefaultExpiryDays); err == nil {
			AppConfig.Token.DefaultExpiryDays = days
		}
	}

	if tokenMaxExpiryDays := os.Getenv("TOKEN_MAX_EXPIRY_DAYS"); tokenMaxExpiryDays != "" {
		if days, err := strconv.Atoi(tokenMaxExpiryDays); err == nil {
			AppConfig.Token.MaxExpiryDays = days
		}
	}

	logger.Info("生产环境配置加载完成")
}

// 加载测试环境配置
func loadTestConfig() {
	AppConfig.UploadDir = filepath.Join(os.TempDir(), "nlip-test-uploads")
	AppConfig.ServerPort = "0" // 随机端口
}

// getEnv 获取环境变量，如果不存在则返回默认值
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

// SaveConfig 保存配置到文件
func SaveConfig() error {
	configFile := getEnv("CONFIG_FILE", "config.yaml")
	data, err := yaml.Marshal(AppConfig)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	return os.WriteFile(configFile, data, 0644)
}
