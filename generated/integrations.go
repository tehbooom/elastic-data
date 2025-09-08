package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/elastic/beats/v7/libbeat/reader/multiline"
)

func main() {
	validIntegrations, allIntegrations, err := getIntegrations()
	if err != nil {
		log.Fatal(err)
	}

	for _, integration := range validIntegrations {
		datasets, err := getDatasets(integration)
		if err != nil {
			log.Fatal(err)
		}
		for _, dataset := range datasets {
			generateIntegrationTemplate(integration, dataset)
		}
	}
	invalidIntegrations := findInvalidIntegrations(validIntegrations, allIntegrations)

	fmt.Println(invalidIntegrations)
}

func findInvalidIntegrations(valid, all []string) []string {
	validMap := make(map[string]bool, len(valid))
	for _, str := range valid {
		validMap[str] = true
	}

	var invalid []string
	for _, str := range all {
		if !validMap[str] {
			invalid = append(invalid, str)
		}
	}

	return invalid
}

// generateIntegrationTemplate generates .tmpl files for each integration/dataset combination
// Each .tmpl file contains multiple events that the application can randomly select from
func generateIntegrationTemplate(integration, dataset string) {
	log.Printf("Processing %s:%s", integration, dataset)

	datasetPath := filepath.Join("./integrations", "packages", integration, "data_stream", dataset)
	basePath := filepath.Join(datasetPath, "_dev", "test", "pipeline")

	// Create templates directory structure
	templatesDir := filepath.Join("../internal/integrations/templates", integration)
	err := os.MkdirAll(templatesDir, 0755)
	if err != nil {
		log.Printf("Failed to create templates directory %s: %v", templatesDir, err)
		return
	}

	// Get multiline config for this dataset
	multilineConfig, err := GetMultiLineConfig(datasetPath)
	if err != nil {
		log.Printf("Warning: failed to get multiline config for %s:%s: %v", integration, dataset, err)
		// Continue without multiline config
	}

	var allEventTemplates []string

	err = filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.Contains(path, "-expected.json") {
			return nil
		}

		switch filepath.Ext(path) {
		case ".json":
			eventTemplates, err := processJSONFileForTemplates(path)
			if err != nil {
				log.Printf("  Warning: failed to process JSON file %s: %v", path, err)
				return nil // Continue processing other files
			}
			allEventTemplates = append(allEventTemplates, eventTemplates...)

		case ".log":
			eventTemplates, err := processLogFileForTemplates(path, multilineConfig)
			if err != nil {
				log.Printf("  Warning: failed to process log file %s: %v", path, err)
				return nil // Continue processing other files
			}
			allEventTemplates = append(allEventTemplates, eventTemplates...)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking directory %s: %v", basePath, err)
		return
	}

	if len(allEventTemplates) > 0 {
		templateFile := filepath.Join(templatesDir, dataset+".tmpl")
		err := writeTemplateFile(templateFile, allEventTemplates)
		if err != nil {
			log.Printf("Failed to write template file %s: %v", templateFile, err)
			return
		}
		log.Printf("  Created %s with %d event templates", templateFile, len(allEventTemplates))
	} else {
		log.Printf("  No templates found for %s:%s", integration, dataset)
	}
}

// getIntegrations returns the list of valid integrations, all integrations, and an error
func getIntegrations() ([]string, []string, error) {
	repoPath := "./integrations/"
	basePath := filepath.Join(repoPath, "packages")
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to access path %s: %w", basePath, err)
	}
	if !fileInfo.IsDir() {
		return nil, nil, fmt.Errorf("%s is not a directory", basePath)
	}
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}
	var validDirectories []string
	var integrations []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// if entry.Name() == "system" {
		// 	continue
		// }

		integrations = append(integrations, entry.Name())

		if hasValidDatasets(repoPath, entry.Name()) {
			validDirectories = append(validDirectories, entry.Name())
		}
	}
	return validDirectories, integrations, nil
}

