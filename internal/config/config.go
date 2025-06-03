package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Connection   ConfigConnection       `mapstructure:"connection"`
	Integrations map[string]Integration `mapstructure:"integrations"`
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
	Enabled  bool `mapstructure:"enabled"`
	Datasets map[string]Dataset
}

type Dataset struct {
	Enabled               bool `mapstructure:"enabled"`
	Threshold             int  `mapstructure:"threshold"`
	PreserveEventOriginal bool `mapstructure:"preserve_original_event"`
	// EPS or bytes
	Unit string `mapstructure:"unit"`
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

	configExists := false
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, creating new one with defaults")
		} else {
			return nil, "", fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		configExists = true
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if isConfigEmpty(config) {
		setDefaults()
		if err := viper.Unmarshal(config); err != nil {
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

	if config.Integrations != nil && len(config.Integrations) > 0 {
		viper.Set("integrations", config.Integrations)
	}

	if configPath != "" {
		viper.SetConfigFile(filepath.Join(configPath, "config.yaml"))
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
	viper.SetDefault("connection.kibana_endpoints", []string{"http://localhost:5601"})
	viper.SetDefault("connection.elasticsearch_endpoints", []string{"http://localhost:9200"})
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
