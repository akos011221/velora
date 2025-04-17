package routing

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
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
			if err := e.enforceNVARoutingForVNet(ctx, *vnet.ID, *vnet.Name, hubCFG); err != nil {
				return err
			}
		}
	}
	return nil
}

// enforceNVARoutingForVNet makes sure that the subnets in the VNet have default route pointing to NVA
func (e *Enforcer) enforceNVARoutingForVNet(ctx context.Context, vnetID, vnetName string, hubCFG *config.HubVNetConfig) error {
	// get resource group from vnetID
	parts := extractResourceIDParts(vnetID)
	if parts["resourceGroups"] == "" {
		return fmt.Errorf("invalid VNet ID format: %s", vnetID)
	}
	resourceGroup := parts["resourceGroups"]

	// subnets client for getting the subnets
	subnetsClient, err := e.clientFactory.NewSubnetsClient(ctx)
	if err != nil {
		return err
	}

	pager := subnetsClient.NewListPager(resourceGroup, vnetName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list subnets: %w", err)
		}

		for _, subnet := range page.Value {
			// some subnets can't have route tables, like "GatewaySubnets"
			// but for now it is assumed that spoke VNets don't have that.
			if subnet.Properties.RouteTable == nil {
				fmt.Println("WARNING: No route table found for subnet:", *subnet.Name)
				// TODO: handle cases when there's no route table,
				// as it shouldn't be allowed
				continue
			}

			// get the resource group and the name of the RT
			rtParts := extractResourceIDParts(*subnet.Properties.RouteTable.ID)
			rtResourceGroup := rtParts["resourceGroups"]
			rtName := rtParts["routeTables"]

			// routes client for route operations
			routesClient, err := e.clientFactory.NewRoutesClient(ctx)
			if err != nil {
				return err
			}

			defaultRouteExists := false
			defaultRouteCorrect := false

			// get all routes in the RT
			routePager := routesClient.NewListPager(rtResourceGroup, rtName, nil)
			for routePager.More() {
				routePage, err := routePager.NextPage(ctx)
				if err != nil {
					return fmt.Errorf("failed to list routes: %w", err)
				}

				for _, route := range routePage.Value {
					// is the default route entry found in the route table?
					if route.Properties.AddressPrefix != nil && *route.Properties.AddressPrefix == "0.0.0.0/0" {
						defaultRouteExists = true

						// is the default route entry pointing to the NVA?
						if route.Properties.NextHopType != nil && *route.Properties.NextHopType == armnetwork.RouteNextHopTypeVirtualAppliance {
							if route.Properties.NextHopIPAddress != nil && *route.Properties.NextHopIPAddress == hubCFG.NVANextHop {
								defaultRouteCorrect = true
								break

							}
						}
						break
					}
				}
			}

			// do necessary operations if the default route is missing or
			// not pointing to the NVA
			if !defaultRouteExists || !defaultRouteCorrect {
				if hubCFG.NVANextHop == "" {
					return fmt.Errorf("no NVA IPs defined for hub %s", hubCFG.Name)
				}
				nvaNH := hubCFG.NVANextHop

				// properties for the default route
				defaultRouteName := "DefaultRoute-To-NVA"
				addressPrefix := "0.0.0.0/0"
				nextHopType := armnetwork.RouteNextHopTypeVirtualAppliance

				routeParams := armnetwork.Route{
					Properties: &armnetwork.RoutePropertiesFormat{
						AddressPrefix:    &addressPrefix,
						NextHopType:      &nextHopType,
						NextHopIPAddress: &nvaNH,
					},
				}

				// create or update the default route
				_, err := routesClient.BeginCreateOrUpdate(ctx, rtResourceGroup, rtName, defaultRouteName, routeParams, nil)
				if err != nil {
					return fmt.Errorf("failed to create or update default route for subnet %s: %w", *subnet.Name, err)
				}
			}
		}
	}

	return nil
}

// extractResourceIDParts is a helper to get resource parts from Azure resource ID.
func extractResourceIDParts(resourceID string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(resourceID, "/")

	for i := 1; i < len(parts)-1; i += 2 {
		result[parts[i]] = parts[i+1]
	}

	return result
}
