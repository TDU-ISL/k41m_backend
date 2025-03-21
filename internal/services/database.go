package services

import (
	"log"
	"os"
	"k41m_backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDatabase() error {
	var err error

	// データベース接続文字列
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=postgresql user=user password=password dbname=mydb port=5432 sslmode=disable"
	}

	// データベース接続
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(
		&models.ScanSummary{},
		&models.ScanTarget{},
		&models.ScanControl{},
		&models.ChecklistItem{},
		&models.MappingRule{},
		&models.SecurityStandard{},
		&models.ChecklistItemSecurityStandard{},
		&models.ChecklistItemStandardSection{},
		&models.MonitorNotification{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
		return err
	}

	log.Println("Database migrated successfully")
	return nil
}

func GetDB() *gorm.DB {
	return db
}
