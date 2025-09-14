package inventory

import (
	"fmt"
	"reflect"
	"sync"

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

	query    []interface{}
	queryMtx *sync.RWMutex
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
	this.queryMtx = &sync.RWMutex{}

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

func (this *InventoryCenter) shouldPrepareQuery() bool {
	this.queryMtx.Lock()
	defer this.queryMtx.Unlock()
	if this.query == nil || len(this.query) != this.elements.Size() {
		fmt.Println("Query length mismatch, recreatng")
		return true
	}
	return false
}

func (this *InventoryCenter) Get(query ifs.IQuery) ([]interface{}, int32) {
	if this.shouldPrepareQuery() {
		fmt.Println("Prepare query")
		localQuery := make([]interface{}, 0)
		this.elements.Collect(func(elem interface{}) (bool, interface{}) {
			localQuery = append(localQuery, elem)
			return true, elem
		})
		this.queryMtx.Lock()
		fmt.Println("local query:", len(localQuery))
		this.query = localQuery
		this.queryMtx.Unlock()
	}

	result := make([]interface{}, 0)

	this.queryMtx.RLock()
	defer this.queryMtx.RUnlock()

	if query.Limit() == 0 {
		for _, elem := range this.query {
			result = append(result, elem)
		}
		return result, 0
	}

	startIndex := int(query.Limit() * query.Page())
	endIndex := startIndex + int(query.Limit())
	fmt.Println("Start Index = ", startIndex, " EndIndex = ", endIndex, " len:", len(this.query))
	for i := startIndex; i < endIndex && i < len(this.query); i++ {
		result = append(result, this.query[i])
	}
	return result, int32(len(this.query)/int(query.Limit()) + 1)
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
