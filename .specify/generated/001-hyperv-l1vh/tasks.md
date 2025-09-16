# Tasks: Hyper-V L1VH (Normalized)

Source: `specs/001-hyperv-l1vh/tasks.md`  
Imported: 2025-09-16

> Condensed to align with `.specify/templates/tasks-template.md` conventions (ID format T###, [P] for parallel). Non-implementation research retained where test-first depends on it. Excess narrative removed; full detail remains in source document.

## Phase 3.1 Setup

- [ ] T001 Validate libvirt & qemu mshv support matrix (docs/l1vh-domain-xml-requirements.md)
- [ ] T002 [P] Add feature gate constant in `pkg/virt-config/featuregate/feature-gates.go` (disabled by default)
- [ ] T003 [P] Prepare Azure L1VH test cluster checklist (docs/hyperv-l1vh-cluster-setup.md)

## Phase 3.2 Tests First (Red)

- [ ] T004 Converter contract test: L1VH selection when gate enabled + /dev/mshv present (`pkg/virt-launcher/virtwrap/converter/converter_test.go`)
- [ ] T005 [P] Converter contract test: Fallback to KVM when device missing (`converter_test.go`)
- [ ] T006 [P] Unit tests: feature gate registration + enable/disable (`feature-gates_test.go`)
- [ ] T007 [P] Unit tests: hasL1VHSupport device presence/absence/permission (`l1vh_test.go`)
- [ ] T008 Integration: run existing VM lifecycle tests with gate off vs on (document deltas)
- [ ] T009 [P] Add one integration test for hypervisor selection visibility (`tests/hyperv_l1vh_test.go`)
- [ ] T010 Performance test plan draft (baseline metrics selection) [doc only]

## Phase 3.3 Core Implementation (Green)

- [ ] T011 Implement feature gate constant + map entry
- [ ] T012 [P] Implement `hasL1VHSupport()` in converter
- [ ] T013 [P] Inject detection + domain type override to mshv in `Convert_*` path
- [ ] T014 Domain XML configuration function `configureL1VHDomain()` (minimal viable)
- [ ] T015 Metrics: `kubevirt_vmi_hypervisor_type_total` & fallback counter
- [ ] T016 Logging: selection + fallback info lines
- [ ] T017 Optional condition emission design doc (decide if needed or metrics-only)

## Phase 3.4 Memory & Resource Validation

- [ ] T018 Research memory overhead parity vs KVM (renderresources.go)
- [ ] T019 [P] Add unit tests for memory overhead when hypervisor=mshv
- [ ] T020 Conditional: adapt `GetMemoryOverhead()` if divergence found
- [ ] T021 Investigate `AdjustQemuProcessMemoryLimits` applicability to mshv
- [ ] T022 Conditional: implement mshv-specific process limit adjustments

## Phase 3.5 Observability

- [ ] T023 Confirm metrics scrape & label cardinality safe
- [ ] T024 [P] Add doc: troubleshooting L1VH detection (docs/hyperv-l1vh.md section)

## Phase 3.6 Documentation & Release Prep

- [ ] T025 User guide draft (transparent operation, enabling gate)
- [ ] T026 [P] Migration/rollback notes (gate disable behavior)
- [ ] T027 [P] Performance benchmark procedure doc
- [ ] T028 Release notes entry (alpha feature)

## Dependencies

- Tests (T004-T010) before implementation tasks (T011-T017)
- Memory research (T018) before potential adaptation (T020, T022)
- Feature gate (T011) before hypervisor selection (T013)

## Parallelization Notes

- Device detection, feature gate tests independent; mark [P]
- Domain XML (T014) after detection logic (T012/T013)
- Memory adaptation tasks conditional; skip if parity confirmed

## Validation Checklist

- [ ] All contract tests express behavior & fail pre-implementation
- [ ] All new metrics documented
- [ ] No API / CRD changes introduced
- [ ] Fallback path verified with missing /dev/mshv simulation
- [ ] Documentation covers enable, observe, troubleshoot, rollback
