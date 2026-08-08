package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	fnov1alpha1 "github.com/fake-network-operator/fake-network-operator/api/v1alpha1"
)

// FakeNicClusterPolicyReconciler reconciles the singleton FakeNicClusterPolicy CR.
// It watches Nodes to count pool membership and updates the CR's status accordingly.
type FakeNicClusterPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles create/update/delete of the FakeNicClusterPolicy CR and
// node label changes that affect pool membership counts.
func (r *FakeNicClusterPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	policy := &fnov1alpha1.FakeNicClusterPolicy{}
	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// List all nodes to count pool membership
	nodes := &corev1.NodeList{}
	if err := r.List(ctx, nodes); err != nil {
		logger.Error(err, "failed to list nodes")
		return ctrl.Result{}, err
	}

	poolSelector := policy.Spec.NodePoolSelector
	if poolSelector == "" {
		poolSelector = "fake-network-operator.io/nic-node-pool"
	}

	// Count nodes per pool
	poolCounts := make(map[string]int)
	for _, node := range nodes.Items {
		if poolName, exists := node.Labels[poolSelector]; exists {
			poolCounts[poolName]++
		}
	}

	// Update status
	if policy.Status.AppliedNodePools == nil {
		policy.Status.AppliedNodePools = make(map[string]fnov1alpha1.NodePoolStatus)
	}

	for poolName := range policy.Spec.NicNodePools {
		count := poolCounts[poolName]
		policy.Status.AppliedNodePools[poolName] = fnov1alpha1.NodePoolStatus{
			TotalNodes: count,
			ReadyNodes: count,
		}
	}

	policy.Status.State = "ready"
	if err := r.Status().Update(ctx, policy); err != nil {
		logger.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// Watches FakeNicClusterPolicy directly, and also watches Nodes so that
// pool membership changes trigger a re-reconcile of the singleton CR.
func (r *FakeNicClusterPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fnov1alpha1.FakeNicClusterPolicy{}).
		Watches(
			&corev1.Node{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
				return []reconcile.Request{
					{NamespacedName: client.ObjectKey{Name: "nic-cluster-policy"}},
				}
			}),
		).
		Complete(r)
}
