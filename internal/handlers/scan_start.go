package handlers

import (
	"net/http"
	"time"
	"k41m_backend/internal/services"
	"k41m_backend/models"
	"k41m_backend/internal/constants"

	"github.com/gin-gonic/gin"
)

func StartScanHandler(c *gin.Context) {
	var req struct {
		Phase   string                 `json:"phase"`
		Tool    string                 `json:"tool"`
		Options map[string]interface{} `json:"options"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// スキャン概要の登録
	scan := models.ScanSummary{
		Phase:    constants.PhaseMap[req.Phase],
		Tool:     req.Tool,
		ScanTime: time.Now(),
		Status:   "running",
	}
	if err := services.SaveScanSummary(&scan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize scan"})
		return
	}

	// フェーズごとに処理を分岐
	go func(scanID int, phase, tool string, options map[string]interface{}) {
		var err error

		switch phase {
		case "Build":
			err = services.RunBuildScan(scanID, tool, options)
		case "Deploy":
			err = services.RunDeployScan(scanID, tool, options)
		case "Post-Deploy":
			err = services.RunPostDeployScan(scanID, tool, options)
		default:
			err = services.UpdateScanStatus(scanID, "failed")
		}

		if err != nil {
			services.UpdateScanStatus(scanID, "failed")
			return
		}

		services.UpdateScanStatus(scanID, "completed")
	}(int(scan.ID), req.Phase, req.Tool, req.Options)

	c.JSON(http.StatusOK, gin.H{"scan_id": scan.ID, "status": "running"})
}