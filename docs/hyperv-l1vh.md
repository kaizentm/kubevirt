# HyperV-Layered Hypervisor (L1VH)

## Overview

HyperV-Layered is an alternative hypervisor implementation for KubeVirt that leverages Microsoft's Layer 1 Virtualization Host (L1VH) technology to eliminate nested virtualization performance penalties on Azure infrastructure. Instead of using traditional nested KVM/QEMU virtualization, HyperV-Layered provides direct hardware access to virtual machines through Microsoft's `/dev/mshv` driver, delivering near-native performance.

### Key Benefits

- **Near-Native Performance**: Eliminates nested virtualization overhead by providing direct hypervisor access
- **Hardware Passthrough**: Enables direct GPU, storage, and networking device assignment to virtual machines
- **Azure Optimization**: Native integration with Azure's Hyper-V infrastructure
- **Improved Resource Efficiency**: Reduced CPU and memory overhead compared to nested virtualization

## Prerequisites

### Infrastructure Requirements

To use the HyperV-Layered hypervisor, your infrastructure must meet the following requirements:

1. **Azure Environment**: 
   - Running on Azure Virtual Machines that support nested virtualization
   - Azure VM sizes with L1 Virtual Hardware (L1VH) capabilities
   - Examples: [Standard_D*v5, Standard_E*v5 series with nested virtualization support]

2. **Operating System**:
   - Linux kernel with `/dev/mshv` device support
   - Microsoft Hyper-V kernel modules loaded

3. **Kubernetes Cluster**:
   - Kubernetes 1.20+ (check [kubernetes-compatibility.md](kubernetes-compatibility.md) for specific version requirements)
   - KubeVirt installed and operational
   - Worker nodes must have `/dev/mshv` device available

4. **Device Plugin**:
   - The cluster must expose `devices.kubevirt.io/mshv` resources on nodes
   - This is handled automatically by KubeVirt when HyperV-Layered is configured

### Verification

Before enabling HyperV-Layered, verify that your nodes support the required hardware:

```bash
# Check if /dev/mshv device exists on worker nodes
kubectl debug node/<node-name> -it --image=ubuntu -- ls -l /dev/mshv
```

Expected output:
```
crw-rw-rw- 1 root root 10, 232 Oct 21 12:00 /dev/mshv
```

## Configuration

### Enabling HyperV-Layered Hypervisor

HyperV-Layered is configured at the cluster level through the KubeVirt Custom Resource (CR). 

1. Edit the KubeVirt CR:

```bash
kubectl edit kubevirt kubevirt -n kubevirt
```

2. Add the hypervisor configuration to the spec:

```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
  namespace: kubevirt
spec:
  configuration:
    hypervisorConfiguration:
      name: hyperv-layered
```

3. Save and exit. KubeVirt will automatically reconfigure to use the HyperV-Layered hypervisor.

### Complete Configuration Example

Here's a complete KubeVirt CR example with HyperV-Layered enabled:

```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
  namespace: kubevirt
spec:
  configuration:
    hypervisorConfiguration:
      name: hyperv-layered
    developerConfiguration:
      featureGates:
        - HotplugVolumes
        - LiveMigration
```

## How It Works

### Hypervisor Selection

When HyperV-Layered is configured:

1. **Domain Type**: KubeVirt generates libvirt domain XML with type `hyperv` instead of `kvm`
2. **Device Resource**: Virt-launcher pods request `devices.kubevirt.io/mshv` instead of `devices.kubevirt.io/kvm`
3. **Backend Driver**: The Microsoft Hyper-V backend (`/dev/mshv`) is used for virtualization instead of KVM (`/dev/kvm`)

### Resource Management

The virt-launcher pod's resource requirements change when using HyperV-Layered:

**With KVM (default):**
```yaml
resources:
  limits:
    devices.kubevirt.io/kvm: "1"
```

**With HyperV-Layered:**
```yaml
resources:
  limits:
    devices.kubevirt.io/mshv: "1"
```

## Using HyperV-Layered with Virtual Machines

