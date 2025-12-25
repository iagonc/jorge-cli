package playbook

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Discovery handles playbook discovery from multiple locations
type Discovery struct {
	logger *zap.Logger
	paths  []string
}

// NewDiscovery creates a new Discovery instance
func NewDiscovery(logger *zap.Logger) *Discovery {
	homeDir, _ := os.UserHomeDir()

	return &Discovery{
		logger: logger,
		paths: []string{
			"./playbooks",
			filepath.Join(homeDir, ".jorge", "playbooks"),
			"/etc/jorge/playbooks",
		},
	}
}

// DiscoverAll finds all playbooks in configured locations
func (d *Discovery) DiscoverAll() ([]models.PlaybookInfo, error) {
	var playbooks []models.PlaybookInfo

	for _, basePath := range d.paths {
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip inaccessible paths
			}

			if info.IsDir() {
				return nil
			}

			ext := filepath.Ext(path)
			if ext == ".yaml" || ext == ".yml" {
				pb, err := d.loadPlaybookInfo(path)
				if err == nil {
					playbooks = append(playbooks, pb)
				} else {
					d.logger.Debug("Failed to load playbook info",
						zap.String("path", path),
						zap.Error(err))
				}
			}

			return nil
		})

		if err != nil {
			d.logger.Warn("Error scanning playbook directory",
				zap.String("path", basePath),
				zap.Error(err))
		}
	}

	return playbooks, nil
}

func (d *Discovery) loadPlaybookInfo(path string) (models.PlaybookInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return models.PlaybookInfo{}, err
	}

	// Only parse minimal fields for discovery
	var partial struct {
		Name        string   `yaml:"name"`
		Version     string   `yaml:"version"`
		Description string   `yaml:"description"`
		Tags        []string `yaml:"tags"`
	}

	if err := yaml.Unmarshal(data, &partial); err != nil {
		return models.PlaybookInfo{}, err
	}

	if partial.Name == "" {
		return models.PlaybookInfo{}, fmt.Errorf("playbook has no name")
	}

	return models.PlaybookInfo{
		Name:        partial.Name,
		Path:        path,
		Description: partial.Description,
		Version:     partial.Version,
		Tags:        partial.Tags,
	}, nil
}

// FindByName locates a playbook by name
func (d *Discovery) FindByName(name string) (string, error) {
	playbooks, err := d.DiscoverAll()
	if err != nil {
		return "", err
	}

	for _, pb := range playbooks {
		if pb.Name == name {
			return pb.Path, nil
		}
	}

	return "", fmt.Errorf("playbook not found: %s", name)
}

// AddSearchPath adds a custom search path
func (d *Discovery) AddSearchPath(path string) {
	d.paths = append([]string{path}, d.paths...)
}
