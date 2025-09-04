# KubeVirt Development Constitution

**Version**: 1.0  
**Effective Date**: September 3, 2025  
**Status**: Active  

## Preamble

This constitution establishes the immutable principles that govern all feature development within the KubeVirt project. These principles ensure consistency, quality, and architectural integrity across all contributions while providing frameworks that **enable** specification-driven development for teams that choose this approach.

KubeVirt exists to bring virtual machine management to Kubernetes through native APIs and patterns. Every feature must honor this mission by extending Kubernetes naturally rather than creating parallel systems.

This constitution supports multiple development approaches while maintaining architectural discipline and quality standards across all contributions.

## Article I: Specification-Driven Development Support

### Section 1.1: Enabling Spec-Driven Development
This constitution establishes the framework to **support** specification-driven development (SDD) for teams and features that choose this approach. When SDD is adopted for a feature:

- **Problem Statement**: Must clearly articulate what limitation is being addressed
- **User Stories**: Must express specific user needs as measurable outcomes  
- **Acceptance Criteria**: Must provide unambiguous success conditions for each user story
- **Requirements**: Must define both functional and non-functional requirements
- **Integration Points**: Must describe how the feature interacts with existing KubeVirt components

### Section 1.2: Implementation Plans for Spec-Driven Features
When a feature follows the specification-driven approach, a detailed implementation plan **MUST** be created that translates business requirements into technical architecture. These implementation plans serve as executable blueprints that can generate consistent, maintainable code.

### Section 1.3: Traditional Development Paths
Features **MAY** follow traditional development approaches including:
- Direct implementation with design documents
- Iterative development with evolving requirements
- Prototype-first development with post-hoc documentation
- Emergency fixes and critical patches

All development approaches **MUST** still comply with the architectural principles defined in Articles II-XII of this constitution.

### Section 1.4: Specification Evolution (SDD Features Only)
For features following specification-driven development, specifications are living documents that evolve through implementation and operational feedback. Changes to specifications drive regeneration of implementation plans and code, maintaining perfect alignment between intent and reality.

## Article II: The KubeVirt Razor

### Section 2.1: Native Kubernetes Integration
**"If something is useful for Pods, we should not implement it only for VMs"**

Every feature **MUST** leverage existing Kubernetes patterns and APIs rather than creating VM-specific alternatives. This principle ensures:

- Consistent user experience across workload types
- Reduced cognitive load for cluster operators
- Natural integration with Kubernetes tooling and workflows
- Future compatibility with Kubernetes evolution

### Section 2.2: Non-Privileged Extensions
Features **MUST NOT** grant users capabilities beyond what Kubernetes already provides. Virtual machine management should feel like native Kubernetes resource management, not privileged system administration.

### Section 2.3: Choreography Over Orchestration
KubeVirt components **MUST** follow the choreography pattern where services act independently based on observed state rather than being centrally orchestrated. This ensures:

- Resilient operation under failure conditions
- Scalable architecture with loose coupling
- Kubernetes-native reconciliation patterns
- Self-healing behavior

## Article III: Feature Gate Discipline

### Section 3.1: Alpha-First Development
All new features **MUST** begin life disabled by default behind a feature gate. No exceptions.

### Section 3.2: Gate Implementation Requirements
Feature gates **MUST**:
- Be implemented in `pkg/virt-config/featuregate/`
- Follow KubeVirt's established feature gate patterns
- Allow complete feature disable/enable without system restart
- Maintain system stability when disabled
- Provide clear documentation of functionality scope

### Section 3.3: Graduation Criteria
Movement between Alpha, Beta, and GA **MUST** follow documented criteria:
- **Alpha to Beta**: Feature complete, stable API, no known critical bugs
- **Beta to GA**: Battle-tested in production, comprehensive documentation, migration paths defined
- **GA Requirements**: Long-term API compatibility commitment, full observability support

## Article IV: Test-First Implementation

### Section 4.1: Test-Driven Development Mandate
**This is NON-NEGOTIABLE**: All implementation **MUST** follow strict Test-Driven Development (TDD).

No implementation code shall be written before:
1. Comprehensive tests are written and approved
2. Tests are validated to correctly express requirements  
3. Tests are confirmed to FAIL (Red phase of TDD)
4. User acceptance of test specifications

### Section 4.2: Test Hierarchy
Tests **MUST** be implemented in this order:
1. **Contract Tests**: Define API behaviors and integration points
2. **Integration Tests**: Validate component interactions in realistic environments
3. **End-to-End Tests**: Verify complete user workflows
4. **Unit Tests**: Validate internal logic and edge cases

