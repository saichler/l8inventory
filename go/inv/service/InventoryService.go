package inventory

import (
	"reflect"

	"github.com/saichler/l8pollaris/go/types"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	types2 "github.com/saichler/l8types/go/types"
	"github.com/saichler/l8utils/go/utils/web"
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
	vnic.Resources().Logger().Info("Post Received inventory item...")
	this.inventoryCenter.Add(elements.Element(), elements.Notification())
	if !elements.Notification() && this.forwardService != nil {
		go func() {
			vnic.Resources().Logger().Info("Forawrding Post to ", this.forwardService.ServiceName, " area ",
				this.forwardService.ServiceArea)
			elem := this.inventoryCenter.ElementByElement(elements.Element())
			resp := this.nic.ProximityRequest(this.forwardService.ServiceName, byte(this.forwardService.ServiceArea),
				ifs.POST, elem)
			if resp != nil && resp.Error() != nil {
				panic(resp.Error())
				vnic.Resources().Logger().Error(resp.Error().Error())
			} else {
				vnic.Resources().Logger().Info("Post Finished to ", this.forwardService.ServiceName, " area ",
					this.forwardService.ServiceArea)
			}
		}()
	}
	return nil
}

func (this *InventoryService) Put(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *InventoryService) Patch(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	vnic.Resources().Logger().Info("Patch Received inventory item...")
	this.inventoryCenter.Update(elements.Element(), elements.Notification())
	if !elements.Notification() && this.forwardService != nil {
		go func() {
			vnic.Resources().Logger().Info("Patch Forawrding to ", this.forwardService.ServiceName, " area ",
				this.forwardService.ServiceArea)
			elem := this.inventoryCenter.ElementByElement(elements.Element())
			resp := this.nic.ProximityRequest(this.forwardService.ServiceName,
				byte(this.forwardService.ServiceArea), ifs.PATCH, elem)
			if resp != nil && resp.Error() != nil {
				vnic.Resources().Logger().Error(resp.Error().Error())
			} else {
				vnic.Resources().Logger().Info("Patch Finished to ", this.forwardService.ServiceName, " area ",
					this.forwardService.ServiceArea)
			}
		}()
	}
	return nil
}
func (this *InventoryService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *InventoryService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	vnic.Resources().Logger().Info("Get Executed...")
	query, err := pb.Query(vnic.Resources())
	if err != nil {
		return object.NewError(err.Error())
	}
	elems := this.inventoryCenter.Get(query)
	vnic.Resources().Logger().Info("Get Completed with ", len(elems), " elements for query:")
	return object.New(nil, elems)
}
func (this *InventoryService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *InventoryService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}
func (this *InventoryService) TransactionMethod() ifs.ITransactionMethod {
	return nil
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

/*
func (this *InventoryService) Replication() bool {
	return false
}
func (this *InventoryService) ReplicationCount() int {
	return 0
}
func (this *InventoryService) KeyOf(elements ifs.IElements) string {
	return ""
}*/
