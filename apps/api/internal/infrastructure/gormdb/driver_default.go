//go:build !cgo

package gormdb

import (
	"fmt"

	"gorm.io/gorm"
)

func resolveDialector(driver, dsn string) (gorm.Dialector, error) {
	switch driver {
	case "postgres":
		return postgresDialector(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported GORM driver: %s (sqlite requires CGO build)", driver)
	}
}
