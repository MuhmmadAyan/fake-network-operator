package main

import (
	"context"
	"flag"
	"os"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/fake-network-operator/fake-network-operator/internal/labeler"
	"github.com/fake-network-operator/fake-network-operator/internal/topology"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

// NicLabelerReconciler watches the per-node topology ConfigMap and patches
// the node with appropriate NIC/NFD labels.
type NicLabelerReconciler struct {
	client.Client
	NodeName  string
	Namespace string
}

// Reconcile handles ConfigMap events, filtering for this node's topology ConfigMap.
func (r *NicLabelerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Only process this node's topology ConfigMap
	expectedName := topology.ConfigMapName(r.NodeName)
	if req.Name != expectedName || req.Namespace != r.Namespace {
		return ctrl.Result{}, nil
	}

	cm := &corev1.ConfigMap{}
	if err := r.Get(ctx, req.NamespacedName, cm); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ConfigMap deleted, cleaning labels")
			if err := labeler.CleanNodeLabels(ctx, r.Client, r.NodeName); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Parse the topology from the ConfigMap
	if _, ok := cm.Data[topology.TopologyDataKey]; !ok {
		logger.Error(nil, "topology ConfigMap missing data key")
		return ctrl.Result{}, nil
	}

	topo, err := topology.ReadConfigMapFromObject(cm)
	if err != nil {
		logger.Error(err, "failed to parse topology from ConfigMap")
		return ctrl.Result{}, err
	}

	labels := labeler.BuildLabels(topo)
	if err := labeler.PatchNodeLabels(ctx, r.Client, r.NodeName, labels); err != nil {
		logger.Error(err, "failed to patch node labels")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully patched node labels")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NicLabelerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}

func main() {
	var topologyNamespace string
	var nodeName string

	flag.StringVar(&topologyNamespace, "topology-namespace", "fake-network-operator", "The namespace to watch for topology ConfigMaps")
	flag.StringVar(&nodeName, "node-name", "", "The name of the node this labeler is running on")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if nodeName == "" {
		nodeName = os.Getenv("NODE_NAME")
	}
	if nodeName == "" {
		setupLog.Error(nil, "node-name flag or NODE_NAME env var is required")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&NicLabelerReconciler{
		Client:    mgr.GetClient(),
		NodeName:  nodeName,
		Namespace: topologyNamespace,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NicLabeler")
		os.Exit(1)
	}

	setupLog.Info("starting nic-labeler", "node", nodeName)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
