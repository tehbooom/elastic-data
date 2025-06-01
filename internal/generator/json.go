package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"
)

func ParseJSONFile(filePath string) ([]LogTemplate, error) {
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
		log.Debug(err)
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	processMap(data, l.Data)

	templateJSON, err := json.Marshal(data)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to marshal template JSON: %w", err)
	}

	var compactBuffer bytes.Buffer
	if err := json.Compact(&compactBuffer, templateJSON); err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to compact JSON: %w", err)
	}
	compactJSON := compactBuffer.Bytes()

	tmpl, err := template.New("jsonevent").Parse(string(compactJSON))
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to create template: %v", err)
	}

	l.IsJSON = true
	l.Template = tmpl

	return nil
}

func processMap(data map[string]interface{}, extractedData map[string]string) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			placeholder := convertStringValue(v, extractedData)
			data[key] = placeholder
		case map[string]interface{}:
			processMap(v, extractedData)
		case []interface{}:
			processSlice(v, extractedData)
		}
	}
}

func processSlice(slice []interface{}, extractedData map[string]string) {
	for i, value := range slice {
		switch v := value.(type) {
		case string:
			placeholder := convertStringValue(v, extractedData)
			slice[i] = placeholder
		case map[string]interface{}:
			processMap(v, extractedData)
		case []interface{}:
			processSlice(v, extractedData)
		}
	}
}

func convertStringValue(value string, extractedData map[string]string) string {
	value = strings.TrimSpace(value)

	if isEmail(value) {
		extractedData["Emails"] = value
		return "{{.Emails}}"
	}

	if isURL(value) {
		extractedData["Domains"] = value
		return "{{.Domains}}"
	}

	if isDomain(value) {
		extractedData["Domains"] = value
		return "{{.Domains}}"
	}

	if isIP(value) {
		extractedData["IPs"] = value
		return "{{.IPs}}"
	}

	if isTimestampISO(value) {
		extractedData["timestamp_iso"] = value
		return "{{.timestamp_iso}}"
	}

	if isTimestampCommon(value) {
		extractedData["timestamp_common"] = value
		return "{{.timestamp_common}}"
	}

	if isTimestampCLF(value) {
		extractedData["timestamp_clf"] = value
		return "{{.timestamp_clf}}"
	}

	if isTimestampSyslog(value) {
		extractedData["timestamp_syslog"] = value
		return "{{.timestamp_syslog}}"
	}

	return value
}

func isEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	regex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return err == nil && regex.MatchString(s)
}

func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isDomain(s string) bool {
	if strings.Contains(s, "://") || strings.Contains(s, "/") {
		return false
	}

	regex := regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return regex.MatchString(s)
}

func isIP(s string) bool {
	return net.ParseIP(s) != nil
}

func isTimestampISO(s string) bool {
	regex := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z?`)

	matches := regex.FindAllString(s, -1)
	if len(matches) > 0 {
		return true
	}

	return false
}

func isTimestampCommon(s string) bool {
	regex := regexp.MustCompile(`\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}`)

	matches := regex.FindAllString(s, -1)
	if len(matches) > 0 {
		return true
	}

	return false
}

func isTimestampCLF(s string) bool {
	regex := regexp.MustCompile(`\[\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}[^\]]*\]`)

	matches := regex.FindAllString(s, -1)
	if len(matches) > 0 {
		return true
	}

	return false
}

func isTimestampSyslog(s string) bool {
	regex := regexp.MustCompile(`[A-Za-z]{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}`)

	matches := regex.FindAllString(s, -1)
	if len(matches) > 0 {
		return true
	}

	return false
}
