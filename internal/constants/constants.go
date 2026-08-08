// Package constants defines shared label keys, annotation keys, and well-known
// names used across all FNO components.
package constants

const (
	// DefaultNodePoolLabelKey is the label applied to nodes to assign them
	// to a simulated NIC node pool.
	DefaultNodePoolLabelKey = "fake-network-operator.io/nic-node-pool"

	// NicReadyLabelKey indicates that the fake NIC stack is fully configured
	// on a node.
	NicReadyLabelKey = "fake-network-operator.io/nic-ready"

	// GPUDirectCapableLabelKey indicates the node pool supports GPUDirect.
	GPUDirectCapableLabelKey = "fake-network-operator.io/gpudirect-capable"

	// ----- NFD-compatible labels -----

	// PCIMellanoxPresentLabel mirrors NFD's PCI vendor detection for Mellanox.
	PCIMellanoxPresentLabel = "feature.node.kubernetes.io/pci-15b3.present"

	// PCISRIOVCapableLabel mirrors NFD's SR-IOV capability detection.
	PCISRIOVCapableLabel = "feature.node.kubernetes.io/pci-15b3.sriov.capable"

	// ----- NIC Feature Discovery compatible labels -----

	NicProductLabel         = "nvidia.com/nic.product"
	NicFirmwareVersionLabel = "nvidia.com/nic.firmware.version"
	NicDriverVersionLabel   = "nvidia.com/nic.driver.version"
	NicRDMACapableLabel     = "nvidia.com/nic.rdma.capable"
	NicSRIOVCapableLabel    = "nvidia.com/nic.sriov.capable"

	// ----- Annotations -----

	// SimulatedRDMAUtilizationAnnotation lets pods override the simulated
	// RDMA utilization reported by the status exporter.
	// Format: "min-max" (e.g., "40-60")
	SimulatedRDMAUtilizationAnnotation = "fake-network-operator.io/simulated-rdma-utilization"

	// MultusNetworksAnnotation is the standard Multus annotation for
	// requesting secondary network attachments.
	MultusNetworksAnnotation = "k8s.v1.cni.cncf.io/networks"

	// ----- Well-known names -----

	// SingletonPolicyName is the required name for the FakeNicClusterPolicy CR.
	SingletonPolicyName = "nic-cluster-policy"

	// TopologyConfigMapPrefix is the prefix for per-node topology ConfigMaps.
	// Full name: "nic-topology-<nodeName>"
	TopologyConfigMapPrefix = "nic-topology-"

	// DefaultNamespace is the default namespace for FNO components.
	DefaultNamespace = "fake-network-operator"

	// DefaultRDMASharedCount is the default number of pods that can share
	// a single RDMA device.
	DefaultRDMASharedCount = 64

	// DefaultOFEDVersion is the simulated MLNX_OFED driver version.
	DefaultOFEDVersion = "5.8-1.0.1.1"

	// ----- Device ID mappings (PCI device ID → product name) -----

	DeviceIDConnectX5   = "1017"
	DeviceIDConnectX5Ex = "1019"
	DeviceIDConnectX6   = "101b"
	DeviceIDConnectX6Dx = "101d"
	DeviceIDConnectX6Lx = "101f"
	DeviceIDConnectX7   = "1021"
	DeviceIDBF2         = "a2d6"
	DeviceIDBF3         = "a2dc"
)

// DeviceIDToProduct maps PCI device IDs to human-readable product names.
var DeviceIDToProduct = map[string]string{
	DeviceIDConnectX5:   "ConnectX-5",
	DeviceIDConnectX5Ex: "ConnectX-5 Ex",
	DeviceIDConnectX6:   "ConnectX-6",
	DeviceIDConnectX6Dx: "ConnectX-6 Dx",
	DeviceIDConnectX6Lx: "ConnectX-6 Lx",
	DeviceIDConnectX7:   "ConnectX-7",
	DeviceIDBF2:         "BlueField-2",
	DeviceIDBF3:         "BlueField-3",
}
