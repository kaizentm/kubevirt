<!--
Sync Impact Report:
- Version change: new → 1.0.0
- Initial constitution creation for KubeVirt project
- Added 5 core principles aligned with Kubernetes virtualization patterns
- Templates requiring updates: ✅ all consistent
- Follow-up TODOs: none
-->

# KubeVirt Constitution

## Core Principles

### I. Kubernetes-Native Architecture

All KubeVirt features MUST extend Kubernetes through CRDs and controllers following cloud-native patterns. Components run as pods with clear separation: virt-api (admission/validation), virt-controller (orchestration), virt-handler (node agent), virt-launcher (VMI execution). Choreography pattern is mandatory - controllers react to CR changes rather than centralized reconciliation. No direct host system modifications outside containerized environments.

**Rationale**: Ensures KubeVirt remains a true Kubernetes add-on, maintainable through standard K8s tooling and lifecycle management.

### II. Reproducible Build System (NON-NEGOTIABLE)

All builds MUST use `make bazel-build-images` through containerized environment (`hack/dockerized`). Image content derived from RPM dependency trees via `make rpm-deps` + `hack/rpm-deps.sh`. NO direct package installation in Dockerfiles - extend RPM lists instead. Generated code changes require `make generate` + `make generate-verify`. Multi-architecture support mandatory without architecture-specific conditionals.

**Rationale**: Guarantees consistent, auditable builds across environments and prevents supply chain vulnerabilities through controlled dependency management.

### III. API Backward Compatibility

All API changes MUST be backward compatible. Add new optional fields only - never rename or remove existing fields. Changes to `api/` types require: edit types → `make generate` → commit all generated files + OpenAPI schema. Validation/defaulting logic added to `pkg/virt-api/webhooks/`. Document changes in `docs/` and update CRD schemas. Feature gating through config CRs preferred over environment variables.

**Rationale**: Protects existing KubeVirt deployments and ensures smooth upgrade paths for users managing production VM workloads.

### IV. Testing Discipline

Functional tests using Ginkgo framework are mandatory for all behavior changes. Unit tests focused on pure logic packages. Changes impacting image content or API surfaces MUST include generated file updates in same PR. Test reproducibility through `hack/dockerized` environment. Integration tests required for: new controller logic, API contract changes, inter-component communication, VM lifecycle operations.

**Rationale**: VM management demands high reliability - comprehensive testing prevents regression in critical virtualization workflows.

### V. Component Separation & Reusability

Prefer reusing existing utilities in `pkg/` over creating new abstractions. Each component (`cmd/virt-*`) follows established flag and logging patterns. Shared logic in libraries, not duplicated across components. Clear boundaries: API helpers, device/network/storage logic, validation, informers. New dependencies justify impact on image size - every RPM inflates multiple container images.

**Rationale**: Maintains codebase coherence and prevents bloat in containerized virtualization stack where resource efficiency directly impacts node capacity.

## Development Workflow

All changes follow Make-based workflow with Bazel backend. Code generation and dependency management through established tooling prevents manual errors. Pull requests require functional test coverage and generated file consistency. Architecture changes require design proposals following community template. Performance-critical paths (launch flows, monitoring) follow established caching/timing patterns.

**Workflow Requirements**: Use `make` targets exclusively, never raw `bazel` commands. Validate via `make bazel-build-verify`. Complex changes start with community design proposal. Generated files committed alongside source changes.

## Quality Assurance

Structured logging with consistent verbosity flags across components. Error handling through return values - controllers use event recording and status updates, never panics. Feature flags through config APIs where possible. Multi-arch support without conditional compilation. Performance awareness in hot paths due to direct impact on VM operations.

**Standards**: Follow existing component patterns for logging/flags. Justify complexity additions. Keep unit tests focused and fast. Ensure E2E test coverage for user-facing features.

## Governance

This constitution supersedes all other development practices. All pull requests and code reviews MUST verify compliance with these principles. Complexity additions require justification against simplicity principle. Changes affecting multiple components need architectural review. Use `.github/copilot-instructions.md` for runtime development guidance and established patterns.

**Amendment Process**: Constitution changes require community discussion and approval. Version increments follow semantic versioning: MAJOR for incompatible governance changes, MINOR for new principles/sections, PATCH for clarifications.

**Version**: 1.0.0 | **Ratified**: 2025-09-25 | **Last Amended**: 2025-09-25
