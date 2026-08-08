package clioutput

import (
	"fmt"
	"strings"

	"github.com/fake-network-operator/fake-network-operator/internal/topology"
)

// RenderIbstat renders output matching real ibstat format for each HCA in topology.
func RenderIbstat(topo *topology.NodeNicTopology) string {
	var sb strings.Builder
	for _, nic := range topo.NICs {
		nodeGuidHex := "0x" + strings.ReplaceAll(nic.NodeGUID, ":", "")
		sysImageGuidHex := "0x" + strings.ReplaceAll(nic.SysImageGUID, ":", "")
		portGuidHex := "0x" + strings.ReplaceAll(nic.PortGUID, ":", "")

		sb.WriteString(fmt.Sprintf("CA '%s'\n", nic.Name))
		sb.WriteString("\tCA type: MT4123\n")
		sb.WriteString("\tNumber of ports: 1\n")
		sb.WriteString(fmt.Sprintf("\tFirmware version: %s\n", nic.FirmwareVersion))
		sb.WriteString("\tHardware version: 0\n")
		sb.WriteString(fmt.Sprintf("\tNode GUID: %s\n", nodeGuidHex))
		sb.WriteString(fmt.Sprintf("\tSystem image GUID: %s\n", sysImageGuidHex))
		sb.WriteString("Port 1:\n")
		sb.WriteString(fmt.Sprintf("\tState: %s\n", nic.State))
		sb.WriteString(fmt.Sprintf("\tPhysical state: %s\n", nic.PhysicalState))
		sb.WriteString(fmt.Sprintf("\tRate: %s\n", nic.Speed))
		sb.WriteString("\tBase lid: 1\n")
		sb.WriteString("\tLMC: 0\n")
		sb.WriteString("\tSM lid: 2\n")
		sb.WriteString("\tCapability mask: 0x2651e848\n")
		sb.WriteString(fmt.Sprintf("\tPort GUID: %s\n", portGuidHex))
		sb.WriteString(fmt.Sprintf("\tLink layer: %s\n", nic.LinkType))
	}
	return sb.String()
}
