package validator

import (
	"path/filepath"
	"strings"
	"nlip/utils/logger"
	"nlip/config"
)

var (
	// 允许的文件类型
	allowedMimeTypes = map[string]bool{
		"text/plain":                true,
		"text/html":                 true,
		"text/css":                  true,
		"text/javascript":           true,
		"application/json":          true,
		"application/xml":           true,
		"application/pdf":           true,
		"image/jpeg":               true,
		"image/png":                true,
		"image/gif":                true,
		"image/webp":               true,
		"image/svg+xml":            true,
	}

	// 允许的文件扩展名
	allowedExtensions = map[string]bool{
		".txt":  true,
		".html": true,
		".css":  true,
		".js":   true,
		".json": true,
		".xml":  true,
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".svg":  true,
	}
)

// ValidateFileType 验证文件类型
func ValidateFileType(filename string, contentType string) bool {
	logger.Debug("验证文件类型: filename=%s, contentType=%s", filename, contentType)

	// 检查文件扩展名
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if !IsAllowedExtension(filename) {
		logger.Warning("不支持的文件扩展名: %s", ext)
		return false
	}

	logger.Debug("文件类型验证通过")
	return true
}

// ValidateFileName 验证文件名
func ValidateFileName(filename string) bool {
	logger.Debug("验证文件名: %s", filename)

	// 检查文件名长度
	if len(filename) > 255 {
		logger.Warning("文件名过长: %s", filename)
		return false
	}

	// 检查文件名是否包含非法字符
	if strings.ContainsAny(filename, "\\/:*?\"<>|") {
		logger.Warning("文件名包含非法字符: %s", filename)
		return false
	}

	logger.Debug("文件名验证通过")
	return true
}

// IsAllowedMimeType 检查是否是允许的MIME类型
func IsAllowedMimeType(mimeType string) bool {
	mainType := strings.Split(mimeType, ";")[0]
	return allowedMimeTypes[mainType]
}

// IsAllowedExtension 检查是否是允许的文件扩展名
func IsAllowedExtension(filename string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	
	// 首先检查黑名单
	for _, denied := range config.AppConfig.FileTypes.DenyList {
		if ext == denied {
			logger.Warning("文件扩展名在黑名单中: %s", ext)
			return false
		}
	}
	
	// 然后检查白名单
	for _, allowed := range config.AppConfig.FileTypes.AllowList {
		if ext == allowed {
			return true
		}
	}
	
	logger.Warning("不支持的文件扩展名: %s", ext)
	return false
}

// RegisterMimeType 注册新的MIME类型
func RegisterMimeType(mimeType string) {
	allowedMimeTypes[mimeType] = true
	logger.Info("注册新的MIME类型: %s", mimeType)
}

// RegisterExtension 注册新的文件扩展名
func RegisterExtension(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ext = strings.ToLower(ext)
	allowedExtensions[ext] = true
	logger.Info("注册新的文件扩展名: %s", ext)
}

// GetAllowedMimeTypes 获取所有允许的MIME类型
func GetAllowedMimeTypes() []string {
	types := make([]string, 0, len(allowedMimeTypes))
	for mimeType := range allowedMimeTypes {
		types = append(types, mimeType)
	}
	return types
}

// GetAllowedExtensions 获取所有允许的文件扩展名
func GetAllowedExtensions() []string {
	exts := make([]string, 0, len(allowedExtensions))
	for ext := range allowedExtensions {
		exts = append(exts, ext)
	}
	return exts
} 