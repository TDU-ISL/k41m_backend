package main

import (
	"log"
	"scan_backend/internal/handlers"
	"scan_backend/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// データベースの初期化
	if err := services.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Ginのルーター作成
	r := gin.Default()

	// APIルートの定義
	api := r.Group("/api")
	{
		api.POST("/scan/start", handlers.StartScan)       // スキャン実行
		api.GET("/scan/history", handlers.GetScanHistory)      // スキャン履歴取得
		api.GET("/scan/:scan_id/details", handlers.GetScanDetails) // スキャン詳細取得
		api.POST("/monitor/notify", handlers.ReceiveMonitorNotification) // モニタリング通知受信
		api.GET("/monitor/details", handlers.GetMonitorDetails) // モニタリング通知取得
		api.GET("/checklist_items", handlers.GetChecklistItems) // チェックリストアイテム取得
		api.POST("/checklist_items/upload", handlers.UploadChecklist) // チェックリストアップロード
	}
	// サーバーの起動
	log.Println("Server running on port 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}