Once HyperV-Layered is configured at the cluster level, it applies automatically to all VirtualMachineInstances. No changes to individual VM definitions are required.

### Example Virtual Machine

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: my-vm
spec:
  running: true
  template:
    metadata:
      labels:
        kubevirt.io/vm: my-vm
    spec:
      domain:
        devices:
          disks:
          - name: containerdisk
            disk:
              bus: virtio
          - name: cloudinitdisk
            disk:
              bus: virtio
        resources:
          requests:
            memory: 1024M
            cpu: 1
      volumes:
      - name: containerdisk
        containerDisk:
          image: quay.io/kubevirt/fedora-cloud-container-disk-demo
      - name: cloudinitdisk
        cloudInitNoCloud:
          userDataBase64: SGkuXG4=
```

This VM will automatically use the HyperV-Layered hypervisor without any additional configuration.

## Verification

### Verifying HyperV-Layered is Active

After enabling HyperV-Layered, verify it's being used:

1. **Check virt-launcher pod resources:**

```bash
# Get a running VMI
kubectl get vmi -A

# Check the virt-launcher pod for the VMI
kubectl get pod virt-launcher-<vmi-name>-xxxxx -n <namespace> -o yaml | grep mshv
```

Expected output should show:
```yaml
devices.kubevirt.io/mshv: "1"
```

2. **Check libvirt domain XML:**

```bash
# Connect to virt-launcher pod
kubectl exec -it virt-launcher-<vmi-name>-xxxxx -n <namespace> -- virsh domxml <domain>
```

The domain XML should show:
```xml
<domain type='hyperv'>
  ...
</domain>
```

3. **Monitor metrics:**

Check the hypervisor type metric:
```bash
kubectl get --raw /api/v1/namespaces/kubevirt/services/kubevirt-prometheus-metrics:metrics/proxy/metrics | grep kubevirt_vmi_hypervisor
```

The metric should show `hypervisor="mshv"` for VMs running with HyperV-Layered.

## Infrastructure Constraints and Assumptions

### Constraints

1. **Azure-Only**: HyperV-Layered is designed for and only works on Azure infrastructure with L1VH support
2. **Architecture**: Currently supports x86_64/AMD64 architecture only
3. **Device Availability**: All worker nodes must have `/dev/mshv` device available
4. **Cluster-Wide Setting**: HyperV-Layered configuration applies to all VMs in the cluster; per-VM hypervisor selection is not supported
5. **Live Migration**: Live migration between KVM and HyperV-Layered hypervisors is not supported

### Assumptions

1. **Homogeneous Cluster**: All worker nodes are assumed to have identical hypervisor capabilities (either all support KVM or all support HyperV-Layered)
2. **Device Plugin**: The Kubernetes device plugin framework is properly configured and operational
3. **Permissions**: The virt-launcher pods have appropriate permissions to access `/dev/mshv`
4. **Libvirt Support**: The libvirt version in use supports the Hyper-V domain type
5. **Memory Overhead**: Memory overhead calculations assume similar characteristics between KVM and HyperV-Layered (this is validated in testing)

### Performance Considerations

1. **Startup Time**: VMs may have different startup characteristics compared to KVM
2. **Memory Management**: While overhead is similar to KVM, actual memory usage patterns may differ
3. **I/O Performance**: Direct hardware access can significantly improve I/O performance for storage and networking
4. **CPU Overhead**: Reduced virtualization overhead compared to nested KVM

## Troubleshooting

### Common Issues

#### 1. VMI Fails to Start - Missing /dev/mshv Device

**Symptom**: VMI stuck in "Scheduling" or "Pending" state

**Cause**: Worker node doesn't have `/dev/mshv` device

**Solution**:
```bash
# Check node for mshv device
kubectl debug node/<node-name> -it --image=ubuntu -- ls -l /dev/mshv

# If missing, verify node is running on Azure L1VH-capable VM
# May need to recreate node with appropriate VM size
```

#### 2. virt-launcher Pod Not Requesting mshv Resource

**Symptom**: Pod stuck in Pending state with "Insufficient devices.kubevirt.io/mshv" error

**Cause**: Device plugin not advertising mshv resources

**Solution**:
```bash
# Check if mshv resources are advertised
kubectl get nodes -o json | jq '.items[].status.allocatable | select(.["devices.kubevirt.io/mshv"] != null)'

