package api

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

func isUniqueViolation(err error) bool {
	var pg *pgconn.PgError
	if ok := errors.As(err, &pg); ok {
		return pg.Code == "23505"
	}
	// GORM may wrap; fallback string match
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
