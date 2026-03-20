# Layer 8 Agnostic Distributed Cache (l8inventory)

(C) 2025 Sharon Aicler (saichler@gmail.com)

Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

---

A high-performance, generic and model-agnostic distributed inventory cache for collected and parsed data, built on the Layer 8 ecosystem. This service provides an advanced distributed caching layer for network devices, Kubernetes resources, and other infrastructure inventory data with enhanced query performance and robust filtering capabilities.

## Overview

l8inventory is a high-performance distributed inventory management system that serves as a cache for collected and parsed infrastructure data. It is designed to be:

- **Generic**: Works with any data model through Protocol Buffers definitions
- **Model Agnostic**: Supports multiple inventory types (network devices, Kubernetes resources, etc.)
- **Distributed**: Built on the Layer 8 ecosystem for distributed operations with service link support
- **Query-enabled**: Enhanced SQL-like queries with improved performance and advanced filtering
- **High Performance**: Optimized query engine with efficient filtering and fetch operations

## Features

- **Multi-Model Support**: Handle network devices, Kubernetes resources, and custom inventory types
- **Distributed Caching**: Uses Layer 8's distributed cache infrastructure with enhanced sync capabilities
- **Query Interface**: Optimized SQL-like query capabilities with improved performance and advanced filtering
- **Real-time Updates**: Support for POST, PUT, PATCH, DELETE operations with notifications and service linking
- **Primary Key Management**: Configurable primary key attributes for different data types
- **Aggregated Forwarding**: Batched forwarding to downstream persistence services via an Aggregator, which collects operations and flushes them in configurable intervals
- **Web Service Interface**: REST API endpoints for external integration with L8Query-based GET requests
- **Metadata Functions**: Extensible metadata capabilities via `AddMetadata` function for custom computed fields on query results

## Architecture

The inventory service consists of three core components:

### Core Components

- **InventoryService**: Main service handler implementing the Layer 8 `IServiceHandler` interface with full CRUD operations, aggregator-based forwarding, and web service endpoint registration
- **InventoryCenter**: Core inventory management engine wrapping a `DistributedCache` with support for queries, pagination, metadata functions, and primary key-based lookups
- **InventoryUtils**: Utility functions for inventory operations including placeholder element creation via reflection

### Data Flow

```
Upstream Clients (REST API, L8 Services)
        |
        v
  InventoryService (CRUD routing, forwarding)
        |
        v
  InventoryCenter (DistributedCache wrapper)
        |                         |
        v                         v
  Local Cache              Aggregator -> Downstream
  (in-memory)              Persistence Service (ORM)
```

### Supported Operations

- `POST`: Add new inventory items (with optional forwarding)
- `PUT`: Replace existing inventory items (with optional forwarding)
- `PATCH`: Update existing inventory items with partial changes (with optional forwarding)
- `DELETE`: Remove inventory items (with optional forwarding)
- `GET`: Query and retrieve inventory data (supports both single element lookup by primary key and SQL-like query-based retrieval with pagination)

### Forwarding Architecture

When activated with a `linksId`, the service uses an `Aggregator` to batch and forward CRUD operations to a downstream persistence service. The forwarding configuration is resolved at runtime via `targets.Links.Persist(linksId)` and `targets.Links.Cache(linksId)`. Notifications (replicated operations from other nodes) are not forwarded, preventing duplicate writes.

## Installation

### Prerequisites

- Go 1.25.4 or later
- Layer 8 ecosystem dependencies

### Dependencies

The project uses several Layer 8 ecosystem modules:

```go
require (
    github.com/saichler/l8bus v0.0.0-20251229161115-ef9b4833c63e
    github.com/saichler/l8pollaris v0.0.0-20251227151640-2d440aed9da1
    github.com/saichler/l8services v0.0.0-20251227145359-8da06cab6a7c
    github.com/saichler/l8srlz v0.0.0-20251226163123-de32dc54dd4b
    github.com/saichler/l8test v0.0.0-20251227041840-6ef7d1910347
    github.com/saichler/l8types v0.0.0-20251229153716-97f3ce136e2a
    github.com/saichler/l8utils v0.0.0-20251229173454-36d44ec87f63
    github.com/saichler/probler v0.0.0-20251228170831-7008bf334cc4
    google.golang.org/protobuf v1.36.11
)
```

