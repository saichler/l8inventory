package inventory

import (
	"reflect"

	"github.com/saichler/l8reflect/go/reflect/introspecting"
	"github.com/saichler/l8services/go/services/dcache"
	"github.com/saichler/l8types/go/ifs"
)

type InventoryCenter struct {
	elements            ifs.IDistributedCache
	elementType         reflect.Type
	primaryKeyAttribute string
	resources           ifs.IResources
	serviceName         string
	serviceArea         byte
	element             interface{}
}

func newInventoryCenter(sla *ifs.ServiceLevelAgreement, vnic ifs.IVNic) *InventoryCenter {
	this := &InventoryCenter{}
	this.serviceName = sla.ServiceName()
	this.serviceArea = sla.ServiceArea()
	this.element = sla.ServiceItem()
	this.elementType = reflect.ValueOf(this.element).Elem().Type()
	this.resources = vnic.Resources()
	this.primaryKeyAttribute = sla.PrimaryKeys()[0]

	node, _ := this.resources.Introspector().Inspect(this.element)
	introspecting.AddPrimaryKeyDecorator(node, this.primaryKeyAttribute)

	this.elements = dcache.NewDistributedCache(this.serviceName, this.serviceArea, this.element, nil,
		nil, this.resources)

	return this
}

func (this *InventoryCenter) Post(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		this.elements.Post(element, elements.Notification())
	}
}

func (this *InventoryCenter) Put(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		this.elements.Put(element, elements.Notification())
	}
}

func (this *InventoryCenter) Patch(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		this.elements.Patch(element, elements.Notification())
	}
}

func (this *InventoryCenter) Delete(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		this.elements.Delete(element, elements.Notification())
	}
}

func (this *InventoryCenter) Get(query ifs.IQuery) ([]interface{}, map[string]int32) {
	result := this.elements.Fetch(int(query.Page()*query.Limit()), int(query.Limit()), query)
	return result, this.elements.Stats()
}

func (this *InventoryCenter) ElementByElement(elem interface{}) interface{} {
	resp, _ := this.elements.Get(elem)
	return resp
}

func (this *InventoryCenter) AddStats(name string, f func(interface{}) bool) {
	this.elements.AddStatFunc(name, f)
}

func Inventory(resource ifs.IResources, serviceName string, serviceArea byte) *InventoryCenter {
	//serviceName = serviceName
	sp, ok := resource.Services().ServiceHandler(serviceName, serviceArea)
	if !ok {
		return nil
	}
	return (sp.(*InventoryService)).inventoryCenter
}
