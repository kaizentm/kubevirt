package capabilities

import (
	arch_capabilities "kubevirt.io/kubevirt/pkg/capabilities/arch"
	core "kubevirt.io/kubevirt/pkg/capabilities/core"
	hypervisor_capabilities "kubevirt.io/kubevirt/pkg/capabilities/hypervisor"
	"kubevirt.io/kubevirt/pkg/virt-config/featuregate"
)

// Function to register all capabilities universal to KubeVirt
func RegisterUniversalCapabilities() {
	// Associate capability keys with their definitions
	core.RegisterCapability(core.CapVsock, core.CapVsockDef)
	core.RegisterCapability(core.CapPanicDevices, core.CapPanicDevicesDef)
	core.RegisterCapability(core.CapPersistentReservation, core.CapPersistentReservationDef)
	core.RegisterCapability(core.CapVideoConfig, core.CapVideoConfigDef)
	core.RegisterCapability(core.CapHostDevicePassthrough, core.CapHostDevicePassthroughDef)

	// Declare cross-platform support level for capabilities
	core.AddPlatformCapabilitySupport(core.Universal, core.CapVsock, core.CapabilitySupport{
		Level:   core.Experimental,
		Message: "Vsock support is experimental on this platform.",
		GatedBy: featuregate.VSOCKGate,
	})
	core.AddPlatformCapabilitySupport(core.Universal, core.CapPanicDevices, core.CapabilitySupport{
		Level:   core.Experimental,
		Message: "PanicDevices experimental on this platform.",
		GatedBy: featuregate.PanicDevicesGate,
	})
	core.AddPlatformCapabilitySupport(core.Universal, core.CapPersistentReservation, core.CapabilitySupport{
		Level:   core.Experimental,
		Message: "Persistent Reservation support is experimental on this platform.",
		GatedBy: featuregate.PersistentReservation,
	})
	core.AddPlatformCapabilitySupport(core.Universal, core.CapVideoConfig, core.CapabilitySupport{
		Level:   core.Experimental,
		Message: "VideoConfig support is experimental on this platform.",
		GatedBy: featuregate.VideoConfig,
	})
	core.AddPlatformCapabilitySupport(core.Universal, core.CapHostDevicePassthrough, core.CapabilitySupport{
		Level:   core.Experimental,
		Message: "HostDevicePassthrough support is experimental on this platform.",
		GatedBy: featuregate.HostDevicesGate,
	})
}

// Function to register all capabilities and their support levels
func Init() {
	RegisterUniversalCapabilities()

	// Declare platform-specific capability support levels

	// Declare hypervisor-specific capability support levels
	hypervisor_capabilities.RegisterKvmCapabilities()
	hypervisor_capabilities.RegisterMshvCapabilities()

	// Declare architecture-specific capability support levels
	arch_capabilities.RegisterAmd64Capabilities()
	arch_capabilities.RegisterArm64Capabilities()
	arch_capabilities.RegisterS390xCapabilities()
}

func Reset() {
	core.Reset()
}
