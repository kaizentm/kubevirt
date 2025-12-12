package virtruntime

import (
	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/log"

	"kubevirt.io/kubevirt/pkg/virt-handler/cgroup"
	"kubevirt.io/kubevirt/pkg/virt-handler/isolation"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

type VirtRuntime interface {
	HandleHousekeeping(vmi *v1.VirtualMachineInstance, cgroupManager cgroup.Manager, domain *api.Domain) error
	AdjustResources(podIsoDetector isolation.PodIsolationDetector, vmi *v1.VirtualMachineInstance, config *v1.KubeVirtConfiguration) error
}

func GetVirtRuntime(podIsolationDetector isolation.PodIsolationDetector) VirtRuntime {
	// TODO L1VH: Extend this to return different VirtRuntimes based on the hypervisor used
	return &KvmVirtRuntime{podIsolationDetector: podIsolationDetector, logger: log.Log.With("controller", "vm")}
}
