package hypervisor

import (
	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/hypervisor/kvm"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/converter/types"
)

func MakeDomainBuilder(hypervisor string, vmi *v1.VirtualMachineInstance, c *types.ConverterContext) *types.DomainBuilder {
	switch hypervisor {
	// Other hypervisors can be added here
	default:
		return kvm.MakeDomainBuilder(vmi, c)
	}
}
