# Integration Contracts: VMI Hypervisor Tracking Metric

**Generated**: October 13, 2025  
**Feature**: VMI Hypervisor Tracking Metric

## Component Contracts

### virt-handler Contract

**Responsibilities**:
- Monitor VMI lifecycle events via informers
- Detect hypervisor type via libvirt domain XML parsing  
- Emit and maintain `kubevirt_vmi_hypervisor_info` metric
- Handle metric lifecycle (create/update/remove)

**Interface Contract**:
```go
// Internal interface for hypervisor detection
type HypervisorDetector interface {
    // DetectHypervisorType queries libvirt for VMI hypervisor type
    DetectHypervisorType(vmiNamespace, vmiName string) (HypervisorType, error)
    
    // IsVMIRunning checks if VMI has an active libvirt domain
    IsVMIRunning(vmiNamespace, vmiName string) bool
}

// Metric emission interface
type HypervisorMetricsEmitter interface {
    // SetHypervisorInfo creates or updates the hypervisor metric for a VMI
    SetHypervisorInfo(namespace, name, node string, hypervisorType HypervisorType)
    
    // RemoveHypervisorInfo removes the hypervisor metric for a VMI
    RemoveHypervisorInfo(namespace, name string)
}
```

### libvirt Integration Contract

**Domain XML Query Pattern**:
```xml
<!-- Expected libvirt domain XML structure -->
<domain type="kvm">  <!-- or type="qemu" -->
  <name>vmi-name</name>
  <!-- ... other domain configuration ... -->
</domain>
```

**Detection Logic Contract**:
```go
// Domain type mapping contract
var hypervisorTypeMapping = map[string]HypervisorType{
    "kvm":  HypervisorTypeKVM,     // Hardware acceleration
    "qemu": HypervisorTypeQEMUTCG, // Software emulation
    // Any other type → HypervisorTypeUnknown
}
```

### Prometheus Metrics Contract

**Metric Specification**:
```yaml
# Metric exposition format
# HELP kubevirt_vmi_hypervisor_info Information about the hypervisor type used by a VirtualMachineInstance
# TYPE kubevirt_vmi_hypervisor_info info
kubevirt_vmi_hypervisor_info{namespace="ns",name="vmi",node="node1",hypervisor_type="kvm"} 1
```

**Label Constraints**:
- All labels must be valid Prometheus label names (no spaces, special chars)
- Values must be non-empty strings
- `hypervisor_type` must be one of: `kvm`, `qemu-tcg`, `unknown`

### VMI Lifecycle Integration Contract

**Event Triggers**:
```go
// VMI events that trigger hypervisor detection
type VMILifecycleEvent string

const (
    VMIStarted    VMILifecycleEvent = "Started"    // VMI enters Running phase
    VMIMigrated   VMILifecycleEvent = "Migrated"   // Post-migration
    VMIResumed    VMILifecycleEvent = "Resumed"    // Post-pause resume
    VMIStopped    VMILifecycleEvent = "Stopped"    // VMI leaves Running phase
    VMIDeleted    VMILifecycleEvent = "Deleted"    // VMI deleted
)
```

**Integration Points**:
- VMI informer event handlers in virt-handler
- Existing VMI monitoring loops
- Standard KubeVirt reconciliation patterns

## Error Handling Contracts

### Detection Error Response

```go
// Error handling contract for hypervisor detection
type DetectionResult struct {
    HypervisorType HypervisorType
    Error          error
    Confidence     ConfidenceLevel
}

type ConfidenceLevel string
const (
    ConfidenceHigh    ConfidenceLevel = "high"    // Successful detection
    ConfidenceLow     ConfidenceLevel = "low"     // Fallback/assumption
    ConfidenceUnknown ConfidenceLevel = "unknown" // Detection failed
)
```

### Fallback Behavior Contract

| Error Condition | Response | Metric Behavior |
|-----------------|----------|-----------------|
| libvirt unavailable | Log error, set unknown | `hypervisor_type="unknown"` |
| Domain not found | Skip metric | No metric created |
| XML parse failure | Log error, set unknown | `hypervisor_type="unknown"` |
| Unknown domain type | Log warning, set unknown | `hypervisor_type="unknown"` |

## Performance Contracts

### Response Time Guarantees
- **Hypervisor Detection**: <100ms per VMI
- **Metric Update**: <10ms per operation
- **Error Recovery**: <1s for retry attempts

### Resource Usage Limits
- **CPU Overhead**: <1% of virt-handler CPU per 100 VMIs
- **Memory Overhead**: <1MB per 1000 VMIs  
- **Network Overhead**: Negligible (local libvirt socket)

### Scalability Contracts
- **VMI Scale**: Support 1000+ concurrent VMIs per node
- **Update Frequency**: Handle 10+ VMI transitions per second
- **Metric Retention**: Maintain metrics for all running VMIs

## Backwards Compatibility Contract

### Non-Breaking Changes
- ✅ Adding new metric does not affect existing functionality
- ✅ No changes to existing APIs or interfaces
- ✅ No changes to VMI or VM resources
- ✅ Optional feature - does not block VMI operations

### Integration Safety
- Metric collection failures must not impact VMI lifecycle
- libvirt query errors must not crash virt-handler
- Unknown hypervisor types handled gracefully

## Testing Contracts

### Unit Test Coverage
- Hypervisor detection logic for all domain types
- Error handling for all failure modes
- Metric emission and lifecycle management
- Label validation and sanitization

### Integration Test Coverage  
- End-to-end VMI lifecycle with metric verification
- libvirt integration with real domain XML
- Prometheus metrics endpoint validation
- Multi-VMI concurrent operation testing

### Performance Test Coverage
- Metric collection overhead measurement
- Scale testing with 100+ VMIs
- Memory usage validation
- CPU impact assessment

This contract specification ensures clear boundaries and expectations for implementing the VMI hypervisor tracking metric while maintaining KubeVirt's reliability and performance standards.