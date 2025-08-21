package gorm

import (
	"fmt"
	"service-mqtt/internal/core/devices"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// New creates a new GORM database instance and runs migrations.
func New(dsn string, lg zerolog.Logger) (*gorm.DB, error) {
	gormLogger := gormlog.New(
		&lg,
		gormlog.Config{
			SlowThreshold: 0, // log all queries
			LogLevel:      gormlog.Info,
			Colorful:      false,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	if err := db.AutoMigrate(&devices.Device{}); err != nil {
		return nil, fmt.Errorf("gorm migrate: %w", err)
	}
	lg.Info().Msg("database migration successful")

	return db, nil
}
