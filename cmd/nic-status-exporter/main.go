package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fake-network-operator/fake-network-operator/internal/metrics"
	"github.com/fake-network-operator/fake-network-operator/internal/topology"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	topologyNamespace string
	nodeName          string
	metricsPort       string
	setupLog          = ctrl.Log.WithName("setup")
)

func init() {
	flag.StringVar(&topologyNamespace, "topology-namespace", "fake-network-operator", "Namespace where the topology ConfigMap is located.")
	flag.StringVar(&nodeName, "node-name", "", "Name of the node this exporter is running on.")
	flag.StringVar(&metricsPort, "metrics-port", "9394", "Port to serve metrics on.")
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

	setupLog.Info("starting nic-status-exporter", "node", nodeName, "port", metricsPort)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	collector := metrics.NewNicMetricsCollector(topologyNamespace, nodeName)

	http.Handle("/metrics", promhttp.Handler())
	
	go func() {
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil && err != http.ErrServerClosed {
			setupLog.Error(err, "failed to start HTTP server")
			os.Exit(1)
		}
	}()

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				topoClient, err := topology.ReadConfigMap(ctx, topologyNamespace, nodeName)
				if err != nil {
					setupLog.Error(err, "failed to read topology ConfigMap")
					continue
				}
				collector.UpdateFromTopology(topoClient)
			}
		}
	}()

	<-ctx.Done()
	setupLog.Info("shutting down nic-status-exporter")
}
