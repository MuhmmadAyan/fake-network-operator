// Package topology handles the per-node NIC topology ConfigMap — the single
// source of truth that every FNO DaemonSet component reads.
package topology

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/fake-network-operator/fake-network-operator/internal/constants"
)

const (
	// TopologyDataKey is the key inside the ConfigMap that holds the JSON payload.
	TopologyDataKey = "topology.json"
)

// NodeNicTopology is the JSON payload stored in the per-node ConfigMap.
// Every fake component on a node reads this to know what to simulate.
type NodeNicTopology struct {
	// NodeName is the Kubernetes node name.
	NodeName string `json:"nodeName"`

	// NodePool is the pool this node belongs to.
	NodePool string `json:"nodePool"`

	// NICs is the list of simulated NICs on this node.
	NICs []NicTopologyEntry `json:"nics"`

	// GPUDirectCapable indicates GPUDirect co-location eligibility.
	GPUDirectCapable bool `json:"gpuDirectCapable"`
}

// NicTopologyEntry describes a single simulated NIC as materialized on a node.
type NicTopologyEntry struct {
	Name            string `json:"name"`
	LinkType        string `json:"linkType"` // "ethernet" | "infiniband"
	Vendor          string `json:"vendor"`
	DeviceID        string `json:"deviceID"`
	FirmwareVersion string `json:"firmwareVersion"`
	BoardID         string `json:"boardID"`
	NUMANode        int    `json:"numaNode"`
	Speed           string `json:"speed"`
	State           string `json:"state"`         // "Active" | "Down"
	PhysicalState   string `json:"physicalState"` // "LinkUp" | "Polling"

	// GUIDs — deterministically generated from node name + NIC index
	PortGUID     string `json:"portGUID"`
	NodeGUID     string `json:"nodeGUID"`
	SysImageGUID string `json:"sysImageGUID"`

	// For RDMA shared device plugin
	RDMA *RDMATopology `json:"rdma,omitempty"`

	// For SR-IOV device plugin
	SRIOV *SRIOVTopology `json:"sriov,omitempty"`
}

// RDMATopology holds RDMA-specific topology for a NIC.
type RDMATopology struct {
	ResourceName string `json:"resourceName"`
	SharedCount  int    `json:"sharedCount"`
}

// SRIOVTopology holds SR-IOV-specific topology for a NIC.
type SRIOVTopology struct {
	TotalVFs       int               `json:"totalVfs"`
	ResourceName   string            `json:"resourceName"`
	ResourcePrefix string            `json:"resourcePrefix"`
	VFs            []VirtualFunction `json:"vfs"`
}

// VirtualFunction describes a single simulated SR-IOV VF.
type VirtualFunction struct {
	ID       string `json:"id"` // Deterministic synthetic UUID
	VFIndex  int    `json:"vfIndex"`
	Driver   string `json:"driver"`
	NUMANode int    `json:"numaNode"`
}

// ConfigMapName returns the name of the topology ConfigMap for a given node.
func ConfigMapName(nodeName string) string {
	return constants.TopologyConfigMapPrefix + nodeName
}

// GenerateGUID generates a deterministic GUID from a seed string.
// The result is formatted as a 16-hex-char colon-separated string
// like "50:6b:4b:03:00:f1:a2:b4".
func GenerateGUID(seed string) string {
	h := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x",
		h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7])
}

// GenerateGUIDHex generates a deterministic GUID as a 0x-prefixed hex string
// like "0x506b4b0300f1a2b4" (used for ibstat output).
func GenerateGUIDHex(seed string) string {
	h := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("0x%02x%02x%02x%02x%02x%02x%02x%02x",
		h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7])
}