func hasValidDatasets(repoPath, integration string) bool {
	basePath := filepath.Join(repoPath, "packages", integration, "data_stream")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if isValidDataset(repoPath, integration, entry.Name()) {
				return true
			}
		}
	}
	return false
}

func isValidDataset(repoPath, integration, dataset string) bool {
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

func getDatasets(integration string) ([]string, error) {
	repoPath := "./integrations/"
	basePath := filepath.Join(repoPath, "packages", integration, "data_stream")
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", basePath, err)
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", basePath)
	}
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}
	var directories []string
	for _, entry := range entries {
		if entry.IsDir() {
			if isValidDataset(repoPath, integration, entry.Name()) {
				directories = append(directories, entry.Name())
			}
		}
	}
	return directories, nil
}

// processJSONFileForTemplates processes JSON files and returns template strings
func processJSONFileForTemplates(filePath string) ([]string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var eventTemplates []string

	var JSONLog struct {
		Events []map[string]any `json:"events"`
	}

	err = json.Unmarshal(file, &JSONLog)
	if err != nil {
		return nil, err
	}

	for _, event := range JSONLog.Events {
		// Remove @timestamp field if it exists
		delete(event, "@timestamp")

		// Special handling for Snort rule field BEFORE template processing - convert string format to object to avoid mapping conflicts
		if ruleValue, exists := event["rule"]; exists {
			if ruleStr, ok := ruleValue.(string); ok && strings.Contains(ruleStr, ":") {
				// Convert "1:10000001:0" format to object structure
				parts := strings.Split(ruleStr, ":")
				if len(parts) == 3 {
					ruleObj := map[string]interface{}{
						"gid": parts[0],
						"sid": parts[1],
						"rev": parts[2],
					}
					event["rule"] = ruleObj
					log.Printf("        Converted rule field from '%s' to object", ruleStr)
				} else {
					log.Printf("        Rule field '%s' split into %d parts, expected 3", ruleStr, len(parts))
				}
			} else {
				log.Printf("        Rule field exists but is not a string or doesn't contain ':'")
			}
		}

		// Process ALL remaining fields to create template patterns
		processMapForTemplating(event)

		// Convert back to JSON template
		templateJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("        Warning: failed to marshal event: %v", err)
			continue
		}

		var compactBuffer bytes.Buffer
		if err := json.Compact(&compactBuffer, templateJSON); err != nil {
			log.Printf("        Warning: failed to compact JSON: %v", err)
			continue
		}

		eventTemplates = append(eventTemplates, compactBuffer.String())
	}

	return eventTemplates, nil
}

// processLogFileForTemplates processes log files and returns template strings
func processLogFileForTemplates(filePath string, multilineConfig *multiline.Config) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Failed to close file: %v\n", err)
		}
	}()

	finalReader, err := createReaderPipeline(file, multilineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader pipeline: %w", err)
	}

	var eventTemplates []string

	for {
		message, err := finalReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read message at position: %w", err)
		}

		if len(message.Content) == 0 {
			continue
		}

		templateEvent := processLogLineForTemplate(message.Content)
		if templateEvent != "" {
			eventTemplates = append(eventTemplates, templateEvent)
		}
	}

	return eventTemplates, nil
}

