package services

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"fmt"

	"k41m_backend/models"
	"k41m_backend/internal/constants"

	"github.com/xuri/excelize/v2"
)

func GetChecklistItems(phase string) ([]map[string]interface{}, error) {
	var checklistItems []models.ChecklistItem
	query := db
	
	// フェーズが指定されている場合はマッピングしてフィルタリング
	if phase != "" {
		mappedPhase, ok := constants.PhaseMap[phase]
		if !ok {
			return nil, fmt.Errorf("Invalid phase value: %s", phase)
		}
		query = query.Where("phase = ?", mappedPhase)
	}

	if err := query.Find(&checklistItems).Error; err != nil {
		return nil, err
	}

	// レスポンス構造を構築
	response := []map[string]interface{}{}
	for _, item := range checklistItems {
		// セキュリティスタンダードを取得
		var standards []models.ChecklistItemSecurityStandard
		db.Preload("Sections").Preload("SecurityStandard").Where("checklist_item_id = ?", item.ID).Find(&standards)

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

		// マッピングルールを取得
		var mappingRules []models.MappingRule
		db.Where("checklist_item_id = ?", item.ID).Find(&mappingRules)

		mappingRulesResp := []map[string]interface{}{}
		for _, rule := range mappingRules {
			mappingRulesResp = append(mappingRulesResp, map[string]interface{}{
				"tool_name":      rule.ToolName,
				"expected_value": rule.ExpectedValue,
			})
		}

		// 各チェックリスト項目をレスポンス形式に整形
		response = append(response, map[string]interface{}{
			"checklist_item_id": item.ID,
			"phase":             item.Phase,
			"category":          item.Category,
			"subcategory":       item.Subcategory,
			"description":       item.Description,
			"severity":          item.Severity,
			"security_standards": securityStandards,
			"mapping_rules":      mappingRulesResp,
		})
	}

	return response, nil
}

// ProcessChecklistFile は、Excel または CSV ファイルを処理し、データを DB に保存する
func ProcessChecklistFile(filePath string) (string, error) {
	ext := filepath.Ext(filePath)
	var csvFilePath string
	var err error

	// ファイル形式を判別
	if ext == ".xlsx" {
		csvFilePath, err = convertExcelToCSV(filePath)
		if err != nil {
			return "", fmt.Errorf("Excel to CSV conversion failed: %v", err)
		}
		defer os.Remove(csvFilePath) // 処理終了後に削除
	} else if ext == ".csv" {
		csvFilePath = filePath
	} else {
		return "", errors.New("Unsupported file format. Only .xlsx and .csv are allowed")
	}

	// データベースを初期化
	if err := clearTables(); err != nil {
		return "", fmt.Errorf("Failed to clear tables: %v", err)
	}

	// CSVを処理してDBに保存
	if err := processAndSaveCSV(csvFilePath); err != nil {
		return "", fmt.Errorf("Failed to process and save CSV data: %v", err)
	}

	return "Checklist items successfully uploaded and saved.", nil
}

// 前処理としてデータベースのテーブルをクリア
func clearTables() error {
	if err := db.Exec("SET session_replication_role = 'replica';").Error; err != nil {
		return fmt.Errorf("Failed to disable foreign key constraints: %v", err)
	}

	tables := []string{
		"checklist_item_security_standards",
		"checklist_item_standard_sections",
		"checklist_items",
		"mapping_rules",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return fmt.Errorf("Failed to clear table %s: %v", table, err)
		}
	}

	if err := db.Exec("SET session_replication_role = 'origin';").Error; err != nil {
		return fmt.Errorf("Failed to re-enable foreign key constraints: %v", err)
	}

	return nil
}


