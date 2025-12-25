package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// LogsUsecase handles log analysis
type LogsUsecase struct {
	logger *zap.Logger
}

// NewLogsUsecase creates a new LogsUsecase
func NewLogsUsecase(logger *zap.Logger) *LogsUsecase {
	return &LogsUsecase{logger: logger}
}

// Analyze analyzes logs and returns statistics
func (u *LogsUsecase) Analyze(reader io.Reader, format models.LogFormat) (*models.LogStats, []models.LogEntry, error) {
	entries := make([]models.LogEntry, 0)
	stats := &models.LogStats{
		LevelCounts: make(map[string]int),
	}

	errorMessages := make(map[string]int)
	sources := make(map[string]int)
	var firstTime, lastTime time.Time

	scanner := bufio.NewScanner(reader)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		stats.TotalLines++

		entry, err := u.parseLine(line, format, lineNum)
		if err != nil {
			stats.FailedLines++
			continue
		}

		stats.ParsedLines++
		entries = append(entries, entry)

		// Update stats
		level := strings.ToUpper(entry.Level)
		if level == "" {
			level = "UNKNOWN"
		}
		stats.LevelCounts[level]++

		if entry.Source != "" {
			sources[entry.Source]++
		}

		// Track errors
		if u.isError(level) {
			errorMessages[entry.Message]++
		}

		// Time range
		if !entry.Timestamp.IsZero() {
			if firstTime.IsZero() || entry.Timestamp.Before(firstTime) {
				firstTime = entry.Timestamp
			}
			if entry.Timestamp.After(lastTime) {
				lastTime = entry.Timestamp
			}
		}
	}

	// Calculate error rate
	errorCount := stats.LevelCounts["ERROR"] + stats.LevelCounts["FATAL"] + stats.LevelCounts["CRITICAL"]
	if stats.ParsedLines > 0 {
		stats.ErrorRate = float64(errorCount) / float64(stats.ParsedLines) * 100
	}

	// Time range
	if !firstTime.IsZero() {
		stats.TimeRange = &models.TimeRange{
			Start: firstTime,
			End:   lastTime,
		}
	}

	// Top errors
	stats.TopErrors = u.topN(errorMessages, 10)

	// Top sources
	stats.TopSources = u.topSources(sources, 10)

	return stats, entries, scanner.Err()
}

