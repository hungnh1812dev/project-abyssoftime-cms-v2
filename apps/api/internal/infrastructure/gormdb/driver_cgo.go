//go:build cgo

package gormdb

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func resolveDialector(driver, dsn string) (gorm.Dialector, error) {
	switch driver {
	case "postgres":
		return postgresDialector(dsn), nil
	case "sqlite":
		return sqlite.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported GORM driver: %s", driver)
	}
}
