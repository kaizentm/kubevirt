# Component Integration: VMI Hypervisor Tracking Metric

**Generated**: October 13, 2025  
**Feature**: VMI Hypervisor Tracking Metric

## KubeVirt Component Integration Architecture

### Primary Component: virt-handler

**Role**: Node-level agent with established metrics infrastructure at `pkg/monitoring/metrics/virt-handler/`

```text
┌───────────────────────────────────────────────────────────────────────────────┐
│                          virt-handler                                         │
│  ┌───────────────────────────────────────────────────────────────────────────┐ │
│  │              pkg/monitoring/metrics/virt-handler/                         │ │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐   │ │
│  │  │ domainstats/    │  │ migration       │  │ hypervisor_metrics.go   │   │ │
│  │  │ (existing)      │  │ domainstats/    │  │ (NEW - static metric)   │   │ │
│  │  │                 │  │ (existing)      │  │                         │   │ │
│  │  │ - Dynamic stats │  │                 │  │ - InfoVec metric        │   │ │
│  │  │ - Collectors    │  │ - Migration     │  │ - Event handlers        │   │ │
│  │  │ - VMI reports   │  │   tracking      │  │ - Hypervisor detection  │   │ │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────────────┘   │ │
│  │                               │                        │                  │ │
│  │                               ▼                        ▼                  │ │
│  │  ┌───────────────────────────────────────────────────────────────────────┐ │ │
│  │  │                    metrics.go (SetupMetrics)                          │ │ │
│  │  │  - RegisterMetrics(versionMetrics, machineTypeMetrics,               │ │ │
│  │  │                    hypervisorMetrics) ← NEW                          │ │ │
│  │  │  - RegisterCollector(domainstats, migrationstats)                    │ │ │
│  │  │  - VMI informer event handlers ← NEW                                 │ │ │
│  │  └───────────────────────────────────────────────────────────────────────┘ │ │
│  └───────────────────────────────────────────────────────────────────────────┘ │
│                                        │                                       │
│                                        ▼                                       │
│  ┌───────────────────────────────────────────────────────────────────────────┐ │
│  │         operator-observability-toolkit /metrics Endpoint                 │ │
│  └───────────────────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────────────────┘
```

### Integration Flow

#### 1. Integration with Existing Static Metrics

```go
// pkg/monitoring/metrics/virt-handler/metrics.go (MODIFY)
func SetupMetrics(nodeName string, MaxRequestsInFlight int, vmiInformer cache.SharedIndexInformer, machines []libvirtxml.CapsGuestMachine) error {
    // ... existing setup ...
    
    // MODIFY: Add hypervisor metrics to static metrics registration
    if err := operatormetrics.RegisterMetrics(versionMetrics, machineTypeMetrics, hypervisorMetrics); err != nil {
        return err
    }
    SetVersionInfo()
    ReportDeprecatedMachineTypes(machines, nodeName)
    
    // NEW: Setup hypervisor metrics with VMI informer event handlers
    if err := SetupHypervisorMetrics(vmiInformer); err != nil {
        return err
    }
    
    // Existing collector registration (no changes)
    return operatormetrics.RegisterCollector(
        domainstats.Collector,
        domainstats.DomainDirtyRateStatsCollector,
        migrationdomainstats.MigrationStatsCollector,
        domainstats.GuestAgentInfoCollector,
    )
}
```

#### 2. VMI Event Handler Integration

```go
// pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go (NEW)
// Following static metrics pattern (like versionInfo, machineTypeMetrics)

func SetupHypervisorMetrics(vmiInformer cache.SharedIndexInformer) error {
    // Add VMI informer event handlers for lifecycle-driven updates
    vmiInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
        AddFunc:    onVMIAdd,
        UpdateFunc: onVMIUpdate,
        DeleteFunc: onVMIDelete,
    })
    return nil
}

func onVMIAdd(obj interface{}) {
    vmi := obj.(*v1.VirtualMachineInstance)
    if vmi.Status.Phase == v1.Running {
        updateHypervisorMetric(vmi)
    }
}

func onVMIUpdate(oldObj, newObj interface{}) {
    oldVMI := oldObj.(*v1.VirtualMachineInstance)
    newVMI := newObj.(*v1.VirtualMachineInstance)
    
    // Handle phase transitions
    if oldVMI.Status.Phase != v1.Running && newVMI.Status.Phase == v1.Running {
        updateHypervisorMetric(newVMI)
    } else if oldVMI.Status.Phase == v1.Running && newVMI.Status.Phase != v1.Running {
        removeHypervisorMetric(oldVMI)
    }
}
```

