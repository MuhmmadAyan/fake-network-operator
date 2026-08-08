# Quick Start Guide

Get Fake Network Operator (FNO) running on a local KinD cluster in under 10 minutes.

## Prerequisites

- **Docker** 20.10+
- **KinD** (Kubernetes in Docker)
- **Helm** 3 (v3.8.0+)
- **kubectl** (v1.26+)

> [!NOTE]
> No Go installation is needed! All Go tools run securely inside Docker containers.

## Step 1: Create a KinD Cluster

Create a local Kubernetes cluster with one control-plane node and one worker node using KinD.

```bash
cat <<EOF | kind create cluster --name fno-demo --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
  - role: worker
EOF
```

## Step 2: Install CRDs

Install the Custom Resource Definitions (CRDs) required by the operator.

```bash
kubectl apply -f deploy/fake-network-operator/crds/
```

## Step 3: Install FNO via Helm

Add the official Helm repository and install Fake Network Operator:

```bash
# Add Helm repository
helm repo add fake-network-operator https://muhmmadayan.github.io/fake-network-operator
helm repo update

# Install FNO operator
helm upgrade --install fake-network-operator fake-network-operator/fake-network-operator \
  --namespace fake-network-operator \
  --create-namespace
```

## Building from Source (Optional)

If you need to build the image locally instead of pulling the pre-built GHCR image:

```bash
make docker-build
kind load docker-image ghcr.io/muhmmadayan/fake-network-operator:0.1.0 --name fno-demo
```

## Step 4: Apply the Sample Policy

Configure the operator by applying the sample network policy.

```bash
kubectl apply -f examples/sample-policy.yaml
```

## Step 5: Label a Worker Node

Label the worker node to trigger the operator's scheduling and resource allocation mechanisms.

```bash
kubectl label node fno-demo-worker fake-network-operator.io/nic-node-pool=default
```

## Step 6: Verify Installation

Ensure that the operator and its components are running correctly.

### Check all pods are running

```bash
kubectl get pods -n fake-network-operator
```

**Expected Output:**
You should see 8 pods, and all should be in the `1/1 Running` state.

### Check node resources

```bash
kubectl describe node fno-demo-worker | grep -A10 'Allocatable'
```

**Expected Output:**
```text
  rdma/rdma_shared_device_a: 64
  nvidia.com/hostdev: 8
```

### Check node labels

```bash
kubectl get node fno-demo-worker --show-labels | tr ',' '\n' | grep -E 'nvidia|feature'
```

### Check Prometheus metrics

Replace `<status-exporter-pod>` with the actual name of your status exporter pod.

```bash
kubectl exec -n fake-network-operator <status-exporter-pod> -- curl -s localhost:9394/metrics | head -20
```

## Step 7: Deploy a Test Workload

Create a test pod that requests RDMA resources to verify the allocation works end-to-end.

Save the following YAML to `test-pod.yaml` and apply it with `kubectl apply -f test-pod.yaml`:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: rdma-test-pod
  annotations:
    k8s.v1.cni.cncf.io/networks: fake-network-operator/rdma-net
spec:
  containers:
  - name: test
    image: busybox
    command: ['sleep', '3600']
    resources:
      limits:
        rdma/rdma_shared_device_a: "1"
```

## Cleanup

When you are done testing, you can clean up the resources and delete the cluster.

```bash
helm uninstall fake-network-operator -n fake-network-operator
kind delete cluster --name fno-demo
```
