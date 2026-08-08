# Design Decisions

## 1. Why Build FNO?
- Testing NVIDIA Network Operator features requires physical Mellanox ConnectX NICs, InfiniBand switches, and RDMA infrastructure ($50K-$200K+ per node).
- CI/CD pipelines cannot validate network-aware scheduling without this hardware.
- Inspired by run-ai/fake-gpu-operator (FGO) which solved the same problem for GPUs.

## 2. Core Design Principles
Inherited from FGO, adapted for networking:
1. **Fake API contracts, not hardware** — implement real Kubelet gRPC, real CRDs, real Prometheus, real CLI output.
2. **Single declarative source of truth** — `topology-server` generates all hardware state from one CRD.
3. **Label-triggered activation** — DaemonSets activate via node label selectors.
4. **Process-level interception** — CLI shims (`ibstat`, `ibv_devinfo`) produce identical text output.
5. **Support modern K8s APIs** — DRA plugin stub for forward compatibility.
6. **Additive & coexistence-friendly** — can run alongside real MNO via label isolation.

## 3. Policy vs Hardware Separation
- `FakeNicClusterPolicy` defines INTENT only (enable RDMA, enable SR-IOV, VF count, component toggles).
- `topology-server` is the HARDWARE source of truth (synthesizes ConnectX-6 specs, board IDs, GUIDs).
- This mirrors real MNO where `NicClusterPolicy` defines policy and NFD+MOFED discover actual hardware.

## 4. Automatic Hardware Synthesis
- When no explicit NICs are specified in the policy, `topology-server` auto-generates a standard dual-port ConnectX-6:
  - **Port 1 (`mlx5_0`)**: InfiniBand, RDMA, 100 Gb/sec.
  - **Port 2 (`mlx5_1`)**: Ethernet, SR-IOV with configurable VFs, 100 Gb/sec.
- Deterministic identity generation using SHA-256 hashing for GUIDs, board IDs, VF IDs.
- Ensures stable identities across pod restarts.

## 5. Two-Tier Fidelity Model
- **`fake` backend (default)**: Lightweight Go binaries, synthetic gRPC responses, no host drivers.
- **`mock` backend (future)**: Real upstream components against mocked sysfs/netlink layer.

## 6. Multi-Binary Docker Image
- All 14 components compiled into a single Docker image.
- Each binary selected via container command override in Helm templates.
- Minimizes registry storage and pull time.

## 7. Why ConfigMaps as Topology Storage?
- No external database dependency.
- Native Kubernetes RBAC and audit logging.
- Watch-based change propagation to DaemonSets.
- Human-readable and kubectl-debuggable.
