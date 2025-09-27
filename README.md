# Layer 8 Agnostic Distributed Cache (l8inventory)

A high-performance, generic and model-agnostic distributed inventory cache for collected and parsed data, built on the Layer 8 ecosystem. This service provides an advanced distributed caching layer for network devices, Kubernetes resources, and other infrastructure inventory data with enhanced query performance and robust filtering capabilities.

## Overview

l8inventory is a high-performance distributed inventory management system that serves as a cache for collected and parsed infrastructure data. With recent performance optimizations and enhanced filtering capabilities, it is designed to be:

- **Generic**: Works with any data model through Protocol Buffers definitions
- **Model Agnostic**: Supports multiple inventory types (network devices, Kubernetes resources, etc.)
- **Distributed**: Built on the Layer 8 ecosystem for distributed operations with service link support
- **Query-enabled**: Enhanced SQL-like queries with improved performance and advanced filtering
- **High Performance**: Optimized query engine with efficient filtering and fetch operations

## Features

- **Multi-Model Support**: Handle network devices, Kubernetes resources, and custom inventory types
- **Distributed Caching**: Uses Layer 8's distributed cache infrastructure with enhanced sync capabilities
- **Query Interface**: Optimized SQL-like query capabilities with improved performance and advanced filtering
- **Real-time Updates**: Support for POST, PATCH operations with notifications and service linking
- **Primary Key Management**: Configurable primary key attributes for different data types
- **Forwarding Support**: Enhanced forwarding to downstream services with service link architecture
- **Web Service Interface**: Comprehensive REST API endpoints for external integration
- **Performance Optimized**: Recent improvements to query performance, filtering, and fetch operations
- **Advanced Statistics**: Built-in statistics collection for monitoring and performance analysis

## Architecture

The inventory service consists of several key components:

### Core Components

- **InventoryService**: Main service handler implementing the Layer 8 service interface
- **InventoryCenter**: Core inventory management with distributed caching
- **Query Engine**: SQL-like query processing for data retrieval

### Supported Operations

- `POST`: Add new inventory items
- `PATCH`: Update existing inventory items
- `GET`: Query and retrieve inventory data
- `DELETE`: Remove inventory items (interface defined)

## Data Models

The service supports multiple data models defined in Protocol Buffers:

### Network Devices (`proto/inventory.proto`)
- NetworkBox (routers, switches, etc.)
- Equipment information (vendor, series, software versions)
- Physical topology (chassis, slots, modules, ports)
- Logical interfaces

### Kubernetes Resources (`proto/kubernetes.proto`)
Comprehensive Kubernetes object support including:
- Core resources (Pods, Services, Nodes, Namespaces)
- Workloads (Deployments, StatefulSets, DaemonSets, Jobs, CronJobs)
- Networking (Ingress, NetworkPolicy, Endpoints)
- Storage (PersistentVolumes, PersistentVolumeClaims, StorageClass)
- Configuration (ConfigMaps, Secrets)
- RBAC (ServiceAccounts, Roles, ClusterRoles, RoleBindings)

## Installation

### Prerequisites

- Go 1.23.8 or later
- Layer 8 ecosystem dependencies

### Dependencies

The project uses several Layer 8 ecosystem modules:

```go
require (
    github.com/saichler/l8bus v0.0.0-20250919233512-9318eab49cf0
    github.com/saichler/l8pollaris v0.0.0-20250922033843-6e532b25a082
    github.com/saichler/l8reflect v0.0.0-20250919234124-8174370f2112
    github.com/saichler/l8services v0.0.0-20250922141647-36d02dfc3f48
    github.com/saichler/l8srlz v0.0.0-20250919234228-5bb968906922
    github.com/saichler/l8test v0.0.0-20250919233411-36e9f6dc3434
    github.com/saichler/l8types v0.0.0-20250922141405-0e75bd0b244a
    github.com/saichler/l8utils v0.0.0-20250918011151-3bbbe0b545ed
    github.com/saichler/probler v0.0.0-20250922022446-c29b793f9262
    google.golang.org/protobuf v1.36.9
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

// Activate inventory service
serviceName := "inventory"
serviceArea := byte(0)
primaryKey := "id"  // Primary key attribute name
elemType := &YourProtobufMessage{}

vnic.Resources().Services().Activate(
    inventory.ServiceType, 
    serviceName, 
    serviceArea, 
    resources, 
    listener,
    primaryKey, 
    elemType,
)
```

### With Forwarding

```go
// Setup forwarding to another service
forwardInfo := &types.DeviceServiceInfo{
    ServiceName: "downstream-service",
    ServiceArea: 1,
}

vnic.Resources().Services().Activate(
    inventory.ServiceType,
    serviceName,
    serviceArea,
    resources,
    listener,
    primaryKey,
    elemType,
    forwardInfo,  // Optional forwarding configuration
)
```

### Querying Data

```go
// Get inventory center instance
inventoryCenter := inventory.Inventory(resources, serviceName, serviceArea)

// Retrieve by key
item := inventoryCenter.ElementByKey("device-id-123")

// Query with conditions (through service interface)
queryElement := object.New(nil, &types.Query{...})
results := inventoryService.Get(queryElement, vnic)
```

## Configuration

### Primary Key Configuration

Each inventory type requires a primary key attribute to be specified during service activation:

```go
// For network devices using "id" field
primaryKey := "id"

// For Kubernetes resources using "metadata.name"
primaryKey := "metadata.name"

// For custom objects
primaryKey := "yourUniqueField"
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

- `Post(elements, vnic)`: Add new inventory items
- `Patch(elements, vnic)`: Update existing items
- `Get(elements, vnic)`: Query and retrieve data
- `Delete(elements, vnic)`: Remove items
- `WebService()`: Get web service interface

### Web Service

The service automatically provides REST API endpoints through the web service interface, supporting:

- Standard HTTP methods (GET, POST, PATCH, DELETE)
- Query parameter support
- JSON serialization

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
- CRUD operations
- Query functionality
- Forwarding behavior
- Mock service integration

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

## Contributing

This project is part of the Layer 8 ecosystem. Please follow the established patterns and conventions when contributing.

## Related Projects

- [l8services](https://github.com/saichler/l8services) - Layer 8 services framework
- [l8types](https://github.com/saichler/l8types) - Common types and interfaces
- [l8pollaris](https://github.com/saichler/l8pollaris) - Polling and data collection
- [layer8](https://github.com/saichler/layer8) - Core Layer 8 framework