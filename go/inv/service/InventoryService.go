package inventory

import (
	"github.com/saichler/l8pollaris/go/types"
	"github.com/saichler/l8types/go/ifs"
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
	return nil
}

func (this *InventoryService) DeActivate() error {
	this.inventoryCenter = nil
	return nil
}

func (this *InventoryService) Post(elements ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	vnic.Resources().Logger().Info("Post Received inventory item...")
	this.inventoryCenter.Add(elements.Element(), elements.Notification())
	if !elements.Notification() {
		go func() {
			if this.forwardService != nil {
				vnic.Resources().Logger().Info("Forawrding Post to ", this.forwardService.ServiceName, " area ",
					this.forwardService.ServiceArea)
				elem := this.inventoryCenter.ElementByElement(elements.Element())
				resp := this.nic.SingleRequest(this.forwardService.ServiceName, byte(this.forwardService.ServiceArea),
					ifs.POST, elem)
				if resp != nil && resp.Error() != nil {
					vnic.Resources().Logger().Error(resp.Error().Error())
				} else {
					vnic.Resources().Logger().Info("Post Finished to ", this.forwardService.ServiceName, " area ",
						this.forwardService.ServiceArea)
				}
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
	if !elements.Notification() {
		go func() {
			if this.forwardService != nil {
				vnic.Resources().Logger().Info("Patch Forawrding to ", this.forwardService.ServiceName, " area ",
					this.forwardService.ServiceArea)
				elem := this.inventoryCenter.ElementByElement(elements.Element())
				resp := this.nic.SingleRequest(this.forwardService.ServiceName,
					byte(this.forwardService.ServiceArea), ifs.POST, elem)
				if resp != nil && resp.Error() != nil {
					vnic.Resources().Logger().Error(resp.Error().Error())
				} else {
					vnic.Resources().Logger().Info("Patch Finished to ", this.forwardService.ServiceName, " area ",
						this.forwardService.ServiceArea)
				}
			}
		}()
	}
	return nil
}
func (this *InventoryService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *InventoryService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
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
	return nil
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
