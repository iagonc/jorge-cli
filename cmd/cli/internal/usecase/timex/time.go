package timex

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// TimeUsecase handles time operations
type TimeUsecase struct {
	logger *zap.Logger
}

// NewTimeUsecase creates a new TimeUsecase
func NewTimeUsecase(logger *zap.Logger) *TimeUsecase {
	return &TimeUsecase{logger: logger}
}

// Now returns the current time in various formats
func (u *TimeUsecase) Now(timezones []string) *models.TimeNow {
	now := time.Now()

	result := &models.TimeNow{
		Unix:      now.Unix(),
		UnixMilli: now.UnixMilli(),
		UnixNano:  now.UnixNano(),
		ISO8601:   now.Format(time.RFC3339),
		RFC3339:   now.Format(time.RFC3339Nano),
		RFC1123:   now.Format(time.RFC1123),
		Human:     now.Format("02/01/2006 15:04:05 MST"),
	}

	// Add timezone conversions
	if len(timezones) > 0 {
		result.Timezones = make([]models.TimezoneResult, 0, len(timezones))
		for _, tz := range timezones {
			loc, err := time.LoadLocation(tz)
			if err != nil {
				continue
			}
			t := now.In(loc)
			_, offset := t.Zone()
			result.Timezones = append(result.Timezones, models.TimezoneResult{
				Timezone: tz,
				Time:     t.Format("02/01/2006 15:04:05"),
				Offset:   fmt.Sprintf("UTC%+d", offset/3600),
			})
		}
	}

	return result
}

// Convert converts a time string or timestamp to various formats
func (u *TimeUsecase) Convert(input string, toTimezones []string) *models.TimeConversion {
	result := &models.TimeConversion{
		Input:   input,
		Outputs: make(map[string]string),
	}

	var parsed time.Time

	// Handle "now" keyword
	if strings.ToLower(input) == "now" {
		parsed = time.Now()
		result.InputFormat = "Now"
		goto outputFormats
	}

	// Try parsing as Unix timestamp
	if ts, e := strconv.ParseInt(input, 10, 64); e == nil {
		if ts > 1e12 {
			// Milliseconds
			parsed = time.UnixMilli(ts)
			result.InputFormat = "Unix Milliseconds"
		} else if ts > 1e15 {
			// Nanoseconds
			parsed = time.Unix(0, ts)
			result.InputFormat = "Unix Nanoseconds"
		} else {
			// Seconds
			parsed = time.Unix(ts, 0)
			result.InputFormat = "Unix Seconds"
		}
	} else {
		// Try various formats
		formats := []struct {
			format string
			name   string
		}{
			{time.RFC3339, "RFC3339"},
			{time.RFC3339Nano, "RFC3339Nano"},
			{time.RFC1123, "RFC1123"},
			{time.RFC822, "RFC822"},
			{"2006-01-02T15:04:05", "ISO8601"},
			{"2006-01-02 15:04:05", "DateTime"},
			{"02/01/2006 15:04:05", "DateTime BR"},
			{"01/02/2006 15:04:05", "DateTime US"},
			{"2006-01-02", "Date"},
			{"02/01/2006", "Date BR"},
			{"01/02/2006", "Date US"},
		}

		for _, f := range formats {
			if t, e := time.Parse(f.format, input); e == nil {
				parsed = t
				result.InputFormat = f.name
				break
			}
		}

		if result.InputFormat == "" {
			result.InputFormat = "Unknown"
			return result
		}
	}

outputFormats:
	result.Parsed = parsed

	// Generate outputs
	result.Outputs["unix"] = fmt.Sprintf("%d", parsed.Unix())
	result.Outputs["unix_milli"] = fmt.Sprintf("%d", parsed.UnixMilli())
	result.Outputs["iso8601"] = parsed.Format(time.RFC3339)
	result.Outputs["rfc3339"] = parsed.Format(time.RFC3339Nano)
	result.Outputs["rfc1123"] = parsed.Format(time.RFC1123)
	result.Outputs["date"] = parsed.Format("2006-01-02")
	result.Outputs["time"] = parsed.Format("15:04:05")
	result.Outputs["human"] = parsed.Format("02/01/2006 15:04:05")
	result.Outputs["human_relative"] = u.relativeTime(parsed)

	// Timezone conversions
	if len(toTimezones) > 0 {
		result.Timezones = make([]models.TimezoneResult, 0, len(toTimezones))
		for _, tz := range toTimezones {
			loc, err := time.LoadLocation(tz)
			if err != nil {
				continue
			}
			t := parsed.In(loc)
			_, offset := t.Zone()
			result.Timezones = append(result.Timezones, models.TimezoneResult{
				Timezone: tz,
				Time:     t.Format("02/01/2006 15:04:05"),
				Offset:   fmt.Sprintf("UTC%+d", offset/3600),
			})
		}
	}

	return result
}

// Diff calculates the difference between two times
func (u *TimeUsecase) Diff(time1, time2 string) (string, error) {
	t1 := u.Convert(time1, nil)
	t2 := u.Convert(time2, nil)

	if t1.InputFormat == "Unknown" {
		return "", fmt.Errorf("não foi possível parsear: %s", time1)
	}
	if t2.InputFormat == "Unknown" {
		return "", fmt.Errorf("não foi possível parsear: %s", time2)
	}

	diff := t2.Parsed.Sub(t1.Parsed)
	return u.formatDuration(diff), nil
}

// Add adds duration to a time
func (u *TimeUsecase) Add(input string, duration string) (*models.TimeConversion, error) {
	t := u.Convert(input, nil)
	if t.InputFormat == "Unknown" {
		return nil, fmt.Errorf("não foi possível parsear: %s", input)
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return nil, fmt.Errorf("duração inválida: %s", duration)
	}

	result := t.Parsed.Add(d)
	return u.Convert(fmt.Sprintf("%d", result.Unix()), nil), nil
}

func (u *TimeUsecase) relativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		diff = -diff
		return "em " + u.formatDuration(diff)
	}

	return u.formatDuration(diff) + " atrás"
}

func (u *TimeUsecase) formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	parts := []string{}

	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d dias", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d horas", hours))
	}
	if minutes > 0 && days == 0 {
		parts = append(parts, fmt.Sprintf("%d minutos", minutes))
	}
	if seconds > 0 && days == 0 && hours == 0 {
		parts = append(parts, fmt.Sprintf("%d segundos", seconds))
	}

	if len(parts) == 0 {
		return "agora"
	}

	return strings.Join(parts, ", ")
}

// CommonTimezones returns a list of common timezones
func (u *TimeUsecase) CommonTimezones() []string {
	return []string{
		"UTC",
		"America/Sao_Paulo",
		"America/New_York",
		"America/Los_Angeles",
		"America/Chicago",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Asia/Singapore",
		"Australia/Sydney",
	}
}
