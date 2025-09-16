# Implementation Plan: Hyper-V L1VH (Transformed)

Source: `specs/001-hyperv-l1vh/implementation-plan.md`  
Imported: 2025-09-16

## Summary

Transparent enablement of Hyper-V L1VH through feature gate without user spec changes. Converter selects mshv when gate enabled and /dev/mshv present; fallback to KVM otherwise.

## Technical Context

Language: Go 1.21+  
Primary Dependencies: libvirt (mshv support), qemu (mshv backend), Linux kernel 6.6+, Azure L1VH-capable hosts  
Storage: Existing KubeVirt storage stack (unchanged)  
Testing: Ginkgo/Gomega, existing integration & E2E harness  
Target Platform: Homogeneous Azure L1VH-capable Kubernetes cluster  
Project Type: KubeVirt extension (converter logic + feature gate)  
Performance Goals: Near-native vs nested virtualization baseline (quantitative targets TBD) [NEEDS CLARIFICATION]  
Constraints: Zero API change, zero RBAC change  
Scale/Scope: Cluster-wide behavior; all VMs benefit when gate enabled

## Constitution Check (Derived)

- Feature Gate Discipline: Present (`HyperVL1VH` alpha, default off)
- Simplicity: ≤2 modified packages (featuregate, converter)
- Test-First: Plan enumerates contract, integration, E2E, unit tests before implementation
- Observability: Metrics proposed (hypervisor selection, fallback counts)
- Security: No new privileged surfaces beyond /dev/mshv device access assumption
- Anti-Abstraction: Reuses existing converter path

Pending Clarifications:

- Memory overhead model differences vs KVM? [NEEDS CLARIFICATION]
- Process memory limit adjustments for mshv vs virtqemud? [NEEDS CLARIFICATION]
- Device resource accounting (new resource name?) [NEEDS CLARIFICATION]

## Phase 0: Research Outline

Will confirm: libvirt/qemu versions for mshv, domain XML requirements, memory overhead, device resource semantics, security implications of exposing /dev/mshv.

## Phase 1: Design & Contracts (Planned Artifacts)

- Converter contract tests (selection + fallback)
- Memory overhead unit tests (baseline vs L1VH)
- Integration: Existing VM lifecycle tests executed with gate toggled
- E2E: Reuse existing flows; add one hypervisor selection visibility test

## Phase 2: Task Planning Approach

See source tasks; will map to `.specify/templates/tasks-template.md` generation rules: tests (Red) → minimal converter additions → metrics → documentation.

## Complexity Tracking

None beyond added detection path. If memory overhead diverges significantly causing added conditional branches, document justification.

## Progress Tracking (Initial)

- [ ] Phase 0 Research
- [ ] Phase 1 Design
- [ ] Phase 2 Task Strategy (described)
- [ ] Phase 3 Tasks Generated (external)

## Open Questions

1. What quantitative performance uplift target defines success?  
2. Are there edge cases for live migration between mshv and kvm nodes (out-of-scope currently)?  
3. How to surface hypervisor type to users (conditions vs metrics vs docs only)?  
4. Memory overhead delta vs KVM baseline for sizing?  
5. GPU passthrough enumeration path under mshv?  

## Next Actions

Populate research.md with version matrix + overhead findings → finalize tests → implement gated detection.
