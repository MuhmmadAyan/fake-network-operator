package clioutput

import (
	"strings"
	"testing"

	"github.com/fake-network-operator/fake-network-operator/internal/topology"
	"github.com/stretchr/testify/assert"
)

func TestGenerateIbstatOutput(t *testing.T) {
	topo := &topology.NodeNicTopology{
		NICs: []topology.NicTopologyEntry{
			{
				Name:            "mlx5_0",
				FirmwareVersion: "1.2.3",
				NodeGUID:        "11:22:33:44:55:66:77:88",
				SysImageGUID:    "aa:bb:cc:dd:ee:ff:11:22",
				PortGUID:        "88:77:66:55:44:33:22:11",
				State:           "Active",
				PhysicalState:   "LinkUp",
				Speed:           "100 Gb/sec",
				LinkType:        "infiniband",
			},
		},
	}

	output := RenderIbstat(topo)

	assert.Contains(t, output, "CA 'mlx5_0'")
	assert.Contains(t, output, "Firmware version: 1.2.3")
	assert.Contains(t, output, "Node GUID: 0x1122334455667788")
	assert.Contains(t, output, "System image GUID: 0xaabbccddeeff1122")
	assert.Contains(t, output, "Port GUID: 0x8877665544332211")
	assert.Contains(t, output, "State: Active")
	assert.Contains(t, output, "Physical state: LinkUp")
	assert.Contains(t, output, "Rate: 100 Gb/sec")
	assert.Contains(t, output, "Link layer: infiniband")
}

func TestGenerateIbvDevinfoOutput(t *testing.T) {
	topo := &topology.NodeNicTopology{
		NICs: []topology.NicTopologyEntry{
			{
				Name:            "mlx5_1",
				FirmwareVersion: "4.5.6",
				NodeGUID:        "11:22:33:44:55:66:77:88",
				SysImageGUID:    "aa:bb:cc:dd:ee:ff:11:22",
				Vendor:          "15b3",
				DeviceID:        "101b",
				BoardID:         "MY_BOARD_ID",
				State:           "Active",
				LinkType:        "ethernet",
			},
		},
	}

	output := RenderIbvDevinfo(topo)

	assert.Contains(t, output, "hca_id:\tmlx5_1")
	assert.Contains(t, output, "fw_ver:\t\t\t\t4.5.6")
	assert.Contains(t, output, "node_guid:\t\t\t11:22:33:44:55:66:77:88")
	assert.Contains(t, output, "sys_image_guid:\t\t\taa:bb:cc:dd:ee:ff:11:22")
	assert.Contains(t, output, "vendor_id:\t\t\t0x15b3")
	assert.Contains(t, output, "vendor_part_id:\t\t\t101b")
	assert.Contains(t, output, "board_id:\t\t\tMY_BOARD_ID")
	assert.Contains(t, output, "state:\t\t\tPORT_ACTIVE (4)")
	assert.Contains(t, output, "link_layer:\t\tethernet")
}

func TestGenerateRdmaLinkOutput(t *testing.T) {
	topo := &topology.NodeNicTopology{
		NICs: []topology.NicTopologyEntry{
			{
				Name:          "mlx5_0",
				State:         "Active",
				PhysicalState: "LinkUp",
			},
			{
				Name:          "mlx5_1",
				State:         "Down",
				PhysicalState: "Polling",
			},
		},
	}

	output := RenderRdmaLink(topo)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 2)
	assert.Equal(t, "link mlx5_0/1 state Active physical_state LinkUp netdev mlx5_0", lines[0])
	assert.Equal(t, "link mlx5_1/2 state Down physical_state Polling netdev mlx5_1", lines[1])
}
