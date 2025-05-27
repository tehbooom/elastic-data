package generator

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"net"
	"net/mail"
	"net/url"
	"os"
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
		lineBytes := []byte(line)

		template := LogTemplate{
			Original:  line,
			IsJSON:    false,
			Size:      len(lineBytes),
			Data:      make(map[string]string),
			DataPools: initializeDataPools(),
		}

		template.AddCommonPatterns()
		template.ParseLogLine()

		templates = append(templates, template)
	}

	return templates, nil
}

func (l *LogTemplate) ParseLogLine() error {
	if l.Data == nil {
		l.Data = make(map[string]string)
	}
	if l.DataPools == nil {
		l.DataPools = initializeDataPools()
	}

	templateStr := l.Original
	for _, pattern := range l.Patterns {
		matches := pattern.Regex.FindAllString(l.Original, -1)
		if len(matches) > 0 {
			l.Data[pattern.Name] = matches[0]
		}

		templateStr = pattern.Regex.ReplaceAllString(templateStr, pattern.Replace)
		log.Debug(templateStr)
	}

	tmpl, err := template.New("logline").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to create template: %v", err)
	}

	l.Template = tmpl

	return nil
}

type JSONTemplateConverter struct {
	emailRegex  *regexp.Regexp
	domainRegex *regexp.Regexp
	timeFormats []string
}

func newJSONTemplateConverter() *JSONTemplateConverter {
	return &JSONTemplateConverter{
		emailRegex:  regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		domainRegex: regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		timeFormats: []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05.000000",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		},
	}
}

func ParseJSONFile(filePath, integration, dataset string) ([]LogTemplate, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	var templates []LogTemplate

	var JSONLog struct {
		Events []map[string]any
	}

	err = json.Unmarshal(file, &JSONLog)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	for _, event := range JSONLog.Events {
		original, err := json.Marshal(event)
		if err != nil {
			log.Debug(err)
			return nil, err
		}

		template := LogTemplate{
			Original:  string(original),
			IsJSON:    true,
			Size:      len(original),
			Data:      make(map[string]string),
			DataPools: initializeDataPools(),
		}

		template.AddCommonPatterns()
		err = template.ParseJSONEvent()
		if err != nil {
			log.Debug(err)
			return nil, err
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func (l *LogTemplate) ParseJSONEvent() error {
	if l.Data == nil {
		l.Data = make(map[string]string)
	}
	if l.DataPools == nil {
		l.DataPools = initializeDataPools()
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(l.Original), &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	converter := newJSONTemplateConverter()

	converter.processMap(data, l.Data)

	templateJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal template JSON: %w", err)
	}

	tmpl, err := template.New("jsonevent").Parse(string(templateJSON))
	if err != nil {
		return fmt.Errorf("failed to create template: %v", err)
	}

	l.Template = tmpl
	log.Debug(string(templateJSON))

	return nil
}

func (jc *JSONTemplateConverter) processMap(data map[string]interface{}, extractedData map[string]string) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			placeholder := jc.convertStringValue(v, extractedData)
			data[key] = placeholder
		case map[string]interface{}:
			jc.processMap(v, extractedData)
		case []interface{}:
			jc.processSlice(v, extractedData)
		}
	}
}

func (jc *JSONTemplateConverter) processSlice(slice []interface{}, extractedData map[string]string) {
	for i, value := range slice {
		switch v := value.(type) {
		case string:
			placeholder := jc.convertStringValue(v, extractedData)
			slice[i] = placeholder
		case map[string]interface{}:
			jc.processMap(v, extractedData)
		case []interface{}:
			jc.processSlice(v, extractedData)
		}
	}
}

func (jc *JSONTemplateConverter) convertStringValue(value string, extractedData map[string]string) string {
	value = strings.TrimSpace(value)

	if jc.isEmail(value) {
		extractedData["Emails"] = value
		return "{{.Emails}}"
	}

	if jc.isURL(value) {
		extractedData["Domains"] = value
		return "{{.Domains}}"
	}

	if jc.isDomain(value) {
		extractedData["Domains"] = value
		return "{{.Domains}}"
	}

	if jc.isIP(value) {
		extractedData["IPs"] = value
		return "{{.IPs}}"
	}

	if jc.isTimestampISO(value) {
		extractedData["timestamp_iso"] = value
		return "{{.timestamp_iso}}"
	}

	if jc.isTimestampCommon(value) {
		extractedData["timestamp_common"] = value
		return "{{.timestamp_common}}"
	}

	if jc.isTimestampCLF(value) {
		extractedData["timestamp_clf"] = value
		return "{{.timestamp_clf}}"
	}

	if jc.isTimestampSyslog(value) {
		extractedData["timestamp_syslog"] = value
		return "{{.timestamp_syslog}}"
	}

	return value
}

func (jc *JSONTemplateConverter) isEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil && jc.emailRegex.MatchString(s)
}

func (jc *JSONTemplateConverter) isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (jc *JSONTemplateConverter) isDomain(s string) bool {
	if strings.Contains(s, "://") || strings.Contains(s, "/") {
		return false
	}
	return jc.domainRegex.MatchString(s)
}

func (jc *JSONTemplateConverter) isIP(s string) bool {
	return net.ParseIP(s) != nil
}

func (jc *JSONTemplateConverter) isTimestampISO(s string) bool {
	regex := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z?`)

	matches := regex.FindAllString(s, -1)
	if len(matches) > 0 {
		return true
	}

	return false
}

func (jc *JSONTemplateConverter) isTimestampCommon(s string) bool {
	regex := regexp.MustCompile(`\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}`)

	matches := regex.FindAllString(s, -1)
	if len(matches) > 0 {
		return true
	}

	return false
}

func (jc *JSONTemplateConverter) isTimestampCLF(s string) bool {
	regex := regexp.MustCompile(`\[\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}[^\]]*\]`)

	matches := regex.FindAllString(s, -1)
	if len(matches) > 0 {
		return true
	}

	return false
}

func (jc *JSONTemplateConverter) isTimestampSyslog(s string) bool {
	regex := regexp.MustCompile(`[A-Za-z]{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}`)

	matches := regex.FindAllString(s, -1)
	if len(matches) > 0 {
		return true
	}

	return false
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

		ext := filepath.Ext(path)
		if ext == ".json" {
			fileTemplates, err := ParseJSONFile(path, integration, dataset)
			if err != nil {
				log.Printf("Error parsing file %s: %v", path, err)
				return nil
			}
			templates = append(templates, fileTemplates...)
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
