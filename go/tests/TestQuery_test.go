package tests

import (
	"fmt"
	"testing"

	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/probler/go/types"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestQuery(t *testing.T) {
	nic := topo.VnicByVnetNum(1, 1)
	nic.Resources().Introspector().Decorators().AddPrimaryKeyDecorator(&types.NetworkDevice{}, "Id")
	elem, err := object.NewQuery("select * from NetworkDevice limit 5 page 2", nic.Resources())
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	jsn, _ := protojson.Marshal(elem.PQuery())
	fmt.Println(string(jsn))
}
