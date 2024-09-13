package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// --------------- HELPERS --------------- //
func (app *application) toHumanReadableTime(val string) string {
	unix, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Error converting time: %v\n", err)
	}

	t := time.Unix(int64(unix), 0)
	return t.Format("Monday 02 January, 2006 03:04 PM")
}

func (app *application) toHumanReadableFileSize(val string) string {
	size, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Error converting file size: %v\n", err)
	}

	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
}

// toHumanReadableTimeDiff - calculate time difference
func (app *application) toHumanReadableTimeDiff(val string) string {
	unixTime, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Error converting time: %v\n", err)
	}

	now := time.Now()
	t := time.Unix(int64(unixTime), 0)

	//get diff
	duration := now.Sub(t)
	seconds := int(duration.Seconds())
	minutes := int(duration.Minutes())
	hours := int(duration.Hours())
	days := int(hours / 24)

	switch {
	case seconds < 60:
		return fmt.Sprintf("%d seconds ago", seconds)
	case minutes < 60:
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours < 24:
		return fmt.Sprintf("%d hours %d minutes ago", hours, minutes%60)
	default:
		return fmt.Sprintf("%d days %d hours ago", days, hours%24)
	}
}
