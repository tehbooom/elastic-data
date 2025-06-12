package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

func GetIntegrations(repoPath string) ([]string, error) {
	basePath := filepath.Join(repoPath, "packages")
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to access path %s: %w", basePath, err)
	}
	if !fileInfo.IsDir() {
		log.Debug(err)
		return nil, fmt.Errorf("%s is not a directory", basePath)
	}
	entries, err := os.ReadDir(basePath)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}
	var directories []string
	for _, entry := range entries {
		if entry.IsDir() {
			if hasValidDatasets(repoPath, entry.Name()) {
				directories = append(directories, entry.Name())
			}
		}
	}
	return directories, nil
}

func GetDatasets(repoPath, integration string) ([]string, error) {
	basePath := filepath.Join(repoPath, "packages", integration, "data_stream")
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to access path %s: %w", basePath, err)
	}
	if !fileInfo.IsDir() {
		log.Debug(err)
		return nil, fmt.Errorf("%s is not a directory", basePath)
	}
	entries, err := os.ReadDir(basePath)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}
	var directories []string
	for _, entry := range entries {
		if entry.IsDir() {
			if IsValidDataset(repoPath, integration, entry.Name()) {
				directories = append(directories, entry.Name())
			}
		}
	}
	return directories, nil
}

func hasValidDatasets(repoPath, integration string) bool {
	basePath := filepath.Join(repoPath, "packages", integration, "data_stream")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if IsValidDataset(repoPath, integration, entry.Name()) {
				return true
			}
		}
	}
	return false
}

func IsValidDataset(repoPath, integration, dataset string) bool {
	basePath := filepath.Join(repoPath, "packages", integration, "data_stream", dataset, "_dev", "test", "pipeline")

	info, err := os.Stat(basePath)
	if err != nil || !info.IsDir() {
		return false
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if strings.Contains(fileName, "-expected.json") {
			continue
		}

		fileExt := strings.ToLower(filepath.Ext(fileName))
		if fileExt == ".json" || fileExt == ".log" {
			return true
		}
	}

	return false
}
