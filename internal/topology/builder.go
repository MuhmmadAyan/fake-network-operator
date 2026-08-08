package topology

import (
	"fmt"

	"github.com/fake-network-operator/fake-network-operator/api/v1alpha1"
)

// BuildNodeTopology converts a NicNodePoolSpec into a NodeNicTopology payload.
// If spec.NICs is empty, it automatically synthesizes standard dual-port Mellanox ConnectX-6 hardware.
// It generates deterministic GUIDs, BoardIDs, VF IDs, and sets defaults.
func BuildNodeTopology(nodeName string, poolName string, spec v1alpha1.NicNodePoolSpec) *NodeNicTopology {
	topo := &NodeNicTopology{
		NodeName:         nodeName,
		NodePool:         poolName,
		GPUDirectCapable: spec.GPUDirectCapable,
	}

	nicsToBuild := spec.NICs

	// If no explicit NICs defined in policy, synthesize standard Mellanox ConnectX-6 dual-port topology
	if len(nicsToBuild) == 0 {
		rdmaResName := "rdma_shared_device_a"
		rdmaCount := 64
		if spec.RDMA != nil {
			if spec.RDMA.ResourceName != "" {
				rdmaResName = spec.RDMA.ResourceName
			}
			if spec.RDMA.SharedCount > 0 {
				rdmaCount = spec.RDMA.SharedCount
			}
		}

		sriovVfs := 8
		sriovResName := "hostdev"
		sriovPrefix := "nvidia.com"
		sriovDriver := "mlx5_core"
		if spec.SRIOV != nil {
			if spec.SRIOV.TotalVFs > 0 {
				sriovVfs = spec.SRIOV.TotalVFs
			}
			if spec.SRIOV.ResourceName != "" {
				sriovResName = spec.SRIOV.ResourceName
			}
			if spec.SRIOV.ResourcePrefix != "" {
				sriovPrefix = spec.SRIOV.ResourcePrefix
			}
			if spec.SRIOV.Driver != "" {
				sriovDriver = spec.SRIOV.Driver
			}
		}

		nicsToBuild = []v1alpha1.NicSpec{
			{
				Name:            "mlx5_0",
				LinkType:        "infiniband",
				Vendor:          "15b3",
				DeviceID:        "101b",
				FirmwareVersion: "20.31.1014",
				BoardID:         fmt.Sprintf("MT_%s_0", nodeName),
				NUMANode:        0,
				Speed:           "100 Gb/sec",
				RDMA: &v1alpha1.RDMASpec{
					ResourceName: rdmaResName,
					SharedCount:  rdmaCount,
				},
			},
			{
				Name:            "mlx5_1",
				LinkType:        "ethernet",
				Vendor:          "15b3",
				DeviceID:        "101b",
				FirmwareVersion: "20.31.1014",
				BoardID:         fmt.Sprintf("MT_%s_1", nodeName),
				NUMANode:        1,
				Speed:           "100 Gb/sec",
				SRIOV: &v1alpha1.SRIOVSpec{
					TotalVFs:       sriovVfs,
					ResourceName:   sriovResName,
					ResourcePrefix: sriovPrefix,
					Driver:         sriovDriver,
				},
			},
		}
	}

	for _, nicSpec := range nicsToBuild {
		vendor := nicSpec.Vendor
		if vendor == "" {
			vendor = "15b3"
		}
		deviceID := nicSpec.DeviceID
		if deviceID == "" {
			deviceID = "101b"
		}
		fwVer := nicSpec.FirmwareVersion
		if fwVer == "" {
			fwVer = "20.31.1014"
		}
		boardID := nicSpec.BoardID
		if boardID == "" {
			boardID = fmt.Sprintf("MT_%s_%s", nodeName, nicSpec.Name)
		}
		speed := nicSpec.Speed
		if speed == "" {
			speed = "100 Gb/sec"
		}

		entry := NicTopologyEntry{
			Name:            nicSpec.Name,
			LinkType:        nicSpec.LinkType,
			Vendor:          vendor,
			DeviceID:        deviceID,
			FirmwareVersion: fwVer,
			BoardID:         boardID,
			NUMANode:        nicSpec.NUMANode,
			Speed:           speed,
			State:           "Active",
			PhysicalState:   "LinkUp",
			PortGUID:        GenerateGUID(nodeName + nicSpec.Name + "port"),
			NodeGUID:        GenerateGUID(nodeName + nicSpec.Name + "node"),
			SysImageGUID:    GenerateGUID(nodeName + "sysimage"),
		}

		if nicSpec.RDMA != nil {
			entry.RDMA = &RDMATopology{
				ResourceName: nicSpec.RDMA.ResourceName,
				SharedCount:  nicSpec.RDMA.SharedCount,
			}
		}

		if nicSpec.SRIOV != nil {
			entry.SRIOV = &SRIOVTopology{
				TotalVFs:       nicSpec.SRIOV.TotalVFs,
				ResourceName:   nicSpec.SRIOV.ResourceName,
				ResourcePrefix: nicSpec.SRIOV.ResourcePrefix,
			}
			for j := 0; j < nicSpec.SRIOV.TotalVFs; j++ {
				entry.SRIOV.VFs = append(entry.SRIOV.VFs, VirtualFunction{
					ID:       GenerateVFID(nodeName, nicSpec.Name, j),
					VFIndex:  j,
					Driver:   nicSpec.SRIOV.Driver,
					NUMANode: nicSpec.NUMANode,
				})
			}
		}

		topo.NICs = append(topo.NICs, entry)
	}

	return topo
}
