package generator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/elastic/beats/v7/libbeat/common/match"
	"github.com/elastic/beats/v7/libbeat/reader"
	"github.com/elastic/beats/v7/libbeat/reader/multiline"
	"github.com/elastic/beats/v7/libbeat/reader/readfile"
	"github.com/elastic/beats/v7/libbeat/reader/readfile/encoding"
	"gopkg.in/yaml.v3"
)

type NestedConfig struct {
	Multiline struct {
		Type         string `yaml:"type,omitempty"`
		Negate       *bool  `yaml:"negate,omitempty"`
		Match        string `yaml:"match,omitempty"`
		MaxLines     *int   `yaml:"max_lines,omitempty"`
		Pattern      string `yaml:"pattern,omitempty"`
		Timeout      string `yaml:"timeout,omitempty"`
		FlushPattern string `yaml:"flush_pattern,omitempty"`
		LinesCount   *int   `yaml:"count_lines,omitempty"`
		SkipNewLine  *bool  `yaml:"skip_newline,omitempty"`
	} `yaml:"multiline,omitempty"`
}

type FlatConfig struct {
	Type         string `yaml:"multiline.type,omitempty"`
	Negate       *bool  `yaml:"multiline.negate,omitempty"`
	Match        string `yaml:"multiline.match,omitempty"`
	MaxLines     *int   `yaml:"multiline.max_lines,omitempty"`
	Pattern      string `yaml:"multiline.pattern,omitempty"`
	Timeout      string `yaml:"multiline.timeout,omitempty"`
	FlushPattern string `yaml:"multiline.flush_pattern,omitempty"`
	LinesCount   *int   `yaml:"multiline.count_lines,omitempty"`
	SkipNewLine  *bool  `yaml:"multiline.skip_newline,omitempty"`
}

func GetMultiLineConfig(datasetPath string) (*multiline.Config, error) {

	handlebarFiles := filepath.Join(datasetPath, "agent", "stream")
	entries, err := os.ReadDir(handlebarFiles)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to read directory %s: %w", handlebarFiles, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".hbs") {
			fullPath := filepath.Join(handlebarFiles, entry.Name())
			yamlContent, err := ExtractYAMLFromHBS(fullPath)
			if err != nil {
				continue
			}

			var nested NestedConfig
			if yaml.Unmarshal(yamlContent, &nested) == nil {
				if nested.Multiline.Pattern != "" {
					config, err := buildMultilineConfigNested(nested)
					if err != nil {
						continue
					}
					return config, nil
				}
			}

			var flat FlatConfig
			if yaml.Unmarshal(yamlContent, &flat) == nil {
				if flat.Pattern != "" {
					log.Debug(fmt.Sprintf("flat multiline config found for %s", handlebarFiles))

					config, err := buildMultilineConfig(flat)
					if err != nil {
						continue
					}
					return config, nil
				}
			}
		}
	}

	manifestPath := filepath.Join(datasetPath, "manifest.yml")
	if _, err := os.Stat(manifestPath); err == nil {
		content, err := os.ReadFile(manifestPath)
		if err == nil {
			var manifest map[string]interface{}
			if yaml.Unmarshal(content, &manifest) == nil {
				if multilineConfig := extractMultilineFromManifest(manifest); multilineConfig != nil {
					log.Debug(fmt.Sprintf("multiline config found for %s", manifestPath))
					log.Debug(multilineConfig)
					return multilineConfig, nil
				}
			}
		}
	}

	return nil, nil
}

func buildMultilineConfig(flat FlatConfig) (*multiline.Config, error) {
	config := &multiline.Config{}

	switch flat.Type {
	case "pattern", "":
		config.Type = 0
	case "count":
		config.Type = 1
	case "while_pattern":
		config.Type = 2
	default:
		log.Debug(fmt.Errorf("unknown multiline type: %s", flat.Type))
		return nil, fmt.Errorf("unknown multiline type: %s", flat.Type)
	}

	if flat.LinesCount != nil {
		config.LinesCount = *flat.LinesCount
	}

	if flat.SkipNewLine != nil {
		config.SkipNewLine = *flat.SkipNewLine
	}

	if flat.Negate != nil {
		config.Negate = *flat.Negate
	} else {
		config.Negate = false
	}

	if config.Type == 0 {
		if flat.Match == "after" || flat.Match == "before" {
			config.Match = flat.Match
		} else {
			return nil, fmt.Errorf("invalid match type '%s': must be 'after' or 'before'", flat.Match)
		}
	}

	if flat.MaxLines != nil {
		config.MaxLines = flat.MaxLines
	} else {
		defaultMaxLines := 500
		config.MaxLines = &defaultMaxLines
	}

	if flat.Pattern != "" {
		matcher, err := match.Compile(flat.Pattern)
		if err != nil {
			log.Debug(err)
			return nil, fmt.Errorf("failed to compile pattern '%s': %w", flat.Pattern, err)
		}
		config.Pattern = &matcher
	} else if config.Type == 0 {
		return nil, fmt.Errorf("pattern is required for pattern mode")
	}

	if flat.Timeout != "" {
		timeout, err := parseTimeout(flat.Timeout)
		if err != nil {
			log.Debug(err)
			return nil, fmt.Errorf("failed to parse timeout '%s': %w", flat.Timeout, err)
		}
		config.Timeout = &timeout
		log.Debug("Timeout set", "timeout", timeout)
	} else {
		defaultTimeout := 5 * time.Second
		config.Timeout = &defaultTimeout
	}

	if flat.FlushPattern != "" {
		flushMatcher, err := match.Compile(flat.FlushPattern)
		if err != nil {
			log.Debug(err)
			return nil, fmt.Errorf("failed to compile flush pattern '%s': %w", flat.FlushPattern, err)
		}
		config.FlushPattern = &flushMatcher
	}

	if err := config.Validate(); err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("invalid multiline config: %w", err)
	}

	return config, nil
}

