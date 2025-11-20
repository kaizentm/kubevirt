/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright The KubeVirt Authors.
 *
 */

package admitters

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"kubevirt.io/kubevirt/pkg/storage/utils"

	admissionv1 "k8s.io/api/admission/v1"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"

	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/downwardmetrics"
	draadmitter "kubevirt.io/kubevirt/pkg/dra/admitter"
	"kubevirt.io/kubevirt/pkg/hooks"
	netadmitter "kubevirt.io/kubevirt/pkg/network/admitter"
	"kubevirt.io/kubevirt/pkg/network/vmispec"
	storageadmitters "kubevirt.io/kubevirt/pkg/storage/admitters"
	"kubevirt.io/kubevirt/pkg/storage/types"
	webhookutils "kubevirt.io/kubevirt/pkg/util/webhooks"
	"kubevirt.io/kubevirt/pkg/virt-api/webhooks"
	hypervisor_validator "kubevirt.io/kubevirt/pkg/virt-api/webhooks/validating-webhook/admitters/hypervisor"
	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
	"kubevirt.io/kubevirt/pkg/virt-config/featuregate"
)

const requiredFieldFmt = "%s is a required field"

var restrictedVmiLabels = map[string]bool{
	v1.CreatedByLabel:               true,
	v1.MigrationJobLabel:            true,
	v1.NodeNameLabel:                true,
	v1.MigrationTargetNodeNameLabel: true,
	v1.NodeSchedulable:              true,
	v1.InstallStrategyLabel:         true,
}

// SpecValidator validates the given VMI spec
type SpecValidator func(*k8sfield.Path, *v1.VirtualMachineInstanceSpec, *virtconfig.ClusterConfig) []metav1.StatusCause

type VMICreateAdmitter struct {
	ClusterConfig           *virtconfig.ClusterConfig
	SpecValidators          []SpecValidator
	KubeVirtServiceAccounts map[string]struct{}
}

func (admitter *VMICreateAdmitter) Admit(_ context.Context, ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	if resp := webhookutils.ValidateSchema(v1.VirtualMachineInstanceGroupVersionKind, ar.Request.Object.Raw); resp != nil {
		return resp
	}

	vmi, _, err := webhookutils.GetVMIFromAdmissionReview(ar)
	if err != nil {
		return webhookutils.ToAdmissionResponseError(err)
	}

	var causes []metav1.StatusCause
	clusterCfg := admitter.ClusterConfig.GetConfig()

	// TODO Use the below validate feature gates to work on the capability mechanism VEP in KubeVirt
	if devCfg := clusterCfg.DeveloperConfiguration; devCfg != nil {
		causes = append(causes, featuregate.ValidateFeatureGates(devCfg.FeatureGates, &vmi.Spec)...)
	}

	for _, validateSpec := range admitter.SpecValidators {
		causes = append(causes, validateSpec(k8sfield.NewPath("spec"), &vmi.Spec, admitter.ClusterConfig)...)
	}

	hypervisor := admitter.ClusterConfig.GetHypervisor()
	validator := hypervisor_validator.NewValidator(hypervisor.Name)
	causes = append(causes, validator.ValidateVirtualMachineInstanceSpec(k8sfield.NewPath("spec"), &vmi.Spec, admitter.ClusterConfig)...)

	/*	MOVE THE FOLLOWING TO THE VALIDATOR INTERFACE'S BASE IMPLEMENTATION
		causes = append(causes, ValidateVirtualMachineInstanceSpec(k8sfield.NewPath("spec"), &vmi.Spec, admitter.ClusterConfig)...)
		// We only want to validate that volumes are mapped to disks or filesystems during VMI admittance, thus this logic is seperated from the above call that is shared with the VM admitter.
		causes = append(causes, validateVirtualMachineInstanceSpecVolumeDisks(k8sfield.NewPath("spec"), &vmi.Spec)...)
		causes = append(causes, ValidateVirtualMachineInstanceMandatoryFields(k8sfield.NewPath("spec"), &vmi.Spec)...)
		causes = append(causes, webhooks.ValidateVirtualMachineInstanceHyperv(k8sfield.NewPath("spec").Child("domain").Child("features").Child("hyperv"), &vmi.Spec)...)
		causes = append(causes, ValidateVirtualMachineInstancePerArch(k8sfield.NewPath("spec"), &vmi.Spec)...)
	*/

	_, isKubeVirtServiceAccount := admitter.KubeVirtServiceAccounts[ar.Request.UserInfo.Username]
	causes = append(causes, ValidateVirtualMachineInstanceMetadata(k8sfield.NewPath("metadata"), &vmi.ObjectMeta, admitter.ClusterConfig, isKubeVirtServiceAccount)...)

	if len(causes) > 0 {
		return webhookutils.ToAdmissionResponse(causes)
	}

	return &admissionv1.AdmissionResponse{
		Allowed:  true,
		Warnings: warnDeprecatedAPIs(&vmi.Spec, admitter.ClusterConfig),
	}
}

