// Package periods computes calendar period boundaries for a comp plan's cadence.
package periods

import "time"

// Range is a half-open period [Start, End).
type Range struct {
	Start time.Time
	End   time.Time
}

// ForDate returns the period containing t for the given cadence
// ("monthly", "quarterly", "annual"). Boundaries are in t's location.
func ForDate(periodType string, t time.Time) Range {
	y, m, _ := t.Date()
	loc := t.Location()

	switch periodType {
	case "annual":
		start := time.Date(y, time.January, 1, 0, 0, 0, 0, loc)
		return Range{Start: start, End: start.AddDate(1, 0, 0)}
	case "quarterly":
		// Quarter starting month: 1, 4, 7, or 10.
		qStartMonth := time.Month((int(m)-1)/3*3 + 1)
		start := time.Date(y, qStartMonth, 1, 0, 0, 0, 0, loc)
		return Range{Start: start, End: start.AddDate(0, 3, 0)}
	default: // monthly
		start := time.Date(y, m, 1, 0, 0, 0, 0, loc)
		return Range{Start: start, End: start.AddDate(0, 1, 0)}
	}
}
