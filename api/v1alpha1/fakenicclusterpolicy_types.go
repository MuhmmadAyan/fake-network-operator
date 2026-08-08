package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// FakeNicClusterPolicy is the Schema for the fakenicclusterpolicies API.
// Only one instance (named "nic-cluster-policy") should exist per cluster.
type FakeNicClusterPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FakeNicClusterPolicySpec   `json:"spec,omitempty"`
	Status FakeNicClusterPolicyStatus `json:"status,omitempty"`
}

// FakeNicClusterPolicySpec defines the desired state of FakeNicClusterPolicy
type FakeNicClusterPolicySpec struct {
	// NodePoolSelector is the label key used to assign nodes to NIC pools.
	// Defaults to "fake-network-operator.io/nic-node-pool"
	// +kubebuilder:default="fake-network-operator.io/nic-node-pool"
	NodePoolSelector string `json:"nodePoolSelector,omitempty"`

	// NicNodePools maps pool names to NIC topology configurations.
	// Each pool describes the simulated NIC hardware for nodes carrying the
	// corresponding pool label value.
	// +kubebuilder:validation:MinProperties=1
	NicNodePools map[string]NicNodePoolSpec `json:"nicNodePools"`

	// Components toggles for each fake sub-component
	Components ComponentsSpec `json:"components,omitempty"`

	// SecondaryNetworks defines NetworkAttachmentDefinitions to generate.
	// Each entry produces a Multus NAD backed by a lightweight CNI delegate
	// (macvlan/bridge) instead of real hardware.
	// +optional
	SecondaryNetworks []SecondaryNetworkSpec `json:"secondaryNetworks,omitempty"`
}

// NicNodePoolSpec describes the simulated NIC topology for a group of nodes.
type NicNodePoolSpec struct {
	// NICs describes custom simulated NICs on nodes in this pool.
	// If omitted, standard dual-port Mellanox ConnectX-6 topology is automatically synthesized.
	// +optional
	NICs []NicSpec `json:"nics,omitempty"`

	// RDMA configures default RDMA shared device settings for this pool.
	// +optional
	RDMA *RDMASpec `json:"rdma,omitempty"`

	// SRIOV configures default SR-IOV VF settings for this pool.
	// +optional
	SRIOV *SRIOVSpec `json:"sriov,omitempty"`

	// GPUDirectCapable marks this pool as co-locatable with a fake GPU pool.
	// When true, the labeler will also set labels indicating GPUDirect readiness.
	// +optional
	GPUDirectCapable bool `json:"gpuDirectCapable,omitempty"`

	// Labels are additional node labels to apply beyond the standard set.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// NicSpec describes a single simulated NIC.
type NicSpec struct {
	// Name is the simulated network interface name (e.g., "mlx5_0", "ens2f0")
	// +kubebuilder:validation:Pattern=`^[a-zA-Z][a-zA-Z0-9_]*$`
	Name string `json:"name"`

	// LinkType is "ethernet" or "infiniband"
	// +kubebuilder:validation:Enum=ethernet;infiniband
	LinkType string `json:"linkType"`

	// Vendor is the PCI vendor ID (e.g., "15b3" for Mellanox/NVIDIA)
	// +kubebuilder:default="15b3"
	Vendor string `json:"vendor,omitempty"`

	// DeviceID is the PCI device ID (e.g., "101b" for ConnectX-6, "1017" for ConnectX-5)
	DeviceID string `json:"deviceID"`

	// FirmwareVersion is the simulated firmware version string
	// +kubebuilder:default="20.31.1014"
	FirmwareVersion string `json:"firmwareVersion,omitempty"`

	// BoardID is the simulated board identifier
	// +optional
	BoardID string `json:"boardID,omitempty"`

	// NUMANode is the NUMA node affinity for topology-aware scheduling
	// +kubebuilder:default=0
	NUMANode int `json:"numaNode,omitempty"`

	// Speed is the simulated link speed (e.g., "100 Gb/sec", "200 Gb/sec")
	// +kubebuilder:default="100 Gb/sec"
	Speed string `json:"speed,omitempty"`

	// RDMA configures RDMA shared device plugin resources for this NIC.
	// If set, a shared RDMA device resource is advertised to the kubelet.
	// +optional
	RDMA *RDMASpec `json:"rdma,omitempty"`

	// SRIOV configures SR-IOV device plugin resources for this NIC.
	// If set, individual VF resources are advertised to the kubelet.
	// +optional
	SRIOV *SRIOVSpec `json:"sriov,omitempty"`
}

// RDMASpec configures RDMA shared device resources.
type RDMASpec struct {
	// ResourceName is the Kubernetes extended resource name.
	// Advertised as rdma/<ResourceName> (e.g., "rdma_shared_device_a").
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_]+$`
	ResourceName string `json:"resourceName"`

	// SharedCount is how many pods can share this RDMA device simultaneously.
	// +kubebuilder:default=64
	// +kubebuilder:validation:Minimum=1
	SharedCount int `json:"sharedCount,omitempty"`
}

// SRIOVSpec configures SR-IOV Virtual Function resources.
type SRIOVSpec struct {
	// TotalVFs is the number of simulated Virtual Functions.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=128
	TotalVFs int `json:"totalVfs"`

	// ResourceName is the Kubernetes extended resource name (e.g., "hostdev").
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_]+$`
	ResourceName string `json:"resourceName"`

	// ResourcePrefix overrides the default resource domain.
	// +kubebuilder:default="nvidia.com"
	ResourcePrefix string `json:"resourcePrefix,omitempty"`

	// Driver is the simulated VF driver name (e.g., "mlx5_core").
	// +kubebuilder:default="mlx5_core"
	Driver string `json:"driver,omitempty"`
}

