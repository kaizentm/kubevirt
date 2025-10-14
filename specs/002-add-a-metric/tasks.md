# Tasks: VMI Hypervisor Tracking Metric

**Feature ID**: 002  
**Feature Name**: VMI Hypervisor Tracking Metric  
**Updated**: October 13, 2025  
**Status**: Ready for Implementation  
**Implementation Plan**: [plan.md](./plan.md)  
**Feature Specification**: [quickstart.md](./quickstart.md)  

**Input**: Design documents from `/specs/002-add-a-metric/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)

```text
1. Load plan.md from feature directory
   → Extract: Static metrics pattern, virt-handler integration, operator-observability-toolkit
   → Components: pkg/monitoring/metrics/virt-handler/ infrastructure
2. Load design documents:
   → data-model.md: InfoVec metric structure, hypervisor type enum → model tasks
   → contracts/: virt-handler integration contracts → integration test tasks  
   → research.md: libvirt domain XML parsing, event-driven updates → setup tasks
3. Verify Constitutional Compliance:
   → ✅ KubeVirt Razor adherence: Pod-VM parity for observability
   → ✅ No feature gate required: metrics addition only
   → ✅ Security boundaries: read-only metric, no new privileges
   → ✅ Integration-first testing: VMI lifecycle and metric accuracy tests
4. Generate tasks by category:
   → Foundation: hypervisor detection logic, metric registration
   → Core Implementation: VMI event handlers, libvirt integration, metric updates
   → Integration Testing: VMI lifecycle validation, metric accuracy verification
   → Release Preparation: documentation, performance validation
5. Apply KubeVirt-specific patterns:
   → Static metrics pattern (like versionInfo, machineTypeMetrics)
   → VMI informer event handlers for lifecycle management
   → Integration tests before unit tests
   → Component separation maintained (virt-handler only)
6. Number tasks sequentially (T001, T002...)
7. Generate dependency graph with constitutional gates
8. Validate task completeness:
   → ✅ Constitutional compliance verified
   → ✅ No feature gate needed (metrics only)
   → ✅ Integration tests cover real VMI environments
   → ✅ No additional security review required
9. Return: SUCCESS (tasks ready for implementation)
```

---

## Constitutional Compliance ✅

Based on the KubeVirt Constitution, verified:

- [x] **KubeVirt Razor**: Provides Pod-VM parity for observability (metrics for VMs like containers)
- [x] **Feature Gate Requirements**: N/A - metrics addition does not require feature gate
- [x] **Security-First**: No new privileges, read-only metrics endpoint, no sensitive data
- [x] **Integration-First Testing**: VMI lifecycle testing prioritized, metric accuracy validation
- [x] **Component Architecture**: Extends virt-handler only, maintains choreography pattern
- [x] **API Backward Compatibility**: No API changes, metrics-only addition
- [x] **Reproducible Build System**: Uses established Make/Bazel workflows

### Compliance Gate

CONSTITUTIONAL COMPLIANCE VERIFIED ✅ - PROCEED WITH IMPLEMENTATION

---

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in KubeVirt structure
- Follow KubeVirt component separation patterns

## KubeVirt Path Conventions

- **Metrics**: `pkg/monitoring/metrics/virt-handler/`
- **Tests**: `tests/` (functional), `pkg/monitoring/metrics/virt-handler/` (unit)
- **Integration**: Existing virt-handler VMI informer patterns

## Phase 1: Foundation

### Task 1.1: Hypervisor Detection Logic

**Definition of Done**: Hypervisor type detection from libvirt domain XML functional

**Deliverables**:

- [x] T001 [P] **Hypervisor Type Enumeration**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go`
  - Task: Define HypervisorType constants (kvm, qemu-tcg, unknown)
  - Pattern: Follow existing metric type definitions

- [x] T002 [P] **libvirt Domain XML Parsing**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go`
  - Task: Implement `detectHypervisorType()` function
  - Integration: Parse `<domain type="...">` attribute from libvirt XML

- [x] T003 [P] **Detection Logic Unit Tests**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics_test.go`
  - Test: Hypervisor type detection for KVM, QEMU-TCG, unknown cases
  - Test: Error handling for invalid/missing domain XML

**Acceptance Criteria**: Hypervisor detection works reliably for all supported types

### Task 1.2: Metric Registration Infrastructure

**Definition of Done**: InfoVec metric properly registered in virt-handler

**Deliverables**:

- [x] T004 **Metric Definition and Registration**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go`
  - Task: Define `kubevirt_vmi_hypervisor_info` InfoVec metric using operatormetrics
  - Labels: namespace, name, node, hypervisor_type

- [x] T005 **Integration with SetupMetrics()**
  - File: `pkg/monitoring/metrics/virt-handler/metrics.go`
  - Task: Add hypervisor metric registration to existing `SetupMetrics()` function
  - Integration: Import and initialize hypervisor metrics package

**Acceptance Criteria**: Metric registered and available on /metrics endpoint

## Phase 2: Integration-First Tests ⚠️ MUST COMPLETE BEFORE PHASE 3

CRITICAL: Integration and E2E tests MUST be written first and MUST FAIL before implementation

- [x] T006 [P] **Integration Test - VMI Lifecycle Events**
  - File: `tests/hypervisor_metric_test.go`
  - Test: Metric appears when VMI enters Running phase
  - Test: Metric disappears when VMI is deleted

- [x] T007 [P] **Integration Test - Hypervisor Type Accuracy**
  - File: `tests/hypervisor_metric_test.go`  
  - Test: KVM-enabled VMIs show hypervisor_type="kvm"
  - Test: Software emulation VMIs show hypervisor_type="qemu-tcg"

- [x] T008 [P] **Integration Test - Multi-Node Scenarios**
  - File: `tests/hypervisor_metric_test.go`
  - Test: VMI migration updates node label correctly
  - Test: Multiple VMIs on same node have separate metrics

- [x] T009 [P] **Integration Test - Error Handling**
  - File: `tests/hypervisor_metric_test.go`
  - Test: VMI without libvirt domain shows hypervisor_type="unknown"
  - Test: libvirt connection failures handled gracefully

## Phase 3: Core Implementation (ONLY after tests are failing)

### Task 3.1: VMI Event Handler Implementation

**Definition of Done**: VMI informer events properly trigger metric updates

**Deliverables**:

- [ ] T010 **VMI Add Event Handler**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go`
  - Task: Implement `handleVMIAdd()` for VMI creation/running events
  - Integration: Detect hypervisor type and create metric

