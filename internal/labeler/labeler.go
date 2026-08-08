package labeler

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fake-network-operator/fake-network-operator/internal/constants"
	"github.com/fake-network-operator/fake-network-operator/internal/topology"
)

// BuildLabels produces all node labels from a topology.
func BuildLabels(topo *topology.NodeNicTopology) map[string]string {
	labels := make(map[string]string)

	labels[constants.NicReadyLabelKey] = "true"
	if topo.GPUDirectCapable {
		labels[constants.GPUDirectCapableLabelKey] = "true"
	}

	for _, nic := range topo.NICs {
		if nic.Vendor == "15b3" {
			labels[constants.PCIMellanoxPresentLabel] = "true"
		}
		if nic.SRIOV != nil {
			labels[constants.PCISRIOVCapableLabel] = "true"
			labels[constants.NicSRIOVCapableLabel] = "true"
		}
		if nic.RDMA != nil {
			labels[constants.NicRDMACapableLabel] = "true"
		}
		if product, ok := constants.DeviceIDToProduct[nic.DeviceID]; ok {
			labels[constants.NicProductLabel] = product
		}
		if nic.FirmwareVersion != "" {
			labels[constants.NicFirmwareVersionLabel] = nic.FirmwareVersion
		}
		labels[constants.NicDriverVersionLabel] = constants.DefaultOFEDVersion
	}

	return labels
}

// PatchNodeLabels patches the node labels onto the kubernetes Node.
func PatchNodeLabels(ctx context.Context, c client.Client, nodeName string, labels map[string]string) error {
	node := &corev1.Node{}
	if err := c.Get(ctx, client.ObjectKey{Name: nodeName}, node); err != nil {
		return err
	}

	patch := client.MergeFrom(node.DeepCopy())

	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	for k, v := range labels {
		node.Labels[k] = v
	}

	return c.Patch(ctx, node, patch)
}

// CleanNodeLabels removes FNO-managed labels from the node.
func CleanNodeLabels(ctx context.Context, c client.Client, nodeName string) error {
	node := &corev1.Node{}
	if err := c.Get(ctx, client.ObjectKey{Name: nodeName}, node); err != nil {
		return client.IgnoreNotFound(err)
	}

	patch := client.MergeFrom(node.DeepCopy())

	toRemove := []string{
		constants.NicReadyLabelKey,
		constants.GPUDirectCapableLabelKey,
		constants.PCIMellanoxPresentLabel,
		constants.PCISRIOVCapableLabel,
		constants.NicProductLabel,
		constants.NicFirmwareVersionLabel,
		constants.NicDriverVersionLabel,
		constants.NicRDMACapableLabel,
		constants.NicSRIOVCapableLabel,
	}

	if node.Labels != nil {
		for _, key := range toRemove {
			delete(node.Labels, key)
		}
	}

	return c.Patch(ctx, node, patch)
}
