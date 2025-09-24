/*
 * This file is part of the KubeVirt project
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
	"fmt"

	k8sv1 "k8s.io/api/core/v1"
	"kubevirt.io/client-go/log"

	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
)

// HypervisorContext represents a hypervisor context with its configuration
type HypervisorContext struct {
	Name       string
	DevicePath string
	DomainType string
}

const (
	// KVM represents the KVM hypervisor
	KVM string = "kvm"
	// HYPERVLAYERED represents the Microsoft Hyper-V
	HYPERVLAYERED string = "mshv"
)

var hypervisorContexts map[string]*HypervisorContext = map[string]*HypervisorContext{
	KVM: {
		Name:       KVM,
		DevicePath: "/dev/kvm",
		DomainType: "kvm",
	},
	HYPERVLAYERED: {
		Name:       HYPERVLAYERED,
		DevicePath: "/dev/mshv",
		DomainType: "hyperv",
	},
}

// GetCurrentHypervisorContext returns the current hypervisor context
func GetCurrentHypervisorContext(clusterConfig *virtconfig.ClusterConfig) *HypervisorContext {
	log.Log.Infof("Checking HyperVLayered feature gate status...")

	// Debug: Print all feature gates
	config := clusterConfig.GetConfig()
	if config.DeveloperConfiguration != nil && config.DeveloperConfiguration.FeatureGates != nil {
		log.Log.Infof("Found feature gates: %v", config.DeveloperConfiguration.FeatureGates)
	} else {
		log.Log.Infof("No feature gates found in config")
	}

	enabled := clusterConfig.HyperVLayeredEnabled()
	log.Log.Infof("HyperVLayeredEnabled() returned: %t", enabled)

	hvCtx, _ := GetHypervisorContextByName(KVM)
	if enabled {
		hvCtx, _ = GetHypervisorContextByName(HYPERVLAYERED)
		log.Log.Infof("HyperVLayered feature gate enabled, selecting hypervisor: %s", HYPERVLAYERED)
	} else {
		log.Log.Infof("HyperVLayered feature gate disabled, selecting hypervisor: %s", KVM)
	}
	return hvCtx
}

func GetHypervisorContextByName(name string) (*HypervisorContext, error) {
	ctx, exists := hypervisorContexts[name]
	if !exists {
		return nil, fmt.Errorf("hypervisor context not found for name: %s", name)
	}
	return ctx, nil
}

func (h *HypervisorContext) K8sResourceName() k8sv1.ResourceName {
	return k8sv1.ResourceName(fmt.Sprintf("devices.kubevirt.io/%s", h.Name))
}
