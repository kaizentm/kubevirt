# Implementation Plan: Hyper-V L1VH Support

**Branch**: `spec/hyperv-l1vh` | **Date**: 2025-09-04 | **Spec**: [feature-spec.md](./feature-spec.md)
**Input**: Feature specification from `/specs/001-hyperv-l1vh/feature-spec.md`

## Prerequisites

Before implementation begins, ensure:
- [x] Feature specification is approved and stable
- [x] All dependencies are available in target versions  
- [x] Development environment supports required tools and libraries
- [x] Required approvals and permissions are obtained

## Architecture Overview

### Component Integration Map
**IMPORTANT**: This implementation plan should remain high-level and readable. Any code samples, detailed algorithms, or extensive technical specifications must be placed in the appropriate `implementation-details/` file.

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  virt-operator  │────│ virt-controller │────│   virt-handler  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
   [NO CHANGES]         [NO CHANGES]              [NO CHANGES]
                                     │                       │
                                     ▼                       ▼
                            ┌─────────────────┐    ┌─────────────────┐
                            │   virt-api      │    │ virt-launcher   │
                            └─────────────────┘    └─────────────────┘
                                     │                       │
                                     ▼                       ▼
                            [NO CHANGES]         [L1VH HYPERVISOR DETECTION]
                                                    [DOMAIN XML CONFIG]
