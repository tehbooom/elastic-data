package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Connection struct {
		Endpoints []string
		APIKey    string
		Username  string
		Password  string
	}
	Integrations []Integration
}

type Integration struct {
	Name     string
	Enabled  bool
	Datasets []Dataset
}

type Dataset struct {
	Name      string
	Enabled   bool
	Threshold struct {
		EPS   *int64 `yaml:"eps,omitempty"`
		Bytes *int64 `yaml:"bytes,omitempty"`
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	cfg := &Config{}

	// Set default connection settings
	cfg.Connection.Endpoints = []string{"http://localhost:9200"}

	return cfg
}

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Follow XDG Base Directory Specification
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to find home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")

		if _, err := os.Stat(configHome); os.IsNotExist(err) {
			if err := os.Mkdir(configHome, 0755); err != nil {
				return nil, fmt.Errorf("failed to create config directory: %w", err)
			}
		}
	}

	appConfigDir := filepath.Join(configHome, "elastic-data")
	if _, err := os.Stat(appConfigDir); os.IsNotExist(err) {
		if err := os.Mkdir(appConfigDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create application config directory: %w", err)
		}
	}

	viper.AddConfigPath(appConfigDir)

	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// SaveConfig saves the current configuration to a file
func SaveConfig(config *Config, configPath string) error {
	viper.Set("connection.endpoints", config.Connection.Endpoints)
	viper.Set("connection.username", config.Connection.Username)
	viper.Set("connection.password", config.Connection.Password)

	viper.Set("integrations", config.Integrations)

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
	viper.SetDefault("connection.urls", []string{"http://localhost:9200"})

	viper.SetDefault("integrations", []string{"system"})
}

// GetConnectionDetails returns the current connection details
func (c *Config) GetConnectionDetails() ([]string, string, string) {
	return c.Connection.Endpoints, c.Connection.Username, c.Connection.Password
}

// SetConnectionDetails sets the connection details
func (c *Config) SetConnectionDetails(endpoints []string, username, password string) {
	c.Connection.Endpoints = endpoints
	c.Connection.Username = username
	c.Connection.Password = password
}
