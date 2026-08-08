package clioutput

import (
	"fmt"
	"strings"

	"github.com/fake-network-operator/fake-network-operator/internal/topology"
)

// RenderIbvDevinfo renders output matching real ibv_devinfo format.
func RenderIbvDevinfo(topo *topology.NodeNicTopology) string {
	var sb strings.Builder
	for _, nic := range topo.NICs {
		sb.WriteString(fmt.Sprintf("hca_id:\t%s\n", nic.Name))
		sb.WriteString("\ttransport:\t\t\tInfiniBand (0)\n")
		sb.WriteString(fmt.Sprintf("\tfw_ver:\t\t\t\t%s\n", nic.FirmwareVersion))
		sb.WriteString(fmt.Sprintf("\tnode_guid:\t\t\t%s\n", nic.NodeGUID))
		sb.WriteString(fmt.Sprintf("\tsys_image_guid:\t\t\t%s\n", nic.SysImageGUID))
		sb.WriteString(fmt.Sprintf("\tvendor_id:\t\t\t0x%s\n", nic.Vendor))
		sb.WriteString(fmt.Sprintf("\tvendor_part_id:\t\t\t%s\n", nic.DeviceID))
		sb.WriteString("\thw_ver:\t\t\t\t0x0\n")
		sb.WriteString(fmt.Sprintf("\tboard_id:\t\t\t%s\n", nic.BoardID))
		sb.WriteString("\tphys_port_cnt:\t\t\t1\n")
		sb.WriteString("\t\tport:\t1\n")
		if nic.State == "Active" {
			sb.WriteString("\t\t\tstate:\t\t\tPORT_ACTIVE (4)\n")
		} else {
			sb.WriteString("\t\t\tstate:\t\t\tPORT_DOWN (1)\n")
		}
		sb.WriteString("\t\t\tmax_mtu:\t\t4096 (5)\n")
		sb.WriteString("\t\t\tactive_mtu:\t\t4096 (5)\n")
		sb.WriteString("\t\t\tsm_lid:\t\t\t2\n")
		sb.WriteString("\t\t\tport_lid:\t\t1\n")
		sb.WriteString("\t\t\tport_lmc:\t\t0x00\n")
		sb.WriteString(fmt.Sprintf("\t\t\tlink_layer:\t\t%s\n", nic.LinkType))
	}
	return sb.String()
}
