package inventory

import (
	"reflect"

	"github.com/saichler/l8srlz/go/serialize/object"
)

func (this *InventoryCenter) AddEmpty(key string) {
	elem := reflect.New(this.elementType)
	field := elem.Elem().FieldByName(this.primaryKeyAttribute)
	field.Set(reflect.ValueOf(key))
	element := object.New(nil, elem.Interface())
	this.Post(element)
}
