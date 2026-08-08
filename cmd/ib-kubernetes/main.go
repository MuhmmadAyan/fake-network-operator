package main

import (
	"context"
	"flag"
	"os"

	"github.com/fake-network-operator/fake-network-operator/internal/ibkubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

type PodReconciler struct {
	client.Client
	Allocator *ibkubernetes.GUIDAllocator
}

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		r.Allocator.Release(req.NamespacedName.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !pod.DeletionTimestamp.IsZero() {
		r.Allocator.Release(req.NamespacedName.String())
		return ctrl.Result{}, nil
	}

	if _, ok := pod.Annotations["k8s.v1.cni.cncf.io/networks"]; ok {
		guid, err := r.Allocator.Allocate(req.NamespacedName.String())
		if err != nil {
			return ctrl.Result{}, err
		}

		if pod.Annotations["mellanox.infiniband.app/guid"] != guid {
			if pod.Annotations == nil {
				pod.Annotations = make(map[string]string)
			}
			pod.Annotations["mellanox.infiniband.app/guid"] = guid
			if err := r.Update(ctx, &pod); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}

func main() {
	var guidRangeStart string
	var guidRangeEnd string
	var namespace string

	flag.StringVar(&guidRangeStart, "guid-range-start", "02:00:00:00:00:00:00:01", "Start of GUID range")
	flag.StringVar(&guidRangeEnd, "guid-range-end", "02:00:00:00:00:00:FF:FF", "End of GUID range")
	flag.StringVar(&namespace, "namespace", "default", "Namespace to watch")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

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

	allocator := ibkubernetes.NewGUIDAllocator(guidRangeStart, guidRangeEnd)

	if err = (&PodReconciler{
		Client:    mgr.GetClient(),
		Allocator: allocator,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pod")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