func parseTimeout(timeoutStr string) (time.Duration, error) {
	timeoutStr = strings.TrimSpace(timeoutStr)

	if duration, err := time.ParseDuration(timeoutStr); err == nil {
		if duration < time.Millisecond {
			return 0, fmt.Errorf("timeout too small: %v (minimum 1ms)", duration)
		}
		if duration > time.Hour {
			return 0, fmt.Errorf("timeout too large: %v (maximum 1h)", duration)
		}
		return duration, nil
	}

	if value, err := strconv.ParseFloat(timeoutStr, 64); err == nil {
		if value <= 0 {
			return 0, fmt.Errorf("timeout must be positive: %v", value)
		}
		if value > 3600 {
			return 0, fmt.Errorf("timeout too large: %vs (maximum 3600s)", value)
		}
		duration := time.Duration(value * float64(time.Second))
		log.Debug("Parsed numeric timeout as seconds", "input", timeoutStr, "duration", duration)
		return duration, nil
	}

	return 0, fmt.Errorf("unable to parse timeout format: %s (expected formats: '5s', '10', '5 seconds')", timeoutStr)
}

func extractMultilineFromManifest(manifest map[string]interface{}) *multiline.Config {
	streams, ok := manifest["streams"].([]interface{})
	if !ok {
		return nil
	}

	for _, stream := range streams {
		streamMap, ok := stream.(map[string]interface{})
		if !ok {
			continue
		}

		parsers, ok := streamMap["parsers"]
		if !ok {
			continue
		}

		if yamlBytes, err := yaml.Marshal(parsers); err == nil {
			var config multiline.Config
			if yaml.Unmarshal(yamlBytes, &config) == nil {
				return &config
			}
		}
	}

	return nil
}

func ExtractYAMLFromHBS(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	handlebarsPattern := regexp.MustCompile(`\{\{[^}]*\}\}`)
	cleaned := handlebarsPattern.ReplaceAllString(string(content), "")

	lines := strings.Split(cleaned, "\n")
	var yamlLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if isYAMLContent(trimmed) {
			yamlLines = append(yamlLines, line)
		}
	}

	return []byte(strings.Join(yamlLines, "\n")), nil
}