```

### Design Principles Alignment

#### KubeVirt Razor Compliance
- **Principle**: "If something is useful for Pods, we should not implement it only for VMs"
- **Application**: L1VH provides transparent hypervisor optimization with no VM-specific APIs - feature operates at infrastructure level

#### Native Workloads Compatibility
- **Requirement**: Feature must not grant permissions users don't already have
- **Implementation**: Zero API changes, zero new RBAC requirements, feature gate controls cluster-wide behavior

#### Choreography Pattern
- **Approach**: Components act independently based on observed state
- **Implementation**: virt-launcher automatically detects L1VH capability and selects optimal hypervisor without coordination

## Implementation Phases

### Phase 0: Environment Setup (Foundation)

**Objective**: Establish secure development environment and validate constitutional compliance

**Deliverables**:
- [ ] **Environment Setup**: Feature gate implementation, contract test framework
- [ ] **Constitutional Verification**: All constitutional gates passed, compliance validated

**Exit Criteria**:
- [ ] Feature gate can be enabled/disabled without errors
- [ ] Contract tests written and failing (TDD Red phase)
- [ ] All constitutional gates verified
- [ ] Development team can proceed with implementation

### Phase 1: Test-First Development (Validation)

**Objective**: Implement contract-driven testing and validation before any production code

**Deliverables**:
- [ ] **Contract Testing**: Complete contract test suite for transparent hypervisor selection
- [ ] **Security Testing**: Security validation and threat model implementation  
- [ ] **Integration Validation**: Integration test framework ready for real environment testing

**Exit Criteria**:
- [ ] All contract tests written and provide clear behavioral specification
- [ ] Security tests validate KubeVirt integration boundaries
- [ ] Integration tests ready for L1VH cluster validation
- [ ] Implementation can begin with clear test-driven requirements

### Phase 2: Core Implementation (Minimal Implementation)

**Objective**: Implement minimal, constitutional L1VH transparent hypervisor selection

**Deliverables**:
- [ ] **Transparent Detection**: L1VH device detection and hypervisor selection in converter
- [ ] **Domain Configuration**: Research and implement libvirt mshv domain requirements  
- [ ] **Graceful Fallback**: QEMU/KVM fallback when L1VH unavailable

**Exit Criteria**:
- [ ] VMs automatically use L1VH when feature gate enabled and `/dev/mshv` available
- [ ] VMs seamlessly fall back to QEMU/KVM when L1VH unavailable
- [ ] All contract tests pass (TDD Green phase)
- [ ] Zero impact on existing functionality when feature gate disabled

### Phase 3: Observability (Monitoring & Debugging)

**Objective**: Minimal observability for L1VH hypervisor selection decisions

**Deliverables**:
- [ ] **Metrics and Logging**: Hypervisor selection metrics and structured logging
- [ ] **Troubleshooting**: Debug capabilities and troubleshooting documentation

**Exit Criteria**:
- [ ] Hypervisor selection decisions visible in metrics and logs
- [ ] Debug endpoints provide L1VH status information
- [ ] Troubleshooting documentation covers common scenarios
- [ ] Operational teams can monitor and debug L1VH operations

### Phase 4: Release Preparation (Finalization)

**Objective**: Final validation and Alpha release preparation

**Deliverables**:
- [ ] **Release Readiness**: E2E validation, documentation, Alpha release preparation

**Exit Criteria**:
- [ ] All existing integration tests pass identically with L1VH vs QEMU
- [ ] Documentation explains transparent operation and setup requirements  
- [ ] Feature ready for Alpha release behind feature gate
- [ ] Constitutional compliance verified for release

---

## Executive Summary

This implementation plan details the technical approach for adding **transparent** Hyper-V L1VH (Level 1 Virtual Host) support to KubeVirt v1.5.0+. The feature enables existing VirtualMachine resources to automatically benefit from L1VH performance optimization when deployed on L1VH-capable clusters, **without requiring any API changes or user configuration**.

### Key Implementation Points

- **Zero API Changes**: Existing VirtualMachine resources work transparently with L1VH optimization
- **Transparent Operation**: VMs automatically use L1VH when feature gate is enabled on L1VH-capable clusters
- **Feature Gate Controlled**: `HyperVL1VH` alpha gate for safe cluster-wide rollout
- **Non-Breaking**: Maintains 100% backward compatibility with existing QEMU/KVM workloads
- **Simplified Architecture**: No complex scheduling, node labeling, or user decision-making required

### L1VH-Specific Implementation Focus

This implementation focuses specifically on L1VH hypervisor integration with minimal changes to existing KubeVirt components. The feature operates transparently, requiring only converter logic modifications and basic observability.

## Technology Stack

### Core Technologies
- **Programming Language**: Go 1.21+
- **Kubernetes API**: v1.31+ - Required for stable CRD features and resource management
- **libvirt**: lsg-pre-release - Needed for Hyper-V integration capabilities through mshv driver
- **QEMU**: lsg-pre-release - Required for hybrid hypervisor support

### KubeVirt Framework Integration
- **Feature Gates**: `HyperVL1VH` alpha feature gate in `pkg/virt-config/featuregate/`
- **CRD Extensions**: NO CRD MODIFICATIONS - transparent operation
- **Controller Patterns**: NO NEW CONTROLLERS - leverages existing converter patterns
- **API Patterns**: NO API CHANGES - transparent hypervisor selection

### External Dependencies
- **Host Dependencies**: `/dev/mshv` device, Microsoft mshv kernel driver, patched QEMU, L1VH-capable Azure VMs
- **Network Dependencies**: Standard Kubernetes networking, no additional requirements

## Component Implementation Strategy
- [ ] NO MODIFICATIONS - Feature operates transparently

**Integration Points**:
- Feature gate configuration: Standard KubeVirt feature gate patterns

### virt-controller
**Role**: NO CHANGES REQUIRED
**Changes Required**:
- [ ] NO MODIFICATIONS - Feature operates transparently

**Integration Points**:
- VMI lifecycle: Unchanged - operates identically with L1VH and QEMU/KVM

### virt-handler
**Role**: NO CHANGES REQUIRED
**Changes Required**:
- [ ] NO MODIFICATIONS - Feature operates transparently

**Integration Points**:
- Domain management: Unchanged - standard libvirt operations

### virt-launcher
**Role**: HYPERVISOR DETECTION AND SELECTION
**Changes Required**:
- [ ] **L1VH Device Detection**: Add `/dev/mshv` detection logic in converter
- [ ] **Domain XML Configuration**: Configure libvirt domain for mshv hypervisor
- [ ] **Metrics and Logging**: Track hypervisor selection decisions

**Integration Points**:
- Converter logic: Integrates with existing `Convert_v1_VirtualMachineInstance_To_api_Domain` function
- Feature gate: Uses existing feature gate framework for enable/disable control

### virt-api
**Role**: NO CHANGES REQUIRED
**Changes Required**:
- [ ] NO MODIFICATIONS - Feature operates transparently

**Integration Points**:
- API validation: NO CHANGES - existing VM specifications work identically

## API Implementation

### CRD Modifications

#### VirtualMachineInstance Spec
```yaml
# NO API CHANGES - L1VH operates transparently
spec:
  domain:
    # Standard KubeVirt VM specification - no changes needed
    # L1VH optimization happens automatically when feature gate is enabled
```

#### VirtualMachine Status
```yaml
# NO STATUS CHANGES - Hypervisor selection is transparent to users
# Observability through metrics and logs, not API status
```

### Validation Logic
- **Location**: NO CHANGES REQUIRED
- **Requirements**: NO NEW VALIDATION - existing validation applies
- **Error Messages**: NO NEW ERRORS - standard VM validation continues

### Defaulting Logic
- **Location**: NO CHANGES REQUIRED
- **Default Values**: NO NEW DEFAULTS - hypervisor selection is automatic
- **Conditions**: NO NEW CONDITIONS - feature gate controls behavior

## Controller Logic Implementation

### Reconciliation Loop Design
```go
// NO NEW CONTROLLERS - Feature integrates into existing converter logic
// Implementation details in pkg/virt-launcher/virtwrap/converter/converter.go