- [ ] T011 **VMI Update Event Handler**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go`
  - Task: Implement `handleVMIUpdate()` for VMI state changes
  - Integration: Update metric on phase transitions or migration

- [ ] T012 **VMI Delete Event Handler**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go`
  - Task: Implement `handleVMIDelete()` for VMI cleanup
  - Integration: Remove metric when VMI is deleted

### Task 3.2: Metric Lifecycle Management

**Definition of Done**: Metric values accurately reflect VMI hypervisor state

**Deliverables**:

- [ ] T013 **Metric Update Functions**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics.go`
  - Task: Implement `updateHypervisorMetric()` and `removeHypervisorMetric()`
  - Pattern: Follow existing static metric patterns (versionInfo style)

- [ ] T014 **Event Handler Registration**
  - File: `pkg/monitoring/metrics/virt-handler/metrics.go`
  - Task: Register VMI informer event handlers in `SetupMetrics()`
  - Integration: Connect handlers to existing VMI informer

## Phase 4: Release Preparation

### Task 4.1: Unit Testing

**Definition of Done**: Comprehensive unit test coverage for all components

**Deliverables**:

- [ ] T015 [P] **Event Handler Unit Tests**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics_test.go`
  - Test: VMI event handlers with mock VMI objects
  - Test: Metric lifecycle management functions

- [ ] T016 [P] **libvirt Integration Unit Tests**
  - File: `pkg/monitoring/metrics/virt-handler/hypervisor_metrics_test.go`
  - Test: Domain XML parsing with mock libvirt responses
  - Test: Error handling for libvirt connection failures

### Task 4.2: Documentation and Performance

**Definition of Done**: Complete documentation and performance validation

**Deliverables**:

- [ ] T017 [P] **Metric Documentation**
  - File: `docs/metrics.md` (extend existing)
  - Content: kubevirt_vmi_hypervisor_info metric description and usage
  - Content: Integration with existing KubeVirt monitoring

- [ ] T018 [P] **Performance Validation**
  - File: `tests/hypervisor_metric_test.go`
  - Test: Performance impact with 100+ concurrent VMIs
  - Test: Memory usage remains within bounds (<1% overhead)

**Acceptance Criteria**: Documentation complete, performance validated

## Dependencies

- Constitutional compliance verified ✅ (no gates required)
- Hypervisor detection (T001-T003) before metric registration (T004-T005)
- Integration tests (T006-T009) before implementation (T010-T014)
- Event handlers (T010-T012) before metric lifecycle (T013-T014)
- Core implementation before release preparation (T015-T018)

## Critical Research Questions

### High Priority ⚠️

1. **VMI Event Timing**: When exactly should hypervisor detection occur during VMI lifecycle?
2. **libvirt Connection Management**: How to handle libvirt connection failures gracefully?
3. **Metric Cardinality**: Impact of per-VMI metrics on Prometheus performance?

### Medium Priority

1. **Node Assignment**: How to handle VMI migration between nodes?
2. **Error Recovery**: What fallback behavior when hypervisor detection fails?
3. **Update Frequency**: Should metric be refreshed periodically or only on events?

## Parallel Example

```text
# Launch integration tests together:
Task: "Integration test VMI lifecycle events in tests/hypervisor_metric_test.go"
Task: "Integration test hypervisor type accuracy in tests/hypervisor_metric_test.go"
Task: "Integration test multi-node scenarios in tests/hypervisor_metric_test.go"
Task: "Integration test error handling in tests/hypervisor_metric_test.go"

# Launch unit tests together:
Task: "Event handler unit tests in pkg/monitoring/metrics/virt-handler/hypervisor_metrics_test.go"
Task: "libvirt integration unit tests in pkg/monitoring/metrics/virt-handler/hypervisor_metrics_test.go"

# Launch documentation tasks together:
Task: "Metric documentation in docs/metrics.md"
Task: "Performance validation in tests/hypervisor_metric_test.go"
```

## KubeVirt-Specific Implementation Notes

### Static Metrics Pattern
This feature follows KubeVirt's static metrics pattern (like `versionInfo`, `machineTypeMetrics`) rather than dynamic collectors (like `domainstats`). The hypervisor type is a static property that doesn't change during VMI lifetime.

### VMI Informer Integration
Uses existing VMI informer infrastructure in virt-handler for event-driven updates. This is more efficient than polling and aligns with KubeVirt's choreography architecture.

### No Feature Gate Required
Metrics additions in KubeVirt don't require feature gates unless they change behavior. This is purely observability enhancement with no functional impact.

### libvirt Integration Pattern
Follows established patterns in virt-handler for libvirt domain management, including error handling and connection management.