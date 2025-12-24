package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronSchedule represents a parsed cron expression.
type CronSchedule struct {
	Minutes     []int // 0-59
	Hours       []int // 0-23
	DaysOfMonth []int // 1-31
	Months      []int // 1-12
	DaysOfWeek  []int // 0-6 (Sunday = 0)
}

// ParseCron parses a cron expression in the format "minute hour day month weekday"
// Examples:
//   - "0 0 * * *" - Daily at midnight
//   - "*/5 * * * *" - Every 5 minutes
//   - "0 9-17 * * 1-5" - Every hour from 9am to 5pm, Monday to Friday
//   - "0 0 1 * *" - First day of every month at midnight
func ParseCron(spec string) (*CronSchedule, error) {
	fields := strings.Fields(spec)
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid cron expression: expected 5 fields, got %d", len(fields))
	}

	cs := &CronSchedule{}
	var err error

	// Parse minutes (0-59)
	cs.Minutes, err = parseField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("invalid minutes: %w", err)
	}

	// Parse hours (0-23)
	cs.Hours, err = parseField(fields[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("invalid hours: %w", err)
	}

	// Parse days of month (1-31)
	cs.DaysOfMonth, err = parseField(fields[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("invalid days of month: %w", err)
	}

	// Parse months (1-12)
	cs.Months, err = parseField(fields[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("invalid months: %w", err)
	}

	// Parse days of week (0-6, Sunday = 0)
	cs.DaysOfWeek, err = parseField(fields[4], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("invalid days of week: %w", err)
	}

	return cs, nil
}

// parseField parses a single cron field.
func parseField(field string, min, max int) ([]int, error) {
	// Handle wildcard
	if field == "*" {
		values := make([]int, max-min+1)
		for i := range values {
			values[i] = min + i
		}
		return values, nil
	}

	// Handle step values (*/5, 0-59/5, etc.)
	if strings.Contains(field, "/") {
		parts := strings.Split(field, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid step expression: %s", field)
		}

		step, err := strconv.Atoi(parts[1])
		if err != nil || step <= 0 {
			return nil, fmt.Errorf("invalid step value: %s", parts[1])
		}

		var base []int
		if parts[0] == "*" {
			base = make([]int, max-min+1)
			for i := range base {
				base[i] = min + i
			}
		} else {
			base, err = parseField(parts[0], min, max)
			if err != nil {
				return nil, err
			}
		}

		var values []int
		for _, v := range base {
			if (v-min)%step == 0 {
				values = append(values, v)
			}
		}
		return values, nil
	}

	// Handle ranges (0-5, 9-17, etc.)
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range: %s", field)
		}

		start, err := strconv.Atoi(parts[0])
		if err != nil || start < min || start > max {
			return nil, fmt.Errorf("invalid range start: %s", parts[0])
		}

		end, err := strconv.Atoi(parts[1])
		if err != nil || end < min || end > max || end < start {
			return nil, fmt.Errorf("invalid range end: %s", parts[1])
		}

		values := make([]int, end-start+1)
		for i := range values {
			values[i] = start + i
		}
		return values, nil
	}

	// Handle comma-separated lists (0,15,30,45)
	if strings.Contains(field, ",") {
		parts := strings.Split(field, ",")
		values := make([]int, 0, len(parts))
		for _, part := range parts {
			vals, err := parseField(part, min, max)
			if err != nil {
				return nil, err
			}
			values = append(values, vals...)
		}
		return values, nil
	}

	// Handle single value
	value, err := strconv.Atoi(field)
	if err != nil || value < min || value > max {
		return nil, fmt.Errorf("invalid value: %s (must be between %d and %d)", field, min, max)
	}

	return []int{value}, nil
}

// Next calculates the next execution time after the given time.
func (cs *CronSchedule) Next(after time.Time) time.Time {
	// Start from the next minute
	t := after.Truncate(time.Minute).Add(time.Minute)

	// Try up to 4 years in the future (to handle Feb 29 and other edge cases)
	maxIterations := 4 * 365 * 24 * 60
	for range maxIterations {
		if cs.matches(t) {
			return t
		}
		t = t.Add(time.Minute)
	}

	// If we couldn't find a match, return far future
	return after.Add(365 * 24 * time.Hour)
}

// matches checks if the given time matches the cron schedule.
func (cs *CronSchedule) matches(t time.Time) bool {
	minute := t.Minute()
	hour := t.Hour()
	day := t.Day()
	month := int(t.Month())
	weekday := int(t.Weekday())

	return contains(cs.Minutes, minute) &&
		contains(cs.Hours, hour) &&
		contains(cs.DaysOfMonth, day) &&
		contains(cs.Months, month) &&
		contains(cs.DaysOfWeek, weekday)
}

// contains checks if a slice contains a value.
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