#### 3. Hypervisor Detection Integration

```go
// pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go (continued)

func updateHypervisorMetric(vmi *v1.VirtualMachineInstance) {
    hypervisorType, err := detectVMIHypervisorType(vmi)
    if err != nil {
        log.Log.V(3).Infof("Cannot detect hypervisor for VMI %s/%s, using unknown: %v", 
            vmi.Namespace, vmi.Name, err)
        hypervisorType = "unknown"
    }
    
    SetVMIHypervisorInfo(vmi.Namespace, vmi.Name, vmi.Status.NodeName, hypervisorType)
}

func removeHypervisorMetric(vmi *v1.VirtualMachineInstance) {
    // Need to determine what hypervisor type was set to delete the right metric
    // This is a limitation of the InfoVec pattern - we'll need to track or query all possible values
    for _, hypervisorType := range []string{"kvm", "qemu-tcg", "unknown"} {
        RemoveVMIHypervisorInfo(vmi.Namespace, vmi.Name, vmi.Status.NodeName, hypervisorType)
    }
}

// Simple libvirt XML parsing - no need for complex connection management
func detectVMIHypervisorType(vmi *v1.VirtualMachineInstance) (string, error) {
    // Use existing libvirt patterns from virt-handler to get domain XML
    // Much simpler than domainstats since we only need type attribute, not stats
    domainName := api.VMINamespaceKeyFunc(vmi)
    return getHypervisorTypeFromDomain(domainName)
}
```

## Component Interaction Patterns

### virt-handler ↔ libvirt Integration

```text
Existing Pattern (Leveraged):
┌─────────────────┐    gRPC/libvirt API    ┌─────────────────┐
│   virt-handler  │◄──────────────────────►│    libvirtd     │
│                 │                        │                 │
│ - Domain mgmt   │                        │ - Domain XML    │
│ - Lifecycle     │                        │ - Type info     │
│ - Monitoring    │                        │ - Status        │
└─────────────────┘                        └─────────────────┘

New Usage:
- Query domain XML for type attribute
- Parse <domain type="kvm|qemu"> 
- Handle connection failures gracefully
```

### virt-handler ↔ Kubernetes API Integration

```text
Existing Pattern (Leveraged):
┌─────────────────┐    K8s API Watch/List   ┌─────────────────┐
│   virt-handler  │◄──────────────────────►│  kube-apiserver │
│                 │                        │                 │
│ - VMI Informer  │                        │ - VMI Resources │
│ - Event Watch   │                        │ - Status Updates│
│ - Status Update │                        │ - Event Stream  │
└─────────────────┘                        └─────────────────┘

New Usage:
- React to VMI phase changes (→ Running, → Terminated)
- Extract VMI metadata (namespace, name, node)
- Use existing informer infrastructure
```

### virt-handler ↔ Prometheus Integration

```text
Existing Pattern (Extended):
┌─────────────────┐    HTTP /metrics        ┌─────────────────┐
│   virt-handler  │◄──────────────────────►│   Prometheus    │
│                 │                        │                 │
│ - Metrics       │                        │ - Scraping      │
│   Endpoint      │                        │ - Service       │
│ - Standard      │                        │   Discovery     │
│   Collectors    │                        │                 │
└─────────────────┘                        └─────────────────┘

New Addition:
+ kubevirt_vmi_hypervisor_info metric
+ InfoVec with standard labels
+ Lifecycle management (create/update/delete)
```

## Choreography Pattern Compliance

### Event-Driven Architecture

