package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"
	"github.com/elastic/beats/v7/libbeat/reader"
	"github.com/elastic/beats/v7/libbeat/reader/multiline"
)

func ParseLogFile(filePath string, multilineConfig *multiline.Config) ([]LogTemplate, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}
	defer file.Close()

	finalReader, err := createReaderPipeline(file, multilineConfig)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to create reader pipeline: %w", err)
	}

	templates, err := parseMessages(finalReader, 100)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to parse messages: %w", err)
	}

	return templates, nil
}

func parseMessages(reader reader.Reader, maxTemplates int) ([]LogTemplate, error) {
	var templates []LogTemplate
	messageCount := 0

	for {
		if maxTemplates > 0 && messageCount >= maxTemplates {
			log.Debug("Reached maximum template limit", "count", maxTemplates)
			break
		}

		message, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				log.Debug("Reached end of file", "templates_parsed", len(templates))
				break
			}
			log.Debug("Error reading message", "error", err, "message_count", messageCount)
			return nil, fmt.Errorf("failed to read message at position %d: %w", messageCount, err)
		}

		messageCount++

		if len(message.Content) == 0 {
			log.Debug("Skipping empty message", "message_count", messageCount)
			continue
		}

		template, err := processLogLine(message.Content)
		if err != nil {
			log.Debug("Failed to process message", "error", err, "message_count", messageCount)
			continue
		}
		templates = append(templates, template)
	}

	if len(templates) == 0 {
		log.Warn("No valid templates generated from file")
	}

	return templates, nil
}

func processLogLine(line []byte) (LogTemplate, error) {
	template := LogTemplate{
		Original:  string(line),
		IsJSON:    false,
		Size:      len(line),
		Data:      make(map[string]string),
		DataPools: initializeDataPools(),
	}

	template.AddCommonPatterns()

	if strings.HasPrefix(strings.TrimSpace(string(line)), "{") && json.Valid(line) {
		err := template.ParseJSONEvent()
		if err != nil {
			log.Debug(err)
			return template, fmt.Errorf("failed to parse valid JSON: %w", err)
		}

		return template, nil
	}

	template.ParseLogLine()

	return template, nil
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
	}

	tmpl, err := template.New("logline").Parse(templateStr)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to create template: %v", err)
	}

	l.Template = tmpl

	return nil
}
