package ws

import (
    "encoding/json"
    "github.com/gofiber/websocket/v2"
    "nlip/utils/jwt"
    "nlip/utils/logger"
    "time"
)

type WSMessage struct {
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
    Timestamp int64       `json:"timestamp"`
}

func HandleWebSocket(c *websocket.Conn) {
    // 获取token
    token := c.Query("token")
    if token == "" {
        logger.Warning("WebSocket连接缺少token")
        return
    }

    // 验证token
    claims, err := jwt.ValidateToken(token)
    if err != nil {
        logger.Warning("WebSocket token验证失败: %v", err)
        return
    }

    logger.Info("WebSocket连接建立: userID=%s", claims.UserID)

    // 保持连接
    for {
        messageType, message, err := c.ReadMessage()
        if err != nil {
            logger.Warning("读取WebSocket消息失败: %v", err)
            break
        }

        if messageType == websocket.TextMessage {
            // 处理消息
            var msg WSMessage
            if err := json.Unmarshal(message, &msg); err != nil {
                logger.Warning("解析WebSocket消息失败: %v", err)
                continue
            }

            // 处理不同类型的消息
            switch msg.Type {
            case "ping":
                // 发送pong响应
                response := WSMessage{
                    Type:      "pong",
                    Timestamp: time.Now().Unix(),
                }
                if err := c.WriteJSON(response); err != nil {
                    logger.Warning("发送WebSocket响应失败: %v", err)
                }
            }
        }
    }
} 