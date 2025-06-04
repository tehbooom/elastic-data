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
	"strconv"
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

		template.AddCommonPatterns(true)
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

		if unixMsRegex.MatchString(valueStr) {
			extractedData["timestamp_unix_ms"] = valueStr
			return "{{.timestamp_unix_ms}}"
		} else if unixSecRegex.MatchString(valueStr) {
			extractedData["timestamp_unix_s"] = valueStr
			return "{{.timestamp_unix_s}}"
		}
	}

	return ""
}

func convertStringValue(k, value string, extractedData map[string]string) string {
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

	if isoRegex.MatchString(value) {
		extractedData["timestamp_iso"] = value
		return "{{.timestamp_iso}}"
	}

	if commonRegex.MatchString(value) {
		extractedData["timestamp_common"] = value
		return "{{.timestamp_common}}"
	}

	if clfRegex.MatchString(value) {
		extractedData["timestamp_clf"] = value
		return "{{.timestamp_clf}}"
	}

	if syslogRegex.MatchString(value) {
		extractedData["timestamp_syslog"] = value
		return "{{.timestamp_syslog}}"
	}
	if snortRegex.MatchString(value) {
		extractedData["timestamp_snort"] = value
		return "{{.timestamp_snort}}"
	}

	key := strings.ToLower(k)

	if strings.Contains(key, "time") || strings.Contains(key, "timestamp") {
		if unixMsRegex.MatchString(value) {
			extractedData["timestamp_unix_ms"] = value
			return "{{.timestamp_unix_ms}}"
		} else if unixSecRegex.MatchString(value) {
			extractedData["timestamp_unix_s"] = value
			return "{{.timestamp_unix_s}}"
		}
	} else if strings.Contains(key, "username") {
		extractedData["Users"] = value
		return "{{.Users}}"
	} else if strings.Contains(key, "hostname") {
		extractedData["Hosts"] = value
		return "{{.Hosts}}"
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
