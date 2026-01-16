package hypervisor

import (
	"kubevirt.io/kubevirt/pkg/hypervisor/kvm"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/converter/builder"
)

type DomainBuilderFactory interface {
	MakeDomainBuilder(c *builder.ConverterContext) *builder.DomainBuilder
}

func NewDomainBuilderFactory(hypervisor string) DomainBuilderFactory {
	switch hypervisor {
	// Other hypervisors can be added here
	default:
		return &kvm.KvmDomainBuilderFactory{} // Default to KVM
	}
}
