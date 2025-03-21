package models

import "time"

// ScanSummary はスキャン全体の情報を格納するモデルです
type ScanSummary struct {
	ID          uint          `gorm:"primaryKey"`
	Phase       string        `gorm:"default:''"`        // フェーズ（ビルド/デプロイ/デプロイ後）
	Tool        string        `gorm:"default:''"`        // 使用したツール
	ToolScanID	string        `gorm:"default:''"`        // ツールが出力したスキャンID
	ScanTime    time.Time     `gorm:"not null"`          // スキャン実行時刻
	Status      string        `gorm:"default:''"`        // スキャンの総合ステータス
	Details     string        `gorm:"type:json;default:'{}'"` // スキャン全体(ツール出力)のjson保存
	ScanTargets []ScanTarget  `gorm:"foreignKey:ScanSummaryID"`
}

// ScanTarget はスキャン対象の情報を格納するモデルです
type ScanTarget struct {
	ID             uint   `gorm:"primaryKey"`
	ScanSummaryID  uint   `gorm:"not null"` // ScanSummaryとの紐付け
	TargetName     string // 対象名
	TargetMetadata string `gorm:"type:json"` // 対象のメタデータ
}

// ScanControl は各スキャン結果を格納するモデルです
type ScanControl struct {
	ID              uint   `gorm:"primaryKey"`
	ScanTargetID    uint   `gorm:"not null"` // ScanTargetとの紐付け
	ControlID       string // ツール項目ID
	ControlName     string // ツール項目名
	ControlStatus   string // スキャン結果
	ControlDetails  string `gorm:"type:json"` // 詳細情報
}
