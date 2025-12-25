package cronx

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// CronUsecase handles cron expression operations
type CronUsecase struct {
	logger *zap.Logger
}

// NewCronUsecase creates a new CronUsecase
func NewCronUsecase(logger *zap.Logger) *CronUsecase {
	return &CronUsecase{logger: logger}
}

// Parse parses a cron expression and returns human-readable description
func (u *CronUsecase) Parse(expression string) *models.CronExpression {
	result := &models.CronExpression{
		Expression: expression,
	}

	// Parse the expression
	fields := strings.Fields(expression)
	if len(fields) < 5 || len(fields) > 6 {
		result.Error = fmt.Sprintf("expressão inválida: esperado 5-6 campos, recebido %d", len(fields))
		return result
	}

	// Handle 6-field cron (with seconds) by removing seconds
	if len(fields) == 6 {
		fields = fields[1:] // Remove seconds field
	}

	result.IsValid = true
	result.Fields = &models.CronFields{
		Minute:     fields[0],
		Hour:       fields[1],
		DayOfMonth: fields[2],
		Month:      fields[3],
		DayOfWeek:  fields[4],
	}

	// Generate human-readable description
	result.Description = u.describe(result.Fields)

	return result
}

// NextRuns calculates the next N runs of a cron expression
func (u *CronUsecase) NextRuns(expression string, count int) *models.CronExpression {
	result := u.Parse(expression)
	if !result.IsValid {
		return result
	}

	// Simple next run calculation
	now := time.Now()
	result.NextRuns = make([]time.Time, 0, count)

	// This is a simplified implementation
	// For production, use a proper cron library like github.com/robfig/cron
	for i := 0; i < count && len(result.NextRuns) < count; i++ {
		next := u.calculateNext(result.Fields, now.Add(time.Duration(i)*time.Minute))
		if next.After(now) && (len(result.NextRuns) == 0 || next.After(result.NextRuns[len(result.NextRuns)-1])) {
			result.NextRuns = append(result.NextRuns, next)
		}
	}

	// Fallback: generate approximate times
	if len(result.NextRuns) == 0 {
		result.NextRuns = u.approximateNextRuns(result.Fields, now, count)
	}

	return result
}

func (u *CronUsecase) calculateNext(fields *models.CronFields, from time.Time) time.Time {
	// Simplified calculation - rounds to next matching minute/hour
	minute := u.parseField(fields.Minute, 0, 59)
	hour := u.parseField(fields.Hour, 0, 23)

	next := time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, from.Location())
	if next.Before(from) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func (u *CronUsecase) approximateNextRuns(fields *models.CronFields, from time.Time, count int) []time.Time {
	runs := make([]time.Time, 0, count)
	current := from.Truncate(time.Minute).Add(time.Minute)

	for i := 0; i < 10080 && len(runs) < count; i++ { // Check up to 1 week
		if u.matches(fields, current) {
			runs = append(runs, current)
		}
		current = current.Add(time.Minute)
	}

	return runs
}

func (u *CronUsecase) matches(fields *models.CronFields, t time.Time) bool {
	return u.fieldMatches(fields.Minute, t.Minute()) &&
		u.fieldMatches(fields.Hour, t.Hour()) &&
		u.fieldMatches(fields.DayOfMonth, t.Day()) &&
		u.fieldMatches(fields.Month, int(t.Month())) &&
		u.fieldMatches(fields.DayOfWeek, int(t.Weekday()))
}

func (u *CronUsecase) fieldMatches(field string, value int) bool {
	if field == "*" {
		return true
	}

	// Handle */n
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(field[2:])
		if err != nil {
			return false
		}
		return value%step == 0
	}

	// Handle ranges (1-5)
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) == 2 {
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			return value >= start && value <= end
		}
	}

	// Handle lists (1,3,5)
	if strings.Contains(field, ",") {
		for _, v := range strings.Split(field, ",") {
			if n, err := strconv.Atoi(v); err == nil && n == value {
				return true
			}
		}
		return false
	}

	// Simple number
	n, err := strconv.Atoi(field)
	return err == nil && n == value
}

