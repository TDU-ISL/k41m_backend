package services

import (
	"fmt"
	"k41m_backend/models"
	"k41m_backend/internal/constants"
)

func GetScanHistory(phase string) ([]map[string]interface{}, error) {
	var scanSummaries []models.ScanSummary
	query := db

	// フェーズが指定されている場合はフィルタリング
	if phase != "" {
		mappedPhase, ok := constants.PhaseMap[phase]
		if !ok {
			return nil, fmt.Errorf("Invalid phase value: %s", phase)
		}
		query = query.Where("phase = ?", mappedPhase)
	}

	if err := query.Preload("ScanTargets").Find(&scanSummaries).Error; err != nil {
		return nil, err
	}

	// レスポンス構造を構築
	history := []map[string]interface{}{}
	for _, summary := range scanSummaries {
		targets := []map[string]interface{}{}
		for _, target := range summary.ScanTargets {
			targets = append(targets, map[string]interface{}{
				"id":   target.ID,
				"name": target.TargetName,
			})
		}

		history = append(history, map[string]interface{}{
			"id":      summary.ID,
			"phase":   summary.Phase,
			"tool":    summary.Tool,
			"scan_time": summary.ScanTime,
			"status":  summary.Status,
			"details": summary.Details,
			"targets": targets,
		})
	}

	return history, nil
}
