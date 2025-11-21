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
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfield "k8s.io/apimachinery/pkg/util/validation/field"

	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/hooks"
	webhookutils "kubevirt.io/kubevirt/pkg/util/webhooks"
	"kubevirt.io/kubevirt/pkg/virt-api/webhooks"
	hypervisor_validator "kubevirt.io/kubevirt/pkg/virt-api/webhooks/validating-webhook/admitters/hypervisor"
	base_validator "kubevirt.io/kubevirt/pkg/virt-api/webhooks/validating-webhook/admitters/hypervisor/base"
	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
	"kubevirt.io/kubevirt/pkg/virt-config/featuregate"
)

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
	baseValidator := base_validator.BaseValidator{}
	causes = append(causes, validator.ValidateVirtualMachineInstanceSpec(k8sfield.NewPath("spec"), &vmi.Spec, admitter.ClusterConfig)...)
	causes = append(causes, validator.ValidateVirtualMachineInstancePerArch(k8sfield.NewPath("spec"), &vmi.Spec, admitter.ClusterConfig)...)

	causes = append(causes, baseValidator.ValidateVirtualMachineInstanceSpecVolumeDisks(k8sfield.NewPath("spec"), &vmi.Spec)...)
	causes = append(causes, baseValidator.ValidateVirtualMachineInstanceMandatoryFields(k8sfield.NewPath("spec"), &vmi.Spec)...)
	// TODO Why is hyperv validation logic in a separate location?
	causes = append(causes, webhooks.ValidateVirtualMachineInstanceHyperv(k8sfield.NewPath("spec").Child("domain").Child("features").Child("hyperv"), &vmi.Spec)...)

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
