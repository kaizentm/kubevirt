package arch_capabilities

import (
	core "kubevirt.io/kubevirt/pkg/capabilities/core"
)

const ARM64Arch = "arm64"

func RegisterArm64Capabilities() {
	// Register capability support levels for ARM64 architecture
	core.RegisterCapability(core.CapSoundDevice, core.CapSoundDeviceDef)
	core.RegisterCapability(core.CapWatchdog, core.CapWatchdogDef)

	core.AddPlatformCapabilitySupport(core.PlatformKeyFromArch(ARM64Arch), core.CapSoundDevice, core.CapabilitySupport{
		Level:   core.Unsupported,
		Message: "Sound device is unsupported on ARM64.",
	})
	core.AddPlatformCapabilitySupport(core.PlatformKeyFromArch(ARM64Arch), core.CapWatchdog, core.CapabilitySupport{
		Level:   core.Unsupported,
		Message: "Watchdog device is unsupported on ARM64.",
	})
}