// GenerateVFID generates a deterministic VF identifier.
func GenerateVFID(nodeName, nicName string, vfIndex int) string {
	seed := fmt.Sprintf("%s/%s/vf/%d", nodeName, nicName, vfIndex)
	h := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uint32(h[0])<<24|uint32(h[1])<<16|uint32(h[2])<<8|uint32(h[3]),
		uint16(h[4])<<8|uint16(h[5]),
		uint16(h[6])<<8|uint16(h[7]),
		uint16(h[8])<<8|uint16(h[9]),
		uint64(h[10])<<40|uint64(h[11])<<32|uint64(h[12])<<24|uint64(h[13])<<16|uint64(h[14])<<8|uint64(h[15]),
	)
}

// WriteConfigMap creates or updates the per-node topology ConfigMap.
func WriteConfigMap(ctx context.Context, c client.Client, namespace string, topo *NodeNicTopology) error {
	data, err := json.MarshalIndent(topo, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling topology: %w", err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName(topo.NodeName),
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":  "fake-network-operator",
				"app.kubernetes.io/component":   "topology",
				"fake-network-operator.io/node": topo.NodeName,
				"fake-network-operator.io/pool": topo.NodePool,
			},
		},
		Data: map[string]string{
			TopologyDataKey: string(data),
		},
	}

	// Try to get existing ConfigMap
	existing := &corev1.ConfigMap{}
	err = c.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, existing)
	if err != nil {
		// Create new
		return c.Create(ctx, cm)
	}

	// Update existing
	existing.Data = cm.Data
	existing.Labels = cm.Labels
	return c.Update(ctx, existing)
}

// ReadConfigMapWithClient reads the topology ConfigMap using an existing client.
func ReadConfigMapWithClient(ctx context.Context, c client.Client, namespace, nodeName string) (*NodeNicTopology, error) {
	cm := &corev1.ConfigMap{}
	err := c.Get(ctx, types.NamespacedName{
		Name:      ConfigMapName(nodeName),
		Namespace: namespace,
	}, cm)
	if err != nil {
		return nil, fmt.Errorf("getting topology ConfigMap for node %s: %w", nodeName, err)
	}

	data, ok := cm.Data[TopologyDataKey]
	if !ok {
		return nil, fmt.Errorf("topology ConfigMap for node %s missing key %s", nodeName, TopologyDataKey)
	}

	topo := &NodeNicTopology{}
	if err := json.Unmarshal([]byte(data), topo); err != nil {
		return nil, fmt.Errorf("unmarshaling topology for node %s: %w", nodeName, err)
	}

	return topo, nil
}

// ReadConfigMap reads the per-node topology ConfigMap by building an in-cluster client.
func ReadConfigMap(ctx context.Context, namespace, nodeName string) (*NodeNicTopology, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return ReadConfigMapWithClient(ctx, c, namespace, nodeName)
}

// ReadConfigMapFromObject reads and parses the topology directly from a ConfigMap object.
func ReadConfigMapFromObject(cm *corev1.ConfigMap) (*NodeNicTopology, error) {
	data, ok := cm.Data[TopologyDataKey]
	if !ok {
		return nil, fmt.Errorf("missing key %s", TopologyDataKey)
	}
	return ReadFromFileBytes([]byte(data))
}

// ReadFromFile reads topology from a local JSON file (used by CLI shims and
// DaemonSet pods that mount the ConfigMap as a volume).
func ReadFromFile(path string) (*NodeNicTopology, error) {
	var topo NodeNicTopology
	return &topo, fmt.Errorf("use ReadFromFileBytes instead with file content: %s", path)
}

// ReadFromFileBytes deserializes topology from raw JSON bytes.
func ReadFromFileBytes(data []byte) (*NodeNicTopology, error) {
	topo := &NodeNicTopology{}
	if err := json.Unmarshal(data, topo); err != nil {
		return nil, fmt.Errorf("unmarshaling topology from bytes: %w", err)
	}
	return topo, nil
}

// DeleteConfigMap removes the per-node topology ConfigMap.
func DeleteConfigMap(ctx context.Context, c client.Client, namespace string, nodeName string) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName(nodeName),
			Namespace: namespace,
		},
	}
	return client.IgnoreNotFound(c.Delete(ctx, cm))
}
