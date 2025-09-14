package inventory

import (
	"reflect"

	"github.com/saichler/l8pollaris/go/types"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	types2 "github.com/saichler/l8types/go/types"
	"github.com/saichler/l8utils/go/utils/web"
	"github.com/saichler/reflect/go/reflect/introspecting"
	"google.golang.org/protobuf/proto"
)

const (
	ServiceType = "InventoryService"
)

type InventoryService struct {
	inventoryCenter *InventoryCenter
	forwardService  *types.DeviceServiceInfo
	nic             ifs.IVNic
	serviceName     string
	serviceArea     byte
	itemSample      interface{}
	itemSampleList  proto.Message
}

func (this *InventoryService) Activate(serviceName string, serviceArea byte,
	r ifs.IResources, l ifs.IServiceCacheListener, args ...interface{}) error {
	r.Logger().Info("Activated Inventory on ", serviceName, " area ", serviceArea)
	primaryKey := args[0].(string)
	this.inventoryCenter = newInventoryCenter(serviceName, serviceArea, primaryKey, args[1], r, l)
	if len(args) == 3 {
		this.forwardService = args[2].(*types.DeviceServiceInfo)
		this.nic = l.(ifs.IVNic)
		this.nic.RegisterServiceBatch(this.forwardService.ServiceName, byte(this.forwardService.ServiceArea), ifs.M_Leader, 5)
		r.Logger().Info("Added forwarding to ", this.forwardService.ServiceName, " area ", this.forwardService.ServiceArea)
	}
	this.serviceName = serviceName
	this.serviceArea = serviceArea
	this.itemSample = args[1]
	this.itemSampleList = ItemListType(r.Registry(), this.itemSample)
	r.Registry().Register(&types2.Query{})
	return nil
}

func (this *InventoryService) DeActivate() error {
	this.inventoryCenter = nil
	return nil
}

func (this *InventoryService) Post(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Post(elements)
	if !elements.Notification() && this.forwardService != nil {
		go func() {
			vnic.Resources().Logger().Debug("Forawrding Post to ", this.forwardService.ServiceName, " area ", this.forwardService.ServiceArea)
			elem := this.inventoryCenter.ElementByElement(elements.Element())
			this.nic.LeaderRequest(this.forwardService.ServiceName, byte(this.forwardService.ServiceArea), ifs.POST, elem, 30)
		}()
	}
	return object.New(nil, this.itemSampleList)
}

func (this *InventoryService) Put(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Put(elements)
	if !elements.Notification() && this.forwardService != nil {
		go func() {
			vnic.Resources().Logger().Debug("Forawrding Put to ", this.forwardService.ServiceName, " area ", this.forwardService.ServiceArea)
			elem := this.inventoryCenter.ElementByElement(elements.Element())
			this.nic.LeaderRequest(this.forwardService.ServiceName, byte(this.forwardService.ServiceArea), ifs.PUT, elem, 30)
		}()
	}
	return object.New(nil, this.itemSampleList)
}

func (this *InventoryService) Patch(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Patch(elements)
	if !elements.Notification() && this.forwardService != nil {
		go func() {
			vnic.Resources().Logger().Debug("Forawrding Patch to ", this.forwardService.ServiceName, " area ", this.forwardService.ServiceArea)
			elem := this.inventoryCenter.ElementByElement(elements.Element())
			this.nic.LeaderRequest(this.forwardService.ServiceName, byte(this.forwardService.ServiceArea), ifs.PATCH, elem, 30)
		}()
	}
	return object.New(nil, this.itemSampleList)
}

func (this *InventoryService) Delete(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Delete(elements)
	if !elements.Notification() && this.forwardService != nil {
		go func() {
			vnic.Resources().Logger().Debug("Forawrding Delete to ", this.forwardService.ServiceName, " area ", this.forwardService.ServiceArea)
			elem := this.inventoryCenter.ElementByElement(elements.Element())
			this.nic.LeaderRequest(this.forwardService.ServiceName, byte(this.forwardService.ServiceArea), ifs.DELETE, elem, 30)
		}()
	}
	return object.New(nil, this.itemSampleList)
}
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
	elems, pages := this.inventoryCenter.Get(query)
	vnic.Resources().Logger().Info("Get Completed with ", len(elems), " elements for query:")
	return object.NewQueryResult(elems, pages)
}
func (this *InventoryService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *InventoryService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}
func (this *InventoryService) TransactionConfig() ifs.ITransactionConfig {
	return nil
}

func (this *InventoryService) Replication() bool {
	return false
}
func (this *InventoryService) ReplicationCount() int {
	return 0
}
func (this *InventoryService) ConcurrentGets() bool {
	return false
}
func (this *InventoryService) KeyOf(elements ifs.IElements, resources ifs.IResources) string {
	return ""
}

func (this *InventoryService) WebService() ifs.IWebService {
	ws := web.New(this.serviceName, this.serviceArea, nil,
		nil, nil, nil, nil, nil, nil, nil,
		&types2.Query{}, this.itemSampleList)
	return ws
}

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

func (this *InventoryService) isSingleElement(pb ifs.IElements, vnic ifs.IVNic) (ifs.IElements, bool) {
	ins, ok := pb.Element().(proto.Message)
	if ok {
		aside := reflect.ValueOf(ins).Elem().Type().Name()
		bside := reflect.ValueOf(this.itemSample).Elem().Type().Name()
		if aside == bside {
			rnode, ok := vnic.Resources().Introspector().NodeByTypeName(bside)
			if ok {
				fields := introspecting.PrimaryKeyDecorator(rnode).([]string)
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
