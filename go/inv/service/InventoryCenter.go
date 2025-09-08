package inventory

import (
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
	this.elements = dcache.NewDistributedCache(this.serviceName, this.serviceArea, this.elementType.Name(),
		resources.SysConfig().LocalUuid, listener, resources)
	node, _ := resources.Introspector().Inspect(element)
	introspecting.AddPrimaryKeyDecorator(node, primaryKeyAttribute)
	return this
}

func (this *InventoryCenter) Post(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		key := primaryKeyValue(this.primaryKeyAttribute, element, this.resources)
		if key != "" {
			this.elements.Post(key, element, elements.Notification())
		}
	}
}

func (this *InventoryCenter) Put(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		key := primaryKeyValue(this.primaryKeyAttribute, element, this.resources)
		if key != "" {
			this.elements.Put(key, element, elements.Notification())
		}
	}
}

func (this *InventoryCenter) Patch(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		key := primaryKeyValue(this.primaryKeyAttribute, element, this.resources)
		if key != "" {
			this.elements.Patch(key, element, elements.Notification())
		}
	}
}

func (this *InventoryCenter) Delete(elements ifs.IElements) {
	for _, element := range elements.Elements() {
		key := primaryKeyValue(this.primaryKeyAttribute, element, this.resources)
		if key != "" {
			this.elements.Delete(key, elements.Notification())
		}
	}
}

func (this *InventoryCenter) Get(query ifs.IQuery) []interface{} {
	result := make([]interface{}, 0)
	this.elements.Collect(func(elem interface{}) (bool, interface{}) {
		match := query.Match(elem)
		if match {
			result = append(result, elem)
		}
		return match, elem
	})
	return result
}

func (this *InventoryCenter) ElementByKey(key string) interface{} {
	return this.elements.Get(key)
}

func (this *InventoryCenter) ElementByElement(elem interface{}) interface{} {
	key := primaryKeyValue(this.primaryKeyAttribute, elem, this.resources)
	return this.elements.Get(key)
}

func Inventory(resource ifs.IResources, serviceName string, serviceArea byte) *InventoryCenter {
	//serviceName = serviceName
	sp, ok := resource.Services().ServiceHandler(serviceName, serviceArea)
	if !ok {
		return nil
	}
	return (sp.(*InventoryService)).inventoryCenter
}
