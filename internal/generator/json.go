package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/internal/common"
)

func ParseJSONFile(filePath string) ([]*LogTemplate, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	var templates []*LogTemplate

	var JSONLog struct {
		Events []map[string]any
	}

	err = json.Unmarshal(file, &JSONLog)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	for _, event := range JSONLog.Events {
		eventToProcess := extractMessageField(event)

		original, err := json.Marshal(eventToProcess)
		if err != nil {
			log.Debug(err)
			return nil, err
		}

		template := &LogTemplate{
			Original:     string(original),
			IsJSON:       true,
			Size:         len(original),
			Data:         make(map[string]string),
			UserProvided: false,
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

	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(l.Original), &rawData); err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	eventToProcess := extractMessageField(rawData)

	processMap(eventToProcess, l.Data)

	templateJSON, err := json.Marshal(eventToProcess)
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
	log.Debug(eventToProcess)

	tmpl, err := template.New("jsonevent").Parse(string(compactJSON))
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to create template: %v", err)
	}

	l.IsJSON = true
	l.Template = tmpl

	return nil
}

// extractMessageField extracts the inner message field if it exists
func extractMessageField(data map[string]any) map[string]any {
	if messageField, exists := data["message"]; exists {
		if messageStr, ok := messageField.(string); ok {
			var innerEvent map[string]interface{}
			if err := json.Unmarshal([]byte(messageStr), &innerEvent); err == nil {
				return innerEvent
			} else {
				return data
			}
		} else {
			return data
		}
	} else {
		return data
	}
}

func processMap(data map[string]interface{}, extractedData map[string]string) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			placeholder := convertStringValue(key, v, extractedData)
			data[key] = placeholder
		case float64:
			placeholder := convertNumericValue(key, v, extractedData)
			if placeholder != "" {
				data[key] = placeholder
			}
		case int64:
			placeholder := convertNumericValue(key, float64(v), extractedData)
			if placeholder != "" {
				data[key] = placeholder
			}
		case map[string]interface{}:
			processMap(v, extractedData)
		case []interface{}:
			processSlice(key, v, extractedData)
		}
	}
}

func processSlice(key string, slice []interface{}, extractedData map[string]string) {
	for i, value := range slice {
		switch v := value.(type) {
		case string:
			placeholder := convertStringValue(key, v, extractedData)
			slice[i] = placeholder
		case map[string]interface{}:
			processMap(v, extractedData)
		case []interface{}:
			processSlice(key, v, extractedData)
		}
	}
}

func convertNumericValue(k string, value float64, extractedData map[string]string) string {
	key := strings.ToLower(k)

	if strings.Contains(key, "time") || strings.Contains(key, "timestamp") {
		valueStr := strconv.FormatFloat(value, 'f', 0, 64)

		if common.UnixMsRegex.MatchString(valueStr) {
			extractedData["timestamp_unix_ms"] = valueStr
			return "{{.timestamp_unix_ms}}"
		} else if common.UnixSecRegex.MatchString(valueStr) {
			extractedData["timestamp_unix_s"] = valueStr
			return "{{.timestamp_unix_s}}"
		}
	}

	return ""
}

func convertStringValue(k, value string, extractedData map[string]string) string {
	value = strings.TrimSpace(value)

	escapedBytes, _ := json.Marshal(value)
	escapedValue := string(escapedBytes[1 : len(escapedBytes)-1])

	if common.IsEmail(value) {
		extractedData["Emails"] = escapedValue
		return "{{.Emails}}"
	}

	if common.IsURL(value) {
		extractedData["Domains"] = escapedValue
		return "{{.Domains}}"
	}

	if common.IsDomain(value) {
		extractedData["Domains"] = escapedValue
		return "{{.Domains}}"
	}

	if common.IsIP(value) {
		extractedData["IPs"] = escapedValue
		return "{{.IPs}}"
	}

	if common.IsoRegex.MatchString(value) {
		extractedData["timestamp_iso"] = escapedValue
		return "{{.timestamp_iso}}"
	}

	if common.CommonRegex.MatchString(value) {
		extractedData["timestamp_common"] = escapedValue
		return "{{.timestamp_common}}"
	}

	if common.ClfRegex.MatchString(value) {
		extractedData["timestamp_clf"] = escapedValue
		return "{{.timestamp_clf}}"
	}

	if common.SyslogRegex.MatchString(value) {
		extractedData["timestamp_syslog"] = escapedValue
		return "{{.timestamp_syslog}}"
	}

	if common.SnortRegex.MatchString(value) {
		extractedData["timestamp_snort"] = escapedValue
		return "{{.timestamp_snort}}"
	}

	if common.SnortNoYearRegex.MatchString(value) {
		extractedData["timestamp_snort"] = escapedValue
		return "{{.timestamp_snort}}"
	}

	if common.UnixMsRegex.MatchString(value) {
		extractedData["timestamp_unix_ms"] = escapedValue
		return "{{.timestamp_unix_ms}}"
	}

	if common.UnixSecRegex.MatchString(value) {
		extractedData["timestamp_unix_s"] = escapedValue
		return "{{.timestamp_unix_s}}"
	}

	key := strings.ToLower(k)

	if strings.Contains(key, "username") {
		extractedData["Users"] = escapedValue
		return "{{.Users}}"
	} else if strings.Contains(key, "hostname") {
		extractedData["Hosts"] = escapedValue
		return "{{.Hosts}}"
	}
	return escapedValue
}
