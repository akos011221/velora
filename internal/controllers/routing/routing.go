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
