# Layer 8 Agnostic Distributed Cache (l8inventory)

© 2025 Sharon Aicler (saichler@gmail.com)

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
- **Forwarding Support**: Enhanced forwarding to downstream services with service link architecture
- **Web Service Interface**: Comprehensive REST API endpoints for external integration
- **Metadata Functions**: Extensible metadata capabilities via `AddMetadata` function
- **Advanced Statistics**: Built-in statistics collection for monitoring and performance analysis

## Architecture

The inventory service consists of several key components:

### Core Components

- **InventoryService**: Main service handler implementing the Layer 8 service interface with full CRUD operations
- **InventoryCenter**: Core inventory management with distributed caching and metadata support
- **InventoryUtils**: Utility functions for inventory operations including empty element creation

### Supported Operations

- `POST`: Add new inventory items
- `PUT`: Replace existing inventory items
- `PATCH`: Update existing inventory items
- `GET`: Query and retrieve inventory data (supports both single element and query-based retrieval)
- `DELETE`: Remove inventory items

## Installation

### Prerequisites

- Go 1.25.4 or later
- Layer 8 ecosystem dependencies

### Dependencies

The project uses several Layer 8 ecosystem modules:

```go
require (
    github.com/saichler/l8bus v0.0.0-20251217195552-7b8e0028d13e
    github.com/saichler/l8pollaris v0.0.0-20251217201759-7262bdbfc272
    github.com/saichler/l8services v0.0.0-20251214040709-a20ffcf7e771
    github.com/saichler/l8srlz v0.0.0-20251212164513-0e6d9cfb21cb
    github.com/saichler/l8test v0.0.0-20251217200904-b3926873b8fc
    github.com/saichler/l8types v0.0.0-20251217202550-21a3478e6096
    github.com/saichler/l8utils v0.0.0-20251217212526-c40717d2420c
    github.com/saichler/probler v0.0.0-20251217022815-9af36d9815f6
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

```go
import (
    inventory "github.com/saichler/l8inventory/go/inv/service"
    "github.com/saichler/l8types/go/ifs"
)

// Define your data types
elemType := &YourProtobufMessage{}
elemTypeList := &YourProtobufMessageList{}
primaryKey := "Id"  // Primary key attribute name

// Activate inventory service using the Activate helper
inventory.Activate(
    "links-id",      // Links ID from pollaris targets
    elemType,        // Service item type
    elemTypeList,    // Service item list type
    vnic,            // Virtual NIC
    primaryKey,      // Primary key field name
)
```

### Manual Service Activation with Forwarding

```go
import (
    inventory "github.com/saichler/l8inventory/go/inv/service"
    "github.com/saichler/l8types/go/ifs"
    "github.com/saichler/l8types/go/types/l8services"
)

// Setup forwarding to another service
forwardInfo := &l8services.L8ServiceLink{
    ZsideServiceName: "downstream-service",
    ZsideServiceArea: 1,
}

serviceName := "inventory"
serviceArea := byte(0)
elemType := &YourProtobufMessage{}
elemTypeList := &YourProtobufMessageList{}
primaryKey := "Id"

// Create and configure SLA
sla := ifs.NewServiceLevelAgreement(&inventory.InventoryService{}, serviceName, serviceArea, true, nil)
sla.SetServiceItem(elemType)
sla.SetServiceItemList(elemTypeList)
sla.SetArgs(forwardInfo)  // Optional forwarding configuration
sla.SetPrimaryKeys(primaryKey)

// Activate the service
vnic.Resources().Services().Activate(sla, vnic)
```

### Querying Data

```go
import (
    inventory "github.com/saichler/l8inventory/go/inv/service"
    "github.com/saichler/l8srlz/go/serialize/object"
)

// Get inventory center instance
inventoryCenter := inventory.Inventory(resources, serviceName, serviceArea)

// Retrieve by element (uses primary key)
elem := &YourProtobufMessage{Id: "device-123"}
result := inventoryCenter.ElementByElement(elem)

// Query with SQL-like syntax
query, err := object.NewQuery("select * from YourType where field=value", resources)
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

```go
// Add custom metadata function to inventory
inventoryCenter.AddMetadata("customField", func(element interface{}) (bool, string) {
    // Process element and return metadata
    if elem, ok := element.(*YourType); ok {
        return true, elem.SomeField
    }
    return false, ""
})
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

## API Reference

### Service Interface

The inventory service implements the standard Layer 8 service interface:

- `Activate(sla, vnic)`: Initialize the service
- `DeActivate()`: Cleanup service resources
- `Post(elements, vnic)`: Add new inventory items
- `Put(elements, vnic)`: Replace inventory items
- `Patch(elements, vnic)`: Update existing items
- `Delete(elements, vnic)`: Remove items
- `Get(elements, vnic)`: Query and retrieve data
- `WebService()`: Get web service interface for REST API

### InventoryCenter API

- `Post(elements)`: Add elements to cache
- `Put(elements)`: Replace elements in cache
- `Patch(elements)`: Update elements in cache
- `Delete(elements)`: Remove elements from cache
- `Get(query)`: Query elements with pagination and filtering
- `ElementByElement(elem)`: Retrieve single element by primary key
- `AddMetadata(name, func)`: Add custom metadata function
- `AddEmpty(key)`: Create empty element with specified key

### Web Service

The service automatically provides REST API endpoints through the web service interface, supporting:

- Standard HTTP methods (GET, POST, PATCH, DELETE)
- Query parameter support via L8Query
- JSON serialization via Protocol Buffers JSON encoding

## Testing

The project includes comprehensive unit tests:

```bash
# Run all tests with coverage
cd go/
./test.sh

# Run specific tests
go test -v ./tests/
```

### Test Coverage

The test suite covers:
- Service activation and deactivation
- CRUD operations (POST, PATCH, GET)
- Query functionality with SQL-like syntax
- Forwarding behavior to downstream services
- Mock service integration

## License

This project is licensed under the Apache License 2.0.

© 2025 Sharon Aicler (saichler@gmail.com)

## Contributing

This project is part of the Layer 8 ecosystem. Please follow the established patterns and conventions when contributing.

## Related Projects

- [l8services](https://github.com/saichler/l8services) - Layer 8 services framework
- [l8types](https://github.com/saichler/l8types) - Common types and interfaces
- [l8pollaris](https://github.com/saichler/l8pollaris) - Polling and data collection
- [l8bus](https://github.com/saichler/l8bus) - Layer 8 message bus
- [l8srlz](https://github.com/saichler/l8srlz) - Serialization utilities
- [l8utils](https://github.com/saichler/l8utils) - Common utilities
- [layer8](https://github.com/saichler/layer8) - Core Layer 8 framework
