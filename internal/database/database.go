package database

import (
	"gorm.io/gorm"
)

// GetDB returns the database connection
// This is a placeholder for database connection management
func GetDB() *gorm.DB {
	// This will be initialized in main.go
	return nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}