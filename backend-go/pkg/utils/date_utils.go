package utils

import "time"

// lastDayOfMonth returns the number of days in the given month/year.
// Uses time.Date overflow trick: day 0 of next month = last day of this month.
func lastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

// clampToDayOfMonth clamps a day to the actual last day of the given month,
// e.g. payday=31 in February → 28 or 29.
func clampToDayOfMonth(year int, month time.Month, day int, loc *time.Location) int {
	last := lastDayOfMonth(year, month, loc)
	if day > last {
		return last
	}
	return day
}

// GetBillingCycle returns the start and end time of the current billing cycle
// based on the current time and the user's payday (1-31).
//
// If payday < 1 or > 31, it defaults to 1 (standard calendar month).
// If payday > last day of the target month, it is clamped to the last day
// (e.g., payday=31 in February → Feb 28/29; in April → Apr 30).
//
// Logic:
//   - Effective payday for the current month is clamped first.
//   - If today >= effective payday → cycle starts on payday of the current month.
//   - If today < effective payday  → cycle starts on payday of the previous month.
//
// End = just before the next cycle's start (23:59:59 of the last day).
func GetBillingCycle(now time.Time, payday int) (time.Time, time.Time) {
	if payday < 1 || payday > 31 {
		payday = 1
	}

	year, month, day := now.Year(), now.Month(), now.Day()
	loc := now.Location()

	// Effective payday for the current calendar month determines which cycle we're in
	effectiveNow := clampToDayOfMonth(year, month, payday, loc)

	var startYear int
	var startMonth time.Month

	if day >= effectiveNow {
		// Cycle started this month
		startYear, startMonth = year, month
	} else {
		// Cycle started last month
		startMonth = month - 1
		startYear = year
		if startMonth < 1 {
			startMonth = 12
			startYear--
		}
	}

	// Clamp payday to actual start month length
	actualStartDay := clampToDayOfMonth(startYear, startMonth, payday, loc)
	start := time.Date(startYear, startMonth, actualStartDay, 0, 0, 0, 0, loc)

	// Compute next cycle's start, then subtract 1 second for end
	nextMonth := startMonth + 1
	nextYear := startYear
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}
	nextStartDay := clampToDayOfMonth(nextYear, nextMonth, payday, loc)
	nextStart := time.Date(nextYear, nextMonth, nextStartDay, 0, 0, 0, 0, loc)
	end := nextStart.Add(-time.Second)

	return start, end
}
