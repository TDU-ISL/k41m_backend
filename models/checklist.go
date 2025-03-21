package models

// ChecklistItem はセキュリティチェックリストの項目を格納するモデルです
type ChecklistItem struct {
	ID                uint   `gorm:"primaryKey"`
	Phase             string // フェーズ
	Category          string // 大分類
	Subcategory       string // 中分類
	Description       string // 小分類の説明
	Severity          string // リスクの重大度
	SecurityStandards []*SecurityStandard `gorm:"many2many:checklist_item_security_standards"`
}

// SecurityStandard はセキュリティスタンダードの情報を格納するモデルです
type SecurityStandard struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string // セキュリティスタンダード名
	Version     string // バージョン
	Description string // 説明
	Section     string // スタンダードのセクション
}

// MappingRule はツールのフィールドとチェックリスト項目のマッピング情報を格納するモデルです
type MappingRule struct {
	ID              uint   `gorm:"primaryKey"`
	ToolName        string
	ExpectedValue   string
	ChecklistItemID uint
}

// ChecklistItemSecurityStandard はChecklistItemとSecurityStandardの関連情報を格納するモデルです
type ChecklistItemSecurityStandard struct {
	ID                 uint                           `gorm:"primaryKey"`
	ChecklistItemID    uint                           `gorm:"not null"` // ChecklistItemとの紐付け
	SecurityStandardID uint                           `gorm:"not null"` // SecurityStandardとの紐付け
	SecurityStandard   SecurityStandard               `gorm:"foreignKey:SecurityStandardID"`
	Sections           []ChecklistItemStandardSection `gorm:"foreignKey:ChecklistItemStandardID"`
}

// ChecklistItemStandardSection はSecurityStandardのセクション情報を格納するモデルです
type ChecklistItemStandardSection struct {
	ID                        uint   `gorm:"primaryKey"`
	ChecklistItemStandardID   uint   `gorm:"not null"` // ChecklistItemSecurityStandardとの紐付け
	Section                   string `gorm:"not null"` // セクション情報
}