func (u *CronUsecase) parseField(field string, min, max int) int {
	if field == "*" {
		return min
	}
	if strings.HasPrefix(field, "*/") {
		return min
	}
	if n, err := strconv.Atoi(field); err == nil {
		return n
	}
	return min
}

func (u *CronUsecase) describe(fields *models.CronFields) string {
	parts := []string{}

	// Minute
	minDesc := u.describeField(fields.Minute, "minuto")
	// Hour
	hourDesc := u.describeField(fields.Hour, "hora")
	// Day of month
	domDesc := u.describeField(fields.DayOfMonth, "dia")
	// Month
	monthDesc := u.describeField(fields.Month, "mês")
	// Day of week
	dowDesc := u.describeDayOfWeek(fields.DayOfWeek)

	// Common patterns
	if fields.Minute == "0" && fields.Hour == "0" && fields.DayOfMonth == "*" && fields.Month == "*" && fields.DayOfWeek == "*" {
		return "Todo dia à meia-noite"
	}
	if fields.Minute == "0" && fields.Hour == "*" && fields.DayOfMonth == "*" && fields.Month == "*" && fields.DayOfWeek == "*" {
		return "A cada hora, no minuto 0"
	}
	if strings.HasPrefix(fields.Minute, "*/") && fields.Hour == "*" && fields.DayOfMonth == "*" && fields.Month == "*" && fields.DayOfWeek == "*" {
		interval := fields.Minute[2:]
		return fmt.Sprintf("A cada %s minutos", interval)
	}
	if fields.Minute == "*" && strings.HasPrefix(fields.Hour, "*/") && fields.DayOfMonth == "*" && fields.Month == "*" && fields.DayOfWeek == "*" {
		interval := fields.Hour[2:]
		return fmt.Sprintf("A cada %s horas", interval)
	}
	if fields.Minute == "0" && fields.Hour == "0" && fields.DayOfMonth == "1" && fields.Month == "*" && fields.DayOfWeek == "*" {
		return "No primeiro dia de cada mês à meia-noite"
	}

	// Build description
	if minDesc != "" {
		parts = append(parts, minDesc)
	}
	if hourDesc != "" {
		parts = append(parts, hourDesc)
	}
	if domDesc != "" {
		parts = append(parts, domDesc)
	}
	if monthDesc != "" {
		parts = append(parts, monthDesc)
	}
	if dowDesc != "" {
		parts = append(parts, dowDesc)
	}

	if len(parts) == 0 {
		return "Executar conforme expressão cron"
	}

	return strings.Join(parts, ", ")
}

func (u *CronUsecase) describeField(field, name string) string {
	if field == "*" {
		return ""
	}
	if strings.HasPrefix(field, "*/") {
		return fmt.Sprintf("a cada %s %ss", field[2:], name)
	}
	if strings.Contains(field, ",") {
		return fmt.Sprintf("nos %ss %s", name, field)
	}
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		return fmt.Sprintf("do %s %s ao %s", name, parts[0], parts[1])
	}
	return fmt.Sprintf("no %s %s", name, field)
}

func (u *CronUsecase) describeDayOfWeek(field string) string {
	days := map[string]string{
		"0": "Domingo", "7": "Domingo",
		"1": "Segunda", "2": "Terça", "3": "Quarta",
		"4": "Quinta", "5": "Sexta", "6": "Sábado",
	}

	if field == "*" {
		return ""
	}
	if day, ok := days[field]; ok {
		return fmt.Sprintf("às %ss", day)
	}
	if field == "1-5" {
		return "de Segunda a Sexta"
	}
	if field == "0,6" || field == "6,0" {
		return "nos fins de semana"
	}
	return fmt.Sprintf("nos dias da semana %s", field)
}
