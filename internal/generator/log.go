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

func ParseLogFile(filePath string, multilineConfig *multiline.Config) ([]*LogTemplate, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Error("Failed to close file:", err)
		}
	}()

	finalReader, err := createReaderPipeline(file, multilineConfig)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to create reader pipeline: %w", err)
	}

	templates, err := parseMessages(finalReader, 100, false)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to parse messages: %w", err)
	}

	for i := range templates {
		templates[i].UserProvided = false
	}

	return templates, nil
}

// ParseUserEvents parses the events for the dataset that the user has provided in the config file
func ParseUserEvents(multilineConfig *multiline.Config, events []string) ([]*LogTemplate, error) {
	log.Debug("parsing user events")
	content := strings.Join(events, "\n")
	log.Debug(fmt.Sprintf("Events: %s", content))

	finalReader, err := createReaderPipelineFromString(content, multilineConfig)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to create reader pipeline: %w", err)
	}

	templates, err := parseMessages(finalReader, 100, true)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to parse messages: %w", err)
	}

	log.Debug(fmt.Sprintf("User provided template: %v", templates[0].Template))

	return templates, nil
}

func parseMessages(reader reader.Reader, maxTemplates int, userProvided bool) ([]*LogTemplate, error) {
	var templates []*LogTemplate
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

		template, err := processLogLine(message.Content, userProvided)
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

func processLogLine(line []byte, userProvided bool) (*LogTemplate, error) {
	template := &LogTemplate{
		Original:     string(line),
		IsJSON:       false,
		Size:         len(line),
		Data:         make(map[string]string),
		UserProvided: userProvided,
	}

	template.AddCommonPatterns()

	if strings.HasPrefix(strings.TrimSpace(string(line)), "{") && json.Valid(line) {
		err := template.ParseJSONEvent()
		if err != nil {
			log.Debug(err)
			return template, fmt.Errorf("failed to parse valid JSON: %w", err)
		}

		return template, nil
	} else {
		if err := template.ParseLogLine(); err != nil {
			log.Debug(err)
			return nil, fmt.Errorf("failed to parse log: %w", err)
		}

		return template, nil
	}
}

func (l *LogTemplate) ParseLogLine() error {
	if l.Data == nil {
		l.Data = make(map[string]string)
	}

	templateStr := l.Original
	for _, pattern := range l.Patterns {
		matches := pattern.Regex.FindAllString(l.Original, -1)
		if len(matches) > 0 {
			l.Data[pattern.Name] = matches[0]
		}

		templateStr = pattern.Regex.ReplaceAllString(templateStr, pattern.Replace)
	}
	log.Debug(fmt.Sprintf("Template is %s", templateStr))

	tmpl, err := template.New("logline").Parse(templateStr)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to create template: %v", err)
	}

	l.Template = tmpl

	return nil
}
