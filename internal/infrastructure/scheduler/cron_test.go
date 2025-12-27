package scheduler

import (
	"testing"
	"time"
)

func TestParseCron(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{
			name:    "daily at midnight",
			spec:    "0 0 * * *",
			wantErr: false,
		},
		{
			name:    "every 5 minutes",
			spec:    "*/5 * * * *",
			wantErr: false,
		},
		{
			name:    "business hours",
			spec:    "0 9-17 * * 1-5",
			wantErr: false,
		},
		{
			name:    "invalid - too few fields",
			spec:    "0 0 * *",
			wantErr: true,
		},
		{
			name:    "invalid - too many fields",
			spec:    "0 0 * * * *",
			wantErr: true,
		},
		{
			name:    "invalid minute",
			spec:    "60 0 * * *",
			wantErr: true,
		},
		{
			name:    "invalid hour",
			spec:    "0 24 * * *",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCron(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCron() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCronSchedule_Next(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		from     time.Time
		expected time.Time
	}{
		{
			name:     "daily at midnight",
			spec:     "0 0 * * *",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "every hour",
			spec:     "0 * * * *",
			from:     time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		},
		{
			name:     "every 15 minutes",
			spec:     "*/15 * * * *",
			from:     time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 12, 15, 0, 0, time.UTC),
		},
		{
			name:     "specific time",
			spec:     "30 14 * * *",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC),
		},
		{
			name:     "first day of month",
			spec:     "0 0 1 * *",
			from:     time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs, err := ParseCron(tt.spec)
			if err != nil {
				t.Fatalf("Failed to parse cron: %v", err)
			}

			next := cs.Next(tt.from)
			if !next.Equal(tt.expected) {
				t.Errorf("Next() = %v, want %v", next, tt.expected)
			}
		})
	}
}

func TestParseField(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		min      int
		max      int
		expected []int
		wantErr  bool
	}{
		{
			name:     "wildcard",
			field:    "*",
			min:      0,
			max:      5,
			expected: []int{0, 1, 2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "single value",
			field:    "3",
			min:      0,
			max:      10,
			expected: []int{3},
			wantErr:  false,
		},
		{
			name:     "range",
			field:    "2-5",
			min:      0,
			max:      10,
			expected: []int{2, 3, 4, 5},
			wantErr:  false,
		},
		{
			name:     "step from wildcard",
			field:    "*/2",
			min:      0,
			max:      6,
			expected: []int{0, 2, 4, 6},
			wantErr:  false,
		},
		{
			name:     "step from range",
			field:    "0-10/3",
			min:      0,
			max:      15,
			expected: []int{0, 3, 6, 9},
			wantErr:  false,
		},
		{
			name:     "comma-separated",
			field:    "1,3,5",
			min:      0,
			max:      10,
			expected: []int{1, 3, 5},
			wantErr:  false,
		},
		{
			name:     "invalid - out of range",
			field:    "20",
			min:      0,
			max:      10,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid - negative",
			field:    "-5",
			min:      0,
			max:      10,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseField(tt.field, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) != len(tt.expected) {
					t.Errorf("parseField() length = %d, want %d", len(result), len(tt.expected))
					return
				}

				for i, v := range result {
					if v != tt.expected[i] {
						t.Errorf("parseField()[%d] = %d, want %d", i, v, tt.expected[i])
					}
				}
			}
		})
	}
}

func TestCronSchedule_Matches(t *testing.T) {
	cs := &CronSchedule{
		Minutes:     []int{0, 15, 30, 45},
		Hours:       []int{9, 10, 11, 12, 13, 14, 15, 16, 17},
		DaysOfMonth: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		Months:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		DaysOfWeek:  []int{1, 2, 3, 4, 5}, // Monday to Friday
	}

	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{
			name:     "matches - Monday 9:00",
			time:     time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), // Monday
			expected: true,
		},
		{
			name:     "matches - Friday 17:15",
			time:     time.Date(2024, 1, 5, 17, 15, 0, 0, time.UTC), // Friday
			expected: true,
		},
		{
			name:     "no match - wrong minute",
			time:     time.Date(2024, 1, 1, 9, 5, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "no match - wrong hour",
			time:     time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "no match - weekend",
			time:     time.Date(2024, 1, 6, 9, 0, 0, 0, time.UTC), // Saturday
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cs.matches(tt.time)
			if result != tt.expected {
				t.Errorf("matches() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkParseCron(b *testing.B) {
	spec := "*/5 9-17 * * 1-5"
	for range b.N {
		_, _ = ParseCron(spec)
	}
}

func BenchmarkCronNext(b *testing.B) {
	cs, _ := ParseCron("*/15 * * * *")
	from := time.Now()

	b.ResetTimer()
	for range b.N {
		_ = cs.Next(from)
	}
}
