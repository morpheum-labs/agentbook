package db

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// tableHasColumn reports whether the table has a column of that name in the current dialect.
func tableHasColumn(gdb *gorm.DB, table, column string) (bool, error) {
	if gdb == nil {
		return false, errors.New("nil db")
	}
	table = strings.TrimSpace(table)
	column = strings.TrimSpace(column)
	if table == "" || column == "" {
		return false, nil
	}
	var n int64
	switch gdb.Dialector.Name() {
	case "postgres":
		if err := gdb.Raw(
			`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ? AND column_name = ?`,
			table, column,
		).Count(&n).Error; err != nil {
			return false, err
		}
	case "sqlite":
		if err := gdb.Raw(`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, table, column).Count(&n).Error; err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("tableHasColumn: unsupported driver %q", gdb.Dialector.Name())
	}
	return n > 0, nil
}

// ensureUncategorized inserts a fallback [Category] row for empty legacy values.
func ensureUncategorized(gdb *gorm.DB) error {
	var c Category
	err := gdb.Where("id = ?", CategoryUncategorized).First(&c).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return gdb.Create(&Category{
		ID:          CategoryUncategorized,
		DisplayName: "Uncategorized",
		SortOrder:   9999,
		IsActive:    true,
	}).Error
}

// EnsureCategory returns an existing or new [Category] id for a non-empty label. Empty input returns ("", nil).
func EnsureCategory(gdb *gorm.DB, label string) (string, error) {
	if gdb == nil {
		return "", errors.New("nil db")
	}
	label = strings.TrimSpace(label)
	if label == "" {
		return "", nil
	}
	var c Category
	if err := gdb.Where("id = ?", label).First(&c).Error; err == nil {
		return c.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	c = Category{ID: label, DisplayName: label, SortOrder: 0, IsActive: true}
	if err := gdb.Create(&c).Error; err != nil {
		return "", err
	}
	return c.ID, nil
}

// seedCategoriesForDistinctIDs creates missing [Category] rows for every non-empty category_id in dependent tables.
func seedCategoriesForDistinctIDs(gdb *gorm.DB) error {
	seen := map[string]struct{}{}
	collect := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" {
			seen[s] = struct{}{}
		}
	}
	var fq []string
	if err := gdb.Model(&FloorQuestion{}).Where("TRIM(COALESCE(category_id, '')) <> ''").Distinct("category_id").Pluck("category_id", &fq).Error; err != nil {
		return err
	}
	for i := range fq {
		collect(fq[i])
	}
	var ft []string
	if err := gdb.Model(&FloorTopicProposal{}).Where("TRIM(COALESCE(category_id, '')) <> ''").Distinct("category_id").Pluck("category_id", &ft).Error; err != nil {
		return err
	}
	for i := range ft {
		collect(ft[i])
	}
	var capIDs []string
	if err := gdb.Model(&CapabilityService{}).Where("category_id IS NOT NULL AND TRIM(COALESCE(category_id, '')) <> ''").Distinct("category_id").Pluck("category_id", &capIDs).Error; err != nil {
		return err
	}
	for i := range capIDs {
		collect(capIDs[i])
	}
	for id := range seen {
		if _, err := EnsureCategory(gdb, id); err != nil {
			return err
		}
	}
	return nil
}

// MigrateCategoryReferences backfills [Category] and category_id from legacy "category" text columns
// when they still exist (pre-schema-change databases). Then fills missing [FloorQuestion] / [FloorTopicProposal]
// category_id with [CategoryUncategorized] if needed.
func MigrateCategoryReferences(gdb *gorm.DB) error {
	if err := ensureUncategorized(gdb); err != nil {
		return err
	}

	// --- Legacy column copy → category_id -------------------------------------------------
	for _, t := range []struct {
		table, legacyCol, idCol, nullable string
	}{
		{"floor_questions", "category", "category_id", "not_null"},
		{"floor_topic_proposals", "category", "category_id", "not_null"},
		{"capability_services", "category", "category_id", "null"},
	} {
		has, err := tableHasColumn(gdb, t.table, t.legacyCol)
		if err != nil {
			return err
		}
		if !has {
			continue
		}
		// copy legacy text into category_id where the new id column is still empty
		qt := `UPDATE "` + t.table + `" SET "` + t.idCol + `" = TRIM(` + t.legacyCol + `) WHERE (`
		if t.nullable == "null" {
			qt += t.idCol + ` IS NULL OR ` + t.idCol + ` = ''`
		} else {
			qt += t.idCol + ` IS NULL OR ` + t.idCol + ` = '' OR TRIM(` + t.idCol + `) = ''`
		}
		qt += `) AND ` + t.legacyCol + ` IS NOT NULL AND TRIM(` + t.legacyCol + `) <> ''`
		if err := gdb.Exec(qt).Error; err != nil {
			return err
		}
		if t.table == "capability_services" {
			// treat empty/whitespace as unset (nullable)
			if err := gdb.Exec(
				`UPDATE "` + t.table + `" SET "` + t.idCol + `" = NULL WHERE ` + t.legacyCol + ` IS NULL OR TRIM(COALESCE(` + t.legacyCol + `, '')) = ''`,
			).Error; err != nil {
				return err
			}
		}
	}

	// ensure category rows for every id referenced
	if err := seedCategoriesForDistinctIDs(gdb); err != nil {
		return err
	}

	// required floor fields: never null
	if err := gdb.Model(&FloorQuestion{}).
		Where("category_id IS NULL OR TRIM(COALESCE(category_id, '')) = ''").
		Update("category_id", CategoryUncategorized).Error; err != nil {
		return err
	}
	if err := gdb.Model(&FloorTopicProposal{}).
		Where("category_id IS NULL OR TRIM(COALESCE(category_id, '')) = ''").
		Update("category_id", CategoryUncategorized).Error; err != nil {
		return err
	}

	return nil
}