func Convert_v1_VirtualMachineInstance_To_api_Domain(vmi *v1.VirtualMachineInstance, domain *api.Domain, c *ConverterContext) error {
    // Existing conversion logic...
    
    // L1VH hypervisor detection (when feature gate enabled)
    if featuregate.DefaultFeatureGate.Enabled(featuregate.HyperVL1VH) {
        if hasL1VHSupport() {
            // Apply L1VH-specific domain configuration
        }
    }
    
    return nil
}
```

### State Machine
- **States**: NO NEW STATES - uses existing VM lifecycle states
- **Transitions**: NO NEW TRANSITIONS - hypervisor selection is transparent
- **Error Handling**: Standard libvirt error handling applies

### Event Handling
- **Watch Targets**: NO NEW WATCHES - uses existing VMI watches
- **Event Types**: NO NEW EVENTS - standard VMI lifecycle events
- **Event Processing**: NO CHANGES - existing event processing continues

## Testing Implementation Strategy

### Unit Testing Approach
- **Framework**: Ginkgo/Gomega (following KubeVirt patterns)
- **Coverage Target**: >80% for new converter logic
- **Mock Strategy**: Mock `/dev/mshv` device detection for unit tests

### Integration Testing Approach
- **Environment**: KubeVirt integration test framework
- **Test Scenarios**: 
  - Transparent hypervisor selection with feature gate enabled
  - QEMU/KVM fallback behavior when L1VH unavailable
  - Existing integration tests run identically on L1VH cluster
- **Data Setup**: Standard VM specifications (no L1VH-specific configuration)

### End-to-End Testing Approach
- **Framework**: KubeVirt E2E test suite
- **User Workflows**: 
  - kubectl apply → VM creation → VM deletion works transparently
  - Existing E2E tests pass identically with L1VH vs QEMU/KVM
- **Environment Requirements**: Azure L1VH-capable cluster with `/dev/mshv` devices

## File Organization Strategy

```
pkg/
├── virt-config/featuregate/
│   └── feature-gates.go (modified - add HyperVL1VH constant)
├── virt-launcher/virtwrap/converter/
│   └── converter.go (modified - add L1VH detection logic)
└── [NO OTHER NEW PACKAGES - follows simplicity principle]

tests/
├── hyperv_l1vh_test.go (minimal L1VH-specific integration test)
└── [LEVERAGE EXISTING TEST FILES]

docs/
├── hyperv-l1vh.md (user guide)
├── hyperv-l1vh-cluster-setup.md (cluster setup guide)
└── l1vh-domain-xml-requirements.md (technical documentation)
```

### Modified Files
- **Feature Gates**: `pkg/virt-config/featuregate/feature-gates.go`
- **Converter Logic**: `pkg/virt-launcher/virtwrap/converter/converter.go`
- **Build Configuration**: `BUILD.bazel` files for conditional compilation

## Deployment Strategy

### Feature Gate Configuration
```yaml
# Example KubeVirt configuration
apiVersion: kubevirt.io/v1
kind: KubeVirt
spec:
  configuration:
    developerConfiguration:
      featureGates:
        - HyperVL1VH
