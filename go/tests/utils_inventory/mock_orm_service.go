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

// Package utils_inventory provides test utilities for the l8inventory package,
// including mock service implementations for integration testing.
package utils_inventory

import (
	"sync"

	"github.com/saichler/l8pollaris/go/types/l8tpollaris"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
)

// ServiceType is the registered service type name for MockOrmService.
const (
	ServiceType = "MockOrmService"
)

// MockOrmService is a mock implementation of the Layer 8 service handler interface
// used for testing inventory service forwarding. It tracks the number of POST and
// PATCH operations received, allowing tests to verify that operations are correctly
// forwarded from the inventory service to downstream services.
type MockOrmService struct {
	// postCount tracks the number of POST operations received
	postCount int
	// patchCount tracks the number of PATCH operations received
	patchCount int
	// mtx provides thread-safe access to the counters
	mtx *sync.Mutex
}

// Activate initializes the mock service. It registers the CJob type and
// initializes the mutex for thread-safe counter access.
func (this *MockOrmService) Activate(sla *ifs.ServiceLevelAgreement, nic ifs.IVNic) error {
	nic.Resources().Registry().Register(&l8tpollaris.CJob{})
	this.mtx = &sync.Mutex{}
	return nil
}

// DeActivate cleans up mock service resources. Currently a no-op.
func (this *MockOrmService) DeActivate() error {
	return nil
}

// Post handles POST requests by incrementing the post counter.
// This allows tests to verify that POST operations are forwarded.
func (this *MockOrmService) Post(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.postCount++
	return object.New(nil, nil)
}

// PostCount returns the number of POST operations received by this mock service.
func (this *MockOrmService) PostCount() int {
	return this.postCount
}

// PatchCount returns the number of PATCH operations received by this mock service.
func (this *MockOrmService) PatchCount() int {
	return this.patchCount
}

// Put handles PUT requests. Currently returns nil as it's not tracked.
func (this *MockOrmService) Put(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}

// Patch handles PATCH requests by incrementing the patch counter.
// This allows tests to verify that PATCH operations are forwarded.
func (this *MockOrmService) Patch(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.patchCount++
	return object.New(nil, nil)
}

// Delete handles DELETE requests. Currently returns nil as it's not tracked.
func (this *MockOrmService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}

// Get handles GET requests. Currently returns nil as it's not implemented.
func (this *MockOrmService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}

// GetCopy handles copy requests. Currently returns nil as it's not implemented.
func (this *MockOrmService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}

// Failed handles failure notifications. Currently returns nil as it's not implemented.
func (this *MockOrmService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}

// TransactionConfig returns the transaction configuration. Returns nil as the
// mock service doesn't support transactions.
func (this *MockOrmService) TransactionConfig() ifs.ITransactionConfig {
	return nil
}

// WebService returns the web service configuration. Returns nil as the mock
// service doesn't expose a web interface.
func (this *MockOrmService) WebService() ifs.IWebService {
	return nil
}