### Build and Test

```bash
# Navigate to the Go directory
cd go/

# Run the test script (includes dependency management and testing)
./test.sh

# Or manually:
go mod init
GOPROXY=direct GOPRIVATE=github.com go mod tidy
go mod vendor
go test -tags=unit -v -coverpkg=./inv/... -coverprofile=cover.html ./...
```

## Usage

### Basic Service Activation

The simplest way to activate an inventory service is via the `Activate` convenience function, which resolves service name and area from the pollaris links cache:

```go
import (
    inventory "github.com/saichler/l8inventory/go/inv/service"
)

// Activate inventory service using the Activate helper
// linksId is used to resolve service name/area from pollaris targets
inventory.Activate(
    "device-cache",  // Links ID for pollaris target lookup
    &Device{},       // Service item prototype
    &DeviceList{},   // Service item list prototype
    vnic,            // Virtual NIC
    "Id",            // Primary key field name
)
```

### Manual Service Activation with Forwarding

For more control, including enabling forwarding to a downstream persistence service:

```go
import (
    inventory "github.com/saichler/l8inventory/go/inv/service"
    "github.com/saichler/l8types/go/ifs"
)

serviceName := "DevCache"
serviceArea := byte(0)
elemType := &Device{}
elemTypeList := &DeviceList{}
primaryKey := "Id"
linksId := "device-persist-link"

// Create and configure SLA
sla := ifs.NewServiceLevelAgreement(&inventory.InventoryService{}, serviceName, serviceArea, true, nil)
sla.SetServiceItem(elemType)
sla.SetServiceItemList(elemTypeList)
sla.SetPrimaryKeys(primaryKey)
sla.SetArgs(linksId)  // Enables aggregated forwarding to persistence service

// Activate the service
vnic.Resources().Services().Activate(sla, vnic)
```

When `linksId` is provided via `SetArgs`, the service creates an `Aggregator` that batches operations and forwards them to the persistence service resolved via `targets.Links.Persist(linksId)`.

### Querying Data

```go
import (
    inventory "github.com/saichler/l8inventory/go/inv/service"
    "github.com/saichler/l8srlz/go/serialize/object"
)

// Get inventory center instance
inventoryCenter := inventory.Inventory(resources, serviceName, serviceArea)

// Retrieve by element (uses primary key)
elem := &Device{Id: "device-123"}
result := inventoryCenter.ElementByElement(elem)

// Query with SQL-like syntax
query, err := object.NewQuery("select * from Device where status=1", resources)
if err != nil {
    // handle error
}
parsedQuery, err := query.Query(resources)
if err != nil {
    // handle error
}
results, metadata := inventoryCenter.Get(parsedQuery)
```

### Adding Custom Metadata

Register custom metadata functions that are evaluated for each element during query operations:

```go
// Add custom metadata function to inventory
inventoryCenter.AddMetadata("statusCount", func(element interface{}) (bool, string) {
    if device, ok := element.(*Device); ok {
        return true, device.Status
    }
    return false, ""
})
```

### Creating Placeholder Elements

```go
// Create an empty element with only the primary key set
inventoryCenter.AddEmpty("device-12345")
```

## Configuration

### Primary Key Configuration

Each inventory type requires a primary key attribute to be specified during service activation:

```go
// For objects using "Id" field
primaryKey := "Id"

// For objects using custom unique field
primaryKey := "UniqueIdentifier"
```

### Service Areas

Services can be partitioned into different areas for organizational purposes:

```go
serviceArea := byte(0)  // Main inventory
serviceArea := byte(1)  // Secondary/staging inventory
```

### Aggregator Configuration

When forwarding is enabled, the `Aggregator` is created with:
- **Batch size**: 5 elements per flush
- **Flush interval**: 30 seconds

These parameters control how operations are batched before being forwarded to the downstream persistence service.

## API Reference

### InventoryService (Layer 8 Service Handler)

