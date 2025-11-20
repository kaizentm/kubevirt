package mshv_validator

import (
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"
	v1 "kubevirt.io/api/core/v1"

	draadmitter "kubevirt.io/kubevirt/pkg/dra/admitter"
	netadmitter "kubevirt.io/kubevirt/pkg/network/admitter"
	"kubevirt.io/kubevirt/pkg/network/vmispec"
	storageadmitters "kubevirt.io/kubevirt/pkg/storage/admitters"
	"kubevirt.io/kubevirt/pkg/virt-api/webhooks"
	base_validator "kubevirt.io/kubevirt/pkg/virt-api/webhooks/validating-webhook/admitters/hypervisor/base"
	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
)

type MshvValidator struct {
	*base_validator.BaseValidator
}

func (mv *MshvValidator) ValidateVirtualMachineInstanceSpec(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	var causes []metav1.StatusCause

	causes = append(causes, mv.ValidateHostNameNotConformingToDNSLabelRules(field, spec)...)
	causes = append(causes, mv.ValidateSubdomainDNSSubdomainRules(field, spec)...)
	causes = append(causes, mv.ValidateMemoryRequestsNegativeOrNull(field, spec)...)
	causes = append(causes, mv.ValidateMemoryLimitsNegativeOrNull(field, spec)...)
	causes = append(causes, mv.ValidateHugepagesMemoryRequests(field, spec)...)
	causes = append(causes, mv.ValidateGuestMemoryLimit(field, spec, config)...)
	causes = append(causes, mv.ValidateEmulatedMachine(field, spec, config)...)
	causes = append(causes, mv.ValidateFirmwareACPI(field.Child("acpi"), spec)...)
	causes = append(causes, mv.ValidateCPURequestNotNegative(field, spec)...)
	causes = append(causes, mv.ValidateCPULimitNotNegative(field, spec)...)
	causes = append(causes, mv.ValidateCpuRequestDoesNotExceedLimit(field, spec)...)
	causes = append(causes, mv.ValidateCpuPinning(field, spec, config)...)
	causes = append(causes, mv.ValidateNUMA(field, spec, config)...)
	causes = append(causes, mv.ValidateCPUIsolatorThread(field, spec)...)
	causes = append(causes, mv.ValidateCPUFeaturePolicies(field, spec)...)
	causes = append(causes, mv.ValidateCPUHotplug(field, spec)...)
	causes = append(causes, mv.ValidateStartStrategy(field, spec)...)
	causes = append(causes, mv.ValidateRealtime(field, spec)...)
	causes = append(causes, mv.ValidateSpecAffinity(field, spec)...)
	causes = append(causes, mv.ValidateSpecTopologySpreadConstraints(field, spec)...)
	causes = append(causes, mv.ValidateArchitecture(field, spec, config)...)

	netValidator := netadmitter.NewValidator(field, spec, config)
	causes = append(causes, netValidator.Validate()...)

	causes = append(causes, draadmitter.ValidateCreation(field, spec, config)...)

	causes = append(causes, mv.ValidateBootOrder(field, spec, config)...)

	causes = append(causes, mv.ValidateInputDevices(field, spec)...)
	causes = append(causes, mv.ValidateIOThreadsPolicy(field, spec)...)
	causes = append(causes, mv.ValidateProbe(field.Child("readinessProbe"), spec.ReadinessProbe)...)
	causes = append(causes, mv.ValidateProbe(field.Child("livenessProbe"), spec.LivenessProbe)...)

	if podNetwork := vmispec.LookupPodNetwork(spec.Networks); podNetwork == nil {
		causes = appendStatusCauseForProbeNotAllowedWithNoPodNetworkPresent(field.Child("readinessProbe"), spec.ReadinessProbe, causes)
		causes = appendStatusCauseForProbeNotAllowedWithNoPodNetworkPresent(field.Child("livenessProbe"), spec.LivenessProbe, causes)
	}

	causes = append(causes, mv.ValidateDomainSpec(field.Child("domain"), &spec.Domain)...)
	causes = append(causes, mv.ValidateVolumes(field.Child("volumes"), spec.Volumes, config)...)
	causes = append(causes, storageadmitters.ValidateContainerDisks(field, spec)...)

	causes = append(causes, mv.ValidateAccessCredentials(field.Child("accessCredentials"), spec.AccessCredentials, spec.Volumes)...)

	if spec.DNSPolicy != "" {
		causes = append(causes, mv.ValidateDNSPolicy(&spec.DNSPolicy, field.Child("dnsPolicy"))...)
	}
	causes = append(causes, mv.ValidatePodDNSConfig(spec.DNSConfig, &spec.DNSPolicy, field.Child("dnsConfig"))...)
	causes = append(causes, mv.ValidateLiveMigration(field, spec, config)...)
	causes = append(causes, mv.ValidateMDEVRamFB(field, spec)...)
	causes = append(causes, mv.ValidateHostDevicesWithPassthroughEnabled(field, spec, config)...)
	causes = append(causes, mv.ValidateSoundDevices(field, spec)...)
	causes = append(causes, mv.ValidateLaunchSecurity(field, spec, config)...)
	causes = append(causes, mv.ValidateVSOCK(field, spec, config)...)
	causes = append(causes, mv.ValidatePersistentReservation(field, spec, config)...)
	causes = append(causes, mv.ValidateDownwardMetrics(field, spec, config)...)
	causes = append(causes, mv.ValidateFilesystemsWithVirtIOFSEnabled(field, spec, config)...)
	causes = append(causes, mv.ValidateVideoConfig(field, spec, config)...)
	causes = append(causes, mv.ValidatePanicDevices(field, spec, config)...)

	// We only want to validate that volumes are mapped to disks or filesystems during VMI admittance, thus this logic is seperated from the above call that is shared with the VM admitter.
	causes = append(causes, mv.ValidateVirtualMachineInstanceSpecVolumeDisks(k8sfield.NewPath("spec"), spec)...)
	causes = append(causes, mv.ValidateVirtualMachineInstanceMandatoryFields(k8sfield.NewPath("spec"), spec)...)

	// TODO Why is hyperv validation logic in a separate location?
	causes = append(causes, webhooks.ValidateVirtualMachineInstanceHyperv(k8sfield.NewPath("spec").Child("domain").Child("features").Child("hyperv"), spec)...)
	causes = append(causes, mv.ValidateVirtualMachineInstancePerArch(k8sfield.NewPath("spec"), spec, config)...)

	return causes
}

func (mv *MshvValidator) ValidateHotplug(oldVmi *v1.VirtualMachineInstance, newVmi *v1.VirtualMachineInstance, cc *virtconfig.ClusterConfig) *admissionv1.AdmissionResponse {
	return mv.BaseValidator.ValidateHotplug(oldVmi, newVmi, cc)
}

func appendStatusCauseForProbeNotAllowedWithNoPodNetworkPresent(field *k8sfield.Path, probe *v1.Probe, causes []metav1.StatusCause) []metav1.StatusCause {
	if probe == nil {
		return causes
	}

	if probe.HTTPGet != nil {
		causes = append(causes, podNetworkRequiredStatusCause(field.Child("httpGet")))
	}

	if probe.TCPSocket != nil {
		causes = append(causes, podNetworkRequiredStatusCause(field.Child("tcpSocket")))
	}
	return causes
}

func podNetworkRequiredStatusCause(field *k8sfield.Path) metav1.StatusCause {
	return metav1.StatusCause{
		Type:    metav1.CauseTypeFieldValueInvalid,
		Message: fmt.Sprintf("%s is only allowed if the Pod Network is attached", field.String()),
		Field:   field.String(),
	}
}
