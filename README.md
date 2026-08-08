<p align="center">
  <h1 align="center">🌐 Fake Network Operator (FNO)</h1>
  <p align="center">
    <strong>Simulate NVIDIA/Mellanox high-performance networking on any Kubernetes cluster — no hardware required.</strong>
  </p>
  <p align="center">
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
    <img src="https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white" alt="Go Version">
    <img src="https://img.shields.io/badge/Kubernetes-1.26+-326CE5?logo=kubernetes&logoColor=white" alt="Kubernetes">
    <img src="https://img.shields.io/badge/Helm-3-0F1689?logo=helm&logoColor=white" alt="Helm 3">
    <img src="https://img.shields.io/badge/Version-0.1.0-green" alt="Version">
  </p>
</p>

---

## Why FNO?

Testing Kubernetes network scheduling for AI/ML workloads (NCCL, PyTorch Distributed, MPI) typically requires **physical Mellanox ConnectX NICs, InfiniBand switches, and RDMA-capable infrastructure** — hardware costing $50K–$200K+ per node.

**Fake Network Operator** eliminates this barrier by simulating the entire [NVIDIA Network Operator (MNO)](https://github.com/Mellanox/network-operator) stack on commodity hardware — even a laptop running KinD.

### Who is this for?

| Persona | Use Case |
|---------|----------|
| **MLOps / Platform Engineers** | Test network-aware scheduling, RDMA resource claims, and SR-IOV VF allocation in CI/CD without GPU nodes |
| **Kubernetes Contributors** | Develop and test network plugins, device plugins, and DRA drivers against realistic hardware topology |
| **AI/ML Researchers** | Prototype distributed training configurations with simulated multi-NIC, multi-rail topologies |
| **DevOps Teams** | Validate Helm charts, Multus CNI configurations, and NetworkAttachmentDefinitions locally |

---

## Features

| Feature | FNO (Simulated) | Real NVIDIA MNO |
|---------|:---:|:---:|
| **CRD-driven policy** (`FakeNicClusterPolicy` / `NicClusterPolicy`) | ✅ | ✅ |
| **Auto-synthesis of ConnectX-6 dual-port NIC topology** | ✅ | ✅ (hardware discovery) |
| **RDMA Device Plugin** (`rdma/rdma_shared_device_a`) | ✅ | ✅ |
| **SR-IOV Device Plugin** (`nvidia.com/hostdev`) | ✅ | ✅ |
| **Node labeling** (`nvidia.com/nic.*`, `feature.node.kubernetes.io/pci-15b3.present`) | ✅ | ✅ (via NFD + MOFED) |
| **Prometheus metrics** (port counters, RDMA utilization, VF allocation) | ✅ (`:9394/metrics`) | ✅ (DOCA Telemetry) |
| **Secondary networks** (Multus NADs via Helm) | ✅ | ✅ |
| **SR-IOV VF lifecycle simulation** | ✅ | ✅ |
| **InfiniBand GUID/PKey allocation** | ✅ | ✅ |
| **CLI tools** (`ibstat`, `ibv_devinfo`, `rdma`) | ✅ (injected fakes) | ✅ (real MOFED) |
| **Requires physical NICs** | ❌ | ✅ |

---

## Architecture

```
                         ┌──────────────────────────┐
                         │   FakeNicClusterPolicy    │  ← CRD (policy intent)
                         └────────────┬─────────────┘
                                      │
                                      ▼
                         ┌──────────────────────────┐
                         │    controller-manager     │  ← Reconciles policy → node components
                         └────────────┬─────────────┘
                                      │
            ┌─────────────────────────┼─────────────────────────┐
            │                         │                         │
            ▼                         ▼                         ▼
  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
  │  topology-server  │    │  fake-nic-labeler │    │ admission-webhook│
  │ (hardware synth)  │    │  (node labels)    │    │  (CRD validation)│
  └────────┬─────────┘    └──────────────────┘    └──────────────────┘
           │
           ▼  Per-node ConfigMaps (source of truth)
           │
  ┌────────┴─────────────────────────────────────────────────┐
  │                         │                         │       │
  ▼                         ▼                         ▼       ▼
┌────────────────┐  ┌────────────────┐  ┌──────────────┐  ┌────────────────┐
│ rdma-device-   │  │ sriov-device-  │  │ sriov-config │  │ nic-status-    │
│ plugin         │  │ plugin         │  │ daemon       │  │ exporter       │
│ (rdma/shared)  │  │ (nvidia.com/)  │  │ (VF lifecycle│  │ (Prometheus)   │
└────────────────┘  └────────────────┘  └──────────────┘  └────────────────┘
```

**Key design principle:** The `FakeNicClusterPolicy` CRD defines _policy intent_ only (RDMA enabled, SR-IOV VF count, component toggles). The `topology-server` is the single **hardware source of truth** — it auto-synthesizes realistic ConnectX-6 dual-port NICs with deterministic board IDs, firmware versions, and GUIDs.

---

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [KinD](https://kind.sigs.k8s.io/) (Kubernetes in Docker)
- [Helm 3](https://helm.sh/docs/intro/install/) (v3.8.0+)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) (v1.26+)

### 1. Create a KinD cluster

```bash
cat <<EOF | kind create cluster --name fno-demo --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
  - role: worker
EOF
```

### 2. Build & load the image

```bash
make docker-build
kind load docker-image ghcr.io/fake-network-operator/fake-network-operator:0.1.0 --name fno-demo
```

### 3. Install FNO

```bash
# Install CRDs
kubectl apply -f deploy/fake-network-operator/crds/

# Install Helm chart
make helm-install
```

### 4. Label a worker node

```bash
kubectl label node fno-demo-worker fake-network-operator.io/nic-node-pool=default
```

### 5. Verify

```bash
# All 8 pods should be Running
kubectl get pods -n fake-network-operator

# Node should advertise simulated NIC resources
kubectl describe node fno-demo-worker | grep -A5 "Allocatable"
# Expected:
#   rdma/rdma_shared_device_a: 64
#   nvidia.com/hostdev:        8
```

> For more details, see the [Configuration](#configuration) section below.

---

## Configuration

All settings are managed via the Helm chart values file:

📁 [`deploy/fake-network-operator/values.yaml`](deploy/fake-network-operator/values.yaml)

| Category | What You Can Configure |
|----------|----------------------|
| **Topology** | NIC models (ConnectX-5/6/7), link types (Ethernet/InfiniBand), speeds, NUMA affinity |
| **RDMA** | Shared device plugin resource names, sharing counts |
| **SR-IOV** | VF counts, resource names, resource prefixes |
| **Secondary Networks** | NetworkAttachmentDefinition generation (macvlan, host-device, bridge) |
| **Components** | Enable/disable individual components independently |

---

## Components

FNO includes 14 modular binary components compiled into a single multi-binary Docker image:

| Component | Category | Description |
|-----------|----------|-------------|
| `controller-manager` | Control Plane | Reconciles `FakeNicClusterPolicy` CRDs and manages daemon lifecycles |
| `topology-server` | Control Plane | Auto-synthesizes per-node NIC topology into ConfigMaps |
| `fake-nic-labeler` | DaemonSet | Labels nodes with NFD and Mellanox PCI feature labels |
| `fake-rdma-device-plugin` | DaemonSet | Kubelet Device Plugin advertising simulated RDMA shared devices |
| `fake-sriov-device-plugin` | DaemonSet | Kubelet Device Plugin advertising simulated SR-IOV VFs |
| `fake-sriov-config-daemon` | DaemonSet | Simulates SR-IOV VF creation lifecycle and state machine |
| `fake-nic-status-exporter` | DaemonSet | Exposes Prometheus metrics for simulated NIC telemetry |
| `fake-cli` | Utility | Multi-call binary emulating `ibstat`, `ibv_devinfo`, `rdma` commands |
| `fake-cli-injector` | DaemonSet | Injects fake CLI tools onto host for pod access |
| `fake-ib-kubernetes` | Controller | Simulates InfiniBand GUID/PKey allocation for pods |
| `fake-ufm-stub` | Service | Emulates NVIDIA UFM REST API endpoints |
| `kwok-nic-device-plugin` | Controller | Publishes fake NIC resources for KWOK simulated nodes |
| `admission-webhook` | Webhook | Validates `FakeNicClusterPolicy` CRDs |
| `fake-nic-dra-plugin` | DaemonSet | Dynamic Resource Allocation plugin (K8s v1.31+ forward-looking) |

---

## Development

### Make Targets

| Target | Description |
|--------|-------------|
| `make docker-build` | Build the multi-stage Docker image containing all 14 binaries |
| `make docker-push` | Push the Docker image to the container registry |
| `make helm-lint` | Lint the Helm chart |
| `make helm-template` | Render Helm templates to stdout (dry-run) |
| `make helm-install` | Deploy FNO to Kubernetes via Helm |
| `make helm-uninstall` | Uninstall the Helm release |
| `make test` | Run unit tests |
| `make fmt` | Format Go source code |
| `make vet` | Run Go static analysis |
| `make clean` | Remove locally built Docker image |

### Project Structure

```
├── api/v1alpha1/        # CRD types with kubebuilder markers
├── cmd/                 # 14 binary entry points
├── internal/            # Core packages (topology, deviceplugin, metrics, controller)
├── deploy/              # Helm chart, CRDs, templates
├── docs/                # Architecture, quickstart, configuration, design docs
├── examples/            # Sample manifests and workload examples
└── hack/                # Code generation scripts
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/architecture.md) | System overview, component deep dive, node onboarding lifecycle |
| [Quick Start Guide](docs/quickstart.md) | Step-by-step setup on KinD |
| [Configuration Reference](docs/configuration.md) | Helm values reference (topology, components, networks) |
| [FNO vs Real MNO](docs/fno-vs-mno.md) | Feature comparison — what's real vs simulated |
| [Design Decisions](docs/design/design-decisions.md) | Architecture rationale and core principles |
| [Contributing Guide](CONTRIBUTING.md) | How to contribute |
| [Changelog](CHANGELOG.md) | Release history |
| [Security Policy](SECURITY.md) | Vulnerability reporting |

---

## Roadmap

- [ ] End-to-end integration tests (KinD-based)
- [ ] GitHub Actions CI/CD pipeline
- [ ] Helm chart publishing to ArtifactHub
- [ ] Grafana dashboard templates for `nic-status-exporter` metrics
- [ ] Fault injection API (simulate NIC failures, link flaps, PFC storms)
- [ ] Multi-rail topology support (4x/8x ConnectX-7)
- [ ] KWOK integration for 1000+ simulated node testing

---

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

Copyright 2024-2026 Mohammad Ayan.

Licensed under the [Apache License, Version 2.0](LICENSE).
