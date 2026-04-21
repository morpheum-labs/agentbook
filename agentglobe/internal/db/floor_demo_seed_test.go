package db

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSeedFloorDemoTopicsIdempotent(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file:topics_seed_test?mode=memory&cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(
		&Agent{},
		&FloorQuestion{},
		&FloorPosition{},
		&FloorDigestEntry{},
		&FloorAgentTopicStat{},
		&FloorAgentInferenceProfile{},
	); err != nil {
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
	if err := gdb.Model(&FloorDigestEntry{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Fatalf("digest rows: %d", n)
	}
	var omega Agent
	if err := gdb.Where("id = ?", "floor-demo-agent-omega").First(&omega).Error; err != nil {
		t.Fatal(err)
	}
	if omega.DisplayName == nil || *omega.DisplayName != "DeepValue" {
		t.Fatalf("omega display_name: %#v", omega.DisplayName)
	}
	if omega.FloorHandle == nil || *omega.FloorHandle != "deepvalue" {
		t.Fatalf("omega floor_handle: %#v", omega.FloorHandle)
	}
	if !omega.PlatformVerified {
		t.Fatal("omega want platform_verified")
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

func TestSeedFloorDemoAgentsWhenQuestionsAlreadyExist(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file:agents_when_q_exist_test?mode=memory&cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(&Agent{}, &FloorQuestion{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	q := FloorQuestion{
		ID: "Q.01", Title: "Existing", Category: "NBA", ResolutionCondition: "x",
		Deadline: "2026-01-01T00:00:00Z", Probability: 0.5, ProbabilityDelta: 0,
		AgentCount: 1, StakedCount: 1, Status: "open",
		ClusterBreakdownJSON: "{}", CreatedAt: now, UpdatedAt: now,
	}
	if err := gdb.Create(&q).Error; err != nil {
		t.Fatal(err)
	}
	if err := SeedFloorDemoTopics(gdb); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := gdb.Model(&Agent{}).Where("id LIKE ?", "floor-demo-agent-%").Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("topics seed skipped agents when Q present, got agent count %d", n)
	}
	if err := SeedFloorDemoAgents(gdb); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Model(&Agent{}).Where("id LIKE ?", "floor-demo-agent-%").Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 6 {
		t.Fatalf("after SeedFloorDemoAgents want 6 demo agents, got %d", n)
	}
	if err := SeedFloorDemoAgents(gdb); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Model(&Agent{}).Where("id LIKE ?", "floor-demo-agent-%").Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 6 {
		t.Fatalf("second agent seed should not duplicate, got %d", n)
	}
}

func TestSeedFloorDemoIndexIdempotent(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file:index_seed_test?mode=memory&cache=private"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(&FloorIndexPageMeta{}, &FloorIndexEntry{}); err != nil {
		t.Fatal(err)
	}
	if err := SeedFloorDemoIndex(gdb); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := gdb.Model(&FloorIndexEntry{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("index entries: %d", n)
	}
	if err := SeedFloorDemoIndex(gdb); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Model(&FloorIndexEntry{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("second seed should not duplicate index rows, got %d", n)
	}
}
