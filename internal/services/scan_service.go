package services

import (
	"scan_backend/internal/utils"
	"scan_backend/models"
)

func SaveScanSummary(scan *models.ScanSummary) error {
	return db.Create(scan).Error
}

func UpdateScanStatus(scanID int, status string) error {
	return db.Model(&models.ScanSummary{}).Where("id = ?", scanID).Update("status", status).Error
}

// リファクタリング前のコード
func GetBuildScanDetails(scanID string) (map[string]interface{}, error) {
	var scan models.ScanSummary

	// スキャン概要を取得
	if err := db.First(&scan, "id = ?", scanID).Error; err != nil {
		return nil, err
	}

	var targets []models.ScanTarget
	db.Where("scan_summary_id = ?", scanID).Find(&targets)

	results := []map[string]interface{}{}

	for _, target := range targets {
		checklistItemMap := map[uint]map[string]interface{}{}

		// 各ターゲットに紐付くコントロールを取得
		var controls []models.ScanControl
		db.Where("scan_target_id = ?", target.ID).Find(&controls)

		for _, control := range controls {
			// マッピングルールからチェックリスト項目を特定
			var mappings []models.MappingRule
			db.Where("tool_name = ? AND expected_value = ?", scan.Tool, control.ControlID).Find(&mappings)

			for _, mapping := range mappings {
				// チェックリスト項目情報を取得
				var checklistItem models.ChecklistItem
				if err := db.First(&checklistItem, "id = ?", mapping.ChecklistItemID).Error; err != nil {
					continue
				}

				// チェックリスト項目がすでにマップに存在するか確認
				if _, exists := checklistItemMap[checklistItem.ID]; !exists {
					// セキュリティスタンダードとセクションを取得
					var standards []models.ChecklistItemSecurityStandard
					db.Preload("Sections").Preload("SecurityStandard").Where("checklist_item_id = ?", checklistItem.ID).Find(&standards)

					securityStandards := []map[string]interface{}{}
					for _, standard := range standards {
						sections := []string{}
						for _, section := range standard.Sections {
							sections = append(sections, section.Section)
						}
						securityStandards = append(securityStandards, map[string]interface{}{
							"ChecklistItemID":   standard.ChecklistItemID,
							"ID":                standard.ID,
							"Sections":          sections,
							"SecurityStandardID": standard.SecurityStandardID,
							"Name":              standard.SecurityStandard.Name,
							"Version":           standard.SecurityStandard.Version,
						})
					}

					// 新しいチェックリスト項目を作成
					checklistItemMap[checklistItem.ID] = map[string]interface{}{
						"checklist_item_id": checklistItem.ID,
						"phase":             checklistItem.Phase,
						"category":          checklistItem.Category,
						"subcategory":       checklistItem.Subcategory,
						"description":       checklistItem.Description,
						"severity":          checklistItem.Severity,
						"controls":          []map[string]interface{}{},
						"security_standards": securityStandards,
					}
				}

				// 現在のコントロールを既存のチェックリスト項目に追加
				checklistItemMap[checklistItem.ID]["controls"] = append(
					checklistItemMap[checklistItem.ID]["controls"].([]map[string]interface{}),
					map[string]interface{}{
						"control_id":   control.ControlID,
						"control_name": control.ControlName,
						"details":      utils.ParseJSONString(control.ControlDetails),
						"id":           control.ID,
						"status":       control.ControlStatus,
					},
				)
			}
		}

		// マップをリスト形式に変換
		checklistItems := []map[string]interface{}{}
		for _, item := range checklistItemMap {
			checklistItems = append(checklistItems, item)
		}

		// 対象情報をレスポンスに追加
		results = append(results, map[string]interface{}{
			"target_name":     target.TargetName,
			"metadata":        utils.ParseJSONString(target.TargetMetadata),
			"checklist_items": checklistItems,
		})
	}

	return map[string]interface{}{
		"scan_id": scan.ID,
		"phase":   scan.Phase,
		"tool":    scan.Tool,
		"status":  scan.Status,
		"results": results,
	}, nil
}