### Section 4.3: Real Environment Testing
Tests **MUST** prefer real environments over mocks:
- Use actual libvirt/QEMU instances over simulators
- Test against real Kubernetes clusters, not test doubles
- Validate storage and networking integration with real implementations
- Only mock external dependencies that cannot be controlled

## Article V: Component Architecture

### Section 5.1: Service-Oriented Design
All features **MUST** respect KubeVirt's service-oriented architecture:

- **virt-operator**: Cluster-wide lifecycle management and configuration
- **virt-controller**: VM resource reconciliation and state management  
- **virt-handler**: Node-specific VM lifecycle and resource management
- **virt-launcher**: VM process management and isolation
- **virt-api**: API validation, defaulting, and webhook implementations

### Section 5.2: Component Responsibility Boundaries
Components **MUST** respect established responsibility boundaries:

- **virt-operator** manages KubeVirt installation and configuration
- **virt-controller** reconciles VM and VMI resources
- **virt-handler** manages per-node VM operations
- **virt-launcher** provides VM process isolation and monitoring
- **virt-api** validates and defaults API requests

### Section 5.3: Inter-Component Communication
Components **MUST** communicate through:
- Kubernetes API resources and events
- Established controller patterns (watch/reconcile)
- Standard Kubernetes inter-service communication
- No direct service-to-service communication

## Article VI: API Design Principles

### Section 6.1: Kubernetes API Conventions
All APIs **MUST** follow [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md):

- Consistent field naming and structure
- Proper API versioning (v1alpha1, v1beta1, v1)
- Standard metadata patterns
- Kubernetes-native validation and defaulting

### Section 6.2: Custom Resource Design
New Custom Resource Definitions **MUST**:
- Extend existing VirtualMachine/VirtualMachineInstance resources when possible
- Follow established KubeVirt CRD patterns
- Provide comprehensive OpenAPI schemas
- Include proper field validation and defaulting

### Section 6.3: API Stability Guarantees  
API design **MUST** provide appropriate stability guarantees:
- **Alpha APIs**: No compatibility guarantees, may change without notice
- **Beta APIs**: Compatible within beta versions, documented migration paths for breaking changes  
- **Stable APIs**: Strong backward compatibility commitment, deprecation policies enforced

## Article VII: Simplicity and Minimalism

### Section 7.1: Minimal Viable Implementation
Features **MUST** start with the minimal viable implementation that satisfies user stories. Avoid:
- Premature optimization
- Speculative features
- Future-proofing beyond documented requirements
- Complex abstractions without proven need

### Section 7.2: Complexity Justification
Any complexity beyond simple, direct implementation **MUST** be explicitly justified:
- Document the specific user problem that requires complexity
- Demonstrate why simpler alternatives are insufficient
- Provide clear criteria for success and failure
- Plan for complexity removal when conditions change

### Section 7.3: Minimal Project Structure
Initial implementations **MUST** use ≤3 Go packages. Additional packages require:
- Documented justification for separation
- Clear responsibility boundaries
- Demonstrated maintainability benefits
- Architecture review approval

## Article VIII: Anti-Abstraction

### Section 8.1: Framework Trust
**Trust the frameworks**: Use KubeVirt, Kubernetes, libvirt, and QEMU features directly rather than creating wrapper layers.

Abstractions are **PROHIBITED** unless they:
- Solve a specific, documented problem
- Cannot be addressed by existing framework features
- Provide clear, measurable benefits
- Have been approved through architecture review

### Section 8.2: Single Model Representation
Domain concepts **MUST** have single, canonical representations:
- One data structure per domain concept
- No parallel modeling systems
- Clear mapping between API resources and internal models
- Consistent field naming across representations

### Section 8.3: Direct Integration Patterns
Prefer direct integration with upstream projects:
- Use libvirt APIs directly for virtualization management
- Integrate with QEMU features without intermediate layers
- Leverage Kubernetes controllers and webhooks natively
- Follow established patterns from successful integrations

## Article IX: Integration-First Testing

### Section 9.1: Integration Test Priority
Integration tests **MUST** be prioritized over unit tests:
- Test component interactions in realistic environments
- Validate end-to-end workflows before testing individual functions
- Use real dependencies (databases, APIs, services) when possible
- Focus on user-visible behaviors rather than internal implementation

### Section 9.2: Realistic Test Environments
Test environments **MUST** reflect production conditions:
- Test with real libvirt and QEMU installations
- Use actual Kubernetes clusters, not test environments
- Include realistic networking and storage configurations
- Test across supported operating systems and versions

