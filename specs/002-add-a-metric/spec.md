# Feature Specification: VMI Hypervisor Tracking Metric

## Metadata

- **Feature ID**: 002
- **Feature Name**: VMI Hypervisor Tracking Metric
- **Author(s)**: AI Assistant
- **Created**: October 13, 2025
- **Status**: Draft
- **KubeVirt Version**: [TARGET_VERSION]
- **Feature Gate**: N/A (metrics addition)

**Feature Branch**: `002-add-a-metric`  
**Input**: User description: "Add a metric to KubeVirt that is collected for each VMI. the metric should track which hypervisor is being used for a given VMI."

## Execution Flow (main)

```text
1. Parse user description from Input
   → Feature: VMI hypervisor tracking metric
2. Extract key concepts from description
   → Actors: cluster operators, monitoring systems
   → Actions: collect, expose, track hypervisor type
   → Data: hypervisor type per VMI
   → Constraints: per-VMI granularity
3. For each unclear aspect:
   → [NEEDS CLARIFICATION: metric naming convention]
   → [NEEDS CLARIFICATION: hypervisor detection mechanism]
4. Fill User Scenarios & Testing section
   → Primary: operators monitoring hypervisor usage
5. Generate Functional Requirements
   → Each requirement must be testable
6. Identify Key Entities: VMI, hypervisor type, metric
7. Run Review Checklist
   → Spec has some uncertainties marked
8. Return: SUCCESS (spec ready for planning)
```

## Executive Summary

**One-sentence description**: Add a Prometheus metric that tracks which hypervisor (QEMU/KVM, software emulation, etc.) is being used for each VirtualMachineInstance in KubeVirt.

**Business justification**: Enables cluster operators to monitor hypervisor distribution across their VM workloads, supporting performance analysis, capacity planning, and troubleshooting of virtualization infrastructure.

**User impact**: Cluster operators and monitoring teams gain visibility into hypervisor usage patterns, enabling better resource optimization and performance tuning decisions.

## Problem Statement

### Current State

- KubeVirt currently lacks visibility into which hypervisor backend is being used for individual VirtualMachineInstances
- Operators cannot easily monitor or analyze the distribution of hypervisor types across their VM workloads
- Performance troubleshooting and capacity planning lack hypervisor-specific context
- No standardized way to track when VMs are running with hardware acceleration vs software emulation

### Desired State

- A Prometheus metric exposed by KubeVirt components that identifies the hypervisor type for each VMI
- Operators can query hypervisor usage patterns across namespaces, nodes, and VM types
- Monitoring dashboards can display hypervisor distribution and correlate with performance metrics
- Automated alerting can detect when VMs unexpectedly fall back to software emulation

### Success Criteria

- **Functional**: Metric accurately reflects the actual hypervisor being used by each running VMI
- **Non-functional**: Metric collection has minimal performance overhead and updates promptly when hypervisor state changes
- **User Experience**: Metric follows Prometheus best practices and integrates seamlessly with existing KubeVirt monitoring

---

## ⚡ Quick Guidelines

- ✅ Focus on WHAT users need and WHY (business value for KubeVirt)
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for KubeVirt stakeholders and business users
- 🎯 Consider KubeVirt's architectural principles and patterns

### Section Requirements

- **Mandatory sections**: Must be completed for every KubeVirt feature
- **Optional sections**: Include only when relevant to the feature
- When a section doesn't apply, remove it entirely (don't leave as "N/A")

### For AI Generation

When creating this spec from a user prompt:

1. **Mark all ambiguities**: Use [NEEDS CLARIFICATION: specific question] for any assumption you'd need to make
2. **Don't guess**: If the prompt doesn't specify something, mark it clearly
3. **Think like a tester**: Every vague requirement should fail the "testable and unambiguous" checklist item
4. **KubeVirt-specific underspecified areas**:
   - User types (cluster operators, VM users, platform developers)
   - Feature gate behavior and lifecycle
   - Component integration (virt-controller, virt-handler, virt-launcher, virt-api)
   - VM lifecycle impact and compatibility
   - Hardware requirements and host environment needs
   - Performance targets and resource overhead
   - Security/compliance considerations for virtualization
   - Kubernetes integration patterns and CRD changes

---

## User Stories *(mandatory)*

### Primary Users

- **Cluster Operators**: Platform administrators managing KubeVirt infrastructure and monitoring VM performance
- **Platform Developers**: Teams extending KubeVirt functionality and analyzing virtualization performance

### User Stories

**Story 1**: As a cluster operator, I want to see which hypervisor is being used for each VMI, so that I can identify performance issues and optimize resource allocation.

- **Acceptance Criteria**:
  - [ ] Metric shows accurate hypervisor type (e.g., "kvm", "qemu", "tcg") for each running VMI
  - [ ] Metric includes VMI name, namespace, and node labels for filtering and aggregation
  - [ ] Metric updates within 30 seconds when hypervisor state changes

