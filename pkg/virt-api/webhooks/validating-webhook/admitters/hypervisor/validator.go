package hypervisor_validator

import (
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"

	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"

	v1 "kubevirt.io/api/core/v1"

	kvm_validator "kubevirt.io/kubevirt/pkg/virt-api/webhooks/validating-webhook/admitters/hypervisor/kvm"
	mshv_validator "kubevirt.io/kubevirt/pkg/virt-api/webhooks/validating-webhook/admitters/hypervisor/mshv"
)

type Validator interface {
	// Validate spec of VirtualMachineInstance
	ValidateVirtualMachineInstanceSpec(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause

	// Validate hot-plug updates to VMI. For example, this would encapsulate functionality in the ValidateHotplugDiskConfiguration function.
	ValidateHotplug(oldVmi *v1.VirtualMachineInstance, newVmi *v1.VirtualMachineInstance, cc *virtconfig.ClusterConfig) *admissionv1.AdmissionResponse
}

func NewValidator(hypervisor string) Validator {
	switch hypervisor {
	case v1.HyperVLayeredHypervisorName:
		return &mshv_validator.MshvValidator{}
	default:
		return &kvm_validator.KvmValidator{}
	}
}
