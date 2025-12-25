package postmortem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// PostmortemUsecase handles postmortem operations
type PostmortemUsecase struct {
	logger     *zap.Logger
	storageDir string
}

// NewPostmortemUsecase creates a new PostmortemUsecase
func NewPostmortemUsecase(logger *zap.Logger) *PostmortemUsecase {
	homeDir, _ := os.UserHomeDir()
	storageDir := filepath.Join(homeDir, ".jorge", "postmortems")
	os.MkdirAll(storageDir, 0755)

	return &PostmortemUsecase{
		logger:     logger,
		storageDir: storageDir,
	}
}

// Create creates a new postmortem
func (u *PostmortemUsecase) Create(incidentID, title string, severity string) (*models.Postmortem, error) {
	pm := &models.Postmortem{
		ID:         fmt.Sprintf("PM-%d", time.Now().Unix()),
		IncidentID: incidentID,
		Title:      title,
		Date:       time.Now(),
		Severity:   severity,
		Status:     models.PostmortemDraft,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Impact: models.ImpactSummary{
			ServicesAffected: []string{},
		},
		Timeline:       []models.TimelineEntry{},
		RootCauses:     []string{},
		ActionItems:    []models.ActionItem{},
		WentWell:       []string{},
		WentPoorly:     []string{},
		LessonsLearned: []string{},
	}

	if err := u.save(pm); err != nil {
		return nil, err
	}

	return pm, nil
}

// Get retrieves a postmortem by ID
func (u *PostmortemUsecase) Get(id string) (*models.Postmortem, error) {
	path := filepath.Join(u.storageDir, id+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("postmortem not found: %s", id)
	}

	var pm models.Postmortem
	if err := yaml.Unmarshal(data, &pm); err != nil {
		return nil, err
	}

	return &pm, nil
}

