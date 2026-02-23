// Package dbutil provides shared pgx type conversion helpers.
package dbutil

import (
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// NumericToFloat64 converts a pgtype.Numeric to float64.
func NumericToFloat64(n pgtype.Numeric) float64 {
	f, _ := n.Float64Value()
	return f.Float64
}

// Float64ToNumeric converts a float64 to pgtype.Numeric.
func Float64ToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	s := new(big.Float).SetFloat64(f).Text('f', -1)
	n.ScanScientific(s)
	return n
}

// TimeToDate converts a time.Time to pgtype.Date.
func TimeToDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}