```text
VMI Lifecycle Event Flow:
┌─────────────┐    VMI Status Change    ┌─────────────────┐
│ K8s API     │────────────────────────▶│ virt-handler    │
│ Server      │                        │ VMI Controller  │
└─────────────┘                        └─────────────────┘
                                                │
                                        Detect & React
                                                ▼
┌─────────────┐    Query Domain XML     ┌─────────────────┐
│ libvirtd    │◄───────────────────────│ Hypervisor      │
│             │                        │ Detector        │
└─────────────┘                        └─────────────────┘
                                                │
                                        Update Metric
                                                ▼
┌─────────────┐    Metric Update        ┌─────────────────┐
│ Prometheus  │◄───────────────────────│ Metrics         │
│ Endpoint    │                        │ Collector       │
└─────────────┘                        └─────────────────┘
```

**No Cross-Component Communication**: Self-contained within virt-handler using existing patterns.

### Error Handling Integration

```go
// Integration with existing error handling patterns
func (c *VMIController) updateHypervisorMetric(vmi *v1.VirtualMachineInstance) error {
    hypervisorType, err := c.detectHypervisorType(vmi)
    if err != nil {
        // Follow existing logging patterns
        log.Log.V(3).Infof("Cannot detect hypervisor for VMI %s/%s, using unknown: %v", 
            vmi.Namespace, vmi.Name, err)
        hypervisorType = HypervisorTypeUnknown
    }
    
    // Update metric (never fails - fire-and-forget)
    c.metricsCollector.SetVMIHypervisorInfo(
        vmi.Namespace, 
        vmi.Name, 
        vmi.Status.NodeName,
        string(hypervisorType),
    )
    
    return nil // Never fail VMI processing due to metrics
}
```

## Configuration Integration

### No Configuration Required

✅ **Zero-Config Feature**: Leverages existing configurations without requiring new settings.

```yaml
# Existing virt-handler configuration (no changes)
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubevirt-config
data:
  # No new configuration needed
  # Uses existing:
  # - libvirt connection settings
  # - metrics collection settings  
  # - VMI monitoring settings
```

### Integration with Feature Gates (N/A)

```go
// No feature gate required - metrics addition is non-disruptive
// Follows KubeVirt pattern: additive observability features don't require gates
```

## Testing Integration

### Unit Testing Integration

```go
// Integration with existing test patterns
func TestVMIHypervisorDetection(t *testing.T) {
    // Use existing test utilities
    vmi := api.NewMinimalVMI("test-vm")
    vmi.Status.Phase = v1.Running
    vmi.Status.NodeName = "test-node"
    
    // Mock libvirt response using established patterns
    mockDomain := &libvirt.Domain{
        XML: `<domain type="kvm"><name>test-vm</name></domain>`,
    }
    
    // Test detection logic
    hypervisorType := detectHypervisorType(mockDomain.XML)
    assert.Equal(t, HypervisorTypeKVM, hypervisorType)
}
```

### Integration Testing Integration

```go
// Integration with existing functional test framework
func TestVMIHypervisorMetricE2E(t *testing.T) {
    // Use existing test VMI creation utilities
    vmi := tests.NewRandomVMI()
    
    By("Creating VMI")
    vmi, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
    Expect(err).ToNot(HaveOccurred())
    
    By("Waiting for VMI to be running")
    tests.WaitForSuccessfulVMIStart(vmi)
    
    By("Checking hypervisor metric exists")
    Eventually(func() bool {
        return checkHypervisorMetricExists(vmi.Namespace, vmi.Name)
    }, 30*time.Second, 1*time.Second).Should(BeTrue())
}
```

## Deployment Integration

### Container Image Integration

```dockerfile
# No changes to virt-handler Dockerfile
# Code integrated into existing virt-handler binary
# Uses existing dependencies (libvirt, prometheus client)
```

### Kubernetes Deployment Integration

```yaml
# No changes to virt-handler DaemonSet
# Existing deployment already includes:
# - libvirt socket access (/var/run/libvirt)
# - metrics port exposure (8443)
# - VMI RBAC permissions
# - Node-level scheduling
```

This component integration design ensures the hypervisor tracking metric integrates seamlessly with KubeVirt's existing architecture, following established patterns and maintaining the choreography-based design principles.