func warnDeprecatedAPIs(spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []string {
	var warnings []string
	for _, fg := range config.GetConfig().DeveloperConfiguration.FeatureGates {
		deprecatedFeature := featuregate.FeatureGateInfo(fg)
		if deprecatedFeature != nil && deprecatedFeature.State == featuregate.Deprecated && deprecatedFeature.VmiSpecUsed != nil {
			if used := deprecatedFeature.VmiSpecUsed(spec); used {
				warnings = append(warnings, deprecatedFeature.Message)
			}
		}
	}
	return warnings
}

func ValidateVirtualMachineInstancePerArch(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause
	arch := spec.Architecture

	switch arch {
	case "amd64":
		causes = append(causes, webhooks.ValidateVirtualMachineInstanceAmd64Setting(field, spec)...)
	case "s390x":
		causes = append(causes, webhooks.ValidateVirtualMachineInstanceS390XSetting(field, spec)...)
	case "arm64":
		causes = append(causes, webhooks.ValidateVirtualMachineInstanceArm64Setting(field, spec)...)
	default:
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("unsupported architecture: %s", arch),
			Field:   field.Child("architecture").String(),
		})
	}

	return causes
}

func ValidateVirtualMachineInstanceSpec(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	var causes []metav1.StatusCause

	causes = append(causes, validateHostNameNotConformingToDNSLabelRules(field, spec)...)
	causes = append(causes, validateSubdomainDNSSubdomainRules(field, spec)...)
	causes = append(causes, validateMemoryRequestsNegativeOrNull(field, spec)...)
	causes = append(causes, validateMemoryLimitsNegativeOrNull(field, spec)...)
	causes = append(causes, validateHugepagesMemoryRequests(field, spec)...)
	causes = append(causes, validateGuestMemoryLimit(field, spec, config)...)
	causes = append(causes, validateEmulatedMachine(field, spec, config)...)
	causes = append(causes, validateFirmwareACPI(field.Child("acpi"), spec)...)
	causes = append(causes, validateCPURequestNotNegative(field, spec)...)
	causes = append(causes, validateCPULimitNotNegative(field, spec)...)
	causes = append(causes, validateCpuRequestDoesNotExceedLimit(field, spec)...)
	causes = append(causes, validateCpuPinning(field, spec, config)...)
	causes = append(causes, validateNUMA(field, spec, config)...)
	causes = append(causes, validateCPUIsolatorThread(field, spec)...)
	causes = append(causes, validateCPUFeaturePolicies(field, spec)...)
	causes = append(causes, validateCPUHotplug(field, spec)...)
	causes = append(causes, validateStartStrategy(field, spec)...)
	causes = append(causes, validateRealtime(field, spec)...)
	causes = append(causes, validateSpecAffinity(field, spec)...)
	causes = append(causes, validateSpecTopologySpreadConstraints(field, spec)...)
	causes = append(causes, validateArchitecture(field, spec, config)...)

	netValidator := netadmitter.NewValidator(field, spec, config)
	causes = append(causes, netValidator.Validate()...)

	causes = append(causes, draadmitter.ValidateCreation(field, spec, config)...)

	causes = append(causes, validateBootOrder(field, spec, config)...)

	causes = append(causes, validateInputDevices(field, spec)...)
	causes = append(causes, validateIOThreadsPolicy(field, spec)...)
	causes = append(causes, validateProbe(field.Child("readinessProbe"), spec.ReadinessProbe)...)
	causes = append(causes, validateProbe(field.Child("livenessProbe"), spec.LivenessProbe)...)

	if podNetwork := vmispec.LookupPodNetwork(spec.Networks); podNetwork == nil {
		causes = appendStatusCauseForProbeNotAllowedWithNoPodNetworkPresent(field.Child("readinessProbe"), spec.ReadinessProbe, causes)
		causes = appendStatusCauseForProbeNotAllowedWithNoPodNetworkPresent(field.Child("livenessProbe"), spec.LivenessProbe, causes)
	}

	causes = append(causes, validateDomainSpec(field.Child("domain"), &spec.Domain)...)
	causes = append(causes, validateVolumes(field.Child("volumes"), spec.Volumes, config)...)
	causes = append(causes, storageadmitters.ValidateContainerDisks(field, spec)...)

	causes = append(causes, validateAccessCredentials(field.Child("accessCredentials"), spec.AccessCredentials, spec.Volumes)...)

	if spec.DNSPolicy != "" {
		causes = append(causes, validateDNSPolicy(&spec.DNSPolicy, field.Child("dnsPolicy"))...)
	}
	causes = append(causes, validatePodDNSConfig(spec.DNSConfig, &spec.DNSPolicy, field.Child("dnsConfig"))...)
	causes = append(causes, validateLiveMigration(field, spec, config)...)
	causes = append(causes, validateMDEVRamFB(field, spec)...)
	causes = append(causes, validateHostDevicesWithPassthroughEnabled(field, spec, config)...)
	causes = append(causes, validateSoundDevices(field, spec)...)
	causes = append(causes, validateLaunchSecurity(field, spec, config)...)
	causes = append(causes, validateVSOCK(field, spec, config)...)
	causes = append(causes, validatePersistentReservation(field, spec, config)...)
	causes = append(causes, validateDownwardMetrics(field, spec, config)...)
	causes = append(causes, validateFilesystemsWithVirtIOFSEnabled(field, spec, config)...)
	causes = append(causes, validateVideoConfig(field, spec, config)...)
	causes = append(causes, validatePanicDevices(field, spec, config)...)

	return causes
}

