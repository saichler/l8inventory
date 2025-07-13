package inventory

import (
	"github.com/saichler/l8types/go/ifs"
	"reflect"
)

func primaryKeyValue(attr string, any interface{}, resources ifs.IResources) string {
	if any == nil {
		resources.Logger().Error("element is nil")
		return ""
	}
	v := reflect.ValueOf(any)
	if v.Kind() != reflect.Ptr {
		resources.Logger().Error("element is not a pointer")
		return ""
	}
	field := v.Elem().FieldByName(attr)
	if !field.IsValid() {
		resources.Logger().Error("attribute " + attr + " not found")
		return ""
	}
	return field.String()
}

func (this *InventoryCenter) AddEmpty(key string) {
	elem := reflect.New(this.elementType)
	field := elem.Elem().FieldByName(this.primaryKeyAttribute)
	field.Set(reflect.ValueOf(key))
	this.Add(elem.Interface(), false)
}
