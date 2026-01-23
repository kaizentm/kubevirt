package hypervisor

import (
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/hypervisor/kvm"
	"kubevirt.io/kubevirt/pkg/hypervisor/mshv"
)

type LauncherResourceRenderer interface {
	GetHypervisorDevice() string
	GetMemoryOverhead(vmi *v1.VirtualMachineInstance, arch string, additionalOverheadRatio *string) resource.Quantity
	GetVirtType() string
	GetHypervisorDeviceMinorNumber() int64
}

func NewLauncherResourceRenderer(hypervisor string) LauncherResourceRenderer {
	switch hypervisor {
	// Other hypervisors can be added here
	case v1.HyperVDirectHypervisorName:
		return mshv.NewMshvLauncherResourceRenderer()
	default:
		return kvm.NewKvmLauncherResourceRenderer() // Default to KVM
	}
}