// processLogLineForTemplate converts a log line into a template string
func processLogLineForTemplate(line []byte) string {
	original := string(line)

	// Apply common pattern replacements
	patterns := []struct {
		name  string
		regex *regexp.Regexp
	}{
		{"timestamp_clf_timezone", regexp.MustCompile(`\[[\d]{2}\/[A-Za-z]{3}\/[\d]{4}:[\d]{2}:[\d]{2}:[\d]{2} [+-][\d]{4}\]`)},
		{"timestamp_clf", regexp.MustCompile(`\[[\d]{2}\/[A-Za-z]{3}\/[\d]{4}:[\d]{2}:[\d]{2}:[\d]{2}\]`)},
		{"timestamp_common", regexp.MustCompile(`[\d]{2}\/[A-Za-z]{3}\/[\d]{4}:[\d]{2}:[\d]{2}:[\d]{2}`)},
		{"timestamp_iso", regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?`)},
		{"timestamp_syslog", regexp.MustCompile(`[A-Za-z]{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}`)},
		{"timestamp_snort", regexp.MustCompile(`\d{2}/\d{2}/\d{2}-\d{2}:\d{2}:\d{2}\.\d{6}`)},
		{"timestamp_snort_no_year", regexp.MustCompile(`\d{2}/\d{2}-\d{2}:\d{2}:\d{2}\.\d{6}`)},
		{"timestamp_unix_s", regexp.MustCompile(`\b1[0-9]{9}\b`)},
		{"timestamp_unix_ms", regexp.MustCompile(`\b1[0-9]{12}\b`)},
		{"IPs", regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)},
		{"Emails", regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)},
		{"Domains", regexp.MustCompile(`[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)},
	}

	templateStr := original
	for _, p := range patterns {
		templateStr = p.regex.ReplaceAllString(templateStr, fmt.Sprintf("{{.%s}}", p.name))
	}

	// Validate template
	_, err := template.New("test").Parse(templateStr)
	if err != nil {
		log.Printf("        Warning: invalid template created from line: %s", original[:min(len(original), 100)])
		return original // Return original if template is invalid
	}

	return templateStr
}

func processMapForTemplating(data map[string]interface{}) {
	// Create a value tracker for this processing session
	valueTracker := make(map[string]map[string]int)
	processMapForTemplatingWithTracker(data, valueTracker)
}

func processMapForTemplatingWithTracker(data map[string]interface{}, valueTracker map[string]map[string]int) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			placeholder := convertStringValueToTemplateWithTracker(key, v, valueTracker)
			data[key] = placeholder
		case float64:
			placeholder := convertNumericValueToTemplateWithTracker(key, v, valueTracker)
			if placeholder != "" {
				data[key] = placeholder
			}
		case int64:
			placeholder := convertNumericValueToTemplateWithTracker(key, float64(v), valueTracker)
			if placeholder != "" {
				data[key] = placeholder
			}
		case map[string]interface{}:
			processMapForTemplatingWithTracker(v, valueTracker)
		case []interface{}:
			processSliceForTemplatingWithTracker(key, v, valueTracker)
		}
	}
}

func processSliceForTemplatingWithTracker(key string, slice []interface{}, valueTracker map[string]map[string]int) {
	for i, value := range slice {
		switch v := value.(type) {
		case string:
			placeholder := convertStringValueToTemplateWithTracker(key, v, valueTracker)
			slice[i] = placeholder
		case map[string]interface{}:
			processMapForTemplatingWithTracker(v, valueTracker)
		case []interface{}:
			processSliceForTemplatingWithTracker(key, v, valueTracker)
		}
	}
}

func convertNumericValueToTemplateWithTracker(k string, value float64, valueTracker map[string]map[string]int) string {
	key := strings.ToLower(k)

	if strings.Contains(key, "time") || strings.Contains(key, "timestamp") {
		valueStr := fmt.Sprintf("%.0f", value)

		if regexp.MustCompile(`\b1[0-9]{12}\b`).MatchString(valueStr) {
			return getUniqueVariableName("timestamp_unix_ms", valueStr, valueTracker)
		} else if regexp.MustCompile(`\b1[0-9]{9}\b`).MatchString(valueStr) {
			return getUniqueVariableName("timestamp_unix_s", valueStr, valueTracker)
		}
	}

	return ""
}

