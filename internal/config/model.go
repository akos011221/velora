package config

import (
	"fmt"
	"net"
)

// Config represents the complete application configuration.
type Config struct {
	Azure         AzureConfig                   `json:"azure"`
	Hubs          []HubVNetConfig               `json:"hubs"`
	Subscriptions map[string]SubscriptionConfig `json:"subscriptions"`
	Features      FeaturesConfig                `json:"features"`
	API           APIConfig                     `json:"api"`
	Logging       LoggingConfig                 `json:"logging"`
}

// AzureConfig represents the Azure-specific configuration.
type AzureConfig struct {
	SubscriptionID   string `json:"subscriptionId"`
	TenantID         string `json:"tenantId"`
	ClientID         string `json:"clientId"`
	ClientSecret     string `json:"clientSecret"`
	UseAzureIdentity bool   `json:"useAzureIdentity"`
}

// HubVNetConfig represents the configuration for a hub VNet.
type HubVNetConfig struct {
	VNetID        string `json:"vnetId"`
	ResourceGroup string `json:"resourceGroup"`
	Name          string `json:"name"`
	NVANextHop    string `json:"nvaNextHop"`
}

// SubscriptionConfig represents the configuration for a subscription.
type SubscriptionConfig struct {
	AllowedCIDRs       []string `json:"allowedCIDRs"`
	HubName            string   `json:"hubName"`
	RequireHubPeering  bool     `json:"requireHubPeering"`
	RequireNVARouting  bool     `json:"requireNVARouting"`
	SubnetToSubnetDeny bool     `json:"subnetToSubnetDeny"`
}

// FeaturesConfig controls enabled features.
type FeaturesConfig struct {
	IPAMEnforcement    bool `json:"ipamEnforcement"`
	RoutingEnforcement bool `json:"routingEnforcement"`
	PeeringEnforcement bool `json:"peeringEnforcement"`
	ComplianceScanning bool `json:"complianceScanning"`
	AutoRemediation    bool `json:"autoRemediation"`
}

// APIConfig represents the API configuration.
type APIConfig struct {
	ListenAddress string `json:"listenAddress"`
	Port          int    `json:"port"`
	TLSEnabled    bool   `json:"tlsEnabled"`
	TLSCertPath   string `json:"tlsCertPath"`
	TLSKeyPath    string `json:"tlsKeyPath"`
}

// LoggingConfig represents the logging configuration.
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputPath string `json:"outputPath"`
}

// Validate performs validation on the configuration.
func (c *Config) Validate() error {
	// validate allowed IP ranges
	for _, subConfig := range c.Subscriptions {
		for _, cidr := range subConfig.AllowedCIDRs {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				return fmt.Errorf("invalid CIDR in allowedIPRanges: %s", cidr)
			}
		}
	}

	// validate hubs
	if len(c.Hubs) == 0 {
		return fmt.Errorf("at least one hub configuration is required")
	}

	// validate NVA IPs
	for _, hub := range c.Hubs {
		if hub.NVANextHop != "" {
			if net.ParseIP(hub.NVANextHop) == nil {
				return fmt.Errorf("invalid NVA IP: %s", hub.NVANextHop)
			}
		}
	}

	return nil
}
