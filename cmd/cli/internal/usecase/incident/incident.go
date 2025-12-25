package incident

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// IncidentUsecase handles incident management
type IncidentUsecase struct {
	logger     *zap.Logger
	storageDir string
}

// NewIncidentUsecase creates a new IncidentUsecase
func NewIncidentUsecase(logger *zap.Logger) *IncidentUsecase {
	homeDir, _ := os.UserHomeDir()
	storageDir := filepath.Join(homeDir, ".jorge", "incidents")
	os.MkdirAll(storageDir, 0755)

	return &IncidentUsecase{
		logger:     logger,
		storageDir: storageDir,
	}
}

// Create creates a new incident
func (u *IncidentUsecase) Create(title, description string, severity models.IncidentSeverity) (*models.Incident, error) {
	incident := &models.Incident{
		ID:          fmt.Sprintf("INC-%d", time.Now().Unix()),
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      models.StatusOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Timeline: []models.TimelineEvent{
			{
				Timestamp:   time.Now(),
				Type:        "created",
				Description: "Incident created",
			},
		},
	}

	if err := u.save(incident); err != nil {
		return nil, err
	}

	return incident, nil
}

// Get retrieves an incident by ID
func (u *IncidentUsecase) Get(id string) (*models.Incident, error) {
	path := filepath.Join(u.storageDir, id+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("incident not found: %s", id)
	}

	var incident models.Incident
	if err := yaml.Unmarshal(data, &incident); err != nil {
		return nil, err
	}

	return &incident, nil
}

// List lists all incidents
func (u *IncidentUsecase) List(status string, limit int) (*models.IncidentList, error) {
	files, err := os.ReadDir(u.storageDir)
	if err != nil {
		return &models.IncidentList{}, nil
	}

	result := &models.IncidentList{
		Incidents: []models.Incident{},
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

		var incident models.Incident
		if err := yaml.Unmarshal(data, &incident); err != nil {
			continue
		}

		if status != "" && string(incident.Status) != status {
			continue
		}

		result.Incidents = append(result.Incidents, incident)

		if incident.Status == models.StatusResolved {
			result.Resolved++
		} else {
			result.Open++
		}
	}

	// Sort by created time (newest first)
	sort.Slice(result.Incidents, func(i, j int) bool {
		return result.Incidents[i].CreatedAt.After(result.Incidents[j].CreatedAt)
	})

	// Apply limit
	if limit > 0 && len(result.Incidents) > limit {
		result.Incidents = result.Incidents[:limit]
	}

	result.Total = len(result.Incidents)
	return result, nil
}

// UpdateStatus updates an incident's status
func (u *IncidentUsecase) UpdateStatus(id string, status models.IncidentStatus, note string) (*models.Incident, error) {
	incident, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	incident.Status = status
	incident.UpdatedAt = time.Now()

	// Add timeline event
	event := models.TimelineEvent{
		Timestamp:   time.Now(),
		Type:        "status_change",
		Description: fmt.Sprintf("Status changed to %s", status),
	}
	if note != "" {
		event.Description += ": " + note
	}
	incident.Timeline = append(incident.Timeline, event)

	// Set resolved time if resolved
	if status == models.StatusResolved {
		now := time.Now()
		incident.ResolvedAt = &now
		incident.Duration = now.Sub(incident.CreatedAt).Round(time.Minute).String()
	}

	if err := u.save(incident); err != nil {
		return nil, err
	}

	return incident, nil
}

// AddNote adds a note to an incident timeline
func (u *IncidentUsecase) AddNote(id, note, author string) (*models.Incident, error) {
	incident, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	incident.UpdatedAt = time.Now()
	incident.Timeline = append(incident.Timeline, models.TimelineEvent{
		Timestamp:   time.Now(),
		Type:        "note",
		Description: note,
		Author:      author,
	})

	if err := u.save(incident); err != nil {
		return nil, err
	}

	return incident, nil
}

// SetCommander sets the incident commander
func (u *IncidentUsecase) SetCommander(id, commander string) (*models.Incident, error) {
	incident, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	incident.Commander = commander
	incident.UpdatedAt = time.Now()
	incident.Timeline = append(incident.Timeline, models.TimelineEvent{
		Timestamp:   time.Now(),
		Type:        "action",
		Description: fmt.Sprintf("Incident commander set to %s", commander),
	})

	if err := u.save(incident); err != nil {
		return nil, err
	}

	return incident, nil
}

// AddService adds an affected service
func (u *IncidentUsecase) AddService(id, service string) (*models.Incident, error) {
	incident, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	incident.Services = append(incident.Services, service)
	incident.UpdatedAt = time.Now()

	if err := u.save(incident); err != nil {
		return nil, err
	}

	return incident, nil
}

// GetSummary returns incident statistics
func (u *IncidentUsecase) GetSummary() (*models.IncidentSummary, error) {
	list, err := u.List("", 0)
	if err != nil {
		return nil, err
	}

	summary := &models.IncidentSummary{
		TotalIncidents: len(list.Incidents),
		ByStatus:       make(map[string]int),
		BySeverity:     make(map[string]int),
	}

	var totalDuration time.Duration
	var resolvedCount int

	for _, inc := range list.Incidents {
		summary.ByStatus[string(inc.Status)]++
		summary.BySeverity[string(inc.Severity)]++

		if inc.ResolvedAt != nil {
			duration := inc.ResolvedAt.Sub(inc.CreatedAt)
			totalDuration += duration
			resolvedCount++
		}
	}

	if resolvedCount > 0 {
		avgDuration := totalDuration / time.Duration(resolvedCount)
		summary.MTTR = avgDuration.Round(time.Minute).String()
		summary.AvgDuration = avgDuration.Round(time.Minute).String()
	}

	return summary, nil
}

// Export exports an incident to JSON or YAML
func (u *IncidentUsecase) Export(id, format string) (string, error) {
	incident, err := u.Get(id)
	if err != nil {
		return "", err
	}

	var data []byte
	switch format {
	case "json":
		data, err = json.MarshalIndent(incident, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(incident)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (u *IncidentUsecase) save(incident *models.Incident) error {
	data, err := yaml.Marshal(incident)
	if err != nil {
		return err
	}

	path := filepath.Join(u.storageDir, incident.ID+".yaml")
	return os.WriteFile(path, data, 0644)
}
