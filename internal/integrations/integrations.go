package integrations

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

//go:embed templates
var templatesFS embed.FS

// GetIntegrationsFromTemplates discovers integrations from embedded template files
func GetIntegrationsFromTemplates() ([]string, error) {
	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to read embedded templates directory: %w", err)
	}

	var integrations []string
	for _, entry := range entries {
		if entry.IsDir() {
			integrations = append(integrations, entry.Name())
		}
	}

	if len(integrations) == 0 {
		return nil, fmt.Errorf("no integrations found in embedded templates")
	}

	return integrations, nil
}

// GetDatasetsFromTemplates gets datasets for an integration from embedded template files
func GetDatasetsFromTemplates(integration string) ([]string, error) {
	integrationDir := filepath.Join("templates", integration)
	
	entries, err := templatesFS.ReadDir(integrationDir)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("integration directory not found: %s", integrationDir)
	}

	var datasets []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
			// Remove .tmpl extension to get dataset name
			dataset := strings.TrimSuffix(entry.Name(), ".tmpl")
			datasets = append(datasets, dataset)
		}
	}

	if len(datasets) == 0 {
		return nil, fmt.Errorf("no datasets found for integration %s", integration)
	}

	return datasets, nil
}
