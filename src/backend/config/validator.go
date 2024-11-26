package config

import (
	"fmt"
	"os"
	"path/filepath"
	"nlip/utils/logger"
)

// ValidateConfig 验证配置
func ValidateConfig() error {
	logger.Debug("开始验证配置")

	// 验证上传目录
	if AppConfig.UploadDir == "" {
		return fmt.Errorf("上传目录不能为空")
	}

	// 确保上传目录存在
	uploadDir := filepath.Clean(AppConfig.UploadDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 验证JWT密钥
	if AppConfig.JWTSecret == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}

	// 验证文件大小限制
	if AppConfig.MaxFileSize <= 0 {
		return fmt.Errorf("文件大小限制必须大于0")
	}

	// 验证令牌过期时间
	if AppConfig.TokenExpiry <= 0 {
		return fmt.Errorf("令牌过期时间必须大于0")
	}

	logger.Info("配置验证通过")
	return nil
} 