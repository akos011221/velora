package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"

	"github.com/akos011221/velora/internal/config"
)

// ClientFactory is for creating factory-like clients for Azure services.
type ClientFactory struct {
	cred           azcore.TokenCredential
	clientOptions  *arm.ClientOptions
	subscriptionID string
}

// NewClientFactory creates a new (Azure) ClientFactory instance.
func NewClientFactory(cfg *config.AzureConfig) (*ClientFactory, error) {
	var cred azcore.TokenCredential
	var err error

	// credential is created based on the configuration
	if cfg.UseAzureIdentity {
		// use managed identity or environment credentials
		cred, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create default azure credential: %w", err)
		}
	} else {
		// use client credentials
		cred, err = azidentity.NewClientSecretCredential(
			cfg.TenantID,
			cfg.ClientID,
			cfg.ClientSecret,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create azure client credential: %w", err)
		}
	}

	clientOptions := &arm.ClientOptions{}

	return &ClientFactory{
		cred:           cred,
		clientOptions:  clientOptions,
		subscriptionID: cfg.SubscriptionID,
	}, nil
}

// GetCredential returns the Azure credential.
func (f *ClientFactory) GetCredential() azcore.TokenCredential {
	return f.cred
}

// GetSubscriptionID returns the current Azure subscription ID.
func (f *ClientFactory) GetSubscriptionID() string {
	return f.subscriptionID
}

// SetSubscriptionID sets the Azure subscription ID.
func (f *ClientFactory) SetSubscriptionID(subscriptionID string) {
	f.subscriptionID = subscriptionID
}

// NewVirtualNeworksClient creates a new VNet client.
func (f *ClientFactory) NewVirtualNeworksClient(ctx context.Context) (*armnetwork.VirtualNetworksClient, error) {
	client, err := armnetwork.NewVirtualNetworksClient(f.subscriptionID, f.cred, f.clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure virtual networks client: %w", err)
	}
	return client, nil
}

// NewSubnetsClient creates a new Subnets client.
func (f *ClientFactory) NewSubnetsClient(ctx context.Context) (*armnetwork.SubnetsClient, error) {
	client, err := armnetwork.NewSubnetsClient(f.subscriptionID, f.cred, f.clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure subnets client: %w", err)
	}
	return client, nil
}

// NewRouteTablesClient creates a new Route Tables client.
func (f *ClientFactory) NewRouteTablesClient(ctx context.Context) (*armnetwork.RouteTablesClient, error) {
	client, err := armnetwork.NewRouteTablesClient(f.subscriptionID, f.cred, f.clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure route tables client: %w", err)
	}
	return client, nil
}

// NewRoutesClient creates a new Routes client.
func (f *ClientFactory) NewRoutesClient(ctx context.Context) (*armnetwork.RoutesClient, error) {
	client, err := armnetwork.NewRoutesClient(f.subscriptionID, f.cred, f.clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure routes client: %w", err)
	}
	return client, nil
}

// NewVirtualNetworkPeeringsClient creates a new VNet peerings client.
func (f *ClientFactory) NewVirtualNetworkPeeringsClient(ctx context.Context) (*armnetwork.VirtualNetworkPeeringsClient, error) {
	client, err := armnetwork.NewVirtualNetworkPeeringsClient(f.subscriptionID, f.cred, f.clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure virtual network peerings client: %w", err)
	}
	return client, nil
}