**Story 2**: As a cluster operator, I want to monitor hypervisor distribution across my cluster, so that I can ensure optimal performance and identify nodes with hardware acceleration issues.

- **Acceptance Criteria**:
  - [ ] Can query total count of VMIs by hypervisor type across cluster
  - [ ] Can filter metrics by namespace, node, or other VMI characteristics
  - [ ] Can create alerts when VMIs unexpectedly use software emulation

**Story 3**: As a platform developer, I want to correlate hypervisor type with performance metrics, so that I can analyze the performance impact of different virtualization backends.

- **Acceptance Criteria**:
  - [ ] Metric can be joined with existing VMI performance metrics
  - [ ] Historical data is preserved for trend analysis
  - [ ] Metric follows standard KubeVirt labeling conventions

### Edge Cases & Error Scenarios

- What happens when hypervisor type cannot be determined? (metric should include "unknown" state)
- How does the metric behave during VMI transitions (starting, stopping)? (should reflect current state)
- How are failed or pending VMIs handled? (metric should only exist for running VMIs)

## Requirements *(mandatory)*

### Functional Requirements

1. **REQ-F-001**: KubeVirt MUST expose a Prometheus metric that indicates the hypervisor type for each running VirtualMachineInstance
2. **REQ-F-002**: Metric MUST include VMI identifying labels (name, namespace, node)  
3. **REQ-F-003**: Metric MUST distinguish between major hypervisor types [NEEDS CLARIFICATION: specific hypervisor types to detect - KVM, QEMU-TCG, others?]
4. **REQ-F-004**: Metric MUST only exist for VMIs in Running phase
5. **REQ-F-005**: Metric MUST update when hypervisor state changes during VMI lifecycle

### Non-Functional Requirements

1. **REQ-NF-001**: **Performance**: Metric collection MUST NOT add more than 1% CPU overhead to virt-handler
2. **REQ-NF-002**: **Reliability**: Metric MUST accurately reflect actual hypervisor in use, not just configuration
3. **REQ-NF-003**: **Compatibility**: Implementation MUST NOT interfere with existing VMI monitoring metrics
4. **REQ-NF-004**: **Scalability**: Metric collection MUST scale to clusters with 1000+ concurrent VMIs

### Kubernetes Integration Requirements

1. **REQ-K8S-001**: MUST follow Prometheus metric naming conventions for Kubernetes
2. **REQ-K8S-002**: MUST integrate with existing KubeVirt metrics collection without breaking changes
3. **REQ-K8S-003**: MUST support standard Prometheus label-based filtering and aggregation
4. **REQ-K8S-004**: MUST be compatible with KubeVirt's existing monitoring architecture

### Key Entities

- **VMI Hypervisor Metric**: Represents the current hypervisor type for a running VMI, with labels for VMI identification and hypervisor classification
- **Hypervisor Type**: Classification of virtualization backend (e.g., hardware-accelerated KVM, software-emulated QEMU)

## Dependencies and Prerequisites

### KubeVirt Component Impact

- **virt-handler**: [NEEDS CLARIFICATION: likely source of hypervisor detection and metric emission]
- **virt-launcher**: May need to provide hypervisor information to virt-handler
- **Existing metrics infrastructure**: Must integrate with current Prometheus metrics collection

### Hypervisor Detection Requirements

- **Detection Method**: [NEEDS CLARIFICATION: mechanism to detect actual hypervisor type - libvirt API, QEMU monitor, /proc inspection?]
- **Update Frequency**: [NEEDS CLARIFICATION: how often to check/update hypervisor type]

## Security Considerations

### Security Impact

- Minimal security impact as metric exposes only hypervisor type information
- No sensitive VM data or configuration details exposed
- Standard Prometheus metrics endpoint security applies

---

## Review & Acceptance Checklist

### Content Quality

- [x] No implementation details (languages, frameworks, APIs, code structure)
- [x] Focused on user value and KubeVirt business needs
- [x] Written for KubeVirt stakeholders and business users
- [x] All mandatory sections completed
- [x] KubeVirt architectural principles respected

### Requirement Completeness

- [ ] No [NEEDS CLARIFICATION] markers remain (3 clarifications needed)
- [x] Requirements are testable and unambiguous  
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified
- [x] Component impact assessed
- [x] Kubernetes integration patterns defined

### KubeVirt-Specific Validation

- [x] User stories cover cluster operators and platform developers
- [x] VM lifecycle impact considered
- [x] Compatibility with existing functionality addressed
- [x] Performance and resource overhead implications identified
- [x] Security implications assessed

---

## Execution Status

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked (3 clarifications needed)
- [x] User scenarios defined
- [x] Requirements generated  
- [x] Entities identified
- [ ] Review checklist passed (pending clarifications)

---
