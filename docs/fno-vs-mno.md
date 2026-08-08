# FNO vs NVIDIA Network Operator (MNO)

The Fake Network Operator (FNO) is designed to simulate the cluster-visible surface of the real NVIDIA Network Operator (MNO) without requiring any physical Mellanox/NVIDIA networking hardware. It emulates the exact same Kubernetes API contracts and node-level interfaces, allowing workloads to be scheduled and interact with the system exactly as they would in a physical InfiniBand/RDMA environment.

## 1. What is Real vs What is Simulated

The following table breaks down the various layers of the network operator stack and identifies whether they are real implementations or simulated by FNO.

| Layer | Real MNO | FNO Equivalent | Real or Simulated? | Notes |
|-------|----------|----------------|--------------------|-------|
| K8s API & CRDs | Real | Real | **REAL** | FNO uses the exact same CRD definitions and controller reconciliation loops as MNO. |
| Kubelet Device Plugin Protocol | Real | Real | **REAL** | FNO implements a real gRPC `v1beta1.DevicePluginServer` to register devices with the Kubelet. |
| Secondary Pod Networking (CNI) | Real | Real | **REAL** | Uses real Multus CNI along with macvlan/bridge to attach actual `net1` interfaces to pods. |
| Prometheus Telemetry | Real | Real | **REAL** | Exposes a real HTTP Prometheus metrics exporter on port `:9394`. |
| Host Kernel Drivers (MOFED) | Real kernel modules | ConfigMap topology | **SIMULATED** | Host drivers are replaced with static ConfigMap topology payloads in FNO. |
| Hardware Virtualization (SR-IOV) | Real `sysfs` VF creation | Go state machine | **SIMULATED** | FNO uses a Go state machine to mimic VF state without actual PCIe SR-IOV configuration. |
| InfiniBand Switch Fabric | Real Subnet Manager (SM) | SHA-256 GUIDs | **SIMULATED** | FNO uses deterministic SHA-256 hashing to generate simulated GUIDs instead of querying a real fabric. |
| Host CLI Tools | Real binaries (e.g., `ibstat`) | `fake-cli` shim | **SIMULATED** | FNO injects shim binaries that produce text output identical to the real Mellanox tools. |
| GPUDirect RDMA (`nvidia-peermem`) | Real | Not needed | **NOT NEEDED** | There is no physical GPU-to-NIC PCIe path to optimize in FNO. |

## 2. Component Mapping Table

This table maps the core components of the NVIDIA Network Operator to their simulated counterparts in the Fake Network Operator.

| Real MNO Component | FNO Equivalent | Emulation Strategy |
|--------------------|----------------|--------------------|
| MOFED Driver Container | `nic-labeler` + `cli-injector` | Patches node labels and injects mock CLI binaries to mimic driver presence. |
| RDMA Shared Device Plugin | `rdma-device-plugin` | Implements real Kubelet gRPC protocol but advertises synthetic RDMA device IDs. |
| SR-IOV Device Plugin | `sriov-device-plugin` | Implements real Kubelet gRPC protocol but advertises synthetic Virtual Function (VF) IDs. |
| SR-IOV Config Daemon | `sriov-config-daemon` | Replaces actual `sysfs` writes with a purely in-memory Go state machine. |
| NFD (Node Feature Discovery) | `nic-labeler` | Applies the exact same node label keys that real NFD would generate for Mellanox hardware. |
| IB-Kubernetes | `ib-kubernetes` | Uses deterministic GUID generation instead of reading from physical physical IB interfaces. |
| DOCA Telemetry Service | `nic-status-exporter` | Generates random-walk Prometheus metrics to simulate telemetry data. |
| Kube-Macpool | Not needed | FNO relies on standard IPAM plugins. |
| NV-IPAM | Not needed | FNO utilizes standard `host-local` or `whereabouts` IPAM for secondary networking. |
| UFM (Unified Fabric Manager) | `ufm-stub` | Provides REST API emulation matching the UFM endpoints. |

## 3. Why This Approach Works

Containerized applications interact with the networking hardware stack through a very specific and limited set of interfaces. From the perspective of a workload pod, it only "sees" the network stack via:

1. **K8s Extended Resources:** For pod scheduling (e.g., requesting `nvidia.com/rdma` or `mellanox.com/sriov`).
2. **Node Labels:** For node selector affinity and anti-affinity rules.
3. **CLI Tool Output:** Standard utilities like `ibstat` or `ibv_devinfo` executed within the container.
4. **Device Files:** Character devices exposed at `/dev/infiniband/*` for RDMA verbs programming.
5. **Prometheus Metrics:** For cluster-level monitoring and alerting.

FNO successfully fakes **all** of these API contracts and presentation layers identically. As a result, workloads cannot distinguish between a node managed by FNO and a node managed by the real MNO at the Kubernetes API and container runtime level.

## 4. Limitations

While FNO is highly effective for control-plane simulation, scheduling tests, and API validation, it operates entirely without physical hardware. Consequently, it has the following fundamental limitations:

*   **No actual RDMA data plane:** Network packets do not actually traverse an InfiniBand fabric or leverage Remote Direct Memory Access.
*   **No real SR-IOV hardware virtualization:** There is no actual hardware-level isolation or Virtual Function (VF) creation.
*   **No GPUDirect RDMA:** The optimized PCIe peer-to-peer path between GPUs and Network Interface Cards does not exist.
*   **Performance Benchmarking:** Any performance, throughput, or latency testing against FNO is meaningless and does not reflect real-world hardware capabilities.
*   **Not a Security Boundary:** FNO simulation should **never** be relied upon as a security mechanism or isolation boundary.
