package inventory

import (
	"reflect"

	"github.com/saichler/l8services/go/services/dcache"
	"github.com/saichler/l8services/go/services/recovery"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8reflect"
	"github.com/saichler/l8types/go/types/l8services"
	"github.com/saichler/l8utils/go/utils/web"
	"google.golang.org/protobuf/proto"
)

const (
	ServiceType = "InventoryService"
)

type InventoryService struct {
	inventoryCenter *InventoryCenter
	link            *l8services.L8ServiceLink
	nic             ifs.IVNic
	serviceName     string
	serviceArea     byte
	itemSample      interface{}
	itemSampleList  proto.Message
}

func (this *InventoryService) Activate(sla *ifs.ServiceLevelAgreement, vnic ifs.IVNic) error {
	vnic.Resources().Logger().Info("Activated Inventory on ", sla.ServiceName(), " area ", sla.ServiceArea())
	this.inventoryCenter = newInventoryCenter(sla, vnic)
	if len(sla.Args()) == 1 {
		this.link = sla.Args()[0].(*l8services.L8ServiceLink)
		this.nic = vnic
		this.nic.RegisterServiceLink(this.link)
		vnic.Resources().Logger().Info("Added forwarding to ", this.link.ZsideServiceName, " area ", this.link.ZsideServiceArea)
	}
	this.serviceName = sla.ServiceName()
	this.serviceArea = sla.ServiceArea()
	this.itemSample = sla.ServiceItem()
	this.itemSampleList = sla.ServiceItemList().(proto.Message)
	vnic.Resources().Registry().Register(&l8api.L8Query{})

	c := this.inventoryCenter.elements.(*dcache.DCache).Cache()

	recovery.RecoveryCheck(this.serviceName, this.serviceArea, c, vnic)

	return nil
}

func (this *InventoryService) DeActivate() error {
	this.inventoryCenter = nil
	return nil
}

func (this *InventoryService) Post(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Post(elements)
	if !elements.Notification() && this.link != nil {
		vnic.Leader(this.link.ZsideServiceName, byte(this.link.ZsideServiceArea), ifs.POST, elements)
	}
	return object.New(nil, this.itemSampleList)
}

func (this *InventoryService) Put(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Put(elements)
	if !elements.Notification() && this.link != nil {
		vnic.Leader(this.link.ZsideServiceName, byte(this.link.ZsideServiceArea), ifs.PUT, elements)
	}
	return object.New(nil, this.itemSampleList)
}

func (this *InventoryService) Patch(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Patch(elements)
	if !elements.Notification() && this.link != nil {
		vnic.Leader(this.link.ZsideServiceName, byte(this.link.ZsideServiceArea), ifs.PATCH, elements)
	}
	return object.New(nil, this.itemSampleList)
}

func (this *InventoryService) Delete(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.inventoryCenter.Delete(elements)
	if !elements.Notification() && this.link != nil {
		vnic.Leader(this.link.ZsideServiceName, byte(this.link.ZsideServiceArea), ifs.DELETE, elements)
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
	elems, stats := this.inventoryCenter.Get(query)
	vnic.Resources().Logger().Info("Get Completed with ", len(elems), " elements for query:")
	return object.NewQueryResult(elems, stats)
}

func (this *InventoryService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *InventoryService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}
func (this *InventoryService) TransactionConfig() ifs.ITransactionConfig {
	return this
}

func (this *InventoryService) Replication() bool {
	return false
}
func (this *InventoryService) ReplicationCount() int {
	return 0
}
func (this *InventoryService) Voter() bool {
	return true
}
func (this *InventoryService) KeyOf(elements ifs.IElements, resources ifs.IResources) string {
	return ""
}

func (this *InventoryService) WebService() ifs.IWebService {
	ws := web.New(this.serviceName, this.serviceArea, nil,
		nil, nil, nil, nil, nil, nil, nil,
		&l8api.L8Query{}, this.itemSampleList)
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
