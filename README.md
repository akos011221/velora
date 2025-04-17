# Velora

Velora is a networking control plane for Microsoft Azure networking. It enables centralized enforcement of networking rules across spoke (workload) VNets, eliminating the need for networking teams to maintain separate GitOps repositories for each spoke. Instead, they can focus solely on defining and enforcing rules centrally, reducing operational overhead.

## Features

Velora allows centralized enforcement of the following rules:

### Routing
- Ensure subnets have a default route pointing to an NVA (either an individual NVA or a load balancer fronting multiple NVAs).
- Restrict direct communication between subnets within the same VNet, requiring traffic to route through the NVA.

### Peering
- Enforce that VNets must be peered exclusively with the hub VPC.

### IP Address Management (IPAM)
- Restrict VNets to use only IP ranges approved by the central networking team for individual subscriptions.