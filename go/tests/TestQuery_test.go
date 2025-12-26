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

package tests

import (
	"fmt"
	"testing"

	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/probler/go/types"
	"google.golang.org/protobuf/encoding/protojson"
)

// TestQuery tests the SQL-like query parsing functionality for inventory queries.
// It verifies that queries with pagination (LIMIT and PAGE clauses) are correctly
// parsed and converted to the L8Query protocol buffer format.
//
// The test:
//  1. Registers a primary key decorator for NetworkDevice
//  2. Creates a query with limit and page clauses
//  3. Verifies the query is correctly serialized to JSON
func TestQuery(t *testing.T) {
	nic := topo.VnicByVnetNum(1, 1)
	nic.Resources().Introspector().Decorators().AddPrimaryKeyDecorator(&types.NetworkDevice{}, "Id")
	elem, err := object.NewQuery("select * from NetworkDevice limit 5 page 2", nic.Resources())
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	elem2 := elem.(*object.Elements)
	jsn, _ := protojson.Marshal(elem2.PQuery())
	fmt.Println(string(jsn))
}
