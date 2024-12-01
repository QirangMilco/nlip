package id

import (
	"crypto/rand"
	"encoding/base32"
	"nlip/config"
	"nlip/utils/logger"
	"strings"
)

// Generate 生成12位短ID
func Generate(byteLen int, base32Len int) string {
	// 生成7字节随机数据(生成12位base32编码)
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		logger.Error("生成随机数据失败: %v", err)
		return ""
	}

	// 使用base32编码(去掉填充=号)
	id := strings.TrimRight(base32.StdEncoding.EncodeToString(b), "=")

	// 返回12位ID
	return strings.ToLower(id[:base32Len])
}

// GenerateSpaceID 生成空间ID前缀
func GenerateSpaceID() string {
	for {
		id := "s_" + Generate(7, 12)

		// 检查ID是否已存在
		var exists bool
		err := config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM nlip_spaces WHERE id = ?)", id).Scan(&exists)
		if err != nil {
			logger.Error("检查空间ID唯一性失败: %v", err)
			continue
		}

		if !exists {
			return id
		}

		logger.Debug("空间ID已存在，重新生成: %s", id)
	}
}

// GenerateClipID 生成剪贴板ID
func GenerateClipID(spaceID string) (string, string) {
	for {
		clipID := Generate(4, 6)
		
		// 组合成完整的ID
		fullID := "c_" + spaceID + "_" + clipID
		
		// 检查同空间内是否存在相同的剪贴板ID
		var exists bool
		err := config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM nlip_clipboard_items WHERE space_id = ? AND clip_id = ?)", spaceID, clipID).Scan(&exists)
		if err != nil {
			logger.Error("检查剪贴板ID唯一性失败: %v", err)
			continue
		}

		if !exists {
			return fullID, clipID
		}

		logger.Debug("剪贴板ID已存在，重新生成: %s", fullID)
	}
}
