
# Implementation Plan: VMI Hypervisor Tracking Metric

**Branch**: `002-add-a-metric` | **Date**: October 13, 2025 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/home/ubuntu/git/ARO/kubevirt/specs/002-add-a-metric/spec.md`

## Execution Flow (/plan command scope)

```text
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect KubeVirt Component Architecture (virtualization, controller, operator patterns)
   → Set Component Integration Strategy based on choreography patterns
3. Fill the Constitution Check section based on the content of the constitution document
   → Validate KubeVirt Razor compliance (Pod-VM parity principle)
   → Check Feature Gate implementation requirements
   → Verify Integration-First testing approach
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If constitutional gates fail: ERROR "Must address constitutional compliance"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md (KubeVirt patterns & dependencies)
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
   → Focus on libvirt/QEMU/Kubernetes integration points
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, component-integration.md, .github/copilot-instructions.md
   → Generate CRD API contracts following Kubernetes conventions
   → Design controller choreography patterns
   → Plan feature gate implementation strategy
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Ensure KubeVirt architectural compliance
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe KubeVirt-specific task generation approach (DO NOT create tasks.md)
   → Include component-specific implementation phases
   → Plan constitutional gate validation throughout implementation
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 8. Implementation phases are executed by other commands:

- Phase 2: /tasks command creates tasks.md with KubeVirt patterns
- Phase 3+: Implementation execution following constitutional principles

## Summary

Add a Prometheus info metric `kubevirt_vmi_hypervisor_info` that tracks which hypervisor (KVM, QEMU-TCG, unknown) is being used for each running VirtualMachineInstance. The implementation extends KubeVirt's existing `pkg/monitoring/metrics/virt-handler/` infrastructure by adding a new hypervisor collector following the established domainstats pattern, reusing VMI informers and libvirt connection management. This enables cluster operators to monitor hypervisor distribution for performance analysis and capacity planning.

## Technical Context

**Language/Version**: Go (KubeVirt standard - current main branch)  
**KubeVirt Framework**: Current main branch (no specific version dependency)  
**Kubernetes API**: Standard Kubernetes API (no new API changes required)  
**Primary Dependencies**: operator-observability-toolkit (existing), libvirt (existing), shared VMI informer (existing)  
**Host Dependencies**: libvirt daemon (already required by KubeVirt), no additional host requirements  
**Component Architecture**: pkg/monitoring/metrics/virt-handler/ (new hypervisor package following domainstats pattern)  
**Testing Framework**: Ginkgo/Gomega (KubeVirt standard)  
**Target Platform**: Linux Kubernetes clusters  
**Feature Gate Strategy**: N/A - metrics addition does not require feature gate  
**Performance Goals**: <1% CPU overhead to virt-handler, scales to 1000+ concurrent VMIs  
**Security Constraints**: Standard Prometheus metrics endpoint security, no sensitive data exposure  
**Scale/Scope**: Per-VMI metric, cluster-wide collection, node-level emission  
**Integration Points**: pkg/monitoring/metrics/virt-handler/metrics.go SetupMetrics(), domainstats collector patterns, shared VMI informer

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Implementation Gates (CRITICAL - Must pass before any implementation)

#### Simplicity Gate (Constitution Article VII)

- [x] Using ≤3 new Go packages for initial implementation? (Only adding metric collection logic to existing virt-handler)
- [x] No future-proofing or premature optimization? (Basic hypervisor types only: kvm, qemu-tcg, unknown)
- [x] Minimal additional complexity to existing codebase? (Single metric addition to existing metrics collection)
- [x] Clear justification for any new abstractions? (No new abstractions - using existing libvirt and Prometheus patterns)

#### Anti-Abstraction Gate (Constitution Article VIII)

- [x] Using existing KubeVirt and Kubernetes patterns directly? (Existing virt-handler metrics collection, libvirt integration)
- [x] Not creating unnecessary wrapper layers? (Direct libvirt domain XML parsing, standard Prometheus metrics)
- [x] Following established libvirt/QEMU integration patterns? (Using existing libvirt connection patterns in virt-handler)
- [x] Single clear representation for feature configuration? (Simple metric with standardized labels)

