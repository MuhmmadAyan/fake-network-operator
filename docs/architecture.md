# Fake Network Operator (FNO) Architecture

This document describes the architecture of the Fake Network Operator (FNO), a Kubernetes operator designed to simulate NVIDIA/Mellanox high-performance networking environments (InfiniBand, RDMA, SR-IOV, ConnectX NICs) without requiring physical hardware.

## 1. System Overview

FNO simulates the entire NVIDIA Network Operator (MNO) stack on commodity hardware. This allows developers to test, validate, and develop Kubernetes applications requiring high-performance networking on any standard Kubernetes cluster, including a local laptop running KinD.

The system revolves around a central Custom Resource Definition (CRD) called `FakeNicClusterPolicy`. This is a cluster-scoped singleton (typically named `nic-cluster-policy`) that dictates the simulated networking configuration across the cluster. The FNO controller-manager reconciles this CRD and orchestrates the deployment and configuration of all node-level DaemonSets required to emulate the networking stack.

## 2. Architecture Diagram

The flow of configuration from the central policy to node-level components is depicted below:

```text
FakeNicClusterPolicy CRD → controller-manager → topology-server → per-node ConfigMaps
                                                                    ↓
                                              nic-labeler, rdma-device-plugin, sriov-device-plugin,
                                              sriov-config-daemon, nic-status-exporter, cli-injector
```

## 3. Core Design Principle: Single Source of Truth

A fundamental design principle of FNO is the use of the `topology-server` as the definitive single source of truth for the simulated hardware layout.

*   **Hardware Synthesis:** The `topology-server` automatically synthesizes realistic ConnectX-6 dual-port NICs when no explicit NICs are specified by the user. By default, it generates:
    *   `mlx5_0`: InfiniBand, RDMA enabled, 100 Gb/sec, attached to NUMA 0.
    *   `mlx5_1`: Ethernet, SR-IOV enabled (8 Virtual Functions), 100 Gb/sec, attached to NUMA 1.
*   **Deterministic Generation:** It generates deterministic hardware identifiers including board IDs, firmware versions, and SHA-256 hashed GUIDs to ensure consistency across cluster restarts and node re-creations.
*   **Configuration Fan-out:** The topology is published into per-node ConfigMaps (e.g., `nic-topology-<node>`).
*   **Downstream Consumption:** ALL downstream FNO components (DaemonSets) read their specific simulated hardware state exclusively from these node-specific ConfigMaps, ensuring a consistent view of the simulated hardware across the cluster.

## 4. Component Deep Dive

The following table details each FNO component and the real-world NVIDIA Network Operator (MNO) component it replaces:

| FNO Component | Replaces Real MNO Component | Description |
| :--- | :--- | :--- |
| **controller-manager** | Network Operator Controller | Reconciles the `FakeNicClusterPolicy` CRD and orchestrates the deployment of all other components. |
| **topology-server** | Node Feature Discovery (NFD) + MOFED Hardware Discovery | Acts as the hardware source of truth, generating and publishing simulated NIC topologies. |
| **nic-labeler** | NFD (NVIDIA NIC Feature Rules) | Patches Kubernetes nodes with appropriate labels (e.g., `nvidia.com/nic.*`, `feature.node.kubernetes.io/pci-15b3.present`) based on the simulated topology. |
| **rdma-device-plugin** | K8s RDMA Shared Device Plugin | Implements the Kubelet gRPC DevicePlugin API, advertising simulated RDMA resources (e.g., `rdma/rdma_shared_device_a`) for pod scheduling. |
| **sriov-device-plugin** | K8s SR-IOV Network Device Plugin | Implements the Kubelet gRPC DevicePlugin API, advertising simulated SR-IOV Virtual Functions (e.g., `nvidia.com/hostdev`). |
| **sriov-config-daemon** | Mellanox SR-IOV Network Operator CNI Daemon | A Go-based state machine that emulates the VF lifecycle (Idle → InProgress → Succeeded), replacing actual `sysfs` VF creation. |
| **nic-status-exporter** | DOCA Telemetry Service | Publishes simulated hardware Prometheus metrics on `:9394/metrics`. |
| **fake-cli + cli-injector** | MOFED CLI Tools | Injects shell shims (`ibstat`, `ibv_devinfo`, `rdma`) onto the host OS to trick user workloads into believing actual drivers are installed. |
| **ib-kubernetes** | InfiniBand Subnet Manager (SM) Integration | Simulates deterministic SHA-256 GUID and PKey allocation for InfiniBand networks. |
| **ufm-stub** | NVIDIA UFM | Emulates the NVIDIA Unified Fabric Manager (UFM) REST API for management plane testing. |
| **kwok-nic-device-plugin** | N/A (KWOK Integration) | Publishes resources for KWOK (Kubernetes WithOut Kubelet) simulated nodes to test scale-out scenarios. |
| **admission-webhook** | Network Operator Validating Webhook | Validates the configuration of `FakeNicClusterPolicy` CRDs. |
| **nic-dra-plugin** | K8s DRA Plugin | Forward-looking stub implementing Kubernetes Dynamic Resource Allocation (DRA) for future resource management. |

## 5. Node Onboarding Lifecycle

When a new node is targeted for simulated networking, the onboarding process follows this deterministic sequence:

1.  **Labeling:** The user or provisioner labels the node with `fake-network-operator.io/nic-node-pool=default`.
2.  **Topology Generation:** The `topology-server` detects the node and generates the corresponding `nic-topology-<node>` ConfigMap.
3.  **Node Labeling:** The `nic-labeler` reads the ConfigMap and patches the Kubernetes Node object with standard `nvidia.com/nic.*` hardware labels.
4.  **RDMA Registration:** The `rdma-device-plugin` registers with the Kubelet gRPC socket, advertising available RDMA devices (e.g., `rdma/rdma_shared_device_a: 64`).
5.  **SR-IOV Registration:** The `sriov-device-plugin` registers with the Kubelet, advertising available SR-IOV VFs (e.g., `nvidia.com/hostdev: 8`).
6.  **VF Lifecycle:** The `sriov-config-daemon` executes the virtual function lifecycle state machine to simulate device readiness.
7.  **Telemetry:** The `nic-status-exporter` begins publishing simulated Prometheus metrics.
8.  **CLI Injection:** The `cli-injector` copies the `fake-cli` binary to the host path (e.g., `/opt/fake-network-operator/bin/`).
9.  **Workload Scheduling:** Pods are scheduled onto the node, successfully requesting and receiving simulated RDMA devices and secondary Multus interfaces.

## 6. Secondary Network Integration

FNO natively supports Kubernetes secondary networking via Multus CNI:

*   **CRD Generation:** FNO generates `NetworkAttachmentDefinition` (NAD) Custom Resources based on its configuration via Helm values.
*   **Multus Attachment:** The Multus CNI reads these NADs and attaches the secondary interfaces (typically `net1`, `net2`, etc.) to scheduled pods.
*   **Supported CNI Types:** FNO is compatible with standard CNI types commonly used in high-performance networking, including `macvlan`, `host-device`, and `bridge`.
