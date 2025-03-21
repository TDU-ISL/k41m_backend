package handlers

import (
	"net/http"
	"k41m_backend/internal/services"

	"github.com/gin-gonic/gin"
)

func GetScanDetailsHandler(c *gin.Context) {
	scanID := c.Param("scan_id")
	phase := c.Query("phase")

	var result map[string]interface{}
	var err error

	switch phase {
	case "Build":
		result, err = services.GetBuildScanDetails(scanID)
	case "Deploy":
		result, err = services.GetDeployScanDetails(scanID)
	// // デプロイ後(運用・監視)フェーズの処理は未実装
	// case "Post-Deploy":
	// 	result, err = services.GetPostDeployScanDetails(scanID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phase"})
		return
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scan_detail": result})
}
