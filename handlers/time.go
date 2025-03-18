package handlers

import (
	"fmt"
	"time"
)

// TimeAgo returns a human-readable string representing the time elapsed since the given time.
func TimeAgo(createdAt time.Time) string {
	now := time.Now()
	diff := now.Sub(createdAt)

	switch {
	case diff < time.Second:
		return "just now"
	case diff < time.Minute:
		seconds := int(diff.Seconds())
		if seconds == 1 {
			return "1 second ago"
		}
		return fmt.Sprintf("%d seconds ago", seconds)
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
