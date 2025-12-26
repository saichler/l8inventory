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

package inventory

import (
	"reflect"

	"github.com/saichler/l8srlz/go/serialize/object"
)

// AddEmpty creates and adds a new empty inventory element with only the primary key
// set. This is useful for reserving a key in the cache before the full element data
// is available, or for creating placeholder entries.
//
// The method uses reflection to create a new instance of the element type, sets the
// primary key field to the provided key value, and posts it to the cache.
//
// Parameters:
//   - key: The primary key value to set on the new element
//
// Example:
//
//	inventoryCenter.AddEmpty("device-12345")
func (this *InventoryCenter) AddEmpty(key string) {
	elem := reflect.New(this.elementType)
	field := elem.Elem().FieldByName(this.primaryKeyAttribute)
	field.Set(reflect.ValueOf(key))
	element := object.New(nil, elem.Interface())
	this.Post(element)
}
