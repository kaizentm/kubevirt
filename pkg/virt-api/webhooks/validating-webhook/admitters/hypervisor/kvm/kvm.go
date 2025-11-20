package kvm_validator

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

type KvmValidator struct {
	*base_validator.BaseValidator
}

func (kv *KvmValidator) ValidateVirtualMachineSpec(field *k8sfield.Path, spec *v1.VirtualMachineSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	return []metav1.StatusCause{}
}

func (kv *KvmValidator) ValidateVirtualMachineInstanceSpec(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	var causes []metav1.StatusCause

	causes = append(causes, kv.ValidateHostNameNotConformingToDNSLabelRules(field, spec)...)
	causes = append(causes, kv.ValidateSubdomainDNSSubdomainRules(field, spec)...)
	causes = append(causes, kv.ValidateMemoryRequestsNegativeOrNull(field, spec)...)
	causes = append(causes, kv.ValidateMemoryLimitsNegativeOrNull(field, spec)...)
	causes = append(causes, kv.ValidateHugepagesMemoryRequests(field, spec)...)
	causes = append(causes, kv.ValidateGuestMemoryLimit(field, spec, config)...)
	causes = append(causes, kv.ValidateEmulatedMachine(field, spec, config)...)
	causes = append(causes, kv.ValidateFirmwareACPI(field.Child("acpi"), spec)...)
	causes = append(causes, kv.ValidateCPURequestNotNegative(field, spec)...)
	causes = append(causes, kv.ValidateCPULimitNotNegative(field, spec)...)
	causes = append(causes, kv.ValidateCpuRequestDoesNotExceedLimit(field, spec)...)
	causes = append(causes, kv.ValidateCpuPinning(field, spec, config)...)
	causes = append(causes, kv.ValidateNUMA(field, spec, config)...)
	causes = append(causes, kv.ValidateCPUIsolatorThread(field, spec)...)
	causes = append(causes, kv.ValidateCPUFeaturePolicies(field, spec)...)
	causes = append(causes, kv.ValidateCPUHotplug(field, spec)...)
	causes = append(causes, kv.ValidateStartStrategy(field, spec)...)
	causes = append(causes, kv.ValidateRealtime(field, spec)...)
	causes = append(causes, kv.ValidateSpecAffinity(field, spec)...)
	causes = append(causes, kv.ValidateSpecTopologySpreadConstraints(field, spec)...)
	causes = append(causes, kv.ValidateArchitecture(field, spec, config)...)

	netValidator := netadmitter.NewValidator(field, spec, config)
	causes = append(causes, netValidator.Validate()...)

	causes = append(causes, draadmitter.ValidateCreation(field, spec, config)...)

	causes = append(causes, kv.ValidateBootOrder(field, spec, config)...)

	causes = append(causes, kv.ValidateInputDevices(field, spec)...)
	causes = append(causes, kv.ValidateIOThreadsPolicy(field, spec)...)
	causes = append(causes, kv.ValidateProbe(field.Child("readinessProbe"), spec.ReadinessProbe)...)
	causes = append(causes, kv.ValidateProbe(field.Child("livenessProbe"), spec.LivenessProbe)...)

	if podNetwork := vmispec.LookupPodNetwork(spec.Networks); podNetwork == nil {
		causes = appendStatusCauseForProbeNotAllowedWithNoPodNetworkPresent(field.Child("readinessProbe"), spec.ReadinessProbe, causes)
		causes = appendStatusCauseForProbeNotAllowedWithNoPodNetworkPresent(field.Child("livenessProbe"), spec.LivenessProbe, causes)
	}

	causes = append(causes, kv.ValidateDomainSpec(field.Child("domain"), &spec.Domain)...)
	causes = append(causes, kv.ValidateVolumes(field.Child("volumes"), spec.Volumes, config)...)
	causes = append(causes, storageadmitters.ValidateContainerDisks(field, spec)...)

	causes = append(causes, kv.ValidateAccessCredentials(field.Child("accessCredentials"), spec.AccessCredentials, spec.Volumes)...)

	if spec.DNSPolicy != "" {
		causes = append(causes, kv.ValidateDNSPolicy(&spec.DNSPolicy, field.Child("dnsPolicy"))...)
	}
	causes = append(causes, kv.ValidatePodDNSConfig(spec.DNSConfig, &spec.DNSPolicy, field.Child("dnsConfig"))...)
	causes = append(causes, kv.ValidateLiveMigration(field, spec, config)...)
	causes = append(causes, kv.ValidateMDEVRamFB(field, spec)...)
	causes = append(causes, kv.ValidateHostDevicesWithPassthroughEnabled(field, spec, config)...)
	causes = append(causes, kv.ValidateSoundDevices(field, spec)...)
	causes = append(causes, kv.ValidateLaunchSecurity(field, spec, config)...)
	causes = append(causes, kv.ValidateVSOCK(field, spec, config)...)
	causes = append(causes, kv.ValidatePersistentReservation(field, spec, config)...)
	causes = append(causes, kv.ValidateDownwardMetrics(field, spec, config)...)
	causes = append(causes, kv.ValidateFilesystemsWithVirtIOFSEnabled(field, spec, config)...)
	causes = append(causes, kv.ValidateVideoConfig(field, spec, config)...)
	causes = append(causes, kv.ValidatePanicDevices(field, spec, config)...)

	// We only want to validate that volumes are mapped to disks or filesystems during VMI admittance, thus this logic is seperated from the above call that is shared with the VM admitter.
	causes = append(causes, kv.ValidateVirtualMachineInstanceSpecVolumeDisks(k8sfield.NewPath("spec"), spec)...)
	causes = append(causes, kv.ValidateVirtualMachineInstanceMandatoryFields(k8sfield.NewPath("spec"), spec)...)

	// TODO Why is hyperv validation logic in a separate location?
	causes = append(causes, webhooks.ValidateVirtualMachineInstanceHyperv(k8sfield.NewPath("spec").Child("domain").Child("features").Child("hyperv"), spec)...)
	causes = append(causes, kv.ValidateVirtualMachineInstancePerArch(k8sfield.NewPath("spec"), spec)...)

	return causes
}

func (kv *KvmValidator) ValidateHotplug(oldVmi *v1.VirtualMachineInstance, newVmi *v1.VirtualMachineInstance, cc *virtconfig.ClusterConfig) *admissionv1.AdmissionResponse {
	return kv.BaseValidator.ValidateHotplug(oldVmi, newVmi, cc)
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
