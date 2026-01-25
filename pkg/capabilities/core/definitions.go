package capabilities

import (
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"
	v1 "kubevirt.io/api/core/v1"
)

// Capability constants - each represents a feature that may need validation or blocking
const (
	CapVsock        CapabilityKey = "domain.devices.vsock"
	CapPanicDevices CapabilityKey = "domain.devices.panicDevices"
	// ... all capabilities declared as constants
)

// Define CapVsock capability
var CapVsockDef = Capability{
	IsRequiredBy: func(vmiSpec *v1.VirtualMachineInstanceSpec) bool {
		return vmiSpec.Domain.Devices.AutoattachVSOCK != nil && *vmiSpec.Domain.Devices.AutoattachVSOCK
	},
	GetField: func(vmiSpecField *k8sfield.Path) string {
		return vmiSpecField.Child("domain").Child("devices").Child("autoattachVSOCK").String()
	},
}

// Define PanicDevices capability
var CapPanicDevicesDef = Capability{
	IsRequiredBy: func(vmiSpec *v1.VirtualMachineInstanceSpec) bool {
		return len(vmiSpec.Domain.Devices.PanicDevices) > 0
	},
	GetField: func(vmiSpecField *k8sfield.Path) string {
		return vmiSpecField.Child("domain").Child("devices").Child("panicDevices").String()
	},
}
