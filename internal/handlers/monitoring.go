package handlers

import (
	"net/http"
	"k41m_backend/internal/services"

	"github.com/gin-gonic/gin"
)

// ツールから通知を受け取り、データベースに保存する
// 現状falcoからの通知のみを想定しているため、他ツールを使う場合には拡張が必要
func ReceiveMonitorNotificationHandler(c *gin.Context) {
	var payload map[string]interface{}

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// 必須フィールドを抽出
	tool, ok := payload["tool"].(string)
	if !ok || tool == "" {
		tool = "Falco" // 今の所falcoのみを想定
	}

	ruleName, ok := payload["rule"].(string)
	if !ok || ruleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid 'rule' field"})
		return
	}

	receivedAt, ok := payload["time"].(string)
	if !ok || receivedAt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid 'time' field"})
		return
	}

	//  不要なフィールドを削除し、その他のデータを details に格納
	delete(payload, "rule")
	delete(payload, "time")

	err := services.SaveMonitorNotification(tool, ruleName, receivedAt, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification received and saved"})
}


func GetMonitorDetailsHandler(c *gin.Context) {
	results, err := services.GetMonitorDetails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monitor details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"monitor_detail": results,
	})
}