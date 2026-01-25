package capabilities

import (
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"
	v1 "kubevirt.io/api/core/v1"
)

type CapabilityKey string // e.g., "graphics.vga", "firmware.secureboot.uefi"
type SupportLevel int

const (
	Unregistered SupportLevel = iota // Not registered (default zero value)
	Unsupported                      // Explicitly blocked on this platform
	Experimental                     // Requires feature gate
	Deprecated                       // Supported but discouraged
)

type Platform string

const (
	Universal Platform = "" // Applies to all platforms
)

type Capability struct {
	// function to check if this capability is required by a given VMI
	IsRequiredBy func(vmi *v1.VirtualMachineInstanceSpec) bool
	// function to get the VMI field associated with this capability (for user messages)
	GetField func(vmiSpecField *k8sfield.Path) string
}

// struct to store the extent to which a given capability is supported
type CapabilitySupport struct {
	Level   SupportLevel
	Message string // User-facing explanation
	GatedBy string // Optional: feature gate name
}

func PlatformKeyFromHypervisor(hypervisor string) Platform {
	return Platform(hypervisor + "/")
}

func PlatformKeyFromArch(arch string) Platform {
	return Platform("/" + arch)
}

func PlatformKeyFromHypervisorAndArch(hypervisor, arch string) Platform {
	return Platform(hypervisor + "/" + arch)
}
