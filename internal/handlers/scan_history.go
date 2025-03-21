package handlers

import (
	"net/http"
	"k41m_backend/internal/services"

	"github.com/gin-gonic/gin"
)

func GetScanHistoryHandler(c *gin.Context) {
	// クエリパラメータからフェーズを取得
	phase := c.DefaultQuery("phase", "")

	// サービス層でデータを取得
	history, err := services.GetScanHistory(phase)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve scan history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scan_history": history})
}
