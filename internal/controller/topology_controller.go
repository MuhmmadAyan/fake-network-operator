package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fnov1alpha1 "github.com/fake-network-operator/fake-network-operator/api/v1alpha1"
	"github.com/fake-network-operator/fake-network-operator/internal/topology"
)

// TopologyReconciler watches Nodes and materializes per-node topology ConfigMaps
// based on the FakeNicClusterPolicy and the node's pool label.
type TopologyReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Namespace string
}

// Reconcile handles node create/update/delete events.
// When a node has a pool label, it looks up the pool spec and writes a topology ConfigMap.
// When the label is removed or node deleted, it cleans up the ConfigMap.
func (r *TopologyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	node := &corev1.Node{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		if apierrors.IsNotFound(err) {
			// Node deleted — clean up its topology ConfigMap
			if err := topology.DeleteConfigMap(ctx, r.Client, r.Namespace, req.Name); err != nil {
				logger.Error(err, "failed to delete topology configmap for deleted node")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the FakeNicClusterPolicy
	policies := &fnov1alpha1.FakeNicClusterPolicyList{}
	if err := r.List(ctx, policies); err != nil {
		logger.Error(err, "failed to list policies")
		return ctrl.Result{}, err
	}

	if len(policies.Items) == 0 {
		return ctrl.Result{}, nil
	}

	policy := policies.Items[0]
	poolSelector := policy.Spec.NodePoolSelector
	if poolSelector == "" {
		poolSelector = "fake-network-operator.io/nic-node-pool"
	}

	poolName, hasLabel := node.Labels[poolSelector]
	if !hasLabel {
		// Label removed — clean up
		if err := topology.DeleteConfigMap(ctx, r.Client, r.Namespace, node.Name); err != nil {
			logger.Error(err, "failed to delete topology configmap")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	poolSpec, poolExists := policy.Spec.NicNodePools[poolName]
	if !poolExists {
		logger.Info("pool not found in policy", "pool", poolName)
		return ctrl.Result{}, nil
	}

	// Build and write the topology
	topo := topology.BuildNodeTopology(node.Name, poolName, poolSpec)
	if err := topology.WriteConfigMap(ctx, r.Client, r.Namespace, topo); err != nil {
		logger.Error(err, "failed to write topology configmap")
		return ctrl.Result{}, err
	}

	logger.Info("wrote topology configmap", "node", node.Name, "pool", poolName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TopologyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
