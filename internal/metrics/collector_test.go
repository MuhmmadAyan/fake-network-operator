package metrics

import (
	"testing"

	"github.com/fake-network-operator/fake-network-operator/internal/topology"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewCollector(t *testing.T) {
	collector := NewNicMetricsCollector("default", "node-1")
	assert.NotNil(t, collector)
	assert.Equal(t, "default", collector.namespace)
	assert.Equal(t, "node-1", collector.nodeName)
}

func TestUpdateFromTopology(t *testing.T) {
	// Clear registry or reset metrics if needed, but since we are checking individual vectors, we can just reset them
	fakeNicPortRxBytesTotal.Reset()
	fakeNicPortTxBytesTotal.Reset()
	fakeNicPortLinkState.Reset()
	fakeRdmaDeviceUtilization.Reset()
	fakeSriovVfAllocated.Reset()
	fakeSriovVfTotal.Reset()

	collector := NewNicMetricsCollector("default", "node-1")

	topo := &topology.NodeNicTopology{
		NodeName: "node-1",
		NICs: []topology.NicTopologyEntry{
			{
				Name:  "mlx5_0",
				State: "Active",
				RDMA: &topology.RDMATopology{
					ResourceName: "rdma_shared_device_a",
					SharedCount:  64,
				},
			},
			{
				Name:  "mlx5_1",
				State: "Down",
				SRIOV: &topology.SRIOVTopology{
					TotalVFs: 8,
				},
			},
		},
	}

	collector.UpdateFromTopology(topo)

	// Check mlx5_0 link state (Active -> 1.0)
	val, err := getMetricValue(fakeNicPortLinkState, "node-1", "mlx5_0")
	assert.NoError(t, err)
	assert.Equal(t, 1.0, val)

	// Check mlx5_1 link state (Down -> 0.0)
	val, err = getMetricValue(fakeNicPortLinkState, "node-1", "mlx5_1")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, val)

	// Check SR-IOV total for mlx5_1
	val, err = getMetricValue(fakeSriovVfTotal, "node-1", "mlx5_1")
	assert.NoError(t, err)
	assert.Equal(t, 8.0, val)

	// We can also verify that metrics like RX/TX are populated
	assert.Equal(t, 2, testutil.CollectAndCount(fakeNicPortRxBytesTotal))
	assert.Equal(t, 2, testutil.CollectAndCount(fakeNicPortTxBytesTotal))
	assert.Equal(t, 1, testutil.CollectAndCount(fakeRdmaDeviceUtilization))
	assert.Equal(t, 1, testutil.CollectAndCount(fakeSriovVfAllocated))
}

func getMetricValue(vec *prometheus.GaugeVec, labels ...string) (float64, error) {
	metric, err := vec.GetMetricWithLabelValues(labels...)
	if err != nil {
		return 0, err
	}
	return testutil.ToFloat64(metric), nil
}
