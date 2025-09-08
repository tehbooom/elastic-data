package generator

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/internal/common"
	"github.com/tehbooom/elastic-data/internal/config"
)

type LogTemplate struct {
	Original     string
	Template     *template.Template
	IsJSON       bool
	Patterns     []PatternRule
	Size         int
	Data         map[string]string
	DataPools    map[string][]string
	UserProvided bool
}

type PatternRule struct {
	Name    string
	Regex   *regexp.Regexp
	Replace string
}

type DataPools struct {
	IPs     []string
	Domains []string
	Emails  []string
	Users   []string
	Hosts   []string
}

func (l *LogTemplate) initializeDataPools(replacements *config.Replacements) {
	if l.DataPools == nil {
		l.DataPools = make(map[string][]string)
	}

	l.DataPools["IPs"] = replacements.IPs
	l.DataPools["Domains"] = replacements.Domains
	l.DataPools["Emails"] = replacements.Emails
	l.DataPools["Users"] = replacements.Users
	l.DataPools["Hosts"] = replacements.Hosts
}

func (l *LogTemplate) AddCommonPatterns() {
	patterns := []struct {
		name  string
		regex *regexp.Regexp
	}{
		{"timestamp_clf_timezone", common.ClfWithTimezoneRegex},
		{"timestamp_clf", common.ClfRegex},
		{"timestamp_common", common.CommonRegex},
		{"timestamp_iso", common.IsoRegex},
		{"timestamp_syslog", common.SyslogRegex},
		{"timestamp_snort", common.SnortRegex},
		{"timestamp_snort_no_year", common.SnortNoYearRegex},
		{"timestamp_unix_s", common.UnixSecRegex},
		{"timestamp_unix_ms", common.UnixMsRegex},
		{"IPs", regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)},
		{"Emails", regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)},
		{"Domains", regexp.MustCompile(`[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)},
	}

	for _, p := range patterns {
		l.AddPattern(p.name, p.regex)
	}
}

func (l *LogTemplate) AddPattern(name string, pattern *regexp.Regexp) {
	l.Patterns = append(l.Patterns, PatternRule{
		Name:    name,
		Regex:   pattern,
		Replace: fmt.Sprintf("{{.%s}}", name),
	})
}

func (l *LogTemplate) UpdateValues() {
	if l.Data == nil {
		l.Data = make(map[string]string)
	}

	// Extract all template variables from the template string
	variableNames := extractTemplateVariables(l.Original)

	// Generate values for all found variables
	for _, varName := range variableNames {
		value := generateValueForVariable(varName, l.DataPools)
		if value != "" {
			l.Data[varName] = value
		}
	}
}

func (l *LogTemplate) ExecuteTemplate() (string, error) {
	if l.Template == nil {
		log.Debug(fmt.Errorf("template not parsed yet, call Parse() first"))
		return "", fmt.Errorf("template not parsed yet, call Parse() first")
	}

	var buf bytes.Buffer
	err := l.Template.Execute(&buf, l.Data)
	if err != nil {
		log.Debug(err)
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

// LoadPreGeneratedTemplatesForDataset loads templates from pre-generated .tmpl files
func LoadPreGeneratedTemplatesForDataset(integration, dataset string, cfg *config.Config) ([]*LogTemplate, error) {
	templateFilePath := filepath.Join("internal", "integrations", "templates", integration, dataset+".tmpl")

	// Check if template file exists
	if _, err := os.Stat(templateFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file not found: %s", templateFilePath)
	}

	// Read the template file
	templateFile, err := os.ReadFile(templateFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templateFilePath, err)
	}

	// Split into individual templates using delimiter
	templateEvents := strings.Split(string(templateFile), "\n---EVENT_DELIMITER---\n")
	var templates []*LogTemplate

	for i, event := range templateEvents {
		event = strings.TrimSpace(event)
		if event == "" {
			continue // Skip empty events
		}

		// Create LogTemplate from the event
		logTemplate, err := createLogTemplateFromString(event, fmt.Sprintf("%s_%s_%d", integration, dataset, i))
		if err != nil {
			log.Debug(fmt.Sprintf("Warning: failed to create template from line %d: %v", i, err))
			continue
		}

		// Initialize data pools with config replacements
		logTemplate.initializeDataPools(&cfg.Replacements)
		templates = append(templates, logTemplate)
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no valid templates found in file %s", templateFilePath)
	}

	log.Debug(fmt.Sprintf("Loaded %d pre-generated templates for %s:%s", len(templates), integration, dataset))
	return templates, nil
}

// createLogTemplateFromString creates a LogTemplate from a template string
func createLogTemplateFromString(templateStr, name string) (*LogTemplate, error) {
	// Determine if this is a JSON template
	isJSON := strings.HasPrefix(strings.TrimSpace(templateStr), "{")

	// Parse the template string
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Create LogTemplate
	logTemplate := &LogTemplate{
		Original:     templateStr,
		Template:     tmpl,
		IsJSON:       isJSON,
		Size:         len(templateStr),
		Data:         make(map[string]string),
		DataPools:    make(map[string][]string),
		UserProvided: false,
	}

	return logTemplate, nil
}

// extractTemplateVariables finds all {{.VariableName}} patterns in a template string
func extractTemplateVariables(templateStr string) []string {
	re := regexp.MustCompile(`\{\{\.([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(templateStr, -1)

	var variables []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if !seen[varName] {
				variables = append(variables, varName)
				seen[varName] = true
			}
		}
	}

	return variables
}

// generateValueForVariable generates appropriate values for template variables including numbered ones
func generateValueForVariable(varName string, dataPools map[string][]string) string {
	now := time.Now().UTC()

	// Handle numbered variables (e.g., IPs_1, Emails_0)
	baseVar := varName
	if idx := strings.LastIndex(varName, "_"); idx != -1 {
		if _, err := strconv.Atoi(varName[idx+1:]); err == nil {
			baseVar = varName[:idx] // Extract base variable name
		}
	}

	switch baseVar {
	case "IPs":
		if array, exists := dataPools["IPs"]; exists && len(array) > 0 {
			return array[rand.Intn(len(array))]
		}
	case "Domains":
		if array, exists := dataPools["Domains"]; exists && len(array) > 0 {
			return array[rand.Intn(len(array))]
		}
	case "Emails":
		if array, exists := dataPools["Emails"]; exists && len(array) > 0 {
			return array[rand.Intn(len(array))]
		}
	case "Users":
		if array, exists := dataPools["Users"]; exists && len(array) > 0 {
			return array[rand.Intn(len(array))]
		}
	case "Hosts":
		if array, exists := dataPools["Hosts"]; exists && len(array) > 0 {
			return array[rand.Intn(len(array))]
		}
	case "timestamp_iso":
		return now.Format("2006-01-02T15:04:05.000Z")
	case "timestamp_common":
		return now.Format("02/Jan/2006:15:04:05")
	case "timestamp_clf_timezone":
		return now.Format("[02/Jan/2006:15:04:05 -0700]")
	case "timestamp_clf":
		return now.Format("[02/Jan/2006:15:04:05]")
	case "timestamp_syslog":
		return now.Format("Jan _2 15:04:05")
	case "timestamp_snort":
		return now.Format("01/02/06-15:04:05.000000")
	case "timestamp_snort_no_year":
		return now.Format("01/02-15:04:05.000000")
	case "timestamp_unix_s":
		return strconv.FormatInt(now.Unix(), 10)
	case "timestamp_unix_ms":
		return strconv.FormatInt(now.UnixMilli(), 10)
	}

	return ""
}
