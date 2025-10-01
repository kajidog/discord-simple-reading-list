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
	case "none", "off", "clear", "なし", "0":
		return nil, nil
	}

	if strings.Contains(trimmed, ":") {
		parts := strings.Split(trimmed, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("無効な時刻です。`08:30` の形式で入力してください。")
		}

		hour, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("時刻の時部分を読み取れませんでした: %w", err)
		}
		minute, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("時刻の分部分を読み取れませんでした: %w", err)
		}

		if hour < 0 || hour > 23 {
			return nil, errors.New("時刻の時は 0〜23 の範囲で指定してください")
		}
		if minute < 0 || minute > 59 {
			return nil, errors.New("時刻の分は 0〜59 の範囲で指定してください")
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
		return nil, errors.New("時間は `30m` や `2h45m` のような形式で入力してください")
	}
	if duration <= 0 {
		return nil, errors.New("リマインドまでの時間は 1分以上にしてください")
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
		desc := fmt.Sprintf("%s に通知（毎日）", target.Format("2006-01-02 15:04"))
		return &Schedule{Time: target, Description: desc}, nil
	case ModeDuration:
		duration := time.Duration(pref.DurationSeconds) * time.Second
		if duration <= 0 {
			return nil, errors.New("リマインドの設定が無効です。再設定してください。")
		}
		target := now.Add(duration)
		desc := fmt.Sprintf("%s後（%s）", formatDuration(duration), target.Format("2006-01-02 15:04"))
		return &Schedule{Time: target, Description: desc}, nil
	case ModeNone:
		return nil, nil
	default:
		return nil, fmt.Errorf("不明なリマインドモードです")
	}
}

// Describe returns a concise textual representation of the reminder configuration for listings.
func Describe(pref *Preference) string {
	if pref == nil {
		return "リマインドなし"
	}

	switch pref.Mode {
	case ModeTimeOfDay:
		return fmt.Sprintf("毎日 %02d:%02d", pref.Hour, pref.Minute)
	case ModeDuration:
		duration := time.Duration(pref.DurationSeconds) * time.Second
		return fmt.Sprintf("保存から %s後", formatDuration(duration))
	default:
		return "リマインドなし"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		minutes := int((d + time.Second/2) / time.Minute)
		if minutes <= 0 {
			minutes = 1
		}
		return fmt.Sprintf("%d分", minutes)
	}

	hours := int(d / time.Hour)
	remainder := d % time.Hour
	minutes := int(remainder / time.Minute)

	var parts []string
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d時間", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d分", minutes))
	}
	if len(parts) == 0 {
		seconds := int((remainder % time.Minute) / time.Second)
		if seconds > 0 {
			parts = append(parts, fmt.Sprintf("%d秒", seconds))
		}
	}

	if len(parts) == 0 {
		return "数分"
	}

	return strings.Join(parts, "")
}
