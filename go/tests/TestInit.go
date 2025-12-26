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
	"github.com/saichler/l8bus/go/overlay/protocol"
	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8test/go/infra/t_topology"
	. "github.com/saichler/l8types/go/ifs"
)

// topo is the shared test topology used across all test cases. It provides
// a distributed Layer 8 environment with multiple virtual network interfaces.
var topo *TestTopology

// init sets the default log level to Trace for detailed test output.
func init() {
	Log.SetLogLevel(Trace_Level)
}

// setup initializes the test environment by creating the test topology.
// Called by TestMain before running any tests.
func setup() {
	setupTopology()
}

// tear cleans up the test environment by shutting down the test topology.
// Called by TestMain after all tests have completed.
func tear() {
	shutdownTopology()
}

// reset resets the test handlers between test cases and logs the completion
// of a test. This ensures test isolation by clearing any accumulated state.
func reset(name string) {
	Log.Info("*** ", name, " end ***")
	topo.ResetHandlers()
}

// setupTopology creates a new test topology with 4 nodes across 3 virtual
// networks on ports 20000, 30000, and 40000. Message logging is enabled
// for debugging purposes.
func setupTopology() {
	protocol.MessageLog = true
	topo = NewTestTopology(4, []int{20000, 30000, 40000}, Info_Level)
}

// shutdownTopology gracefully shuts down the test topology and releases
// all associated resources.
func shutdownTopology() {
	topo.Shutdown()
}