func validateFilesystemsWithVirtIOFSEnabled(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) (causes []metav1.StatusCause) {
	if spec.Domain.Devices.Filesystems == nil {
		return causes
	}

	volumes := types.GetVolumesByName(spec)

	for _, fs := range spec.Domain.Devices.Filesystems {
		volume, ok := volumes[fs.Name]
		if !ok {
			continue
		}

		switch {
		case utils.IsConfigVolume(volume) && (!config.VirtiofsConfigVolumesEnabled() && !config.OldVirtiofsEnabled()):
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: "virtiofs is not allowed: virtiofs feature gate is not enabled for config volumes",
				Field:   field.Child("domain", "devices", "filesystems").String(),
			})
		case utils.IsStorageVolume(volume) && (!config.VirtiofsStorageEnabled() && !config.OldVirtiofsEnabled()):
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: "virtiofs is not allowed: virtiofs feature gate is not enabled for PVC",
				Field:   field.Child("domain", "devices", "filesystems").String(),
			})
		}
	}

	return causes
}

func validateDownwardMetrics(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	var causes []metav1.StatusCause

	// Check if serial and feature gate is enabled
	if downwardmetrics.HasDevice(spec) && !config.DownwardMetricsEnabled() {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: "downwardMetrics virtio serial is not allowed: DownwardMetrics feature gate is not enabled",
			Field:   field.Child("domain", "devices", "downwardMetrics").String(),
		})
	}

	return causes
}

func validateVirtualMachineInstanceSpecVolumeDisks(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause

	diskAndFilesystemNames := make(map[string]struct{})

	for _, disk := range spec.Domain.Devices.Disks {
		diskAndFilesystemNames[disk.Name] = struct{}{}
	}

	for _, fs := range spec.Domain.Devices.Filesystems {
		diskAndFilesystemNames[fs.Name] = struct{}{}
	}

	// Validate that volumes match disks and filesystems correctly
	for idx, volume := range spec.Volumes {
		if volume.MemoryDump != nil {
			continue
		}
		if _, matchingDiskExists := diskAndFilesystemNames[volume.Name]; !matchingDiskExists {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: fmt.Sprintf(nameOfTypeNotFoundMessagePattern, field.Child("domain", "volumes").Index(idx).Child("name").String(), volume.Name),
				Field:   field.Child("domain", "volumes").Index(idx).Child("name").String(),
			})
		}
	}
	return causes
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

func isValidEvictionStrategy(evictionStrategy *v1.EvictionStrategy) bool {
	return evictionStrategy == nil ||
		*evictionStrategy == v1.EvictionStrategyLiveMigrate ||
		*evictionStrategy == v1.EvictionStrategyLiveMigrateIfPossible ||
		*evictionStrategy == v1.EvictionStrategyNone ||
		*evictionStrategy == v1.EvictionStrategyExternal
}

func validateLiveMigration(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	var causes []metav1.StatusCause
	evictionStrategy := config.GetConfig().EvictionStrategy

	if spec.EvictionStrategy != nil {
		evictionStrategy = spec.EvictionStrategy
	}
	if !isValidEvictionStrategy(evictionStrategy) {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s is set with an unrecognized option: %s", field.Child("evictionStrategy").String(), *spec.EvictionStrategy),
			Field:   field.Child("evictionStrategy").String(),
		})
	}
	return causes
}

func validateEmulatedMachine(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	var causes []metav1.StatusCause
	if machine := spec.Domain.Machine; machine != nil && len(machine.Type) > 0 {
		supportedMachines := config.GetEmulatedMachines(spec.Architecture)
		var match = false
		for _, val := range supportedMachines {
			// The pattern are hardcoded, so this should not throw an error
			if ok, _ := filepath.Match(val, machine.Type); ok {
				match = true
				break
			}
		}
		if !match {
			causes = append(causes, metav1.StatusCause{
				Type: metav1.CauseTypeFieldValueInvalid,
				Message: fmt.Sprintf("%s is not supported: %s (allowed values: %v)",
					field.Child("domain", "machine", "type").String(),
					machine.Type,
					supportedMachines,
				),
				Field: field.Child("domain", "machine", "type").String(),
			})
		}
	}
	return causes
}

func validateGuestMemoryLimit(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	var causes []metav1.StatusCause
	if config.IsVMRolloutStrategyLiveUpdate() {
		return causes
	}
	if spec.Domain.Memory == nil || spec.Domain.Memory.Guest == nil {
		return causes
	}
	limits := spec.Domain.Resources.Limits.Memory().Value()
	guest := spec.Domain.Memory.Guest.Value()
	if limits < guest && limits != 0 {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s '%s' must be equal to or less than the memory limit %s '%s'",
				field.Child("domain", "memory", "guest").String(),
				spec.Domain.Memory.Guest,
				field.Child("domain", "resources", "limits", "memory").String(),
				spec.Domain.Resources.Limits.Memory(),
			),
			Field: field.Child("domain", "memory", "guest").String(),
		})
	}
	return causes
}

