package inventory

import (
	"math"
	"reflect"

	"github.com/saichler/l8services/go/services/dcache"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/reflect/go/reflect/introspecting"
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

func newInventoryCenter(serviceName string, serviceArea byte, primaryKeyAttribute string,
	element interface{}, resources ifs.IResources, listener ifs.IServiceCacheListener) *InventoryCenter {
	this := &InventoryCenter{}
	this.serviceName = serviceName
	this.serviceArea = serviceArea
	this.element = element
	this.elementType = reflect.ValueOf(element).Elem().Type()
	this.resources = resources
	this.primaryKeyAttribute = primaryKeyAttribute
	node, _ := resources.Introspector().Inspect(element)
	introspecting.AddPrimaryKeyDecorator(node, primaryKeyAttribute)

	this.elements = dcache.NewDistributedCache(this.serviceName, this.serviceArea, this.element, nil,
		listener, resources)

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

func (this *InventoryCenter) Get(query ifs.IQuery) []interface{} {
	result := make([]interface{}, 0)
	startRec := 0
	endRec := math.MaxInt
	if query.Limit() > 0 {
		startRec = int(query.Page() * query.Limit())
		endRec = int((query.Page() + 1) * query.Limit())
	}
	currRec := 0
	this.elements.Collect(func(elem interface{}) (bool, interface{}) {
		match := query.Match(elem)
		if match {
			currRec++
			if currRec >= startRec && currRec <= endRec {
				result = append(result, elem)
			}
		}
		return match, elem
	})
	return result
}

func (this *InventoryCenter) ElementByElement(elem interface{}) interface{} {
	resp, _ := this.elements.Get(elem)
	return resp
}

func Inventory(resource ifs.IResources, serviceName string, serviceArea byte) *InventoryCenter {
	//serviceName = serviceName
	sp, ok := resource.Services().ServiceHandler(serviceName, serviceArea)
	if !ok {
		return nil
	}
	return (sp.(*InventoryService)).inventoryCenter
}