// List lists all postmortems
func (u *PostmortemUsecase) List(status string, limit int) (*models.PostmortemList, error) {
	files, err := os.ReadDir(u.storageDir)
	if err != nil {
		return &models.PostmortemList{}, nil
	}

	result := &models.PostmortemList{
		Postmortems: []models.Postmortem{},
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(u.storageDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var pm models.Postmortem
		if err := yaml.Unmarshal(data, &pm); err != nil {
			continue
		}

		if status != "" && string(pm.Status) != status {
			continue
		}

		result.Postmortems = append(result.Postmortems, pm)
	}

	// Sort by date (newest first)
	sort.Slice(result.Postmortems, func(i, j int) bool {
		return result.Postmortems[i].Date.After(result.Postmortems[j].Date)
	})

	if limit > 0 && len(result.Postmortems) > limit {
		result.Postmortems = result.Postmortems[:limit]
	}

	result.Total = len(result.Postmortems)
	return result, nil
}

// UpdateStatus updates postmortem status
func (u *PostmortemUsecase) UpdateStatus(id string, status models.PostmortemStatus) (*models.Postmortem, error) {
	pm, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	pm.Status = status
	pm.UpdatedAt = time.Now()

	if status == models.PostmortemApproved || status == models.PostmortemPublished {
		now := time.Now()
		pm.ReviewedAt = &now
	}

	if err := u.save(pm); err != nil {
		return nil, err
	}

	return pm, nil
}

// SetSummary sets the postmortem summary
func (u *PostmortemUsecase) SetSummary(id, summary string) (*models.Postmortem, error) {
	pm, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	pm.Summary = summary
	pm.UpdatedAt = time.Now()

	if err := u.save(pm); err != nil {
		return nil, err
	}

	return pm, nil
}

// AddRootCause adds a root cause
func (u *PostmortemUsecase) AddRootCause(id, cause string) (*models.Postmortem, error) {
	pm, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	pm.RootCauses = append(pm.RootCauses, cause)
	pm.UpdatedAt = time.Now()

	if err := u.save(pm); err != nil {
		return nil, err
	}

	return pm, nil
}

// AddActionItem adds an action item
func (u *PostmortemUsecase) AddActionItem(id string, item models.ActionItem) (*models.Postmortem, error) {
	pm, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	item.ID = fmt.Sprintf("AI-%d", len(pm.ActionItems)+1)
	item.Status = "todo"
	pm.ActionItems = append(pm.ActionItems, item)
	pm.UpdatedAt = time.Now()

	if err := u.save(pm); err != nil {
		return nil, err
	}

	return pm, nil
}

// AddTimeline adds a timeline entry
func (u *PostmortemUsecase) AddTimeline(id string, entry models.TimelineEntry) (*models.Postmortem, error) {
	pm, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	pm.Timeline = append(pm.Timeline, entry)
	// Sort by time
	sort.Slice(pm.Timeline, func(i, j int) bool {
		return pm.Timeline[i].Time.Before(pm.Timeline[j].Time)
	})
	pm.UpdatedAt = time.Now()

	if err := u.save(pm); err != nil {
		return nil, err
	}

	return pm, nil
}

// AddLesson adds a lesson learned
func (u *PostmortemUsecase) AddLesson(id, lesson string) (*models.Postmortem, error) {
	pm, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	pm.LessonsLearned = append(pm.LessonsLearned, lesson)
	pm.UpdatedAt = time.Now()

	if err := u.save(pm); err != nil {
		return nil, err
	}

	return pm, nil
}

// Export exports a postmortem to various formats
func (u *PostmortemUsecase) Export(id, format string) (string, error) {
	pm, err := u.Get(id)
	if err != nil {
		return "", err
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(pm, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "yaml":
		data, err := yaml.Marshal(pm)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "markdown", "md":
		return u.exportMarkdown(pm), nil

	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func (u *PostmortemUsecase) exportMarkdown(pm *models.Postmortem) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", pm.Title))
	sb.WriteString(fmt.Sprintf("**Incident ID:** %s\n", pm.IncidentID))
	sb.WriteString(fmt.Sprintf("**Date:** %s\n", pm.Date.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("**Severity:** %s\n", pm.Severity))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", pm.Status))

	if len(pm.Authors) > 0 {
		sb.WriteString(fmt.Sprintf("**Authors:** %s\n\n", strings.Join(pm.Authors, ", ")))
	}

	sb.WriteString("## Summary\n\n")
	sb.WriteString(pm.Summary + "\n\n")

	sb.WriteString("## Impact\n\n")
	sb.WriteString(fmt.Sprintf("- **Duration:** %s\n", pm.Impact.Duration))
	if pm.Impact.UsersAffected != "" {
		sb.WriteString(fmt.Sprintf("- **Users Affected:** %s\n", pm.Impact.UsersAffected))
	}
	if len(pm.Impact.ServicesAffected) > 0 {
		sb.WriteString(fmt.Sprintf("- **Services Affected:** %s\n", strings.Join(pm.Impact.ServicesAffected, ", ")))
	}
	sb.WriteString("\n")

	if len(pm.Timeline) > 0 {
		sb.WriteString("## Timeline\n\n")
		for _, entry := range pm.Timeline {
			sb.WriteString(fmt.Sprintf("- **%s** - %s\n", entry.Time.Format("15:04"), entry.Description))
		}
		sb.WriteString("\n")
	}

	if len(pm.RootCauses) > 0 {
		sb.WriteString("## Root Causes\n\n")
		for _, cause := range pm.RootCauses {
			sb.WriteString(fmt.Sprintf("- %s\n", cause))
		}
		sb.WriteString("\n")
	}

	if len(pm.WentWell) > 0 {
		sb.WriteString("## What Went Well\n\n")
		for _, item := range pm.WentWell {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(pm.WentPoorly) > 0 {
		sb.WriteString("## What Went Poorly\n\n")
		for _, item := range pm.WentPoorly {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(pm.ActionItems) > 0 {
		sb.WriteString("## Action Items\n\n")
		sb.WriteString("| ID | Description | Type | Priority | Owner | Status |\n")
		sb.WriteString("|---|---|---|---|---|---|\n")
		for _, item := range pm.ActionItems {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
				item.ID, item.Description, item.Type, item.Priority, item.Owner, item.Status))
		}
		sb.WriteString("\n")
	}

	if len(pm.LessonsLearned) > 0 {
		sb.WriteString("## Lessons Learned\n\n")
		for _, lesson := range pm.LessonsLearned {
			sb.WriteString(fmt.Sprintf("- %s\n", lesson))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GenerateTemplate generates a postmortem template
func (u *PostmortemUsecase) GenerateTemplate() string {
	return `# Postmortem: [Incident Title]

## Incident Details
- **Incident ID:** INC-XXX
- **Date:** YYYY-MM-DD
- **Severity:** SEV1/SEV2/SEV3
- **Duration:** X hours Y minutes
- **Authors:** @author1, @author2

## Summary
[Brief description of what happened]

## Impact
- **Users Affected:** X users
- **Requests Failed:** X%
- **Revenue Impact:** $X
- **SLA Breach:** Yes/No
- **Services Affected:** service1, service2

## Timeline
- **HH:MM** - [Event description]
- **HH:MM** - [Event description]
- **HH:MM** - [Event description]

## Root Causes
1. [Primary root cause]
2. [Contributing factor]

## What Went Well
- [Positive observation]

## What Went Poorly
- [Area for improvement]

## Where We Got Lucky
- [Lucky factor that prevented worse outcome]

## Action Items
| ID | Description | Type | Priority | Owner | Due Date | Status |
|---|---|---|---|---|---|---|
| AI-1 | [Action description] | prevent | P0 | @owner | YYYY-MM-DD | todo |

## Lessons Learned
1. [Key takeaway]

---
*Generated by Jorge CLI*
`
}

func (u *PostmortemUsecase) save(pm *models.Postmortem) error {
	data, err := yaml.Marshal(pm)
	if err != nil {
		return err
	}

	path := filepath.Join(u.storageDir, pm.ID+".yaml")
	return os.WriteFile(path, data, 0644)
}
