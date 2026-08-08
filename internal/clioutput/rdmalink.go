package clioutput

import (
	"fmt"
	"strings"

	"github.com/fake-network-operator/fake-network-operator/internal/topology"
)

// RenderRdmaLink renders output matching real 'rdma link' command format.
func RenderRdmaLink(topo *topology.NodeNicTopology) string {
	var sb strings.Builder
	for i, nic := range topo.NICs {
		sb.WriteString(fmt.Sprintf("link %s/%d state %s physical_state %s netdev %s\n",
			nic.Name, i+1, nic.State, nic.PhysicalState, nic.Name))
	}
	return sb.String()
}
