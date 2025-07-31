package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"math/rand"
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
	now := time.Now().UTC()

	if l.IsJSON {
		for key := range l.Data {
			var value string
			switch key {
			case "IPs":
				if array, exists := l.DataPools["IPs"]; exists && len(array) > 0 {
					value = array[rand.Intn(len(array))]
				}
			case "Domains":
				if array, exists := l.DataPools["Domains"]; exists && len(array) > 0 {
					value = array[rand.Intn(len(array))]
				}
			case "Emails":
				if array, exists := l.DataPools["Emails"]; exists && len(array) > 0 {
					value = array[rand.Intn(len(array))]
				}
			case "timestamp_iso":
				value = now.Format("2006-01-02T15:04:05.000Z")
			case "timestamp_common":
				value = now.Format("02/Jan/2006:15:04:05")
			case "timestamp_clf_timezone":
				value = now.Format("[02/Jan/2006:15:04:05 -0700]")
			case "timestamp_clf":
				value = now.Format("[02/Jan/2006:15:04:05]")
			case "timestamp_syslog":
				value = now.Format("Jan _2 15:04:05")
			case "timestamp_snort":
				value = now.Format("01/02/06-15:04:05.000000")
			case "timestamp_snort_no_year":
				value = now.Format("01/02-15:04:05.000000")
			case "timestamp_unix_s":
				value = strconv.FormatInt(now.Unix(), 10)
			case "timestamp_unix_ms":
				value = strconv.FormatInt(now.UnixMilli(), 10)
			}
			if value != "" {
				l.Data[key] = value
			}
		}
		return
	}

	for _, pattern := range l.Patterns {
		log.Debug(fmt.Sprintf("Processing pattern: %s", pattern.Name))
		matches := pattern.Regex.FindAllString(l.Original, -1)
		log.Debug(fmt.Sprintf("Matches found for %s: %v", pattern.Name, matches))
		if len(matches) > 0 {
			var value string
			switch pattern.Name {
			case "IPs":
				if array, exists := l.DataPools["IPs"]; exists && len(array) > 0 {
					value = array[rand.Intn(len(array))]
				}
			case "Domains":
				if array, exists := l.DataPools["Domains"]; exists && len(array) > 0 {
					value = array[rand.Intn(len(array))]
				}
			case "Emails":
				if array, exists := l.DataPools["Emails"]; exists && len(array) > 0 {
					value = array[rand.Intn(len(array))]
				}
			case "timestamp_iso":
				value = now.Format("2006-01-02T15:04:05.000Z")
			case "timestamp_common":
				value = now.Format("02/Jan/2006:15:04:05")
			case "timestamp_clf_timezone":
				value = now.Format("[02/Jan/2006:15:04:05 -0700]")
			case "timestamp_clf":
				value = now.Format("[02/Jan/2006:15:04:05 -0700]")
			case "timestamp_syslog":
				value = now.Format("Jan _2 15:04:05")
			case "timestamp_snort":
				value = now.Format("01/02/06-15:04:05.000000")
			case "timestamp_snort_no_year":
				value = now.Format("01/02-15:04:05.000000")
			case "timestamp_unix_s":
				value = strconv.FormatInt(now.Unix(), 10)
			case "timestamp_unix_ms":
				value = strconv.FormatInt(now.UnixMilli(), 10)
			}
			if value != "" {
				l.Data[pattern.Name] = value
			}
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

func LoadTemplatesForDataset(cfgPath, integration, dataset string, cfg *config.Config) ([]*LogTemplate, error) {
	var templates []*LogTemplate

	datasetPath := filepath.Join(cfgPath, "integrations", "packages", integration, "data_stream", dataset)

	basePath := filepath.Join(datasetPath, "_dev", "test", "pipeline")

	multilineConfig, err := GetMultiLineConfig(datasetPath)
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Debug(err)
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

			fileTemplates, err := ParseJSONFile(path)
			if err != nil {
				log.Debug(fmt.Sprintf("Error parsing file %s: %v", path, err))
				return nil
			}

			templates = append(templates, fileTemplates...)
		case ".log":

			fileTemplates, err := ParseLogFile(path, multilineConfig)
			if err != nil {
				log.Debug(fmt.Sprintf("Error parsing file %s: %v", path, err))
				return nil
			}

			templates = append(templates, fileTemplates...)
		}

		return nil
	})

	// add user provided events
	if len(cfg.Integrations[integration].Datasets[dataset].Events) > 0 {
		userTemplates, err := ParseUserEvents(multilineConfig, cfg.Integrations[integration].Datasets[dataset].Events)
		if err != nil {
			log.Debug(fmt.Sprintf("Error user templates: %v", err))
			return nil, err
		}
		templates = append(templates, userTemplates...)
	}

	for i := range templates {
		templates[i].initializeDataPools(&cfg.Replacements)
	}

	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("error walking directory %s: %w", basePath, err)
	}

	log.Debug(fmt.Sprintf("Loaded %d templates for %s:%s", len(templates), integration, dataset))
	return templates, nil
}
