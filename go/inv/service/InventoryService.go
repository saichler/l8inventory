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

package inventory

import (
	"reflect"

	"github.com/saichler/l8pollaris/go/pollaris/targets"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8reflect"
	"github.com/saichler/l8types/go/types/l8services"
	"github.com/saichler/l8utils/go/utils/web"
	"google.golang.org/protobuf/proto"
)

// InventoryService is the Layer 8 service handler that implements the standard
// IServiceHandler interface for distributed inventory management. It wraps an
// InventoryCenter and provides service-level operations including CRUD methods,
// forwarding to downstream services, and web service endpoint registration.
//
// InventoryService supports optional service linking, which allows operations
// to be automatically forwarded to a downstream service (e.g., for persistence
// to a database via an ORM service).
type InventoryService struct {
	// inventoryCenter is the core inventory management engine
	inventoryCenter *InventoryCenter
	// link contains optional forwarding configuration to a downstream service
	link *l8services.L8ServiceLink
	// nic is the virtual network interface for this service
	nic ifs.IVNic
	// sla contains the service level agreement configuration
	sla *ifs.ServiceLevelAgreement
}

// Activate is a convenience function to activate an inventory service for a given
// data type. It retrieves the service name and area from the pollaris links cache
// and configures the service level agreement automatically.
//
// Parameters:
//   - linksId: The links identifier used to look up service name and area from pollaris
//   - serviceItem: A prototype instance of the inventory item type (e.g., &MyProto{})
//   - serviceItemList: A prototype instance of the list type (e.g., &MyProtoList{})
//   - vnic: The virtual network interface to activate the service on
//   - primaryKeys: One or more field names that form the primary key (first is used)
//
// Example:
//
//	inventory.Activate("device-cache", &Device{}, &DeviceList{}, vnic, "Id")
func Activate(linksId string, serviceItem, serviceItemList interface{}, vnic ifs.IVNic, primaryKeys ...string) {
	svName, svArea := targets.Links.Cache(linksId)
	sla := ifs.NewServiceLevelAgreement(&InventoryService{}, svName, svArea, true, nil)
	sla.SetServiceItem(serviceItem)
	sla.SetServiceItemList(serviceItemList)
	sla.SetPrimaryKeys(primaryKeys...)
	vnic.Resources().Services().Activate(sla, vnic)
}

// Activate initializes the InventoryService with the provided service level agreement
// and virtual network interface. This method is called by the Layer 8 service framework
// when the service is activated.
//
// If the SLA contains a service link argument, the service will automatically forward
// operations to the linked downstream service (e.g., for persistence).
//
// Returns nil on success, or an error if initialization fails.
func (this *InventoryService) Activate(sla *ifs.ServiceLevelAgreement, vnic ifs.IVNic) error {
	this.sla = sla
	vnic.Resources().Logger().Info("Activated Inventory on ", sla.ServiceName(), " area ", sla.ServiceArea())
	this.inventoryCenter = newInventoryCenter(sla, vnic)
	if len(sla.Args()) == 1 {
		this.link = sla.Args()[0].(*l8services.L8ServiceLink)
		this.nic = vnic
		this.nic.RegisterServiceLink(this.link)
		vnic.Resources().Logger().Info("Added forwarding to ", this.link.ZsideServiceName, " area ", this.link.ZsideServiceArea)
	}
	vnic.Resources().Registry().Register(&l8api.L8Query{})

	return nil
}

// DeActivate cleans up resources when the service is deactivated. This method is
// called by the Layer 8 service framework when the service is being shut down.
//
// Returns nil on success.
func (this *InventoryService) DeActivate() error {
	this.inventoryCenter = nil
	return nil
}

// Post handles POST requests to add new inventory items. It stores the elements
// in the local cache and optionally forwards the operation to a linked downstream
// service if configured.
//
// Returns an empty elements container of the service item list type.
func (this *InventoryService) Post(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Post(elements)
	if !elements.Notification() && this.link != nil {
		vnic.Leader(this.link.ZsideServiceName, byte(this.link.ZsideServiceArea), ifs.POST, elements)
	}
	return object.New(nil, this.sla.ServiceItemList())
}

// Put handles PUT requests to replace existing inventory items. It replaces the
// elements in the local cache and optionally forwards the operation to a linked
// downstream service if configured.
//
// Returns an empty elements container of the service item list type.
func (this *InventoryService) Put(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Put(elements)
	if !elements.Notification() && this.link != nil {
		vnic.Leader(this.link.ZsideServiceName, byte(this.link.ZsideServiceArea), ifs.PUT, elements)
	}
	return object.New(nil, this.sla.ServiceItemList())
}

// Patch handles PATCH requests to update existing inventory items with partial
// changes. It merges the elements in the local cache and optionally forwards the
// operation to a linked downstream service if configured.
//
// Returns an empty elements container of the service item list type.
func (this *InventoryService) Patch(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Patch(elements)
	if !elements.Notification() && this.link != nil {
		vnic.Leader(this.link.ZsideServiceName, byte(this.link.ZsideServiceArea), ifs.PATCH, elements)
	}
	return object.New(nil, this.sla.ServiceItemList())
}

