package reminders

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Mode represents how a reminder should be interpreted.
type Mode string

const (
	// ModeNone indicates that no reminder is configured.
	ModeNone Mode = ""
	// ModeTimeOfDay schedules the reminder at the next occurrence of the provided HH:MM time.
	ModeTimeOfDay Mode = "time_of_day"
	// ModeDuration schedules the reminder relative to the saved time using a duration such as 30m or 2h.
	ModeDuration Mode = "duration"
)

// Preference stores the reminder configuration for a bookmark.
type Preference struct {
	Mode             Mode  `json:"mode"`
	Hour             int   `json:"hour,omitempty"`
	Minute           int   `json:"minute,omitempty"`
	DurationSeconds  int64 `json:"durationSeconds,omitempty"`
	RemoveOnComplete bool  `json:"removeOnComplete"`
}

// Schedule represents the next reminder instance together with human friendly text.
type Schedule struct {
	Time        time.Time
	Description string
}

// Parse converts raw user input into a reminder preference. Returning nil indicates the reminder should be cleared.
func Parse(raw string) (*Preference, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	lowered := strings.ToLower(trimmed)
	switch lowered {
	case "none", "off", "clear", "0":
		return nil, nil
	}

	if strings.Contains(trimmed, ":") {
		parts := strings.Split(trimmed, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid time. Use HH:MM such as 08:30")
		}

		hour, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("unable to read hour value: %w", err)
		}
		minute, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("unable to read minute value: %w", err)
		}

		if hour < 0 || hour > 23 {
			return nil, errors.New("hour must be between 0 and 23")
		}
		if minute < 0 || minute > 59 {
			return nil, errors.New("minute must be between 0 and 59")
		}

		return &Preference{
			Mode:   ModeTimeOfDay,
			Hour:   hour,
			Minute: minute,
		}, nil
	}

	cleaned := lowered
	cleaned = strings.TrimPrefix(cleaned, "in ")
	cleaned = strings.TrimPrefix(cleaned, "after ")

	duration, err := time.ParseDuration(cleaned)
	if err != nil {
		return nil, errors.New("use durations like `30m` or `2h45m`")
	}
	if duration <= 0 {
		return nil, errors.New("reminders must be at least 1m in the future")
	}

	return &Preference{
		Mode:            ModeDuration,
		DurationSeconds: int64(duration / time.Second),
	}, nil
}

// Next determines the next reminder time and a textual description for the embed display.
func Next(pref *Preference, now time.Time) (*Schedule, error) {
	if pref == nil {
		return nil, nil
	}

	switch pref.Mode {
	case ModeTimeOfDay:
		target := time.Date(now.Year(), now.Month(), now.Day(), pref.Hour, pref.Minute, 0, 0, now.Location())
		if !target.After(now) {
			target = target.Add(24 * time.Hour)
		}
		desc := fmt.Sprintf("Next alert at %s (daily)", target.Format("2006-01-02 15:04"))
		return &Schedule{Time: target, Description: desc}, nil
	case ModeDuration:
		duration := time.Duration(pref.DurationSeconds) * time.Second
		if duration <= 0 {
			return nil, errors.New("the reminder configuration is invalid. Please set it again")
		}
		target := now.Add(duration)
		desc := fmt.Sprintf("Reminder in %s (%s)", formatDuration(duration), target.Format("2006-01-02 15:04"))
		return &Schedule{Time: target, Description: desc}, nil
	case ModeNone:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown reminder mode")
	}
}

// Describe returns a concise textual representation of the reminder configuration for listings.
func Describe(pref *Preference) string {
	if pref == nil {
		return "No reminder"
	}

	switch pref.Mode {
	case ModeTimeOfDay:
		return fmt.Sprintf("Every day at %02d:%02d", pref.Hour, pref.Minute)
	case ModeDuration:
		duration := time.Duration(pref.DurationSeconds) * time.Second
		return fmt.Sprintf("%s after saving", formatDuration(duration))
	default:
		return "No reminder"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		minutes := int((d + time.Second/2) / time.Minute)
		if minutes <= 0 {
			minutes = 1
		}
		return fmt.Sprintf("%d min", minutes)
	}

	hours := int(d / time.Hour)
	remainder := d % time.Hour
	minutes := int(remainder / time.Minute)

	var parts []string
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if len(parts) == 0 {
		seconds := int((remainder % time.Minute) / time.Second)
		if seconds > 0 {
			parts = append(parts, fmt.Sprintf("%ds", seconds))
		}
	}

	if len(parts) == 0 {
		return "a few minutes"
	}

	return strings.Join(parts, " ")
}
