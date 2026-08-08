package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	"github.com/fake-network-operator/fake-network-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

type FakeNicClusterPolicyValidator struct {
	decoder admission.Decoder
}

func (v *FakeNicClusterPolicyValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	policy := &v1alpha1.FakeNicClusterPolicy{}

	err := v.decoder.Decode(req, policy)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if policy.Name != "nic-cluster-policy" {
		return admission.Denied("FakeNicClusterPolicy name must be 'nic-cluster-policy'")
	}

	nicNames := make(map[string]bool)
	for _, np := range policy.Spec.NicNodePools {
		for _, nic := range np.NICs {
			if nicNames[nic.Name] {
				return admission.Denied("NIC names must be unique within a node pool")
			}
			nicNames[nic.Name] = true

			if nic.SRIOV != nil && nic.SRIOV.ResourceName == "" {
				return admission.Denied("SRIOV resource name cannot be empty")
			}
			if nic.RDMA != nil && nic.RDMA.ResourceName == "" {
				return admission.Denied("RDMA resource name cannot be empty")
			}
		}
	}

	return admission.Allowed("")
}

func main() {
	var port int
	var certDir string

	flag.IntVar(&port, "port", 9443, "Webhook server port.")
	flag.StringVar(&certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "Webhook cert dir.")

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
		WebhookServer: webhook.NewServer(webhook.Options{
			Port:    port,
			CertDir: certDir,
		}),
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	hookServer := mgr.GetWebhookServer()
	decoder := admission.NewDecoder(scheme)

	hookServer.Register("/validate-fake-network-operator-v1alpha1-fakenicclusterpolicy", &webhook.Admission{Handler: &FakeNicClusterPolicyValidator{decoder: decoder}})

	setupLog.Info("starting webhook manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