// Delete handles DELETE requests to remove inventory items. It removes the elements
// from the local cache and optionally forwards the operation to a linked downstream
// service if configured.
//
// Returns an empty elements container of the service item list type.
func (this *InventoryService) Delete(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Delete(elements)
	if !elements.Notification() && this.link != nil {
		vnic.Leader(this.link.ZsideServiceName, byte(this.link.ZsideServiceArea), ifs.DELETE, elements)
	}
	return object.New(nil, this.sla.ServiceItemList())
}

// Get handles GET requests to retrieve inventory items. It supports two modes:
//  1. Single element lookup: If the request contains an element of the service item
//     type, it performs a primary key lookup and returns the matching element.
//  2. Query-based retrieval: If the request contains a query, it executes the query
//     and returns matching elements with pagination and metadata.
//
// Returns the matching elements or an error container if the query fails.
func (this *InventoryService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	vnic.Resources().Logger().Info("Get Executed...")

	result, ok := this.isSingleElement(pb, vnic)
	if ok {
		return result
	}

	query, err := pb.Query(vnic.Resources())
	if err != nil {
		return object.NewError(err.Error())
	}
	elems, stats := this.inventoryCenter.Get(query)
	vnic.Resources().Logger().Info("Get Completed with ", len(elems), " elements for query:")
	return object.NewQueryResult(elems, stats)
}

// GetCopy handles requests for a copy of inventory items. Currently not implemented.
// Returns nil.
func (this *InventoryService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}

// Failed handles failure notifications for operations that could not be completed.
// Currently not implemented. Returns nil.
func (this *InventoryService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}

// TransactionConfig returns the transaction configuration for this service.
// InventoryService implements ITransactionConfig, so it returns itself.
func (this *InventoryService) TransactionConfig() ifs.ITransactionConfig {
	return this
}

// Replication returns whether this service requires replication.
// Currently returns false as inventory caching doesn't require replication.
func (this *InventoryService) Replication() bool {
	return false
}

// ReplicationCount returns the number of replicas required.
// Returns 0 as replication is not enabled.
func (this *InventoryService) ReplicationCount() int {
	return 0
}

// Voter returns whether this service participates in leader election voting.
// Returns true, indicating this service can vote in elections.
func (this *InventoryService) Voter() bool {
	return true
}

// KeyOf extracts a transaction key from the elements. Currently returns an empty
// string as inventory operations don't use transaction keys.
func (this *InventoryService) KeyOf(elements ifs.IElements, resources ifs.IResources) string {
	return ""
}

// WebService returns the web service configuration for REST API endpoints.
// It registers a GET endpoint that accepts L8Query and returns the service item list.
func (this *InventoryService) WebService() ifs.IWebService {
	ws := web.New(this.sla.ServiceName(), this.sla.ServiceArea(), 0)
	ws.AddEndpoint(&l8api.L8Query{}, ifs.GET, this.sla.ServiceItemList().(proto.Message))
	return ws
}

// ItemListType is a utility function that creates a new instance of the list type
// for a given element type. It looks up the list type by appending "List" to the
// element type name.
//
// Example: For an element type "Device", it looks up "DeviceList" in the registry.
//
// Panics if the list type is not found or cannot be instantiated.
func ItemListType(r ifs.IRegistry, any interface{}) proto.Message {
	v := reflect.ValueOf(any).Elem()
	listName := v.Type().Name() + "List"
	info, err := r.Info(listName)
	if err != nil {
		panic(err)
	}
	list, err := info.NewInstance()
	if err != nil {
		panic(err)
	}
	return list.(proto.Message)
}

// isSingleElement checks if the request is for a single element lookup (as opposed
// to a query). If the element type matches the service item type, it performs a
// primary key lookup and returns the result.
//
// Returns (result, true) if single element lookup was performed, (nil, false) otherwise.
func (this *InventoryService) isSingleElement(pb ifs.IElements, vnic ifs.IVNic) (ifs.IElements, bool) {
	ins, ok := pb.Element().(proto.Message)
	if ok {
		aside := reflect.ValueOf(ins).Elem().Type().Name()
		bside := reflect.ValueOf(this.sla.ServiceItem()).Elem().Type().Name()
		if aside == bside {
			rnode, ok1 := vnic.Resources().Introspector().NodeByTypeName(bside)
			if ok1 {
				fields, _ := vnic.Resources().Introspector().Decorators().Fields(rnode, l8reflect.L8DecoratorType_Primary)
				v := reflect.ValueOf(ins).Elem().FieldByName(fields[0])
				gsql := "select * from " + bside + " where " + fields[0] + "=" + v.String()
				q1, err := object.NewQuery(gsql, vnic.Resources())
				if err != nil {
					panic(gsql + " " + err.Error())
				}
				q2, err := q1.Query(vnic.Resources())
				if err != nil {
					panic(gsql + " " + err.Error())
				}
				result, _ := this.inventoryCenter.Get(q2)
				return object.New(nil, result), true
			}
		}
	}
	return nil, false
}
