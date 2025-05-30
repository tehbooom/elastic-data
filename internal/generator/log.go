package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"
)

func ParseLogFile(filePath, integration, dataset string) ([]LogTemplate, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var templates []LogTemplate

	scanner := bufio.NewScanner(file)
	// TODO: Handle multiline log files
	// Need to look in two places */data-stream/*/agent/stream/*.hbs and */data_stream/*/manifest.yml looking for - multiline:
	// if multiline we need to scan() differently
	for scanner.Scan() {
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

		if bytes.HasPrefix(lineBytes, []byte("{")) {
			template.IsJSON = true
			err = template.ParseJSONEvent()
			if err != nil {
				log.Debug(err)
				return nil, err
			}
		} else {
			template.ParseLogLine()
		}

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
	}

	tmpl, err := template.New("logline").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to create template: %v", err)
	}

	l.Template = tmpl

	return nil
}
