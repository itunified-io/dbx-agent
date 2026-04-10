package oracle

import (
	"context"
	"database/sql"
)

// OracleQuerier abstracts *sql.DB for Oracle queries.
type OracleQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// LicenseChecker verifies if an Oracle feature is licensed for this target.
type LicenseChecker interface {
	IsLicensed(option string) bool
}

// noopLicense always returns true (for targets where license checking is disabled).
type noopLicense struct{}

func (n *noopLicense) IsLicensed(string) bool { return true }

// NewNoopLicense returns a license checker that approves everything.
func NewNoopLicense() LicenseChecker { return &noopLicense{} }
