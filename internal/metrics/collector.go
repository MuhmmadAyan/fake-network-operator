package metrics

import (
	"math/rand"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/fake-network-operator/fake-network-operator/internal/topology"
)

var (
	fakeNicPortRxBytesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fake_nic_port_rx_bytes_total",
			Help: "Fake NIC port RX bytes total",
		},
		[]string{"node", "nic"},
	)
	fakeNicPortTxBytesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fake_nic_port_tx_bytes_total",
			Help: "Fake NIC port TX bytes total",
		},
		[]string{"node", "nic"},
	)
	fakeNicPortLinkState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fake_nic_port_link_state",
			Help: "Fake NIC port link state (1.0 for Active, 0.0 otherwise)",
		},
		[]string{"node", "nic"},
	)
	fakeRdmaDeviceUtilization = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fake_rdma_device_utilization",
			Help: "Fake RDMA device utilization",
		},
		[]string{"node", "nic"},
	)
	fakeSriovVfAllocated = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fake_sriov_vf_allocated",
			Help: "Fake SR-IOV VF allocated",
		},
		[]string{"node", "nic"},
	)
	fakeSriovVfTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fake_sriov_vf_total",
			Help: "Fake SR-IOV VF total",
		},
		[]string{"node", "nic"},
	)
)

func init() {
	prometheus.MustRegister(fakeNicPortRxBytesTotal)
	prometheus.MustRegister(fakeNicPortTxBytesTotal)
	prometheus.MustRegister(fakeNicPortLinkState)
	prometheus.MustRegister(fakeRdmaDeviceUtilization)
	prometheus.MustRegister(fakeSriovVfAllocated)
	prometheus.MustRegister(fakeSriovVfTotal)
}

// NicMetricsCollector wraps metadata about the metrics collection
type NicMetricsCollector struct {
	namespace string
	nodeName  string
}

// NewNicMetricsCollector creates a new NicMetricsCollector
func NewNicMetricsCollector(namespace, nodeName string) *NicMetricsCollector {
	return &NicMetricsCollector{
		namespace: namespace,
		nodeName:  nodeName,
	}
}

// GenerateSyntheticValue produces a random value between min and max.
func GenerateSyntheticValue(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// UpdateFromTopology updates the Prometheus metrics based on the provided topology.
func (c *NicMetricsCollector) UpdateFromTopology(topo *topology.NodeNicTopology) {
	for _, nic := range topo.NICs {
		fakeNicPortRxBytesTotal.WithLabelValues(topo.NodeName, nic.Name).Set(GenerateSyntheticValue(1000, 100000))
		fakeNicPortTxBytesTotal.WithLabelValues(topo.NodeName, nic.Name).Set(GenerateSyntheticValue(1000, 100000))

		stateValue := 1.0
		if nic.State != "Active" {
			stateValue = 0.0
		}
		fakeNicPortLinkState.WithLabelValues(topo.NodeName, nic.Name).Set(stateValue)

		if nic.RDMA != nil {
			fakeRdmaDeviceUtilization.WithLabelValues(topo.NodeName, nic.Name).Set(GenerateSyntheticValue(0, 100))
		}

		if nic.SRIOV != nil {
			fakeSriovVfTotal.WithLabelValues(topo.NodeName, nic.Name).Set(float64(nic.SRIOV.TotalVFs))
			fakeSriovVfAllocated.WithLabelValues(topo.NodeName, nic.Name).Set(GenerateSyntheticValue(0, float64(nic.SRIOV.TotalVFs)))
		}
	}
}
