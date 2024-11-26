package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"nlip/config"
	"nlip/utils/logger"
)

// SaveFile 保存文件到本地存储
func SaveFile(data []byte, fileName string) (string, error) {
	// 确保上传目录存在
	if err := os.MkdirAll(config.AppConfig.UploadDir, 0755); err != nil {
		logger.Error("创建上传目录失败: %v", err)
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 生成文件路径
	filePath := filepath.Join(config.AppConfig.UploadDir, fileName)
	logger.Debug("准备保存文件: %s", filePath)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		logger.Error("创建文件失败: %v", err)
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入数据
	if _, err := file.Write(data); err != nil {
		logger.Error("写入文件失败: %v", err)
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	logger.Info("文件保存成功: %s (大小: %d bytes)", filePath, len(data))
	return filePath, nil
}

// DeleteFile 从本地存储删除文件
func DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	logger.Debug("准备删除文件: %s", filePath)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			logger.Warning("要删除的文件不存在: %s", filePath)
			return nil
		}
		logger.Error("删除文件失败: %v", err)
		return fmt.Errorf("删除文件失败: %w", err)
	}

	logger.Info("文件删除成功: %s", filePath)
	return nil
}

// GetFile 从本地存储读取文件
func GetFile(filePath string) ([]byte, error) {
	logger.Debug("准备读取文件: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("打开文件失败: %v", err)
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		logger.Error("读取文件内容失败: %v", err)
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	logger.Debug("文件读取成功: %s (大小: %d bytes)", filePath, len(data))
	return data, nil
}

// CleanupFiles 清理指定目录下的所有文件
func CleanupFiles(dir string) error {
	logger.Info("开始清理目录: %s", dir)

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warning("要清理的目录不存在: %s", dir)
			return nil
		}
		logger.Error("读取目录失败: %v", err)
		return fmt.Errorf("读取目录失败: %w", err)
	}

	var errorCount int
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		if err := DeleteFile(filePath); err != nil {
			errorCount++
			logger.Error("清理文件失败: %s - %v", filePath, err)
		}
	}

	if errorCount > 0 {
		logger.Warning("目录清理完成，但有 %d 个文件删除失败", errorCount)
	} else {
		logger.Info("目录清理完成: %s", dir)
	}

	return nil
}

// EnsureUploadDir 确保上传目录存在
func EnsureUploadDir() error {
	logger.Debug("检查上传目录: %s", config.AppConfig.UploadDir)

	if err := os.MkdirAll(config.AppConfig.UploadDir, 0755); err != nil {
		logger.Error("创建上传目录失败: %v", err)
		return fmt.Errorf("创建上传目录失败: %w", err)
	}

	logger.Info("上传目录就绪: %s", config.AppConfig.UploadDir)
	return nil
} 