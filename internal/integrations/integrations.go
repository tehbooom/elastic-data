package integrations

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetIntegrations(repoPath string) ([]string, error) {
	basePath := filepath.Join(repoPath, "packages")
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", basePath, err)
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", basePath)
	}

	// Read directory entries
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	// Filter for directories only
	var directories []string
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	return directories, nil
}

func GetDatasets(repoPath, integration string) ([]string, error) {
	basePath := filepath.Join(repoPath, "packages", integration, "data_stream")
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", basePath, err)
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", basePath)
	}

	// Read directory entries
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	// Filter for directories only
	var directories []string
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	return directories, nil
}
