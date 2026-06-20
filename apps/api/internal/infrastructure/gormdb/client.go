package gormdb

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

func NewClient(driver, dsn string) (*gorm.DB, error) {
	dialector, err := resolveDialector(driver, dsn)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm connect (%s): %w", driver, err)
	}

	if driver == "postgres" {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("gorm sql.DB: %w", err)
		}
		if err := sqlDB.Ping(); err != nil {
			return nil, fmt.Errorf("gorm ping (%s): %w", driver, err)
		}
	}

	return db, nil
}

func postgresDialector(dsn string) gorm.Dialector {
	return postgres.Open(dsn)
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.ContentType{},
		&entity.Document{},
		&entity.MediaAsset{},
		&entity.RoleEntity{},
		&entity.Invite{},
		&entity.AccessToken{},
	)
}
