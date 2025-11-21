package kvm_validator

import (
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"
	v1 "kubevirt.io/api/core/v1"

	base_validator "kubevirt.io/kubevirt/pkg/virt-api/webhooks/validating-webhook/admitters/hypervisor/base"
	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
)

type KvmValidator struct {
	*base_validator.BaseValidator
}

func (kv *KvmValidator) ValidateVirtualMachineInstanceSpec(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	return kv.BaseValidator.ValidateVirtualMachineInstanceSpec(field, spec, config)
}

func (kv *KvmValidator) ValidateHotplug(oldVmi *v1.VirtualMachineInstance, newVmi *v1.VirtualMachineInstance, cc *virtconfig.ClusterConfig) *admissionv1.AdmissionResponse {
	return kv.BaseValidator.ValidateHotplug(oldVmi, newVmi, cc)
}
