package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Connection   ConfigConnection
	Integrations map[string]Integration
}

type ConfigConnection struct {
	KibanaEndpoints        []string
	ElasticsearchEndpoints []string
	APIKey                 string
	Username               string
	Password               string
	Unsafe                 bool
	CACert                 string
	Cert                   string
	Key                    string
}

type Integration struct {
	Enabled  bool
	Datasets map[string]Dataset
}

type Dataset struct {
	Enabled   bool
	Threshold int
	// EPS or bytes
	Unit string
}

func LoadConfig() (*Config, string, error) {
	config := &Config{}
	// Follow XDG Base Directory Specification
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, "", fmt.Errorf("failed to find home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
		if _, err := os.Stat(configHome); os.IsNotExist(err) {
			if err := os.Mkdir(configHome, 0755); err != nil {
				return nil, "", fmt.Errorf("failed to create config directory: %w", err)
			}
		}
	}
	appConfigDir := filepath.Join(configHome, "elastic-data")
	if _, err := os.Stat(appConfigDir); os.IsNotExist(err) {
		if err := os.Mkdir(appConfigDir, 0755); err != nil {
			return nil, "", fmt.Errorf("failed to create application config directory: %w", err)
		}
	}

	viper.AddConfigPath(appConfigDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	configFilePath := filepath.Join(appConfigDir, "config.yaml")

	// First try to read existing config
	configExists := false
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, we'll create a new one with defaults
			log.Println("Config file not found, creating new one with defaults")
		} else {
			// Some other error occurred while reading the config file
			return nil, "", fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		configExists = true
	}

	// Unmarshal the config
	if err := viper.Unmarshal(config); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Check if config is empty and set defaults if needed
	if isConfigEmpty(config) {
		setDefaults()
		// Re-unmarshal after setting defaults
		if err := viper.Unmarshal(config); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal config with defaults: %w", err)
		}
		// Save the config with defaults
		SaveConfig(config, "")
	} else if !configExists {
		// If the config didn't exist but somehow wasn't empty after unmarshaling,
		// still save it to create the file
		SaveConfig(config, "")
	}

	return config, configFilePath, nil
}

// Helper function to check if the config is empty
func isConfigEmpty(config *Config) bool {
	// Check if Endpoints slice is empty or if first endpoint is empty
	kibanaEndpointsEmpty := len(config.Connection.KibanaEndpoints) == 0 ||
		(len(config.Connection.KibanaEndpoints) > 0 && config.Connection.KibanaEndpoints[0] == "")

	esEndpointsEmpty := len(config.Connection.ElasticsearchEndpoints) == 0 ||
		(len(config.Connection.ElasticsearchEndpoints) > 0 && config.Connection.ElasticsearchEndpoints[0] == "")

	endpointsEmpty := true

	if !kibanaEndpointsEmpty && !esEndpointsEmpty {
		endpointsEmpty = false
	}
	// Check if username is empty
	usernameEmpty := config.Connection.Username == ""

	return endpointsEmpty && usernameEmpty
}

// SaveConfig saves the current configuration to a file
func SaveConfig(config *Config, configPath string) error {
	viper.Set("connection.kibana_endpoints", config.Connection.KibanaEndpoints)
	viper.Set("connection.elasticsearch_endpoints", config.Connection.ElasticsearchEndpoints)
	viper.Set("connection.username", config.Connection.Username)
	viper.Set("connection.password", config.Connection.Password)

	if config.Integrations != nil && len(config.Integrations) > 0 {
		viper.Set("integrations", config.Integrations)
	}

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to find home directory: %w", err)
		}

		configDir := filepath.Join(home, ".config", "elastic-data")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		viper.SetConfigFile(filepath.Join(configDir, "config.yaml"))
	}

	if err := viper.WriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.SafeWriteConfig()
		}
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("connection.endpoints", []string{"http://localhost:9200"})
	viper.SetDefault("connection.password", "changeme")
	viper.SetDefault("connection.username", "elastic")
}

// GetConnectionDetails returns the current connection details
func (c *Config) GetConnectionDetails() ([]string, []string, string, string) {
	return c.Connection.KibanaEndpoints, c.Connection.ElasticsearchEndpoints, c.Connection.Username, c.Connection.Password
}

// SetConnectionDetails sets the connection details
func (c *Config) SetConnectionDetails(esEndpoints, kibanaEndpoints []string, username, password string) {
	c.Connection.ElasticsearchEndpoints = esEndpoints
	c.Connection.KibanaEndpoints = kibanaEndpoints
	c.Connection.Username = username
	c.Connection.Password = password
}