func validateHugepagesMemoryRequests(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause
	if spec.Domain.Memory == nil || spec.Domain.Memory.Hugepages == nil {
		return causes
	}
	hugepagesSize, err := resource.ParseQuantity(spec.Domain.Memory.Hugepages.PageSize)
	if err != nil {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s '%s': %s",
				field.Child("domain", "hugepages", "size").String(),
				spec.Domain.Memory.Hugepages.PageSize,
				resource.ErrFormatWrong,
			),
			Field: field.Child("domain", "hugepages", "size").String(),
		})
		return causes
	}
	vmMemory := spec.Domain.Resources.Requests.Memory().Value()
	if vmMemory < hugepagesSize.Value() {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s '%s' must be equal to or larger than page size %s '%s'",
				field.Child("domain", "resources", "requests", "memory").String(),
				spec.Domain.Resources.Requests.Memory(),
				field.Child("domain", "hugepages", "size").String(),
				spec.Domain.Memory.Hugepages.PageSize,
			),
			Field: field.Child("domain", "resources", "requests", "memory").String(),
		})
	} else if vmMemory%hugepagesSize.Value() != 0 {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s '%s' is not a multiple of the page size %s '%s'",
				field.Child("domain", "resources", "requests", "memory").String(),
				spec.Domain.Resources.Requests.Memory(),
				field.Child("domain", "hugepages", "size").String(),
				spec.Domain.Memory.Hugepages.PageSize,
			),
			Field: field.Child("domain", "resources", "requests", "memory").String(),
		})
	}

	return causes
}

func validateMemoryLimitsNegativeOrNull(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause
	if spec.Domain.Resources.Limits.Memory().Value() < 0 {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf(valueMustBePositiveMessagePattern, field.Child("domain", "resources", "limits", "memory").String(),
				spec.Domain.Resources.Limits.Memory()),
			Field: field.Child("domain", "resources", "limits", "memory").String(),
		})
	}

	if spec.Domain.Resources.Limits.Memory().Value() > 0 &&
		spec.Domain.Resources.Requests.Memory().Value() > spec.Domain.Resources.Limits.Memory().Value() {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s '%s' is greater than %s '%s'", field.Child("domain", "resources", "requests", "memory").String(),
				spec.Domain.Resources.Requests.Memory(),
				field.Child("domain", "resources", "limits", "memory").String(),
				spec.Domain.Resources.Limits.Memory()),
			Field: field.Child("domain", "resources", "requests", "memory").String(),
		})
	}
	return causes
}

func validateMemoryRequestsNegativeOrNull(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause
	if spec.Domain.Resources.Requests.Memory().Value() < 0 {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf(valueMustBePositiveMessagePattern, field.Child("domain", "resources", "requests", "memory").String(),
				spec.Domain.Resources.Requests.Memory()),
			Field: field.Child("domain", "resources", "requests", "memory").String(),
		})
	} else if spec.Domain.Resources.Requests.Memory().Value() > 0 && spec.Domain.Resources.Requests.Memory().Cmp(resource.MustParse("1M")) < 0 {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s '%s': must be greater than or equal to 1M.", field.Child("domain", "resources", "requests", "memory").String(),
				spec.Domain.Resources.Requests.Memory()),
			Field: field.Child("domain", "resources", "requests", "memory").String(),
		})
	}
	return causes
}

func validateSubdomainDNSSubdomainRules(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause
	if spec.Subdomain == "" {
		return causes
	}
	if errors := validation.IsDNS1123Subdomain(spec.Subdomain); len(errors) != 0 {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s does not conform to the kubernetes DNS_SUBDOMAIN rules : %s",
				field.Child("subdomain").String(), strings.Join(errors, ", ")),
			Field: field.Child("subdomain").String(),
		})
	}

	return causes
}

func validateHostNameNotConformingToDNSLabelRules(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause
	if spec.Hostname == "" {
		return causes
	}
	if errors := validation.IsDNS1123Label(spec.Hostname); len(errors) != 0 {
		causes = appendNewStatusCauseForHostNameNotConformingToDNSLabelRules(field, causes, errors)
	}

	return causes
}

func validateRealtime(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause
	if spec.Domain.CPU != nil && spec.Domain.CPU.Realtime != nil {
		causes = append(causes, validateCPURealtime(field, spec)...)
		causes = append(causes, validateMemoryRealtime(field, spec)...)
	}
	return causes
}

// ValidateVirtualMachineInstanceMandatoryFields should be invoked after all defaults and presets are applied.
// It is only meant to be used for VMI reviews, not if they are templates on other objects
func ValidateVirtualMachineInstanceMandatoryFields(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause

	requests := spec.Domain.Resources.Requests.Memory().Value()
	if requests != 0 {
		return causes
	}

	if spec.Domain.Memory == nil || spec.Domain.Memory != nil &&
		spec.Domain.Memory.Guest == nil && spec.Domain.Memory.Hugepages == nil {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueRequired,
			Message: fmt.Sprintf("no memory requested, at least one of '%s', '%s' or '%s' must be set",
				field.Child("domain", "memory", "guest").String(),
				field.Child("domain", "memory", "hugepages", "size").String(),
				field.Child("domain", "resources", "requests", "memory").String()),
		})
	}
	return causes
}

