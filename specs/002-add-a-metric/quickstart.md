# Quickstart Guide: VMI Hypervisor Tracking Metric

**Generated**: October 13, 2025  
**Feature**: VMI Hypervisor Tracking Metric

## Developer Quick Reference

### Overview
Add Prometheus metric `kubevirt_vmi_hypervisor_info` to track hypervisor type (KVM/QEMU-TCG/unknown) for each running VMI via virt-handler.

### Key Integration Points
- **Framework**: KubeVirt's `pkg/monitoring/metrics/virt-handler/` infrastructure
- **Pattern**: Static metric (like `versionInfo`, `machineTypeMetrics`) with VMI event handlers
- **Registration**: Add to existing `operatormetrics.RegisterMetrics()` call in `SetupMetrics()`
- **Updates**: Event-driven via VMI informer, not continuous collection

## Development Environment Setup

### Prerequisites
```bash
# Standard KubeVirt development environment
make cluster-up
```

### Build and Test Commands
```bash
# Generate code (if needed)
make generate

# Build images
make bazel-build-images

# Run unit tests
make bazel-test

# Run integration tests (specific to this feature)
make test-functional WHAT=hypervisor-metric
```

## Implementation Checklist

### Phase 1: Core Detection Logic
- [ ] Add hypervisor detection function to virt-handler
- [ ] Implement libvirt domain XML parsing
- [ ] Add hypervisor type enumeration (kvm/qemu-tcg/unknown)
- [ ] Unit tests for detection logic

### Phase 2: Metric Integration
- [ ] Create static hypervisor InfoVec metric in virt-handler
- [ ] Implement VMI informer event handlers (Add/Update/Delete)
- [ ] Register metric and handlers in SetupMetrics()
- [ ] Integration tests with VMI lifecycle events

### Phase 3: Error Handling & Polish  
- [ ] Handle libvirt connection failures gracefully
- [ ] Add proper logging and error recovery
- [ ] Performance testing and optimization
- [ ] End-to-end functional tests

## Code Structure Preview

### Primary Files to Modify
```text
pkg/monitoring/metrics/virt-handler/
├── hypervisor_metrics.go     # NEW: Static hypervisor InfoVec metric
├── hypervisor_metrics_test.go # NEW: Tests for event handlers and detection
└── metrics.go                # MODIFY: Add hypervisor metric registration
```

### Key Functions to Implement
```go
// pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go
func detectHypervisorType(vmi *v1.VirtualMachineInstance) (string, error)
func updateHypervisorMetric(vmi *v1.VirtualMachineInstance, hypervisorType string)
func handleVMIAdd(obj interface{})
func handleVMIUpdate(oldObj, newObj interface{})
func handleVMIDelete(obj interface{})

// pkg/monitoring/metrics/virt-handler/metrics.go (modify)
func SetupHypervisorMetrics() error // Add to existing SetupMetrics()
```

## Testing Strategy

### Unit Testing
```go
// Test hypervisor detection with various domain XML inputs
func TestDetectHypervisorType(t *testing.T) {
    testCases := []struct{
        domainXML string
        expected  HypervisorType
    }{
        {`<domain type="kvm">`, HypervisorTypeKVM},
        {`<domain type="qemu">`, HypervisorTypeQEMUTCG},
        {`<domain type="unknown">`, HypervisorTypeUnknown},
        {`invalid xml`, HypervisorTypeUnknown},
    }
    // ... test implementation
}
```

### Integration Testing
```go
// Test metric lifecycle with real VMI
func TestVMIHypervisorMetricLifecycle(t *testing.T) {
    // 1. Create VMI
    // 2. Wait for Running phase
    // 3. Verify metric exists with correct hypervisor_type
    // 4. Delete VMI
    // 5. Verify metric removed
}
```

### Functional Testing
```bash
# E2E test with actual VM creation
make cluster-up
kubectl apply -f test-vmi.yaml
# Wait for VMI Running
curl -k https://virt-handler:8443/metrics | grep kubevirt_vmi_hypervisor_info
# Verify metric appears with expected labels
```

## Debugging Guide

### Common Issues
1. **Metric not appearing**: Check VMI is in Running phase, verify libvirt domain exists
2. **Wrong hypervisor_type**: Check libvirt domain XML content, verify parsing logic
3. **Metric not removed**: Check VMI deletion events, verify cleanup logic

### Debug Commands
```bash
# Check VMI status
kubectl get vmi -A

# Check virt-handler logs
kubectl logs -n kubevirt daemonset/virt-handler

# Check metrics endpoint
kubectl port-forward -n kubevirt ds/virt-handler 8443:8443
curl -k https://localhost:8443/metrics | grep hypervisor

# Check libvirt domains on node
virsh list --all
virsh dumpxml <domain-name>
```

### Log Messages to Add
```go
log.Log.V(2).Infof("Detected hypervisor type %s for VMI %s/%s", hypervisorType, namespace, name)
log.Log.V(4).Infof("Domain XML for VMI %s/%s: %s", namespace, name, domainXML)
log.Log.Warningf("Failed to detect hypervisor for VMI %s/%s: %v", namespace, name, err)
```

## Performance Considerations

### Optimization Targets
- **Detection Latency**: <100ms per VMI
- **CPU Overhead**: <1% of virt-handler CPU usage
- **Memory Overhead**: <1MB for 1000 VMIs
- **Update Frequency**: Event-driven (not polling)

### Monitoring Performance
```promql
# Monitor metric collection performance
rate(kubevirt_vmi_hypervisor_detection_duration_seconds[5m])
histogram_quantile(0.95, kubevirt_vmi_hypervisor_detection_duration_seconds)
```

## Rollout Strategy

### Development Phases
1. **Local Development**: Single-node cluster testing
2. **CI Integration**: Automated testing in CI pipeline  
3. **Staging Deployment**: Multi-node cluster validation
4. **Production Rollout**: Gradual rollout with monitoring

### Feature Flag (Not Required)
No feature gate needed - metrics addition is non-disruptive.

### Rollback Plan
Remove metric collection code and redeploy virt-handler - no persistent state to clean up.

## Validation Checklist

Before submitting PR:
- [ ] All unit tests passing
- [ ] Integration tests added and passing
- [ ] Manual testing with various VMI configurations
- [ ] Performance impact measured and acceptable
- [ ] Documentation updated
- [ ] Code follows KubeVirt conventions
- [ ] No breaking changes to existing functionality

This quickstart provides the essential information for developers to implement the VMI hypervisor tracking metric efficiently and correctly.