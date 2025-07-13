package tests

import (
	inventory "github.com/saichler/l8inventory/go/inv/service"
	"github.com/saichler/l8pollaris/go/types"
	"github.com/saichler/l8types/go/testtypes"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

func TestParser(t *testing.T) {
	forwardInfo := &types.DeviceServiceInfo{}
	forwardInfo.ServiceName = "MockOrm"
	forwardInfo.ServiceArea = 0
	vnic := topo.VnicByVnetNum(2, 2)
	vnic.Resources().Registry().Register(&inventory.InventoryService{})
	vnic.Resources().Services().Activate(inventory.ServiceType, "inventory", 0, vnic.Resources(), vnic,
		"MyString", &testtypes.TestProto{}, forwardInfo)
}
