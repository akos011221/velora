package routing

import (
	"context"
	"fmt"

	"github.com/akos011221/velora/internal/azure"
	"github.com/akos011221/velora/internal/config"
)

// Enforcer handles routing enforcement in Azure.
type Enforcer struct {
	clientFactory *azure.ClientFactory
	config        *config.Config
}

// NewEnforcer creates a new routing enforcer instance.
func NewEnforcer(clientFactory *azure.ClientFactory, config *config.Config) *Enforcer {
	return &Enforcer{
		clientFactory: clientFactory,
		config:        config,
	}
}

// EnforceAll applies routing enforcement to all subscriptions.
func (e *Enforcer) EnforceAll(ctx context.Context) error {
	for subID, subCFG := range e.config.Subscriptions {
		// sets the subscription ID for the client factory
		e.clientFactory.SetSubscriptionID(subID)

		/* enforcement logic, if required for the subscription */

		if e.config.Features.RoutingEnforcement {
			// find the relevant hub
			var hubCFG *config.HubVNetConfig
			for _, hub := range e.config.Hubs {
				if hub.Name == subCFG.HubName {
					hubCFG = &hub
					break
				}
			}

			if hubCFG == nil {
				return fmt.Errorf("hub %s not found for subscription %s", subCFG.HubName, subID)
			}

			if subCFG.RequireNVARouting {
				if err := e.enforceNVARouting(ctx, subID, hubCFG); err != nil {
					return fmt.Errorf("failed to enforce NVA routing for subscription %s: %w", subID, err)
				}
			}

			if subCFG.SubnetToSubnetDeny {
				if err := e.enforceSubnetIsolation(ctx, subID); err != nil {
					return fmt.Errorf("failed to enforce subnet isolation for subscription %s: %w", subID, err)
				}
			}
		}
	}
	return nil
}

// enforceNVARouting makes sure that all subnets using the NVAs as the default route next hop.
func (e *Enforcer) enforceNVARouting(ctx context.Context, subscriptionID string, hubCFG *config.HubVNetConfig) error {
	// get all VNets in the subscription
	vnetsClient, err := e.clientFactory.NewVirtualNeworksClient(ctx)
	if err != nil {
		return err
	}

	// list all VNets in the subscription
	pager := vnetsClient.NewListAllPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list virtual networks: %w", err)
		}

		// process each VNet
		for _, vnet := range page.Value {
			if err := e.enforceNVARoutingForVNet(...); err != nil {
				return err
			}
		}
	}
	return nil
}

// TODO: create enforceNVARoutingForVNet for VNet level processing
