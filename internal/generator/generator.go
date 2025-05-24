package generator

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/charmbracelet/log"
)

type LogTemplate struct {
	Original string
	Template *template.Template
	IsJSON   bool
	Patterns []PatternRule
	Data     map[string]string
}

type PatternRule struct {
	Name    string
	Regex   *regexp.Regexp
	Replace string
}

type DataPools struct {
	IPs         []string
	Domains     []string
	Usernames   []string
	Hostnames   []string
	UserAgents  []string
	StatusCodes []string
	Methods     []string
	Paths       []string
	mu          sync.RWMutex
}

func initializeDataPools() *DataPools {
	return &DataPools{
		IPs: []string{
			"192.168.1.100", "10.0.0.1", "172.16.0.1", "203.0.113.1",
			"198.51.100.1", "127.0.0.1", "192.168.0.1", "10.1.1.1",
		},
		Domains: []string{
			"example.com", "test.org", "mycompany.net", "service.io",
			"app.local", "api.service.com", "web.example.org",
		},
		Usernames: []string{
			"alice", "bob", "charlie", "diana", "admin", "user", "guest",
			"service", "system", "root",
		},
		Hostnames: []string{
			"web-01", "api-gateway", "db-primary", "cache-01", "worker-01",
			"load-balancer", "monitor-srv", "backup-system",
		},
		UserAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			"curl/7.68.0", "PostmanRuntime/7.28.4",
		},
		StatusCodes: []string{"200", "201", "400", "401", "403", "404", "500", "502"},
		Methods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		Paths: []string{
			"/api/users", "/login", "/dashboard", "/health", "/search",
			"/api/orders", "/profile", "/admin", "/logout",
		},
	}
}

func (l *LogTemplate) AddCommonPatterns() {
	commonPatterns := map[string]*regexp.Regexp{
		"ip":        regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
		"timestamp": regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`),
		"email":     regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		"url":       regexp.MustCompile(`https?://[^\s"]+`),
		"domain":    regexp.MustCompile(`[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		"UUID":      regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`),
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

func ParseLogFile(filePath, integration, dataset string) ([]LogTemplate, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var templates []LogTemplate

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// TODO: Handle multiline log files
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		template := LogTemplate{
			Original: line,
			Template: nil,
			IsJSON:   false,
		}
		templates = append(templates, template)
	}

	return templates, nil
}

func (l *LogTemplate) ParseLogLine() error {
	templateStr := l.Original
	for _, pattern := range l.Patterns {
		matches := pattern.Regex.FindAllString(l.Original, -1)
		if len(matches) > 0 {
			l.Data[pattern.Name] = matches[0]
		}

		templateStr = pattern.Regex.ReplaceAllString(templateStr, pattern.Replace)
	}

	tmpl, err := template.New("logline").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to create template: %v", err)
	}

	l.Template = tmpl

	return nil
}

func LoadTemplatesForDataset(integration, dataset string) ([]LogTemplate, error) {
	var templates []LogTemplate

	basePath := filepath.Join("data_stream", dataset, "_dev", "test", "pipeline")

	err := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".json" {
		} else if ext == ".log" {
			fileTemplates, err := ParseLogFile(path, integration, dataset)
			if err != nil {
				log.Printf("Error parsing file %s: %v", path, err)
				return nil
			}
			templates = append(templates, fileTemplates...)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", basePath, err)
	}

	log.Printf("Loaded %d templates for %s:%s", len(templates), integration, dataset)
	return templates, nil
}