```

### Rollout Plan
1. **Development Clusters**: Enable feature gate in dev environment with mocked L1VH
2. **Staging Clusters**: Test on real Azure L1VH-capable clusters
3. **Production Clusters**: Alpha release with explicit feature gate enablement

### Rollback Plan
- **Trigger Conditions**: Any regressions in existing functionality, L1VH detection failures
- **Rollback Process**: Disable HyperVL1VH feature gate - automatic fallback to QEMU/KVM
- **Data Migration**: NO DATA MIGRATION NEEDED - VMs continue with existing hypervisor

## Monitoring and Observability

### Metrics Implementation
- **Prometheus Metrics**: 
  - `kubevirt_vmi_hypervisor_type_total{hypervisor="mshv|kvm"}`
  - `kubevirt_l1vh_fallback_total{reason="device_unavailable|feature_disabled"}`
- **Collection Points**: Converter logic during hypervisor selection
- **Alerting Rules**: Track L1VH device availability and fallback rates

### Logging Strategy
- **Log Levels**: INFO for hypervisor selection decisions, DEBUG for detailed detection logic
- **Log Format**: JSON-structured logging with correlation IDs (existing KubeVirt pattern)
- **Sensitive Data**: NO SENSITIVE DATA - device detection and hypervisor selection only

### Debugging Support
- **Debug Endpoints**: Integrate L1VH status into existing virt-launcher debug output
- **Troubleshooting Tools**: Document L1VH validation using existing `virtctl` and `kubectl` commands
- **Support Information**: L1VH device availability, feature gate status, hypervisor selection decisions

## L1VH Documentation

### Code Documentation
- **Package Documentation**: Update converter package docs to describe L1VH integration
- **API Documentation**: NO NEW APIs - existing documentation continues to apply
- **Inline Comments**: Document L1VH detection logic and domain configuration decisions

### User Documentation
- **User Guide Updates**: Create `docs/hyperv-l1vh.md` explaining transparent operation
- **API Reference**: NO CHANGES - existing VM specifications work identically
- **Examples**: Standard VM YAML examples with explanation of automatic L1VH optimization

### Developer Documentation
- **Architecture Documents**: Update converter architecture to include L1VH detection
- **Integration Guides**: Document L1VH cluster setup requirements
- **Troubleshooting Guides**: Debug procedures for L1VH device detection and fallback scenarios

## L1VH Risk Assessment

### Technical Risks
1. **libvirt mshv driver compatibility**: 
   - **Probability**: Low
   - **Impact**: High
   - **Mitigation**: Research phase validates libvirt mshv requirements before implementation

2. **L1VH device detection reliability**:
   - **Probability**: Medium
   - **Impact**: Medium
   - **Mitigation**: Robust fallback to QEMU/KVM, comprehensive testing of detection logic

### Dependency Risks
- **Upstream Changes**: Monitor libvirt and QEMU releases for mshv driver changes
- **Version Conflicts**: Document and validate specific version combinations
- **External Service Dependencies**: Azure L1VH VM availability for testing

### Integration Risks
- **Existing Feature Conflicts**: Leverage existing tests to validate no regressions
- **Performance Impact**: Monitor converter performance with L1VH detection logic
- **Security Vulnerabilities**: Trust libvirt security model for hypervisor isolation

## L1VH Success Criteria

### Functional Success
- [x] All user stories from feature specification are implementable (transparent operation)
- [ ] All acceptance criteria are met (VM lifecycle works identically)
- [ ] Feature works correctly when enabled via feature gate
- [ ] System works correctly when feature is disabled (graceful fallback)

### Technical Success
- [ ] Code follows KubeVirt coding standards and patterns
- [ ] All tests pass in CI/CD pipeline (contract, integration, E2E)
- [ ] Performance meets non-functional requirements (no converter overhead)
- [ ] Security review confirms no new attack surfaces

### Integration Success
- [ ] Feature integrates cleanly with existing KubeVirt components (converter only)
- [ ] No regressions in existing functionality (validated by existing tests)
- [ ] Backward compatibility is maintained (zero API changes)
- [ ] Feature follows Kubernetes API conventions (no API changes needed)

### Documentation Success
- [ ] User documentation explains transparent operation clearly
- [ ] Developer documentation enables easy contribution
- [ ] Troubleshooting information covers L1VH-specific scenarios
- [ ] Setup documentation provides clear cluster requirements

## L1VH Implementation Execution

**IMPORTANT**: Follow this order to ensure dependencies are satisfied:

### Phase 0: Infrastructure
1. Create feature gate implementation in `pkg/virt-config/featuregate/feature-gates.go`
2. Create contract test files in `pkg/virt-launcher/virtwrap/converter/converter_test.go`
3. Create Azure L1VH test environment setup
4. Update build configuration for conditional compilation

### Phase 1: Core Implementation
1. Research and document libvirt mshv requirements in `docs/l1vh-domain-xml-requirements.md`
2. Implement L1VH detection logic in `pkg/virt-launcher/virtwrap/converter/converter.go`
3. Implement domain XML configuration for mshv hypervisor
4. Add integration test for hypervisor selection in `tests/hyperv_l1vh_test.go`

### Phase 2: Validation and Documentation
1. Validate existing integration tests on L1VH cluster
2. Implement minimal metrics and logging for hypervisor selection
3. Create user documentation in `docs/hyperv-l1vh.md`
4. Create cluster setup guide in `docs/hyperv-l1vh-cluster-setup.md`

---

## Implementation Plan Execution

### Creating the L1VH Implementation

1. **Simplified Scope**: Implementation focuses on minimal, transparent hypervisor selection
2. **Test-First Approach**: Contract tests drive implementation (TDD Red → Green → Refactor)
3. **Existing Pattern Reuse**: Leverages established KubeVirt converter and feature gate patterns
4. **Zero API Changes**: Maintains backward compatibility with no new APIs or CRDs

### Quality Standards

- **Code Quality**: Follow KubeVirt's established patterns for converter logic and feature gates
- **Test Coverage**: Achieve >80% coverage for new L1VH detection logic
- **Documentation**: Document transparent operation and cluster setup requirements
- **Security**: Trust libvirt security model, no new attack surfaces introduced
- **Performance**: Zero overhead when feature disabled, minimal overhead when enabled
- **Zero User Impact**: Users submit identical VM specifications regardless of underlying hypervisor
- **No VM-Specific Configuration**: Follows existing Kubernetes patterns without new abstractions

## L1VH Converter Implementation

1. Feature Gate Registration
2. Transparent Hypervisor Detection
3. L1VH Domain Configuration

## L1VH Testing Strategy

### Test-First Development

All tests must be written and approved before implementation begins.

### L1VH Test Hierarchy

1. Contract Tests
2. Integration Tests
3. End-to-End Tests
4. Unit Tests

## L1VH Build Integration

### 1. Bazel Build Integration

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

## L1VH Observability

### Metrics Implementation
```go
// Hypervisor selection metrics
kubevirt_vmi_hypervisor_type_total{hypervisor="mshv"} 
kubevirt_vmi_hypervisor_type_total{hypervisor="qemu"}