func ValidateVirtualMachineInstanceMetadata(field *k8sfield.Path, metadata *metav1.ObjectMeta, config *virtconfig.ClusterConfig, isKubeVirtServiceAccount bool) []metav1.StatusCause {
	var causes []metav1.StatusCause
	annotations := metadata.Annotations
	labels := metadata.Labels
	// Validate kubevirt.io labels presence. Restricted labels allowed
	// to be created only by known service accounts
	if !isKubeVirtServiceAccount {
		if len(filterKubevirtLabels(labels)) > 0 {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueNotSupported,
				Message: "creation of the following reserved kubevirt.io/ labels on a VMI object is prohibited",
				Field:   field.Child("labels").String(),
			})
		}
	}

	// Validate ignition feature gate if set when the corresponding annotation is found
	if annotations[v1.IgnitionAnnotation] != "" && !config.IgnitionEnabled() {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("ExperimentalIgnitionSupport feature gate is not enabled in kubevirt-config, invalid entry %s",
				field.Child("annotations").Child(v1.IgnitionAnnotation).String()),
			Field: field.Child("annotations").String(),
		})
	}

	// Validate sidecar feature gate if set when the corresponding annotation is found
	if annotations[hooks.HookSidecarListAnnotationName] != "" && !config.SidecarEnabled() {
		causes = append(causes, metav1.StatusCause{
			Type: metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("sidecar feature gate is not enabled in kubevirt-config, invalid entry %s",
				field.Child("annotations", hooks.HookSidecarListAnnotationName).String()),
			Field: field.Child("annotations").String(),
		})
	}

	return causes
}

// Copied from kubernetes/pkg/apis/core/validation/validation.go
func validatePodDNSConfig(dnsConfig *k8sv1.PodDNSConfig, dnsPolicy *k8sv1.DNSPolicy, field *k8sfield.Path) []metav1.StatusCause {
	var causes []metav1.StatusCause

	// Validate DNSNone case. Must provide at least one DNS name server.
	if dnsPolicy != nil && *dnsPolicy == k8sv1.DNSNone {
		if dnsConfig == nil {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueRequired,
				Message: fmt.Sprintf("must provide `dnsConfig` when `dnsPolicy` is %s", k8sv1.DNSNone),
				Field:   field.String(),
			})
			return causes
		}
		if len(dnsConfig.Nameservers) == 0 {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueRequired,
				Message: fmt.Sprintf("must provide at least one DNS nameserver when `dnsPolicy` is %s", k8sv1.DNSNone),
				Field:   "nameservers",
			})
			return causes
		}
	}

	if dnsConfig == nil {
		return causes
	}
	// Validate nameservers.
	if len(dnsConfig.Nameservers) > maxDNSNameservers {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("must not have more than %v nameservers: %s", maxDNSNameservers, dnsConfig.Nameservers),
			Field:   "nameservers",
		})
	}
	for _, ns := range dnsConfig.Nameservers {
		if ip := net.ParseIP(ns); ip == nil {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: fmt.Sprintf("must be valid IP address: %s", ns),
				Field:   "nameservers",
			})
		}
	}
	// Validate searches.
	if len(dnsConfig.Searches) > maxDNSSearchPaths {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("must not have more than %v search paths", maxDNSSearchPaths),
			Field:   "searchDomains",
		})
	}
	// Include the space between search paths.
	if len(strings.Join(dnsConfig.Searches, " ")) > maxDNSSearchListChars {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("must not have more than %v characters (including spaces) in the search list", maxDNSSearchListChars),
			Field:   "searchDomains",
		})
	}
	for _, search := range dnsConfig.Searches {
		for _, msg := range validation.IsDNS1123Subdomain(search) {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: fmt.Sprintf("%v", msg),
				Field:   "searchDomains",
			})
		}
	}
	// Validate options.
	for _, option := range dnsConfig.Options {
		if len(option.Name) == 0 {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: fmt.Sprintf("Option.Name must not be empty"),
				Field:   "options",
			})
		}
	}

	return causes
}

// Copied from kubernetes/pkg/apis/core/validation/validation.go
func validateDNSPolicy(dnsPolicy *k8sv1.DNSPolicy, field *k8sfield.Path) []metav1.StatusCause {
	var causes []metav1.StatusCause
	switch *dnsPolicy {
	case k8sv1.DNSClusterFirstWithHostNet, k8sv1.DNSClusterFirst, k8sv1.DNSDefault, k8sv1.DNSNone, "":
	default:
		validValues := []string{string(k8sv1.DNSClusterFirstWithHostNet), string(k8sv1.DNSClusterFirst), string(k8sv1.DNSDefault), string(k8sv1.DNSNone), ""}
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueNotSupported,
			Message: fmt.Sprintf("DNSPolicy: %s is not supported, valid values: %s", *dnsPolicy, validValues),
			Field:   field.String(),
		})
	}
	return causes
}

func validateBootloader(field *k8sfield.Path, bootloader *v1.Bootloader) []metav1.StatusCause {
	var causes []metav1.StatusCause

	if bootloader != nil && bootloader.EFI != nil && bootloader.BIOS != nil {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s has both EFI and BIOS configured, but they are mutually exclusive.", field.String()),
			Field:   field.String(),
		})
	}

	return causes
}

func validateFirmwareACPI(field *k8sfield.Path, spec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause

	if spec.Domain.Firmware == nil || spec.Domain.Firmware.ACPI == nil {
		return causes
	}

	acpi := spec.Domain.Firmware.ACPI
	if acpi.SlicNameRef == "" && acpi.MsdmNameRef == "" {
		return append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("ACPI was set but no SLIC nor MSDM volume reference was set"),
			Field:   field.String(),
		})
	}

	causes = append(causes, validateACPIRef(field, acpi.SlicNameRef, spec.Volumes, "slicNameRef")...)
	causes = append(causes, validateACPIRef(field, acpi.MsdmNameRef, spec.Volumes, "msdmNameRef")...)
	return causes
}

