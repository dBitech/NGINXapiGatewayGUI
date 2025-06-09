package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Port                 string `mapstructure:"port"`
	NginxConfigPath      string `mapstructure:"nginx_config_path"`
	NginxExecutablePath  string `mapstructure:"nginx_executable_path"`
	BackupPath           string `mapstructure:"backup_path"`
	ApiGatewayConfigPath string `mapstructure:"apigateway_config_path"`
	WebAssetsPath        string `mapstructure:"web_assets_path"`
	TemplatesPath        string `mapstructure:"templates_path"`
	DefaultTemplate      string `mapstructure:"default_template"`
	LogLevel             string `mapstructure:"log_level"`
}

// Load initializes configuration from file, environment variables, and defaults
func Load(configPath string) (*Config, error) {
	// Set defaults
	viper.SetDefault("port", "8080")
	viper.SetDefault("nginx_config_path", "/etc/nginx/nginx.conf")
	viper.SetDefault("nginx_executable_path", "nginx")
	viper.SetDefault("backup_path", "./backups")
	viper.SetDefault("apigateway_config_path", "./apigateway-config.yaml")
	viper.SetDefault("web_assets_path", "./web")
	viper.SetDefault("templates_path", "./templates")
	viper.SetDefault("default_template", "nginx.conf.j2")
	viper.SetDefault("log_level", "info")

	// Environment variable support with prefix
	viper.SetEnvPrefix("APIGATEWAY")
	viper.AutomaticEnv()

	// Handle config file
	if configPath != "" {
		// Use specified config file
		viper.SetConfigFile(configPath)
	} else {
		// Search for config in multiple locations
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/.apigateway")
		viper.AddConfigPath("/etc/apigateway")
	}

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Printf("No config file found, using defaults and environment variables")
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	// Unmarshal config
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Resolve relative paths to absolute paths
	if err := cfg.resolveRelativePaths(); err != nil {
		return nil, fmt.Errorf("error resolving paths: %w", err)
	}

	return &cfg, nil
}

// resolveRelativePaths converts relative paths to absolute paths
func (c *Config) resolveRelativePaths() error {
	var err error

	if c.BackupPath, err = filepath.Abs(c.BackupPath); err != nil {
		return fmt.Errorf("error resolving backup path: %w", err)
	}

	if c.ApiGatewayConfigPath, err = filepath.Abs(c.ApiGatewayConfigPath); err != nil {
		return fmt.Errorf("error resolving apigateway config path: %w", err)
	}

	if c.WebAssetsPath, err = filepath.Abs(c.WebAssetsPath); err != nil {
		return fmt.Errorf("error resolving web assets path: %w", err)
	}

	if c.TemplatesPath, err = filepath.Abs(c.TemplatesPath); err != nil {
		return fmt.Errorf("error resolving templates path: %w", err)
	}

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(c.ApiGatewayConfigPath), 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	if err := os.MkdirAll(c.BackupPath, 0755); err != nil {
		return fmt.Errorf("error creating backup directory: %w", err)
	}

	return nil
}

// GetConfigPath returns the path that should be used for the apigateway config
func (c *Config) GetConfigPath() string {
	return c.ApiGatewayConfigPath
}
