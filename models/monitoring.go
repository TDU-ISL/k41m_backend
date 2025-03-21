package models

import "time"

// MonitorNotification は運用・監視フェーズの通知データを保存するモデル
type MonitorNotification struct {
	ID         uint      `gorm:"primaryKey"`
	Tool       string    `gorm:"default:''"`         // ツール名 (例: Falco)
	RuleName   string    `gorm:"default:''"`         // ルール名
	ReceivedAt time.Time                             // 受信時刻（初期値は自動設定）
	Details    string    `gorm:"type:json;default:'{}'"` // 通知データのJSON (デフォルトは空のJSON)
}