### Section 9.3: Contract-Driven Testing
All inter-component interactions **MUST** be defined by contracts:
- API contracts specify exact request/response formats
- Service contracts define behavioral expectations
- Integration contracts verify component boundaries
- Contracts are tested independently of implementations

## Article X: Constitutional Enforcement

### Section 10.1: Pre-Implementation Gates
All implementations **MUST** pass constitutional gates before code development:

#### Simplicity Gate
- [ ] Using ≤3 Go packages for initial implementation?
- [ ] No future-proofing or premature optimization?
- [ ] Complexity is justified and documented?

#### Anti-Abstraction Gate  
- [ ] Using existing KubeVirt/Kubernetes patterns directly?
- [ ] Not creating unnecessary wrapper layers?
- [ ] Single model representation per domain concept?

#### Integration-First Gate
- [ ] Integration tests planned before unit tests?
- [ ] Real environment testing prioritized?
- [ ] Contracts defined for all component interactions?

#### KubeVirt Razor Gate
- [ ] Feature follows "useful for Pods" principle?
- [ ] No privileged capabilities beyond Kubernetes?
- [ ] Choreography pattern respected?

### Section 10.2: Gate Failure Protocol
If any constitutional gate fails:
1. Implementation **MUST** pause until issues are resolved
2. Failures **MUST** be documented with specific remediation plans
3. Architecture review **MUST** approve any exceptions
4. Remediation **MUST** be verified before proceeding

### Section 10.3: Template Enforcement
Implementation plan templates **MUST** enforce constitutional compliance through:
- Embedded gate checkpoints that cannot be bypassed
- Automatic validation of constitutional principles
- Required documentation of complexity justifications
- Mandatory architectural review triggers

## Article XI: Observability and Operations

### Section 11.1: Metrics and Monitoring
All features **MUST** provide comprehensive observability:
- Prometheus metrics for operational health
- Structured logging for troubleshooting
- Kubernetes events for user notification
- Performance metrics for capacity planning

### Section 11.2: Debugging Support
Features **MUST** include debugging capabilities:
- CLI tools for operational inspection
- Debug endpoints for internal state
- Comprehensive error messages with remediation guidance
- Integration with existing KubeVirt debugging tools

### Section 11.3: Security Posture
Security **MUST** be designed in, not added later:
- Threat modeling during specification phase
- Security review before implementation approval
- Regular security scanning and vulnerability assessment
- Incident response procedures for security issues

## Article XII: Documentation as Code

### Section 12.1: Documentation Requirements
Documentation **MUST** be treated as first-class deliverable:
- User documentation for all public features
- API documentation generated from OpenAPI schemas
- Operational runbooks for troubleshooting
- Architecture documentation for maintainers

### Section 12.2: Documentation Evolution
Documentation **MUST** evolve with specifications:
- Examples updated when APIs change
- Troubleshooting guides reflect current behavior
- Migration guides for breaking changes
- Deprecation notices with clear timelines

## Article XIII: Amendment Process

### Section 13.1: Constitutional Modification
This constitution may only be modified through:
- Explicit documentation of rationale for change
- Review and approval by KubeVirt maintainers
- Backward compatibility impact assessment
- Community feedback period
- Formal adoption through governance process

### Section 13.2: Constitutional Evolution
Constitutional amendments **MUST**:
- Be dated and versioned
- Include migration guidance for existing features
- Provide grandfathering policies for existing implementations
- Document lessons learned that motivated changes

### Section 13.3: Constitutional Supremacy
In case of conflicts between this constitution and other project documents, constitutional principles take precedence. Implementation plans and feature specifications **MUST** align with constitutional requirements.

---

## Enforcement Mechanisms

This constitution is enforced through:

1. **Template Integration**: Implementation plan templates embed constitutional gates
2. **Automated Validation**: CI/CD pipelines validate constitutional compliance  
3. **Review Processes**: Code reviews verify constitutional adherence
4. **Community Standards**: Community members uphold constitutional principles
5. **Mentorship Programs**: New contributors learn constitutional requirements

## Constitutional Benefits

This constitution provides:

- **Consistency**: All features follow the same architectural principles
- **Quality**: Built-in quality gates prevent architectural debt
- **Maintainability**: Simple, direct implementations are easier to maintain
- **Evolution**: Clear patterns enable rapid, safe feature development
- **Integration**: Native Kubernetes patterns ensure seamless operation

## Living Document

This constitution is a living document that evolves with the KubeVirt project while maintaining core architectural integrity. It serves as both the foundation for current development and the guide for future evolution.

---

*"Democracy is the worst form of government, except for all the others that have been tried."* - Winston Churchill

*"A constitution is the worst form of software architecture, except for all the others that have been tried."* - KubeVirt Community
