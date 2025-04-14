package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// ConfigEnvVar is the environment variable that specifies the path to the configuration file
	ConfigEnvVar = "VELORA_CONFIG"
	// DefaultConfigPath is the default path to the configuration file
	DefaultConfigPath = "~/.velora/config.json"
	// EnvPrefix is the prefix for environment variables that override config
	EnvPrefix = "VELORA_"
)

// LoadConfig loads the configuration from the specified path or environment variable.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv(ConfigEnvVar)
		if path == "" {
			path = DefaultConfigPath
		}
	}

	cfg, err := loadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from file: %w", err)
	}

	if err := overrideFromEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads configuration from a JSON file
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &cfg, nil
}

// overrideFromEnv overrides configuration values with environment variables
func overrideFromEnv(cfg *Config) error {
	// azure config overrides
	if val := os.Getenv(EnvPrefix + "AZURE_TENANT_ID"); val != "" {
		cfg.Azure.TenantID = val
	}
	if val := os.Getenv(EnvPrefix + "AZURE_CLIENT_ID"); val != "" {
		cfg.Azure.ClientID = val
	}
	if val := os.Getenv(EnvPrefix + "AZURE_SUBSCRIPTION_ID"); val != "" {
		cfg.Azure.SubscriptionID = val
	}

	// feature flag overrides
	if val := os.Getenv(EnvPrefix + "FEATURE_IPAM_ENFORCEMENT"); val != "" {
		cfg.Features.IPAMEnforcement = strings.ToLower(val) == "true"
	}
	if val := os.Getenv(EnvPrefix + "FEATURE_ROUTING_ENFORCEMENT"); val != "" {
		cfg.Features.RoutingEnforcement = strings.ToLower(val) == "true"
	}
	if val := os.Getenv(EnvPrefix + "FEATURE_PEERING_ENFORCEMENT"); val != "" {
		cfg.Features.PeeringEnforcement = strings.ToLower(val) == "true"
	}
	if val := os.Getenv(EnvPrefix + "FEATURE_COMPLIANCE_SCANNING"); val != "" {
		cfg.Features.ComplianceScanning = strings.ToLower(val) == "true"
	}
	if val := os.Getenv(EnvPrefix + "FEATURE_AUTO_REMEDIATION"); val != "" {
		cfg.Features.AutoRemediation = strings.ToLower(val) == "true"
	}

	// api config overrides
	if val := os.Getenv(EnvPrefix + "API_LISTEN_ADDRESS"); val != "" {
		cfg.API.ListenAddress = val
	}
	if val := os.Getenv(EnvPrefix + "API_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.API.Port = port
		} else {
			return fmt.Errorf("invalid API port: %s", val)
		}
	}

	// logging config overrides
	if val := os.Getenv(EnvPrefix + "LOGGING_LEVEL"); val != "" {
		cfg.Logging.Level = val
	}
	if val := os.Getenv(EnvPrefix + "LOGGING_FORMAT"); val != "" {
		cfg.Logging.Format = val
	}

	return nil
}
