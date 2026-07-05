package db

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"quota/internal/config"
	"quota/internal/models"
)

// Connect opens the database. It uses Postgres when DATABASE_URL is set,
// otherwise a local SQLite file for zero-setup local development.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	if cfg.DatabaseURL != "" {
		log.Println("connecting to Postgres")
		dialector = postgres.Open(cfg.DatabaseURL)
	} else {
		log.Printf("DATABASE_URL not set; using local SQLite at %s", cfg.SQLitePath)
		dialector = sqlite.Open(cfg.SQLitePath)
	}

	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, err
	}
	return gdb, nil
}

// Migrate runs auto-migration for all models.
func Migrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(models.AllModels()...)
}
