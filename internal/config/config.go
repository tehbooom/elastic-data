package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Connection struct {
		URLs     []string
		APIKey   string
		Username string
		Password string
	}
	Integrations []string
	Defaults     struct {
		Datasets map[string][]string
		Metrics  map[string]string
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	cfg := &Config{}

	// Set default connection settings
	cfg.Connection.URLs = []string{"http://localhost:9200"}

	// Set default integrations
	cfg.Integrations = []string{"system"}

	// Initialize maps
	cfg.Defaults.Datasets = make(map[string][]string)
	cfg.Defaults.Metrics = make(map[string]string)

	return cfg
}

// LoadConfig loads the configuration from file
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Follow XDG Base Directory Specification
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to find home directory: %w", err)
			}
			configHome = filepath.Join(home, ".config")

			// Create ~/.config if it doesn't exist
			if _, err := os.Stat(configHome); os.IsNotExist(err) {
				if err := os.Mkdir(configHome, 0755); err != nil {
					return nil, fmt.Errorf("failed to create config directory: %w", err)
				}
			}
		}

		// Create application config directory if it doesn't exist
		appConfigDir := filepath.Join(configHome, "elastic-data")
		if _, err := os.Stat(appConfigDir); os.IsNotExist(err) {
			if err := os.Mkdir(appConfigDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create application config directory: %w", err)
			}
		}

		// Add standard XDG config paths
		viper.AddConfigPath(appConfigDir)

		// Also check in XDG_CONFIG_DIRS if set
		configDirs := os.Getenv("XDG_CONFIG_DIRS")
		if configDirs == "" {
			configDirs = "/etc/xdg" // Default according to spec
		}

		// Add each directory in XDG_CONFIG_DIRS
		for _, dir := range filepath.SplitList(configDirs) {
			viper.AddConfigPath(filepath.Join(dir, "elastic-data"))
		}

		viper.SetConfigType("yaml")
		viper.SetConfigName("config") // Use 'config.yaml' instead of 'elastic-data.yaml'
	}

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal the config into our struct
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// SaveConfig saves the current configuration to a file
func SaveConfig(config *Config, configPath string) error {
	// Set all config values in Viper
	viper.Set("connection.urls", config.Connection.URLs)
	viper.Set("connection.username", config.Connection.Username)
	viper.Set("connection.password", config.Connection.Password)

	viper.Set("integrations", config.Integrations)

	viper.Set("defaults.datasets", config.Defaults.Datasets)
	viper.Set("defaults.metrics", config.Defaults.Metrics)

	// Set the config file path if provided
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to find home directory: %w", err)
		}

		// Ensure the config directory exists
		configDir := filepath.Join(home, ".config", "elastic-data")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		viper.SetConfigFile(filepath.Join(configDir, "config.yaml"))
	}

	// Write the config file
	if err := viper.WriteConfig(); err != nil {
		// If the config file doesn't exist, write it
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.SafeWriteConfig()
		}
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// setDefaults sets the default values in Viper
func setDefaults() {
	viper.SetDefault("connection.urls", []string{"http://localhost:9200"})

	viper.SetDefault("integrations", []string{"system"})

}

// GetConnectionDetails returns the current connection details
func (c *Config) GetConnectionDetails() ([]string, string, string) {
	return c.Connection.URLs, c.Connection.Username, c.Connection.Password
}

// SetConnectionDetails sets the connection details
func (c *Config) SetConnectionDetails(urls []string, username, password string) {
	c.Connection.URLs = urls
	c.Connection.Username = username
	c.Connection.Password = password
}

// GetIntegrations returns the available integrations
func (c *Config) GetIntegrations() []string {
	return c.Integrations
}

// SetSelectedDatasets sets the selected datasets for each integration
func (c *Config) SetSelectedDatasets(datasets map[string][]string) {
	c.Defaults.Datasets = datasets
}

// GetSelectedDatasets returns the selected datasets
func (c *Config) GetSelectedDatasets() map[string][]string {
	return c.Defaults.Datasets
}

// SetSelectedMetrics sets the selected metrics
func (c *Config) SetSelectedMetrics(metrics map[string]string) {
	c.Defaults.Metrics = metrics
}

// GetSelectedMetrics returns the selected metrics
func (c *Config) GetSelectedMetrics() map[string]string {
	return c.Defaults.Metrics
}