// ComponentsSpec toggles individual FNO sub-components.
type ComponentsSpec struct {
	NicLabeler        *ComponentToggle `json:"nicLabeler,omitempty"`
	RDMADevicePlugin  *ComponentToggle `json:"rdmaDevicePlugin,omitempty"`
	SRIOVDevicePlugin *ComponentToggle `json:"sriovDevicePlugin,omitempty"`
	SRIOVConfigDaemon *ComponentToggle `json:"sriovConfigDaemon,omitempty"`
	StatusExporter    *ComponentToggle `json:"statusExporter,omitempty"`
	CLIInjector       *ComponentToggle `json:"cliInjector,omitempty"`
	IBKubernetes      *ComponentToggle `json:"ibKubernetes,omitempty"`
	UFMStub           *ComponentToggle `json:"ufmStub,omitempty"`
	AdmissionWebhook  *ComponentToggle `json:"admissionWebhook,omitempty"`
}

// ComponentToggle controls whether a component is deployed and with what image.
type ComponentToggle struct {
	// Deploy enables or disables this component.
	Deploy bool `json:"deploy"`

	// Image overrides the default container image for this component.
	// +optional
	Image string `json:"image,omitempty"`
}

// SecondaryNetworkSpec defines a NetworkAttachmentDefinition to be generated.
type SecondaryNetworkSpec struct {
	// Name of the NetworkAttachmentDefinition to generate.
	Name string `json:"name"`

	// Type is the CNI delegate plugin to use: "macvlan", "ipvlan", "bridge".
	// +kubebuilder:validation:Enum=macvlan;ipvlan;bridge
	Type string `json:"type"`

	// Master interface on the host to attach to (for macvlan/ipvlan).
	// +kubebuilder:default="eth0"
	Master string `json:"master,omitempty"`

	// IPAMRange is the CIDR range for Whereabouts IPAM (e.g., "192.168.2.0/24").
	// +optional
	IPAMRange string `json:"ipamRange,omitempty"`

	// Namespace to create the NAD in.
	// +kubebuilder:default="default"
	Namespace string `json:"namespace,omitempty"`
}

// FakeNicClusterPolicyStatus defines the observed state of FakeNicClusterPolicy
type FakeNicClusterPolicyStatus struct {
	// State is the overall reconciliation state: "ready", "notReady", "error"
	// +kubebuilder:validation:Enum=ready;notReady;error
	State string `json:"state,omitempty"`

	// Reason provides human-readable detail for non-ready states.
	// +optional
	Reason string `json:"reason,omitempty"`

	// AppliedNodePools tracks which pools are active and on how many nodes.
	// +optional
	AppliedNodePools map[string]NodePoolStatus `json:"appliedNodePools,omitempty"`

	// ComponentStatuses tracks readiness of each sub-component.
	// +optional
	ComponentStatuses map[string]string `json:"componentStatuses,omitempty"`
}

// NodePoolStatus reports the readiness of a single node pool.
type NodePoolStatus struct {
	ReadyNodes int `json:"readyNodes"`
	TotalNodes int `json:"totalNodes"`
}

// +kubebuilder:object:root=true

// FakeNicClusterPolicyList contains a list of FakeNicClusterPolicy
type FakeNicClusterPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FakeNicClusterPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FakeNicClusterPolicy{}, &FakeNicClusterPolicyList{})
}
