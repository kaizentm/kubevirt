package arch_validators

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"

	v1 "kubevirt.io/api/core/v1"

	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
)

type ArchValidator interface {
	ValidateVirtualMachineInstanceArchSetting(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, cc *virtconfig.ClusterConfig) []metav1.StatusCause
}

func NewArchValidator(arch string) ArchValidator {
	switch arch {
	case "amd64":
		return &Amd64Validator{}
	case "arm64":
		return &Arm64Validator{}
	case "s390x":
		return &S390xValidator{}
	default:
		return nil
	}
}
