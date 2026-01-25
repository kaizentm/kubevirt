package capabilities

import (
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"
	v1 "kubevirt.io/api/core/v1"
)

// Capability constants - each represents a feature that may need validation or blocking
const (
	CapVsock                 CapabilityKey = "domain.devices.vsock"
	CapPanicDevices          CapabilityKey = "domain.devices.panicDevices"
	CapPersistentReservation CapabilityKey = "domain.devices.disks.luns.reservation"
	CapVideoConfig           CapabilityKey = "domain.devices.video"
	CapHostDevicePassthrough CapabilityKey = "domain.devices.hostDevices.passthrough"
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

var CapPersistentReservationDef = Capability{
	IsRequiredBy: func(vmiSpec *v1.VirtualMachineInstanceSpec) bool {
		for _, disk := range vmiSpec.Domain.Devices.Disks {
			if disk.DiskDevice.LUN != nil && disk.DiskDevice.LUN.Reservation {
				return true
			}
		}
		return false
	},
	GetField: func(vmiSpecField *k8sfield.Path) string {
		return vmiSpecField.Child("domain", "devices", "disks", "luns", "reservation").String()
	},
}

var CapVideoConfigDef = Capability{
	IsRequiredBy: func(vmiSpec *v1.VirtualMachineInstanceSpec) bool {
		return vmiSpec.Domain.Devices.Video != nil
	},
	GetField: func(vmiSpecField *k8sfield.Path) string {
		return vmiSpecField.Child("video").String()
	},
}

var CapHostDevicePassthroughDef = Capability{
	IsRequiredBy: func(vmiSpec *v1.VirtualMachineInstanceSpec) bool {
		return vmiSpec.Domain.Devices.HostDevices != nil
	},
	GetField: func(vmiSpecField *k8sfield.Path) string {
		return vmiSpecField.Child("HostDevices").String()
	},
}
