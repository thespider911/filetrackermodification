package helpers

import (
	"fmt"
	"testing"
	"time"
)

func TestToHumanReadableTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1633027200", "Thursday 30 September, 2021 09:40 PM"},
		{"1640995200", "Saturday 01 January, 2022 03:00 AM"},
		{"1672531200", "Sunday 01 January, 2023 03:00 AM"},
	}

	for _, test := range tests {
		result := ToHumanReadableTime(test.input)
		if result != test.expected {
			t.Errorf("ToHumanReadableTime(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestToHumanReadableFileSize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"500", "500 B"},
		{"1500", "1.5 KB"},
		{"1500000", "1.4 MB"},
		{"1500000000", "1.4 GB"},
	}

	for _, test := range tests {
		result := ToHumanReadableFileSize(test.input)
		if result != test.expected {
			t.Errorf("ToHumanReadableFileSize(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestToHumanReadableTimeDiff(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    string
		expected string
	}{
		{fmt.Sprintf("%d", now.Add(-30*time.Second).Unix()), "30 seconds ago"},
		{fmt.Sprintf("%d", now.Add(-45*time.Minute).Unix()), "45 minutes ago"},
		{fmt.Sprintf("%d", now.Add(-3*time.Hour).Unix()), "3 hours 0 minutes ago"},
		{fmt.Sprintf("%d", now.Add(-49*time.Hour).Unix()), "2 days 1 hours ago"},
	}

	for _, test := range tests {
		result := ToHumanReadableTimeDiff(test.input)
		if result != test.expected {
			t.Errorf("ToHumanReadableTimeDiff(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}
