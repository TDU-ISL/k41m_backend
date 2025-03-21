package services

import (
	"scan_backend/models"
	"scan_backend/internal/utils"
	"time"
)

func SaveMonitorNotification(tool, ruleName, receivedAt string, details map[string]interface{}) error {
	parsedTime, err := time.Parse(time.RFC3339, receivedAt)
	if err != nil {
		return err
	}

	notification := models.MonitorNotification{
		Tool:       tool,
		RuleName:   ruleName,
		ReceivedAt: parsedTime,
		Details:    utils.ToJSON(details),
	}

	return db.Create(&notification).Error
}

func GetMonitorDetails() (map[string]interface{}, error) {
	// MonitorNotifications を取得
	var notifications []models.MonitorNotification
	if err := db.Find(&notifications).Error; err != nil {
		return nil, err
	}

	// MappingRules を取得
	var mappingRules []models.MappingRule
	if err := db.Where("tool_name = ?", "Falco").Find(&mappingRules).Error; err != nil {
		return nil, err
	}

	// チェックリスト項目をマッピング
	checklistItemMap := map[uint]map[string]interface{}{}

	for _, notification := range notifications {
		// MappingRule に基づいてチェックリスト項目を特定
		var matchedRule *models.MappingRule
		for _, rule := range mappingRules {
			if rule.ExpectedValue == notification.RuleName {
				matchedRule = &rule
				break
			}
		}

		if matchedRule == nil {
			continue // マッチするルールがない場合はスキップ
		}

		// チェックリスト項目を取得
		var checklistItem models.ChecklistItem
		if err := db.First(&checklistItem, "id = ?", matchedRule.ChecklistItemID).Error; err != nil {
			continue
		}

		// チェックリスト項目がすでにマップに存在するか確認
		if _, exists := checklistItemMap[checklistItem.ID]; !exists {
			// セキュリティスタンダードを取得
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
				"notifications":     []map[string]interface{}{},
				"security_standards": securityStandards,
			}
		}

		// 通知を既存のチェックリスト項目に追加
		checklistItemMap[checklistItem.ID]["notifications"] = append(
			checklistItemMap[checklistItem.ID]["notifications"].([]map[string]interface{}),
			map[string]interface{}{
				"rule_name":   notification.RuleName,
				"received_at": notification.ReceivedAt,
				"tool":		   notification.Tool,
				"details":     utils.ParseJSONString(notification.Details),
			},
		)
	}

	// マップをリスト形式に変換
	checklistItems := []map[string]interface{}{}
	for _, item := range checklistItemMap {
		checklistItems = append(checklistItems, item)
	}

	// 結果を構築
	result := map[string]interface{}{
		"phase": "PostDeploy",
		"result": map[string]interface{}{
			"checklist_items": checklistItems,
		},
	}

	return result, nil
}
