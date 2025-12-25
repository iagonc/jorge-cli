package testutil

import (
	"testing"

	"github.com/iagonc/jorge-cli/internal/schemas"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	err = db.AutoMigrate(&schemas.Resource{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// SeedTestData populates the database with test data
func SeedTestData(t *testing.T, db *gorm.DB) []*schemas.Resource {
	t.Helper()

	resources := []*schemas.Resource{
		{Name: "test-resource-1", Dns: "test1.example.com"},
		{Name: "test-resource-2", Dns: "test2.example.com"},
		{Name: "test-resource-3", Dns: "test3.example.com"},
	}

	for _, r := range resources {
		if err := db.Create(r).Error; err != nil {
			t.Fatalf("Failed to seed test data: %v", err)
		}
	}

	return resources
}

// CleanupTestData removes all test data from the database
func CleanupTestData(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Exec("DELETE FROM resources")
}
