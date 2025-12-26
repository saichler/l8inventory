// © 2025 Sharon Aicler (saichler@gmail.com)
//
// Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package inventory provides a high-performance, model-agnostic distributed inventory cache
// for the Layer 8 ecosystem. It enables storage, retrieval, and management of any Protocol
// Buffer-based data model with SQL-like query capabilities, automatic service linking,
// and real-time notifications.
//
// The package consists of two main components:
//   - InventoryService: The Layer 8 service handler that implements the standard service interface
//   - InventoryCenter: The core inventory management engine with distributed caching capabilities
//
// Example usage:
//
//	// Activate an inventory service for a custom data type
//	inventory.Activate("my-links-id", &MyProto{}, &MyProtoList{}, vnic, "Id")
//
//	// Query the inventory
//	center := inventory.Inventory(resources, "serviceName", 0)
//	results, metadata := center.Get(query)
package inventory

import (
	"reflect"

	"github.com/saichler/l8services/go/services/dcache"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
)

// InventoryCenter is the core inventory management engine that provides distributed
// caching capabilities for any Protocol Buffer-based data model. It wraps a Layer 8
// distributed cache and provides CRUD operations with support for queries, metadata
// functions, and primary key-based lookups.
//
// InventoryCenter is created internally by InventoryService during activation and
// can be accessed via the Inventory() function for direct cache operations.
type InventoryCenter struct {
	// elements is the underlying distributed cache that stores the inventory items
	elements ifs.IDistributedCache
	// elementType is the reflect.Type of the inventory item for creating new instances
	elementType reflect.Type
	// primaryKeyAttribute is the name of the field used as the primary key
	primaryKeyAttribute string
	// resources provides access to Layer 8 system resources (logger, registry, etc.)
	resources ifs.IResources
	// serviceName is the registered name of this inventory service
	serviceName string
	// serviceArea is the partition/area identifier for this service instance
	serviceArea byte
	// element is a prototype instance of the inventory item type
	element interface{}
}

// newInventoryCenter creates a new InventoryCenter instance from the service level agreement
// and virtual network interface. It initializes the distributed cache, registers the primary
// key decorator with the introspector, and configures the cache for the specified service.
//
// This is an internal constructor called by InventoryService.Activate().
func newInventoryCenter(sla *ifs.ServiceLevelAgreement, vnic ifs.IVNic) *InventoryCenter {
	this := &InventoryCenter{}
	this.serviceName = sla.ServiceName()
	this.serviceArea = sla.ServiceArea()
	this.element = sla.ServiceItem()
	this.elementType = reflect.ValueOf(this.element).Elem().Type()
	this.resources = vnic.Resources()
	this.primaryKeyAttribute = sla.PrimaryKeys()[0]

	vnic.Resources().Introspector().Decorators().AddPrimaryKeyDecorator(this.element, this.primaryKeyAttribute)

	this.elements = dcache.NewDistributedCache(this.serviceName, this.serviceArea, this.element, nil,
		nil, this.resources)

	return this
}

// Post adds new inventory items to the distributed cache. Each element in the
// IElements collection is posted individually to the cache. The notification flag
// from elements determines whether change notifications are propagated.
//
// This operation is idempotent - posting an element with the same primary key
// will update the existing entry.
func (this *InventoryCenter) Post(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		this.elements.Post(element, elements.Notification())
	}
}

// Put replaces existing inventory items in the distributed cache with the provided
// elements. Each element in the IElements collection is put individually. Unlike Post,
// Put performs a full replacement of the existing entry.
//
// The notification flag determines whether change notifications are propagated.
func (this *InventoryCenter) Put(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		this.elements.Put(element, elements.Notification())
	}
}

// Patch updates existing inventory items with partial changes. Only the non-zero
// fields in the provided elements are merged into the existing entries. This is
// useful for updating specific fields without replacing the entire object.
//
// The notification flag determines whether change notifications are propagated.
func (this *InventoryCenter) Patch(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		this.elements.Patch(element, elements.Notification())
	}
}

// Delete removes inventory items from the distributed cache. Each element in the
// IElements collection is deleted based on its primary key.
//
// The notification flag determines whether deletion notifications are propagated.
func (this *InventoryCenter) Delete(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		this.elements.Delete(element, elements.Notification())
	}
}

// Get retrieves inventory items matching the provided query. It supports pagination
// through the query's Page() and Limit() methods. The query can include SQL-like
// conditions for filtering results.
//
// Returns:
//   - []interface{}: Slice of matching inventory items
//   - *l8api.L8MetaData: Metadata about the query results (total count, etc.)
func (this *InventoryCenter) Get(query ifs.IQuery) ([]interface{}, *l8api.L8MetaData) {
	return this.elements.Fetch(int(query.Page()*query.Limit()), int(query.Limit()), query)
}

// ElementByElement retrieves a single inventory item by matching its primary key.
// The provided element should have its primary key field set; other fields are ignored.
//
// Returns the matching element from the cache, or nil if not found.
func (this *InventoryCenter) ElementByElement(elem interface{}) interface{} {
	resp, _ := this.elements.Get(elem)
	return resp
}

// AddMetadata registers a custom metadata function that will be called for each
// element during query operations. The function receives an element and should
// return (true, value) if it produces metadata, or (false, "") otherwise.
//
// Example:
//
//	center.AddMetadata("status", func(elem interface{}) (bool, string) {
//	    if device, ok := elem.(*Device); ok {
//	        return true, device.Status
//	    }
//	    return false, ""
//	})
func (this *InventoryCenter) AddMetadata(name string, f func(interface{}) (bool, string)) {
	this.elements.AddMetadataFunc(name, f)
}

// Inventory retrieves the InventoryCenter for a registered inventory service.
// This allows direct access to the cache operations without going through the
// service interface.
//
// Parameters:
//   - resource: The Layer 8 resources interface
//   - serviceName: The registered name of the inventory service
//   - serviceArea: The partition/area identifier for the service
//
// Returns nil if the service is not found or not an InventoryService.
func Inventory(resource ifs.IResources, serviceName string, serviceArea byte) *InventoryCenter {
	sp, ok := resource.Services().ServiceHandler(serviceName, serviceArea)
	if !ok {
		return nil
	}
	return (sp.(*InventoryService)).inventoryCenter
}