| Method | Description |
|--------|-------------|
| `Activate(sla, vnic)` | Initialize the service, set up cache and optional forwarding |
| `DeActivate()` | Cleanup service resources |
| `Post(elements, vnic)` | Add new items to cache, forward if configured |
| `Put(elements, vnic)` | Replace items in cache, forward if configured |
| `Patch(elements, vnic)` | Update existing items, forward if configured |
| `Delete(elements, vnic)` | Remove items, forward if configured |
| `Get(elements, vnic)` | Query/retrieve data (single element or query-based) |
| `WebService()` | Get web service interface for REST API |
| `TransactionConfig()` | Returns transaction config (self) |
| `Voter()` | Returns true (participates in leader election) |
| `Replication()` | Returns false (no replication) |

### Convenience Functions

| Function | Description |
|----------|-------------|
| `Activate(linksId, serviceItem, serviceItemList, vnic, primaryKeys...)` | Activate service from pollaris links |
| `Inventory(resources, serviceName, serviceArea)` | Get InventoryCenter for direct cache access |
| `ItemListType(registry, element)` | Create list type instance from element type |

### InventoryCenter API

| Method | Description |
|--------|-------------|
| `Post(elements)` | Add elements to cache |
| `Put(elements)` | Replace elements in cache |
| `Patch(elements)` | Update elements in cache (partial merge) |
| `Delete(elements)` | Remove elements from cache |
| `Get(query)` | Query elements with pagination and filtering |
| `ElementByElement(elem)` | Retrieve single element by primary key |
| `AddMetadata(name, func)` | Register custom metadata function |
| `AddEmpty(key)` | Create placeholder element with specified primary key |

### Web Service

The service registers a GET endpoint via `WebService()` that accepts `L8Query` and returns the service item list type. REST API integration supports:

- L8Query-based GET requests
- JSON serialization via Protocol Buffers JSON encoding

## Project Structure

```
l8inventory/
├── README.md
├── LICENSE
├── go/
│   ├── go.mod
│   ├── go.sum
│   ├── test.sh                         # Build and test script
│   ├── vendor/                         # Vendored dependencies
│   ├── inv/
│   │   └── service/
│   │       ├── InventoryService.go     # Layer 8 service handler (283 lines)
│   │       ├── InventoryCenter.go      # Core cache engine (183 lines)
│   │       └── InventoryUtils.go       # Helper utilities (41 lines)
│   └── tests/
│       ├── Inventory_test.go           # Integration tests
│       ├── TestInit.go                 # Test topology setup (4 nodes, 3 vnets)
│       ├── TestQuery_test.go           # Query parsing tests
│       └── utils_inventory/
│           └── mock_orm_service.go     # Mock persistence service for forwarding tests
```

## Testing

The project includes integration tests that exercise the full service stack through the Layer 8 virtual network:

```bash
# Run all tests with coverage
cd go/
./test.sh

# Run specific tests
go test -tags=unit -v ./tests/
```

### Test Coverage

The test suite covers:
- Service activation with forwarding configuration
- CRUD operations (POST, PATCH, GET, DELETE)
- Query functionality with SQL-like syntax and pagination
- Single element lookup by primary key
- Forwarding behavior verification via a mock ORM service
- Distributed topology with 4 nodes across 3 virtual networks

## License

This project is licensed under the Apache License 2.0.

(C) 2025 Sharon Aicler (saichler@gmail.com)

## Contributing

This project is part of the Layer 8 ecosystem. Please follow the established patterns and conventions when contributing.

## Related Projects

- [l8services](https://github.com/saichler/l8services) - Layer 8 services framework (distributed cache, SLA)
- [l8types](https://github.com/saichler/l8types) - Common types and interfaces
- [l8pollaris](https://github.com/saichler/l8pollaris) - Polling, targets, and service link management
- [l8bus](https://github.com/saichler/l8bus) - Layer 8 virtual network and message protocol
- [l8srlz](https://github.com/saichler/l8srlz) - Serialization and query object utilities
- [l8utils](https://github.com/saichler/l8utils) - Common utilities (aggregator, web)
- [l8reflect](https://github.com/saichler/l8reflect) - Reflection utilities and decorators
- [l8orm](https://github.com/saichler/l8orm) - ORM persistence service
- [probler](https://github.com/saichler/probler) - Network device types and test utilities
- [layer8](https://github.com/saichler/layer8) - Core Layer 8 framework