func isYAMLContent(line string) bool {
	if regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*:`).MatchString(line) {
		return true
	}

	if strings.HasPrefix(line, "-") {
		return true
	}

	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return true
	}

	return false
}

func createReaderPipeline(file *os.File, multilineConfig *multiline.Config) (reader.Reader, error) {
	log.Debug("Creating reader pipeline", "multiline_enabled", multilineConfig != nil)

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to read file into buffer: %w", err)
	}

	content := buf.Bytes()
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}

	encodingFactory := encoding.Plain
	enc, err := encodingFactory(file)
	if err != nil {
		log.Debug("Failed to create encoding", "error", err)
		return nil, fmt.Errorf("failed to create encoding: %w", err)
	}

	return createReaderFromContent(content, enc, multilineConfig)
}

func createReaderPipelineFromString(content string, multilineConfig *multiline.Config) (reader.Reader, error) {
	log.Debug("Creating reader pipeline from string", "multiline_enabled", multilineConfig != nil)

	if len(content) > 0 && content[len(content)-1] != '\n' {
		content += "\n"
	}

	encodingFactory := encoding.Plain
	enc, err := encodingFactory(strings.NewReader(""))
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to create encoding: %w", err)
	}

	return createReaderFromContent([]byte(content), enc, multilineConfig)
}

func createReaderFromContent(content []byte, enc encoding.Encoding, multilineConfig *multiline.Config) (reader.Reader, error) {
	bufReader := bytes.NewReader(content)
	encodeReader, err := readfile.NewEncodeReader(io.NopCloser(bufReader), readfile.Config{
		Codec:        enc,
		BufferSize:   16 * 1024,
		Terminator:   readfile.LineFeed,
		CollectOnEOF: true,
		MaxBytes:     1024 * 1024,
	})
	if err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("failed to create encode reader: %w", err)
	}

	stripReader := readfile.NewStripNewline(encodeReader, readfile.LineFeed)
	var finalReader reader.Reader = stripReader

	if multilineConfig != nil {
		if patternExistsInContent(content, multilineConfig) {
			log.Debug("Applying multiline configuration",
				"type", multilineConfig.Type,
				"pattern", getPatternString(multilineConfig.Pattern),
				"negate", multilineConfig.Negate,
				"match", multilineConfig.Match)
			finalReader, err = multiline.New(stripReader, "\n", 1024*1024, multilineConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create multiline reader: %w", err)
			}
		} else {
			log.Debug("Multiline pattern not found in content, skipping multiline processing",
				"pattern", getPatternString(multilineConfig.Pattern))
		}
	}

	return readfile.NewLimitReader(finalReader, 1024*1024), nil
}

func patternExistsInContent(content []byte, config *multiline.Config) bool {
	if config == nil || config.Pattern == nil {
		return false
	}

	contentStr := string(content)

	lines := strings.Split(contentStr, "\n")

	for _, line := range lines {
		if config.Pattern.MatchString(line) {
			log.Debug("Pattern found in content", "line", line[:min(len(line), 100)]) // Truncate long lines
			return true
		}
	}

	log.Debug("Pattern not found in content", "pattern", getPatternString(config.Pattern))
	return false
}

func getPatternString(pattern *match.Matcher) string {
	if pattern == nil {
		return "<nil>"
	}
	return pattern.String()
}

func buildMultilineConfigNested(flat NestedConfig) (*multiline.Config, error) {
	config := &multiline.Config{}

	switch flat.Multiline.Type {
	case "pattern", "":
		config.Type = 0
	case "count":
		config.Type = 1
	case "while_pattern":
		config.Type = 2
	default:
		log.Debug(fmt.Errorf("unknown multiline type: %s", flat.Multiline.Type))
		return nil, fmt.Errorf("unknown multiline type: %s", flat.Multiline.Type)
	}

	if flat.Multiline.LinesCount != nil {
		config.LinesCount = *flat.Multiline.LinesCount
	}

	if flat.Multiline.SkipNewLine != nil {
		config.SkipNewLine = *flat.Multiline.SkipNewLine
	}

	if flat.Multiline.Negate != nil {
		config.Negate = *flat.Multiline.Negate
	} else {
		config.Negate = false
	}

	if config.Type == 0 {
		if flat.Multiline.Match == "after" || flat.Multiline.Match == "before" {
			config.Match = flat.Multiline.Match
		} else {
			log.Debug(fmt.Errorf("invalid match type '%s': must be 'after' or 'before'", flat.Multiline.Match))
			return nil, fmt.Errorf("invalid match type '%s': must be 'after' or 'before'", flat.Multiline.Match)
		}
	}

	if flat.Multiline.MaxLines != nil {
		config.MaxLines = flat.Multiline.MaxLines
	} else {
		defaultMaxLines := 500
		config.MaxLines = &defaultMaxLines
	}

	if flat.Multiline.Pattern != "" {
		matcher, err := match.Compile(flat.Multiline.Pattern)
		if err != nil {
			log.Debug(err)
			return nil, fmt.Errorf("failed to compile pattern '%s': %w", flat.Multiline.Pattern, err)
		}
		config.Pattern = &matcher
	} else if config.Type == 0 {
		return nil, fmt.Errorf("pattern is required for pattern mode")
	}

	if flat.Multiline.Timeout != "" {
		timeout, err := parseTimeout(flat.Multiline.Timeout)
		if err != nil {
			log.Debug(fmt.Errorf("failed to parse timeout '%s': %w", flat.Multiline.Timeout, err))
			return nil, fmt.Errorf("failed to parse timeout '%s': %w", flat.Multiline.Timeout, err)
		}
		config.Timeout = &timeout
		log.Debug("Timeout set", "timeout", timeout)
	} else {
		defaultTimeout := 5 * time.Second
		config.Timeout = &defaultTimeout
	}

	if flat.Multiline.FlushPattern != "" {
		flushMatcher, err := match.Compile(flat.Multiline.FlushPattern)
		if err != nil {
			log.Debug(err)
			return nil, fmt.Errorf("failed to compile flush pattern '%s': %w", flat.Multiline.FlushPattern, err)
		}
		config.FlushPattern = &flushMatcher
	}

	if err := config.Validate(); err != nil {
		log.Debug(err)
		return nil, fmt.Errorf("invalid multiline config: %w", err)
	}

	return config, nil
}