func validateACPIRef(field *k8sfield.Path, nameRef string, volumes []v1.Volume, fieldName string) []metav1.StatusCause {
	if nameRef == "" {
		return nil
	}

	for _, volume := range volumes {
		if nameRef != volume.Name {
			continue
		}

		if volume.Secret != nil {
			return nil
		}

		return []metav1.StatusCause{{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s refers to Volume of unsupported type.", field.String()),
			Field:   field.Child(fieldName).String(),
		}}
	}

	return []metav1.StatusCause{{
		Type:    metav1.CauseTypeFieldValueInvalid,
		Message: fmt.Sprintf("%s does not have a matching Volume.", field.String()),
		Field:   field.Child(fieldName).String(),
	}}
}

func validateFirmware(field *k8sfield.Path, firmware *v1.Firmware) []metav1.StatusCause {
	var causes []metav1.StatusCause

	if firmware != nil {
		causes = append(causes, validateBootloader(field.Child("bootloader"), firmware.Bootloader)...)
		causes = append(causes, validateKernelBoot(field.Child("kernelBoot"), firmware.KernelBoot)...)
	}

	return causes
}

func validateDomainSpec(field *k8sfield.Path, spec *v1.DomainSpec) []metav1.StatusCause {
	var causes []metav1.StatusCause

	causes = append(causes, storageadmitters.ValidateDisks(field.Child("devices").Child("disks"), spec.Devices.Disks)...)
	causes = append(causes, validateFirmware(field.Child("firmware"), spec.Firmware)...)

	if secureBootEnabled(spec.Firmware) && !smmFeatureEnabled(spec.Features) {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s has EFI SecureBoot enabled. SecureBoot requires SMM, which is currently disabled.", field.String()),
			Field:   field.String(),
		})
	}

	return causes
}

func validateAccessCredentials(field *k8sfield.Path, accessCredentials []v1.AccessCredential, volumes []v1.Volume) []metav1.StatusCause {
	var causes []metav1.StatusCause

	hasNoCloudVolume := false
	for _, volume := range volumes {
		if volume.CloudInitNoCloud != nil {
			hasNoCloudVolume = true
			break
		}
	}
	hasConfigDriveVolume := false
	for _, volume := range volumes {
		if volume.CloudInitConfigDrive != nil {
			hasConfigDriveVolume = true
			break
		}
	}

	for idx, accessCred := range accessCredentials {

		count := 0
		// one access cred type must be selected
		if accessCred.SSHPublicKey != nil {
			count++

			sourceCount := 0
			methodCount := 0
			if accessCred.SSHPublicKey.Source.Secret != nil {
				sourceCount++
			}

			if accessCred.SSHPublicKey.PropagationMethod.NoCloud != nil {
				methodCount++
				if !hasNoCloudVolume {
					causes = append(causes, metav1.StatusCause{
						Type:    metav1.CauseTypeFieldValueInvalid,
						Message: fmt.Sprintf("%s requires a noCloud volume to exist when the noCloud propagationMethod is in use.", field.Index(idx).String()),
						Field:   field.Index(idx).Child("sshPublicKey", "propagationMethod").String(),
					})

				}
			}

			if accessCred.SSHPublicKey.PropagationMethod.ConfigDrive != nil {
				methodCount++
				if !hasConfigDriveVolume {
					causes = append(causes, metav1.StatusCause{
						Type:    metav1.CauseTypeFieldValueInvalid,
						Message: fmt.Sprintf("%s requires a configDrive volume to exist when the configDrive propagationMethod is in use.", field.Index(idx).String()),
						Field:   field.Index(idx).Child("sshPublicKey", "propagationMethod").String(),
					})

				}
			}

			if accessCred.SSHPublicKey.PropagationMethod.QemuGuestAgent != nil {

				if len(accessCred.SSHPublicKey.PropagationMethod.QemuGuestAgent.Users) == 0 {
					causes = append(causes, metav1.StatusCause{
						Type:    metav1.CauseTypeFieldValueInvalid,
						Message: fmt.Sprintf("%s requires at least one user to be present in the users list", field.Index(idx).String()),
						Field:   field.Index(idx).Child("sshPublicKey", "propagationMethod", "qemuGuestAgent", "users").String(),
					})
				}

				methodCount++
			}

			if sourceCount != 1 {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s must have exactly one source set", field.Index(idx).String()),
					Field:   field.Index(idx).Child("sshPublicKey", "source").String(),
				})
			}
			if methodCount != 1 {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s must have exactly one propagationMethod set", field.Index(idx).String()),
					Field:   field.Index(idx).Child("sshPublicKey", "propagationMethod").String(),
				})
			}
		}

		if accessCred.UserPassword != nil {
			count++

			sourceCount := 0
			methodCount := 0
			if accessCred.UserPassword.Source.Secret != nil {
				sourceCount++
			}

			if accessCred.UserPassword.PropagationMethod.QemuGuestAgent != nil {
				methodCount++
			}

			if sourceCount != 1 {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s must have exactly one source set", field.Index(idx).String()),
					Field:   field.Index(idx).Child("userPassword", "source").String(),
				})
			}
			if methodCount != 1 {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s must have exactly one propagationMethod set", field.Index(idx).String()),
					Field:   field.Index(idx).Child("userPassword", "propagationMethod").String(),
				})
			}
		}

		if count != 1 {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: fmt.Sprintf("%s must have exactly one access credential type set", field.Index(idx).String()),
				Field:   field.Index(idx).String(),
			})
		}

	}

	return causes
}

