package db

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSeedFloorDemoTopicsIdempotent(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(&Agent{}, &FloorQuestion{}, &FloorPosition{}, &FloorDigestEntry{}); err != nil {
		t.Fatal(err)
	}
	if err := SeedFloorDemoTopics(gdb); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := gdb.Model(&FloorQuestion{}).Where("id IN ?", []string{"Q.01", "Q.02", "Q.03", "Q.04"}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Fatalf("questions: %d", n)
	}
	if err := gdb.Model(&FloorPosition{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 6 {
		t.Fatalf("positions: %d", n)
	}
	if err := SeedFloorDemoTopics(gdb); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Model(&FloorPosition{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 6 {
		t.Fatalf("second seed should not duplicate positions, got %d", n)
	}
}