// Filter filters log entries by criteria
func (u *LogsUsecase) Filter(entries []models.LogEntry, level string, pattern string, since, until time.Time) []models.LogEntry {
	var filtered []models.LogEntry
	var re *regexp.Regexp

	if pattern != "" {
		re, _ = regexp.Compile(pattern)
	}

	for _, entry := range entries {
		// Filter by level
		if level != "" && !strings.EqualFold(entry.Level, level) {
			continue
		}

		// Filter by pattern
		if re != nil && !re.MatchString(entry.Message) && !re.MatchString(entry.Raw) {
			continue
		}

		// Filter by time
		if !since.IsZero() && entry.Timestamp.Before(since) {
			continue
		}
		if !until.IsZero() && entry.Timestamp.After(until) {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

// DetectFormat attempts to detect the log format
func (u *LogsUsecase) DetectFormat(sample string) models.LogFormat {
	// Try JSON
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(sample), &js); err == nil {
		return models.LogFormatJSON
	}

	// Check for common patterns
	if strings.Contains(sample, " - - [") {
		return models.LogFormatNginx
	}

	if strings.Contains(sample, "] \"") && strings.Contains(sample, "\" ") {
		return models.LogFormatApache
	}

	if regexp.MustCompile(`^\w+\s+\d+\s+\d+:\d+:\d+`).MatchString(sample) {
		return models.LogFormatSyslog
	}

	return models.LogFormatCommon
}

func (u *LogsUsecase) parseLine(line string, format models.LogFormat, lineNum int) (models.LogEntry, error) {
	switch format {
	case models.LogFormatJSON:
		return u.parseJSON(line, lineNum)
	case models.LogFormatNginx:
		return u.parseNginx(line, lineNum)
	case models.LogFormatSyslog:
		return u.parseSyslog(line, lineNum)
	default:
		return u.parseCommon(line, lineNum)
	}
}

func (u *LogsUsecase) parseJSON(line string, lineNum int) (models.LogEntry, error) {
	entry := models.LogEntry{
		Raw:     line,
		LineNum: lineNum,
		Fields:  make(map[string]interface{}),
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return entry, err
	}

	// Extract common fields
	for _, key := range []string{"level", "lvl", "severity"} {
		if v, ok := data[key].(string); ok {
			entry.Level = v
			delete(data, key)
			break
		}
	}

	for _, key := range []string{"message", "msg", "text"} {
		if v, ok := data[key].(string); ok {
			entry.Message = v
			delete(data, key)
			break
		}
	}

	for _, key := range []string{"timestamp", "time", "ts", "@timestamp"} {
		if v, ok := data[key]; ok {
			if t, err := u.parseTimestamp(v); err == nil {
				entry.Timestamp = t
			}
			delete(data, key)
			break
		}
	}

	for _, key := range []string{"source", "logger", "component"} {
		if v, ok := data[key].(string); ok {
			entry.Source = v
			delete(data, key)
			break
		}
	}

	entry.Fields = data
	return entry, nil
}

func (u *LogsUsecase) parseNginx(line string, lineNum int) (models.LogEntry, error) {
	entry := models.LogEntry{
		Raw:     line,
		LineNum: lineNum,
	}

	// Combined log format pattern
	re := regexp.MustCompile(`^(\S+) - (\S+) \[([^\]]+)\] "([^"]*)" (\d+) (\d+)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 6 {
		entry.Fields = map[string]interface{}{
			"remote_addr": matches[1],
			"user":        matches[2],
			"request":     matches[4],
			"status":      matches[5],
			"bytes":       matches[6],
		}

		// Parse timestamp
		if t, err := time.Parse("02/Jan/2006:15:04:05 -0700", matches[3]); err == nil {
			entry.Timestamp = t
		}

		// Determine level based on status code
		status := matches[5]
		if strings.HasPrefix(status, "5") {
			entry.Level = "ERROR"
		} else if strings.HasPrefix(status, "4") {
			entry.Level = "WARN"
		} else {
			entry.Level = "INFO"
		}

		entry.Message = matches[4]
	} else {
		entry.Message = line
	}

	return entry, nil
}

func (u *LogsUsecase) parseSyslog(line string, lineNum int) (models.LogEntry, error) {
	entry := models.LogEntry{
		Raw:     line,
		LineNum: lineNum,
	}

	// Syslog pattern: Month Day HH:MM:SS hostname process[pid]: message
	re := regexp.MustCompile(`^(\w+\s+\d+\s+\d+:\d+:\d+)\s+(\S+)\s+(\S+?)(?:\[\d+\])?\s*:\s*(.*)$`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 5 {
		// Parse timestamp (add current year)
		timeStr := fmt.Sprintf("%d %s", time.Now().Year(), matches[1])
		if t, err := time.Parse("2006 Jan 2 15:04:05", timeStr); err == nil {
			entry.Timestamp = t
		}

		entry.Source = matches[3]
		entry.Message = matches[4]

		// Try to detect level from message
		entry.Level = u.detectLevel(matches[4])
	} else {
		entry.Message = line
	}

	return entry, nil
}

func (u *LogsUsecase) parseCommon(line string, lineNum int) (models.LogEntry, error) {
	entry := models.LogEntry{
		Raw:     line,
		LineNum: lineNum,
	}

	// Try common log patterns
	patterns := []struct {
		re    *regexp.Regexp
		level int
		msg   int
		time  int
	}{
		// [2023-01-01 10:00:00] [ERROR] message
		{regexp.MustCompile(`\[([^\]]+)\]\s*\[(\w+)\]\s*(.+)`), 2, 3, 1},
		// 2023-01-01 10:00:00 ERROR message
		{regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+(\w+)\s+(.+)`), 2, 3, 1},
		// ERROR: message
		{regexp.MustCompile(`^(\w+):\s*(.+)`), 1, 2, 0},
	}

	for _, p := range patterns {
		if matches := p.re.FindStringSubmatch(line); matches != nil {
			if p.level > 0 && p.level < len(matches) {
				entry.Level = matches[p.level]
			}
			if p.msg > 0 && p.msg < len(matches) {
				entry.Message = matches[p.msg]
			}
			if p.time > 0 && p.time < len(matches) {
				if t, err := u.parseTimestamp(matches[p.time]); err == nil {
					entry.Timestamp = t
				}
			}
			return entry, nil
		}
	}

	entry.Message = line
	entry.Level = u.detectLevel(line)
	return entry, nil
}

func (u *LogsUsecase) parseTimestamp(v interface{}) (time.Time, error) {
	switch t := v.(type) {
	case string:
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"02/01/2006 15:04:05",
		}
		for _, f := range formats {
			if parsed, err := time.Parse(f, t); err == nil {
				return parsed, nil
			}
		}
	case float64:
		return time.Unix(int64(t), 0), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse timestamp")
}

func (u *LogsUsecase) detectLevel(text string) string {
	text = strings.ToUpper(text)
	levels := []string{"FATAL", "ERROR", "WARN", "WARNING", "INFO", "DEBUG", "TRACE"}
	for _, level := range levels {
		if strings.Contains(text, level) {
			if level == "WARNING" {
				return "WARN"
			}
			return level
		}
	}
	return "INFO"
}

func (u *LogsUsecase) isError(level string) bool {
	return level == "ERROR" || level == "FATAL" || level == "CRITICAL"
}

func (u *LogsUsecase) topN(m map[string]int, n int) []models.ErrorCount {
	type kv struct {
		k string
		v int
	}
	var items []kv
	for k, v := range m {
		items = append(items, kv{k, v})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].v > items[j].v })

	result := make([]models.ErrorCount, 0, n)
	for i := 0; i < len(items) && i < n; i++ {
		result = append(result, models.ErrorCount{
			Message: items[i].k,
			Count:   items[i].v,
		})
	}
	return result
}

func (u *LogsUsecase) topSources(m map[string]int, n int) []models.SourceCount {
	type kv struct {
		k string
		v int
	}
	var items []kv
	for k, v := range m {
		items = append(items, kv{k, v})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].v > items[j].v })

	result := make([]models.SourceCount, 0, n)
	for i := 0; i < len(items) && i < n; i++ {
		result = append(result, models.SourceCount{
			Source: items[i].k,
			Count:  items[i].v,
		})
	}
	return result
}
