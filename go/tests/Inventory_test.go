package tests

import (
	"fmt"
	"testing"
	"time"

	inventory "github.com/saichler/l8inventory/go/inv/service"
	"github.com/saichler/l8inventory/go/tests/utils_inventory"
	"github.com/saichler/l8types/go/types/l8services"

	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/testtypes"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func TestInventory(t *testing.T) {
	forwardInfo := &l8services.L8ServiceLink{}
	forwardInfo.ZsideServiceName = "MockOrm"
	forwardInfo.ZsideServiceArea = 0
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
		forwardInfo.ZsideServiceName, byte(forwardInfo.ZsideServiceArea), vnic.Resources(), vnic)

	time.Sleep(time.Second)

	ci := topo.VnicByVnetNum(1, 1)
	ci.ProximityRequest(serviceName, serviceArea, ifs.POST, elem, 30)

	time.Sleep(time.Second * 5)

	m, ok := vnic.Resources().Services().ServiceHandler(forwardInfo.ZsideServiceName, byte(forwardInfo.ZsideServiceArea))
	if !ok {
		vnic.Resources().Logger().Fail(t, "Cannot find mock service")
		return
	}
	mock := m.(*utils_inventory.MockOrmService)
	if mock.PostCount() != 1 {
		vnic.Resources().Logger().Fail(t, "Expected 1 post count in mock ", mock.PostCount())
		return
	}

	elem = &testtypes.TestProto{MyString: "Hello World", MyInt32: 13}
	ci.ProximityRequest(serviceName, serviceArea, ifs.PATCH, elem, 30)

	time.Sleep(time.Second * 5)

	if mock.PatchCount() != 1 {
		vnic.Resources().Logger().Fail(t, "Expected 1 patch count in mock")
		return
	}

	inventoryCenter := inventory.Inventory(vnic.Resources(), serviceName, serviceArea)
	elem = inventoryCenter.ElementByElement(elem).(*testtypes.TestProto)
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
	all, _ := inventoryCenter.Get(q)
	fmt.Println(all)
}
