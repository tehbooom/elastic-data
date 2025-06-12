package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Connection   ConfigConnection       `yaml:"connection"`
	Integrations map[string]Integration `yaml:"integrations,omitempty"`
	Replacements Replacements           `yaml:"replacements"`
}

type ConfigConnection struct {
	KibanaEndpoints        []string `yaml:"kibana_endpoints"`
	ElasticsearchEndpoints []string `yaml:"elasticsearch_endpoints"`
	APIKey                 string   `yaml:"api_key,omitempty"`
	Username               string   `yaml:"username"`
	Password               string   `yaml:"password"`
	Unsafe                 bool     `yaml:"unsafe,omitempty"`
	CACert                 string   `yaml:"ca_cert,omitempty"`
	Cert                   string   `yaml:"cert,omitempty"`
	Key                    string   `yaml:"key,omitempty"`
}

type Integration struct {
	Enabled  bool               `yaml:"enabled"`
	Datasets map[string]Dataset `yaml:"datasets,omitempty"`
}

type Dataset struct {
	Enabled               bool     `yaml:"enabled"`
	Threshold             int      `yaml:"threshold"`
	PreserveEventOriginal bool     `yaml:"preserve_original_event"`
	Unit                  string   `yaml:"unit"`
	Events                []string `yaml:"events,omitempty"`
}