// L1VH performance metrics  
kubevirt_l1vh_vm_startup_duration_seconds
kubevirt_l1vh_device_access_total{device="mshv"}
kubevirt_l1vh_optimization_enabled_total

// Resource utilization comparison
kubevirt_l1vh_memory_overhead_bytes
kubevirt_l1vh_cpu_overhead_percentage

// Error and fallback metrics
kubevirt_l1vh_fallback_to_kvm_total{reason="device_unavailable|feature_disabled"}
kubevirt_l1vh_errors_total{operation="domain_create|vm_start|device_access"}
```

#### 1.2 Performance Monitoring

**SLA Metrics** (Constitutional Article XIV):
```go
// VM lifecycle performance
kubevirt_vmi_creation_duration_seconds{hypervisor="mshv"}
kubevirt_vmi_start_duration_seconds{hypervisor="mshv"}  
kubevirt_vmi_migration_duration_seconds{hypervisor="mshv"}

// Resource efficiency metrics
kubevirt_l1vh_resource_utilization_ratio{resource="cpu|memory|network|storage"}
kubevirt_l1vh_performance_improvement_ratio{metric="startup|throughput|latency"}
```

### 2. Structured Logging Requirements

#### 2.1 L1VH Event Logging

**Required Log Events**:
```go
// Hypervisor selection events
log.WithFields(log.Fields{
    "vmi": vmi.Name,
    "hypervisor_selected": "mshv",
    "feature_gate_enabled": true,
    "mshv_available": true,
    "selection_reason": "automatic_optimization",
}).Info("L1VH hypervisor selected for VMI")

// Performance and optimization events
log.WithFields(log.Fields{
    "vmi": vmi.Name,
    "startup_time_ms": duration.Milliseconds(),
    "hypervisor": "mshv",
    "optimization_level": "l1vh_enabled",
}).Info("L1VH VM startup completed")

// Fallback events  
log.WithFields(log.Fields{
    "vmi": vmi.Name,
    "fallback_reason": "mshv_device_unavailable",
    "fallback_hypervisor": "qemu",
    "performance_impact": "nested_virtualization_penalty",
}).Warn("L1VH unavailable, falling back to QEMU/KVM")

// Security events
log.WithFields(log.Fields{
    "vmi": vmi.Name,
    "security_context": vmi.Status.SelinuxContext,
    "mshv_access_granted": true,
    "isolation_verified": true,
}).Info("L1VH VM security context validated")
```

#### 2.2 Error Handling and Remediation

**Error Logging with Remediation Guidance**:
```go
// L1VH initialization failures
log.WithFields(log.Fields{
    "error": err.Error(),
    "remediation": "verify_mshv_kernel_module_loaded",
    "documentation": "https://kubevirt.io/troubleshooting/l1vh",
}).Error("L1VH device initialization failed")

// Feature gate configuration errors
log.WithFields(log.Fields{
    "feature_gate": "HyperVL1VH",
    "cluster_ready": false,
    "remediation": "ensure_all_nodes_l1vh_capable",
    "validation_command": "kubectl get nodes -o yaml | grep l1vh-capable",
}).Error("L1VH feature gate enabled but cluster not ready")
```

### 3. Debugging and Troubleshooting

#### 3.1 Debug Endpoints

**L1VH Status Endpoints**:
```go
// Add to virt-launcher debug interface
func (s *DebugServer) handleL1VHStatus(w http.ResponseWriter, r *http.Request) {
    status := L1VHStatus{
        DeviceAvailable: s.hasL1VHDevice(),
        FeatureEnabled: s.featureGateEnabled("HyperVL1VH"),
        ActiveVMs: s.getL1VHVMCount(),
        LastError: s.getLastL1VHError(),
        Performance: s.getL1VHPerformanceMetrics(),
    }
    json.NewEncoder(w).Encode(status)
}
```

#### 3.2 CLI Debugging Tools

**L1VH Validation Commands**:
```bash
# Cluster L1VH readiness check
kubectl get nodes -o jsonpath='{.items[*].status.allocatable.devices\.kubevirt\.io/mshv}'

# VMI hypervisor type inspection
kubectl get vmi <name> -o jsonpath='{.status.hypervisor}'

