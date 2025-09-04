# Implementation Plan: Hyper-V L1VH Support in KubeVirt

**Branch**: `001-hyperv-l1vh` | **Date**: 2025-09-03 | **Spec**: [./feature-spec.md](./feature-spec.md)

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Constitutional Compliance](#constitutional-compliance)
- [Implementation Architecture](#implementation-architecture)
- [Component Analysis](#component-analysis)
- [No API Changes Required](#no-api-changes-required)
- [Code Implementation](#code-implementation)
- [Testing Strategy](#testing-strategy)
- [Build and Packaging](#build-and-packaging)
- [Documentation](#documentation)
- [Rollout Plan](#rollout-plan)
- [Success Criteria](#success-criteria)

---

## Executive Summary

This implementation plan details the technical approach for adding **transparent** Hyper-V L1VH (Level 1 Virtual Host) support to KubeVirt v1.5.0+. The feature enables existing VirtualMachine resources to automatically benefit from L1VH performance optimization when deployed on L1VH-capable clusters, **without requiring any API changes or user configuration**.

### Key Implementation Points

- **Zero API Changes**: Existing VirtualMachine resources work transparently with L1VH optimization
- **Transparent Operation**: VMs automatically use L1VH when feature gate is enabled on L1VH-capable clusters
- **Cluster-Wide Assumption**: All nodes are assumed L1VH-capable when feature gate is enabled
- **Feature Gate Controlled**: `HyperVL1VH` alpha gate for safe cluster-wide rollout
- **Non-Breaking**: Maintains 100% backward compatibility with existing QEMU/KVM workloads
- **Simplified Architecture**: No complex scheduling, node labeling, or user decision-making required

---

## Constitutional Compliance

This implementation strictly adheres to the **KubeVirt Constitution v1.0**:

### ✅ Article II: The KubeVirt Razor
- **Native Kubernetes Integration**: No new APIs or VM-specific abstractions created
- **Non-Privileged Extensions**: No additional capabilities beyond existing Kubernetes patterns
- **Choreography Pattern**: Components act independently based on observed state

### ✅ Article III: Feature Gate Discipline  
- **Alpha-First Development**: Feature begins disabled behind `HyperVL1VH` gate
- **Complete Disable/Enable**: System functions identically to pre-feature state when disabled

### ✅ Article VII: Simplicity and Minimalism
- **Minimal Implementation**: Simplest possible approach using existing VirtualMachine resources
- **No Premature Optimization**: Direct integration without unnecessary abstractions

### ✅ Article VIII: Anti-Abstraction
- **Framework Trust**: Direct use of libvirt and QEMU L1VH capabilities 
- **Single Model**: No parallel hypervisor selection or configuration systems

---

## Implementation Architecture

### High-Level Flow (Transparent to Users)

```
VirtualMachine (Standard API - NO CHANGES)
    ↓ (user submits existing VM spec)
virt-controller 
    ↓ (creates VMI with no special handling required)
virt-handler 
    ↓ (standard lifecycle management, assumes L1VH cluster)
virt-launcher 
    ↓ (detects L1VH capability, automatically uses optimal hypervisor)
libvirt + QEMU 
    ↓ (automatically connects to /dev/mshv when available, QEMU/KVM otherwise)
Microsoft HyperV L1VH (when feature gate enabled) | QEMU/KVM (when disabled)
```

### Simplified Integration Points

The **minimal** integration points required:

1. **Feature Gate Detection** (virt-launcher converter)
2. **Automatic Hypervisor Selection** (converter.go hypervisor type detection)
3. **Transparent Operation** (no component changes needed for scheduling or validation)

---

## Component Analysis

### 1. virt-controller

**Location:** `cmd/virt-controller/`  
**Responsibilities:** VirtualMachine resource reconciliation

#### Changes Required: **MINIMAL TO NONE**

Per the current spec and KubeVirt Constitution Article VII (Simplicity), virt-controller requires **NO CHANGES** for L1VH support since:

- No API validation changes needed (existing VM specs work transparently)
- No scheduling complexity (cluster-wide L1VH assumption eliminates node selection)
- No admission webhook extensions (hypervisor selection happens in virt-launcher)

**Implementation Impact**: **Zero code changes** - existing VM lifecycle management continues unchanged.

### 2. virt-handler

**Location:** `cmd/virt-handler/`  
**Responsibilities:** Node-level VM lifecycle management

#### Changes Required: **NONE**

Per constitutional principles and cluster-wide L1VH assumption:

- **No Node Capability Detection**: All nodes assumed L1VH-capable when feature gate enabled
- **No Device Management**: `/dev/mshv` access handled by existing device patterns
- **No Special Lifecycle**: Standard VM start/stop procedures work transparently
- **No Node Labeling**: Cluster-wide capability assumption eliminates node labeling complexity

**Implementation Impact**: **Zero code changes** - virt-handler operates identically for all VMs.

### 3. virt-launcher (PRIMARY INTEGRATION POINT)

**Location:** `cmd/virt-launcher/`  
**Responsibilities:** VM process management and hypervisor integration

#### Changes Required: **MINIMAL, TARGETED**

**Single Integration Point**: Extend converter's hypervisor detection logic to automatically select L1VH when:
1. Feature gate `HyperVL1VH` is enabled
2. `/dev/mshv` device is available
3. No user configuration required

**Key Files:**
- `pkg/virt-launcher/virtwrap/converter/converter.go` (hypervisor type detection)

### 4. virt-operator

**Location:** `cmd/virt-operator/`  
**Responsibilities:** KubeVirt configuration and feature gate management

#### Changes Required: **MINIMAL**

- **Feature Gate Registration**: Add `HyperVL1VH` to available feature gates
- **L1VH Cluster Validation**: Validate cluster-wide L1VH readiness when feature gate is enabled

**Key Files:**
- `pkg/virt-config/featuregate/feature-gates.go`

---

## No API Changes Required

**CRITICAL**: This feature requires **ZERO API changes**. Per the current specification and constitutional Article II (KubeVirt Razor), existing VirtualMachine resources work transparently with L1VH optimization.

### Existing API Works Transparently

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: example-vm
spec:
  template:
    spec:
      domain:
        # Standard KubeVirt VM specification - NO CHANGES NEEDED
        # L1VH optimization happens automatically when feature gate is enabled
        resources:
          requests:
            memory: "2Gi"
            cpu: "1"
        devices:
          disks:
          - name: containerdisk
            disk:
              bus: virtio
      volumes:
      - name: containerdisk
        containerDisk:
          image: kubevirt/fedora-cloud-container-disk-demo
```

### Transparent Status Reporting

```yaml
status:
  # Standard status fields work identically
  phase: "Running"
  interfaces: [...]
  # Optional: Internal observability field
  conditions:
  - type: "HypervisorOptimized"
    status: "True"
    reason: "L1VHEnabled" 
    message: "VM automatically optimized with L1VH"
```

### No User Configuration Required

- **Automatic Optimization**: VMs use L1VH when feature gate enabled and `/dev/mshv` available
- **Fallback Behavior**: VMs use QEMU/KVM when L1VH unavailable
- **Zero User Impact**: Users submit identical VM specifications regardless of underlying hypervisor
- **Constitutional Compliance**: Follows "useful for Pods" principle - no VM-specific configuration

---

## Code Implementation

### 1. Feature Gate Registration (virt-operator)

**File:** `pkg/virt-config/featuregate/feature-gates.go`

```go
const (
    // Existing feature gates...
    HyperVL1VH featuregate.FeatureGate = "HyperVL1VH"
)

var defaultKubeVirtFeatureGates = map[featuregate.FeatureGate]featuregate.FeatureSpec{
    // Existing gates...
    HyperVL1VH: {Stage: featuregate.Alpha, Default: false},
}
```

### 2. Transparent Hypervisor Detection (virt-launcher - PRIMARY CHANGE)

**File:** `pkg/virt-launcher/virtwrap/converter/converter.go`

**Constitutional Compliance**: Single, minimal change following Article VII (Simplicity) and Article VIII (Anti-Abstraction).

```go
// Extend existing hypervisor type detection (around existing logic)
func (c *ConverterContext) detectHypervisorType(vmi *v1.VirtualMachineInstance) (string, error) {
    // L1VH Auto-Detection (when feature gate enabled)
    if featuregate.DefaultFeatureGate.Enabled(featuregate.HyperVL1VH) {
        if c.HasL1VHSupport() {
            log.Info("L1VH detected and enabled, using optimized hypervisor")
            return "mshv", nil
        }
        log.Info("L1VH feature gate enabled but /dev/mshv unavailable, using QEMU/KVM")
    }
    
    // Existing QEMU/KVM detection logic (unchanged)
    return c.detectQEMUKVMType(vmi)
}

func (c *ConverterContext) HasL1VHSupport() bool {
    // Simple, direct check for L1VH capability
    if _, err := os.Stat("/dev/mshv"); err != nil {
        return false
    }
    return true
}
```

### 3. L1VH Domain Configuration (virt-launcher)

**File:** `pkg/virt-launcher/virtwrap/converter/converter.go`

```go
// Extend Convert_v1_VirtualMachineInstance_To_api_Domain function
func Convert_v1_VirtualMachineInstance_To_api_Domain(vmi *v1.VirtualMachineInstance, c *ConverterContext) (*api.Domain, error) {
    // Existing conversion logic...
    
    // Automatic hypervisor type selection
    hypervisorType, err := c.detectHypervisorType(vmi)
    if err != nil {
        return nil, err
    }
    
    domain.Spec.Type = hypervisorType
    
    // L1VH-specific optimizations (when using mshv hypervisor)
    if hypervisorType == "mshv" {
        if err := c.optimizeForL1VH(domain, vmi); err != nil {
            log.Infof("L1VH optimization failed, continuing with standard config: %v", err)
        }
    }
    
    return domain, nil
}

func (c *ConverterContext) optimizeForL1VH(domain *api.Domain, vmi *v1.VirtualMachineInstance) error {
    // Apply L1VH-specific optimizations
    // This follows existing converter patterns for hypervisor-specific features
    return nil
}
```

### Implementation Scope

**Total Changes Required:**
- **1 file modified**: `pkg/virt-launcher/virtwrap/converter/converter.go`
- **1 file modified**: `pkg/virt-config/featuregate/feature-gates.go`
- **~50 lines of code total**
- **Zero breaking changes**
- **Constitutional compliance**: Minimal, direct implementation

---

## Testing Strategy

### 1. Constitutional Test-First Development

**CRITICAL**: Per KubeVirt Constitution Article IV, **ALL TESTS MUST BE WRITTEN AND APPROVED BEFORE IMPLEMENTATION**.

### 2. Test Hierarchy (Constitutional Requirement)

Following constitutional Article IV (Test-First Implementation), tests **MUST** be implemented in this order:

#### 2.1 Contract Tests (First Priority)
```go
// pkg/virt-launcher/virtwrap/converter/converter_test.go
func TestL1VHTransparentSelection(t *testing.T) {
    // Test transparent hypervisor selection with feature gate enabled
    vmi := libvmi.New(
        libvmi.WithContainerDisk("test", "kubevirt/cirros-container-disk-demo"),
    )
    
    // Mock L1VH availability
    ctx := &ConverterContext{
        L1VHAvailable: true,
        FeatureGate: mockFeatureGate(HyperVL1VH, true),
    }
    
    domain, err := Convert_v1_VirtualMachineInstance_To_api_Domain(vmi, ctx)
    Expect(err).ToNot(HaveOccurred())
    Expect(domain.Spec.Type).To(Equal("mshv"))
}

func TestL1VHFallbackToQEMU(t *testing.T) {
    // Test fallback to QEMU/KVM when L1VH unavailable
    vmi := libvmi.New()
    
    ctx := &ConverterContext{
        L1VHAvailable: false,
        FeatureGate: mockFeatureGate(HyperVL1VH, true),
    }
    
    domain, err := Convert_v1_VirtualMachineInstance_To_api_Domain(vmi, ctx)
    Expect(err).ToNot(HaveOccurred()) 
    Expect(domain.Spec.Type).To(Equal("qemu"))
}
```

#### 2.2 Integration Tests (Second Priority)
```go
// tests/hyperv_l1vh_test.go
var _ = Describe("[sig-compute]HyperV L1VH Transparent", decorators.SigCompute, func() {
    BeforeEach(func() {
        checks.SkipTestIfNoFeatureGate(featuregate.HyperVL1VH)
        checks.SkipIfNoL1VHSupport()
    })
    
    It("should transparently create VM with L1VH optimization", func() {
        // Use standard VM specification - no L1VH-specific config
        vm := libvmifact.NewCirros()
        
        vm, err = virtClient.VirtualMachine(testsuite.GetTestNamespace(nil)).Create(context.Background(), vm, metav1.CreateOptions{})
        Expect(err).ToNot(HaveOccurred())
        
        // VM should start and run successfully with transparent L1VH optimization
        Eventually(func() bool {
            vm, err := virtClient.VirtualMachine(vm.Namespace).Get(context.Background(), vm.Name, metav1.GetOptions{})
            return err == nil && vm.Status.Ready
        }, 300*time.Second, 1*time.Second).Should(BeTrue())
    })
})
```

#### 2.3 End-to-End Tests (Third Priority)
```go
// tests/e2e_hyperv_l1vh_test.go
It("should handle complete VM lifecycle with L1VH", func() {
    // Standard VM lifecycle with transparent L1VH optimization
    vm := libvmifact.NewAlpine()
    
    // Create, start, verify, and cleanup
    vm, err = virtClient.VirtualMachine(namespace).Create(context.Background(), vm, metav1.CreateOptions{})
    Expect(err).ToNot(HaveOccurred())
    
    vm = tests.StartVirtualMachine(vm)
    libwait.WaitForSuccessfulVMIStart(vm)
    
    // Verify transparent optimization
    vmi, err := virtClient.VirtualMachineInstance(namespace).Get(context.Background(), vm.Name, metav1.GetOptions{})
    Expect(err).ToNot(HaveOccurred())
    
    // Cleanup
    err = virtClient.VirtualMachine(namespace).Delete(context.Background(), vm.Name, metav1.DeleteOptions{})
    Expect(err).ToNot(HaveOccurred())
})
```

### 3. Test Environment Requirements

**Constitutional Requirement**: Real environments over mocks (Article IV)

- **L1VH Cluster**: Azure VMs with L1VH capability or nested VMs with `/dev/mshv` support
- **Feature Gate Control**: Ability to enable/disable `HyperVL1VH` gate
- **Transparent Testing**: Validate existing test suites work unchanged with L1VH

### 4. Performance Validation Tests

```go
func TestL1VHPerformanceOptimization(t *testing.T) {
    // Validate L1VH provides performance benefits over nested virtualization
    // Compare VM startup times, I/O performance, CPU performance
}
```

---

## Build and Packaging

### 1. Bazel Build Integration

**File:** `cmd/virt-launcher/BUILD.bazel`

```bazel
# Extend existing virt-launcher build
go_library(
    name = "go_default_library",
    srcs = glob(["*.go"]),
    deps = [
        # Existing dependencies...
        "//pkg/virt-launcher/virtwrap/converter:go_default_library",
    ],
)
```

### 2. Container Image Updates

**File:** `images/BUILD.bazel`

No changes required - HyperV support is provided through:
- Host kernel module (`/dev/mshv`)
- libvirt driver (container runtime dependency)
- QEMU mshv backend (container runtime dependency)

### 3. Multi-Architecture Support

Following KubeVirt's existing patterns in `BUILD.bazel`:

```bazel
# HyperV L1VH only supported on x86_64
select({
    "@io_bazel_rules_go//go/platform:linux_amd64": [
        "//pkg/hyperv:go_default_library",
    ],
    "//conditions:default": [],
})
```

### 4. RPM Dependencies

**File:** `rpm/BUILD.bazel`

HyperV support requires updated dependencies:
- libvirt (mshv driver support)
- QEMU (mshv backend support)
- Linux kernel (mshv module)

---

## Documentation

### 1. User Documentation

**File:** `docs/hyperv-l1vh.md`

```markdown
# HyperV L1VH Support in KubeVirt

## Overview
KubeVirt supports Microsoft HyperV L1VH (Level 1 Virtual Host) as an alternative hypervisor to QEMU/KVM.

## Prerequisites
- Nodes with /dev/mshv device
- libvirt with HyperV driver
- Feature gate HyperVL1VH enabled

## Usage
```yaml
spec:
  domain:
    features:
      hyperv:
        l1vh:
          enabled: true
```

### 2. Developer Documentation

**File:** `docs/architecture-hyperv.md`

Architecture documentation explaining:
- HyperV integration points
- libvirt domain XML differences
- Performance characteristics
- Troubleshooting guide

### 3. API Documentation

**OpenAPI Spec Updates:**
- `api/openapi-spec/swagger.json`
- Field descriptions for HyperV L1VH configuration
- Example manifests

---

## Rollout Plan

### Phase 1: Alpha Release (v1.5.0)

**Scope:** Transparent L1VH support behind feature gate

**Constitutional Gates (MUST PASS BEFORE IMPLEMENTATION):**
- [x] **Simplicity Gate**: ≤3 Go packages, minimal implementation ✅
- [x] **Anti-Abstraction Gate**: Direct libvirt/QEMU integration ✅  
- [x] **Integration-First Gate**: Real environment testing planned ✅
- [x] **KubeVirt Razor Gate**: No VM-specific APIs, native Kubernetes patterns ✅

**Deliverables:**
- [ ] **Tests Written First** (Constitutional requirement)
- [ ] Feature gate registration (`HyperVL1VH`)
- [ ] Transparent hypervisor detection in virt-launcher converter
- [ ] Integration tests with L1VH clusters
- [ ] Basic documentation
- [ ] Constitutional compliance validation

**Success Criteria:**
- VMs automatically use L1VH when feature gate enabled and `/dev/mshv` available
- VMs fall back to QEMU/KVM when L1VH unavailable
- Feature gate controls access cluster-wide
- Zero regression in existing QEMU/KVM functionality
- **Constitutional Compliance**: All constitutional gates passed

### Phase 2: Beta Release (v1.6.0)

**Scope:** Production-ready transparent L1VH with performance validation

**Deliverables:**
- [ ] Comprehensive end-to-end test coverage
- [ ] Performance benchmarking against nested virtualization
- [ ] Hardware passthrough validation (GPU, storage, networking)
- [ ] Production deployment guides for L1VH clusters
- [ ] Monitoring and observability integration

**Success Criteria:**
- Feature gate promoted to beta (default: false)
- L1VH performance benefits measured and documented
- Hardware passthrough capabilities validated
- Production deployment patterns established

### Phase 3: GA Release (v1.7.0)

**Scope:** General availability with full L1VH optimization

**Deliverables:**
- [ ] Feature gate promoted to GA (default: true for L1VH-capable clusters)
- [ ] Complete hardware acceleration support
- [ ] Migration support for L1VH VMs
- [ ] Advanced monitoring and troubleshooting tools

**Success Criteria:**
- Feature enabled by default on L1VH-capable clusters
- Full parity with QEMU/KVM feature set
- Enterprise-ready with complete operational support

---

## Success Criteria

### Constitutional Compliance ✅

This implementation meets all constitutional requirements:

- **Article II (KubeVirt Razor)**: No VM-specific APIs, uses existing Kubernetes patterns
- **Article III (Feature Gate Discipline)**: Alpha-first with proper gate implementation  
- **Article IV (Test-First Implementation)**: TDD mandated before any code development
- **Article VII (Simplicity)**: Minimal implementation using ≤2 Go packages, no premature optimization
- **Article VIII (Anti-Abstraction)**: Direct libvirt/QEMU integration without wrapper layers

### Technical Success Criteria

1. **Transparent Operation**
   - [ ] VMs automatically use L1VH when feature gate enabled and `/dev/mshv` available
   - [ ] VMs seamlessly fall back to QEMU/KVM when L1VH unavailable  
   - [ ] Users remain unaware of underlying hypervisor selection
   - [ ] Standard kubectl commands work identically across hypervisors

2. **Performance Objectives**
   - [ ] L1VH eliminates nested virtualization performance penalty
   - [ ] Hardware passthrough enables direct GPU/storage/network access
   - [ ] VM startup times comparable to or better than QEMU/KVM
   - [ ] Memory usage impact <5% compared to QEMU/KVM baseline

3. **Feature Integration**
   - [ ] All existing KubeVirt features work transparently with L1VH
   - [ ] Live migration functions correctly between L1VH-capable nodes
   - [ ] Storage and networking integrations work without modification
   - [ ] Monitoring and observability work identically

### User Experience Success Criteria

1. **Zero Configuration Required**
   - [ ] Existing VM specifications work without modification
   - [ ] No additional RBAC permissions or cluster setup beyond feature gate
   - [ ] Clear documentation for L1VH cluster requirements
   - [ ] Transparent optimization without user decision-making

2. **Operational Simplicity**
   - [ ] Cluster-wide L1VH assumption eliminates complex scheduling
   - [ ] No node-level capability detection or labeling required
   - [ ] Clear error messages when L1VH requirements not met
   - [ ] Rollback procedures for feature gate disable

### Quality Assurance Criteria

1. **Test Coverage**
   - [ ] Unit test coverage >80% for modified components
   - [ ] Integration tests cover transparent hypervisor selection
   - [ ] E2E tests validate complete VM lifecycle with L1VH
   - [ ] Performance tests demonstrate L1VH benefits

2. **Production Readiness**
   - [ ] Security review completed for `/dev/mshv` integration
   - [ ] Documentation complete for cluster setup and troubleshooting
   - [ ] Monitoring integration provides L1VH-specific metrics
   - [ ] Support procedures defined for L1VH-related issues

---
## Open Questions
1. **Libvirt and QEMU Version Dependencies**: How can we retrieve the minimum required versions of libvirt and QEMU for L1VH support?

## Risk Assessment and Mitigation

### Implementation Risks

1. **Constitutional Compliance Risk**
   - *Risk*: Implementation becomes complex, violating constitutional simplicity
   - *Mitigation*: Strict adherence to ≤2 Go packages, regular constitutional review

2. **L1VH Driver Stability**
   - *Risk*: `/dev/mshv` driver instability affecting VM operations
   - *Mitigation*: Graceful fallback to QEMU/KVM, comprehensive error handling

3. **Transparent Operation Complexity**
   - *Risk*: Automatic hypervisor selection introduces unexpected behaviors
   - *Mitigation*: Clear logging, comprehensive testing, predictable fallback rules

### Operational Risks

1. **Cluster-Wide Assumption Risk**
   - *Risk*: L1VH assumption fails if some nodes lack capability
   - *Mitigation*: Clear documentation of L1VH cluster requirements, validation tools

2. **Performance Regression Risk**
   - *Risk*: L1VH integration impacts QEMU/KVM performance
   - *Mitigation*: Isolated code paths, performance monitoring, feature gate isolation

---

**Next Steps:**
1. **Constitutional Validation**: Verify all constitutional gates are met
2. **Test-First Development**: Write and approve all tests before implementation
3. **Minimal Implementation**: Implement only the converter changes required
4. **Performance Validation**: Establish benchmarking framework for L1VH benefits

---

*This implementation plan strictly follows the KubeVirt Constitution v1.0 and implements transparent Hyper-V L1VH support that eliminates nested virtualization performance penalties while maintaining 100% backward compatibility with existing KubeVirt workloads.*