func GetDeployScanDetails(scanID string) (map[string]interface{}, error) {
	var scan models.ScanSummary

	// スキャン概要を取得
	if err := db.First(&scan, "id = ?", scanID).Error; err != nil {
		return nil, err
	}

	var targets []models.ScanTarget
	db.Where("scan_summary_id = ?", scanID).Find(&targets)

	results := []map[string]interface{}{}

	for _, target := range targets {
		checklistItemMap := map[uint]map[string]interface{}{}

		// 各ターゲットに紐付くコントロールを取得
		var controls []models.ScanControl
		db.Where("scan_target_id = ?", target.ID).Find(&controls)

		for _, control := range controls {
			// マッピングルールからチェックリスト項目を特定
			var mappings []models.MappingRule
			db.Where("tool_name = ? AND expected_value = ?", scan.Tool, control.ControlID).Find(&mappings)

			for _, mapping := range mappings {
				// チェックリスト項目情報を取得
				var checklistItem models.ChecklistItem
				if err := db.First(&checklistItem, "id = ?", mapping.ChecklistItemID).Error; err != nil {
					continue
				}

				// チェックリスト項目がすでにマップに存在するか確認
				if _, exists := checklistItemMap[checklistItem.ID]; !exists {
					// セキュリティスタンダードとセクションを取得
					var standards []models.ChecklistItemSecurityStandard
					db.Preload("Sections").Preload("SecurityStandard").Where("checklist_item_id = ?", checklistItem.ID).Find(&standards)

					securityStandards := []map[string]interface{}{}
					for _, standard := range standards {
						sections := []string{}
						for _, section := range standard.Sections {
							sections = append(sections, section.Section)
						}
						securityStandards = append(securityStandards, map[string]interface{}{
							"ChecklistItemID":   standard.ChecklistItemID,
							"ID":                standard.ID,
							"Sections":          sections,
							"SecurityStandardID": standard.SecurityStandardID,
							"Name":              standard.SecurityStandard.Name,
							"Version":           standard.SecurityStandard.Version,
						})
					}

					// 新しいチェックリスト項目を作成
					checklistItemMap[checklistItem.ID] = map[string]interface{}{
						"checklist_item_id": checklistItem.ID,
						"phase":             checklistItem.Phase,
						"category":          checklistItem.Category,
						"subcategory":       checklistItem.Subcategory,
						"description":       checklistItem.Description,
						"severity":          checklistItem.Severity,
						"controls":          []map[string]interface{}{},
						"security_standards": securityStandards,
						"status":            "",
					}
				}

				// 現在のコントロールを既存のチェックリスト項目に追加
				checklistItemMap[checklistItem.ID]["controls"] = append(
					checklistItemMap[checklistItem.ID]["controls"].([]map[string]interface{}),
					map[string]interface{}{
						"control_id":   control.ControlID,
						"control_name": control.ControlName,
						"details":      utils.ParseJSONString(control.ControlDetails),
						"status":       control.ControlStatus, // コントロールのステータス
					},
				)
			}
		}

		// マップをリスト形式に変換し、statusを計算
		checklistItems := []map[string]interface{}{}
		for _, item := range checklistItemMap {
			// controls のステータスを集計して checklist_item のステータスを設定
			controls := item["controls"].([]map[string]interface{})
			item["status"] = utils.CalculateChecklistItemStatus(controls)

			checklistItems = append(checklistItems, item)
		}

		// 対象情報をレスポンスに追加
		results = append(results, map[string]interface{}{
			"target_name":     target.TargetName,
			"metadata":        utils.ParseJSONString(target.TargetMetadata),
			"checklist_items": checklistItems,
		})
	}

	return map[string]interface{}{
		"scan_id": scan.ID,
		"phase":   scan.Phase,
		"tool":    scan.Tool,
		"status":  scan.Status,
		"results": results,
	}, nil
}