func convertStringValueToTemplateWithTracker(k, value string, valueTracker map[string]map[string]int) string {
	value = strings.TrimSpace(value)

	// Check if this is a JSON string - if so, parse and template it recursively
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		var innerJSON map[string]interface{}
		if err := json.Unmarshal([]byte(value), &innerJSON); err == nil {
			// This is valid JSON, process it recursively with the same tracker
			processMapForTemplatingWithTracker(innerJSON, valueTracker)
			// Convert back to JSON string
			if templatedJSON, err := json.Marshal(innerJSON); err == nil {
				return string(templatedJSON)
			}
		}
	}

	// Apply pattern matching for non-JSON strings
	if regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`).MatchString(value) {
		return getUniqueVariableName("Emails", value, valueTracker)
	}

	if regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`).MatchString(value) {
		return getUniqueVariableName("IPs", value, valueTracker)
	}

	if regexp.MustCompile(`[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`).MatchString(value) {
		return getUniqueVariableName("Domains", value, valueTracker)
	}

	if regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?`).MatchString(value) {
		return getUniqueVariableName("timestamp_iso", value, valueTracker)
	}

	if regexp.MustCompile(`[\d]{2}\/[A-Za-z]{3}\/[\d]{4}:[\d]{2}:[\d]{2}:[\d]{2}`).MatchString(value) {
		return getUniqueVariableName("timestamp_common", value, valueTracker)
	}

	if regexp.MustCompile(`\[[\d]{2}\/[A-Za-z]{3}\/[\d]{4}:[\d]{2}:[\d]{2}:[\d]{2}\]`).MatchString(value) {
		return getUniqueVariableName("timestamp_clf", value, valueTracker)
	}

	if regexp.MustCompile(`\[[\d]{2}\/[A-Za-z]{3}\/[\d]{4}:[\d]{2}:[\d]{2}:[\d]{2} [+-][\d]{4}\]`).MatchString(value) {
		return getUniqueVariableName("timestamp_clf_timezone", value, valueTracker)
	}

	if regexp.MustCompile(`[A-Za-z]{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}`).MatchString(value) {
		return getUniqueVariableName("timestamp_syslog", value, valueTracker)
	}

	if regexp.MustCompile(`\d{2}/\d{2}/\d{2}-\d{2}:\d{2}:\d{2}\.\d{6}`).MatchString(value) {
		return getUniqueVariableName("timestamp_snort", value, valueTracker)
	}

	if regexp.MustCompile(`\d{2}/\d{2}-\d{2}:\d{2}:\d{2}\.\d{6}`).MatchString(value) {
		return getUniqueVariableName("timestamp_snort_no_year", value, valueTracker)
	}

	if regexp.MustCompile(`\b1[0-9]{12}\b`).MatchString(value) {
		return getUniqueVariableName("timestamp_unix_ms", value, valueTracker)
	}

	if regexp.MustCompile(`\b1[0-9]{9}\b`).MatchString(value) {
		return getUniqueVariableName("timestamp_unix_s", value, valueTracker)
	}

	key := strings.ToLower(k)

	if strings.Contains(key, "username") || strings.Contains(key, "user") {
		return getUniqueVariableName("Users", value, valueTracker)
	} else if strings.Contains(key, "hostname") || strings.Contains(key, "host") {
		return getUniqueVariableName("Hosts", value, valueTracker)
	}

	return value
}

// getUniqueVariableName tracks unique values for each variable type and returns numbered variants
func getUniqueVariableName(varType, value string, valueTracker map[string]map[string]int) string {
	// Initialize the map for this variable type if it doesn't exist
	if valueTracker[varType] == nil {
		valueTracker[varType] = make(map[string]int)
	}

	// Check if we've seen this exact value before
	if index, exists := valueTracker[varType][value]; exists {
		// Return the same indexed variable for the same value
		if index == 0 {
			return fmt.Sprintf("{{.%s}}", varType)
		}
		return fmt.Sprintf("{{.%s_%d}}", varType, index)
	}

	// New unique value - assign it the next available index
	index := len(valueTracker[varType])
	valueTracker[varType][value] = index

	// Return the variable name
	if index == 0 {
		return fmt.Sprintf("{{.%s}}", varType)
	}
	return fmt.Sprintf("{{.%s_%d}}", varType, index)
}

func writeTemplateFile(filePath string, events []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Failed to close file: %v\n", err)
		}
	}()

	// Use a special delimiter between events that won't conflict with log content
	delimiter := "\n---EVENT_DELIMITER---\n"

	for i, event := range events {
		if i > 0 {
			if _, err := file.WriteString(delimiter); err != nil {
				return err
			}
		}
		if _, err := file.WriteString(event); err != nil {
			return err
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