# L1VH performance metrics query
curl http://virt-launcher:8080/debug/l1vh/status
```

### 4. Alerting and SLA Management

#### 4.1 Operational Alerts

**Critical Alerts**:
```yaml
# L1VH device unavailability
- alert: L1VHDeviceUnavailable
  expr: kubevirt_l1vh_fallback_to_qemu_total{reason="device_unavailable"} > 0
  for: 1m
  labels:
    severity: warning
  annotations:
    summary: "L1VH device unavailable on node"
    description: "L1VH feature enabled but /dev/mshv device unavailable"
    runbook: "https://kubevirt.io/runbooks/l1vh-device-unavailable"

# L1VH performance degradation  
- alert: L1VHPerformanceDegradation
  expr: kubevirt_l1vh_performance_improvement_ratio{metric="startup"} < 1.1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "L1VH not providing expected performance improvement"
    description: "L1VH performance benefit below 10% threshold"
```

#### 4.2 SLA Monitoring

**Performance SLAs**:
- **VM Startup Time**: L1VH VMs start ≤90% of QEMU/KVM time
- **Resource Overhead**: L1VH adds ≤5% memory overhead vs QEMU/KVM
- **Availability**: L1VH fallback success rate ≥99.9%
- **Error Rate**: L1VH-related errors ≤0.1% of VM operations

### 5. Performance Benchmarking

#### 5.1 Continuous Performance Monitoring

**Automated Benchmarking**:
```go
func BenchmarkL1VHPerformance(b *testing.B) {
    // Continuous benchmarking in CI/CD pipeline
    scenarios := []BenchmarkScenario{
        {Name: "VM_Startup", Hypervisor: "mshv"},
        {Name: "VM_Startup", Hypervisor: "qemu"},
        {Name: "Memory_Bandwidth", Hypervisor: "mshv"},
        {Name: "Memory_Bandwidth", Hypervisor: "qemu"},
        {Name: "Network_Throughput", Hypervisor: "mshv"},
        {Name: "Network_Throughput", Hypervisor: "qemu"},
    }
    
    for _, scenario := range scenarios {
        b.Run(scenario.Name+"_"+scenario.Hypervisor, func(b *testing.B) {
            // Benchmark implementation
        })
    }
}
```

#### 5.2 Performance Regression Detection

**Automated Performance Gates**:
- Performance regression >5% fails CI/CD pipeline
- Performance improvement <10% for L1VH triggers investigation
- Resource utilization increase >3% requires optimization review

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

## Success Criteria

### Constitutional Compliance ✅

This implementation meets all constitutional requirements from **KubeVirt Constitution v2.0**:

- **Article II (KubeVirt Razor)**: No VM-specific APIs, uses existing Kubernetes patterns
- **Article III (Feature Gate Discipline)**: Alpha-first with proper graduation criteria  
- **Article IV (Test-First Implementation)**: TDD mandated with specific test hierarchy
- **Article XI (Security by Design)**: Comprehensive threat modeling and security controls
- **Article XII (Observability)**: Full metrics, logging, and debugging capabilities
- **Article XIV (Performance Excellence)**: Performance monitoring and SLA management
- **Article VII (Simplicity)**: Minimal implementation using ≤2 Go packages
- **Article VIII (Anti-Abstraction)**: Direct libvirt/QEMU integration without wrapper layers

### Technical Success Criteria

1. **Transparent Operation**
   - [ ] VMs automatically use L1VH when feature gate enabled and `/dev/mshv` available
   - [ ] VMs seamlessly fall back to QEMU/KVM when L1VH unavailable  
   - [ ] Users remain unaware of underlying hypervisor selection
   - [ ] Standard kubectl commands work identically across hypervisors

2. **Performance Objectives** (Constitutional Article XIV)
   - [ ] L1VH eliminates nested virtualization performance penalty (≥10% improvement)
   - [ ] Hardware passthrough enables direct GPU/storage/network access
   - [ ] VM startup times ≤90% of QEMU/KVM baseline
   - [ ] Memory usage overhead ≤5% compared to QEMU/KVM baseline
   - [ ] Performance regression detection prevents degradation

3. **Feature Integration**
   - [ ] All existing KubeVirt features work transparently with L1VH
   - [ ] Live migration functions correctly between L1VH-capable nodes
   - [ ] Storage and networking integrations work without modification
   - [ ] Monitoring and observability work identically

### Security Success Criteria (Constitutional Article XI)

1. **Security by Design**
   - [ ] **Threat Model Complete**: All attack vectors identified and mitigated
   - [ ] **Security Review Passed**: Independent security team approval
   - [ ] **Vulnerability Assessment**: No critical or high-severity findings
   - [ ] **Penetration Testing**: Red team assessment completed successfully

2. **Runtime Security**
   - [ ] **Privilege Non-Escalation**: L1VH grants no additional capabilities
   - [ ] **Isolation Maintained**: VM-to-VM and VM-to-host boundaries verified
   - [ ] **Input Validation**: All L1VH configurations properly sanitized
   - [ ] **Security Monitoring**: Anomaly detection and incident response ready

### Observability Success Criteria (Constitutional Article XII)

1. **Comprehensive Monitoring**
   - [ ] **Metrics Complete**: Prometheus metrics for all L1VH operations
   - [ ] **Structured Logging**: JSON logs with correlation IDs
   - [ ] **Debug Capabilities**: CLI tools and debug endpoints functional
   - [ ] **Alerting Configured**: Actionable alerts for operational issues

2. **Performance Monitoring**
   - [ ] **SLA Tracking**: Real-time performance against defined SLAs
   - [ ] **Regression Detection**: Automated performance change detection
   - [ ] **Capacity Planning**: Resource utilization trends and forecasting
   - [ ] **Cost Optimization**: Resource efficiency recommendations

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

1. **Test Coverage** (Constitutional Article IV)
   - [ ] **Contract Tests**: API behavior and integration contracts validated
   - [ ] **Integration Tests**: Component interactions in realistic environments
   - [ ] **End-to-End Tests**: Complete user workflows verified  
   - [ ] **Unit Tests**: Internal logic and edge cases covered
   - [ ] **Security Tests**: Threat scenarios and boundaries validated
   - [ ] **Performance Tests**: Scale and performance characteristics proven

2. **Production Readiness**
   - [ ] **Security Review**: Comprehensive security analysis completed
   - [ ] **Documentation Complete**: User guides, runbooks, troubleshooting
   - [ ] **Monitoring Integration**: L1VH-specific metrics and alerting
   - [ ] **Support Procedures**: L1VH incident response and troubleshooting

### Constitutional Gate Validation

**Pre-Implementation Constitutional Gates** (MUST PASS):

#### Simplicity Gate (Article VII)
- [x] Using ≤3 Go packages for initial implementation ✅
- [x] No future-proofing or premature optimization ✅
- [x] Complexity is justified and documented ✅

#### Anti-Abstraction Gate (Article VIII)
- [x] Using existing KubeVirt/Kubernetes patterns directly ✅
- [x] Not creating unnecessary wrapper layers ✅
- [x] Single model representation per domain concept ✅

#### Integration-First Gate (Article IX)
- [x] Integration tests planned before unit tests ✅
- [x] Real environment testing prioritized ✅
- [x] Contracts defined for all component interactions ✅

#### KubeVirt Razor Gate (Article II)
- [x] Feature follows "useful for Pods" principle ✅
- [x] No privileged capabilities beyond Kubernetes ✅
- [x] Choreography pattern respected ✅

#### Security Gate (Article XI)
- [ ] **Threat modeling completed and approved**
- [ ] **Security review scheduled and resourced**
- [ ] **Security testing strategy defined**
- [ ] **Incident response procedures established**

#### Observability Gate (Article XII)
- [ ] **Metrics strategy defined and approved**
- [ ] **Logging standards compliance verified**
- [ ] **Debug capabilities designed**
- [ ] **Alerting and SLA definitions complete**

**Constitutional Compliance**: All gates must pass before implementation begins per KubeVirt Constitution v2.0.

---
## Open Questions
1. **Libvirt and QEMU Version Dependencies**: How can we retrieve the minimum required versions of libvirt and QEMU for L1VH support?

## Risk Assessment and Mitigation

### Constitutional Requirement (Article III)

Per KubeVirt Constitution v2.0 Article III, **comprehensive risk analysis is mandatory** for all new features.

### KubeVirt Implementation Risks

#### 1. Constitutional Compliance Risk
- **Risk**: Implementation becomes complex, violating constitutional simplicity
- **Mitigation**: Strict adherence to ≤2 Go packages, regular constitutional review
- **KubeVirt Scope**: Design and implementation complexity management

#### 2. Configuration Validation Risk  
- **Risk**: Malformed L1VH configurations bypass validation and cause failures
- **Mitigation**: Comprehensive input validation, error handling, fuzzing tests
- **KubeVirt Scope**: VMI specification validation and domain conversion security

#### 3. Feature Gate Logic Risk
- **Risk**: Hypervisor selection logic introduces unexpected behaviors
- **Mitigation**: Clear logging, comprehensive testing, predictable fallback rules
- **KubeVirt Scope**: Feature gate implementation and validation logic

#### 4. Resource Management Risk
- **Risk**: L1VH VMs consume resources differently than QEMU VMs
- **Mitigation**: Identical resource limits enforcement, monitoring, testing
- **KubeVirt Scope**: Resource quota and limit enforcement consistency

### Operational Risks (KubeVirt Responsibility)

#### 1. Documentation and Support Risk
- **Risk**: Insufficient L1VH documentation leads to misconfigurations and support burden
- **Mitigation**: Comprehensive documentation, clear limitation statements, troubleshooting guides
- **KubeVirt Scope**: User education and operational guidance

#### 2. Cluster Configuration Risk
- **Risk**: L1VH assumption fails if some nodes lack capability
- **Mitigation**: Clear documentation of L1VH cluster requirements, validation tools
- **KubeVirt Scope**: Deployment guidance and validation tooling

#### 3. Performance Monitoring Risk
- **Risk**: L1VH performance impact on overall cluster performance
- **Mitigation**: Isolated code paths, performance monitoring, feature gate isolation
- **KubeVirt Scope**: Performance measurement and optimization

### Trust Dependencies (Outside KubeVirt Control)

#### 1. Hyper-V Hypervisor Security
- **Trust Assumption**: Microsoft Hyper-V provides enterprise-grade VM isolation
- **Validation**: Microsoft security certifications, industry adoption, security research
- **KubeVirt Position**: **Explicitly accepts** Hyper-V security model as foundational
- **Monitoring**: Industry security advisories, Microsoft security updates

#### 2. mshv Kernel Driver Security  
- **Trust Assumption**: Linux `/dev/mshv` driver follows kernel security standards
- **Validation**: Linux kernel community review, Microsoft engineering practices
- **KubeVirt Position**: **Explicitly accepts** kernel driver security is kernel responsibility
- **Monitoring**: Kernel security advisories, driver update notifications

#### 3. libvirt mshv Backend Security
- **Trust Assumption**: libvirt mshv driver maintains proven libvirt security model
- **Validation**: libvirt community review, established virtualization patterns
- **KubeVirt Position**: **Explicitly accepts** libvirt security model across all hypervisors
- **Monitoring**: libvirt security advisories, community security discussions

### Risk Acceptance Framework

#### Explicit Risk Acceptances Required:

1. **VM Isolation Security**: KubeVirt **accepts** that VM-to-VM and VM-to-host isolation is **Hyper-V's responsibility**

2. **Hypervisor Vulnerabilities**: KubeVirt **accepts** that hypervisor-level security issues are **Microsoft's responsibility** to address

3. **Kernel Driver Security**: KubeVirt **accepts** that `/dev/mshv` device security is **Linux kernel community's responsibility**

4. **Integration Security Only**: KubeVirt **takes responsibility** only for secure integration within existing KubeVirt security boundaries

#### Risk Mitigation Focus (KubeVirt Scope):

**Configuration Security**:
```go
func validateL1VHConfiguration(vmi *v1.VirtualMachineInstance) error {
    // Validate L1VH configurations don't bypass KubeVirt security policies
    // Ensure resource limits identical to QEMU VMs
    // Verify no additional privileges required
    return nil
}
```

**Access Control Parity**:
```go
func ensureSecurityParity(vmi *v1.VirtualMachineInstance) error {
    // Verify L1VH VMs have identical RBAC requirements to QEMU VMs
    // Validate security contexts match QEMU VM patterns
    // Ensure no capability escalation
    return nil
}
```

**Audit and Monitoring**:
```go
func auditL1VHOperation(vmi *v1.VirtualMachineInstance, operation string) {
    // Log L1VH operations for security monitoring
    // Track resource usage patterns
    // Monitor for configuration anomalies
}
```

### Emergency Procedures

#### Immediate Risk Response:
```yaml
Critical_Security_Advisory:
  trigger: CVE affecting Hyper-V, mshv, or libvirt
  response: Evaluate impact on KubeVirt integration
  action: Disable L1VH feature gate if integration affected
  communication: Security advisory to KubeVirt users

Configuration_Validation_Bypass:
  trigger: L1VH configuration validation failure
  response: Investigate root cause, patch validation logic
  action: Enhanced input validation, regression testing
  communication: Security notice and patching guidance
```

#### Risk Monitoring:
- **Monthly**: Review security advisories for all L1VH dependencies
- **Quarterly**: Assess KubeVirt integration security posture  
- **Annually**: Comprehensive trust model validation and documentation update

### Required Approvals

**Stakeholder Sign-offs Required**:
- [ ] **Security Team**: Approves trust model and integration security scope
- [ ] **Operations Team**: Accepts dependency on external security models
- [ ] **Product Team**: Approves risk/benefit balance and user communication strategy

---

**Next Steps:**
1. **Constitutional Validation**: Verify all constitutional gates are met
2. **Test-First Development**: Write and approve all tests before implementation
3. **Minimal Implementation**: Implement only the converter changes required
4. **Performance Validation**: Establish benchmarking framework for L1VH benefits

---

*This implementation plan strictly follows the KubeVirt Constitution v1.0 and implements transparent Hyper-V L1VH support that eliminates nested virtualization performance penalties while maintaining 100% backward compatibility with existing KubeVirt workloads.*
