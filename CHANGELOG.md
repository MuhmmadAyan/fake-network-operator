# Changelog
All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.1.0] - 2026-08-09
### Added
- Initial release of Fake Network Operator (FNO)
- FakeNicClusterPolicy CRD with controller-manager reconciliation
- Topology server with automatic ConnectX-6 dual-port NIC synthesis
- Fake RDMA device plugin (advertises rdma/rdma_shared_device_a)
- Fake SR-IOV device plugin (advertises nvidia.com/hostdev)
- NIC labeler (applies nvidia.com/nic.* and feature.node.kubernetes.io labels)
- NIC status exporter with Prometheus metrics on :9394/metrics
- SR-IOV config daemon for VF lifecycle simulation
- Helm chart with CRDs, DaemonSets, and NetworkAttachmentDefinitions
- Secondary network management via Helm values (macvlan, host-device NADs)
- Multi-binary distroless Docker image