#### KubeVirt Architectural Gates

- [x] Follows the KubeVirt Razor principle (Pod-VM feature parity)? (Provides same observability for VMs as Pods have for containers)
- [x] Feature gate implemented and disabled by default? (N/A - metrics addition doesn't require feature gate)
- [x] Integration and e2e tests planned before unit tests? (Functional tests verifying metric accuracy and lifecycle behavior)
- [x] Uses choreography pattern (components react to observed state)? (virt-handler reacts to VMI lifecycle events)
- [x] Security considerations documented with threat model? (Minimal security impact documented)
- [x] Does not grant VM users capabilities beyond what Pods already have? (Read-only observability metric only)
- [x] Backward compatibility maintained for existing APIs? (No API changes required)
- [x] Component integration follows established KubeVirt patterns? (Standard virt-handler node agent pattern)

#### Dependency Gate

- [x] All upstream dependencies confirmed available? (libvirt and Prometheus client already used)
- [x] Host environment requirements validated? (No additional host requirements)
- [x] libvirt/QEMU feature availability confirmed? (Domain XML parsing available in all supported versions)
- [x] Kubernetes API compatibility verified? (No Kubernetes API changes needed)

**If any gate fails, implementation must pause until issues are resolved.**

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md                    # This file (/plan command output)
├── research.md                # Phase 0 output (/plan command)
├── data-model.md              # Phase 1 output (/plan command)
├── quickstart.md              # Phase 1 output (/plan command)
├── component-integration.md   # Phase 1 output (/plan command)
├── contracts/                 # Phase 1 output (/plan command)
│   ├── api-schema.yaml        # CRD modifications
│   └── integration-contracts.md
├── implementation-details/    # Detailed technical specs
└── tasks.md                   # Phase 2 output (/tasks command - NOT created by /plan)
```

### KubeVirt Component Architecture

```text
# KubeVirt Component Structure (choose applicable components)
pkg/
├── virt-config/
│   └── featuregate/          # Feature gate implementation
├── virt-operator/            # [IF_OPERATOR_CHANGES]
│   ├── resources/
│   └── [FEATURE_COMPONENTS]
├── virt-controller/          # [IF_CONTROLLER_CHANGES]
│   ├── watch/
│   └── [FEATURE_CONTROLLERS]
├── virt-handler/             # [IF_HANDLER_CHANGES]
│   ├── node/
│   └── [FEATURE_HANDLERS]
├── virt-launcher/            # [IF_LAUNCHER_CHANGES]
│   ├── domain/
│   └── [FEATURE_DOMAIN_LOGIC]
├── virt-api/                 # [IF_API_CHANGES]
│   ├── webhooks/
│   │   ├── validating-webhook/
│   │   └── mutating-webhook/
│   └── [FEATURE_VALIDATIONS]
└── [OTHER_SHARED_PACKAGES]

api/
├── core/v1/                  # [IF_CRD_CHANGES]
│   ├── types.go             # VMI/VM spec modifications
│   └── [FEATURE_TYPES]
└── openapi-spec/            # Generated API documentation

tests/
├── [FEATURE_UNIT_TESTS]     # Component-specific unit tests
├── [FEATURE_INTEGRATION_TESTS] # KubeVirt integration tests
└── [FEATURE_E2E_TESTS]      # End-to-end user workflows

cmd/
└── [IF_NEW_BINARIES_NEEDED] # New command-line tools (rare)
```

### Component Integration Map

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  virt-operator  │────│ virt-controller │────│   virt-handler  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
[OPERATOR_ROLE]         [CONTROLLER_ROLE]         [HANDLER_ROLE]
                                 │                       │
                                 ▼                       ▼
                        ┌─────────────────┐    ┌─────────────────┐
                        │   virt-api      │    │ virt-launcher   │
                        └─────────────────┘    └─────────────────┘
                                 │                       │
                                 ▼                       ▼
                        [API_VALIDATION_ROLE]    [VM_EXECUTION_ROLE]
```

**Structure Decision**: 
- **Primary Component**: virt-handler (node agent) - detects hypervisor type and emits metric
- **Integration Pattern**: Event-driven detection on VMI lifecycle changes (startup, resume, migration)
- **Choreography**: virt-handler observes VMI status changes and queries libvirt domain XML to determine hypervisor type
- **Metric Emission**: Integrated with existing virt-handler metrics collection infrastructure
- **No API Changes**: Uses existing VMI resources and libvirt APIs, no new CRDs or controllers required

## Phase 0: KubeVirt Architecture Research

1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → KubeVirt-specific research task
   - For each component dependency → choreography patterns task
   - For each integration point → existing KubeVirt patterns task
   - For each host dependency → libvirt/QEMU capability research

2. **Generate and dispatch KubeVirt research agents**:

   ```text
   For each unknown in Technical Context:
     Task: "Research {unknown} for KubeVirt {component} integration"
   For each component interaction:
     Task: "Find KubeVirt choreography patterns for {interaction}"
   For each virtualization feature:
     Task: "Research libvirt/QEMU capabilities for {feature}"
   For each API change:
     Task: "Research Kubernetes CRD best practices for {change}"
   ```

3. **Consolidate findings** in `research.md` using KubeVirt format:
   - **KubeVirt Razor Analysis**: How feature maintains Pod-VM parity
   - **Component Choreography**: How components will coordinate
   - **Technology Decisions**: libvirt/QEMU/K8s API choices with rationale
   - **Feature Gate Strategy**: Implementation approach and backward compatibility
   - **Security Model**: How feature maintains Pod security equivalence
   - **Integration Points**: Existing KubeVirt features that interact
   - **Alternatives Considered**: What approaches were evaluated and rejected

**Output**: research.md with all NEEDS CLARIFICATION resolved and KubeVirt architectural decisions documented

## Phase 1: KubeVirt Design & Integration Contracts

Prerequisites: research.md complete with all constitutional gates passed

1. **Extract KubeVirt entities from feature spec** → `data-model.md`:
   - **CRD Modifications**: VMI/VM spec and status field additions
   - **Controller State Models**: State machines for reconciliation loops
   - **API Validation Rules**: Webhook validation requirements
   - **Feature Gate Integration**: How feature toggles affect API behavior
   - **Component Data Flow**: How data flows between virt-* components

2. **Generate KubeVirt API contracts** from functional requirements:
   - **CRD Schema Contracts**: OpenAPI schema for new API fields
   - **Controller Reconciliation Contracts**: Expected state transitions
   - **Component Integration Contracts**: How virt-* components coordinate
   - **Feature Gate Contracts**: API behavior with feature enabled/disabled
   - Output Kubernetes-compliant schemas to `/contracts/`

3. **Generate KubeVirt integration contracts** from choreography patterns:
   - **Watch Contracts**: What resources each controller observes
   - **Event Contracts**: What events trigger reconciliation
   - **Status Update Contracts**: How components communicate via status
   - **Error Handling Contracts**: How failures propagate through system

4. **Generate contract tests** following KubeVirt patterns:
   - **API Validation Tests**: CRD schema validation (must fail initially)
   - **Controller Integration Tests**: Component interaction scenarios
   - **Feature Gate Tests**: Behavior verification with feature on/off
   - **End-to-End Tests**: User workflow validation scenarios
   - Use Ginkgo/Gomega framework following KubeVirt conventions

5. **Create component integration documentation** → `component-integration.md`:
   - **Architecture Diagram**: Component interaction visualization
   - **Choreography Flows**: How components react to state changes
   - **Security Boundaries**: How feature maintains Pod security equivalence
   - **Performance Characteristics**: Expected resource usage and scaling

6. **Extract test scenarios** from user stories:
   - Each story → KubeVirt E2E test scenario
   - Quickstart test = VMI lifecycle validation steps
   - Integration test = component coordination validation

7. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh copilot`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW KubeVirt patterns from current plan
   - Include component integration context
   - Update constitutional compliance markers
   - Keep under 150 lines for token efficiency
   - Output to repository root (`.github/copilot-instructions.md`)

**Output**: data-model.md, component-integration.md, /contracts/*, failing KubeVirt integration tests, quickstart.md, .github/copilot-instructions.md

## Phase 2: KubeVirt Task Planning Approach

This section describes what the /tasks command will do - DO NOT execute during /plan

**KubeVirt Task Generation Strategy for VMI Hypervisor Metric**:

- Load `.specify/templates/tasks-template.md` as base with KubeVirt patterns
- Generate tasks from Phase 1 design docs (component-integration, contracts, data-model)
- **Foundation Tasks**: Hypervisor detection logic, libvirt XML parsing [P]
- **Component Tasks**: virt-handler metrics integration, VMI event handling
- **Integration Tasks**: Metrics lifecycle management, error handling patterns
- **Testing Tasks**: Integration-first approach (E2E VMI metrics → unit tests)
- **Validation Tasks**: Performance testing, constitutional compliance verification

**KubeVirt Ordering Strategy**:

- **Integration-First Order**: E2E tests → Integration tests → Unit tests
- **Component Dependency Order**: APIs → Controllers → Handlers → Launchers
- **Constitutional Gates**: Validate compliance at each major milestone
- **Feature Gate Strategy**: Disabled-by-default implementation with toggle validation
- Mark [P] for parallel execution (independent components/files)

**Feature-Specific Task Categories**:

- **Pre-Implementation Gates**: Constitutional compliance validation (already passed)
- **Foundation Phase**: Hypervisor detection logic, type enumeration, XML parsing
- **Core Implementation Phase**: virt-handler metrics collector integration, VMI lifecycle hooks
- **Integration Testing Phase**: VMI creation/deletion metric validation, libvirt integration testing  
- **Production Readiness Phase**: Performance validation, error handling, logging, documentation

**Estimated Output**: 25-35 numbered, ordered tasks in tasks.md (simpler than typical due to single-component, additive nature)

**IMPORTANT**: This phase is executed by the /tasks command following KubeVirt constitutional principles, NOT by /plan

## Phase 3+: Future Implementation

These phases are beyond the scope of the /plan command

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking

Fill ONLY if Constitution Check has violations that must be justified

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |

## Progress Tracking

This checklist is updated during execution flow

**Phase Status**:

- [x] Phase 0: Research complete (/plan command) - Generated research.md
- [x] Phase 1: Design complete (/plan command) - Generated data-model.md, contracts/, quickstart.md, component-integration.md
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:

- [x] Initial Constitution Check: PASS (all gates satisfied)
- [x] Post-Design Constitution Check: PASS (no design violations)
- [x] All NEEDS CLARIFICATION resolved (via clarification session)
- [x] Complexity deviations documented (none - simple additive feature)

---

## KubeVirt Implementation Guidelines

### Constitutional Compliance Throughout Implementation

- **KubeVirt Razor Principle**: Continuously validate Pod-VM feature parity
- **Feature Gate Discipline**: All new functionality behind disabled-by-default gates
- **Integration-First Testing**: E2E and integration tests before unit tests
- **Component Choreography**: Independent component reactions to observed state
- **Security Equivalence**: VM users cannot exceed Pod user capabilities
- **Simplicity Maintenance**: Resist abstraction layers and complexity

### Quality Standards for KubeVirt Features

- **Code Quality**: Follow established KubeVirt patterns and conventions
- **Test Coverage**: Comprehensive integration and E2E test coverage
- **Documentation**: User guides, API docs, and troubleshooting information
- **Security Review**: Threat model analysis and security boundary validation
- **Performance Validation**: Resource usage and scaling characteristics
- **Backward Compatibility**: Maintain API compatibility and migration paths

### Implementation Execution Guidelines

- **Follow Constitutional Gates**: Validate compliance at each phase
- **Use Established Patterns**: Leverage existing KubeVirt architectural patterns
- **Integration-First Development**: Build and test component interactions early
- **Feature Gate Compliance**: Ensure feature can be safely disabled
- **Security-First Approach**: Maintain Pod security model equivalence

---
*Based on KubeVirt Constitution v1.0.0 - See `.specify/memory/constitution.md`*
