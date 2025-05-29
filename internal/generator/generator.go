package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/log"
)

type LogTemplate struct {
	Original  string
	Template  *template.Template
	IsJSON    bool
	Patterns  []PatternRule
	Size      int
	Data      map[string]string
	DataPools map[string][]string
}

type PatternRule struct {
	Name    string
	Regex   *regexp.Regexp
	Replace string
}

type DataPools struct {
	IPs     []string
	Domains []string
	// Usernames  []string
	// Hostnames  []string
	Email []string
}

func initializeDataPools() map[string][]string {
	dataPools := make(map[string][]string)
	dataPools["IPs"] = []string{
		"192.168.1.100", "10.0.0.1", "172.16.0.1", "203.0.113.1",
		"198.51.100.1", "127.0.0.1", "192.168.0.1", "10.1.1.1",
	}

	dataPools["Domains"] = []string{
		"example.com", "test.org", "mycompany.net", "service.io",
		"app.local", "api.service.com", "web.example.org",
	}
	dataPools["Emails"] = []string{
		"admin@example.com",
		"user@hello.world.com",
		"alice@test.org",
		"root@service.io",
	}
	return dataPools
}

func (l *LogTemplate) AddCommonPatterns() {
	commonPatterns := map[string]*regexp.Regexp{
		"IPs":              regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
		"timestamp_iso":    regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z?`),
		"timestamp_common": regexp.MustCompile(`\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}`),
		"timestamp_clf":    regexp.MustCompile(`\[\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}[^\]]*\]`),
		"timestamp_syslog": regexp.MustCompile(`[A-Za-z]{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}`),
		"Emails":           regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		"Domains":          regexp.MustCompile(`[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
	}

	for name, pattern := range commonPatterns {
		l.AddPattern(name, pattern)
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
	if l.DataPools == nil {
		l.DataPools = initializeDataPools()
	}

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
				now := time.Now()
				value = now.UTC().Format("2006-01-02T15:04:05.000Z")
			case "timestamp_common":
				now := time.Now()
				value = now.Format("02/Jan/2006:15:04:05")
			case "timestamp_clf":
				now := time.Now()
				value = now.Format("[02/Jan/2006:15:04:05 -0700]")
			case "timestamp_syslog":
				now := time.Now()
				value = now.Format("Jan _2 15:04:05")
			}
			if value != "" {
				l.Data[key] = value
			}
		}
		return
	}

	for _, pattern := range l.Patterns {
		matches := pattern.Regex.FindAllString(l.Original, -1)
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
				now := time.Now()
				value = now.UTC().Format("2006-01-02T15:04:05.000Z")
			case "timestamp_common":
				now := time.Now()
				value = now.Format("02/Jan/2006:15:04:05")
			case "timestamp_clf":
				now := time.Now()
				value = now.Format("[02/Jan/2006:15:04:05 -0700]")
			case "timestamp_syslog":
				now := time.Now()
				value = now.Format("Jan _2 15:04:05")
			}
			if value != "" {
				l.Data[pattern.Name] = value
			}
		}
	}
}

func (l *LogTemplate) ExecuteTemplate() (string, error) {
	if l.Template == nil {
		return "", fmt.Errorf("template not parsed yet, call Parse() first")
	}

	var buf bytes.Buffer
	err := l.Template.Execute(&buf, l.Data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

func LoadTemplatesForDataset(cfgPath, integration, dataset string) ([]LogTemplate, error) {
	var templates []LogTemplate

	basePath := filepath.Join(cfgPath, "integrations", "packages", integration, "data_stream", dataset, "_dev", "test", "pipeline")

	err := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.Contains(path, "-expected.json") {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".json" {
			fileTemplates, err := ParseJSONFile(path, integration, dataset)
			if err != nil {
				log.Debug(fmt.Sprintf("Error parsing file %s: %v", path, err))
				return nil
			}
			templates = append(templates, fileTemplates...)
		} else if ext == ".log" {
			fileTemplates, err := ParseLogFile(path, integration, dataset)
			if err != nil {
				log.Debug(fmt.Sprintf("Error parsing file %s: %v", path, err))
				return nil
			}
			templates = append(templates, fileTemplates...)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", basePath, err)
	}

	log.Debug(fmt.Sprintf("Loaded %d templates for %s:%s", len(templates), integration, dataset))
	return templates, nil
}
