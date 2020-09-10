package date

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Format string

const (
	RFC2822  Format = "rfc2822"
	ISO8601  Format = "iso8601"
	Relative Format = "relative"
)

var now = time.Now

func SetNow(f func() time.Time) {
	if f == nil {
		panic("nil function")
	}
	now = f
}

const iso8601Layout = "2006-01-02 15:04:05 Z0700"

func FormatWithType(t time.Time, f Format) string {
	switch f {
	case RFC2822:
		return FormatRFC2822(t)
	case ISO8601:
		return FormatISO8601(t)
	case Relative:
		return FormatRelative(t)
	default:
		return "format " + string(f) + " not supported"
	}
}

// Format timestamps in a ISO 8601-like format (same as in Git command)
func FormatISO8601(t time.Time) string {
	return t.Format(iso8601Layout)
}

// Format timestamps in RFC 2822 format, often found in email messages (same as in Git command)
func FormatRFC2822(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

func FormatRelative(t time.Time) string {
	var (
		timePassed = time.Since(t)
		seconds    = int(timePassed.Seconds())
		minutes    = int(timePassed.Minutes())
		hours      = int(timePassed.Hours() + 0.5)
	)
	switch {
	case seconds < 1:
		return "Less than a second ago"
	case seconds == 1:
		return "1 second ago"
	case seconds < 60:
		return fmt.Sprintf("%d seconds ago", seconds)
	case minutes == 1:
		return "About a minute ago"
	case minutes < 60:
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours == 1:
		return "About an hour ago"
	case hours < 48:
		return fmt.Sprintf("%d hours ago", hours)
	case hours < 24*7*2:
		return fmt.Sprintf("%d days ago", hours/24)
	case hours < 24*30*2:
		return fmt.Sprintf("%d weeks ago", hours/24/7)
	case hours < 24*365*2:
		return fmt.Sprintf("%d months ago", hours/24/30)
	default:
		return fmt.Sprintf("%d years ago", int(timePassed.Hours())/24/365)
	}
}

func ParseAbsolute(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(time.RFC1123Z, value)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(iso8601Layout, value)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-01-02", value, time.Local)
	if err == nil {
		return t, nil
	}
	return time.Parse(time.UnixDate, value)
}

func Parse(value string) (time.Time, error) {
	t, err := ParseAbsolute(value)
	if err == nil {
		return t, nil
	}
	var amount time.Duration
	var unit string
	_, err = fmt.Sscanf(value, "%d %s ago", &amount, &unit)
	if err == nil {
		unit = strings.ToLower(unit)
		switch unit {
		case "seconds", "second":
			return now().Add(-time.Second * amount), nil
		case "minutes", "minute":
			return now().Add(-time.Minute * amount), nil
		case "hours", "hour":
			return now().Add(-time.Hour * amount), nil
		case "days", "day":
			return now().Add(-time.Hour * 24 * amount), nil
		case "weeks", "week":
			return now().Add(-time.Hour * 24 * 7 * amount), nil
		case "months", "month":
			return now().Add(-time.Hour * 30 * 24 * 7 * amount), nil
		case "years", "year":
			return now().Add(-time.Hour * 365 * 30 * 24 * 7 * amount), nil
		}
	}
	if value == "now" {
		return now(), nil
	}
	if value == "today" {
		return midnight(now()), nil
	}
	if value == "yesterday" {
		yesterday := now().AddDate(0, 0, -1)
		return midnight(yesterday), nil
	}
	if value == "tomorrow" {
		tomorrow := now().AddDate(0, 0, 1)
		return midnight(tomorrow), nil
	}
	return time.Time{}, errors.New("not supported date format: " + value)
}

func midnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