# If not present, check virt-handler logs
kubectl logs -n kubevirt -l kubevirt.io=virt-handler
```

#### 3. Domain XML Shows type='kvm' Instead of type='hyperv'

**Symptom**: VMs still using KVM despite HyperV-Layered configuration

**Cause**: Configuration not applied or virt-launcher using cached configuration

**Solution**:
```bash
# Verify KubeVirt CR configuration
kubectl get kubevirt kubevirt -n kubevirt -o jsonpath='{.spec.configuration.hypervisorConfiguration}'

# Restart virt-controller and virt-handler
kubectl delete pod -n kubevirt -l kubevirt.io=virt-controller
kubectl delete pod -n kubevirt -l kubevirt.io=virt-handler

# Delete and recreate VMI
kubectl delete vmi <vmi-name> -n <namespace>
```

#### 4. Performance Issues

**Symptom**: VM performance is not as expected

**Diagnostics**:
```bash
# Check if VM is actually using mshv
kubectl exec -it virt-launcher-<vmi-name>-xxxxx -n <namespace> -- cat /proc/*/stat | grep mshv

# Monitor resource usage
kubectl top pod virt-launcher-<vmi-name>-xxxxx -n <namespace>

# Check kernel logs on node
kubectl debug node/<node-name> -it --image=ubuntu -- dmesg | grep -i mshv
```

## Migration from KVM to HyperV-Layered

To migrate an existing KubeVirt installation from KVM to HyperV-Layered:

1. **Prerequisites**: Ensure all nodes support `/dev/mshv` device

2. **Enable Configuration**: Update KubeVirt CR as shown in the Configuration section

3. **Restart VMs**: Existing VMs must be restarted to use the new hypervisor
   ```bash
   # For VirtualMachine objects (recommended)
   kubectl patch vm <vm-name> -n <namespace> --type merge -p '{"spec":{"running":false}}'
   kubectl patch vm <vm-name> -n <namespace> --type merge -p '{"spec":{"running":true}}'
   
   # For VirtualMachineInstance objects
   kubectl delete vmi <vmi-name> -n <namespace>
   # Recreate the VMI
   ```

4. **Verify**: Use verification steps above to confirm HyperV-Layered is active

**Note**: There is no automatic live migration between hypervisor types. VMs must be restarted.

## Reverting to KVM

To revert back to KVM:

1. Edit the KubeVirt CR:
   ```bash
   kubectl edit kubevirt kubevirt -n kubevirt
   ```

2. Change the hypervisor configuration:
   ```yaml
   spec:
     configuration:
       hypervisorConfiguration:
         name: kvm
   ```
   
   Or remove the hypervisorConfiguration section entirely (KVM is the default).

3. Restart all VMs as described in the migration section above.

## Additional Resources

- [Architecture Documentation](architecture.md)
- [Software Emulation](software-emulation.md)
- [KubeVirt Components](components.md)
- [L1HV Infrastructure Documentation](l1hv-infra/README.md)
- [HyperV-Layered Code Analysis](../specs/001-hyperv-layered/code-analysis.md)

## Support and Feedback

HyperV-Layered is an alpha feature. For issues, questions, or feedback:

- GitHub Issues: [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt/issues)
- Slack: [#kubevirt on Kubernetes Slack](https://kubernetes.slack.com/messages/kubevirt)
- Mailing List: [kubevirt-dev@googlegroups.com](mailto:kubevirt-dev@googlegroups.com)

## Feature Status

**Current Status**: Alpha

**Known Limitations**:
- Azure-only support
- No live migration between hypervisor types
- Cluster-wide configuration only
- Limited to x86_64 architecture

**Planned Enhancements**:
- Per-VM hypervisor selection
- Enhanced hardware passthrough capabilities
- Additional Azure VM size support
- Performance optimization for specific workloads
