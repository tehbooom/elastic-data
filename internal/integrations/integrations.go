package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

// GetIntegrationsFromTemplates discovers integrations from pre-generated template files
func GetIntegrationsFromTemplates() ([]string, error) {
	templatesDir := "internal/integrations/templates"

	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("templates directory not found: %s", templatesDir)
	}

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to read templates directory %s: %w", templatesDir, err)
	}

	var integrations []string
	for _, entry := range entries {
		if entry.IsDir() {
			integrations = append(integrations, entry.Name())
		}
	}

	if len(integrations) == 0 {
		return nil, fmt.Errorf("no integrations found in templates directory")
	}

	return integrations, nil
}

// GetDatasetsFromTemplates gets datasets for an integration from template files
func GetDatasetsFromTemplates(integration string) ([]string, error) {
	integrationDir := filepath.Join("internal/integrations/templates", integration)

	// Check if integration directory exists
	if _, err := os.Stat(integrationDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("integration directory not found: %s", integrationDir)
	}

	entries, err := os.ReadDir(integrationDir)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to read integration directory %s: %w", integrationDir, err)
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