// ExcelファイルをCSVに変換
func convertExcelToCSV(excelFilePath string) (string, error) {
	f, err := excelize.OpenFile(excelFilePath)
	if err != nil {
		return "", fmt.Errorf("Failed to open Excel file: %v", err)
	}
	defer f.Close()

	sheetNames := f.GetSheetList()
	if len(sheetNames) == 0 {
		return "", errors.New("No valid sheets found in Excel file")
	}

	sheetName := sheetNames[0]
	if sheetName == "" {
		return "", errors.New("Sheet name is empty")
	}

	csvFilePath := "temp.csv"
	csvFile, err := os.Create(csvFilePath)
	if err != nil {
		return "", fmt.Errorf("Failed to create CSV file: %v", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return "", fmt.Errorf("Failed to read rows from sheet '%s': %v", sheetName, err)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("Sheet '%s' has no data", sheetName)
	}

	// 最大列数を計算して行の長さを揃える
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	for _, row := range rows {
		for len(row) < maxCols {
			row = append(row, "")
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("Error writing to CSV: %v", err)
		}
	}
	writer.Flush()

	return csvFilePath, nil
}

// CSVを処理してDBに保存
func processAndSaveCSV(csvFilePath string) error {
	file, err := os.Open(csvFilePath)
	if err != nil {
		return fmt.Errorf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	data, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("Failed to read CSV data: %v", err)
	}

	if len(data) < 2 {
		return errors.New("CSV file has no valid data")
	}

	previousRow := make([]string, len(data[0])) // 前回の行を保持
	groupedData := make(map[int][][]string)    // checklist_item_id ごとにデータをグループ化

	// データをグループ化
	for i, row := range data {
		if i < 2 {
			continue // ヘッダー行をスキップ
		}

		// 欠損値を直前の行データで埋める
		for j := 0; j < len(row); j++ {
			if row[j] == "" {
				row[j] = previousRow[j]
			}
		}
		copy(previousRow, row)

		checklistItemID, err := strconv.Atoi(row[0]) // A列 (checklist_item_id)
		if err != nil {
			return fmt.Errorf("Invalid checklist_item_id in row %d: %v", i+1, err)
		}
		groupedData[checklistItemID] = append(groupedData[checklistItemID], row)
	}

	// グループ化されたデータを処理
	for _, rows := range groupedData {
		// ChecklistItem の保存または取得
		var item models.ChecklistItem
		if err := db.FirstOrCreate(&item, models.ChecklistItem{
			Phase:       rows[0][1],
			Category:    rows[0][2],
			Subcategory: rows[0][3],
			Description: rows[0][4],
			Severity:    rows[0][5],
		}).Error; err != nil {
			return fmt.Errorf("Failed to save or retrieve checklist item: %v", err)
		}

		// MappingRules の保存
		for _, row := range rows {
			if row[6] != "" && row[7] != "" {
				var mappingRule models.MappingRule
				if err := db.FirstOrCreate(&mappingRule, models.MappingRule{
					ToolName:      row[6],
					ExpectedValue: row[7],
					ChecklistItemID: item.ID,
				}).Error; err != nil {
					return fmt.Errorf("Failed to save or retrieve mapping rule: %v", err)
				}
			}
		}

		// SecurityStandards の保存
		for _, row := range rows {
			if row[8] != "" && row[9] != "" && row[10] != "" {
				var standard models.SecurityStandard
				if err := db.FirstOrCreate(&standard, models.SecurityStandard{
					Name:    row[8],
					Version: row[9],
				}).Error; err != nil {
					return fmt.Errorf("Failed to save or retrieve security standard: %v", err)
				}

				// ChecklistItemSecurityStandard の保存
				var checklistStandard models.ChecklistItemSecurityStandard
				if err := db.FirstOrCreate(&checklistStandard, models.ChecklistItemSecurityStandard{
					ChecklistItemID:    item.ID,
					SecurityStandardID: standard.ID,
				}).Error; err != nil {
					return fmt.Errorf("Failed to save or retrieve checklist item security standard: %v", err)
				}

				// ChecklistItemStandardSection の保存
				if err := db.FirstOrCreate(&models.ChecklistItemStandardSection{}, models.ChecklistItemStandardSection{
					ChecklistItemStandardID: checklistStandard.ID,
					Section:                 row[10],
				}).Error; err != nil {
					return fmt.Errorf("Failed to save or retrieve checklist item standard section: %v", err)
				}
			}
		}
	}

	return nil
}