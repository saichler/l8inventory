// © 2025 Sharon Aicler (saichler@gmail.com)
//
// Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package tests provides unit tests for the l8inventory distributed cache service.
// It tests the core functionality including service activation, CRUD operations,
// service linking, and query execution.
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

// TestMain is the test entry point that sets up the test topology before running
// tests and tears it down afterward. It uses the Layer 8 test infrastructure to
// create a distributed environment for integration testing.
func TestMain(m *testing.M) {
	setup()
	m.Run()
	tear()
}

// TestInventory is the main integration test for the inventory service. It tests:
//   - Service activation with forwarding configuration
//   - POST operation and verification that it forwards to the mock ORM service
//   - PATCH operation and verification of forwarding
//   - Element retrieval by primary key
//   - Query execution with SQL-like syntax
//
// The test uses a mock ORM service to verify that operations are correctly
// forwarded to downstream services when service linking is configured.
func TestInventory(t *testing.T) {
	forwardInfo := &l8services.L8ServiceLink{}
	forwardInfo.ZsideServiceName = "MockOrm"
	forwardInfo.ZsideServiceArea = 0
	serviceName := "inventory"
	serviceArea := byte(0)
	primaryKey := "MyString"
	elemType := &testtypes.TestProto{}
	elemTypeList := &testtypes.TestProtoList{}
	elem := &testtypes.TestProto{MyString: "Hello World", MyInt64: 67}

	vnic := topo.VnicByVnetNum(2, 2)
	sla := ifs.NewServiceLevelAgreement(&inventory.InventoryService{}, serviceName, serviceArea, true, nil)
	sla.SetServiceItem(elemType)
	sla.SetServiceItemList(elemTypeList)
	sla.SetArgs(forwardInfo)
	sla.SetPrimaryKeys(primaryKey)
	vnic.Resources().Services().Activate(sla, vnic)

	sla = ifs.NewServiceLevelAgreement(&utils_inventory.MockOrmService{}, forwardInfo.ZsideServiceName,
		byte(forwardInfo.ZsideServiceArea), false, nil)
	vnic.Resources().Services().Activate(sla, vnic)

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
