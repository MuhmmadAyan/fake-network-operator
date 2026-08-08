package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fake-network-operator/fake-network-operator/internal/clioutput"
	"github.com/fake-network-operator/fake-network-operator/internal/topology"
)

func main() {
	if len(os.Args) < 1 {
		fmt.Fprintln(os.Stderr, "no command specified")
		os.Exit(1)
	}

	cmdName := filepath.Base(os.Args[0])
	topoPath := "/etc/fake-network-operator/topology.json"

	topoData, err := os.ReadFile(topoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read topology file: %v\n", err)
		os.Exit(1)
	}

	topoNode, err := topology.ReadFromFileBytes(topoData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse topology: %v\n", err)
		os.Exit(1)
	}

	switch cmdName {
	case "ibstat":
		fmt.Print(clioutput.RenderIbstat(topoNode))
	case "ibv_devinfo":
		fmt.Print(clioutput.RenderIbvDevinfo(topoNode))
	case "ibv_devices":
		fmt.Println("mock ibv_devices output")
	case "rdma":
		fmt.Print(clioutput.RenderRdmaLink(topoNode))
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmdName)
		os.Exit(1)
	}
}
