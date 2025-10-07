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
    v1 "kubevirt.io/api/core/v1"
    "kubevirt.io/client-go/log"

    "kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

type HyperVLayeredHypervisor struct{}

func (h *HyperVLayeredHypervisor) AdjustDomain(vmi *v1.VirtualMachineInstance, domain *api.Domain) {
	if domain == nil {
		return
	}
	domain.Spec.Type = "hyperv"

	// If user did not request a specific CPU model (converter likely filled host-model),
	// force a stable minimal baseline for mshv to reduce feature-surface while debugging.
	// NOTE: virtwrap api.DomainSpec.CPU is a value; we only mutate fields.
	//if domain.Spec.CPU.Mode == "" || domain.Spec.CPU.Mode == "host-model" || domain.Spec.CPU.Model == "" {
		// Use libvirt custom mode with qemu64 model.
		domain.Spec.CPU.Mode = "custom"
		domain.Spec.CPU.Model = "qemu64"
	//}

	log.Log.Infof("Adjusting domain for HyperV Layered (name=%s, cpuMode=%s, cpuModel=%s)", domain.Spec.Name, domain.Spec.CPU.Mode, domain.Spec.CPU.Model)
}

func (*HyperVLayeredHypervisor) GetDevice() string {
	return "mshv"
}