// LoadConfig returns the config, configuration directory and errors
func LoadConfig() (*Config, string, error) {
	// Follow XDG Base Directory Specification
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Debug(err)
			return nil, "", fmt.Errorf("failed to find home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
		if _, err := os.Stat(configHome); os.IsNotExist(err) {
			if err := os.Mkdir(configHome, 0755); err != nil {
				log.Debug(err)
				return nil, "", fmt.Errorf("failed to create config directory: %w", err)
			}
		}
	}

	appConfigDir := filepath.Join(configHome, "elastic-data")
	if _, err := os.Stat(appConfigDir); os.IsNotExist(err) {
		if err := os.Mkdir(appConfigDir, 0755); err != nil {
			log.Debug(err)
			return nil, "", fmt.Errorf("failed to create application config directory: %w", err)
		}
	}

	configPath := filepath.Join(appConfigDir, "config.yaml")

	config := &Config{}
	configExists := false

	// Try to read existing config file
	if data, err := os.ReadFile(configPath); err != nil {
		if !os.IsNotExist(err) {
			log.Debug(err)
			return nil, "", fmt.Errorf("failed to read config file: %w", err)
		}
		log.Debug("Config file not found, will create new one with defaults")
	} else {
		configExists = true
		if err := yaml.Unmarshal(data, config); err != nil {
			log.Debug(err)
			return nil, "", fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	if config.Replacements.isEmpty() {
		config.Replacements.setDefaults()
	}

	if isConfigEmpty(config) {
		setDefaults(config)
		if err := SaveConfig(config, appConfigDir); err != nil {
			log.Debug(err)
			return nil, "", fmt.Errorf("failed to save config: %w", err)
		}
	} else if !configExists {
		if err := SaveConfig(config, appConfigDir); err != nil {
			log.Debug(err)
			return nil, "", fmt.Errorf("failed to save config: %w", err)
		}
	}

	if err := ValidateConfig(config); err != nil {
		log.Debug(err)
		return nil, "", fmt.Errorf("config validation failed: %w", err)
	}

	return config, appConfigDir, nil
}

func isConfigEmpty(config *Config) bool {
	kibanaEndpointsEmpty := len(config.Connection.KibanaEndpoints) == 0 ||
		(len(config.Connection.KibanaEndpoints) > 0 && config.Connection.KibanaEndpoints[0] == "")
	esEndpointsEmpty := len(config.Connection.ElasticsearchEndpoints) == 0 ||
		(len(config.Connection.ElasticsearchEndpoints) > 0 && config.Connection.ElasticsearchEndpoints[0] == "")
	endpointsEmpty := kibanaEndpointsEmpty && esEndpointsEmpty
	usernameEmpty := config.Connection.Username == ""
	return endpointsEmpty && usernameEmpty
}

func SaveConfig(config *Config, configDir string) error {
	if validReplacements, err := config.Replacements.validReplacements(); !validReplacements {
		return fmt.Errorf("invalid replacement: %v", err)
	}

	var configPath string
	if configDir != "" {
		configPath = filepath.Join(configDir, "config.yaml")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Debug(err)
			return fmt.Errorf("failed to find home directory: %w", err)
		}
		configDirPath := filepath.Join(home, ".config", "elastic-data")
		if err := os.MkdirAll(configDirPath, 0755); err != nil {
			log.Debug(err)
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		configPath = filepath.Join(configDirPath, "config.yaml")
	}

	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Failed to close file: %v\n", err)
		}
	}()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	defer func() {
		if err := encoder.Close(); err != nil {
			log.Fatalf("Failed to close encoder: %v\n", err)
		}
	}()

	if err := encoder.Encode(config); err != nil {
		log.Debug(err)
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

func setDefaults(config *Config) {
	if len(config.Connection.KibanaEndpoints) == 0 {
		config.Connection.KibanaEndpoints = []string{"http://localhost:5601"}
	}
	if len(config.Connection.ElasticsearchEndpoints) == 0 {
		config.Connection.ElasticsearchEndpoints = []string{"http://localhost:9200"}
	}
	if config.Connection.Username == "" {
		config.Connection.Username = "elastic"
	}
	if config.Connection.Password == "" {
		config.Connection.Password = "changeme"
	}
}

// ValidateConfig validates the loaded configuration and returns an error if invalid
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if err := validateEndpoints(config.Connection.KibanaEndpoints, "kibana_endpoints"); err != nil {
		return err
	}

	if err := validateEndpoints(config.Connection.ElasticsearchEndpoints, "elasticsearch_endpoints"); err != nil {
		return err
	}

	hasAPIKey := config.Connection.APIKey != ""
	hasUserPass := config.Connection.Username != "" && config.Connection.Password != ""

	if !hasAPIKey && !hasUserPass {
		return fmt.Errorf("authentication required: must provide either api_key or both username and password")
	}

	if err := validateTLSConfig(&config.Connection); err != nil {
		return err
	}

	if err := validateIntegrations(config.Integrations); err != nil {
		return err
	}

	if validReplacements, err := config.Replacements.validReplacements(); !validReplacements {
		return fmt.Errorf("invalid replacement configuration: %v", err)
	}

	return nil
}

func validateEndpoints(endpoints []string, fieldName string) error {
	if len(endpoints) == 0 {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	for i, endpoint := range endpoints {
		if endpoint == "" {
			return fmt.Errorf("%s[%d] cannot be empty", fieldName, i)
		}

		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			return fmt.Errorf("%s[%d] must be a valid URL starting with http:// or https://: %s", fieldName, i, endpoint)
		}
	}

	return nil
}

// validateTLSConfig ensures TLS configuration is consistent
func validateTLSConfig(conn *ConfigConnection) error {
	hasCACert := conn.CACert != ""
	hasCert := conn.Cert != ""
	hasKey := conn.Key != ""

	if hasCert && !hasKey {
		return fmt.Errorf("cert specified but key is missing - both cert and key are required for client certificate authentication")
	}

	if hasKey && !hasCert {
		return fmt.Errorf("key specified but cert is missing - both cert and key are required for client certificate authentication")
	}

	if hasCACert {
		if _, err := os.Stat(conn.CACert); os.IsNotExist(err) {
			log.Debug(err)
			return fmt.Errorf("ca_cert file does not exist: %s", conn.CACert)
		}
	}

	if hasCert {
		if _, err := os.Stat(conn.Cert); os.IsNotExist(err) {
			log.Debug(err)
			return fmt.Errorf("cert file does not exist: %s", conn.Cert)
		}
	}

	if hasKey {
		if _, err := os.Stat(conn.Key); os.IsNotExist(err) {
			log.Debug(err)
			return fmt.Errorf("key file does not exist: %s", conn.Key)
		}
	}

	return nil
}

// validateIntegrations validates the integrations configuration
func validateIntegrations(integrations map[string]Integration) error {
	for integrationName, integration := range integrations {
		if integrationName == "" {
			return fmt.Errorf("integration name cannot be empty")
		}

		for datasetName, dataset := range integration.Datasets {
			if datasetName == "" {
				return fmt.Errorf("dataset name cannot be empty in integration %s", integrationName)
			}

			if dataset.Enabled && dataset.Threshold <= 0 {
				return fmt.Errorf("threshold must be positive for enabled dataset %s in integration %s", datasetName, integrationName)
			}

			if dataset.Unit != "" {
				if dataset.Unit != "eps" && dataset.Unit != "bytes" {
					return fmt.Errorf("invalid unit %s for dataset %s in integration %s. Valid units are eps or bytes", dataset.Unit, datasetName, integrationName)
				}
			}
		}
	}

	return nil
}
