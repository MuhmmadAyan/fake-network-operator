package topology

import (
	"testing"

	"github.com/fake-network-operator/fake-network-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestBuildNodeTopology_Default(t *testing.T) {
	spec := v1alpha1.NicNodePoolSpec{
		GPUDirectCapable: true,
	}
	topo := BuildNodeTopology("node-1", "pool-1", spec)

	assert.NotNil(t, topo)
	assert.Equal(t, "node-1", topo.NodeName)
	assert.Equal(t, "pool-1", topo.NodePool)
	assert.True(t, topo.GPUDirectCapable)
	assert.Len(t, topo.NICs, 2)

	// Verify first NIC (mlx5_0 - InfiniBand RDMA)
	nic0 := topo.NICs[0]
	assert.Equal(t, "mlx5_0", nic0.Name)
	assert.Equal(t, "infiniband", nic0.LinkType)
	assert.Equal(t, "15b3", nic0.Vendor)
	assert.Equal(t, "101b", nic0.DeviceID)
	assert.Equal(t, 0, nic0.NUMANode)
	assert.Equal(t, "100 Gb/sec", nic0.Speed)
	assert.NotNil(t, nic0.RDMA)
	assert.Equal(t, "rdma_shared_device_a", nic0.RDMA.ResourceName)
	assert.Equal(t, 64, nic0.RDMA.SharedCount)
	assert.Nil(t, nic0.SRIOV)

	// Verify second NIC (mlx5_1 - Ethernet SR-IOV)
	nic1 := topo.NICs[1]
	assert.Equal(t, "mlx5_1", nic1.Name)
	assert.Equal(t, "ethernet", nic1.LinkType)
	assert.Equal(t, "15b3", nic1.Vendor)
	assert.Equal(t, "101b", nic1.DeviceID)
	assert.Equal(t, 1, nic1.NUMANode)
	assert.Equal(t, "100 Gb/sec", nic1.Speed)
	assert.Nil(t, nic1.RDMA)
	assert.NotNil(t, nic1.SRIOV)
	assert.Equal(t, 8, nic1.SRIOV.TotalVFs)
	assert.Equal(t, "hostdev", nic1.SRIOV.ResourceName)
	assert.Equal(t, "nvidia.com", nic1.SRIOV.ResourcePrefix)
	assert.Len(t, nic1.SRIOV.VFs, 8)
}

func TestBuildNodeTopology_Explicit(t *testing.T) {
	spec := v1alpha1.NicNodePoolSpec{
		NICs: []v1alpha1.NicSpec{
			{
				Name:            "custom_nic",
				LinkType:        "ethernet",
				Vendor:          "1234",
				DeviceID:        "5678",
				FirmwareVersion: "1.2.3",
				BoardID:         "MY_BOARD",
				NUMANode:        2,
				Speed:           "200 Gb/sec",
			},
		},
	}
	topo := BuildNodeTopology("node-2", "pool-2", spec)

	assert.NotNil(t, topo)
	assert.Len(t, topo.NICs, 1)

	nic := topo.NICs[0]
	assert.Equal(t, "custom_nic", nic.Name)
	assert.Equal(t, "ethernet", nic.LinkType)
	assert.Equal(t, "1234", nic.Vendor)
	assert.Equal(t, "5678", nic.DeviceID)
	assert.Equal(t, "1.2.3", nic.FirmwareVersion)
	assert.Equal(t, "MY_BOARD", nic.BoardID)
	assert.Equal(t, 2, nic.NUMANode)
	assert.Equal(t, "200 Gb/sec", nic.Speed)
	assert.Equal(t, "Active", nic.State)
	assert.Equal(t, "LinkUp", nic.PhysicalState)
	assert.NotEmpty(t, nic.PortGUID)
	assert.NotEmpty(t, nic.NodeGUID)
	assert.NotEmpty(t, nic.SysImageGUID)
}

func TestDeterministicGUIDs(t *testing.T) {
	guid1 := GenerateGUID("test-seed-1")
	guid2 := GenerateGUID("test-seed-1")
	guid3 := GenerateGUID("test-seed-2")

	assert.Equal(t, guid1, guid2)
	assert.NotEqual(t, guid1, guid3)

	vfid1 := GenerateVFID("node-1", "mlx5_0", 0)
	vfid2 := GenerateVFID("node-1", "mlx5_0", 0)
	vfid3 := GenerateVFID("node-1", "mlx5_0", 1)
	vfid4 := GenerateVFID("node-2", "mlx5_0", 0)

	assert.Equal(t, vfid1, vfid2)
	assert.NotEqual(t, vfid1, vfid3)
	assert.NotEqual(t, vfid1, vfid4)
}
