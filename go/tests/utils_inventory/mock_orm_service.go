package utils_inventory

import (
	"sync"

	
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
)

const (
	ServiceType = "MockOrmService"
)

type MockOrmService struct {
	postCount  int
	patchCount int
	mtx        *sync.Mutex
}

func (this *MockOrmService) Activate(serviceName string, serviceArea byte,
	r ifs.IResources, l ifs.IServiceCacheListener, args ...interface{}) error {
	r.Registry().Register(&types.CJob{})
	this.mtx = &sync.Mutex{}
	return nil
}

func (this *MockOrmService) DeActivate() error {
	return nil
}

func (this *MockOrmService) Post(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.postCount++
	return object.New(nil, nil)
}

func (this *MockOrmService) PostCount() int {
	return this.postCount
}

func (this *MockOrmService) PatchCount() int {
	return this.patchCount
}

func (this *MockOrmService) Put(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *MockOrmService) Patch(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.patchCount++
	return object.New(nil, nil)
}
func (this *MockOrmService) Delete(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *MockOrmService) Get(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *MockOrmService) GetCopy(pb ifs.IElements, vnic ifs.IVNic) ifs.IElements {
	return nil
}
func (this *MockOrmService) Failed(pb ifs.IElements, vnic ifs.IVNic, msg *ifs.Message) ifs.IElements {
	return nil
}
func (this *MockOrmService) TransactionConfig() ifs.ITransactionConfig {
	return nil
}
func (this *MockOrmService) WebService() ifs.IWebService {
	return nil
}
