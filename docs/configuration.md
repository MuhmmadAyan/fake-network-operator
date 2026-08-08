# Configuration Reference

All FNO settings are managed via Helm chart values.

## Helm Values Reference

### Image Configuration

| Parameter | Default | Description |
|---|---|---|
| `image.repository` | `ghcr.io/fake-network-operator/fake-network-operator` | Container image repository |
| `image.tag` | `0.1.0` | Image tag |
| `image.pullPolicy` | `IfNotPresent` | Image pull policy |

### Topology Configuration

| Parameter | Default | Description |
|---|---|---|
| `topology.nicNodePools.default.gpuDirectCapable` | `true` | Whether NICs support GPUDirect RDMA |
| `topology.nicNodePools.default.rdma.resourceName` | `rdma_shared_device_a` | RDMA device plugin resource name |
| `topology.nicNodePools.default.rdma.sharingCount` | `64` | Number of shared RDMA device slots |
| `topology.nicNodePools.default.sriov.totalVfs` | `8` | Number of SR-IOV Virtual Functions |
| `topology.nicNodePools.default.sriov.resourceName` | `hostdev` | SR-IOV resource name |
| `topology.nicNodePools.default.sriov.resourcePrefix` | `nvidia.com` | SR-IOV resource domain prefix |

### Component Toggles

| Component | Default State |
|---|---|
| `nicLabeler` | enabled |
| `rdmaDevicePlugin` | enabled |
| `sriovDevicePlugin` | enabled |
| `sriovConfigDaemon` | enabled |
| `statusExporter` | enabled |
| `cliInjector` | enabled |
| `ibKubernetes` | enabled |
| `ufmStub` | enabled |
| `admissionWebhook` | enabled |
| `kwokPlugin` | disabled |
| `draNicPlugin` | disabled |

### Secondary Networks

To add NetworkAttachmentDefinitions, configure the `secondaryNetworks` array:

```yaml
secondaryNetworks:
  - name: rdma-net
    type: macvlan
  - name: hostdevice-net
    type: host-device
```
