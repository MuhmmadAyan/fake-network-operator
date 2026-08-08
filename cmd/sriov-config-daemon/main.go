package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fake-network-operator/fake-network-operator/internal/sriov"
	"github.com/fake-network-operator/fake-network-operator/internal/topology"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	topologyNamespace string
	nodeName          string
	drainDelay        time.Duration
	configDelay       time.Duration
	failureRate       float64
	setupLog          = ctrl.Log.WithName("setup")
)

func init() {
	flag.StringVar(&topologyNamespace, "topology-namespace", "fake-network-operator", "Namespace where the topology ConfigMap is located.")
	flag.StringVar(&nodeName, "node-name", "", "Name of the node this daemon is running on.")
	flag.DurationVar(&drainDelay, "drain-delay", 5*time.Second, "Simulated delay for draining a node.")
	flag.DurationVar(&configDelay, "config-delay", 10*time.Second, "Simulated delay for configuring SR-IOV.")
	flag.Float64Var(&failureRate, "failure-rate", 0.0, "Simulated failure rate for configuration (0.0 to 1.0).")
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

	setupLog.Info("starting sriov-config-daemon", "node", nodeName)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	_, err := topology.ReadConfigMap(ctx, topologyNamespace, nodeName)
	if err != nil {
		setupLog.Error(err, "failed to read initial topology ConfigMap")
		os.Exit(1)
	}

	stateMachine := sriov.NewStateMachine(nodeName, drainDelay, configDelay, failureRate)

	go func() {
		if err := stateMachine.Run(ctx); err != nil {
			setupLog.Error(err, "state machine error")
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, err := topology.ReadConfigMap(ctx, topologyNamespace, nodeName)
				if err != nil {
					setupLog.Error(err, "failed to read topology ConfigMap during watch")
				}
			}
		}
	}()

	<-ctx.Done()
	setupLog.Info("shutting down sriov-config-daemon")
}
