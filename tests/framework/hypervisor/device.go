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

package hypervisor

import (
	"strings"

	k8sv1 "k8s.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"

	"kubevirt.io/kubevirt/pkg/hypervisor"
	"kubevirt.io/kubevirt/pkg/virt-config/featuregate"
	"kubevirt.io/kubevirt/tests/libkubevirt"
)

// GetDevice returns the appropriate hypervisor device resource name
// based on the current KubeVirt configuration using the hypervisor package.
func GetDevice(virtClient kubecli.KubevirtClient) k8sv1.ResourceName {
	// Check if HyperVLayered feature gate is enabled
	kv := libkubevirt.GetCurrentKv(virtClient)
	hypervisorName := hypervisor.KVM // Default to KVM

	if kv.Spec.Configuration.DeveloperConfiguration != nil {
		featureGates := kv.Spec.Configuration.DeveloperConfiguration.FeatureGates
		for _, fg := range featureGates {
			if fg == strings.ToLower(featuregate.HyperVLayered) {
				hypervisorName = hypervisor.HYPERVLAYERED
				break
			}
		}
	}

	// Get hypervisor context using the hypervisor package
	hvContext, err := hypervisor.GetHypervisorContextByName(hypervisorName)
	if err != nil {
		// Fallback to KVM if error occurs
		hvContext, _ = hypervisor.GetHypervisorContextByName(hypervisor.KVM)
	}

	return hvContext.K8sResourceName()
}
