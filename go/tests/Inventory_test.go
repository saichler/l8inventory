package tests

import (
	"fmt"
	inventory "github.com/saichler/l8inventory/go/inv/service"
	"github.com/saichler/l8inventory/go/tests/utils_inventory"
	"github.com/saichler/l8pollaris/go/types"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/testtypes"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func TestInventory(t *testing.T) {
	forwardInfo := &types.DeviceServiceInfo{}
	forwardInfo.ServiceName = "MockOrm"
	forwardInfo.ServiceArea = 0
	serviceName := "inventory"
	serviceArea := byte(0)
	primaryKey := "MyString"
	elemType := &testtypes.TestProto{}
	elem := &testtypes.TestProto{MyString: "Hello World", MyInt64: 67}

	vnic := topo.VnicByVnetNum(2, 2)
	vnic.Resources().Registry().Register(&inventory.InventoryService{})
	vnic.Resources().Services().Activate(inventory.ServiceType, serviceName, serviceArea, vnic.Resources(), vnic,
		primaryKey, elemType, forwardInfo)
	vnic.Resources().Registry().Register(&utils_inventory.MockOrmService{})
	vnic.Resources().Services().Activate(utils_inventory.ServiceType,
		forwardInfo.ServiceName, byte(forwardInfo.ServiceArea), vnic.Resources(), vnic)

	time.Sleep(time.Second)

	ci := topo.VnicByVnetNum(1, 1)
	ci.Proximity(serviceName, serviceArea, ifs.POST, elem)

	time.Sleep(time.Second)

	m, ok := vnic.Resources().Services().ServiceHandler(forwardInfo.ServiceName, byte(forwardInfo.ServiceArea))
	if !ok {
		vnic.Resources().Logger().Fail(t, "Cannot find mock service")
		return
	}
	mock := m.(*utils_inventory.MockOrmService)
	if mock.PostCount() != 1 {
		vnic.Resources().Logger().Fail(t, "Expected 1 post count in mock")
		return
	}

	elem = &testtypes.TestProto{MyString: "Hello World", MyInt32: 13}
	ci.Proximity(serviceName, serviceArea, ifs.PATCH, elem)
	time.Sleep(time.Second)

	if mock.PatchCount() != 1 {
		vnic.Resources().Logger().Fail(t, "Expected 1 patch count in mock")
		return
	}

	inventoryCenter := inventory.Inventory(vnic.Resources(), serviceName, serviceArea)
	elem = inventoryCenter.ElementByKey(elem.MyString).(*testtypes.TestProto)
	if elem.MyInt64 != 67 || elem.MyInt32 != 13 {
		vnic.Resources().Logger().Fail(t, "Expected values to match")
		return
	}

	elems, e := object.NewQuery("select * from testproto where mystring=*", vnic.Resources())
	if e != nil {
		vnic.Resources().Logger().Fail(t, "Unable to create query", e.Error())
		return
	}
	q, e := elems.Query(vnic.Resources())
	if e != nil {
		vnic.Resources().Logger().Fail(t, "Unable to create query", e.Error())
		return
	}
	all := inventoryCenter.Get(q)
	fmt.Println(all)
}
