Under `pkg/utils/webhooks/validating-webhooks`, the `admitter` interface is defined, which has only 1 function: `Admit(context.Context, *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse`

`vmi-create-admitter.go` contains the `VMICreateAdmitter` struct that implements the `admitter` interface mentioned above.

Here are the actions taken by the `Admit` function in `VMICreateAdmitter`:

- webhookutils.ValidateSchema


- The the following code validates the network admitter (`netadmitter`) to validate multiple networking related specs.

```go
    for _, validateSpec := range admitter.SpecValidators {
        causes = append(causes, validateSpec(k8sfield.NewPath("spec"), &vmi.Spec, admitter.ClusterConfig)...)
    }
```

The purpose of this extra code for validation is to allow extensions to the validation logic - i.e., perform extra validation apart from the VMICreateAdmitter. PR for this change: https://github.com/kubevirt/kubevirt/pull/13388.
There is a follow-up task for this PR.

- Then, `ValidateVirtualMachineInstanceSpec` function is called for an intensive validation of VMI spec. This function is also called when validating the parent object `VirtualMachine`.
  - This function is also called from other Admitters, e.g., VMsAdmitter, VMPoolAdmitter, VMIRSAdmitter, VMIUpdateAdmitter.


- Then we validate that volumes are mapped to disks or filesystems during VMI admittance. This is done in the fn `validateVirtualMachineInstanceSpecVolumeDisks`.

- Then `ValidateVirtualMachineInstanceMandatoryFields`.

- Then `ValidateVirtualMachineInstanceMetadata`

- `ValidateVirtualMachineInstanceHyperv`: validate if hyperv features are supported.

- `ValidateVirtualMachineInstancePerArch`: 
  - For each architecture, there are different implementations of validations. The diff validations are also validating diff parts of the spec, based on the peculiarities of the hypervisor.



-------------------------------


For extensible validation webhooks, we define the `Validator` interface, for validating multiple aspects of `VirtualMachine` and `VirtualMachineInstance` definitions.

```go
type Validator interface {
    // Validate spec of VirtualMachine
    ValidateVirtualMachineSpec(field *k8sfield.Path, spec *v1.VirtualMachineSpec, config *virtconfig.ClusterConfig, isKubeVirtServiceAccount bool) []metav1.StatusCause

    // Validate spec of VirtualMachineInstance
    ValidateVirtualMachineInstanceSpec(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause
    
    // Validate hot-plug updates to VMI
    ValidateHotplug(oldVmi *v1.VirtualMachineInstance, newVmi *v1.VirtualMachineInstance, cc *virtconfig.ClusterConfig) []metav1.StatusCause
}
```

The above `Validator` interface would be implemented by the `BaseValidator` that contains validation functionality common across hypervisors and architectures.

```go
type BaseValidator struct {}

func (b *BaseValidator) ValidateVirtualMachineInstanceSpec (field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
    var causes []metav1.StatusCause

    ... // more validation functions
    causes = append(causes, b.validateNUMA(field, spec)...)
    causes = append(causes, b.validateGuestMemoryLimit(field, spec)...)
    ... // more validation functions

}

func (b *BaseValidator) validateNUMA(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
    // generic (hypervisor/arch-agnostic) logic for validation of VMI's NUMA spec
}
```

Following is a hypervisor-specific validator.

```go
type MshvValidator struct { *BaseValidator }

func (m *MshvValidator) ValidateVirtualMachineInstanceSpec (field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
    var causes []metav1.StatusCause

    // Execute common validation logic
    causes = append(causes, m.BaseValidator.ValidateVirtualMachineInstanceSpec(field, spec, config))

    // Run hypervisor-specific check
    causes = append(causes, m.validateCPUModel(field, spec, config))

    return causes
}

func (m *MshvValidator) validateCPUModel (field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
    // For MSHV hypervisor, the guest's CPU model has to be "qemu64-v1"
    var causes []metav1.StatusCause

    if spec.Domain.CPU.model != "qemu64-v1" {
        // append validation failure cause to causes
    }

    return causes
}

```

This can be further extended to perform architecture-specific checks for a given hypervisor.

```go
type Amd64Validator struct {}

func (a *Amd64Validator) ValidateVirtualMachineInstanceSpec (field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
    var causes []metav1.StatusCause
    causes = append(causes, a.validateWathdogAmd64(field, spec, config))
    return causes
}

type MshvAmd64Validator struct { *MshvValidator *Amd64Validator }

func (ma *MshvAmd64Validator) ValidateVirtualMachineInstanceSpec (field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
    var causes []metav1.StatusCause

    // Execute hypervisor-specific validation logic
    causes = append(causes, m.MshvValidator.ValidateVirtualMachineInstanceSpec(field, spec, config))

    // Execute arch-specific validation logic
    causes = append(causes, m.Amd64Validator.ValidateVirtualMachineInstanceSpec(field, spec, config))

    // Run hypervisor-arch-specific check
    causes = append(causes, m.validateMshvAmd64SpecificConfig(field, spec, config))

    return causes
}
```