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

// Package wsnotify tests InventoryService's live client notifications: every
// write used to multicast an untargeted, record-less notification to the
// websocket service (broadcast to every browser, each re-fetching its page);
// now only the cache's targeted notification is sent, carrying the record,
// and only for writes on a subscriber's displayed page
// (probler/plans/websocket-page-scoped-notifications.md Phase 3).
package wsnotify

import (
	"fmt"
	"sync"
	"testing"

	inventory "github.com/saichler/l8inventory/go/inv/service"
	"github.com/saichler/l8reflect/go/reflect/introspecting"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/testtypes"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8notify"
	"github.com/saichler/l8types/go/types/l8sysconfig"
	"github.com/saichler/l8utils/go/utils/logger"
	"github.com/saichler/l8utils/go/utils/registry"
	"github.com/saichler/l8utils/go/utils/resources"
)

// captureVnic provides real resources and records Multicast calls; every
// other IVNic method is unused here (the embedded nil interface would panic).
type captureVnic struct {
	ifs.IVNic
	res  ifs.IResources
	mtx  sync.Mutex
	sent []*l8notify.L8NotificationSet
}

func (this *captureVnic) Resources() ifs.IResources { return this.res }

func (this *captureVnic) Multicast(serviceName string, serviceArea byte, action ifs.Action, data interface{}) error {
	if serviceName != "websock" || serviceArea != 0 {
		return fmt.Errorf("unexpected multicast to %s/%d", serviceName, serviceArea)
	}
	n, ok := data.(*l8notify.L8NotificationSet)
	if !ok {
		return fmt.Errorf("unexpected multicast payload %T", data)
	}
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.sent = append(this.sent, n)
	return nil
}

func (this *captureVnic) take() []*l8notify.L8NotificationSet {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	s := this.sent
	this.sent = nil
	return s
}

func newVnic() *captureVnic {
	res := resources.NewResources(logger.NewLoggerDirectImpl(&logger.FmtLogMethod{}))
	res.Set(registry.NewRegistry())
	res.Set(&l8sysconfig.L8SysConfig{})
	res.Set(introspecting.NewIntrospect(res.Registry()))
	return &captureVnic{res: res}
}

func item(i int, v float64) *testtypes.TestProto {
	return &testtypes.TestProto{MyString: fmt.Sprintf("dev-%05d", i), MyInt32: int32(i), MyFloat64: v}
}

func newService(t *testing.T, vnic *captureVnic, n int) *inventory.InventoryService {
	t.Helper()
	sla := ifs.NewServiceLevelAgreement(&inventory.InventoryService{}, "NetDev", 0, true, nil)
	sla.SetServiceItem(&testtypes.TestProto{})
	sla.SetServiceItemList(&testtypes.TestProtoList{})
	sla.SetPrimaryKeys("MyString")
	svc := &inventory.InventoryService{}
	if err := svc.Activate(sla, vnic); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		// Initial load as replication (Notification) so nothing is sent.
		svc.Post(object.NewNotify(item(i, 1)), vnic)
	}
	vnic.take()
	return svc
}

// registerPage fetches the first page the way a live table does, which
// registers aaaId's subscription to that page.
func registerPage(t *testing.T, svc *inventory.InventoryService, vnic *captureVnic, aaaId string) {
	t.Helper()
	q, err := object.NewFromQuery(&l8api.L8Query{Text: "select * from TestProto limit 15 page 0",
		AaaId: aaaId, Register: l8api.L8Register_L8_REGISTER_REGISTER, RegisterId: "table"}, vnic.res)
	if err != nil {
		t.Fatal(err)
	}
	result := svc.Get(q, vnic)
	if result.Error() != nil {
		t.Fatal(result.Error())
	}
	if n := len(result.Elements()); n != 15 {
		t.Fatalf("expected a first page of 15, got %d", n)
	}
}

func TestInventoryNotifiesOnlyDisplayedPage(t *testing.T) {
	vnic := newVnic()
	svc := newService(t, vnic, 3000)
	registerPage(t, svc, vnic, "aaa-1")

	// Writes to devices off the displayed page send nothing at all -- before
	// the fix each one was a broadcast to every browser.
	for i := 15; i < 3000; i += 7 {
		svc.Patch(object.New(nil, item(i, 2)), vnic)
	}
	svc.Put(object.New(nil, item(2000, 3)), vnic)
	svc.Delete(object.New(nil, item(2500, 0)), vnic)
	if sent := vnic.take(); len(sent) != 0 {
		t.Fatalf("expected no websocket notifications for off-page writes, got %d", len(sent))
	}

	// A write to a displayed device sends exactly one notification, targeted
	// at the subscriber and carrying the record.
	svc.Patch(object.New(nil, item(3, 2)), vnic)
	sent := vnic.take()
	if len(sent) != 1 {
		t.Fatalf("expected 1 websocket notification for an on-page patch, got %d", len(sent))
	}
	n := sent[0]
	if len(n.AaaIds) != 1 || !n.AaaIds["aaa-1"] {
		t.Errorf("expected the notification to target only aaa-1, got %v", n.AaaIds)
	}
	if len(n.NotificationList) == 0 || n.NotificationList[0].NewValue == nil {
		t.Error("expected the notification to carry the record")
	}
	if n.ModelKey != "dev-00003" {
		t.Errorf("expected ModelKey dev-00003, got %q", n.ModelKey)
	}

	// On-page replace and delete send one each.
	svc.Put(object.New(nil, item(4, 5)), vnic)
	svc.Delete(object.New(nil, item(5, 0)), vnic)
	if sent := vnic.take(); len(sent) != 2 {
		t.Errorf("expected 2 websocket notifications for on-page put+delete, got %d", len(sent))
	}

	// Replicated writes (Notification) never notify clients.
	svc.Patch(object.NewNotify(item(6, 9)), vnic)
	if sent := vnic.take(); len(sent) != 0 {
		t.Errorf("expected no websocket notification for a replicated write, got %d", len(sent))
	}
}

// With no subscriber, writes send nothing (the old notifyWs broadcast every
// write regardless).
func TestInventoryNoSubscriberSendsNothing(t *testing.T) {
	vnic := newVnic()
	svc := newService(t, vnic, 50)
	svc.Patch(object.New(nil, item(3, 2)), vnic)
	svc.Post(object.New(nil, item(999, 1)), vnic)
	if sent := vnic.take(); len(sent) != 0 {
		t.Errorf("expected no websocket notifications without subscribers, got %d", len(sent))
	}
}
