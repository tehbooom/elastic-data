package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

type Config struct {
	Connection   ConfigConnection       `mapstructure:"connection"`
	Integrations map[string]Integration `mapstructure:"integrations"`
	Replacements Replacements           `mapstructure:"replacements"`
}

type ConfigConnection struct {
	KibanaEndpoints        []string `mapstructure:"kibana_endpoints"`
	ElasticsearchEndpoints []string `mapstructure:"elasticsearch_endpoints"`
	APIKey                 string   `mapstructure:"api_key"`
	Username               string   `mapstructure:"username"`
	Password               string   `mapstructure:"password"`
	Unsafe                 bool     `mapstructure:"unsafe"`
	CACert                 string   `mapstructure:"ca_cert"`
	Cert                   string   `mapstructure:"cert"`
	Key                    string   `mapstructure:"key"`
}

type Integration struct {
	Enabled  bool               `mapstructure:"enabled"`
	Datasets map[string]Dataset `mapstructure:"datasets"`
}

type Dataset struct {
	Enabled               bool `mapstructure:"enabled"`
	Threshold             int  `mapstructure:"threshold"`
	PreserveEventOriginal bool `mapstructure:"preserve_original_event"`
	// EPS or bytes
	Unit   string   `mapstructure:"unit"`
	Events []string `mapstructure:"events"`
}

// LoadConfig returns the config, configuration directory and errors
func LoadConfig() (*Config, string, error) {
	config := &Config{}
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

	viper.AddConfigPath(appConfigDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	configExists := false
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("Config file not found, creating new one with defaults")
		} else {
			log.Debug(err)
			return nil, "", fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		configExists = true
	}

	if err := viper.Unmarshal(config); err != nil {
		log.Debug(err)
		return nil, "", fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if config.Replacements.isEmpty() {
		config.Replacements.setDefaults()
	}

	if isConfigEmpty(config) {
		setDefaults()
		if err := viper.Unmarshal(config); err != nil {
			log.Debug(err)
			return nil, "", fmt.Errorf("failed to unmarshal config with defaults: %w", err)
		}
		SaveConfig(config, "")
	} else if !configExists {
		SaveConfig(config, "")
	}

	return config, appConfigDir, nil
}

// Helper function to check if the config is empty
func isConfigEmpty(config *Config) bool {
	kibanaEndpointsEmpty := len(config.Connection.KibanaEndpoints) == 0 ||
		(len(config.Connection.KibanaEndpoints) > 0 && config.Connection.KibanaEndpoints[0] == "")

	esEndpointsEmpty := len(config.Connection.ElasticsearchEndpoints) == 0 ||
		(len(config.Connection.ElasticsearchEndpoints) > 0 && config.Connection.ElasticsearchEndpoints[0] == "")

	endpointsEmpty := true

	if !kibanaEndpointsEmpty && !esEndpointsEmpty {
		endpointsEmpty = false
	}
	usernameEmpty := config.Connection.Username == ""

	return endpointsEmpty && usernameEmpty
}

// SaveConfig saves the current configuration to a file
func SaveConfig(config *Config, configPath string) error {

	viper.Set("connection.kibana_endpoints", config.Connection.KibanaEndpoints)
	viper.Set("connection.elasticsearch_endpoints", config.Connection.ElasticsearchEndpoints)
	viper.Set("connection.username", config.Connection.Username)
	viper.Set("connection.password", config.Connection.Password)
	viper.Set("connection.api_key", config.Connection.APIKey)
	viper.Set("connection.unsafe", config.Connection.Unsafe)
	viper.Set("connection.ca_cert", config.Connection.CACert)
	viper.Set("connection.cert", config.Connection.Cert)
	viper.Set("connection.key", config.Connection.Key)

	if config.Integrations != nil && len(config.Integrations) > 0 {
		integrations := make(map[string]interface{})
		for name, integration := range config.Integrations {
			intMap := map[string]interface{}{
				"enabled": integration.Enabled,
			}
			if integration.Datasets != nil && len(integration.Datasets) > 0 {
				datasets := make(map[string]interface{})
				for dsName, dataset := range integration.Datasets {
					datasets[dsName] = map[string]interface{}{
						"enabled":                 dataset.Enabled,
						"threshold":               dataset.Threshold,
						"preserve_original_event": dataset.PreserveEventOriginal,
						"unit":                    dataset.Unit,
						"events":                  dataset.Events,
					}
				}
				intMap["datasets"] = datasets
			}
			integrations[name] = intMap
		}
		viper.Set("integrations", integrations)
	}

	var replacementErr error
	validReplacements, replacementErr := config.Replacements.validReplacements()

	if !validReplacements {
		replacementErr = fmt.Errorf("Invalid replacement: %v", replacementErr)
	}

	replacements := map[string]interface{}{
		"ip_addresses": config.Replacements.IPs,
		"usernames":    config.Replacements.Users,
		"domains":      config.Replacements.Domains,
		"hostnames":    config.Replacements.Hosts,
		"emails":       config.Replacements.Emails,
	}

	viper.Set("replacements", replacements)

	if configPath != "" {
		viper.SetConfigFile(filepath.Join(configPath, "config.yaml"))
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Debug(err)
			return fmt.Errorf("failed to find home directory: %w", err)
		}

		configDir := filepath.Join(home, ".config", "elastic-data")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Debug(err)
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		viper.SetConfigFile(filepath.Join(configDir, "config.yaml"))
	}

	if err := viper.WriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.SafeWriteConfig()
		}
		log.Debug(err)
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if replacementErr != nil {
		return replacementErr
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("connection.kibana_endpoints", []string{"http://localhost:5601"})
	viper.SetDefault("connection.elasticsearch_endpoints", []string{"http://localhost:9200"})
	viper.SetDefault("connection.password", "changeme")
	viper.SetDefault("connection.username", "elastic")
}