func validateVolumes(field *k8sfield.Path, volumes []v1.Volume, config *virtconfig.ClusterConfig) []metav1.StatusCause {
	var causes []metav1.StatusCause
	nameMap := make(map[string]int)

	// check that we have max 1 instance of below disks
	serviceAccountVolumeCount := 0
	downwardMetricVolumeCount := 0
	memoryDumpVolumeCount := 0

	for idx, volume := range volumes {
		// verify name is unique
		otherIdx, ok := nameMap[volume.Name]
		if !ok {
			nameMap[volume.Name] = idx
		} else {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: fmt.Sprintf("%s and %s must not have the same Name.", field.Index(idx).String(), field.Index(otherIdx).String()),
				Field:   field.Index(idx).Child("name").String(),
			})
		}

		// Verify exactly one source is set
		volumeSourceSetCount := 0
		if volume.PersistentVolumeClaim != nil {
			volumeSourceSetCount++
		}
		if volume.Sysprep != nil {
			volumeSourceSetCount++
		}
		if volume.CloudInitNoCloud != nil {
			volumeSourceSetCount++
		}
		if volume.CloudInitConfigDrive != nil {
			volumeSourceSetCount++
		}
		if volume.ContainerDisk != nil {
			volumeSourceSetCount++
		}
		if volume.Ephemeral != nil {
			volumeSourceSetCount++
		}
		if volume.EmptyDisk != nil {
			volumeSourceSetCount++
		}
		if volume.HostDisk != nil {
			volumeSourceSetCount++
		}
		if volume.DataVolume != nil {
			if volume.DataVolume.Name == "" {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueRequired,
					Message: "DataVolume 'name' must be set",
					Field:   field.Index(idx).Child("name").String(),
				})
			}
			volumeSourceSetCount++
		}
		if volume.ConfigMap != nil {
			volumeSourceSetCount++
		}
		if volume.Secret != nil {
			volumeSourceSetCount++
		}
		if volume.DownwardAPI != nil {
			volumeSourceSetCount++
		}
		if volume.ServiceAccount != nil {
			volumeSourceSetCount++
			serviceAccountVolumeCount++
		}
		if volume.DownwardMetrics != nil {
			downwardMetricVolumeCount++
			volumeSourceSetCount++
		}
		if volume.MemoryDump != nil {
			memoryDumpVolumeCount++
			volumeSourceSetCount++
		}

		if volumeSourceSetCount != 1 {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: fmt.Sprintf("%s must have exactly one source type set", field.Index(idx).String()),
				Field:   field.Index(idx).String(),
			})
		}

		// Verify cloud init data is within size limits
		if volume.CloudInitNoCloud != nil || volume.CloudInitConfigDrive != nil {
			var userDataSecretRef, networkDataSecretRef *k8sv1.LocalObjectReference
			var dataSourceType, userData, userDataBase64, networkData, networkDataBase64 string
			if volume.CloudInitNoCloud != nil {
				dataSourceType = "cloudInitNoCloud"
				userDataSecretRef = volume.CloudInitNoCloud.UserDataSecretRef
				userDataBase64 = volume.CloudInitNoCloud.UserDataBase64
				userData = volume.CloudInitNoCloud.UserData
				networkDataSecretRef = volume.CloudInitNoCloud.NetworkDataSecretRef
				networkDataBase64 = volume.CloudInitNoCloud.NetworkDataBase64
				networkData = volume.CloudInitNoCloud.NetworkData
			} else if volume.CloudInitConfigDrive != nil {
				dataSourceType = "cloudInitConfigDrive"
				userDataSecretRef = volume.CloudInitConfigDrive.UserDataSecretRef
				userDataBase64 = volume.CloudInitConfigDrive.UserDataBase64
				userData = volume.CloudInitConfigDrive.UserData
				networkDataSecretRef = volume.CloudInitConfigDrive.NetworkDataSecretRef
				networkDataBase64 = volume.CloudInitConfigDrive.NetworkDataBase64
				networkData = volume.CloudInitConfigDrive.NetworkData
			}

			userDataLen := 0
			userDataSourceCount := 0
			networkDataLen := 0
			networkDataSourceCount := 0

			if userDataSecretRef != nil && userDataSecretRef.Name != "" {
				userDataSourceCount++
			}
			if userDataBase64 != "" {
				userDataSourceCount++
				userData, err := base64.StdEncoding.DecodeString(userDataBase64)
				if err != nil {
					causes = append(causes, metav1.StatusCause{
						Type:    metav1.CauseTypeFieldValueInvalid,
						Message: fmt.Sprintf("%s.%s.userDataBase64 is not a valid base64 value.", field.Index(idx).Child(dataSourceType, "userDataBase64").String(), dataSourceType),
						Field:   field.Index(idx).Child(dataSourceType, "userDataBase64").String(),
					})
				}
				userDataLen = len(userData)
			}
			if userData != "" {
				userDataSourceCount++
				userDataLen = len(userData)
			}

			if userDataSourceCount > 1 {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s must have only one userdatasource set.", field.Index(idx).Child(dataSourceType).String()),
					Field:   field.Index(idx).Child(dataSourceType).String(),
				})
			}

			if userDataLen > cloudInitUserMaxLen {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s userdata exceeds %d byte limit. Should use UserDataSecretRef for larger data.", field.Index(idx).Child(dataSourceType).String(), cloudInitUserMaxLen),
					Field:   field.Index(idx).Child(dataSourceType).String(),
				})
			}

			if networkDataSecretRef != nil && networkDataSecretRef.Name != "" {
				networkDataSourceCount++
			}
			if networkDataBase64 != "" {
				networkDataSourceCount++
				networkData, err := base64.StdEncoding.DecodeString(networkDataBase64)
				if err != nil {
					causes = append(causes, metav1.StatusCause{
						Type:    metav1.CauseTypeFieldValueInvalid,
						Message: fmt.Sprintf("%s.%s.networkDataBase64 is not a valid base64 value.", field.Index(idx).Child(dataSourceType, "networkDataBase64").String(), dataSourceType),
						Field:   field.Index(idx).Child(dataSourceType, "networkDataBase64").String(),
					})
				}
				networkDataLen = len(networkData)
			}
			if networkData != "" {
				networkDataSourceCount++
				networkDataLen = len(networkData)
			}

			if networkDataSourceCount > 1 {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s must have only one networkdata source set.", field.Index(idx).Child(dataSourceType).String()),
					Field:   field.Index(idx).Child(dataSourceType).String(),
				})
			}

			if networkDataLen > cloudInitNetworkMaxLen {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s networkdata exceeds %d byte limit. Should use NetworkDataSecretRef for larger data.", field.Index(idx).Child(dataSourceType).String(), cloudInitNetworkMaxLen),
					Field:   field.Index(idx).Child(dataSourceType).String(),
				})
			}

			if userDataSourceCount == 0 && networkDataSourceCount == 0 {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s must have at least one userdatasource or one networkdatasource set.", field.Index(idx).Child(dataSourceType).String()),
					Field:   field.Index(idx).Child(dataSourceType).String(),
				})
			}
		}

		if volume.DownwardMetrics != nil && !config.DownwardMetricsEnabled() {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: "downwardMetrics disks are not allowed: DownwardMetrics feature gate is not enabled.",
				Field:   field.Index(idx).String(),
			})
		}

		// validate HostDisk data
		if hostDisk := volume.HostDisk; hostDisk != nil {
			if !config.HostDiskEnabled() {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: "HostDisk feature gate is not enabled",
					Field:   field.Index(idx).String(),
				})
			}
			if hostDisk.Path == "" {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueNotFound,
					Message: fmt.Sprintf("%s is required for hostDisk volume", field.Index(idx).Child("hostDisk", "path").String()),
					Field:   field.Index(idx).Child("hostDisk", "path").String(),
				})
			}

			if hostDisk.Type != v1.HostDiskExists && hostDisk.Type != v1.HostDiskExistsOrCreate {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s has invalid value '%s', allowed are '%s' or '%s'", field.Index(idx).Child("hostDisk", "type").String(), hostDisk.Type, v1.HostDiskExists, v1.HostDiskExistsOrCreate),
					Field:   field.Index(idx).Child("hostDisk", "type").String(),
				})
			}

			// if disk.img already exists and user knows that by specifying type 'Disk' it is pointless to set capacity
			if hostDisk.Type == v1.HostDiskExists && !hostDisk.Capacity.IsZero() {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf("%s is allowed to pass only with %s equal to '%s'", field.Index(idx).Child("hostDisk", "capacity").String(), field.Index(idx).Child("hostDisk", "type").String(), v1.HostDiskExistsOrCreate),
					Field:   field.Index(idx).Child("hostDisk", "capacity").String(),
				})
			}
		}

		if volume.ConfigMap != nil {
			if volume.ConfigMap.LocalObjectReference.Name == "" {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf(requiredFieldFmt, field.Index(idx).Child("configMap", "name").String()),
					Field:   field.Index(idx).Child("configMap", "name").String(),
				})
			}
		}

		if volume.Secret != nil {
			if volume.Secret.SecretName == "" {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf(requiredFieldFmt, field.Index(idx).Child("secret", "secretName").String()),
					Field:   field.Index(idx).Child("secret", "secretName").String(),
				})
			}
		}

		if volume.ServiceAccount != nil {
			if volume.ServiceAccount.ServiceAccountName == "" {
				causes = append(causes, metav1.StatusCause{
					Type:    metav1.CauseTypeFieldValueInvalid,
					Message: fmt.Sprintf(requiredFieldFmt, field.Index(idx).Child("serviceAccount", "serviceAccountName").String()),
					Field:   field.Index(idx).Child("serviceAccount", "serviceAccountName").String(),
				})
			}
		}
	}

	if serviceAccountVolumeCount > 1 {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s must have max one serviceAccount volume set", field.String()),
			Field:   field.String(),
		})
	}

	if downwardMetricVolumeCount > 1 {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s must have max one downwardMetric volume set", field.String()),
			Field:   field.String(),
		})
	}
	if memoryDumpVolumeCount > 1 {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("%s must have max one memory dump volume set", field.String()),
			Field:   field.String(),
		})
	}

	return causes
}
