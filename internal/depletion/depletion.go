// Package depletion provides pure functions for prescription depletion calculations.
package depletion

import (
	"math"
	"time"
)

// Status constants for prescription depletion classification.
const (
	StatusOk          = "ok"
	StatusApproaching = "approaching"
	StatusDepleted    = "depleted"
)

// EstimatedDate returns the date when a box is expected to run out.
// Formula: boxStartDate + floor(unitsPerBox / dailyConsumption) days.
func EstimatedDate(unitsPerBox int, dailyConsumption float64, boxStartDate time.Time) time.Time {
	days := math.Floor(float64(unitsPerBox) / dailyConsumption)
	return boxStartDate.AddDate(0, 0, int(days))
}

// DaysRemaining returns the number of days until depletionDate relative to now.
// Negative values mean the prescription is past depletion.
func DaysRemaining(depletionDate, now time.Time) int {
	now = now.Truncate(24 * time.Hour)
	depletionDate = depletionDate.Truncate(24 * time.Hour)
	return int(depletionDate.Sub(now).Hours() / 24)
}

// Status classifies based on days remaining:
// "depleted" (<=0), "approaching" (<=7), "ok" (>7).
func Status(daysRemaining int) string {
	switch {
	case daysRemaining <= 0:
		return StatusDepleted
	case daysRemaining <= 7:
		return StatusApproaching
	default:
		return StatusOk
	}
}
