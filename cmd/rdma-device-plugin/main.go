package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fake-network-operator/fake-network-operator/internal/deviceplugin"
	"github.com/fake-network-operator/fake-network-operator/internal/topology"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	topologyNamespace string
	nodeName          string
	kubeletSocket     string
	setupLog          = ctrl.Log.WithName("setup")
)

func init() {
	flag.StringVar(&topologyNamespace, "topology-namespace", "fake-network-operator", "Namespace where the topology ConfigMap is located.")
	flag.StringVar(&nodeName, "node-name", "", "Name of the node this plugin is running on.")
	flag.StringVar(&kubeletSocket, "kubelet-socket", "/var/lib/kubelet/device-plugins/kubelet.sock", "Path to kubelet socket.")
}

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if nodeName == "" {
		setupLog.Error(nil, "--node-name flag is required")
		os.Exit(1)
	}

	setupLog.Info("starting rdma-device-plugin", "node", nodeName)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	topoClient, err := topology.ReadConfigMap(ctx, topologyNamespace, nodeName)
	if err != nil {
		setupLog.Error(err, "failed to read topology ConfigMap")
		os.Exit(1)
	}

	for _, nicTopo := range topoClient.NICs {
		if nicTopo.RDMA != nil {
			var deviceIDs []string
			for i := 0; i < nicTopo.RDMA.SharedCount; i++ {
				deviceIDs = append(deviceIDs, fmt.Sprintf("%s-rdma-%d", nicTopo.Name, i))
			}

			resName := nicTopo.RDMA.ResourceName
			if !strings.Contains(resName, "/") {
				resName = "rdma/" + resName
			}

			server := deviceplugin.NewServer(resName, deviceIDs, "RDMA_DEVICE")
			socketFileName := strings.ReplaceAll(resName, "/", "_") + ".sock"
			socketPath := filepath.Join("/var/lib/kubelet/device-plugins", socketFileName)

			go func(socketPath string, resourceName string) {
				setupLog.Info("serving device plugin", "resource", resourceName)
				if err := deviceplugin.Serve(server, socketPath); err != nil {
					setupLog.Error(err, "failed to serve device plugin", "resource", resourceName)
				}
			}(socketPath, resName)

			setupLog.Info("registering device plugin", "resource", resName)
			if err := deviceplugin.Register(kubeletSocket, resName, socketPath); err != nil {
				setupLog.Error(err, "failed to register device plugin", "resource", resName)
			}
		}
	}

	<-ctx.Done()
	setupLog.Info("shutting down rdma-device-plugin")
}
